package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"minivmm"
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

	if f.FromPort == "" {
		rangeMin, rangeMax := portRangePerUser(minivmm.GetUserName(r))
		port, err := minivmm.GetRandomForwardPort(f.Proto, rangeMin, rangeMax)
		if err != nil {
			writeInternalServerError(err, w)
			return
		}
		f.FromPort = port
	}

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

	err := restrictForwardOperationByOwner(w, r, f.Proto, f.FromPort)
	if err != nil {
		return
	}

	err = minivmm.StopForward(f.Proto, f.FromPort)
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

func restrictForwardOperationByOwner(w http.ResponseWriter, r *http.Request, proto, fromPort string) error {
	metaData, err := minivmm.ReadForwardFile(proto, fromPort)
	if err != nil {
		writeInternalServerError(err, w)
		return err
	}

	if metaData.Owner != minivmm.GetUserName(r) {
		writeForbidden(w)
		return fmt.Errorf("forbidden")
	}

	return nil
}

func portRangePerUser(userName string) (int, int) {
	// Auto-numbering port range is from 30000 to 55999(30000+256*100-1).
	// The size of the range per user is 100, its range caluculated by the user name hash.
	hash := sha256.Sum256([]byte(userName))
	base := int(hash[0])
	return 100*base + 30000, 100*(base+1) - 1 + 30000
}
