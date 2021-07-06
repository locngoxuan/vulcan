package main

import (
	"fmt"
	"os"

	"github.com/locngoxuan/vulcan/builtin"
)

func main() {
	err := builtin.RunVSet()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
