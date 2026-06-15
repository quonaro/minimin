package runner

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// GetClient creates a new Docker client from environment settings.
func GetClient() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}
	return cli, nil
}

// ImageExists checks whether the managed Minecraft server image is already present locally.
func ImageExists(ctx context.Context, cli *client.Client) (bool, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list docker images: %w", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == ImageName {
				return true, nil
			}
		}
	}
	return false, nil
}

// PullImageIfNeeded pulls the managed image if it is not already available locally.
func PullImageIfNeeded(ctx context.Context, cli *client.Client) error {
	exists, err := ImageExists(ctx, cli)
	if err != nil {
		return err
	}

	if !exists {
		slog.Info("pulling docker image", "image", ImageName)
		reader, err := cli.ImagePull(ctx, ImageName, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
		defer func() { _ = reader.Close() }()
		_, _ = io.Copy(io.Discard, reader)
		slog.Info("docker image pulled", "image", ImageName)
	}
	return nil
}
