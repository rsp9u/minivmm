package minivmm

import (
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	dhcp "github.com/krolaw/dhcp4"
	"github.com/krolaw/dhcp4/conn"
)

func parseNameServers() []byte {
	servers := C.NameServers
	addresses := []byte{}
	for _, serverIP := range servers {
		ip := net.ParseIP(serverIP)
		ip = ip.To4()
		if ip == nil {
			panic("could not parse the string as IPv4: " + serverIP)
		}
		addresses = append(addresses, ip...)
	}
	return addresses
}

// ServeDHCP serves DHCP.
func ServeDHCP() {
	nwInfo, err := newNetworkInfo()
	if err != nil {
		log.Fatal(err)
	}

	dnsIPs := parseNameServers()
	handler := &dhcpHandler{
		ip:            nwInfo.gwIP,
		start:         nwInfo.startIP,
		leaseRange:    (1 << uint(32-nwInfo.cidrLen)) - 4,
		leaseDuration: 2 * time.Hour,
		leases:        make(map[int]lease, 32),
		macVendor:     "52:54:00",
		options: dhcp.Options{
			dhcp.OptionSubnetMask:       []byte(nwInfo.cidrIPNet.Mask),
			dhcp.OptionRouter:           []byte(nwInfo.gwIP),
			dhcp.OptionDomainNameServer: dnsIPs,
		},
	}

	pc, err := conn.NewUDP4BoundListener(vethNames[0], ":67")
	if err != nil {
		panic(err)
	}
	log.Fatal(dhcp.Serve(pc, handler))
}

type lease struct {
	nic    string    // Client's CHAddr
	expiry time.Time // When the lease expires
}

type dhcpHandler struct {
	ip            net.IP        // Server IP to use
	options       dhcp.Options  // Options to send to DHCP Clients
	start         net.IP        // Start of IP range to distribute
	leaseRange    int           // Number of IPs to distribute (starting from start)
	leaseDuration time.Duration // Lease period
	leases        map[int]lease // Map to keep track of leases
	macVendor     string
}

func (h *dhcpHandler) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) (d dhcp.Packet) {
	switch msgType {

	case dhcp.Discover:
		log.Println("[dhcp] INFO macaddr:", p.CHAddr().String())
		if !strings.HasPrefix(p.CHAddr().String(), h.macVendor) {
			log.Println("[dhcp] WARN received unexpected vendor's DISCOVER, discard it")
			return
		}
		free, nic := -1, p.CHAddr().String()
		for i, v := range h.leases { // Find previous lease
			if v.nic == nic {
				free = i
				goto reply
			}
		}
		if free = h.freeLease(); free == -1 {
			return
		}
	reply:
		return dhcp.ReplyPacket(p, dhcp.Offer, h.ip, dhcp.IPAdd(h.start, free), h.leaseDuration,
			h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))

	case dhcp.Request:
		if server, ok := options[dhcp.OptionServerIdentifier]; ok && !net.IP(server).Equal(h.ip) {
			log.Println("[dhcp] WARN missing to OptionServerIdentifier or mismatch")
		}
		reqIP := net.IP(options[dhcp.OptionRequestedIPAddress])
		if reqIP == nil {
			reqIP = net.IP(p.CIAddr())
		}
		log.Println("[dhcp] INFO ipaddr:", reqIP.String())

		if len(reqIP) == 4 && !reqIP.Equal(net.IPv4zero) {
			if leaseNum := dhcp.IPRange(h.start, reqIP) - 1; leaseNum >= 0 && leaseNum < h.leaseRange {
				if l, exists := h.leases[leaseNum]; !exists || l.nic == p.CHAddr().String() {
					// update VM metadata
					VMIPAddressUpdateChan <- &VMMetaData{
						IPAddress:  reqIP.String(),
						MacAddress: p.CHAddr().String(),
					}
					// lease
					h.leases[leaseNum] = lease{nic: p.CHAddr().String(), expiry: time.Now().Add(h.leaseDuration)}
					return dhcp.ReplyPacket(p, dhcp.ACK, h.ip, reqIP, h.leaseDuration,
						h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
				}
			}
		}
		return dhcp.ReplyPacket(p, dhcp.NAK, h.ip, nil, 0, nil)

	case dhcp.Release, dhcp.Decline:
		nic := p.CHAddr().String()
		for i, v := range h.leases {
			if v.nic == nic {
				delete(h.leases, i)
				break
			}
		}
	}
	return nil
}

func (h *dhcpHandler) freeLease() int {
	now := time.Now()
	b := rand.Intn(h.leaseRange) // Try random first
	for _, v := range [][]int{[]int{b, h.leaseRange}, []int{0, b}} {
		for i := v[0]; i < v[1]; i++ {
			if l, ok := h.leases[i]; !ok || l.expiry.Before(now) {
				return i
			}
		}
	}
	return -1
}
