package main

import (
	"os"

	"github.com/hiasinho/specter/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
