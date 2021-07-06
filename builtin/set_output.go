package builtin

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/locngoxuan/vulcan/core"
)

func RunVSet() error {
	var kv core.StringList
	flag.Var(&kv, "kv", "specify a key-value pair")
	stepId := flag.String("step-id", "", "specific where stores key-value pairs")
	flag.Parse()
	if *stepId = strings.TrimSpace(*stepId); *stepId == "" {
		return fmt.Errorf(`step id is missing`)
	}
	if kv == nil || len(kv) == 0 {
		return nil
	}

	p := filepath.Join(stepOutput, *stepId)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	w := bufio.NewWriter(f)
	for _, pair := range kv {
		pair = strings.TrimSpace(pair)
		elems := strings.Split(pair, "=")
		if len(elems) == 1 {
			return fmt.Errorf(`pair %s is malformed`, pair)
		}
		w.WriteString(pair)
		w.WriteString("\n")
	}
	w.Flush()
	return nil
}
