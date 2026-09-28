package main

import (
	"context"
	"fmt"
	"os"

	"github.com/michaelmjhhhh/typeit/internal/cli"
)

func main() {
	if err := cli.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "typeit:", err)
		os.Exit(1)
	}
}
