package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
)

// ImagePullProgress holds aggregated pull progress.
type ImagePullProgress struct {
	Current int64 `json:"current"`
	Total   int64 `json:"total"`
}

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
	return PullImageWithProgress(ctx, cli, ImageName, func(_, _ int64) {})
}

// PullImageWithProgress pulls the given image and calls onProgress with
// aggregated download+extract bytes.
func PullImageWithProgress(ctx context.Context, cli *client.Client, imageName string, onProgress func(current, total int64)) error {
	slog.Info("pulling docker image", "image", imageName)
	reader, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer func() { _ = reader.Close() }()

	type layer struct {
		current int64
		total   int64
	}
	layers := make(map[string]*layer)

	dec := json.NewDecoder(reader)
	for {
		type result struct {
			msg jsonmessage.JSONMessage
			err error
		}
		ch := make(chan result, 1)
		go func() {
			var msg jsonmessage.JSONMessage
			if err := dec.Decode(&msg); err != nil {
				ch <- result{err: err}
				return
			}
			ch <- result{msg: msg}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case res := <-ch:
			if res.err != nil {
				if res.err == io.EOF {
					slog.Info("docker image pulled", "image", imageName)
					return nil
				}
				return fmt.Errorf("failed to decode pull progress: %w", res.err)
			}
			msg := res.msg
			if msg.Error != nil {
				return fmt.Errorf("pull error: %s", msg.Error.Message)
			}
			if msg.ID == "" {
				continue
			}

			l := layers[msg.ID]
			if l == nil {
				l = &layer{}
				layers[msg.ID] = l
			}

			switch msg.Status {
			case "Already exists", "Pull complete", "Download complete":
				if l.total > 0 {
					l.current = l.total
				}
			default:
				if msg.Progress != nil && msg.Progress.Total > 0 {
					l.current = msg.Progress.Current
					l.total = msg.Progress.Total
				}
			}

			var totalCurrent, totalSize int64
			for _, ll := range layers {
				totalCurrent += ll.current
				totalSize += ll.total
			}
			onProgress(totalCurrent, totalSize)
		}
	}
}
