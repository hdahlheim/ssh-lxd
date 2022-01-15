package main

import (
	"os"

	"github.com/hdahlheim/ssh-lxd/cmd"
)

func main() {
	os.Exit(cmd.Run())
}
