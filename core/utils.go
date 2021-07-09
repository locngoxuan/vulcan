package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadEnvVariableIfHas(str string) string {
	origin := strings.TrimSpace(str)
	if strings.HasPrefix(origin, "$") {
		result := os.ExpandEnv(origin)
		if result != "" {
			return result
		}
	}
	return origin
}

func UpdateEnvFromFile(envFile string) ([]string, error) {
	f, err := os.Open(envFile)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf(`destination of env is not file`)
	}

	scanner := bufio.NewScanner(f)
	var envs []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) < 2 {
			continue
		}
		envs = append(envs, line)
		err = os.Setenv(parts[0], parts[1])
		if err != nil {
			return nil, err
		}
	}
	return envs, nil
}
