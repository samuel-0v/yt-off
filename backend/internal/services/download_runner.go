package services

import (
	"context"
	"fmt"
	"io"

	dockerclient "yt-off/backend/internal/docker"
)

type dockerDownloadRunner struct {
	referenceContainerName string
}

func newDockerDownloadRunner(referenceContainerName string) *dockerDownloadRunner {
	return &dockerDownloadRunner{referenceContainerName: referenceContainerName}
}

func (runner *dockerDownloadRunner) Run(ctx context.Context, downloadID string, command []string, stdoutWriter io.Writer, stderrWriter io.Writer, onContainerCreated func(string)) (dockerclient.ContainerRunResult, error) {
	cli, err := dockerclient.NewClient()
	if err != nil {
		return dockerclient.ContainerRunResult{}, err
	}
	defer cli.Close()

	return dockerclient.RunCommandContainer(
		ctx,
		cli,
		runner.referenceContainerName,
		fmt.Sprintf("yt-off-download-%s", downloadID),
		command,
		stdoutWriter,
		stderrWriter,
		onContainerCreated,
	)
}

func (runner *dockerDownloadRunner) Stop(ctx context.Context, containerID string) error {
	cli, err := dockerclient.NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	return dockerclient.StopContainer(ctx, cli, containerID)
}
