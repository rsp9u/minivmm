package minivmm

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"
)

var (
	stopChannels = make(map[string]chan struct{})

	nameToIP   = map[string]string{"localhost": "127.0.0.1"}
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
func StartForward(id, proto, fromPort, toName, toPort string) error {
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
func StopForward(id string) error {
	c, ok := stopChannels[id]
	if !ok {
		return fmt.Errorf("unknown forwarding: %s", id)
	}
	c <- struct{}{}
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
