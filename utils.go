package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var shellBegin = `# exit when any command fails
set -e

`

func writeShellScript(jobId string, cmd []string) (string, error) {
	p := filepath.Join(pwd, fmt.Sprintf(`%s.sh`, jobId))
	f, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return p, err
	}
	defer func() {
		_ = f.Close()
	}()
	writer := bufio.NewWriter(f)
	writer.WriteString(shellBegin)
	writer.WriteString("\n")
	for _, line := range cmd {
		writer.WriteString(line)
		writer.WriteString("\n")
		writer.WriteString("\n")
	}
	writer.Flush()
	return p, nil
}

func removeShellScript(p string) error {
	return os.RemoveAll(p)
}

func updateEnvFromFile(envFile string) error {
	f, err := os.Open(envFile)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return fmt.Errorf(`destination of env is not file`)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		err = os.Setenv(parts[0], parts[1])
		if err != nil {
			return err
		}
	}
	return nil
}
