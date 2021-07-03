package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/locngoxuan/vulcan/core"
)

var dockerCli core.DockerClient

//vulcan --config-docker {config-file} --env-file {} --env --action {action-name}
func main() {
	configDocker := flag.String("config-docker", "", "specify location of docker configuration file.")
	envFile := flag.String("env-file", "", "specify location of environment file.")
	action := flag.String("action", "", "specify action for running.")
	var envs core.EnvPairs
	flag.Var(&envs, "env", "set environment variables")
	flag.Parse()

	if *action = strings.TrimSpace(*action); *action == "" {
		flag.PrintDefaults()
		return
	}

	if *configDocker = strings.TrimSpace(*configDocker); *configDocker != "" {
		//read docker configuration
	}

	if *envFile = strings.TrimSpace(*envFile); *envFile != "" {
		//read then export environment variables
	}

	if envs != nil {
		for _, envPair := range envs {
			parts := strings.Split(strings.TrimSpace(envPair), "=")
			if len(parts) == 2 {
				err := os.Setenv(parts[0], parts[1])
				if err != nil {
					log.Fatalf("failed to set env variable: %v", err)
				}
			} else if len(parts) > 2 {
				err := os.Setenv(parts[0], strings.Join(parts[1:], "="))
				if err != nil {
					log.Fatalf("failed to set env variable: %v", err)
				}
			}
		}
	}
	var err error
	dockerCli, err = core.ConnectDockerHost(context.Background(),
		[]string{core.DefaultDockerUnixSock, core.DefaultDockerTCPSock})
	if err != nil {
		log.Fatalf("failed to connect docker host: %v", err)
	}

	pwd, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("failed to get present working directory: %v", err)
	}

	vulCanDir := filepath.Join(pwd, ".vulcan")
	st, err := os.Stat(vulCanDir)
	if err != nil {
		log.Fatalf("failed to read .vulcan directory: %v", err)
	}

	if !st.IsDir() {
		log.Fatalf("%s is not directory", vulCanDir)
	}

	fileInfos, err := ioutil.ReadDir(vulCanDir)
	if err != nil {
		log.Fatalf("failed to list file in .vulcan directory: %v", err)
	}

	for _, fileInfo := range fileInfos {
		p := filepath.Join(vulCanDir, fileInfo.Name())
		ext := filepath.Ext(p)
		fileName := strings.TrimSuffix(fileInfo.Name(), ext)
		if fileName == *action {
			c, err := core.ReadProjectConfig(p)
			if err != nil {
				log.Fatalf("failed to read project configuration file: %v", err)
			}
			//prepare environemnt before runing build
			for _, job := range c.Jobs {
				if job.Args != nil {
					job.Args.ReplaceEnv()
				}
				for _, step := range job.Steps {
					if step.With != nil {
						step.With.ReplaceEnv()
					}
				}

				err = runJob(pwd, job)
				if err != nil {
					log.Fatalf("failed to run job %s: %v", job.Name, err)
				}
			}
		}
	}
}
