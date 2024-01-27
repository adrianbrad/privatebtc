package tview

import (
	"context"
	"fmt"

	"github.com/adrianbrad/privatebtc"
	"golang.org/x/sync/errgroup"
)

type data struct {
	btcpn          *privatebtc.PrivateNetwork
	nodesDetails   []*nodeDetails
	networkMempool privatebtc.NetworkMempool
}

const burnAddress = "bcrt1qzlfc3dw3ecjncvkwmwpvs84ejqzp4fr4agghm8"

func (d *data) toFormAddresses() []string {
	addrs := []string{"burn address"}

	for _, n := range d.nodesDetails {
		for _, a := range n.addresses {
			addrs = append(addrs, fmt.Sprintf("%d:%s", n.id, a))
		}
	}

	return addrs
}

func (d *data) update(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	for nodeID := range d.btcpn.Nodes() {
		nodeID := nodeID

		eg.Go(func() error {
			if err := d.updateNodeData(egCtx, nodeID); err != nil {
				return fmt.Errorf("update node %d data: %w", nodeID, err)
			}

			return nil
		})
	}

	eg.Go(func() error {
		mp, err := d.btcpn.Nodes().NetworkMempool(ctx)
		if err != nil {
			return fmt.Errorf("network mempool: %w", err)
		}

		d.networkMempool = mp

		return nil
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("eg wait: %w", err)
	}

	return nil
}

func (d *data) updateNodeData(ctx context.Context, nodeID int) error {
	eg, egCtx := errgroup.WithContext(ctx)

	node := d.btcpn.Nodes()[nodeID]

	eg.Go(func() error {
		hashes, err := node.RPCClient().GetRawMempool(egCtx)
		if err != nil {
			return fmt.Errorf("get raw mempool: %w", err)
		}

		d.nodesDetails[nodeID].mempoolTxs = hashes

		return nil
	})

	eg.Go(func() error {
		bal, err := node.RPCClient().GetBalance(egCtx)
		if err != nil {
			return fmt.Errorf("get balance: %w", err)
		}

		d.nodesDetails[nodeID].balance = bal

		return nil
	})

	eg.Go(func() error {
		blockCount, err := node.RPCClient().GetBlockCount(egCtx)
		if err != nil {
			return fmt.Errorf("get block count: %w", err)
		}

		d.nodesDetails[nodeID].blockCount = blockCount

		return nil
	})

	eg.Go(func() error {
		addresses, err := node.RPCClient().ListAddresses(egCtx)
		if err != nil {
			return fmt.Errorf("list addresses: %w", err)
		}

		d.nodesDetails[nodeID].addresses = addresses

		return nil
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("eg wait: %w", err)
	}

	return nil
}
