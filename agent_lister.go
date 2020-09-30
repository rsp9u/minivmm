package minivmm

import (
	"context"
	"github.com/grandcat/zeroconf"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
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
	// agent-url to last-seen-time map
	agentList map[string]time.Time
}

func NewZeroconfAgentLister(originUrl string, port int) (*ZeroconfAgentLister, error) {
	origin, err := url.Parse(originUrl)
	if err != nil {
		return nil, err
	}
	hostname, _ := os.Hostname()

	server, err := zeroconf.Register(origin.Host, "_minivmm._tcp", "local.", port, []string{"api=" + hostname + "=" + originUrl}, nil)
	if err != nil {
		return nil, err
	}

	return &ZeroconfAgentLister{
		server:    server,
		agentList: map[string]time.Time{},
	}, nil
}

func (l *ZeroconfAgentLister) refreshAgent() error {
	now := time.Now()

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			for _, txt := range entry.Text {
				if strings.HasPrefix(txt, "api=") {
					keyval := strings.SplitN(txt, "=", 2)
					agentUrl := keyval[1]
					l.agentList[agentUrl] = now
				}
			}
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(3))
	defer cancel()

	err = resolver.Browse(ctx, "_minivmm._tcp", "local.", entries)
	if err != nil {
		return err
	}
	<-ctx.Done()

	// cleanup old agents
	for agentUrl, lastSeen := range l.agentList {
		if now.Sub(lastSeen).Seconds() > 10 {
			delete(l.agentList, agentUrl)
		}
	}

	return nil
}

func (l *ZeroconfAgentLister) GetAgents() []string {
	go l.refreshAgent()

	agents := make([]string, 0, len(l.agentList))
	for key := range l.agentList {
		// FIXME: origin, agentURL form unification
		agents = append(agents, key+"/api/v1")
	}
	return agents
}

func (l ZeroconfAgentLister) Cleanup() {
	log.Println("Cleanup")
	l.server.Shutdown()
}
