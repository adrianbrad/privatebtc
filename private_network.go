package privatebtc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/exp/slices"
)

// NodeService is a service for creating Bitcoin Nodes.
type NodeService interface {
	CreateNodes(
		ctx context.Context,
		nodeRequests []CreateNodeRequest,
	) ([]NodeHandler, error)
}

// PrivateNetwork is a Bitcoin private network.
type PrivateNetwork struct {
	logger           *slog.Logger
	nodeService      NodeService
	rpcClientFactory RPCClientFactory
	nodes            Nodes
	nodeRequests     []CreateNodeRequest
	timeout          *time.Duration
	walletName       *string
	rpcUser          string
	rpcPassword      string
}

// Default Bitcoin Core ports for regtest.
const (
	RPCRegtestDefaultPort = "18443"
	P2PRegtestDefaultPort = "18444"
)

// NewPrivateNetwork creates a new private network with the given number of nodes.
func NewPrivateNetwork(
	nodeService NodeService,
	rpcClientFactory RPCClientFactory,
	nodes int,
	opts ...Option,
) (*PrivateNetwork, error) {
	options := defaultOptions()

	for i := range opts {
		opts[i].apply(options)
	}

	rpcAuth, err := newRPCAuth(options.rpcUser, options.rpcPass)
	if err != nil {
		return nil, fmt.Errorf("new rpc auth: %w", err)
	}

	nodeRequests := make([]CreateNodeRequest, nodes)

	for i := range nodeRequests {
		nodeRequests[i] = CreateNodeRequest{
			RPCAuth:     rpcAuth,
			FallbackFee: options.fallbackFee,
		}
	}

	return &PrivateNetwork{
		logger:           slog.New(options.handler),
		nodeService:      nodeService,
		rpcClientFactory: rpcClientFactory,
		nodes:            nil,
		nodeRequests:     nodeRequests,
		timeout:          options.timeout,
		walletName:       options.walletName,
		rpcUser:          options.rpcUser,
		rpcPassword:      options.rpcPass,
	}, nil
}

// Start creates the private network nodes and connects them.
func (n *PrivateNetwork) Start(ctx context.Context) error {
	n.logger.Info("⌛ Creating nodes")

	nodes, err := n.nodeService.CreateNodes(ctx, n.nodeRequests)
	if err != nil {
		return fmt.Errorf("create nodes: %w", err)
	}

	n.logger.Info("🐳✅ Successfully created nodes")

	n.nodes = make([]Node, len(nodes))

	for i, nodeHandler := range nodes {
		rpcClient, err := n.rpcClientFactory.NewRPCClient(
			nodeHandler.HostRPCPort(),
			n.rpcUser,
			n.rpcPassword,
		)
		if err != nil {
			return fmt.Errorf("new rpc client: %w", err)
		}

		if n.walletName != nil {
			if err := rpcClient.CreateWallet(ctx, *n.walletName); err != nil {
				return fmt.Errorf("create wallet: %w", err)
			}
		}

		n.nodes[i] = Node{
			id:          i,
			name:        fmt.Sprintf("Node %d", i),
			rpcClient:   rpcClient,
			nodeHandler: nodeHandler,
			pn:          n,
		}
	}

	n.logger.Info("🔗⌛ Connecting nodes")

	if err := connectNodes(ctx, n.nodes); err != nil {
		return fmt.Errorf("connect nodes: %w", err)
	}

	n.logger.Info("🔗✅ Successfully connected nodes")

	return nil
}

// Nodes returns a copy of the nodes in the private network.
func (n *PrivateNetwork) Nodes() Nodes {
	return slices.Clone(n.nodes)
}

// CreateNodeRequest is used to create a node.
type CreateNodeRequest struct {
	RPCAuth     string
	FallbackFee float64
}

// Close terminates all nodes in the private network.
func (n *PrivateNetwork) Close() error {
	var errs error

	for i := range n.nodes {
		if err := n.nodes[i].nodeHandler.Close(); err != nil {
			errs = errors.Join(errs, fmt.Errorf("terminate node %d: %w", i, err))
		}
	}

	return errs
}
