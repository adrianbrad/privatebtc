package privatebtc_test

import (
	"context"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/adrianbrad/privatebtc"
	"github.com/adrianbrad/privatebtc/mock"
	"github.com/stretchr/testify/require"
)

func newPrivateNetworkStartSuccessRPCClientFactory(
	mockRPCClient *mock.RPCClient,
) *mock.RPCClientFactory {
	return &mock.RPCClientFactory{
		NewRPCClientFunc: func(
			hostRPCPort string,
			rpcUser string,
			rpcPass string,
		) (privatebtc.RPCClient, error) {
			return mockRPCClient, nil
		},
	}
}

func newRPCClientFactoryWithDetachedNode(
	defaultMockRPCClient *mock.RPCClient,
	disconnectedNodePort string,
	disconnectedNodeRPCClient *mock.RPCClient,
) *mock.RPCClientFactory {
	return &mock.RPCClientFactory{
		NewRPCClientFunc: func(
			hostRPCPort string,
			rpcUser string,
			rpcPass string,
		) (privatebtc.RPCClient, error) {
			if hostRPCPort == disconnectedNodePort {
				return disconnectedNodeRPCClient, nil
			}

			return defaultMockRPCClient, nil
		},
	}
}

func newPrivateNetworkStartSuccessDockerService(
	containers ...privatebtc.NodeHandler,
) *mock.NodeService {
	return &mock.NodeService{
		CreateNodesFunc: func(
			ctx context.Context,
			containerRequests []privatebtc.CreateNodeRequest,
		) ([]privatebtc.NodeHandler, error) {
			return containers, nil
		},
	}
}

func newPrivateNetworkStartSuccessNodeHandler() *mock.NodeHandler {
	return &mock.NodeHandler{
		HostRPCPortFunc: func() string {
			return "1234"
		},
		InternalIPFunc: func() string {
			return "1234"
		},
		NameFunc: func() string {
			return nodeName
		},
	}
}

func newPrivateNetworkStartSuccessContainerWithPort(port string) *mock.NodeHandler {
	return &mock.NodeHandler{
		HostRPCPortFunc: func() string {
			return port
		},
		InternalIPFunc: func() string {
			return "1234"
		},
		NameFunc: func() string {
			return nodeName
		},
	}
}

type newChainReorgFunc func(
	t *testing.T,
	dockerService privatebtc.NodeService,
	rpcClientFactory privatebtc.RPCClientFactory,
	nodes int,
	disconnectedNodeIndex int,
) (*privatebtc.PrivateNetwork, privatebtc.ChainReorgManager)

func newChainReorg(
	t *testing.T,
	dockerService privatebtc.NodeService,
	rpcClientFactory privatebtc.RPCClientFactory,
	nodes int,
	disconnectedNodeIndex int,
) (*privatebtc.PrivateNetwork, privatebtc.ChainReorgManager) {
	t.Helper()

	req := require.New(t)

	pn, err := privatebtc.NewPrivateNetwork(
		dockerService,
		rpcClientFactory,
		nodes,
	)
	req.NoError(err)

	err = pn.Start(context.Background())
	req.NoError(err)

	crm, err := pn.NewChainReorgWithAssertion(disconnectedNodeIndex)
	req.NoError(err)

	return pn, crm
}

func newChainReorgWithDisconnect(
	t *testing.T,
	dockerService privatebtc.NodeService,
	rpcClientFactory privatebtc.RPCClientFactory,
	nodes int,
	disconnectedNodeIndex int,
) (*privatebtc.PrivateNetwork, privatebtc.ChainReorgManager) {
	t.Helper()

	req := require.New(t)

	pn, crm := newChainReorg(t, dockerService, rpcClientFactory, nodes, disconnectedNodeIndex)

	_, err := crm.DisconnectNode(context.TODO())
	req.NoError(err)

	return pn, crm
}

func newChainReorgSuccessRPCClient(peerCount *atomic.Int64) *mock.RPCClient {
	if peerCount == nil {
		peerCount = new(atomic.Int64)
	}

	return &mock.RPCClient{
		CreateWalletFunc: func(context.Context, string) error {
			return nil
		},

		GetConnectionCountFunc: func(context.Context) (int, error) {
			return int(peerCount.Load()), nil
		},

		AddPeerFunc: func(context.Context, privatebtc.Node) error {
			peerCount.Add(1)

			return nil
		},

		RemovePeerFunc: func(context.Context, privatebtc.Node) error {
			peerCount.Add(-1)

			return nil
		},
	}
}

func getCallerFunction() string {
	pc, _, _, ok := runtime.Caller(3)
	if !ok {
		panic("failed to get caller")
	}

	details := runtime.FuncForPC(pc)

	return details.Name()
}

func callStackFunctionContains(s string) bool {
	rpc := make([]uintptr, 16)

	n := runtime.Callers(0, rpc)

	frames := runtime.CallersFrames(rpc[:n])

	for {
		frame, more := frames.Next()

		if strings.Contains(frame.Function, s) {
			return true
		}

		if !more {
			break
		}
	}

	return false
}
