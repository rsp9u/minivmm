package ws

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
	"minivmm"
)

// HandleWsVNC proxies between websocket and VNC protocols.
func HandleWsVNC(wsconn *websocket.Conn) {
	defer wsconn.Close()

	// get the destination VM name
	vmName := wsconn.Request().URL.Query().Get("name")
	vncSocketPath := filepath.Join(minivmm.C.VMDir, vmName, "vnc.socket")

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
	err := handshakeWsVNC(config, r)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func handshakeWsVNC(config *websocket.Config, r *http.Request) error {
	vmName := r.URL.Query().Get("name")
	if vmName == "" {
		return fmt.Errorf("missing query parameter 'name'")
	}
	log.Printf("ws connect query name=%s\n", vmName)

	vmMetaData, err := minivmm.GetVM(vmName)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("no such a VM named '%s'", vmName))
	}

	config.Protocol = []string{"binary"}

	if !minivmm.C.NoAuth {
		// get access token
		cookie, err := r.Cookie(minivmm.CookieName)
		if err != nil {
			return errors.Wrap(err, "failed to get the access token from cookie")
		}
		token := cookie.Value

		// verify token
		payload, err := minivmm.VerifyToken(token)
		if err != nil {
			return errors.Wrap(err, "failed to verify the access token")
		}

		// check ownership for VM
		if vmMetaData.Owner != payload.Subject {
			return fmt.Errorf("forbidden")
		}
	}

	log.Printf("ws connected name=%s\n", vmName)
	return nil
}
