package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"minivmm"
)

type agent struct {
	Name string `json:"name"`
	API  string `json:"api"`
}

// HandleAgents handles agent resource request.
func HandleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ListAgents(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// ListAgents returns a list of agent defined in environment variable.
func ListAgents(w http.ResponseWriter, r *http.Request) {
	agents := []*agent{}
	definedSelf := false

	hostname, _ := os.Hostname()
	for _, a := range minivmm.C.Agents.GetAgents() {
		name := strings.Split(a, "=")[0]
		api := strings.Split(a, "=")[1]
		agents = append(agents, &agent{name, api})
		if name == hostname {
			definedSelf = true
		}
	}

	if !definedSelf {
		api := minivmm.C.Origin + "/api/v1/"
		agents = append(agents, &agent{hostname, api})
	}

	ret := map[string][]*agent{"agents": agents}
	b, _ := json.Marshal(ret)
	w.Write(b)
}
