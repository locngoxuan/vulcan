package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/locngoxuan/vulcan/core"
)

func runJob(jobConfig core.JobConfig) error {
	//aggregate run command
	var shellContents []string
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
		shellContents = append(shellContents,
			fmt.Sprintf(`echo 'Run step: %s'`, step.Name),
			buf.String())
	}
	//write shell script
	shFile, err := writeShellScript(jobConfig.Id, shellContents)
	if err != nil {
		return err
	}
	//remove shell script after done
	defer func(p string) {
		_ = removeShellScript(p)
	}(shFile)
	//run build inside docker
	return runBuildInDocker(jobConfig.Id, jobConfig.RunOn)
}

func runBuildInDocker(jobId, image string) error {
	//check and pull image if it is necessary
	mounts := make([]mount.Mount, 0)
	fileInfos, err := ioutil.ReadDir(pwd)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: filepath.Join(pwd, fileInfo.Name()),
			Target: filepath.Join("/workdir", fileInfo.Name()),
		})
	}

	dockerCommandArg := make([]string, 0)
	shFile := filepath.Join("/workdir", fmt.Sprintf(`%s.sh`, jobId))
	dockerCommandArg = append(dockerCommandArg, "/bin/sh", "-c", shFile)
	containerConfig := &container.Config{
		Image:        image,
		Cmd:          dockerCommandArg,
		WorkingDir:   "/workdir",
		Tty:          true,
		AttachStdout: true,
	}
	hostConfig := &container.HostConfig{
		Mounts: mounts,
	}

	cli := dockerCli.Client
	cont, err := cli.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}

	defer core.RemoveAfterDone(cli, cont.ID)

	err = cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	if verbose {
		out, err := cli.ContainerLogs(context.Background(), cont.ID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Timestamps: false,
			Follow:     true,
			Tail:       "20",
		})
		if err != nil {
			return err
		}
		core.StreamDockerLog(out, func(s string) {
			log.Println(s)
		})
	} else {
		statusCh, errCh := cli.ContainerWait(context.Background(), cont.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				duration := 30 * time.Second
				_ = cli.ContainerStop(context.Background(), cont.ID, &duration)
				return err
			}
		case status := <-statusCh:
			//due to status code just takes either running (0) or exited (1) and I can not find a constants or variable
			//in docker sdk that represents for both two state. Then I hard-code value 1 here
			if status.StatusCode != 0 {
				var buf bytes.Buffer
				defer buf.Reset()
				out, err := cli.ContainerLogs(context.Background(), cont.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
				if err != nil {
					return err
				}
				_, err = stdcopy.StdCopy(&buf, &buf, out)
				if err != nil {
					return err
				}
				return fmt.Errorf(buf.String())
			}
		}
	}
	return nil
}
