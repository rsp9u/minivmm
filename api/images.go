package api

import (
	"encoding/json"
	"net/http"

	"minivmm"
)

type image struct {
	Name string `json:"name"`
}

// HandleImages handles image resource request.
func HandleImages(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ListImages(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// ListImages returns a list of images.
func ListImages(w http.ResponseWriter, r *http.Request) {
	imgs := []*image{}
	names := minivmm.ListBaseImages()
	for _, n := range names {
		imgs = append(imgs, &image{n})
	}
	ret := map[string][]*image{"images": imgs}
	b, _ := json.Marshal(ret)
	w.Write(b)
}
