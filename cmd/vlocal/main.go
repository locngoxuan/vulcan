package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/locngoxuan/vulcan/core"
)

var dockerCli core.DockerClient
var pwd string
var verbose bool
var toolChains string
var plugins string

func main() {
	configDocker := flag.String("config-docker", "", "specify location of docker configuration file.")
	envFile := flag.String("env-file", "", "specify location of environment file.")
	action := flag.String("action", "", "specify action for running.")
	jobId := flag.String("job", "", "specify job for running.")
	flag.StringVar(&toolChains, "toolchain", "", "specify location of toolchains directory.")
	flag.StringVar(&plugins, "plugin", "", "specify location of plugins directory.")
	flag.BoolVar(&verbose, "verbose", false, "print detail of build.")
	var envs core.StringList
	flag.Var(&envs, "env", "set environment variables")
	flag.Parse()

	if *action = strings.TrimSpace(*action); *action == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of: vulcan\n")
		flag.PrintDefaults()
		return
	}

	*jobId = strings.TrimSpace(*jobId)

	if toolChains = strings.TrimSpace(toolChains); toolChains == "" {
		vucalHome := strings.TrimSpace(os.Getenv("VULCAN_HOME"))
		if vucalHome == "" {
			p, err := exec.LookPath("vlocal")
			if err != nil {
				log.Fatalln(err)
			}
			//return bin
			vucalHome = filepath.Dir(filepath.Dir(p))
		}
		toolChains = filepath.Join(vucalHome, "toolchains")
	}

	if plugins = strings.TrimSpace(toolChains); plugins == "" {
		vucalHome := strings.TrimSpace(os.Getenv("VULCAN_HOME"))
		if vucalHome == "" {
			p, err := exec.LookPath("vlocal")
			if err != nil {
				log.Fatalln(err)
			}
			vucalHome = filepath.Dir(filepath.Dir(p))
		}
		plugins = filepath.Join(vucalHome, "plugins")
	}

	if *configDocker = strings.TrimSpace(*configDocker); *configDocker != "" {
		//read docker configuration
	}

	if *envFile = strings.TrimSpace(*envFile); *envFile != "" {
		//read then export environment variables
		err := core.UpdateEnvFromFile(*envFile)
		if err != nil {
			log.Fatalf("failed to update env variables: %v", err)
		}
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

	pwd, err = filepath.Abs(".")
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
		//exclude not yaml file
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		fileName := strings.TrimSuffix(fileInfo.Name(), ext)
		if fileName == *action {
			c, err := core.ReadProjectConfig(p)
			if err != nil {
				log.Fatalf("failed to read project configuration file: %v", err)
			}
			for id, job := range c.Jobs {
				job.Id = strings.TrimSpace(id)
				//run job
				if *jobId == "" || *jobId == job.Id {
					err = runJob(fileInfo.Name(), id, job.RunOn)
					if err != nil {
						log.Printf("failed to run job: %s", job.Name)
						log.Println("=== BEGIN: Error Message ===")
						log.Printf("%v", err)
						log.Println("=== END: Error Message ===")
						os.Exit(1)
					}
				}
			}
		}
	}
}
