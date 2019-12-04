package api

import (
	"encoding/json"
	"net/http"
)

func writeInternalServerError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	ret := map[string]string{"error": err.Error()}
	b, _ := json.Marshal(ret)
	w.Write(b)
}
