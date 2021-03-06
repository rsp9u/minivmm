package minivmm

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	stopChannels = make(map[string]chan struct{})

	nameToIP   = map[string]string{}
	ipChannels = make(map[string]map[string]chan struct{})
)

func proxyUDPStream(fromPort, toIP, toPort string) (*net.UDPConn, *net.UDPConn, error) {
	laddr, err := net.ResolveUDPAddr("udp", ":"+fromPort)
	if err != nil {
		log.Println("[forwarder] WARN ResolveUDPAddr error: ", err.Error())
		return nil, nil, err
	}

	src, err := net.ListenUDP("udp", laddr)
	if err != nil {
		log.Println("[forwarder] WARN ListenUDP error: ", err.Error())
		return nil, nil, err
	}

	raddr, err := net.ResolveUDPAddr("udp", toIP+":"+toPort)
	if err != nil {
		log.Println("[forwarder] WARN ResolveUDPAddr error: ", err.Error())
		return nil, nil, err
	}

	dst, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Println("[forwarder] WARN DialUDP error: ", err.Error())
		return nil, nil, err
	}

	go io.Copy(dst, src)
	go io.Copy(src, dst)

	return src, dst, nil
}

func proxyUDP(stopChan chan struct{}, id, fromPort, toName, toPort string) {
	ipChan := makeIPChannel(toName, id)
	defer deleteIPChannel(toName, id)

	for {
		toIP, err := resolveName(toName)
		if err != nil {
			log.Printf("[forwarder] WARN could not get IP address for %s\n", toName)
			return
		}

		src, dst, err := proxyUDPStream(fromPort, toIP, toPort)
		if err != nil {
			return
		}
		defer src.Close()
		defer dst.Close()

		// Wait for address updating or stopping
		select {
		case <-ipChan:
			src.Close()
			dst.Close()
			log.Println("[forwarder] INFO update udp forwarder dest address, reopen")
			continue
		case <-stopChan:
			log.Println("[forwarder] INFO shutdown udp proxy")
			return
		}
	}
}

func isUDPBindable(port string) error {
	laddr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		return err
	}
	c, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}
	defer c.Close()
	return nil
}

func proxyTCPSession(src net.Conn, toIP, toPort string) {
	dst, err := net.Dial("tcp", toIP+":"+toPort)
	if err != nil {
		log.Println("[forwarder] WARN dial error: ", err.Error())
		return
	}

	done := make(chan struct{})

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(dst, src)
		done <- struct{}{}
	}()

	go func() {
		defer src.Close()
		defer dst.Close()
		io.Copy(src, dst)
		done <- struct{}{}
	}()

	<-done
	<-done
}

func proxyTCP(stopChan chan struct{}, fromPort, toName, toPort string) {
	ln, err := net.Listen("tcp", ":"+fromPort)
	if err != nil {
		log.Println("[forwarder] WARN listen error: ", err.Error())
		return
	}
	defer ln.Close()

	for {
		// Accept in the background
		var conn net.Conn
		acc := make(chan struct{})
		go func() {
			for {
				conn, err = ln.Accept()
				if err != nil {
					if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
						log.Println("[forwarder] WARN listen temporary error: ", err.Error())
						continue
					} else {
						log.Println("[forwarder] INFO shutdown listen")
						return
					}
				}
				acc <- struct{}{}
				return
			}
		}()

		// Wait for accepting or stopping
		select {
		case <-acc:
			toIP, err := resolveName(toName)
			if err != nil {
				log.Printf("[forwarder] WARN could not get IP address for %s\n", toName)
				continue
			}
			go proxyTCPSession(conn, toIP, toPort)
		case <-stopChan:
			log.Println("[forwarder] INFO shutdown tcp proxy")
			return
		}
	}
}

func isTCPBindable(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer ln.Close()
	return nil
}

func resolveName(name string) (string, error) {
	for i := 0; i < 10; i++ {
		ip, ok := nameToIP[name]
		if ok {
			return ip, nil
		}
		time.Sleep(6 * time.Second)
		log.Printf("[forwarder] INFO waiting for resolution for %s..\n", name)
	}
	return "", errors.New("waiting for the address resolution is timed out")
}

func makeIPChannel(name, id string) chan struct{} {
	_, ok := ipChannels[name]
	if !ok {
		ipChannels[name] = map[string]chan struct{}{}
	}

	c := make(chan struct{})
	ipChannels[name][id] = c
	return c
}

func deleteIPChannel(name, id string) {
	delete(ipChannels[name], id)
}

// StartForward starts new forwarding.
func StartForward(proto, fromPort, toName, toPort string) error {
	id := generateForwardID(proto, fromPort)

	ch := make(chan struct{})
	if proto == "udp" {
		if err := isUDPBindable(fromPort); err != nil {
			return errors.Wrap(err, "failed to bind to udp port")
		}
		go proxyUDP(ch, id, fromPort, toName, toPort)
	} else {
		if err := isTCPBindable(fromPort); err != nil {
			return errors.Wrap(err, "failed to bind to tcp port")
		}
		go proxyTCP(ch, fromPort, toName, toPort)
	}
	stopChannels[id] = ch
	return nil
}

// StopForward stop forwarding.
func StopForward(proto, fromPort string) error {
	id := generateForwardID(proto, fromPort)

	c, ok := stopChannels[id]
	if !ok {
		return fmt.Errorf("unknown forwarding: %s", id)
	}
	c <- struct{}{}
	return nil
}

// ForwardMetaData is forwarding settings.
type ForwardMetaData struct {
	Owner       string `json:"owner"`
	Hypervisor  string `json:"hypervisor"`
	Proto       string `json:"proto"`
	FromPort    string `json:"from_port"`
	ToName      string `json:"to_name"`
	ToPort      string `json:"to_port"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func generateForwardID(proto, fromPort string) string {
	return proto + "-" + fromPort
}

// ResumeForwards resumes forwardings from file.
func ResumeForwards() error {
	// get existing VM's addresses
	vms, err := ListVMs()
	if err != nil {
		return err
	}
	for _, vm := range vms {
		UpdateIPAddressInForwarder(vm.Name, vm.IPAddress)
	}

	// resume forwards
	fws, err := ReadAllForwardFiles()
	if err != nil {
		return err
	}
	for _, f := range fws {
		err := StartForward(f.Proto, f.FromPort, f.ToName, f.ToPort)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateIPAddressInForwarder updates the IP address associated to VM.
func UpdateIPAddressInForwarder(name, ip string) {
	nameToIP[name] = ip

	channels, ok := ipChannels[name]
	if !ok {
		return
	}
	for _, c := range channels {
		c <- struct{}{}
	}
}

// WriteForwardFile creates or updates the forwarding settings file.
// The file name will be joined string of protocol and listen port.
func WriteForwardFile(fw *ForwardMetaData) error {
	recordPath := filepath.Join(C.ForwardDir, generateForwardID(fw.Proto, fw.FromPort)+".json")

	f, err := os.OpenFile(recordPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.Marshal(fw)
	if err != nil {
		return err
	}

	lockpath := recordPath + ".lock"
	WriteWithLock(f, lockpath, b)

	return nil
}

// RemoveForwardFile removes a forwarding settings file.
func RemoveForwardFile(fw *ForwardMetaData) error {
	recordPath := filepath.Join(C.ForwardDir, generateForwardID(fw.Proto, fw.FromPort)+".json")
	return os.Remove(recordPath)
}

// ReadAllForwardFiles returns a list of forwarding settings.
func ReadAllForwardFiles() ([]*ForwardMetaData, error) {
	dirEntries, err := os.ReadDir(C.ForwardDir)
	if err != nil {
		return nil, err
	}

	var ret []*ForwardMetaData
	for _, f := range dirEntries {
		if !f.IsDir() {
			fw, err := readForwardFileByFileName(f.Name())
			if err != nil {
				log.Println("Ignore ReadForwardFile error:", err)
				continue
			}
			ret = append(ret, fw)
		}
	}

	return ret, nil
}

// ReadForwardFile returns a forwarding setting.
func ReadForwardFile(proto, fromPort string) (*ForwardMetaData, error) {
	return readForwardFileByFileName(generateForwardID(proto, fromPort) + ".json")
}

// GetRandomForwardPort choices a random number in range and it's unused port as forward port.
func GetRandomForwardPort(proto string, rangeMin, rangeMax int) (string, error) {
	dirEntries, err := os.ReadDir(C.ForwardDir)
	if err != nil {
		return "", err
	}
	existsSet := map[string]struct{}{}
	for _, f := range dirEntries {
		if !f.IsDir() && strings.HasPrefix(f.Name(), proto) {
			n := f.Name()
			n = strings.TrimSuffix(n, ".json")
			existsSet[n] = struct{}{}
		}
	}

	for i := rangeMin; i <= rangeMax; i++ {
		id := generateForwardID(proto, strconv.Itoa(i))
		_, exists := existsSet[id]
		if !exists && checkPortIsBindable(proto, strconv.Itoa(i)) {
			return strconv.Itoa(i), nil
		}
	}

	return "", fmt.Errorf("failed to get forwarding port. it's exhausted.")
}

func checkPortIsBindable(proto, port string) bool {
	switch proto {
	case "tcp":
		conn, err := net.Listen(proto, ":"+port)
		if err != nil {
			log.Printf("failed a port bind check; %v", err)
			return false
		} else {
			conn.Close()
			return true
		}
	case "udp":
		addr, _ := net.ResolveUDPAddr("udp", ":"+port)
		conn, err := net.ListenUDP(proto, addr)
		if err != nil {
			log.Printf("failed a port bind check; %v", err)
			return false
		} else {
			conn.Close()
			return true
		}
	}
	return false
}

func readForwardFileByFileName(fileName string) (*ForwardMetaData, error) {
	fw := ForwardMetaData{}
	b, err := os.ReadFile(filepath.Join(C.ForwardDir, fileName))
	if err != nil {
		return nil, err
	}
	json.Unmarshal(b, &fw)
	return &fw, nil
}
