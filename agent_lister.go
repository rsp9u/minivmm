package minivmm

import (
    "time"
    "context"
    "log"
	"github.com/grandcat/zeroconf"
)

type AgentLister interface {
    GetAgents() []string
    Cleanup()
}

type StaticAgentLister struct{}

func (l StaticAgentLister) GetAgents() []string {
    return C.StaticAgents
}

func (l StaticAgentLister) Cleanup() {}

type ZeroconfAgentLister struct {
    server *zeroconf.Server
}

func NewZeroconfAgentLister(origin string, port int) (*ZeroconfAgentLister, error) {
	server, err := zeroconf.Register(origin, "_minivmm._tcp", "local.", port, []string{}, nil)
    if err != nil {
        return nil, err
    }

    return &ZeroconfAgentLister{
        server: server,
    }, nil
}

func (l ZeroconfAgentLister) GetAgents() []string {
    // maybe zeroconf library does not support continuous discovery
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
        log.Println(err)
        return []string{}
	}
	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Println(entry.ServiceRecord.Instance)
		}
		log.Println("No more entries.")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(1))
	defer cancel()
    err = resolver.Browse(ctx, "_minivmm._tcp", "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}
    <-ctx.Done()

    return []string{}
}

func (l ZeroconfAgentLister) Cleanup() {
	log.Println("Cleanup")
    l.server.Shutdown()
}
