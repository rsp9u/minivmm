package minivmm

var (
	nsName    = "minivmm"
	brName    = "br-minivmm"
	vethNames = []string{"minivmm", "minivmm-peer"}
	cidr      = "192.168.200.0/24"
	brIP      = "192.168.200.1/24"
	gwIP      = "192.168.200.254/24"
	startIP   = "192.168.200.101"
)

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
	return Execs([][]string{
		{"sudo", "ip", "link", "set", "up", "dev", vethNames[0]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "up", "dev", vethNames[1]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "promisc", "on", "dev", vethNames[1]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "link", "set", "up", "dev", brName},

		{"sudo", "ip", "addr", "add", gwIP, "dev", vethNames[0]},
		{"sudo", "ip", "netns", "exec", nsName, "ip", "addr", "add", brIP, "dev", brName},
	})
}
