package tview

import (
	"github.com/rivo/tview"
)

type actionsFlex struct {
	*tview.Flex
}

func newActionsFlex(
	outputView *outputView,
	nodeActionsList *nodeActionsList,
) *actionsFlex {
	flex := tview.NewFlex()

	const borderWidth = 2

	flex.
		SetDirection(tview.FlexRow).
		AddItem(outputView, 0, 1, false).
		AddItem(nodeActionsList, nodeActionsList.GetItemCount()+borderWidth, 1, true)

	return &actionsFlex{
		Flex: flex,
	}
}

type mempoolFlex struct {
	*tview.Flex
}

func newMempoolFlex(
	mempoolTransactionsList *mempoolTransactionsList,
	mempoolTransactionDetails *mempoolTransactionDetails,
) *mempoolFlex {
	flex := tview.NewFlex()

	flex.
		AddItem(mempoolTransactionsList, 0, 1, true).
		AddItem(mempoolTransactionDetails, 0, 1, false).
		SetBorder(true).
		SetTitle("Mempool")

	return &mempoolFlex{
		Flex: flex,
	}
}

type nodesFlex struct {
	*tview.Flex
}

func newNodesFlex(
	nodesList *nodesList,
	nodeDetails *nodeDetailsView,
) *nodesFlex {
	flex := tview.NewFlex()

	const (
		nodesListProportion   = 1
		nodeDetailsProportion = 2 * nodesListProportion
	)

	flex.
		AddItem(nodesList, 0, nodesListProportion, true).
		AddItem(nodeDetails, 0, nodeDetailsProportion, false).
		SetBorder(true).
		SetTitle("Nodes")

	return &nodesFlex{
		Flex: flex,
	}
}

type appFlex struct {
	*tview.Flex
}

func newAppFlex(
	flex *tview.Flex,
	mempoolFlex *mempoolFlex,
	actionsFlex *actionsFlex,
	nodesFlex *nodesFlex,
	footer *tview.TextView,
) *appFlex {
	flex.
		AddItem(mempoolFlex, 0, 1, false).
		AddItem(actionsFlex, 0, 1, true).
		AddItem(nodesFlex, 0, 1, false)

	outerFlex := tview.NewFlex().
		AddItem(flex, 0, 1, true). // Add the inner flex to the outer flex
		AddItem(footer, 1, 0, false).
		SetDirection(tview.FlexRow)

	return &appFlex{
		Flex: outerFlex,
	}
}
