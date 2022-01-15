package server

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/gorilla/websocket"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
)

var lxdClient lxd.InstanceServer

func initLXDClient() error {
	url := os.Getenv("LXD_HOST_URL")

	certFile := os.Getenv("LXD_CLIENT_CERT")
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return err
	}

	key, err := ioutil.ReadFile(os.Getenv("LXD_CLIENT_KEY"))
	if err != nil {
		return err
	}

	// Connect to LXD over the HTTPS
	lxdClient, err = lxd.ConnectLXD(url, &lxd.ConnectionArgs{
		InsecureSkipVerify: true,
		TLSClientCert:      string(cert),
		TLSClientKey:       string(key),
	})
	if err != nil {
		return err
	}

	return nil
}

func connectToShell(instance string, s ssh.Session) error {
	env := make(map[string]string)

	for _, l := range s.Environ() {
		kv := strings.Split(l, "=")
		env[kv[0]] = kv[1]
	}

	env["PATH"] = ":/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	env["TERM"] = "xterm"

	// if on comand is provided start a bash shell
	cmd := s.Command()
	if len(cmd) == 0 {
		cmd = []string{"bash"}
	}

	pty, windowChanges, isPty := s.Pty()

	if pty.Term != "" {
		env["TERM"] = pty.Term
	}

	// Setup the exec request
	req := api.InstanceExecPost{
		Command:     cmd,
		WaitForWS:   true,
		Interactive: isPty,
		Width:       pty.Window.Width,
		Height:      pty.Window.Height,
		Environment: env,
	}

	// Setup the exec arguments
	args := lxd.InstanceExecArgs{
		Stdin:  s,
		Stdout: s,
		Stderr: s,
		Control: func(conn *websocket.Conn) {
			for window := range windowChanges {
				req := api.InstanceExecControl{}
				req.Command = "window-resize"
				req.Args = map[string]string{
					"width":  strconv.Itoa(window.Width),
					"height": strconv.Itoa(window.Height),
				}

				if err := conn.WriteJSON(req); err != nil {
					log.Println(err)
				}
			}
		},
	}

	// Get the current state
	op, err := lxdClient.ExecInstance(instance, req, &args)
	if err != nil {
		return err
	}

	// Wait for it to complete
	err = op.Wait()
	if err != nil {
		return err
	}

	return nil
}
