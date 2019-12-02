package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"minivmm"
)

var (
	recordPath = filepath.Join(minivmm.ForwardDir, "forward.json")
)

type forward struct {
	Owner       string `json:"owner"`
	Hypervisor  string `json:"hypervisor"`
	Proto       string `json:"proto"`
	FromPort    string `json:"from_port"`
	ToName      string `json:"to_name"`
	ToPort      string `json:"to_port"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

func parseForwardBody(body io.ReadCloser) *forward {
	defer body.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, body)

	var f forward
	json.Unmarshal(buf.Bytes(), &f)
	return &f
}

func uniq(f *forward) string {
	return f.Hypervisor + f.FromPort + f.ToName + f.ToPort
}

// HandleForwards handles forward resource request.
func HandleForwards(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ListForwards(w, r)
		return
	}
	if r.Method == http.MethodPost {
		CreateForward(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		DeleteForward(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// ListForwards returns a list of forwards.
func ListForwards(w http.ResponseWriter, r *http.Request) {
	// read forward list from file
	forwards := readFile()

	// filter by owner
	ownedForwards := []*forward{}
	for i := 0; i < len(forwards); i++ {
		if forwards[i].Owner != minivmm.GetUserName(r) {
			continue
		}
		ownedForwards = append(ownedForwards, &forwards[i])
	}

	// reponse
	ret := map[string][]*forward{"forwards": ownedForwards}
	b, _ := json.Marshal(ret)
	w.Write(b)
}

// CreateForward sets up forwarding and writes its metadata.
func CreateForward(w http.ResponseWriter, r *http.Request) {
	f := parseForwardBody(r.Body)
	f.Owner = minivmm.GetUserName(r)
	log.Println(f)

	err := minivmm.StartForward(uniq(f), f.Proto, f.FromPort, f.ToName, f.ToPort)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ret := map[string]string{"error": err.Error()}
		b, _ := json.Marshal(ret)
		w.Write(b)
		return
	}

	err = appendToFile(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ret := map[string]string{"error": err.Error()}
		b, _ := json.Marshal(ret)
		w.Write(b)
		return
	}
}

// DeleteForward shuts down forwarding and removes its metadata.
func DeleteForward(w http.ResponseWriter, r *http.Request) {
	f := parseForwardBody(r.Body)

	err := minivmm.StopForward(uniq(f))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ret := map[string]string{"error": err.Error()}
		b, _ := json.Marshal(ret)
		w.Write(b)
		return
	}

	err = removeFromFile(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ret := map[string]string{"error": err.Error()}
		b, _ := json.Marshal(ret)
		w.Write(b)
		return
	}
}

// ResumeForwards resumes forwardings from file.
func ResumeForwards() error {
	// get existing VM's addresses
	vms, err := minivmm.ListVMs()
	if err != nil {
		return err
	}
	for _, vm := range vms {
		minivmm.UpdateIPAddressInForwarder(vm.Name, vm.IPAddress)
	}

	// resume forwards
	fws := readFile()
	for _, f := range fws {
		err := minivmm.StartForward(uniq(&f), f.Proto, f.FromPort, f.ToName, f.ToPort)
		if err != nil {
			return err
		}
	}

	return nil
}

func appendToFile(fw *forward) error {
	fws := readFile()
	fws = append(fws, *fw)
	err := writeFile(fws)
	return err
}

func removeFromFile(rmFw *forward) error {
	fws := readFile()
	newFws := []forward{}
	for _, fw := range fws {
		if uniq(&fw) != uniq(rmFw) {
			newFws = append(newFws, fw)
		}
	}
	err := writeFile(newFws)

	return err
}

func readFile() []forward {
	f, err := os.Open(recordPath)
	if err != nil {
		return []forward{}
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, f)

	var fws []forward
	json.Unmarshal(buf.Bytes(), &fws)

	return fws
}

func writeFile(fws []forward) error {
	f, err := os.OpenFile(recordPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.Marshal(fws)
	if err != nil {
		return err
	}

	// NOTE: seriously, read lock is also needed.
	lockpath := recordPath + ".lock"
	minivmm.WriteWithLock(f, lockpath, b)

	return nil
}
