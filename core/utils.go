package core

import (
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
