package privatebtc

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Transaction represents a BTC transaction.
type Transaction struct {
	TxID      string
	Hash      string
	BlockHash string
	Vout      []TransactionVout
	Vin       []TransactionVin
}

// GetTransactionFee returns the transaction fee.
func (tx Transaction) GetTransactionFee(totalInputs float64) float64 {
	var totalOutputs float64

	for _, vout := range tx.Vout {
		totalOutputs += vout.Value
	}

	return totalInputs - totalOutputs
}

// TotalInputsValue sums up the value of all inputs in the transaction by checking the value of
// the outputs that they spend from.
func (tx Transaction) TotalInputsValue(ctx context.Context, client RPCClient) (float64, error) {
	var totalInputs float64

	// sync
	var (
		inputsMutex sync.Mutex
		eg, egCtx   = errgroup.WithContext(ctx)
	)

	for _, vin := range tx.Vin {
		vin := vin

		eg.Go(func() error {
			vinTx, err := client.GetTransaction(egCtx, vin.TxID)
			if err != nil {
				return fmt.Errorf("get transaction %s: %w", vin.TxID, err)
			}

			for _, vout := range vinTx.Vout {
				if vout.N == vin.Vout {
					inputsMutex.Lock()
					totalInputs += vout.Value
					inputsMutex.Unlock()
				}
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return 0, err
	}

	return totalInputs, nil
}

// TransactionVin represents a BTC transaction input.
type TransactionVin struct {
	TxID string
	Vout uint32
}

// TransactionVout represents a BTC transaction output.
type TransactionVout struct {
	Value        float64
	N            uint32
	ScriptPubKey struct {
		Address string
	}
}

// MempoolTransaction represents a BTC mempool transaction.
type MempoolTransaction struct {
	Hash    string
	Outputs []MempoolTransactionOutput
}

// MempoolTransactionOutput represents a BTC mempool transaction output.
type MempoolTransactionOutput struct {
	Address string
	Value   float64
}
