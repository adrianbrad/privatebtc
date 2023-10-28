package privatebtc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
)

// Chain Reorganisation
// 1. Disconnect a node from the network.
// 2. Send a transaction on the network and ensure transaction is in all mempools
// except the disconnected node.
// 3. Send a transaction on the disconnected node and ensure transaction is in
// the mempool of the disconnected node and not in the other nodes mempool.
// 4a. Mine a block on the network and ensure the transaction is confirmed.
// 4b. Mine a block on the disconnected node and ensure the transaction is confirmed.
// 5. Reconnect the disconnected node to the network.
// 6. At this point both chains have the same length but different best block.
// 7. Mine a block either on the network or the disconnected node to resolve the reorg,
// depending on which chain the block was added to, that chain will be considered the canonic chain
// and the other chain will be considered the orphaned chain with the corresponding transactions
// invalidated and sent back to the mempool.

// ChainReorgManager defines methods for handling a chain reorg.
// nolint: revive
type ChainReorgManager interface {
	DisconnectNode(ctx context.Context) (Node, error)
	SendTransactionOnNetwork(
		ctx context.Context,
		receiverAddress string,
		amount float64,
	) (string, error)
	SendTransactionOnDisconnectedNode(ctx context.Context, receiverAddress string, amount float64) (string, error)
	MineBlocksOnNetwork(ctx context.Context, numBlocks int64) ([]string, error)
	MineBlocksOnDisconnectedNode(ctx context.Context, numBlocks int64) ([]string, error)
	ReconnectNode(ctx context.Context) error
}

var _ ChainReorgManager = (*ChainReorg)(nil)

// ChainReorg represents a chain reorg manager.
type ChainReorg struct {
	disconnectedNode Node
	networkNodes     Nodes
	logger           *slog.Logger

	disconnected bool
}

// NewChainReorg creates a new chain reorg manager.
func (n *PrivateNetwork) NewChainReorg(
	disconnectedNodeIndex int,
) (*ChainReorg, error) {
	if disconnectedNodeIndex < 0 || disconnectedNodeIndex >= len(n.nodes) {
		return nil, fmt.Errorf("index %d: %w", disconnectedNodeIndex, ErrNodeIndexOutOfRange)
	}

	return &ChainReorg{
		disconnectedNode: n.nodes[disconnectedNodeIndex],
		networkNodes:     slices.Delete(n.Nodes(), disconnectedNodeIndex, disconnectedNodeIndex+1),
		logger:           n.logger,
	}, nil
}

// NewChainReorgWithAssertion creates a new chain reorg manager with assertion.
func (n *PrivateNetwork) NewChainReorgWithAssertion(
	disconnectedNodeIndex int,
) (*ChainReorgWithAssertion, error) {
	cr, err := n.NewChainReorg(disconnectedNodeIndex)
	if err != nil {
		return nil, fmt.Errorf("new chain reorg: %w", err)
	}

	return &ChainReorgWithAssertion{ChainReorg: cr}, nil
}

// DisconnectNode disconnects a node from the network.
func (c *ChainReorg) DisconnectNode(ctx context.Context) (Node, error) {
	c.logger.Info(
		"🔌⌛ Disconnecting node from network",
		"disconnected_node_id",
		c.disconnectedNode.Name(),
	)

	if err := c.disconnectedNode.DisconnectFromNetwork(ctx); err != nil {
		return Node{}, fmt.Errorf("disconnect from network: %w", err)
	}

	c.logger.Info(
		"🔌✅ Successfully disconnected node from network",
		"disconnected_node_id",
		c.disconnectedNode.Name(),
	)

	c.disconnected = true

	return c.disconnectedNode, nil
}

// SendTransactionOnNetwork sends a transaction on the network.
func (c *ChainReorg) SendTransactionOnNetwork(
	ctx context.Context,
	receiverAddress string,
	amount float64,
) (string, error) {
	if !c.disconnected {
		return "", ErrChainReorgMustDisconnectNodeFirst
	}

	c.logger.Info(
		"⬆️⬆️⌛ Sending transaction on network",
		"sender_node_id",
		c.networkNodes[0].Name(),
		"receiver_address",
		receiverAddress,
		"amount",
		amount,
	)

	hash, err := c.networkNodes[0].RPCClient().SendToAddress(ctx, receiverAddress, amount)
	if err != nil {
		return "", fmt.Errorf("send to address: %w", err)
	}

	c.logger.Info(
		"⬆️⬆️✅ Successfully sent transaction on network",
		"sender_node_id",
		c.networkNodes[0].Name(),
		"tx_hash",
		hash,
	)

	return hash, nil
}

// SendTransactionOnDisconnectedNode sends a transaction on the disconnected node.
func (c *ChainReorg) SendTransactionOnDisconnectedNode(
	ctx context.Context,
	receiverAddress string,
	amount float64,
) (string, error) {
	if !c.disconnected {
		return "", ErrChainReorgMustDisconnectNodeFirst
	}

	c.logger.Info(
		"⬆️⌛ Sending transaction on disconnected node",
		"sender_node_id",
		c.disconnectedNode.Name(),
		"receiver_address",
		receiverAddress,
		"amount",
		amount,
	)

	hash, err := c.disconnectedNode.RPCClient().SendToAddress(ctx, receiverAddress, amount)
	if err != nil {
		return "", fmt.Errorf("send to address: %w", err)
	}

	c.logger.Info(
		"⬆️✅ Successfully sent transaction on disconnected node",
		"sender_node_id",
		c.disconnectedNode.Name(),
		"tx_hash",
		hash,
	)

	return hash, nil
}

// MineBlocksOnNetwork mines blocks on the network.
func (c *ChainReorg) MineBlocksOnNetwork(ctx context.Context, numBlocks int64) ([]string, error) {
	if !c.disconnected {
		return nil, ErrChainReorgMustDisconnectNodeFirst
	}

	const (
		burningAddr = "bcrt1qzlfc3dw3ecjncvkwmwpvs84ejqzp4fr4agghm8"
	)

	c.logger.Info(
		"⏹️⏹️⌛ Mine blocks on network",
		"miner_node_id",
		c.networkNodes[0].Name(),
		"num_blocks",
		numBlocks,
	)

	blockHashes, err := c.networkNodes[0].RPCClient().GenerateToAddress(ctx, numBlocks, burningAddr)
	if err != nil {
		return nil, fmt.Errorf("generate to address: %w", err)
	}

	c.logger.Info(
		"⏹️⏹️✅ Successfully mined blocks on network",
		"miner_node_id",
		c.networkNodes[0].Name(),
		"num_blocks",
		numBlocks,
		"block_hashes",
		blockHashes,
	)

	return blockHashes, nil
}

// MineBlocksOnDisconnectedNode mines blocks on the disconnected node.
func (c *ChainReorg) MineBlocksOnDisconnectedNode(
	ctx context.Context,
	numBlocks int64,
) ([]string, error) {
	if !c.disconnected {
		return nil, ErrChainReorgMustDisconnectNodeFirst
	}

	const (
		burningAddr = "bcrt1qzlfc3dw3ecjncvkwmwpvs84ejqzp4fr4agghm8"
	)

	c.logger.Info(
		"⏹️⌛ Mine blocks on disconnected node",
		"miner_node_id",
		c.disconnectedNode.Name(),
		"num_blocks",
		numBlocks,
	)

	blockHashes, err := c.disconnectedNode.RPCClient().GenerateToAddress(
		ctx,
		numBlocks,
		burningAddr,
	)
	if err != nil {
		return nil, err
	}

	c.logger.Info(
		"⏹️✅ Successfully mined blocks on disconnected node",
		"miner_node_id",
		c.disconnectedNode.Name(),
		"num_blocks",
		numBlocks,
		"block_hashes",
		blockHashes,
	)

	return blockHashes, nil
}

// ReconnectNode reconnects the disconnected node to the network.
func (c *ChainReorg) ReconnectNode(ctx context.Context) error {
	if !c.disconnected {
		return ErrChainReorgMustDisconnectNodeFirst
	}

	c.logger.Info(
		"🔁⌛ Reconnect disconnected node",
		"disconnected_node_id",
		c.disconnectedNode.Name(),
	)

	if err := c.disconnectedNode.ConnectToNetwork(ctx); err != nil {
		return fmt.Errorf("connect to network: %w", err)
	}

	c.logger.Info(
		"🔁✅ Successfully reconnected node",
		"disconnected_node_id",
		c.disconnectedNode.Name(),
	)

	return nil
}

var _ ChainReorgManager = (*ChainReorgWithAssertion)(nil)

// ChainReorgWithAssertion represents a chain reorg manager with assertion.
type ChainReorgWithAssertion struct {
	*ChainReorg
}

// DisconnectNode disconnects a node from the network.
// It is expected that the node will have 0 peers after disconnecting.
func (c *ChainReorgWithAssertion) DisconnectNode(ctx context.Context) (Node, error) {
	disconnectedNode, err := c.ChainReorg.DisconnectNode(ctx)
	if err != nil {
		return Node{}, fmt.Errorf("disconnect node: %w", err)
	}

	const tickEvery = 100 * time.Millisecond

	t := time.NewTicker(tickEvery)
	defer t.Stop()

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var lastConnectionCount int

loop:
	for {
		select {
		case <-t.C:
			lastConnectionCount, err = disconnectedNode.RPCClient().GetConnectionCount(ctx)
			if err != nil {
				return Node{}, fmt.Errorf("get disconnected node connection count: %w", err)
			}

			if lastConnectionCount == 0 {
				break loop
			}

		case <-ctxTimeout.Done():
			return Node{}, fmt.Errorf("disconnected node: %w", &UnexpectedPeerCountError{
				nodeName: disconnectedNode.Name(),
				expected: 0,
				got:      lastConnectionCount,
			})
		}
	}

	eg, egCtx := errgroup.WithContext(ctx)

	expectedConnectionCount := len(c.networkNodes) - 1

	for _, n := range c.networkNodes {
		n := n

		eg.Go(func() error {
			cc, err := n.RPCClient().GetConnectionCount(egCtx)
			if err != nil {
				return fmt.Errorf("get connection count for node %s: %w", n.Name(), err)
			}

			if cc != expectedConnectionCount {
				return fmt.Errorf(
					"assert peer connection count: %w",
					&UnexpectedPeerCountError{
						nodeName: n.Name(),
						expected: expectedConnectionCount,
						got:      cc,
					},
				)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return Node{}, err
	}

	return disconnectedNode, nil
}

// SendTransactionOnNetwork sends a transaction on the network.
// It is expected that the transaction is in every node mempool except the disconnected node.
func (c *ChainReorgWithAssertion) SendTransactionOnNetwork(
	ctx context.Context,
	receiverAddress string,
	amount float64,
) (string, error) {
	hash, err := c.ChainReorg.SendTransactionOnNetwork(ctx, receiverAddress, amount)
	if err != nil {
		return "", fmt.Errorf("send transaction on network: %w", err)
	}

	if err := c.networkNodes.EnsureTransactionInEveryMempool(ctx, hash); err != nil {
		return "", fmt.Errorf("ensure transaction in every network node mempool: %w", err)
	}

	ok, err := c.disconnectedNode.IsTransactionInMempool(ctx, hash)
	if err != nil {
		return "", fmt.Errorf("is transaction in disconnected node mempool: %w", err)
	}

	if ok {
		return "", fmt.Errorf("transaction %s: %w", hash, ErrTxFoundInMempool)
	}

	return hash, nil
}

// SendTransactionOnDisconnectedNode sends a transaction on the disconnected node.
// It is expected that the transaction is not in any network node mempool.
func (c *ChainReorgWithAssertion) SendTransactionOnDisconnectedNode(
	ctx context.Context,
	receiverAddress string,
	amount float64,
) (string, error) {
	hash, err := c.ChainReorg.SendTransactionOnDisconnectedNode(ctx, receiverAddress, amount)
	if err != nil {
		return "", fmt.Errorf("send transaction on network: %w", err)
	}

	ok, err := c.disconnectedNode.IsTransactionInMempool(ctx, hash)
	if err != nil {
		return "", fmt.Errorf("is transaction in disconnected node mempool: %w", err)
	}

	if !ok {
		return "", fmt.Errorf("transaction %s: %w", hash, ErrTxFoundInMempool)
	}

	if err := c.networkNodes.EnsureTransactionNotInAnyMempool(ctx, hash); err != nil {
		return "", fmt.Errorf("ensure transaction not in any network node mempool: %w", err)
	}

	return hash, nil
}

// MineBlocksOnNetwork mines blocks on the network.
// It is expected that the disconnected node will not see the new blocks.
func (c *ChainReorgWithAssertion) MineBlocksOnNetwork(
	ctx context.Context,
	numBlocks int64,
) ([]string, error) {
	blockHashes, err := c.ChainReorg.MineBlocksOnNetwork(ctx, numBlocks)
	if err != nil {
		return nil, fmt.Errorf("mine blocks on network: %w", err)
	}

	bestBlockHash := blockHashes[len(blockHashes)-1]

	const hardTimeout = 5 * time.Second

	ctxTimeout, cancel := context.WithTimeout(ctx, hardTimeout)
	defer cancel()

	if err := c.networkNodes.Sync(ctxTimeout, bestBlockHash); err != nil {
		return nil, fmt.Errorf("sync network nodes: %w", err)
	}

	discBestBlockHash, err := c.ChainReorg.disconnectedNode.RPCClient().GetBestBlockHash(ctx)
	if err != nil {
		return nil, fmt.Errorf("get disconnected node best block hash: %w", err)
	}

	if bestBlockHash == discBestBlockHash {
		return nil, ErrChainsShouldNotBeSynced
	}

	return blockHashes, nil
}

// MineBlocksOnDisconnectedNode mines blocks on the disconnected node.
// It is expected that the network nodes do not will not see the new blocks.
func (c *ChainReorgWithAssertion) MineBlocksOnDisconnectedNode(
	ctx context.Context,
	numBlocks int64,
) ([]string, error) {
	blockHashes, err := c.ChainReorg.MineBlocksOnDisconnectedNode(ctx, numBlocks)
	if err != nil {
		return nil, fmt.Errorf("mine blocks on disconnected node: %w", err)
	}

	bestBlockHash := blockHashes[len(blockHashes)-1]

	if !c.disconnected {
		if err := c.networkNodes.Sync(ctx, bestBlockHash); err != nil {
			return nil, fmt.Errorf("sync network nodes to %s: %w", bestBlockHash, err)
		}
	}

	return blockHashes, nil
}

// ReconnectNode reconnects the disconnected node to the network.
// It is expected that the disconnected node will reorg to the network chain.
func (c *ChainReorgWithAssertion) ReconnectNode(ctx context.Context) error {
	var (
		// disconnected block count
		dcBc int
		// network block count
		nBc int
	)

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error

		dcBc, err = c.disconnectedNode.RPCClient().GetBlockCount(egCtx)
		if err != nil {
			return fmt.Errorf("get disconnected node block count: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		var err error

		nBc, err = c.networkNodes[0].RPCClient().GetBlockCount(egCtx)
		if err != nil {
			return fmt.Errorf("get network node block count: %w", err)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	if err := c.ChainReorg.ReconnectNode(ctx); err != nil {
		return fmt.Errorf("reconnect node: %w", err)
	}

	if syncNetwork := dcBc > nBc; syncNetwork {
		blockHash, err := c.disconnectedNode.RPCClient().GetBestBlockHash(ctx)
		if err != nil {
			return fmt.Errorf("get disconnected node best block hash: %w", err)
		}

		const timeout = 5 * time.Second

		ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if err := c.networkNodes.Sync(ctxTimeout, blockHash); err != nil {
			return fmt.Errorf("sync network nodes: %w", err)
		}
	}

	if syncDisconnected := nBc > dcBc; syncDisconnected {
		blockHash, err := c.networkNodes[0].RPCClient().GetBestBlockHash(ctx)
		if err != nil {
			return fmt.Errorf("get network best block hash: %w", err)
		}

		const timeout = 5 * time.Second

		ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if err := (Nodes{c.disconnectedNode}.Sync(ctxTimeout, blockHash)); err != nil {
			return fmt.Errorf("sync disconnected node: %w", err)
		}
	}

	return nil
}
