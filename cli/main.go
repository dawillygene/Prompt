package main

import (
	"fmt"
	"os"

	"github.com/dawillygene/my-prompt-repository/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
