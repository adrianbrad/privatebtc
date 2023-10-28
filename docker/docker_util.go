package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// NewClient returns a new docker client.
func NewClient() (*client.Client, error) {
	host, err := GetDockerHost()
	if err != nil {
		return nil, fmt.Errorf("get docker host: %w", err)
	}

	return client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithHost(host),
	)
}

// ImageNotFoundError is used to whenever the ruimarinho/bitcoin-core docker Image
// is not found in the local Image cache.
type ImageNotFoundError struct {
	Image string
}

func (e *ImageNotFoundError) Error() string {
	return fmt.Sprintf(
		"Image %[1]s not found, consider pulling it using: docker pull %[1]s",
		e.Image,
	)
}

// Is implements errors.Is.
func (e *ImageNotFoundError) Is(target error) bool {
	var t *ImageNotFoundError
	ok := errors.As(target, &t)

	return ok && t.Image == e.Image
}

// CheckImageExistsInLocalCache returns whether the given docker Image
// exists in the local Image cache.
func CheckImageExistsInLocalCache(
	ctx context.Context,
	dockerClient *client.Client,
	imageName string,
) error {
	_, _, err := dockerClient.ImageInspectWithRaw(ctx, imageName)
	if client.IsErrNotFound(err) {
		return &ImageNotFoundError{Image: imageName}
	}

	if err != nil {
		return fmt.Errorf("err while inspecting Image %s: %w", imageName, err)
	}

	return nil
}

// PullImage pulls the docker Image with the given name.
func PullImage(
	ctx context.Context,
	dockerClient *client.Client,
	imageName string,
) error {
	reader, err := dockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pull image: %w", err)
	}

	defer reader.Close()

	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("discard Image pull reader: %w", err)
	}

	if err := reader.Close(); err != nil {
		return fmt.Errorf("close reader: %w", err)
	}

	return nil
}

// GetDockerHost returns the actual docker host from different alternatives.
// Windows is not supported.
func GetDockerHost() (string, error) {
	// Retrieve DOCKER_HOST environment variable
	dockerHost := os.Getenv("DOCKER_HOST")

	// If it's set, use that
	if dockerHost != "" {
		return dockerHost, nil
	}

	// Otherwise, based on the OS, return potential defaults
	switch runtime.GOOS {
	case "darwin": // macOS
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("get current user: %w", err)
		}

		dockerHost = fmt.Sprintf("%s/.docker/run/docker.sock", usr.HomeDir)

	case "linux":
		dockerHost = "/var/run/docker.sock"

	default:
		return "", os.ErrNotExist
	}

	if err := unixSocketExists(dockerHost); err != nil {
		return "", fmt.Errorf("unix socket does not exists at %q: %w", dockerHost, err)
	}

	return "unix://" + dockerHost, nil
}

func unixSocketExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat unix socket: %w", err)
	}

	isSocket := info.Mode()&os.ModeSocket != 0

	if !isSocket {
		return os.ErrInvalid
	}

	return nil
}
