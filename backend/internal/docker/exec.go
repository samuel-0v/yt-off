package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func ExecCommand(ctx context.Context, cli *client.Client, containerName string, command []string) (ExecResult, error) {
	return ExecCommandWithWriters(ctx, cli, containerName, command, nil, nil)
}

func ExecCommandWithWriters(ctx context.Context, cli *client.Client, containerName string, command []string, stdoutWriter io.Writer, stderrWriter io.Writer) (ExecResult, error) {
	if len(command) == 0 {
		return ExecResult{}, fmt.Errorf("command cannot be empty")
	}

	item, found, err := FindContainerByName(ctx, cli, containerName)
	if err != nil {
		return ExecResult{}, err
	}
	if !found {
		return ExecResult{}, fmt.Errorf("container %q not found", containerName)
	}
	if item.State != "running" {
		return ExecResult{}, fmt.Errorf("container %q is not running", containerName)
	}

	created, err := cli.ContainerExecCreate(ctx, item.ID, container.ExecOptions{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return ExecResult{}, err
	}

	attached, err := cli.ContainerExecAttach(ctx, created.ID, container.ExecAttachOptions{})
	if err != nil {
		return ExecResult{}, err
	}
	defer attached.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(outputWriter(&stdout, stdoutWriter), outputWriter(&stderr, stderrWriter), attached.Reader); err != nil {
		return ExecResult{}, err
	}

	inspect, err := cli.ContainerExecInspect(ctx, created.ID)
	if err != nil {
		return ExecResult{}, err
	}

	return ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: inspect.ExitCode,
	}, nil
}

func outputWriter(capture *bytes.Buffer, stream io.Writer) io.Writer {
	if stream == nil {
		return capture
	}

	return io.MultiWriter(capture, stream)
}
