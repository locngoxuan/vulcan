package core

import (
	"fmt"
)

type StringList []string

// Implement the flag.Value interface
func (s *StringList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *StringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}
