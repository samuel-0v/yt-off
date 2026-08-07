package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func ListContainers(ctx context.Context, cli *client.Client) ([]container.Summary, error) {
	return cli.ContainerList(ctx, container.ListOptions{All: true})
}

func FindContainerByName(ctx context.Context, cli *client.Client, name string) (container.Summary, bool, error) {
	containers, err := ListContainers(ctx, cli)
	if err != nil {
		return container.Summary{}, false, err
	}

	for _, item := range containers {
		if hasContainerName(item.Names, name) {
			return item, true, nil
		}
	}

	return container.Summary{}, false, nil
}

func ContainerExists(ctx context.Context, cli *client.Client, name string) (bool, error) {
	_, found, err := FindContainerByName(ctx, cli, name)
	return found, err
}

func ContainerRunning(ctx context.Context, cli *client.Client, name string) (bool, error) {
	item, found, err := FindContainerByName(ctx, cli, name)
	if err != nil || !found {
		return false, err
	}

	return item.State == "running", nil
}

type ContainerRunResult struct {
	ContainerID string
	Stdout      string
	Stderr      string
	ExitCode    int
}

func RunCommandContainer(ctx context.Context, cli *client.Client, referenceContainerName string, containerName string, command []string, stdoutWriter io.Writer, stderrWriter io.Writer, onContainerCreated func(string)) (ContainerRunResult, error) {
	if len(command) == 0 {
		return ContainerRunResult{}, fmt.Errorf("command cannot be empty")
	}

	reference, found, err := FindContainerByName(ctx, cli, referenceContainerName)
	if err != nil {
		return ContainerRunResult{}, err
	}
	if !found {
		return ContainerRunResult{}, fmt.Errorf("container %q not found", referenceContainerName)
	}

	inspected, err := cli.ContainerInspect(ctx, reference.ID)
	if err != nil {
		return ContainerRunResult{}, err
	}

	created, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:        inspected.Config.Image,
			Cmd:          command,
			WorkingDir:   inspected.Config.WorkingDir,
			AttachStdout: true,
			AttachStderr: true,
			Labels: map[string]string{
				"app": "yt-off",
			},
		},
		&container.HostConfig{
			Mounts:      sharedMounts(inspected.Mounts),
			NetworkMode: firstNetworkMode(inspected),
		},
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return ContainerRunResult{}, err
	}

	containerID := created.ID
	if onContainerCreated != nil {
		onContainerCreated(containerID)
	}
	defer RemoveContainer(context.Background(), cli, containerID, true)

	attached, err := cli.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return ContainerRunResult{ContainerID: containerID}, err
	}
	defer attached.Close()

	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return ContainerRunResult{ContainerID: containerID}, err
	}

	stopWatcherDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = StopContainer(context.Background(), cli, containerID)
		case <-stopWatcherDone:
		}
	}()
	defer close(stopWatcherDone)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	copyDone := make(chan error, 1)
	go func() {
		_, err := stdcopy.StdCopy(outputWriter(&stdout, stdoutWriter), outputWriter(&stderr, stderrWriter), attached.Reader)
		copyDone <- err
	}()

	waitCh, errCh := cli.ContainerWait(context.Background(), containerID, container.WaitConditionNotRunning)

	var exitCode int
	select {
	case err := <-errCh:
		if err != nil {
			return ContainerRunResult{ContainerID: containerID, Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}, err
		}
	case response := <-waitCh:
		exitCode = int(response.StatusCode)
		if response.Error != nil {
			return ContainerRunResult{ContainerID: containerID, Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}, fmt.Errorf("%s", response.Error.Message)
		}
	}

	attached.Close()
	if err := <-copyDone; err != nil && ctx.Err() == nil {
		return ContainerRunResult{ContainerID: containerID, Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}, err
	}
	if ctx.Err() != nil {
		return ContainerRunResult{ContainerID: containerID, Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}, ctx.Err()
	}

	return ContainerRunResult{
		ContainerID: containerID,
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
		ExitCode:    exitCode,
	}, nil
}

func StopContainer(ctx context.Context, cli *client.Client, containerID string) error {
	timeout := 10
	stopCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	return cli.ContainerStop(stopCtx, containerID, container.StopOptions{Timeout: &timeout})
}

func RemoveContainer(ctx context.Context, cli *client.Client, containerID string, force bool) error {
	return cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: force})
}

func sharedMounts(mountPoints []container.MountPoint) []mount.Mount {
	mounts := make([]mount.Mount, 0, 2)
	for _, item := range mountPoints {
		if item.Destination != "/downloads" && item.Destination != "/cookies" {
			continue
		}

		source := item.Source
		if item.Name != "" {
			source = item.Name
		}
		if source == "" {
			continue
		}

		mounts = append(mounts, mount.Mount{
			Type:     mount.Type(item.Type),
			Source:   source,
			Target:   item.Destination,
			ReadOnly: item.RW == false,
		})
	}

	return mounts
}

func firstNetworkMode(inspected container.InspectResponse) container.NetworkMode {
	for networkName := range inspected.NetworkSettings.Networks {
		return container.NetworkMode(networkName)
	}

	return ""
}

func hasContainerName(names []string, expected string) bool {
	expected = strings.TrimPrefix(expected, "/")

	for _, name := range names {
		if strings.TrimPrefix(name, "/") == expected {
			return true
		}
	}

	return false
}
