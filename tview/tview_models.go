package tview

import (
	"fmt"
	"strings"

	"github.com/adrianbrad/privatebtc"
)

type mempoolTxDetailsOutputs []privatebtc.MempoolTransactionOutput

func (s mempoolTxDetailsOutputs) String() string {
	var b strings.Builder

	b.WriteString("\n")

	for _, output := range s {
		b.WriteString(fmt.Sprintf("%s: %f\n", output.Address, output.Value))
	}

	return b.String()
}

type mempoolTxDetails privatebtc.NetworkMempoolTransaction

func (m mempoolTxDetails) String() string {
	return fmt.Sprintf(
		"Hash: %s\nNodes: %v\nOutputs: %v",
		m.Hash,
		m.Nodes,
		mempoolTxDetailsOutputs(m.Outputs),
	)
}

type nodeDetails struct {
	id         int
	blockCount int
	connected  bool
	balance    privatebtc.Balance
	addresses  []string
	mempoolTxs []string
}

func (n nodeDetails) String() string {
	return fmt.Sprintf(
		"ID: %d\nConnected: %t\n"+
			"Block Count: %d\n"+
			"Balance: T:[green]%.2f[-] P:%.2f I:%.2f \n"+
			"Addresses:\n%s\n"+
			"Mempool Transactions:\n%s",
		n.id,
		n.connected,
		n.blockCount,
		n.balance.Trusted, n.balance.Pending, n.balance.Immature,
		strings.Join(n.addresses, "\n"),
		strings.Join(n.mempoolTxs, "\n"),
	)
}
