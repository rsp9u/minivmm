package ws

import (
	"net/http"

	"golang.org/x/net/websocket"
)

// RegisterHandlers registers all Websocket handlers.
func RegisterHandlers(mux *http.ServeMux) {
	server := websocket.Server{Handshake: HandshakeWsVNC, Handler: websocket.Handler(HandleWsVNC)}
	mux.Handle("/ws/vnc", server)
}
