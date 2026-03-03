package main

import (
	"fmt"
	"os"

	"github.com/tenlisboa/tempo/internal/config"
	"github.com/tenlisboa/tempo/internal/daemon"
)

func main() {
	cfg := config.Load()
	if err := daemon.Run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
