package builtin

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/locngoxuan/vulcan/core"
)

//entry point of executor
func RunVExec() error {
	configFile := flag.String("config", "", "")
	jobId := flag.String("job-id", "", "")
	flag.Parse()

	if *configFile = strings.TrimSpace(*configFile); *configFile == "" {
		return fmt.Errorf(`path of config file is missing`)
	}

	if *jobId = strings.TrimSpace(*jobId); *jobId == "" {
		return fmt.Errorf(`job id is missing`)
	}

	err := core.CreateTmpDir()
	if err != nil {
		return fmt.Errorf(`failed to create temporary directory %v`, err)
	}

	fmt.Printf("Run job: config-file=%s id=%s\n", *configFile, *jobId)
	config, err := core.ReadProjectConfig(*configFile)
	if err != nil {
		return err
	}
	for k, c := range config.Jobs {
		if k != *jobId {
			continue
		}
		c.Id = *jobId
		if c.Args != nil {
			c.Args.ReplaceEnv()
		}
		err = runJob(c)
		if err != nil {
			return err
		}
	}
	return nil
}

func runJob(c *core.JobConfig) error {
	globalArgs := make(map[string]string)
	if c.Args != nil {
		for k, v := range *c.Args {
			globalArgs[k] = v
		}
	}

	p := os.Getenv("PATH")
	p = fmt.Sprintf(`%s:%s:%s`, p, core.ToolChainInsideContainer, core.PluginInsideContainer)
	err := os.Setenv("PATH", p)
	if err != nil {
		return err
	}
	err = os.Setenv("GOTMPDIR", "/tmp")
	if err != nil {
		return err
	}

	for _, step := range c.Steps {
		//build local arguments
		args := make(map[string]string)
		for k, v := range globalArgs {
			args[k] = v
		}
		if step.Args != nil {
			step.Args.ReplaceEnv()
			for k, v := range *step.Args {
				args[k] = v
			}
		}
		if v := strings.TrimSpace(step.Name); v != "" {
			fmt.Printf("Step: %s\n", v)
		}

		if step.Id != "" {
			err = core.SetCurrentStep(step.Id)
			if err != nil {
				return err
			}
		}

		if v := strings.TrimSpace(step.Run); v != "" {
			cmdlines := strings.Split(v, "\n")
			for _, cmdLine := range cmdlines {
				err = runCommandLine(cmdLine, args)
				if err != nil {
					return err
				}
			}
		} else if v := strings.TrimSpace(step.Use); v != "" {
			//use plugin
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
			cmdLine := b.String()
			b.Reset()
			err = runCommandLine(cmdLine, args)
			if err != nil {
				return err
			}
		} else {
			//throw error
			return fmt.Errorf(`either run or use must be specified`)
		}

		if step.Id != "" {
			//load input from /tmp/vulcan/output/step-id
			outputs, err := core.GetAllOutputs()
			if err != nil {
				return err
			}
			for _, outp := range outputs {
				globalArgs[outp.Key] = outp.Value
			}
		}
	}
	return nil
}

func runCommandLine(cmdLine string, args map[string]string) error {
	cmdLine = strings.TrimSpace(cmdLine)
	fmt.Printf("Run: %s\n", cmdLine)
	var buf bytes.Buffer
	t, err := template.New("tmpl").Parse(cmdLine)
	if err != nil {
		return err
	}
	err = t.Execute(&buf, args)
	if err != nil {
		return err
	}

	cmdArgs, err := core.ParseCommandLine(buf.String())
	if err != nil {
		return err
	}
	execFile := ""
	argStart := 0
	for i, arg := range cmdArgs {
		execFile = strings.TrimSpace(arg)
		if execFile != "" {
			argStart = i + 1
			break
		}
	}
	if argStart > len(cmdArgs) {
		argStart = len(cmdArgs)
	}
	cmd := exec.Command(execFile, cmdArgs[argStart:]...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
