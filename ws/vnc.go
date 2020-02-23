package ws

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/net/websocket"
	"minivmm"
)

// HandleWsVNC proxies between websocket and VNC protocols.
func HandleWsVNC(wsconn *websocket.Conn) {
	defer wsconn.Close()

	// get the destination VM name
	vmName := wsconn.Request().URL.Query().Get("name")
	vncSocketPath := filepath.Join(minivmm.VMDir, vmName, "vnc.socket")

	// connect to vnc socket
	vncconn, err := net.Dial("unix", vncSocketPath)
	if err != nil {
		log.Printf("failed to open vnc socket: %v\n", err)
		return
	}
	defer vncconn.Close()

	wsconn.PayloadType = websocket.BinaryFrame

	// proxy between websocket and VNC
	done := make(chan struct{})
	go func() {
		io.Copy(wsconn, vncconn)
		wsconn.Close()
		vncconn.Close()
		done <- struct{}{}
	}()
	go func() {
		io.Copy(vncconn, wsconn)
		wsconn.Close()
		vncconn.Close()
		done <- struct{}{}
	}()
	<-done
	<-done

	log.Printf("ws disconnected name=%s\n", vmName)
}

// HandshakeWsVNC checks parameters and authorizes the websocket connection request.
func HandshakeWsVNC(config *websocket.Config, r *http.Request) error {
	// TODO: remove debug log
	for k, v := range r.Header {
		log.Printf("ws connect header {%s: %s}\n", k, v)
	}

	vmName := r.URL.Query().Get("name")
	if vmName == "" {
		return fmt.Errorf("missing query parameter 'name'")
	}
	log.Printf("ws connect query name=%s\n", vmName)

	config.Protocol = []string{"binary"}

	envNoAuth := os.Getenv(minivmm.EnvNoAuth)
	if envNoAuth != "1" && envNoAuth != "true" {
		// TODO: Impl verification of the access token.
		msg := "failed to verify the access token"
		log.Println(msg)
		return fmt.Errorf(msg)
	}

	log.Printf("ws connected name=%s\n", vmName)
	return nil
}
