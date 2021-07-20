package builtin

import (
	"flag"
	"fmt"
	"strings"

	"github.com/locngoxuan/vulcan/core"
)

func RunVSet() error {
	var kv core.StringList
	flag.Var(&kv, "kv", "specify a key-value pair")
	flag.Parse()
	if kv == nil || len(kv) == 0 {
		return nil
	}

	for _, pair := range kv {
		pair = strings.TrimSpace(pair)
		elems := strings.Split(pair, "=")
		if len(elems) == 1 {
			return fmt.Errorf(`pair %s is malformed`, pair)
		}
		//w.WriteString(pair)
		//w.WriteString("\n")
		core.SetOutput(elems[0], strings.Join(elems[1:], "="))
	}
	return nil
}
