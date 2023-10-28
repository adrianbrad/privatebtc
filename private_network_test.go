package privatebtc_test

import (
	"context"
	"crypto/rand"
	"io"
	"testing"
	"testing/iotest"
	"time"

	"github.com/adrianbrad/privatebtc"
	"github.com/adrianbrad/privatebtc/btcsuite"
	"github.com/adrianbrad/privatebtc/docker/testcontainers"
	"github.com/adrianbrad/privatebtc/mock"
	"github.com/matryer/is"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

const nodeName = "name"

func TestPrivateNetwork(t *testing.T) {
	t.Run("NewPrivateNetwork", func(t *testing.T) {
		tests := map[string]struct {
			saltRandReader io.Reader
			opts           []privatebtc.Option
			errorAssertion require.ErrorAssertionFunc
		}{
			"NoOptions": {
				saltRandReader: rand.Reader,
				opts:           nil,
				errorAssertion: require.NoError,
			},
			"SaltReadError": {
				saltRandReader: iotest.ErrReader(assert.AnError),
				opts:           nil,
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, assert.AnError, i...)
				},
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				r := rand.Reader

				rand.Reader = test.saltRandReader

				t.Cleanup(func() { rand.Reader = r })

				_, err := privatebtc.NewPrivateNetwork(
					nil,
					nil,
					0,
					test.opts...,
				)

				test.errorAssertion(t, err)
			})
		}
	})

	t.Run("Start", func(t *testing.T) {
		t.Parallel()

		tests := map[string]struct {
			mockNodeService      *mock.NodeService
			mockRPCClientFactory *mock.RPCClientFactory
			nodes                int
			opts                 []privatebtc.Option
			errorAssertion       require.ErrorAssertionFunc
		}{
			"CreateContainersError": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return nil, assert.AnError
					},
				},
				mockRPCClientFactory: nil,
				nodes:                0,
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "create nodes", i...)
				},
			},
			"CannotConnectToDockerDaemonError": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return nil, privatebtc.ErrCannotConnectToDockerAPI
					},
				},
				mockRPCClientFactory: nil,
				nodes:                1,
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, privatebtc.ErrCannotConnectToDockerAPI, i...)
				},
			},
			"NewRPCClientError": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
						}, nil
					},
				},
				mockRPCClientFactory: &mock.RPCClientFactory{
					NewRPCClientFunc: func(
						hostRPCPort string,
						rpcUser string,
						rpcPass string,
					) (privatebtc.RPCClient, error) {
						return nil, assert.AnError
					},
				},
				nodes: 1,
				opts:  nil,
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "new rpc client", i...)
				},
			},
			"CreateWalletError": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
						}, nil
					},
				},
				mockRPCClientFactory: &mock.RPCClientFactory{
					NewRPCClientFunc: func(
						hostRPCPort string,
						rpcUser string,
						rpcPass string,
					) (privatebtc.RPCClient, error) {
						return &mock.RPCClient{
							CreateWalletFunc: func(ctx context.Context, walletName string) error {
								return assert.AnError
							},
						}, nil
					},
				},
				nodes: 1,
				opts:  []privatebtc.Option{privatebtc.WithWallet("test")},
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "create wallet", i...)
				},
			},
			"GetConnectionCountError": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
						}, nil
					},
				},
				mockRPCClientFactory: &mock.RPCClientFactory{
					NewRPCClientFunc: func(
						hostRPCPort string,
						rpcUser string,
						rpcPass string,
					) (privatebtc.RPCClient, error) {
						return &mock.RPCClient{
							CreateWalletFunc: func(ctx context.Context, walletName string) error {
								return nil
							},
							GetConnectionCountFunc: func(ctx context.Context) (int, error) {
								return 0, assert.AnError
							},
						}, nil
					},
				},
				nodes: 1,
				opts:  []privatebtc.Option{privatebtc.WithWallet("test")},
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(
						t,
						err,
						"get connection count for first node",
						i...,
					)
				},
			},
			"ConnectionCountShouldBe0": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
						}, nil
					},
				},
				mockRPCClientFactory: &mock.RPCClientFactory{
					NewRPCClientFunc: func(
						hostRPCPort string,
						rpcUser string,
						rpcPass string,
					) (privatebtc.RPCClient, error) {
						return &mock.RPCClient{
							CreateWalletFunc: func(ctx context.Context, walletName string) error {
								return nil
							},
							GetConnectionCountFunc: func(ctx context.Context) (int, error) {
								return 1, nil
							},
						}, nil
					},
				},
				nodes: 1,
				opts:  []privatebtc.Option{privatebtc.WithWallet("test")},
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(
						t,
						err,
						"node 0 should not have any peers, got: 1",
						i...,
					)
				},
			},
			"AddPeerError": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
						}, nil
					},
				},
				mockRPCClientFactory: &mock.RPCClientFactory{
					NewRPCClientFunc: func(
						hostRPCPort string,
						rpcUser string,
						rpcPass string,
					) (privatebtc.RPCClient, error) {
						return &mock.RPCClient{
							CreateWalletFunc: func(ctx context.Context, walletName string) error {
								return nil
							},
							GetConnectionCountFunc: func(ctx context.Context) (int, error) {
								return 0, nil
							},
							AddPeerFunc: func(ctx context.Context, peer privatebtc.Node) error {
								return assert.AnError
							},
						}, nil
					},
				},
				nodes: 2,
				opts:  []privatebtc.Option{privatebtc.WithWallet("test")},
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "add node", i...)
				},
			},
			"GetConnectionCountForConnectedNodeError": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
						}, nil
					},
				},
				mockRPCClientFactory: &mock.RPCClientFactory{
					NewRPCClientFunc: func(
						hostRPCPort string,
						rpcUser string,
						rpcPass string,
					) (privatebtc.RPCClient, error) {
						return &mock.RPCClient{
							CreateWalletFunc: func(_ context.Context, walletName string) error {
								return nil
							},
							GetConnectionCountFunc: func() func(ctx context.Context) (int, error) {
								var calls int

								return func(context.Context) (int, error) {
									if calls == 1 {
										return 0, assert.AnError
									}

									calls++

									return 0, nil
								}
							}(),
							AddPeerFunc: func(_ context.Context, peer privatebtc.Node) error {
								return nil
							},
						}, nil
					},
				},
				nodes: 1,
				opts:  []privatebtc.Option{privatebtc.WithWallet("test")},
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "get connection count for node", i...)
				},
			},
			"UnexpectedPeerCountError": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
						}, nil
					},
				},
				mockRPCClientFactory: &mock.RPCClientFactory{
					NewRPCClientFunc: func(
						hostRPCPort string,
						rpcUser string,
						rpcPass string,
					) (privatebtc.RPCClient, error) {
						return &mock.RPCClient{
							CreateWalletFunc: func(_ context.Context, walletName string) error {
								return nil
							},
							GetConnectionCountFunc: func(_ context.Context) (int, error) {
								return 0, nil
							},
							AddPeerFunc: func(_ context.Context, peer privatebtc.Node) error {
								return nil
							},
						}, nil
					},
				},
				nodes: 2,
				opts:  []privatebtc.Option{privatebtc.WithWallet("test")},
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(
						t,
						err,
						"should have 1 peers, got: 0",
						i...,
					)
				},
			},
			"Success": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								NameFunc: func() string {
									return nodeName
								},
							},
						}, nil
					},
				},
				mockRPCClientFactory: newPrivateNetworkStartSuccessRPCClientFactory(
					newChainReorgSuccessRPCClient(nil),
				),
				nodes:          2,
				opts:           []privatebtc.Option{privatebtc.WithWallet("test")},
				errorAssertion: require.NoError,
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ctx := context.Background()
				req := require.New(t)

				pn, err := privatebtc.NewPrivateNetwork(
					test.mockNodeService,
					test.mockRPCClientFactory,
					test.nodes,
					test.opts...,
				)
				req.NoError(err)

				err = pn.Start(ctx)
				test.errorAssertion(t, err)
			})
		}
	})

	t.Run("Close", func(t *testing.T) {
		tests := map[string]struct {
			mockNodeService *mock.NodeService
			errorAssertion  require.ErrorAssertionFunc
		}{
			"ClosesContainersError": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								CloseFunc: func() error {
									return assert.AnError
								},
							},
						}, nil
					},
				},
				errorAssertion: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "terminate node", i...)
				},
			},
			"Success": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								CloseFunc: func() error {
									return nil
								},
							},
						}, nil
					},
				},
				errorAssertion: require.NoError,
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ctx := context.Background()
				req := require.New(t)

				pn, err := privatebtc.NewPrivateNetwork(
					test.mockNodeService,
					newPrivateNetworkStartSuccessRPCClientFactory(
						newChainReorgSuccessRPCClient(nil),
					),
					2,
					[]privatebtc.Option{}...,
				)
				req.NoError(err)

				err = pn.Start(ctx)
				req.NoError(err)

				err = pn.Close()
				test.errorAssertion(t, err)
			})
		}
	})

	t.Run("Nodes", func(t *testing.T) {
		tests := map[string]struct {
			mockNodeService *mock.NodeService
			updateNodesFunc func(nodes []privatebtc.Node)
		}{
			"AssertNodesSliceDoesNotChange": {
				mockNodeService: &mock.NodeService{
					CreateNodesFunc: func(
						ctx context.Context,
						containerRequests []privatebtc.CreateNodeRequest,
					) ([]privatebtc.NodeHandler, error) {
						return []privatebtc.NodeHandler{
							&mock.NodeHandler{
								HostRPCPortFunc: func() string {
									return "1234"
								},
								CloseFunc: func() error {
									return nil
								},
							},
						}, nil
					},
				},
				updateNodesFunc: func(nodes []privatebtc.Node) {
					nodes[0] = privatebtc.Node{}
				},
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ctx := context.Background()
				req := require.New(t)

				pn, err := privatebtc.NewPrivateNetwork(
					test.mockNodeService,
					newPrivateNetworkStartSuccessRPCClientFactory(
						newChainReorgSuccessRPCClient(nil),
					),
					1,
					[]privatebtc.Option{}...,
				)
				req.NoError(err)

				err = pn.Start(ctx)
				req.NoError(err)

				initialNodes := pn.Nodes()

				test.updateNodesFunc(pn.Nodes())

				req.Equal(initialNodes, pn.Nodes())
			})
		}
	})
}

func TestBitcoinPrivateNetwork(t *testing.T) {
	// t.Skip()
	is := is.New(t)

	ctx := context.Background()

	pn, err := privatebtc.NewPrivateNetwork(
		&testcontainers.NodeService{},
		btcsuite.RPCClientFactory{},
		4,
		privatebtc.WithWallet(t.Name()),
	)
	is.NoErr(err)

	const (
		burningAddr = "bcrt1qzlfc3dw3ecjncvkwmwpvs84ejqzp4fr4agghm8"
	)

	startCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	err = pn.Start(startCtx)
	is.NoErr(err)

	t.Cleanup(func() {
		_ = pn.Close()
	})

	testNode := pn.Nodes()[0]

	blockHash, err := testNode.Fund(ctx)
	is.NoErr(err)

	syncCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = pn.Nodes().Sync(syncCtx, blockHash)
	is.NoErr(err)

	testNodeAddr, err := testNode.RPCClient().GetNewAddress(ctx, "test")
	is.NoErr(err)

	t.Run("ReplaceByFee", func(t *testing.T) {
		is := is.New(t)

		h, err := testNode.RPCClient().SendToAddress(ctx, burningAddr, 0.1)
		is.NoErr(err)

		mp, err := testNode.RPCClient().GetRawMempool(ctx)
		is.NoErr(err)

		is.Equal([]string{h}, mp)

		h2, err := privatebtc.ReplaceTransactionDrainToAddress(
			ctx,
			testNode.RPCClient(),
			h,
			testNodeAddr,
		)
		is.NoErr(err)

		mp, err = testNode.RPCClient().GetRawMempool(ctx)
		is.NoErr(err)

		is.Equal([]string{h2}, mp)
	})

	t.Run("Mining", func(t *testing.T) {
		is := is.New(t)

		initialBalance, err := testNode.RPCClient().GetBalance(ctx)
		is.NoErr(err)

		initialBlockCount, err := testNode.RPCClient().GetBlockCount(ctx)
		is.NoErr(err)

		t.Logf("balance before mining: %f", initialBalance)

		_, err = testNode.RPCClient().GenerateToAddress(ctx, 1, burningAddr)
		is.NoErr(err)

		bal, err := testNode.RPCClient().GetBalance(ctx)
		is.NoErr(err)

		expectedBalance := initialBalance

		expectedBalance.Trusted += 50
		expectedBalance.Immature -= 50

		is.Equal(bal, expectedBalance)

		t.Logf("balance after mining: %+v", bal)

		blockCount, err := testNode.RPCClient().GetBlockCount(ctx)
		is.NoErr(err)

		is.Equal(blockCount, initialBlockCount+1)
	})

	t.Run("Transfer", func(t *testing.T) {
		is := is.New(t)

		receiverNode := pn.Nodes()[1]

		receiverAddr, err := receiverNode.RPCClient().GetNewAddress(ctx, t.Name())
		is.NoErr(err)

		const amount = 0.1

		txHash, err := testNode.RPCClient().SendToAddress(ctx, receiverAddr, amount)
		is.NoErr(err)

		start := time.Now()

		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		err = pn.Nodes().EnsureTransactionInEveryMempool(ctx, txHash)
		is.NoErr(err)

		t.Logf("tx in all mempools after %s", time.Since(start))

		blockHashes, err := pn.Nodes()[3].RPCClient().GenerateToAddress(ctx, 1, burningAddr)
		is.NoErr(err)

		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err = pn.Nodes().Sync(ctx, blockHashes[0])
		is.NoErr(err)
	})

	t.Run("ChainReorg", func(t *testing.T) {
		is := is.New(t)

		receiverNode := pn.Nodes()[2]

		cr, err := pn.NewChainReorgWithAssertion(1)
		is.NoErr(err)

		reorgNode, err := cr.DisconnectNode(ctx)
		is.NoErr(err)

		initialBlockCount, err := reorgNode.RPCClient().GetBlockCount(ctx)
		is.NoErr(err)

		t.Logf("initial block count: %d", initialBlockCount)

		nodesWithoutReorg := slices.Delete(pn.Nodes(), 1, 2)

		receiverInitialBalance, err := receiverNode.RPCClient().GetBalance(ctx)
		is.NoErr(err)

		receiverNodeAddr, err := receiverNode.RPCClient().GetNewAddress(ctx, t.Name())
		is.NoErr(err)

		const amount = 0.1

		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		txHash, err := cr.SendTransactionOnNetwork(ctx, receiverNodeAddr, amount)
		is.NoErr(err)

		receiverPendingBalanceAfterTx, err := receiverNode.RPCClient().GetBalance(ctx)
		is.NoErr(err)

		is.Equal(
			receiverPendingBalanceAfterTx.Pending,
			receiverInitialBalance.Pending+amount,
		)

		blockHashes, err := cr.MineBlocksOnNetwork(ctx, 1)
		is.NoErr(err)

		txAfterMine, err := receiverNode.RPCClient().GetTransaction(ctx, txHash)
		is.NoErr(err)

		is.Equal(txAfterMine.BlockHash, blockHashes[0])

		receiverBalanceAfterTransfer, err := receiverNode.RPCClient().GetBalance(ctx)
		is.NoErr(err)
		is.Equal(
			receiverBalanceAfterTransfer.Trusted,
			receiverInitialBalance.Trusted+amount,
		)

		_, err = cr.MineBlocksOnDisconnectedNode(ctx, 2)
		is.NoErr(err)

		err = cr.ReconnectNode(ctx)
		is.NoErr(err)

		txAfterReorg, err := receiverNode.RPCClient().GetTransaction(ctx, txHash)
		is.NoErr(err)
		is.Equal(txAfterReorg.BlockHash, "")

		receiverBalanceAfterReorg, err := receiverNode.RPCClient().GetBalance(ctx)
		is.NoErr(err)

		receiverBalanceAfterReorg.Pending -= amount

		is.Equal(receiverBalanceAfterReorg, receiverInitialBalance)

		syncCtx, cancel = context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		err = nodesWithoutReorg.EnsureTransactionInEveryMempool(syncCtx, txHash)
		is.NoErr(err)
	})
}
