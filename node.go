package privatebtc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/avast/retry-go"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

// Node represents a node in the private network.
// It contains an RPC client and a Node Handler.
type Node struct {
	id          int
	name        string
	rpcClient   RPCClient
	nodeHandler NodeHandler

	pn *PrivateNetwork
}

// ID returns the ID of the node.
func (n Node) ID() int {
	return n.id
}

// Name returns the Name of the node.
func (n Node) Name() string {
	return n.name
}

// NodeHandler returns the nodeHandler of the node.
func (n Node) NodeHandler() NodeHandler {
	return n.nodeHandler
}

// RPCClient returns the RPC client of the node.
func (n Node) RPCClient() RPCClient {
	return n.rpcClient
}

// Fund is a helper function for funding a node.
// It generates a new address to the block and mines 101 blocks to it,
// returning the Hash of the last block that was mined
// This method will fund the node wallet with 50 BTC.
func (n Node) Fund(ctx context.Context) (string, error) {
	addr, err := n.RPCClient().GetNewAddress(ctx, "fund")
	if err != nil {
		return "", fmt.Errorf("get new address: %w", err)
	}

	const numBlocks = 101

	hashes, err := n.RPCClient().GenerateToAddress(ctx, numBlocks, addr)
	if err != nil {
		return "", fmt.Errorf("generate to address: %w", err)
	}

	return hashes[len(hashes)-1], nil
}

// DisconnectFromNetwork disconnects the node from the network (the other nodes).
func (n Node) DisconnectFromNetwork(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	for i := range n.pn.nodes {
		node := n.pn.nodes[i]

		if node.id == n.id {
			continue
		}

		eg.Go(func() error {
			if err := node.RPCClient().RemovePeer(egCtx, n); err != nil {
				return fmt.Errorf("remove peer %d from node %d: %w", n.id, node.id, err)
			}

			return nil
		})
	}

	return eg.Wait()
}

// ConnectToNetwork connects the node to the network (the other nodes).
func (n Node) ConnectToNetwork(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	for i, node := range n.pn.nodes {
		if node.id == n.id {
			continue
		}

		node := node
		i := i

		eg.Go(func() error {
			err := node.RPCClient().AddPeer(egCtx, n)
			if err != nil {
				return fmt.Errorf("add node %d: %w", i, err)
			}

			return nil
		})
	}

	return eg.Wait()
}

// IsTransactionInMempool checks if a transaction is in the mempool of the node.
func (n Node) IsTransactionInMempool(ctx context.Context, txHash string) (bool, error) {
	mempool, err := n.RPCClient().GetRawMempool(ctx)
	if err != nil {
		return false, fmt.Errorf("get raw mempool: %w", err)
	}

	for _, mempoolTxHash := range mempool {
		if strings.EqualFold(mempoolTxHash, txHash) {
			return true, nil
		}
	}

	return false, nil
}

// NodeHandler represents manager for a bitcoin node.
type NodeHandler interface {
	io.Closer
	InternalIP() string
	HostRPCPort() string
	Name() string
}

// Nodes is a slice of nodes.
type Nodes []Node

// Sync waits until all nodes are on the same block height.
// nolint: gocognit
func (nodes Nodes) Sync(ctx context.Context, toBlockHash string) error {
	const tickEvery = 10 * time.Millisecond

	ticker := time.NewTicker(tickEvery)
	defer ticker.Stop()

	for {
		var cont atomic.Bool

		select {
		case <-ctx.Done():
			return ErrTimeoutAndChainsAreNotSynced

		case <-ticker.C:
			eg, egCtx := errgroup.WithContext(ctx)

			for i := range nodes {
				i := i

				eg.Go(func() error {
					blockHash, err := nodes[i].RPCClient().GetBestBlockHash(egCtx)
					if err != nil {
						return fmt.Errorf("get best block Hash for node %d: %w", i, err)
					}

					if blockHash != toBlockHash {
						cont.Store(true)
					}

					return nil
				})
			}

			if err := eg.Wait(); err != nil {
				return err
			}

			if synced := !cont.Load(); synced {
				return nil
			}
		}
	}
}

// EnsureTransactionInEveryMempool ensures that a transaction is in the mempool of every node.
// nolint: gocognit
func (nodes Nodes) EnsureTransactionInEveryMempool(
	ctx context.Context,
	txHash string,
) error {
	const attempts = 10

	return retry.Do(func() error {
		eg, egCtx := errgroup.WithContext(ctx)

		for i, node := range nodes {
			i, node := i, node

			eg.Go(func() error {
				ok, err := node.IsTransactionInMempool(egCtx, txHash)
				if err != nil {
					return fmt.Errorf(
						"is tx in node %d mempool: %w",
						i,
						err,
					)
				}

				if !ok {
					var errs error

					rawMempool, err := node.RPCClient().GetRawMempool(ctx)
					if err != nil {
						errs = errors.Join(errs, fmt.Errorf("get raw mempool: %w", err))
					}

					connCount, err := node.RPCClient().GetConnectionCount(ctx)
					if err != nil {
						errs = errors.Join(errs, fmt.Errorf("get connection count: %w", err))
					}

					tx, err := node.RPCClient().GetTransaction(ctx, txHash)
					if err != nil {
						errs = errors.Join(errs, fmt.Errorf("get raw transaction: %w", err))
					}

					return errors.Join(errs, fmt.Errorf(
						"get tx %q for node %d, "+
							"raw mempool: %q, "+
							"connection count: %d, "+
							"raw tx %+v: %w",
						txHash,
						i,
						rawMempool,
						connCount,
						tx,
						ErrTxNotFoundInMempool,
					))
				}

				return nil
			})
		}

		return eg.Wait()
	}, retry.Context(ctx), retry.Attempts(attempts), retry.MaxDelay(time.Second))
}

// EnsureTransactionNotInAnyMempool ensures that transaction is not in any mempool of the nodes.
func (nodes Nodes) EnsureTransactionNotInAnyMempool(ctx context.Context, txHash string) error {
	eg, egCtx := errgroup.WithContext(ctx)

	for i := range nodes {
		i := i
		node := nodes[i]

		eg.Go(func() error {
			ok, err := node.IsTransactionInMempool(egCtx, txHash)
			if err != nil {
				return fmt.Errorf("is tx in node %d mempool: %w", i, err)
			}

			if ok {
				return fmt.Errorf("node %d mempool: %w", i, ErrTxFoundInMempool)
			}

			return nil
		})
	}

	return eg.Wait()
}

// NetworkMempoolTransaction represents a transaction in the network mempool.
type NetworkMempoolTransaction struct {
	MempoolTransaction
	Nodes []int
}

// NetworkMempool represents the network mempool.
type NetworkMempool map[string]*NetworkMempoolTransaction

// Hashes returns the hashes of the network mempool transactions.
func (m NetworkMempool) Hashes() []string {
	hashes := maps.Keys(m)

	slices.Sort(hashes)

	return hashes
}

// NetworkMempool returns all the transactions aggregated from the nodes mempools.
func (nodes Nodes) NetworkMempool(ctx context.Context) (NetworkMempool, error) {
	mempoolTransactions := NetworkMempool{}

	var mutex sync.Mutex

	eg, egCtx := errgroup.WithContext(ctx)

	for i := range nodes {
		i := i

		eg.Go(func() error {
			nodeMempoolTxs, err := nodes[i].RPCClient().GetRawMempool(egCtx)
			if err != nil {
				return fmt.Errorf("get mempool for node %d: %w", i, err)
			}

			for _, txHash := range nodeMempoolTxs {
				mutex.Lock()

				if mpTx, exists := mempoolTransactions[txHash]; exists {
					mpTx.Nodes = append(mpTx.Nodes, i)

					mutex.Unlock()

					continue
				}

				txOutputs, err := nodes[i].RPCClient().GetTransactionOutputs(egCtx, txHash)
				if err != nil {
					mutex.Unlock()

					return fmt.Errorf("get transaction outputs: %w", err)
				}

				mempoolTransactions[txHash] = &NetworkMempoolTransaction{
					MempoolTransaction: MempoolTransaction{
						Hash:    txHash,
						Outputs: txOutputs,
					},
					Nodes: []int{i},
				}

				mutex.Unlock()
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return NetworkMempool{}, fmt.Errorf("fetch nodes mempools: %w", err)
	}

	for i := range mempoolTransactions {
		slices.Sort(mempoolTransactions[i].Nodes)
	}

	return mempoolTransactions, nil
}

// Balance is a balance of a node.
type Balance struct {
	Trusted  float64
	Pending  float64
	Immature float64
}

func connectNodes(ctx context.Context, nodes []Node) error {
	connectionCount, err := nodes[0].RPCClient().GetConnectionCount(ctx)
	if err != nil {
		return fmt.Errorf("get connection count for first node: %w", err)
	}

	if connectionCount != 0 {
		return &peerCountShouldBeZeroError{got: connectionCount}
	}

	eg, egCtx := errgroup.WithContext(ctx)

	for nodeIdx := range nodes {
		for nextNodeIdx := nodeIdx + 1; nextNodeIdx < len(nodes); nextNodeIdx++ {
			nextNode := nodes[nextNodeIdx]
			node := nodes[nodeIdx]

			eg.Go(func() error {
				if err := node.RPCClient().AddPeer(egCtx, nextNode); err != nil {
					return fmt.Errorf("add node %d: %w", nextNode.id, err)
				}

				return nil
			})
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	eg, egCtx = errgroup.WithContext(ctx)

	for i := range nodes {
		i := i

		eg.Go(func() error {
			const attempts = 5

			return retry.Do(func() error {
				peerCount, err := nodes[i].RPCClient().GetConnectionCount(egCtx)
				if err != nil {
					return fmt.Errorf(
						"get connection count for node %d: %w",
						i,
						err,
					)
				}

				if peerCount != len(nodes)-1 {
					return &UnexpectedPeerCountError{
						nodeName: nodes[i].Name(),
						expected: len(nodes) - 1,
						got:      peerCount,
					}
				}

				return nil
			}, retry.Context(ctx), retry.Attempts(attempts))
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("eg wait: %w", err)
	}

	return nil
}
