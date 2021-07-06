package core

import (
	"fmt"
	"strings"
)

type cmdParserState int

type cmdParser func(i int, line string, args *[]string) (int, cmdParserState)

const (
	PARSER_CMD_BEGIN cmdParserState = iota
	PARSER_SPACE
	PARSER_FLAG_KEY
	PARSER_FLAG_VALUE
	PARSER_ARGS
	PARSER_ERROR
)

var states = map[cmdParserState]cmdParser{
	PARSER_CMD_BEGIN: func(i int, line string, args *[]string) (int, cmdParserState) {
		j := i
		var b strings.Builder
		for j = i; j < len(line); j++ {
			r := rune(line[j])
			if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') {
				b.WriteRune(r)
				continue
			}
			break
		}
		word := b.String()
		*args = append(*args, word)
		return j, PARSER_SPACE
	},
	PARSER_SPACE: func(i int, line string, _ *[]string) (int, cmdParserState) {
		r := rune(line[i])
		if r == ' ' {
			return i + 1, PARSER_SPACE
		}
		if r == '-' {
			return i, PARSER_FLAG_KEY
		}
		return i, PARSER_ARGS
	},
	PARSER_ARGS: func(i int, line string, args *[]string) (int, cmdParserState) {
		j := i
		var b strings.Builder
		openQuote := rune(0)
		for j = i; j < len(line); j++ {
			r := rune(line[j])
			if r == '\'' || r == '"' {
				if openQuote == rune(0) {
					openQuote = r
					j++
					continue
				}
				if rune(line[j-1]) == '\\' {
					b.WriteRune(r)
					continue
				}
				if openQuote == r {
					j += 1
					break
				}
				continue
			}
			if r == ' ' && openQuote == rune(0) {
				break
			}
			b.WriteRune(r)
		}
		word := b.String()
		*args = append(*args, word)
		return j, PARSER_SPACE
	},
	PARSER_FLAG_KEY: func(i int, line string, args *[]string) (int, cmdParserState) {
		j := i
		var b strings.Builder
		for j = i; j < len(line); j++ {
			r := rune(line[j])
			if ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == '-' {
				b.WriteRune(r)
				continue
			}
			if r != ' ' && r != '=' {
				*args = []string{fmt.Sprintf(`invalid flag key %s`, b.String())}
				return 0, PARSER_ERROR
			}
			break
		}
		word := b.String()
		*args = append(*args, word)
		return j, PARSER_FLAG_VALUE
	},
	PARSER_FLAG_VALUE: func(i int, line string, args *[]string) (int, cmdParserState) {
		if rune(line[i]) == '-' {
			return i, PARSER_FLAG_KEY
		}
		if rune(line[i]) == ' ' || rune(line[i]) == '=' {
			i = i + 1
		}
		var b strings.Builder
		j := i
		openQuote := rune(0)
		for j = i; j < len(line); j++ {
			r := rune(line[j])
			if r == '"' || r == '\'' {
				if openQuote == rune(0) {
					openQuote = r
					continue
				}
				if openQuote == r && rune(line[j-1]) != '\\' {
					j = j + 1
					break
				}
			}
			if r == ' ' && openQuote == rune(0) {
				j = j + 1
				break
			}
			b.WriteRune(r)
		}
		word := b.String()
		*args = append(*args, word)
		return j, PARSER_SPACE
	},
}

func ParseCommandLine(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	state := PARSER_CMD_BEGIN
	var args []string
	for i := 0; i < len(line); {
		if state == PARSER_ERROR {
			return nil, fmt.Errorf(args[0])
		}
		f, ok := states[state]
		if !ok {
			return nil, fmt.Errorf(`no such state`)
		}
		i, state = f(i, line, &args)
	}
	return args, nil
}
