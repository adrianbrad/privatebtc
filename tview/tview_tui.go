package tview

import (
	"fmt"
	"time"

	"github.com/adrianbrad/privatebtc"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TUI is a Terminal User Interface for a bitcoin private network.
// It starts a new TUI app and runs it.
type TUI struct {
	app       *app
	btcPrvNet *privatebtc.PrivateNetwork
}

const bitcoinMaxAddressLength = 44

// NewTUI creates a new Terminal User Interface for the given bitcoin private network.
func NewTUI(btcPrivateNetwork *privatebtc.PrivateNetwork, version string) *TUI {
	mempoolTxs := map[string]*mempoolTxDetails{}

	nodesDetails := make([]*nodeDetails, len(btcPrivateNetwork.Nodes()))

	for i := range btcPrivateNetwork.Nodes() {
		nodesDetails[i] = &nodeDetails{
			id:        btcPrivateNetwork.Nodes()[i].ID(),
			connected: true,
		}
	}

	data := &data{
		btcpn:        btcPrivateNetwork,
		nodesDetails: nodesDetails,
		mempool:      mempoolTxs,
	}

	// Mempool
	mempoolTransactionDetails := newMempoolTransactionDetails(data)
	mempoolTransactionsList := newMempoolTransactionsList(mempoolTransactionDetails, data)

	// Nodes
	nodeDetailsView := newNodeDetailsView(data)
	nodesList := newNodesList(btcPrivateNetwork.Nodes(), nodeDetailsView)

	// Shared widgets
	tviewAppFlex := tview.NewFlex()
	tviewAppPages := tview.NewPages()

	// Actions
	outputView := newOutputView()
	actionsHandler := newActionsHandler(data, btcPrivateNetwork)
	nodeActionsList := newNodeActionsList(
		tviewAppPages,
		nodesList,
		mempoolTransactionsList,
		actionsHandler,
		outputView,
		nodeDetailsView,
		data,
	)
	actionsFlex := newActionsFlex(outputView, nodeActionsList)

	// init child flexes
	mempoolFlex := newMempoolFlex(mempoolTransactionsList, mempoolTransactionDetails)
	nodesFlex := newNodesFlex(nodesList, nodeDetailsView)

	// main app flex
	footer := tview.NewTextView().
		SetText(fmt.Sprintf(
			"Version: %s. Navigate with ←↑↓→(arrows); use ↹(Tab) for modals.",
			version,
		)).
		SetTextAlign(tview.AlignCenter)

	appFlex := newAppFlex(tviewAppFlex, mempoolFlex, actionsFlex, nodesFlex, footer)

	appPages := newAppPages(tviewAppPages, nodeActionsList, appFlex)

	app := newApp(appPages, nodesList, actionsFlex, mempoolTransactionsList, nodeActionsList)

	go func() {
		for {
			const delay = time.Second / 2

			<-time.After(delay)

			app.QueueUpdateDraw(func() {
				if err := data.update(); err != nil {
					outputView.AddError(fmt.Sprintf("update error: %s", err))
					return
				}

				mempoolTransactionsList.refresh()
				nodeDetailsView.refresh()
			})
		}
	}()

	return &TUI{
		app:       app,
		btcPrvNet: btcPrivateNetwork,
	}
}

func centeredForm(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

// Run starts the Terminal User Interface.
func (t *TUI) Run() error {
	if err := t.app.Run(); err != nil {
		return fmt.Errorf("run tview app: %w", err)
	}

	return nil
}

type app struct {
	*tview.Application
}

func newApp(
	appPages *appPages,
	nodes *nodesList,
	actionsFlex *actionsFlex,
	mempoolTransactions *mempoolTransactionsList,
	nodeActions *nodeActionsList,
) *app {
	application := tview.NewApplication()

	application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// nolint: exhaustive
		switch event.Key() {
		case tcell.KeyLeft:
			switch application.GetFocus() {
			case nodes:
				application.SetFocus(actionsFlex)
			case mempoolTransactions:
				application.SetFocus(nodes)
			case nodeActions:
				application.SetFocus(mempoolTransactions)
			}

		case tcell.KeyRight:
			switch application.GetFocus() {
			case nodes:
				application.SetFocus(mempoolTransactions)
			case mempoolTransactions:
				application.SetFocus(actionsFlex)
			case nodeActions, actionsFlex:
				application.SetFocus(nodes)
			}
		}

		return event
	})

	application.SetRoot(appPages, true)

	return &app{
		Application: application,
	}
}

func inputCaptureIgnoreLeftRight(event *tcell.EventKey) *tcell.EventKey {
	// nolint: exhaustive
	switch event.Key() {
	case tcell.KeyLeft, tcell.KeyRight:
		return nil
	}

	return event
}
