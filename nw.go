package minivmm

import (
	"net"
	"os"

	"github.com/apparentlymart/go-cidr/cidr"
)

type vmNetworkInfo struct {
	cidrIPNet *net.IPNet
	brIP      net.IP
	gwIP      net.IP
	startIP   net.IP
}

var (
	nsName    = "minivmm"
	brName    = "br-minivmm"
	vethNames = []string{"minivmm", "minivmm-peer"}
)

func newNetworkInfo() (*vmNetworkInfo, error) {
	_, cidrIPNet, err := net.ParseCIDR(os.Getenv(EnvSubnetCIDR))
	if err != nil {
		return nil, err
	}

	cnt := int(cidr.AddressCount(cidrIPNet) - 1)
	brIP, err := cidr.Host(cidrIPNet, 1)
	if err != nil {
		return nil, err
	}
	gwIP, err := cidr.Host(cidrIPNet, cnt-1)
	if err != nil {
		return nil, err
	}
	startIP, err := cidr.Host(cidrIPNet, 2)
	if err != nil {
		return nil, err
	}

	return &vmNetworkInfo{
		cidrIPNet,
		brIP,
		gwIP,
		startIP,
	}, nil
}

// InitNetns initializes netns.
func InitNetns() error {
	return Execs([][]string{
		{"sudo", "ip", "netns", "add", nsName},

		{"sudo", "ip", "link", "add", vethNames[0], "type", "veth", "peer", "name", vethNames[1]},
		{"sudo", "ip", "link", "set", "netns", nsName, "dev", vethNames[1]},

		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "add", brName, "type", "bridge"},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "master", brName, "dev", vethNames[1]},
	})
}

// ResetNetns removes all netns and interfaces.
func ResetNetns() error {
	return Execs([][]string{
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "down", "dev", vethNames[1]},
		{"sudo", "ip", "link", "set", "down", "dev", vethNames[0]},

		{"sudo", "ip", "link", "delete", "dev", vethNames[0]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "delete", brName},

		{"sudo", "ip", "netns", "delete", nsName},
	})
}

// StartNetwork set up interfaces.
func StartNetwork() error {
	nwInfo, err := newNetworkInfo()
	if err != nil {
		return err
	}

	return Execs([][]string{
		{"sudo", "ip", "link", "set", "up", "dev", vethNames[0]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "up", "dev", vethNames[1]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "promisc", "on", "dev", vethNames[1]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "up", "dev", brName},

		{"sudo", "ip", "addr", "add", nwInfo.gwIP.String(), "dev", vethNames[0]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "addr", "add", nwInfo.brIP.String(), "dev", brName},
	})
}
