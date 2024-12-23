package main

import (
	"fmt"
	"os"

	"github.com/tasiov/golulo/cmd/golulo/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
