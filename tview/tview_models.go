package tview

import (
	"fmt"
	"sort"
	"strings"

	"github.com/adrianbrad/privatebtc"
	"golang.org/x/exp/maps"
)

type mempoolTxDetailsOutputs map[string]float64

func (s mempoolTxDetailsOutputs) String() string {
	ks := maps.Keys(s)

	sort.Strings(ks)

	var b strings.Builder

	b.WriteString("\n")

	for _, k := range ks {
		b.WriteString(fmt.Sprintf("%s: %f\n", k, s[k]))
	}

	return b.String()
}

type mempoolTxDetails struct {
	hash    string
	nodes   map[int]struct{}
	outputs mempoolTxDetailsOutputs
}

func (m mempoolTxDetails) String() string {
	nodes := maps.Keys(m.nodes)

	sort.Ints(nodes)

	return fmt.Sprintf(
		"Hash: %s\nNodes: %v\nOutputs: %v",
		m.hash,
		nodes,
		m.outputs,
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
