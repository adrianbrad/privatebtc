package tview

import (
	"fmt"

	"github.com/adrianbrad/privatebtc"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

type data struct {
	btcpn        *privatebtc.PrivateNetwork
	nodesDetails []*nodeDetails
	mempool      map[string]*mempoolTxDetails
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

func (d *data) update() error {
	eg := new(errgroup.Group)

	maps.Clear(d.mempool)

	for nodeID := range d.btcpn.Nodes() {
		nodeID := nodeID

		eg.Go(func() error {
			if err := d.updateNodeData(nodeID); err != nil {
				return fmt.Errorf("update node %d data: %w", nodeID, err)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	for _, details := range d.nodesDetails {
		for _, txHash := range details.mempoolTxs {
			tx, ok := d.mempool[txHash]
			if !ok {
				outputs, err := getTransactionOutputs(
					d.btcpn.Nodes()[details.id].RPCClient(),
					txHash,
				)
				if err != nil {
					return fmt.Errorf("get transaction outputs: %w", err)
				}

				d.mempool[txHash] = &mempoolTxDetails{
					hash: txHash,
					nodes: map[int]struct{}{
						details.id: {},
					},
					outputs: outputs,
				}

				continue
			}

			tx.nodes[details.id] = struct{}{}
		}
	}

	return nil
}

func (d *data) updateNodeData(nodeID int) error {
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

func getTransactionOutputs(
	client privatebtc.RPCClient,
	txHash string,
) (mempoolTxDetailsOutputs, error) {
	tx, err := client.GetTransaction(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}

	outputs := make(mempoolTxDetailsOutputs, len(tx.Vout))

	for _, v := range tx.Vout {
		outputs[v.ScriptPubKey.Address] = v.Value
	}

	return outputs, nil
}
