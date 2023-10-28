package testcontainers

import (
	"context"
	"fmt"

	"github.com/adrianbrad/privatebtc"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/sync/errgroup"
)

var _ privatebtc.NodeHandler = (*NodeHandler)(nil)

// NodeHandler represents a bitcoin node running in a docker container.
type NodeHandler struct {
	cont        testcontainers.Container
	containerIP string
	hostRPCPort string
	name        string
}

func newNodeHandler(
	ctx context.Context,
	testCont testcontainers.Container,
) (*NodeHandler, error) {
	var (
		containerIP string
		hostRPCPort nat.Port
		name        string
	)

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error

		containerIP, err = testCont.ContainerIP(egCtx)
		if err != nil {
			return fmt.Errorf("get container ip: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		var err error

		hostRPCPort, err = testCont.MappedPort(egCtx, privatebtc.RPCRegtestDefaultPort)
		if err != nil {
			return fmt.Errorf("get rpc port: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		var err error

		name, err = testCont.Name(egCtx)
		if err != nil {
			return fmt.Errorf("get container name: %w", err)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &NodeHandler{
		cont:        testCont,
		containerIP: containerIP,
		hostRPCPort: hostRPCPort.Port(),
		name:        name,
	}, nil
}

// InternalIP returns the IP address of the container.
func (c NodeHandler) InternalIP() string {
	return c.containerIP
}

// HostRPCPort returns the host RPC port of the container.
func (c NodeHandler) HostRPCPort() string {
	return c.hostRPCPort
}

// Name returns the container name.
func (c NodeHandler) Name() string {
	return c.name
}

// Close terminates the container.
func (c NodeHandler) Close() error {
	return c.cont.Terminate(context.Background())
}
