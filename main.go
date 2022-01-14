package main

import (
	"os"

	"github.com/hdahlheim/ssh-lxd/internal/server"
)

func main() {
	os.Exit(server.Run())
}
