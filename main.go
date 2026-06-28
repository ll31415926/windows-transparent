package main

import (
	"fmt"
	"os"

	"windows-transparent/frontend"
	"windows-transparent/internal/gui"
)

func main() {
	if err := gui.Run(gui.Config{Assets: frontend.Assets}); err != nil {
		fmt.Fprintf(os.Stderr, "error: start GUI: %v\n", err)
		os.Exit(1)
	}
}
