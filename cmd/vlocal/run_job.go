package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/locngoxuan/vulcan/core"
)

func runJob(configFile, jobId, runOn string) error {
	//check and pull image if it is necessary
	log.Printf("Job: %s", jobId)
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
	vexecLocated := filepath.Join(toolChains, "vexec")
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: vexecLocated,
		Target: filepath.Join("/bin", "vexec"),
	})

	configFile = filepath.Join("/workdir", ".vulcan", configFile)
	dockerCommandArg = append(dockerCommandArg, "/bin/vexec",
		"--config", configFile,
		"--job-id", jobId)

	containerConfig := &container.Config{
		Image:        runOn,
		Cmd:          dockerCommandArg,
		WorkingDir:   "/workdir",
		Tty:          verbose,
		AttachStdout: verbose,
		Env:          os.Environ(),
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
			if status.StatusCode == 1 {
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
