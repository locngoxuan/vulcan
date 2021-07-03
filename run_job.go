package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/locngoxuan/vulcan/core"
)

func runJob(pwd string, jobConfig core.JobConfig) error {
	//aggregate run command
	var shellCmds []string
	globalArgs := make(map[string]string)

	if jobConfig.Args != nil {
		for k, v := range *jobConfig.Args {
			globalArgs[k] = v
		}
	}
	for i, step := range jobConfig.Steps {
		//build local arguments
		args := make(map[string]string)
		for k, v := range globalArgs {
			args[k] = v
		}
		if step.With != nil {
			for k, v := range *step.With {
				args[k] = v
			}
		}

		//creaet command
		var run string
		if v := strings.TrimSpace(step.Run); v != "" {
			run = v
		} else if v := strings.TrimSpace(step.Use); v != "" {
			var b strings.Builder
			b.WriteString(v)
			b.WriteString(" ")
			if step.With != nil {
				for k, v := range *step.With {
					b.WriteString(fmt.Sprintf(`--%s`, k))
					b.WriteString("=")
					b.WriteString(v)
					b.WriteString(" ")
				}
			}
			run = b.String()
		} else {
			run = ""
		}
		if run == "" {
			return fmt.Errorf(`step is not specified for running`)
		}
		var buf bytes.Buffer
		t, err := template.New(fmt.Sprintf(`tmpl_%02d`, i)).Parse(run)
		if err != nil {
			return err
		}
		err = t.Execute(&buf, args)
		if err != nil {
			return err
		}
		shellCmds = append(shellCmds, buf.String())
	}
	//write shell script
	writeShellScript(pwd, shellCmds)
	return nil
}

var shellBegin = `# exit when any command fails
set -e

# keep track of the last executed command
trap 'last_command=$current_command; current_command=$BASH_COMMAND' DEBUG
# echo an error message before exiting
trap 'echo "\"${last_command}\" command filed with exit code $?."' EXIT
`

func writeShellScript(pwd string, cmd []string) error {
	p := filepath.Join(pwd, "vulcan.sh")
	f, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
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
	return nil
}
