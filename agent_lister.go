package minivmm

import (
	"context"
	"fmt"
	"github.com/grandcat/zeroconf"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	MDnsServiceName       = "_minivmm._tcp"
	MDnsApiUrlTxt         = "api"
	MDnsDomain            = "local."
	MDnsDiscoveryWaitTime = 3000 // ms
	MDnsNodeTTL           = 10   // sec
)
var Agents AgentLister

type AgentLister interface {
	GetAgents() []string
	Cleanup()
}

func InitAgentLister() (AgentLister, error) {
	if !C.NoAgentsDiscover {
		l, err := NewZeroconfAgentLister(C.Origin, C.Port)
		if err != nil {
			return nil, err
		}
		return l, nil
	}
	return &StaticAgentLister{}, nil
}

func GetApiUrl(origin string) string {
	if !strings.HasSuffix(origin, "/api/v1") {
		return origin + "/api/v1"
	}
	return origin
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
	apiUrlTxt := fmt.Sprintf("%s=%s=%s", MDnsApiUrlTxt, hostname, GetApiUrl(originUrl))
	log.Printf("[zeroconf] my service text: %s\n", apiUrlTxt)

	server, err := zeroconf.Register(origin.Host, MDnsServiceName, MDnsDomain, port, []string{apiUrlTxt}, nil)
	if err != nil {
		return nil, err
	}

	return &ZeroconfAgentLister{
		server:    server,
		agentList: map[string]time.Time{},
	}, nil
}

func (l *ZeroconfAgentLister) refreshAgent() error {
	log.Println("[zeroconf] refreshing agent list")

	now := time.Now()

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Printf("[zeroconf] here comes an entry: %v\n", entry)
			for _, txt := range entry.Text {
				if strings.HasPrefix(txt, MDnsApiUrlTxt) {
					keyval := strings.SplitN(txt, "=", 2)
					agentUrl := keyval[1]
					l.agentList[agentUrl] = now
					log.Printf("[zeroconf] registered an agent (name: %s, url: %s)\n", keyval[0], keyval[1])
				}
			}
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(MDnsDiscoveryWaitTime))
	defer cancel()

	err = resolver.Browse(ctx, MDnsServiceName, MDnsDomain, entries)
	if err != nil {
		return err
	}
	<-ctx.Done()

	// cleanup old agents
	for agentUrl, lastSeen := range l.agentList {
		if int(now.Sub(lastSeen).Seconds()) > MDnsNodeTTL {
			delete(l.agentList, agentUrl)
		}
	}

	return nil
}

func (l *ZeroconfAgentLister) GetAgents() []string {
	go l.refreshAgent()

	agents := make([]string, 0, len(l.agentList))
	for key := range l.agentList {
		agents = append(agents, key)
	}
	return agents
}

func (l ZeroconfAgentLister) Cleanup() {
	l.server.Shutdown()
}
