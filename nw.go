package minivmm

import (
	"fmt"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
)

type vmNetworkInfo struct {
	cidrIPNet *net.IPNet
	cidrLen   int
	gwIP      net.IP
	startIP   net.IP
}

var (
	nsName    = "minivmm"
	brName    = "br-minivmm"
	vethNames = []string{"minivmm", "minivmm-peer"}
)

func newNetworkInfo() (*vmNetworkInfo, error) {
	_, cidrIPNet, err := net.ParseCIDR(C.SubnetCIDR)
	if err != nil {
		return nil, err
	}
	cidrLen, _ := cidrIPNet.Mask.Size()
	if cidrLen >= 30 {
		return nil, fmt.Errorf("Subnet size is too small (least '/29')")
	}

	cnt := int(cidr.AddressCount(cidrIPNet) - 1)
	gwIP, err := cidr.Host(cidrIPNet, cnt-1)
	if err != nil {
		return nil, err
	}
	startIP, err := cidr.Host(cidrIPNet, 1)
	if err != nil {
		return nil, err
	}

	return &vmNetworkInfo{
		cidrIPNet,
		cidrLen,
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

	ExecsIgnoreErr([][]string{
		{"sudo", "ip", "link", "set", "up", "dev", vethNames[0]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "up", "dev", vethNames[1]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "promisc", "on", "dev", vethNames[1]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "up", "dev", brName},

		{"sudo", "ip", "addr", "add", fmt.Sprintf("%s/%d", nwInfo.gwIP.String(), nwInfo.cidrLen), "dev", vethNames[0]},
	})
	return nil
}
