package privatebtc_test

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adrianbrad/privatebtc"
	"github.com/adrianbrad/privatebtc/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint: goconst
func TestChainReorg(t *testing.T) {
	t.Parallel()

	type chainReorgArgs struct {
		dockerService         privatebtc.NodeService
		rpcClientFactory      privatebtc.RPCClientFactory
		nodes                 int
		disconnectedNodeIndex int
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	t.Run("DisconnectNode", func(t *testing.T) {
		t.Parallel()

		type args struct {
			ctx context.Context
		}

		tests := map[string]struct {
			chainReorgArgs    chainReorgArgs
			args              args
			assertErr         require.ErrorAssertionFunc
			expectedNodeIndex int
		}{
			"Success": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessNodeHandler(),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: newPrivateNetworkStartSuccessRPCClientFactory(
						newChainReorgSuccessRPCClient(nil),
					),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				assertErr:         require.NoError,
				expectedNodeIndex: 0,
			},
			"GetConnectionCountErr": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessNodeHandler(),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						f := c.GetConnectionCountFunc

						c.GetConnectionCountFunc = func(ctx context.Context) (int, error) {
							if strings.Contains(getCallerFunction(), "DisconnectNode") {
								return 0, assert.AnError
							}

							return f(ctx)
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, assert.AnError, i...)
				},
				expectedNodeIndex: -1,
			},
			"TimeoutErr": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessNodeHandler(),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						f := c.GetConnectionCountFunc

						c.GetConnectionCountFunc = func(ctx context.Context) (int, error) {
							if strings.Contains(getCallerFunction(), "DisconnectNode") {
								return 1, nil
							}

							return f(ctx)
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(
						t,
						err,
						"disconnected node: unexpected peer count",
						i...,
					)
				},
				expectedNodeIndex: -1,
			},
			"AssertConnectionCountError": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessNodeHandler(),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						f := c.GetConnectionCountFunc

						c.GetConnectionCountFunc = func(ctx context.Context) (int, error) {
							if len(c.GetConnectionCountCalls()) > 4 {
								return 0, assert.AnError
							}

							return f(ctx)
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(
						t,
						err,
						"get connection count for node",
						i...,
					)
				},
				expectedNodeIndex: -1,
			},
			"UnexpectedPeerCount": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessNodeHandler(),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						f := c.GetConnectionCountFunc

						c.GetConnectionCountFunc = func(ctx context.Context) (int, error) {
							if len(c.GetConnectionCountCalls()) > 4 {
								return 2, nil
							}

							return f(ctx)
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(
						t,
						err,
						"assert peer connection count",
						i...,
					)
				},
				expectedNodeIndex: -1,
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				pn, crm := newChainReorg(
					t,
					test.chainReorgArgs.dockerService,
					test.chainReorgArgs.rpcClientFactory,
					test.chainReorgArgs.nodes,
					test.chainReorgArgs.disconnectedNodeIndex,
				)

				n, err := crm.DisconnectNode(test.args.ctx)
				test.assertErr(t, err)

				var expectedNode privatebtc.Node

				if test.expectedNodeIndex >= 0 {
					expectedNode = pn.Nodes()[test.expectedNodeIndex]
				}

				require.Equal(t, expectedNode, n)
			})
		}
	})

	t.Run("SendTransactionOnNetwork", func(t *testing.T) {
		t.Parallel()

		// nolint: containedctx
		type args struct {
			ctx             context.Context
			receiverAddress string
			amount          float64
		}

		tests := map[string]struct {
			chainReorgArgs    chainReorgArgs
			args              args
			newChainReorgFunc newChainReorgFunc
			assertErr         require.ErrorAssertionFunc
			expectedHash      string
		}{
			"Success": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						c := newChainReorgSuccessRPCClient(peerCount)

						c.SendToAddressFunc = func(
							context.Context,
							string,
							float64,
						) (string, error) {
							return "Hash", nil
						}

						c.GetRawMempoolFunc = func(ctx context.Context) ([]string, error) {
							return []string{"Hash"}, nil
						}

						disconnectedNodeRPCClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeRPCClient.GetRawMempoolFunc = func(
							ctx context.Context,
						) ([]string, error) {
							return []string{}, nil
						}

						return newRPCClientFactoryWithDetachedNode(
							c,
							"disconnected",
							disconnectedNodeRPCClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr:         require.NoError,
				expectedHash:      "Hash",
			},
			"MustDisconnectFirstError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: newPrivateNetworkStartSuccessRPCClientFactory(
						newChainReorgSuccessRPCClient(nil),
					),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorg,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, privatebtc.ErrChainReorgMustDisconnectNodeFirst)
				},
				expectedHash: "",
			},
			"SendToAddressError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						c.SendToAddressFunc = func(
							context.Context,
							string,
							float64,
						) (string, error) {
							return "", assert.AnError
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(
							c,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, assert.AnError)
				},
				expectedHash: "",
			},
			"GetRawMempoolError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						c.SendToAddressFunc = func(
							context.Context,
							string,
							float64,
						) (string, error) {
							return "Hash", nil
						}

						c.GetRawMempoolFunc = func(ctx context.Context) ([]string, error) {
							return nil, assert.AnError
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "get raw mempool")
				},
				expectedHash: "",
			},

			"IsTransactionInDisconnectedNodeMempoolError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						c := newChainReorgSuccessRPCClient(peerCount)

						c.SendToAddressFunc = func(
							context.Context,
							string,
							float64,
						) (string, error) {
							return "Hash", nil
						}

						c.GetRawMempoolFunc = func(ctx context.Context) ([]string, error) {
							return []string{"Hash"}, nil
						}

						disconnectedNodeRPCClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeRPCClient.GetRawMempoolFunc = func(
							ctx context.Context,
						) ([]string, error) {
							return nil, assert.AnError
						}

						return newRPCClientFactoryWithDetachedNode(
							c,
							"disconnected",
							disconnectedNodeRPCClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(
						t,
						err,
						"is transaction in disconnected node mempool",
					)
				},
				expectedHash: "",
			},
			"IsTransactionInDisconnectedNodeMempoolTrue": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						c := newChainReorgSuccessRPCClient(peerCount)

						c.SendToAddressFunc = func(
							context.Context,
							string,
							float64,
						) (string, error) {
							return "Hash", nil
						}

						c.GetRawMempoolFunc = func(ctx context.Context) ([]string, error) {
							return []string{"Hash"}, nil
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(
							c,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(
						t,
						err,
						privatebtc.ErrTxFoundInMempool,
					)
				},
				expectedHash: "",
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, crm := test.newChainReorgFunc(
					t,
					test.chainReorgArgs.dockerService,
					test.chainReorgArgs.rpcClientFactory,
					test.chainReorgArgs.nodes,
					test.chainReorgArgs.disconnectedNodeIndex,
				)

				hash, err := crm.SendTransactionOnNetwork(
					test.args.ctx,
					test.args.receiverAddress,
					test.args.amount,
				)
				test.assertErr(t, err)

				require.Equal(t, test.expectedHash, hash)
			})
		}
	})
	t.Run("SendTransactionOnDisconnectedNode", func(t *testing.T) {
		t.Parallel()

		type args struct {
			ctx             context.Context
			receiverAddress string
			amount          float64
		}

		tests := map[string]struct {
			chainReorgArgs    chainReorgArgs
			args              args
			newChainReorgFunc newChainReorgFunc
			assertErr         require.ErrorAssertionFunc
			expectedHash      string
		}{
			"Success": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						disconenctedNodeRPCClient := newChainReorgSuccessRPCClient(peerCount)

						disconenctedNodeRPCClient.SendToAddressFunc = func(
							context.Context,
							string,
							float64,
						) (string, error) {
							return "Hash", nil
						}

						disconenctedNodeRPCClient.GetRawMempoolFunc = func(
							ctx context.Context,
						) ([]string, error) {
							return []string{"Hash"}, nil
						}

						networkRPCClient := newChainReorgSuccessRPCClient(peerCount)

						networkRPCClient.GetRawMempoolFunc = func(
							ctx context.Context,
						) ([]string, error) {
							return []string{}, nil
						}

						return newRPCClientFactoryWithDetachedNode(
							networkRPCClient,
							"disconnected",
							disconenctedNodeRPCClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr:         require.NoError,
				expectedHash:      "Hash",
			},
			"MustDisconnectFirstError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: newPrivateNetworkStartSuccessRPCClientFactory(
						newChainReorgSuccessRPCClient(nil),
					),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorg,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, privatebtc.ErrChainReorgMustDisconnectNodeFirst)
				},
				expectedHash: "",
			},
			"SendToAddressError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						c.SendToAddressFunc = func(
							context.Context,
							string,
							float64,
						) (string, error) {
							return "", assert.AnError
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(
							c,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, assert.AnError)
				},
				expectedHash: "",
			},
			"GetRawMempoolError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						c.SendToAddressFunc = func(
							context.Context,
							string,
							float64,
						) (string, error) {
							return "Hash", nil
						}

						c.GetRawMempoolFunc = func(ctx context.Context) ([]string, error) {
							return nil, assert.AnError
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "get raw mempool")
				},
				expectedHash: "",
			},
			"IsTransactionInNetworkMempoolError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						disconnectedNodeRPCClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeRPCClient.SendToAddressFunc = func(
							context.Context,
							string,
							float64,
						) (string, error) {
							return "Hash", nil
						}

						disconnectedNodeRPCClient.GetRawMempoolFunc = func(
							ctx context.Context,
						) ([]string, error) {
							return []string{"Hash"}, nil
						}

						networkRPCClient := newChainReorgSuccessRPCClient(peerCount)

						networkRPCClient.GetRawMempoolFunc = func(
							ctx context.Context,
						) ([]string, error) {
							return nil, assert.AnError
						}

						return newRPCClientFactoryWithDetachedNode(
							networkRPCClient,
							"disconnected",
							disconnectedNodeRPCClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:             ctx,
					receiverAddress: "addr",
					amount:          1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(
						t,
						err,
						"ensure transaction not in any network node mempool",
					)
				},
				expectedHash: "",
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, crm := test.newChainReorgFunc(
					t,
					test.chainReorgArgs.dockerService,
					test.chainReorgArgs.rpcClientFactory,
					test.chainReorgArgs.nodes,
					test.chainReorgArgs.disconnectedNodeIndex,
				)

				hash, err := crm.SendTransactionOnDisconnectedNode(
					test.args.ctx,
					test.args.receiverAddress,
					test.args.amount,
				)
				test.assertErr(t, err)

				require.Equal(t, test.expectedHash, hash)
			})
		}
	})
	t.Run("MineBlocksOnNetwork", func(t *testing.T) {
		t.Parallel()

		type args struct {
			ctx       context.Context
			numBlocks int64
		}

		tests := map[string]struct {
			chainReorgArgs    chainReorgArgs
			args              args
			newChainReorgFunc newChainReorgFunc
			assertErr         require.ErrorAssertionFunc
			expectedHashes    []string
		}{
			"Success": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						disconnectedNodeClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "hash1", nil
						}

						networkClient := newChainReorgSuccessRPCClient(peerCount)

						networkClient.GenerateToAddressFunc = func(
							context.Context,
							int64,
							string,
						) ([]string, error) {
							return []string{"Hash"}, nil
						}

						networkClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "Hash", nil
						}

						return newRPCClientFactoryWithDetachedNode(
							networkClient,
							"disconnected",
							disconnectedNodeClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx: ctx,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr:         require.NoError,
				expectedHashes:    []string{"Hash"},
			},
			"MustDisconnectFirstError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: newPrivateNetworkStartSuccessRPCClientFactory(
						newChainReorgSuccessRPCClient(nil),
					),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:       ctx,
					numBlocks: 1,
				},
				newChainReorgFunc: newChainReorg,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, privatebtc.ErrChainReorgMustDisconnectNodeFirst)
				},
				expectedHashes: nil,
			},
			"GenerateToAddressError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						c.GenerateToAddressFunc = func(
							context.Context,
							int64,
							string,
						) ([]string, error) {
							return nil, assert.AnError
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:       ctx,
					numBlocks: 1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, assert.AnError)
				},
				expectedHashes: nil,
			},
			"SyncError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						c.GenerateToAddressFunc = func(
							context.Context,
							int64,
							string,
						) ([]string, error) {
							return []string{"Hash"}, nil
						}

						c.GetBestBlockHashFunc = func(ctx context.Context) (string, error) {
							return "", assert.AnError
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:       ctx,
					numBlocks: 1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "sync network nodes")
					require.ErrorIs(t, err, assert.AnError)
				},
				expectedHashes: nil,
			},
			"GetDisconnectedNodeBestBlockHashError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						disconnectedNodeClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "", assert.AnError
						}

						networkClient := newChainReorgSuccessRPCClient(peerCount)

						networkClient.GenerateToAddressFunc = func(
							context.Context,
							int64,
							string,
						) ([]string, error) {
							return []string{"Hash"}, nil
						}

						networkClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "Hash", nil
						}

						return newRPCClientFactoryWithDetachedNode(
							networkClient,
							"disconnected",
							disconnectedNodeClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx: ctx,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "get disconnected node best block Hash")
					require.ErrorIs(t, err, assert.AnError)
				},
				expectedHashes: nil,
			},
			"ChainsShouldNotBeSyncedError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						c.GenerateToAddressFunc = func(
							context.Context,
							int64,
							string,
						) ([]string, error) {
							return []string{"Hash"}, nil
						}

						c.GetBestBlockHashFunc = func(ctx context.Context) (string, error) {
							return "Hash", nil
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:       ctx,
					numBlocks: 1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, privatebtc.ErrChainsShouldNotBeSynced)
				},
				expectedHashes: nil,
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, crm := test.newChainReorgFunc(
					t,
					test.chainReorgArgs.dockerService,
					test.chainReorgArgs.rpcClientFactory,
					test.chainReorgArgs.nodes,
					test.chainReorgArgs.disconnectedNodeIndex,
				)

				hashes, err := crm.MineBlocksOnNetwork(test.args.ctx, test.args.numBlocks)
				test.assertErr(t, err)

				require.Equal(t, test.expectedHashes, hashes)
			})
		}
	})

	t.Run("MineBlocksOnDisconnectedNode", func(t *testing.T) {
		t.Parallel()

		type args struct {
			ctx       context.Context
			numBlocks int64
		}

		tests := map[string]struct {
			chainReorgArgs    chainReorgArgs
			args              args
			newChainReorgFunc newChainReorgFunc
			assertErr         require.ErrorAssertionFunc
			reconnect         bool
			expectedHashes    []string
		}{
			"Success": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						networkNodeClient := newChainReorgSuccessRPCClient(peerCount)

						networkNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "Hash", nil
						}

						disconnectedNodeClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeClient.GenerateToAddressFunc = func(
							context.Context,
							int64,
							string,
						) ([]string, error) {
							return []string{"Hash"}, nil
						}

						disconnectedNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "Hash", nil
						}

						return newRPCClientFactoryWithDetachedNode(
							networkNodeClient,
							"disconnected",
							disconnectedNodeClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx: ctx,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr:         require.NoError,
				expectedHashes:    []string{"Hash"},
			},
			"MustDisconnectFirstError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: newPrivateNetworkStartSuccessRPCClientFactory(
						newChainReorgSuccessRPCClient(nil),
					),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:       ctx,
					numBlocks: 1,
				},
				newChainReorgFunc: newChainReorg,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, privatebtc.ErrChainReorgMustDisconnectNodeFirst)
				},
				expectedHashes: nil,
			},
			"GenerateToAddressError": {
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						c := newChainReorgSuccessRPCClient(nil)

						c.GenerateToAddressFunc = func(
							context.Context,
							int64,
							string,
						) ([]string, error) {
							return nil, assert.AnError
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				args: args{
					ctx:       ctx,
					numBlocks: 1,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorIs(t, err, assert.AnError)
				},
				expectedHashes: nil,
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, crm := test.newChainReorgFunc(
					t,
					test.chainReorgArgs.dockerService,
					test.chainReorgArgs.rpcClientFactory,
					test.chainReorgArgs.nodes,
					test.chainReorgArgs.disconnectedNodeIndex,
				)

				hashes, err := crm.MineBlocksOnDisconnectedNode(test.args.ctx, test.args.numBlocks)
				test.assertErr(t, err)

				require.Equal(t, test.expectedHashes, hashes)
			})
		}
	})

	t.Run("ReconnectNode", func(t *testing.T) {
		t.Parallel()

		type args struct {
			ctx context.Context
		}

		tests := map[string]struct {
			chainReorgArgs    chainReorgArgs
			args              args
			newChainReorgFunc newChainReorgFunc
			assertErr         require.ErrorAssertionFunc
		}{
			"DisconnectedNodeGetBlockCountError": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						networkNodeClient := newChainReorgSuccessRPCClient(peerCount)

						networkNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 1, nil
						}

						disconnectedNodeClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 0, assert.AnError
						}

						return newRPCClientFactoryWithDetachedNode(
							networkNodeClient,
							"disconnected",
							disconnectedNodeClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "get disconnected node block count")
					require.ErrorIs(t, err, assert.AnError)
				},
			},
			"NetworkNodeGetBlockCountError": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						networkNodeClient := newChainReorgSuccessRPCClient(peerCount)

						networkNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 0, assert.AnError
						}

						disconnectedNodeClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 1, nil
						}

						return newRPCClientFactoryWithDetachedNode(
							networkNodeClient,
							"disconnected",
							disconnectedNodeClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "get network node block count")
					require.ErrorIs(t, err, assert.AnError)
				},
			},
			"ConnectToNetworkError": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						c := newChainReorgSuccessRPCClient(peerCount)

						c.GetBlockCountFunc = func(ctx context.Context) (int, error) {
							return 1, nil
						}

						f := c.AddPeerFunc

						c.AddPeerFunc = func(ctx context.Context, n privatebtc.Node) error {
							if callStackFunctionContains("ConnectToNetwork") {
								return assert.AnError
							}

							return f(ctx, n)
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "reconnect node")
					require.ErrorIs(t, err, assert.AnError)
				},
			},
			"GetDisconnectedNodeBestBlockHashError": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						networkNodeClient := newChainReorgSuccessRPCClient(peerCount)

						networkNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 1, nil
						}

						disconnectedNodeClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 2, nil
						}

						disconnectedNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "", assert.AnError
						}

						return newRPCClientFactoryWithDetachedNode(
							networkNodeClient,
							"disconnected",
							disconnectedNodeClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "get disconnected node best block Hash")
					require.ErrorIs(t, err, assert.AnError)
				},
			},
			"SyncNetworkNodesError": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						networkNodeClient := newChainReorgSuccessRPCClient(peerCount)

						networkNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 1, nil
						}

						networkNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "", assert.AnError
						}

						disconnectedNodeClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 2, nil
						}

						disconnectedNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "Hash", nil
						}

						return newRPCClientFactoryWithDetachedNode(
							networkNodeClient,
							"disconnected",
							disconnectedNodeClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "sync network nodes")
					require.ErrorIs(t, err, assert.AnError)
				},
			},

			"GetNetworkNodeBestBlockHashError": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						networkNodeClient := newChainReorgSuccessRPCClient(peerCount)

						networkNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 2, nil
						}

						networkNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "", assert.AnError
						}

						disconnectedNodeClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 1, nil
						}

						return newRPCClientFactoryWithDetachedNode(
							networkNodeClient,
							"disconnected",
							disconnectedNodeClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "get network best block Hash")
					require.ErrorIs(t, err, assert.AnError)
				},
			},
			"SyncDisconnectedNodeError": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						networkNodeClient := newChainReorgSuccessRPCClient(peerCount)

						networkNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 2, nil
						}

						networkNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "Hash", nil
						}

						disconnectedNodeClient := newChainReorgSuccessRPCClient(peerCount)

						disconnectedNodeClient.GetBlockCountFunc = func(
							ctx context.Context,
						) (int, error) {
							return 1, nil
						}

						disconnectedNodeClient.GetBestBlockHashFunc = func(
							ctx context.Context,
						) (string, error) {
							return "", assert.AnError
						}

						return newRPCClientFactoryWithDetachedNode(
							networkNodeClient,
							"disconnected",
							disconnectedNodeClient,
						)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr: func(t require.TestingT, err error, i ...any) {
					require.ErrorContains(t, err, "sync disconnected node")
					require.ErrorIs(t, err, assert.AnError)
				},
			},
			"Success": {
				args: args{
					ctx: ctx,
				},
				chainReorgArgs: chainReorgArgs{
					dockerService: newPrivateNetworkStartSuccessDockerService(
						newPrivateNetworkStartSuccessContainerWithPort("disconnected"),
						newPrivateNetworkStartSuccessNodeHandler(),
					),
					rpcClientFactory: func() *mock.RPCClientFactory {
						peerCount := new(atomic.Int64)

						c := newChainReorgSuccessRPCClient(peerCount)

						c.GetBlockCountFunc = func(ctx context.Context) (int, error) {
							return 1, nil
						}

						return newPrivateNetworkStartSuccessRPCClientFactory(c)
					}(),
					nodes:                 2,
					disconnectedNodeIndex: 0,
				},
				newChainReorgFunc: newChainReorgWithDisconnect,
				assertErr:         require.NoError,
			},
		}

		for name, test := range tests {
			test := test

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, crm := test.newChainReorgFunc(
					t,
					test.chainReorgArgs.dockerService,
					test.chainReorgArgs.rpcClientFactory,
					test.chainReorgArgs.nodes,
					test.chainReorgArgs.disconnectedNodeIndex,
				)

				err := crm.ReconnectNode(test.args.ctx)
				test.assertErr(t, err)
			})
		}
	})
}
