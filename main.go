package main

import (
	"fmt"
	"os"

	"github.com/delvop-dev/delvop/cmd/delvop"
)

func main() {
	if err := delvop.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
