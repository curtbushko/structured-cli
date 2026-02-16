// Package docker provides parsers for docker command output.
// This package is in the adapters layer and implements parsers for
// converting raw docker command output into structured domain types.
package docker

// Command constants.
const (
	dockerCommand = "docker"
	subContainer  = "container"
	subCompose    = "compose"
	subPs         = "ps"
	subLs         = "ls"
	subList       = "list"
	subLogs       = "logs"
	subImages     = "images"
	subRun        = "run"
	subExec       = "exec"
	subPull       = "pull"
	subUp         = "up"
	subDown       = "down"
)

// PSResult represents the structured output of 'docker ps'.
type PSResult struct {
	// Success indicates whether the command completed successfully.
	Success bool `json:"success"`

	// Containers is the list of containers.
	Containers []Container `json:"containers"`
}

// Container represents a single Docker container.
type Container struct {
	// ID is the container ID.
	ID string `json:"id"`

	// Names is the container name(s).
	Names string `json:"names"`

	// Image is the image used to create the container.
	Image string `json:"image"`

	// Command is the command running in the container.
	Command string `json:"command"`

	// Created is when the container was created.
	Created string `json:"created"`

	// Status is the container status.
	Status string `json:"status"`

	// Ports is the port mappings.
	Ports string `json:"ports"`

	// Size is the container size (if -s flag used).
	Size string `json:"size,omitempty"`

	// State is the container state (running, exited, etc.).
	State string `json:"state"`
}

// BuildResult represents the structured output of 'docker build'.
type BuildResult struct {
	// Success indicates whether the build completed successfully.
	Success bool `json:"success"`

	// ImageID is the ID of the built image.
	ImageID string `json:"image_id"`

	// Tags is the list of tags applied to the image.
	Tags []string `json:"tags"`

	// Steps is the list of build steps executed.
	Steps []BuildStep `json:"steps"`

	// TotalSteps is the total number of build steps.
	TotalSteps int `json:"total_steps"`

	// Cached is the number of cached steps.
	Cached int `json:"cached"`

	// Duration is the build duration.
	Duration string `json:"duration,omitempty"`

	// Errors is the list of build errors.
	Errors []string `json:"errors"`

	// Warnings is the list of build warnings.
	Warnings []string `json:"warnings"`
}

// BuildStep represents a single step in a Docker build.
type BuildStep struct {
	// Number is the step number.
	Number int `json:"number"`

	// Instruction is the Dockerfile instruction (FROM, RUN, COPY, etc.).
	Instruction string `json:"instruction"`

	// Cached indicates if the step used cache.
	Cached bool `json:"cached"`

	// Duration is the step duration.
	Duration string `json:"duration,omitempty"`
}

// LogsResult represents the structured output of 'docker logs'.
type LogsResult struct {
	// Success indicates whether the command completed successfully.
	Success bool `json:"success"`

	// ContainerID is the container ID.
	ContainerID string `json:"container_id"`

	// Lines is the list of log lines.
	Lines []LogLine `json:"lines"`

	// TotalLines is the total number of log lines.
	TotalLines int `json:"total_lines"`
}

// LogLine represents a single log line from a container.
type LogLine struct {
	// Timestamp is the log timestamp (if --timestamps used).
	Timestamp string `json:"timestamp,omitempty"`

	// Stream is the output stream (stdout or stderr).
	Stream string `json:"stream"`

	// Message is the log message content.
	Message string `json:"message"`
}

// ImagesResult represents the structured output of 'docker images'.
type ImagesResult struct {
	// Success indicates whether the command completed successfully.
	Success bool `json:"success"`

	// Images is the list of images.
	Images []Image `json:"images"`
}

// Image represents a single Docker image.
type Image struct {
	// ID is the image ID.
	ID string `json:"id"`

	// Repository is the image repository.
	Repository string `json:"repository"`

	// Tag is the image tag.
	Tag string `json:"tag"`

	// Created is when the image was created.
	Created string `json:"created"`

	// Size is the image size.
	Size string `json:"size"`

	// Digest is the image digest.
	Digest string `json:"digest,omitempty"`
}

// RunResult represents the structured output of 'docker run'.
type RunResult struct {
	// Success indicates whether the container ran successfully.
	Success bool `json:"success"`

	// ContainerID is the ID of the created/running container.
	ContainerID string `json:"container_id"`

	// Output is the container output (if not detached).
	Output string `json:"output,omitempty"`

	// ExitCode is the container exit code (if completed).
	ExitCode int `json:"exit_code"`

	// Detached indicates if the container is running in background.
	Detached bool `json:"detached"`

	// Errors contains any error messages.
	Errors []string `json:"errors"`
}

// ExecResult represents the structured output of 'docker exec'.
type ExecResult struct {
	// Success indicates whether the command completed successfully.
	Success bool `json:"success"`

	// ContainerID is the container ID.
	ContainerID string `json:"container_id"`

	// Output is the command output.
	Output string `json:"output"`

	// ExitCode is the command exit code.
	ExitCode int `json:"exit_code"`

	// Errors contains any error messages.
	Errors []string `json:"errors"`
}

// PullResult represents the structured output of 'docker pull'.
type PullResult struct {
	// Success indicates whether the pull completed successfully.
	Success bool `json:"success"`

	// Image is the pulled image reference.
	Image string `json:"image"`

	// Digest is the image digest.
	Digest string `json:"digest"`

	// Status is the pull status (downloaded, up to date, etc.).
	Status string `json:"status"`

	// Layers is the list of layer statuses.
	Layers []LayerStatus `json:"layers"`

	// Errors contains any error messages.
	Errors []string `json:"errors"`
}

// LayerStatus represents the status of a single layer during pull.
type LayerStatus struct {
	// ID is the layer ID.
	ID string `json:"id"`

	// Status is the layer status (downloading, extracting, etc.).
	Status string `json:"status"`

	// Progress is the download progress percentage.
	Progress int `json:"progress"`
}

// ComposeUpResult represents the structured output of 'docker compose up'.
type ComposeUpResult struct {
	// Success indicates whether all services started successfully.
	Success bool `json:"success"`

	// Services is the list of services and their status.
	Services []ComposeService `json:"services"`

	// Networks is the list of created networks.
	Networks []string `json:"networks"`

	// Volumes is the list of created volumes.
	Volumes []string `json:"volumes"`

	// Errors contains any error messages.
	Errors []string `json:"errors"`

	// Warnings contains any warning messages.
	Warnings []string `json:"warnings"`
}

// ComposeService represents a service in a docker compose project.
type ComposeService struct {
	// Name is the service name.
	Name string `json:"name"`

	// Status is the service status.
	Status string `json:"status"`

	// ContainerID is the container ID for the service.
	ContainerID string `json:"container_id,omitempty"`

	// Image is the image used by the service.
	Image string `json:"image"`

	// State is the container state.
	State string `json:"state"`

	// Ports is the port mappings.
	Ports string `json:"ports,omitempty"`

	// Health is the health status.
	Health string `json:"health,omitempty"`
}

// ComposeDownResult represents the structured output of 'docker compose down'.
type ComposeDownResult struct {
	// Success indicates whether all services stopped successfully.
	Success bool `json:"success"`

	// StoppedContainers is the list of stopped containers.
	StoppedContainers []string `json:"stopped_containers"`

	// RemovedContainers is the list of removed containers.
	RemovedContainers []string `json:"removed_containers"`

	// RemovedNetworks is the list of removed networks.
	RemovedNetworks []string `json:"removed_networks"`

	// RemovedVolumes is the list of removed volumes.
	RemovedVolumes []string `json:"removed_volumes"`

	// Errors contains any error messages.
	Errors []string `json:"errors"`
}

// ComposePSResult represents the structured output of 'docker compose ps'.
type ComposePSResult struct {
	// Success indicates whether the command completed successfully.
	Success bool `json:"success"`

	// ProjectName is the compose project name.
	ProjectName string `json:"project_name"`

	// Services is the list of services and their status.
	Services []ComposeService `json:"services"`
}
