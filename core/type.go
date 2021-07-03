package core

import (
	"fmt"
)

type EnvPairs []string

// Implement the flag.Value interface
func (s *EnvPairs) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *EnvPairs) Set(value string) error {
	*s = append(*s, value)
	return nil
}
