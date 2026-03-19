package main

import (
	"fmt"
	"os"

	"github.com/delvop-dev/delvop/cmd/delvop"
)

var version = "dev"

func main() {
	delvop.SetVersion(version)
	if err := delvop.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
