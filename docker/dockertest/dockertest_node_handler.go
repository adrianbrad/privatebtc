package dockertest

import (
	"fmt"
	"net"

	"github.com/adrianbrad/privatebtc"
	"github.com/ory/dockertest/v3"
)

var _ privatebtc.NodeHandler = (*NodeHandler)(nil)

// NodeHandler represents a docker container.
type NodeHandler struct {
	res         *dockertest.Resource
	containerIP string
	hostRPCPort string
	name        string
}

func newNodeHandler(res *dockertest.Resource) (*NodeHandler, error) {
	host := res.GetHostPort(privatebtc.RPCRegtestDefaultPort + "/tcp")

	_, hostRPCPort, err := net.SplitHostPort(host)
	if err != nil {
		return nil, fmt.Errorf("split host port: %w", err)
	}

	containerIP := res.Container.NetworkSettings.IPAddress

	return &NodeHandler{
		res:         res,
		hostRPCPort: hostRPCPort,
		containerIP: containerIP,
		name:        res.Container.Name,
	}, nil
}

// InternalIP returns the internal IP of the container.
func (n NodeHandler) InternalIP() string {
	return n.containerIP
}

// HostRPCPort returns the host RPC port of the container.
func (n NodeHandler) HostRPCPort() string {
	return n.hostRPCPort
}

// Name returns the container name.
func (n NodeHandler) Name() string {
	return n.name
}

// Close closes the container.
func (n NodeHandler) Close() error {
	return n.res.Close()
}
