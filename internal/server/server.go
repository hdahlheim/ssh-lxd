package server

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"

	"github.com/gliderlabs/ssh"
)

var iLS = &intLXDServer{}

func Run() int {
	certFile := os.Getenv("LXD_CLIENT_CERT")
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		log.Println(certFile)
		log.Println(err)
		return 1
	}

	key, err := ioutil.ReadFile(os.Getenv("LXD_CLIENT_KEY"))
	if err != nil {
		log.Println(err)
		return 1
	}

	iLS = &intLXDServer{
		URL:           os.Getenv("LXD_HOST_URL"),
		TLSClientCert: string(cert),
		TLSClientKey:  string(key),
	}

	s := ssh.Server{
		Addr:    ":6666",
		Handler: sessionHandler,
	}

	// publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
	// 	return true // allow all keys, or use ssh.KeysEqual() to compare against known keys
	// })

	log.Println("starting server on :6666")
	if err := s.ListenAndServe(); err != nil {
		log.Println(err)
		return 1
	}
	return 0
}

func sessionHandler(s ssh.Session) {
	instance := s.User()
	log.Println()
	log.Println(iLS.connect(instance, s))

	// authorizedKey := gossh.MarshalAuthorizedKey(s.PublicKey())
	// io.WriteString(s, fmt.Sprintf("public key used by %s:\n", s.User()))
	// s.Write(authorizedKey)
}

type intLXDServer struct {
	URL           string
	TLSClientCert string
	TLSClientKey  string
}

func (iLS *intLXDServer) connect(instance string, s ssh.Session) error {
	// Connect to LXD over the HTTPS
	c, err := lxd.ConnectLXD(iLS.URL, &lxd.ConnectionArgs{
		InsecureSkipVerify: true,
		TLSClientCert:      iLS.TLSClientCert,
		TLSClientKey:       iLS.TLSClientKey,
	})
	if err != nil {
		return err
	}

	// Setup the exec request
	req := api.ContainerExecPost{
		Command:     []string{"bash"},
		WaitForWS:   true,
		Interactive: true,
		Width:       80,
		Height:      60,
		Environment: map[string]string{
			"PATH": ":/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			"TERM": "xterm",
		},
	}

	_, windowChannel, _ := s.Pty()

	// Setup the exec arguments
	args := lxd.ContainerExecArgs{
		Stdin:  s,
		Stdout: s,
		Stderr: s,
		Control: func(conn *websocket.Conn) {
			for window := range windowChannel {
				req := api.InstanceExecControl{}
				req.Command = "window-resize"
				req.Args = map[string]string{
					"width":  strconv.Itoa(window.Width),
					"height": strconv.Itoa(window.Height),
				}

				if err := conn.WriteJSON(req); err != nil {
					log.Panicln(err)
				}
			}
		},
	}

	// Get the current state
	op, err := c.ExecContainer(instance, req, &args)
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
