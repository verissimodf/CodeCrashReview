package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	"github.com/rs/zerolog"

	"codecrashreview/backend/internal/config"
)

type SandboxManager struct {
	dockerClient *client.Client
	config       *config.Config
	logger       zerolog.Logger
	images       map[string]string // language -> image name
}

type SandboxConfig struct {
	Language        string
	Image           string
	CPULimit        float64
	MemoryLimit     string
	PIDsLimit       int64
	NetworkDisabled bool
	ReadonlyRootfs  bool
	SeccompProfile  string
	TimeoutSeconds  int
	EnvVars         map[string]string
	WorkingDir      string
	Mounts          []mount.Mount
}

type ExecutionRequest struct {
	Files          []CodeFile
	Command        string
	Args           []string
	Stdin          string
	TimeoutSeconds int
	ResourceLimits ResourceLimits
}

type ResourceLimits struct {
	CPUSeconds  int
	MemoryBytes int64
	OutputBytes int64
}

type CodeFile struct {
	Path     string
	Language string
	Content  string
	Encoding string
}

type ExecutionResult struct {
	ExitCode   int
	Stdout     string
	Stderr     string
	DurationMs int64
	MemoryUsed int64
	TimedOut   bool
	OOMKilled  bool
}

var languageImages = map[string]string{
	"cpp":        "codecrash-sandbox-cpp:latest",
	"python":     "codecrash-sandbox-python:latest",
	"java":       "codecrash-sandbox-java:latest",
	"javascript": "codecrash-sandbox-javascript:latest",
	"typescript": "codecrash-sandbox-typescript:latest",
	"rust":       "codecrash-sandbox-rust:latest",
	"shell":      "codecrash-sandbox-shell:latest",
	"html":       "codecrash-sandbox-html:latest",
	"gherkin":    "codecrash-sandbox-gherkin:latest",
}

var languageCommands = map[string]struct {
	Compile   string
	Run       string
	Extension string
}{
	"cpp":        {Compile: "g++ -std=c++17 -O2 -pipe -static -s -o /tmp/main %s", Run: "/tmp/main", Extension: ".cpp"},
	"python":     {Compile: "", Run: "python3 %s", Extension: ".py"},
	"java":       {Compile: "javac %s", Run: "java -cp /tmp Main", Extension: ".java"},
	"javascript": {Compile: "", Run: "node %s", Extension: ".js"},
	"typescript": {Compile: "tsc %s", Run: "node %s", Extension: ".ts"},
	"rust":       {Compile: "rustc -O %s -o /tmp/main", Run: "/tmp/main", Extension: ".rs"},
	"shell":      {Compile: "", Run: "bash %s", Extension: ".sh"},
	"html":       {Compile: "", Run: "cat %s", Extension: ".html"},
	"gherkin":    {Compile: "", Run: "behave %s", Extension: ".feature"},
}

func NewSandboxManager(cfg *config.Config) (*SandboxManager, error) {
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

	return &SandboxManager{
		dockerClient: cli,
		config:       cfg,
		logger:       *zerolog.Ctx(context.Background()),
		images:       languageImages,
	}, nil
}

func (s *SandboxManager) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
	if len(req.Files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	// Determine language from first file
	lang := req.Files[0].Language
	langConfig, ok := languageCommands[lang]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	imageName := s.images[lang]
	if imageName == "" {
		return nil, fmt.Errorf("no image for language: %s", lang)
	}

	// Create container
	containerID, err := s.createContainer(ctx, imageName, req.Files, langConfig)
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	defer func() {
		// Cleanup container
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.dockerClient.ContainerRemove(cleanupCtx, containerID, container.RemoveOptions{Force: true})
	}()

	// Start container
	if err := s.dockerClient.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	// Wait for completion or timeout
	startTime := time.Now()
	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	statusCh, errCh := s.dockerClient.ContainerWait(waitCtx, containerID, container.WaitConditionNotRunning)

	var exitCode int64
	select {
	case err := <-errCh:
		if err != nil {
			return nil, fmt.Errorf("container wait: %w", err)
		}
	case status := <-statusCh:
		exitCode = status.StatusCode
	}

	duration := time.Since(startTime)

	// Get logs
	stdout, stderr, err := s.getContainerLogs(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("get logs: %w", err)
	}

	// Get stats
	memUsed, _ := s.getContainerMemoryUsage(ctx, containerID)

	// Check if timed out or OOM killed
	timedOut := duration >= timeout
	oomKilled := exitCode == 137 // SIGKILL

	return &ExecutionResult{
		ExitCode:   int(exitCode),
		Stdout:     stdout,
		Stderr:     stderr,
		DurationMs: duration.Milliseconds(),
		MemoryUsed: memUsed,
		TimedOut:   timedOut,
		OOMKilled:  oomKilled,
	}, nil
}

func (s *SandboxManager) createContainer(ctx context.Context, imageName string, files []CodeFile, langConfig struct {
	Compile   string
	Run       string
	Extension string
}) (string, error) {
	// Prepare file content
	var fileContent strings.Builder
	for _, f := range files {
		fileContent.WriteString(f.Content)
		fileContent.WriteString("\n")
	}

	// Create container config
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode(s.config.SandboxNetwork),
		Resources: container.Resources{
			CPUQuota:  int64(s.config.SandboxCPULimit * 100000),
			CPUPeriod: 100000,
			Memory:    parseMemoryLimit(s.config.SandboxMemoryLimit),
			PidsLimit: &s.config.SandboxPIDsLimit,
		},
		SecurityOpt: []string{
			"no-new-privileges:true",
		},
		ReadonlyRootfs: true,
		Tmpfs: map[string]string{
			"/tmp": "rw,noexec,nosuid,size=100m",
		},
	}

	if s.config.SandboxSeccompProfile != "" {
		hostConfig.SecurityOpt = append(hostConfig.SecurityOpt, "seccomp="+s.config.SandboxSeccompProfile)
	}

	containerConfig := &container.Config{
		Image:        imageName,
		Cmd:          []string{"sleep", "infinity"}, // Keep container running
		WorkingDir:   "/workspace",
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		OpenStdin:    true,
		StdinOnce:    false,
		Tty:          false,
		Env:          []string{"LANG=C.UTF-8"},
	}

	resp, err := s.dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}

	containerID := resp.ID

	// Copy files to container
	for _, f := range files {
		dstPath := "/workspace/" + f.Path
		if err := s.copyToContainer(ctx, containerID, dstPath, f.Content); err != nil {
			return "", fmt.Errorf("copy file %s: %w", f.Path, err)
		}
	}

	// Compile if needed
	if langConfig.Compile != "" {
		compileCmd := fmt.Sprintf(langConfig.Compile, "/workspace/main"+langConfig.Extension)
		if err := s.execInContainer(ctx, containerID, []string{"sh", "-c", compileCmd}); err != nil {
			return "", fmt.Errorf("compile failed: %w", err)
		}
	}

	return containerID, nil
}

func (s *SandboxManager) copyToContainer(ctx context.Context, containerID, dstPath, content string) error {
	// Create tar archive
	// Simplified: use docker exec with tee
	cmd := []string{"sh", "-c", fmt.Sprintf("cat > %s", dstPath)}
	return s.execInContainerWithInput(ctx, containerID, cmd, content)
}

func (s *SandboxManager) execInContainer(ctx context.Context, containerID string, cmd []string) error {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		WorkingDir:   "/workspace",
	}

	execResp, err := s.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return err
	}

	execAttach, err := s.dockerClient.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return err
	}
	defer execAttach.Close()

	// Wait for completion
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("exec timeout")
	}
}

func (s *SandboxManager) execInContainerWithInput(ctx context.Context, containerID string, cmd []string, input string) error {
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		WorkingDir:   "/workspace",
	}

	execResp, err := s.dockerClient.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return err
	}

	execAttach, err := s.dockerClient.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return err
	}
	defer execAttach.Close()

	// Write input
	if _, err := execAttach.Conn.Write([]byte(input)); err != nil {
		return err
	}
	execAttach.CloseWrite()

	// Wait for completion
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("exec timeout")
	}
}

func (s *SandboxManager) getContainerLogs(ctx context.Context, containerID string) (stdout, stderr string, err error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	}

	reader, err := s.dockerClient.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", "", err
	}
	defer reader.Close()

	// Docker logs are multiplexed, need to demux
	// Simplified: read all
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", "", err
	}

	// For simplicity, return all as stdout
	// Real implementation would parse the multiplexed stream
	return string(data), "", nil
}

func (s *SandboxManager) getContainerMemoryUsage(ctx context.Context, containerID string) (int64, error) {
	stats, err := s.dockerClient.ContainerStats(ctx, containerID, false)
	if err != nil {
		return 0, err
	}
	defer stats.Body.Close()

	var statsData struct {
		MemoryStats struct {
			Usage int64 `json:"usage"`
			Limit int64 `json:"limit"`
		} `json:"memory_stats"`
	}

	if err := json.NewDecoder(stats.Body).Decode(&statsData); err != nil {
		return 0, err
	}

	return statsData.MemoryStats.Usage, nil
}

func (s *SandboxManager) PrewarmContainers(ctx context.Context) error {
	for _, imageName := range s.images {
		// Pull image if not present
		_, _, err := s.dockerClient.ImageInspectWithRaw(ctx, imageName)
		if err != nil {
			s.logger.Info().Str("image", imageName).Msg("Pulling sandbox image")
			reader, err := s.dockerClient.ImagePull(ctx, imageName, image.PullOptions{})
			if err != nil {
				return fmt.Errorf("pull image %s: %w", imageName, err)
			}
			io.Copy(io.Discard, reader)
			reader.Close()
		}

		// Create and start a warm container
		// This would be done in a pool for production
	}

	return nil
}

func parseMemoryLimit(limit string) int64 {
	limit = strings.TrimSpace(limit)
	if strings.HasSuffix(limit, "g") || strings.HasSuffix(limit, "G") {
		var val float64
		fmt.Sscanf(limit, "%f", &val)
		return int64(val * 1024 * 1024 * 1024)
	}
	if strings.HasSuffix(limit, "m") || strings.HasSuffix(limit, "M") {
		var val float64
		fmt.Sscanf(limit, "%f", &val)
		return int64(val * 1024 * 1024)
	}
	if strings.HasSuffix(limit, "k") || strings.HasSuffix(limit, "K") {
		var val float64
		fmt.Sscanf(limit, "%f", &val)
		return int64(val * 1024)
	}
	var val int64
	fmt.Sscanf(limit, "%d", &val)
	return val
}

func (s *SandboxManager) Close() error {
	return s.dockerClient.Close()
}
