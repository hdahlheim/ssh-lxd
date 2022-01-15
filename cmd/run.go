package cmd

import (
	"log"

	"github.com/hdahlheim/ssh-lxd/internal/config"
	"github.com/hdahlheim/ssh-lxd/internal/server"
)

func Run() int {
	err := config.LoadConfig()
	if err != nil {
		log.Println(err)
		return 1
	}

	cfg := config.GetConfig()

	if err := server.Run(cfg); err != nil {
		log.Println(err)
		return 1
	}
	return 0
}
