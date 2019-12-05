package api

import (
	"net/http"
)

// RegisterHandlers registers all API handlers.
func RegisterHandlers(mux *http.ServeMux) {
	prefix := "/api/v1"

	registerWithAuth(mux, prefix+"/auth", HandleAuth)
	registerWithAuth(mux, prefix+"/agents", HandleAgents)
	registerWithAuth(mux, prefix+"/vms", HandleVMs)
	registerWithAuth(mux, prefix+"/vms/", HandleVMs)
	registerWithAuth(mux, prefix+"/forwards", HandleForwards)
	registerWithAuth(mux, prefix+"/images", HandleImages)

	mux.HandleFunc(prefix+"/login", HandleOIDCCallback)
}

func registerWithAuth(mux *http.ServeMux, url string, f func(http.ResponseWriter, *http.Request)) {
	mux.Handle(url, AuthMiddleware(http.HandlerFunc(f)))
}
