package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"minivmm"
)

var (
	recordPath = filepath.Join(minivmm.ForwardDir, "forward.json")
)

func parseForwardBody(body io.ReadCloser) *minivmm.ForwardMetaData {
	defer body.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, body)

	var f minivmm.ForwardMetaData
	json.Unmarshal(buf.Bytes(), &f)
	return &f
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
	forwards, err := minivmm.ReadAllForwardFiles()
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	// filter by owner
	ownedForwards := []*minivmm.ForwardMetaData{}
	for i := 0; i < len(forwards); i++ {
		if forwards[i].Owner != minivmm.GetUserName(r) {
			continue
		}
		ownedForwards = append(ownedForwards, forwards[i])
	}

	// reponse
	ret := map[string][]*minivmm.ForwardMetaData{"forwards": ownedForwards}
	b, _ := json.Marshal(ret)
	w.Write(b)
}

// CreateForward sets up forwarding and writes its metadata.
func CreateForward(w http.ResponseWriter, r *http.Request) {
	f := parseForwardBody(r.Body)
	f.Owner = minivmm.GetUserName(r)
	log.Println(f)

	err := minivmm.StartForward(f.Proto, f.FromPort, f.ToName, f.ToPort)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	err = minivmm.WriteForwardFile(f)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}
}

// DeleteForward shuts down forwarding and removes its metadata.
func DeleteForward(w http.ResponseWriter, r *http.Request) {
	f := parseForwardBody(r.Body)

	err := minivmm.StopForward(f.Proto, f.FromPort)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	err = minivmm.RemoveForwardFile(f)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}
}
