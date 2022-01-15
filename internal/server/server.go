package server

import (
	"log"
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/hdahlheim/ssh-lxd/internal/config"
)

var cfg *config.Config

func Run(c *config.Config) error {
	cfg = c

	if err := initLXDClient(); err != nil {
		return err
	}

	s := ssh.Server{
		Addr:             ":6666",
		Handler:          sessionHandler,
		PublicKeyHandler: authHandler,
		PasswordHandler:  nil,
	}

	// use hostkey file if set
	if path := os.Getenv("HOST_KEY_FILE"); path != "" {
		s.SetOption(ssh.HostKeyFile(path))
	}

	log.Println("Starting server on", s.Addr)

	return s.ListenAndServe()
}

func sessionHandler(s ssh.Session) {
	instance := s.User()

	if err := connectToShell(instance, s); err != nil {
		log.Println(err)
	}
}

func authHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	user := ctx.User()

	var passed bool
	for _, keyStr := range cfg.Auth[user].Keys {
		authKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyStr))
		if err != nil {
			log.Println(err)
			continue
		}

		if passed = ssh.KeysEqual(key, authKey); passed {
			break
		}
	}
	return passed
}
