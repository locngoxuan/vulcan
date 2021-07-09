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

func runJob(configFile string, jobConfig core.JobConfig, envs []string) error {
	//check and pull image if it is necessary
	log.Printf("Job: %s", jobConfig.Id)
	mounts := make([]mount.Mount, 0)
	baseDir := filepath.Join(pwd, jobConfig.BaseDir)
	vulcanConfig := filepath.Join(pwd, ".vulcan")
	set := make(map[string]struct{})
	err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if strings.HasPrefix(path, vulcanConfig) {
			return nil
		}
		if baseDir == path {
			return nil
		}
		if !info.IsDir() {
			if filepath.Dir(path) != baseDir {
				return nil
			}
		}
		for p := range set {
			if strings.HasPrefix(path, p) {
				return nil
			}
		}
		f := strings.TrimPrefix(path, baseDir)
		set[path] = struct{}{}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: path,
			Target: filepath.Join(workDir, f),
		})
		return nil
	})

	if _, ok := set[vulcanConfig]; !ok {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: vulcanConfig,
			Target: filepath.Join(workDir, ".vulcan"),
		})
	}

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

			if _, ok := set[h]; !ok {
				mounts = append(mounts, mount.Mount{
					Type:   mount.TypeBind,
					Source: h,
					Target: t,
				})
			}

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
		Env:          envs,
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
