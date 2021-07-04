package main

import (
	"fmt"
	"os"

	"github.com/locngoxuan/vulcan/builtin"
)

func main() {
	err := builtin.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
