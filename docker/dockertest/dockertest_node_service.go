package dockertest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/adrianbrad/privatebtc"
	pbtcdocker "github.com/adrianbrad/privatebtc/docker"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"golang.org/x/sync/errgroup"
)

var _ privatebtc.NodeService = (*NodeService)(nil)

// NodeService is an ory/dockertest implementation of privatebtc.NodeService.
// It is used to create containers.
type NodeService struct {
	SlogHandler slog.Handler

	logger *slog.Logger

	initOnce sync.Once // guards init of NodeService
}

// CreateNodes creates docker container nodes in parallel.
func (s *NodeService) CreateNodes(
	ctx context.Context,
	nodeRequests []privatebtc.CreateNodeRequest,
) ([]privatebtc.NodeHandler, error) {
	s.initOnce.Do(s.init)

	dockerHost, err := pbtcdocker.GetDockerHost()
	if err != nil {
		return nil, fmt.Errorf("get docker host: %w", err)
	}

	pool, err := dockertest.NewPool(dockerHost)
	if err != nil {
		return nil, fmt.Errorf("create docker pool: %w", err)
	}

	if err := pool.Client.Ping(); err != nil {
		return nil, fmt.Errorf("ping docker: %w", err)
	}

	containers := make([]*dockertest.Resource, len(nodeRequests))

	eg, _ := errgroup.WithContext(ctx)

	for i, nodeReq := range nodeRequests {
		i, nodeReq := i, nodeReq

		eg.Go(func() error {
			containerName := fmt.Sprintf("privatebtc_node_%d", i)

			s.logger.Info("🐳⌛ Creating container", "name", containerName)

			sub := strings.Split(pbtcdocker.BitcoinImage, ":")

			imageName := sub[0]
			imageTag := sub[1]

			res, err := pool.RunWithOptions(
				&dockertest.RunOptions{
					Name:       fmt.Sprintf("privatebtc_node_%d", i),
					Repository: imageName,
					Tag:        imageTag,
					Cmd: []string{
						"-regtest=1",
						"-rpcallowip=172.17.0.0/16", // allow requests coming from the docker host
						"-rpcbind=0.0.0.0",
						"-dnsseed=0",
						"-txindex",
						fmt.Sprintf("-rpcauth=%s", nodeReq.RPCAuth),
						fmt.Sprintf("-fallbackfee=%f", nodeReq.FallbackFee),
						// "blocksonly=1", // use this flag in order to disable mempool and
						// cause walletnotify to trigger when transaction has only 1 confirmation
					},
					ExposedPorts: []string{
						privatebtc.RPCRegtestDefaultPort + "/tcp",
					},
				},
				func(hostConfig *docker.HostConfig) {
					hostConfig.AutoRemove = true
					hostConfig.RestartPolicy = docker.RestartPolicy{Name: "no"}
				},
			)
			if err != nil {
				return fmt.Errorf("run with options: %w", err)
			}

			s.logger.Info("🐳✅ NodeHandler created", "name", containerName)

			containers[i] = res

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		err = fmt.Errorf("wait for containers: %w", err)

		if closeErr := closeContainers(containers); closeErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf("close containers: %w", closeErr),
			)
		}

		return nil, err
	}

	conts := make([]privatebtc.NodeHandler, len(containers))

	for i, res := range containers {
		conts[i], err = newNodeHandler(res)
		if err != nil {
			err = fmt.Errorf("new container: %w", err)

			if closeErr := closeContainers(containers); closeErr != nil {
				err = errors.Join(
					err,
					fmt.Errorf("close containers: %w", closeErr),
				)
			}

			return nil, err
		}
	}

	return conts, nil
}

func (s *NodeService) init() {
	s.logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	if s.SlogHandler == nil {
		s.logger = slog.New(s.SlogHandler)
	}
}

func closeContainers(containers []*dockertest.Resource) error {
	var err error

	for i, res := range containers {
		if res != nil {
			closeErr := res.Close()
			if closeErr != nil {
				err = errors.Join(err, fmt.Errorf("close container %d: %w", i, closeErr))
			}
		}
	}

	return err
}
