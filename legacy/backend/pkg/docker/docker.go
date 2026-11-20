package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
)

type Client struct {
	cli     *client.Client
	log     zerolog.Logger
	timeout time.Duration
}

func NewClient(timeout time.Duration) (*Client, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := cli.Ping(ctx); err != nil {
		return nil, fmt.Errorf("docker ping: %w", err)
	}

	return &Client{
		cli:     cli,
		log:     *zerolog.Ctx(context.Background()),
		timeout: timeout,
	}, nil
}

func (c *Client) PullImage(ctx context.Context, imageName string) error {
	c.log.Info().Str("image", imageName).Msg("Pulling image")
	reader, err := c.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull image %s: %w", imageName, err)
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)
	return nil
}

func (c *Client) ImageExists(ctx context.Context, imageName string) (bool, error) {
	_, _, err := c.cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Client) CreateContainer(ctx context.Context, config *container.Config, hostConfig *container.HostConfig) (string, error) {
	resp, err := c.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (c *Client) WaitContainer(ctx context.Context, containerID string) (int64, error) {
	statusCh, errCh := c.cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return -1, err
		}
	case status := <-statusCh:
		return status.StatusCode, nil
	}
	return -1, nil
}

func (c *Client) GetContainerLogs(ctx context.Context, containerID string) (string, string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	}

	reader, err := c.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", "", err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", "", err
	}

	// Simplified - real implementation would demux stdout/stderr
	return string(data), "", nil
}

func (c *Client) RemoveContainer(ctx context.Context, containerID string) error {
	return c.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

func (c *Client) ExecInContainer(ctx context.Context, containerID string, cmd []string) error {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/workspace",
	}

	execResp, err := c.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return err
	}

	execAttach, err := c.cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return err
	}
	defer execAttach.Close()

	// Wait for completion with timeout
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	select {
	case <-waitCtx.Done():
		return waitCtx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("exec timeout")
	}
}

func (c *Client) ExecInContainerWithInput(ctx context.Context, containerID string, cmd []string, input string) error {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		WorkingDir:   "/workspace",
	}

	execResp, err := c.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return err
	}

	execAttach, err := c.cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return err
	}
	defer execAttach.Close()

	// Write input
	if _, err := execAttach.Conn.Write([]byte(input)); err != nil {
		return err
	}
	execAttach.CloseWrite()

	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	select {
	case <-waitCtx.Done():
		return waitCtx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("exec timeout")
	}
}

func (c *Client) CopyToContainer(ctx context.Context, containerID, dstPath, content string) error {
	// Use exec with tee to copy file
	cmd := []string{"sh", "-c", fmt.Sprintf("cat > %s", dstPath)}
	return c.ExecInContainerWithInput(ctx, containerID, cmd, content)
}

func (c *Client) GetContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	stats, err := c.cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	defer stats.Body.Close()

	var statsData struct {
		MemoryStats struct {
			Usage int64 `json:"usage"`
			Limit int64 `json:"limit"`
		} `json:"memory_stats"`
		CPUStats struct {
			CPUUsage struct {
				TotalUsage  uint64   `json:"total_usage"`
				PercpuUsage []uint64 `json:"percpu_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage  uint64   `json:"total_usage"`
				PercpuUsage []uint64 `json:"percpu_usage"`
			} `json:"cpu_usage"`
			SystemCPUUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
	}

	if err := json.NewDecoder(stats.Body).Decode(&statsData); err != nil {
		return nil, err
	}

	return &ContainerStats{
		MemoryUsage: statsData.MemoryStats.Usage,
		MemoryLimit: statsData.MemoryStats.Limit,
		CPUUsage:    statsData.CPUStats.CPUUsage.TotalUsage,
	}, nil
}

type ContainerStats struct {
	MemoryUsage int64
	MemoryLimit int64
	CPUUsage    uint64
}

func (c *Client) Close() error {
	return c.cli.Close()
}

// ListContainers lists containers with optional filters
func (c *Client) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	return c.cli.ContainerList(ctx, container.ListOptions{All: all})
}

// Prune removes unused containers, networks, images
func (c *Client) Prune(ctx context.Context) error {
	// Container prune
	_, err := c.cli.ContainersPrune(ctx, filters.NewArgs())
	if err != nil {
		return err
	}
	// Image prune
	_, err = c.cli.ImagesPrune(ctx, filters.NewArgs())
	return err
}
