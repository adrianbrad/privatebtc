package tview

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/adrianbrad/privatebtc"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/rivo/tview"
)

const inputFieldWithd = bitcoinMaxAddressLength

type nodesList struct {
	*tview.List
}

func newNodesList(
	nodes []privatebtc.Node,
	nodeDetailsView *nodeDetailsView,
) *nodesList {
	list := tview.NewList()

	list.
		SetChangedFunc(func(nodeID int, _, _ string, _ rune) {
			nodeDetailsView.updateDisplayedNode(nodeID)
		}).
		SetSelectedFocusOnly(false).
		ShowSecondaryText(false).
		SetInputCapture(inputCaptureIgnoreLeftRight).
		SetBorder(true).
		SetTitle("Nodes")

	for i := range nodes {
		list.AddItem(nodes[i].Name(), strconv.Itoa(i), 0, nil)
	}

	return &nodesList{
		List: list,
	}
}

type nodeDetailsView struct {
	*tview.TextView
	currentNode int
	data        *data
}

func (v *nodeDetailsView) refresh() {
	v.SetText(v.data.nodesDetails[v.currentNode].String())
}

func (v *nodeDetailsView) updateDisplayedNode(nodeID int) {
	v.currentNode = nodeID
	v.SetText(v.data.nodesDetails[v.currentNode].String())
}

func newNodeDetailsView(data *data) *nodeDetailsView {
	tv := tview.NewTextView()

	tv.
		SetDynamicColors(true).
		SetInputCapture(nil).
		SetBorder(true).
		SetTitle("Details")

	return &nodeDetailsView{
		TextView:    tv,
		currentNode: 0,
		data:        data,
	}
}

type mempoolTransactionsList struct {
	*tview.List
	mempoolTransactionDetails *mempoolTransactionDetails
	data                      *data
}

func newMempoolTransactionsList(
	mempoolTransactionDetails *mempoolTransactionDetails,
	data *data,
) *mempoolTransactionsList {
	list := tview.NewList()

	list.
		SetSelectedFocusOnly(false).
		ShowSecondaryText(false).
		SetChangedFunc(func(_ int, txHash, _ string, _ rune) {
			mempoolTransactionDetails.updateDisplayedTx(txHash)
		}).
		SetInputCapture(inputCaptureIgnoreLeftRight).
		SetBorder(true).
		SetTitle("Transactions")

	return &mempoolTransactionsList{
		List:                      list,
		mempoolTransactionDetails: mempoolTransactionDetails,
		data:                      data,
	}
}

func (l *mempoolTransactionsList) refresh() {
	l.Clear()

	if len(l.data.mempool) == 0 {
		l.mempoolTransactionDetails.Clear()
		return
	}

	for k := range l.data.mempool {
		l.AddItem(k, "", 0, nil)
	}
}

type mempoolTransactionDetails struct {
	*tview.TextView
	currentTxHash string
	data          *data
}

func (d *mempoolTransactionDetails) updateDisplayedTx(txHash string) {
	d.currentTxHash = txHash
	d.SetText(d.data.mempool[d.currentTxHash].String())
}

func newMempoolTransactionDetails(data *data) *mempoolTransactionDetails {
	tv := tview.NewTextView()

	tv.SetInputCapture(nil).
		SetBorder(true).
		SetTitle("Details")

	return &mempoolTransactionDetails{
		TextView:      tv,
		currentTxHash: "",
		data:          data,
	}
}

type nodeActionsList struct {
	*tview.List
	sendBitcoinForm   *sendBitcoinForm
	mineBlocksForm    *mineBlocksForm
	rbfDrainToAddress *replaceByFeeDrainToAddressForm
}

// nolint: gocognit
func newNodeActionsList(
	appPages *tview.Pages,
	nodesList *nodesList,
	mempoolTxList *mempoolTransactionsList,
	actionsHandler *actionsHandler,
	outputView *outputView,
	nodeDetailsView *nodeDetailsView,
	data *data,
) *nodeActionsList {
	sendBitcoinForm := newSendBitcoinForm(
		appPages,
		actionsHandler,
		outputView,
		nodesList,
	)

	mineBlocksForm := newMineBlocksForm(
		appPages,
		actionsHandler,
		outputView,
		nodesList,
	)

	rbfDrainToAdress := newReplaceByFeeDrainToAddressForm(
		appPages,
		actionsHandler,
		outputView,
		nodesList,
	)

	list := tview.NewList()

	list.
		AddItem("Create Address", "", 0, nil).
		AddItem("Send to address", "", 0, nil).
		AddItem("Mine to address", "", 0, nil).
		AddItem("Disconnect from Network", "", 0, nil).
		AddItem("Connect to Network", "", 0, nil).
		AddItem("Replace By Fee Drain To Address", "", 0, nil).
		SetSelectedFunc(func(actionIndex int, _, _ string, _ rune) {
			currentNodeIndex := nodesList.GetCurrentItem()

			// nolint: gomnd
			switch actionIndex {
			case 0: // Create Address
				addr, err := actionsHandler.handleCreateAddress(currentNodeIndex)
				if err != nil {
					outputView.AddError(fmt.Sprintf(
						"error while creating address for node %d: %s",
						actionIndex,
						err,
					))
					return
				}

				actionsHandler.data.nodesDetails[currentNodeIndex].addresses = append(
					actionsHandler.data.nodesDetails[currentNodeIndex].addresses,
					addr,
				)

				outputView.AddSuccess(fmt.Sprintf(
					"created new address for node %d: %s",
					currentNodeIndex,
					addr,
				))

				nodeDetailsView.updateDisplayedNode(currentNodeIndex)

			case 1: // Send to address
				appPages.ShowPage("sendBitcoinForm")

				sendBitcoinForm.GetFormItem(0).(*tview.TextView).SetText(
					fmt.Sprintf(
						"Node %d Balance: %.2f",
						currentNodeIndex,
						data.nodesDetails[currentNodeIndex].balance.Trusted,
					),
				)

				sendBitcoinForm.GetFormItem(1).(*tview.DropDown).
					SetOptions(
						data.toFormAddresses(),
						func(option string, i int) {
							if i < 0 {
								return
							}

							addr := burnAddress
							if a := strings.Split(option, ":"); len(a) == 2 {
								addr = a[1]
							}

							sendBitcoinForm.
								GetFormItem(2).(*tview.InputField).
								SetText(addr)
						}).
					SetCurrentOption(-1)

			case 2: // Mine to address
				appPages.ShowPage("mineBlocksForm")

				mineBlocksForm.GetFormItem(0).(*tview.TextView).SetText(
					fmt.Sprintf("Node %d", currentNodeIndex),
				)

				coinbase, err := actionsHandler.btcpn.Nodes()[currentNodeIndex].
					RPCClient().GetCoinbaseValue(ctx)
				if err != nil {
					outputView.AddError(fmt.Sprintf(
						"get coinbase from node %d: %v",
						currentNodeIndex,
						err,
					))

					return
				}

				mineBlocksForm.GetFormItem(4).(*tview.TextView).SetText(
					fmt.Sprintf("%.2f BTC", btcutil.Amount(coinbase).ToBTC()),
				)

				mineBlocksForm.GetFormItem(1).(*tview.DropDown).
					SetOptions(
						data.toFormAddresses(),
						func(option string, i int) {
							if i < 0 {
								return
							}

							addr := burnAddress
							if a := strings.Split(option, ":"); len(a) == 2 {
								addr = a[1]
							}

							mineBlocksForm.
								GetFormItem(2).(*tview.InputField).
								SetText(addr)
						}).
					SetCurrentOption(-1)
			case 3: // Disconnect from network
				if err := actionsHandler.handleDisconnectFromNetwork(currentNodeIndex); err != nil {
					outputView.AddError(fmt.Sprintf(
						"disconnect node %d from network: %v",
						currentNodeIndex,
						err,
					))

					return
				}

				outputView.AddSuccess(fmt.Sprintf(
					"node %d disconncted from network",
					currentNodeIndex,
				))

				actionsHandler.data.nodesDetails[currentNodeIndex].connected = false
			case 4: // Connect to network
				if err := actionsHandler.handleConnectToNetwork(currentNodeIndex); err != nil {
					outputView.AddError(fmt.Sprintf(
						"connect node %d to network: %v",
						currentNodeIndex,
						err,
					))

					return
				}

				outputView.AddSuccess(fmt.Sprintf("node %d connected to network", currentNodeIndex))

				actionsHandler.data.nodesDetails[currentNodeIndex].connected = true
			case 5: // Replace by fee drain to address
				appPages.ShowPage("replaceByFeeDrainToAddressForm")

				rbfDrainToAdress.GetFormItem(0).(*tview.TextView).SetText(
					fmt.Sprintf("Node %d", currentNodeIndex),
				)

				txID, _ := mempoolTxList.GetItemText(mempoolTxList.GetCurrentItem())

				rbfDrainToAdress.GetFormItem(1).(*tview.TextView).SetText(txID)

				nodeRPCClient := actionsHandler.btcpn.Nodes()[currentNodeIndex].RPCClient()

				tx, err := nodeRPCClient.GetTransaction(ctx, txID)
				if err != nil {
					outputView.AddError(fmt.Sprintf(
						"retrieve tx %q: %v",
						txID,
						err,
					))

					return
				}

				totalInputs, err := tx.TotalInputsValue(ctx, nodeRPCClient)
				if err != nil {
					outputView.AddError(fmt.Sprintf(
						"retrieve total inputs value for tx %q: %v",
						txID,
						err,
					))

					return
				}

				rbfDrainToAdress.GetFormItem(2).(*tview.TextView).
					SetText(fmt.Sprintf("%.2f", totalInputs))

				rbfDrainToAdress.GetFormItem(3).(*tview.DropDown).
					SetOptions(
						data.toFormAddresses(),
						func(option string, i int) {
							if i < 0 {
								return
							}

							addr := burnAddress
							if a := strings.Split(option, ":"); len(a) == 2 {
								addr = a[1]
							}

							rbfDrainToAdress.
								GetFormItem(4).(*tview.InputField).
								SetText(addr)
						}).
					SetCurrentOption(-1)
			}
		}).
		ShowSecondaryText(false).
		SetSelectedFocusOnly(true).
		SetInputCapture(inputCaptureIgnoreLeftRight).
		SetBorder(true).
		SetTitle("Actions for the selected node(on the Nodes list)")

	return &nodeActionsList{
		List:              list,
		sendBitcoinForm:   sendBitcoinForm,
		mineBlocksForm:    mineBlocksForm,
		rbfDrainToAddress: rbfDrainToAdress,
	}
}

type sendBitcoinForm struct {
	*tview.Form
}

func newSendBitcoinForm(
	appPages *tview.Pages,
	actionsHandler *actionsHandler,
	output *outputView,
	nodesList *nodesList,
) *sendBitcoinForm {
	form := tview.NewForm()

	hide := hideForm(appPages, "sendBitcoinForm", form)

	const (
		labelSenderNode      = "Sender Node"
		labelAddresses       = "Addresses"
		labelReceiverAddress = "Receiver Address"
		labelAmount          = "Amount"
	)

	form.
		AddTextView(
			labelSenderNode,
			"",
			0,
			1,
			true,
			false,
		).
		AddDropDown(
			labelAddresses,
			[]string{},
			0,
			nil,
		).
		AddInputField(
			labelReceiverAddress,
			"",
			inputFieldWithd,
			tview.InputFieldMaxLength(bitcoinMaxAddressLength),
			nil,
		).
		AddInputField(
			labelAmount,
			"",
			inputFieldWithd,
			tview.InputFieldFloat,
			nil,
		).
		AddButton("Send", func() {
			defer hide()

			addr := form.GetFormItemByLabel(labelReceiverAddress).(*tview.InputField).GetText()

			amount := form.GetFormItemByLabel(labelAmount).(*tview.InputField).GetText()

			nodeID := nodesList.GetCurrentItem()

			txHash, err := actionsHandler.handleSendToAddress(
				nodeID,
				addr,
				amount,
			)
			if err != nil {
				output.AddError(fmt.Sprintf(
					"send to address %q, amount %s, from node %d: %s",
					addr,
					amount,
					nodeID,
					err,
				))
				return
			}

			output.AddSuccess(fmt.Sprintf(
				"sent %s BTC to %s from node %d with tx hash %s",
				amount,
				addr,
				nodeID,
				txHash,
			))
		}).
		AddButton("Cancel", hide).
		SetCancelFunc(hide).
		SetBorder(true).
		SetTitle("Send Bitcoin")

	return &sendBitcoinForm{
		Form: form,
	}
}

type mineBlocksForm struct {
	*tview.Form
}

func newMineBlocksForm(
	appPages *tview.Pages,
	actionsHandler *actionsHandler,
	output *outputView,
	nodesList *nodesList,
) *mineBlocksForm {
	form := tview.NewForm()

	hide := hideForm(appPages, "mineBlocksForm", form)

	const (
		labelMinerNode = "Miner Node"
		labelAddresses = "Addresses"
		labelReceiver  = "Receiver Address"
		labelBlocks    = "Blocks"
		labelCoinbase  = "Coinbase"
	)

	form.
		AddTextView(
			labelMinerNode,
			"",
			inputFieldWithd,
			1,
			true,
			false,
		).
		AddDropDown(
			labelAddresses,
			[]string{},
			0,
			nil,
		).
		AddInputField(
			labelReceiver,
			"",
			inputFieldWithd,
			tview.InputFieldMaxLength(bitcoinMaxAddressLength),
			nil,
		).
		AddInputField(
			labelBlocks,
			"",
			inputFieldWithd,
			tview.InputFieldInteger,
			nil,
		).
		AddTextView(
			labelCoinbase,
			"",
			inputFieldWithd,
			1,
			true,
			false,
		).
		AddButton("Mine", func() {
			defer hide()

			addr := form.GetFormItemByLabel(labelReceiver).(*tview.InputField).GetText()

			blocks := form.GetFormItemByLabel(labelBlocks).(*tview.InputField).GetText()

			nodeID := nodesList.GetCurrentItem()

			if err := actionsHandler.handleMineToAddress(
				nodesList.GetCurrentItem(),
				blocks,
				addr,
			); err != nil {
				output.AddError(fmt.Sprintf(
					"mine to address %q, %s blocks, from node %d: %s",
					addr,
					blocks,
					nodeID,
					err,
				))
				return
			}

			output.AddSuccess(fmt.Sprintf(
				"mined %s blocks to %s from node %d",
				blocks,
				addr,
				nodeID,
			))
		}).
		AddButton("Cancel", hide).
		SetCancelFunc(hide).
		SetBorder(true).
		SetTitle("Mine Blocks")

	return &mineBlocksForm{
		Form: form,
	}
}

type replaceByFeeDrainToAddressForm struct {
	*tview.Form
}

func newReplaceByFeeDrainToAddressForm(
	appPages *tview.Pages,
	actionsHandler *actionsHandler,
	output *outputView,
	nodesList *nodesList,
) *replaceByFeeDrainToAddressForm {
	form := tview.NewForm()

	hide := hideForm(appPages, "replaceByFeeDrainToAddressForm", form)

	const (
		labelNode               = "Node"
		labelTxID               = "Transaction ID"
		labelAmount             = "Amount"
		labelAddresses          = "Addresses"
		labelDestinationAddress = "Destination Address"
	)

	form.
		AddTextView(
			labelNode,
			"",
			0,
			1,
			true,
			false,
		).
		AddTextView(
			labelTxID,
			"",
			0,
			1,
			true,
			false,
		).
		AddTextView(
			labelAmount,
			"",
			0,
			1,
			true,
			false,
		).
		AddDropDown(
			labelAddresses,
			[]string{},
			0,
			nil,
		).
		AddInputField(
			labelDestinationAddress,
			"",
			inputFieldWithd,
			tview.InputFieldMaxLength(bitcoinMaxAddressLength),
			nil,
		).
		AddButton("Send", func() {
			defer hide()

			addr := form.GetFormItemByLabel(labelDestinationAddress).(*tview.InputField).GetText()
			txID := form.GetFormItemByLabel(labelTxID).(*tview.TextView).GetText(false)

			nodeID := nodesList.GetCurrentItem()

			newTxID, err := actionsHandler.replaceByFeeDrainToAddress(
				nodeID,
				txID,
				addr,
			)
			if err != nil {
				output.AddError(fmt.Sprintf(
					"replace tx %q, from node %d: %s",
					txID,
					nodeID,
					err,
				))
				return
			}

			output.AddSuccess(fmt.Sprintf(
				"replaced tx %q with tx %q from node %d",
				txID,
				newTxID,
				nodeID,
			))
		}).
		AddButton("Cancel", hide).
		SetCancelFunc(hide).
		SetBorder(true).
		SetTitle("Replace by Fee Drain to Address")

	return &replaceByFeeDrainToAddressForm{
		Form: form,
	}
}

func hideForm(appPages *tview.Pages, pageName string, form *tview.Form) func() {
	return func() {
		appPages.HidePage(pageName)

		for i := 0; i < form.GetFormItemCount(); i++ {
			switch formItem := form.GetFormItem(i).(type) {
			case *tview.DropDown:
				formItem.SetCurrentOption(0)
			case *tview.InputField:
				formItem.SetText("")
			}
		}

		form.SetFocus(1)
	}
}

type outputView struct {
	*tview.TextView
}

func (v *outputView) AddOutput(text string) {
	text = time.Now().Format(time.TimeOnly) + " " + text
	v.SetText(text + "\n" + v.GetText(false))
}

func (v *outputView) AddError(text string) {
	v.AddOutput("[red]ERROR[-] " + text)
}

func (v *outputView) AddSuccess(text string) {
	v.AddOutput("[green]SUCCESS[-] " + text)
}

func newOutputView() *outputView {
	v := tview.NewTextView()
	v.
		SetDynamicColors(true).
		SetBorder(true).
		SetTitle("Actions output")

	return &outputView{TextView: v}
}

type appPages struct {
	*tview.Pages
}

func newAppPages(
	pages *tview.Pages,
	nodeActionsList *nodeActionsList,
	appFlex *appFlex,
) *appPages {
	const (
		width  = 100
		height = 13
	)

	pages.
		AddPage("background", appFlex, true, true).
		AddPage("sendBitcoinForm", centeredForm(
			nodeActionsList.sendBitcoinForm,
			width,
			height,
		), true, false).
		AddPage("mineBlocksForm", centeredForm(
			nodeActionsList.mineBlocksForm,
			width,
			height,
		), true, false).
		AddPage("replaceByFeeDrainToAddressForm", centeredForm(
			nodeActionsList.rbfDrainToAddress,
			width,
			height,
		), true, false)

	return &appPages{
		Pages: pages,
	}
}
