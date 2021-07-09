package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/locngoxuan/vulcan/core"
)

var workDir = "/workdir"

func runJob(configFile string, jobConfig core.JobConfig) error {
	//check and pull image if it is necessary
	log.Printf("Job: %s", jobConfig.Id)
	mounts := make([]mount.Mount, 0)
	err := filepath.Walk(pwd, func(path string, info fs.FileInfo, err error) error {
		if pwd == path {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		f := strings.TrimPrefix(path, pwd)
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: path,
			Target: filepath.Join(workDir, f),
		})
		return nil
	})

	if jobConfig.Artifacts != nil {
		for _, artifact := range jobConfig.Artifacts {
			parts := strings.Split(artifact, ":")
			host := strings.TrimSpace(parts[0])
			target := strings.TrimSpace(parts[1])
			if strings.HasPrefix(host, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				host = filepath.Join(home, host[1:])
			}

			if strings.HasPrefix(target, "/") {

			} else if strings.HasPrefix(target, "~") {
				target = filepath.Join("/root", target)
			} else {
				target = filepath.Join(workDir, target)
			}

			h, err := filepath.Abs(host)
			if err != nil {
				return err
			}
			t, err := filepath.Abs(target)
			if err != nil {
				return err
			}

			_, err = os.Stat(h)
			if err != nil {
				if os.IsNotExist(err) {
					err = os.MkdirAll(h, 0755)
					if err != nil {
						return err
					}
				}
				return err
			}

			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: h,
				Target: t,
			})
		}
	}

	dockerCommandArg := make([]string, 0)
	fis, err := ioutil.ReadDir(toolChains)
	if err != nil {
		return err
	}
	//mount all binaries from toolchains to /bin
	for _, fi := range fis {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: filepath.Join(toolChains, fi.Name()),
			Target: filepath.Join("/bin", fi.Name()),
		})
	}

	configFile = filepath.Join(workDir, ".vulcan", configFile)
	dockerCommandArg = append(dockerCommandArg, "/bin/vexec",
		"--config", configFile,
		"--job-id", jobConfig.Id)

	containerConfig := &container.Config{
		Image:        jobConfig.RunOn,
		Cmd:          dockerCommandArg,
		WorkingDir:   workDir,
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
