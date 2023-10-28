package testcontainers

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"sync"

	"github.com/adrianbrad/privatebtc"
	"github.com/adrianbrad/privatebtc/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"
)

// Ensure NodeService implements go-privatebtc.NodeService.
var _ privatebtc.NodeService = (*NodeService)(nil)

// NodeService is a testcontainers implementation of go-privatebtc.NodeService.
// It is used to create containers.
type NodeService struct {
	SlogHandler slog.Handler

	initOnce       sync.Once // guards init of NodeService
	testcontLogger testcontainers.Logging
}

// CreateNodes creates bitcoin node containers in parallel.
func (s *NodeService) CreateNodes(
	ctx context.Context,
	nodeRequests []privatebtc.CreateNodeRequest,
) ([]privatebtc.NodeHandler, error) {
	s.initOnce.Do(s.init)

	reqs := make([]testcontainers.GenericContainerRequest, len(nodeRequests))

	for i, nodeReq := range nodeRequests {
		reqs[i] = testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image: docker.BitcoinImage,
				ExposedPorts: []string{
					privatebtc.RPCRegtestDefaultPort + "/tcp",
				},
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
				WaitingFor: wait.ForLog("init message: Done loading"),
				Name:       fmt.Sprintf("privatebtc_node_%d", i),
				HostConfigModifier: func(config *container.HostConfig) {
					config.AutoRemove = true
					config.RestartPolicy = container.RestartPolicy{Name: "no"}
				},
			},
			Started: true,
			Logger:  s.testcontLogger,
		}
	}

	testConts, err := testcontainers.ParallelContainers(
		ctx,
		reqs,
		testcontainers.ParallelContainersOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("create containers in parallel: %w", err)
	}

	eg, egCtx := errgroup.WithContext(ctx)

	conts := make([]privatebtc.NodeHandler, len(testConts))

	for i := range testConts {
		i := i

		eg.Go(func() error {
			var err error

			conts[i], err = newNodeHandler(egCtx, testConts[i])
			if err != nil {
				return fmt.Errorf("create node %d: %w", i, err)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return conts, nil
}

func (s *NodeService) init() {
	s.testcontLogger = log.New(io.Discard, "", 0)

	if s.SlogHandler != nil {
		s.testcontLogger = logger{slog.New(s.SlogHandler)}
	}

	testcontainers.Logger = s.testcontLogger
}

type logger struct {
	*slog.Logger
}

func (l logger) Printf(format string, v ...any) {
	l.Info(fmt.Sprintf(format, v...))
}
