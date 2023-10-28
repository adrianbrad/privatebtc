package tview

import (
	"context"
	"fmt"
	"strconv"

	"github.com/adrianbrad/privatebtc"
)

var ctx = context.Background()

type actionsHandler struct {
	data  *data
	btcpn *privatebtc.PrivateNetwork
}

func newActionsHandler(
	data *data,
	btcpn *privatebtc.PrivateNetwork,
) *actionsHandler {
	return &actionsHandler{
		data:  data,
		btcpn: btcpn,
	}
}

func (a *actionsHandler) handleCreateAddress(nodeID int) (string, error) {
	addr, err := a.btcpn.Nodes()[nodeID].RPCClient().GetNewAddress(ctx, "acc")
	if err != nil {
		return "", fmt.Errorf("get new address: %w", err)
	}

	return addr, nil
}

func (a *actionsHandler) handleSendToAddress(
	nodeID int,
	address,
	amount string,
) (string, error) {
	amountBTC, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return "", fmt.Errorf("parse amount: %w", err)
	}

	txHash, err := a.btcpn.Nodes()[nodeID].RPCClient().SendToAddress(ctx, address, amountBTC)
	if err != nil {
		return "", fmt.Errorf("send to address: %w", err)
	}

	return txHash, nil
}

func (a *actionsHandler) handleMineToAddress(
	nodeID int,
	numBlocks,
	address string,
) error {
	numBlocksInt, err := strconv.ParseInt(numBlocks, 10, 64)
	if err != nil {
		return fmt.Errorf("parse num blocks: %w", err)
	}

	if _, err := a.btcpn.Nodes()[nodeID].RPCClient().GenerateToAddress(
		ctx,
		numBlocksInt,
		address,
	); err != nil {
		return fmt.Errorf("generate to address: %w", err)
	}

	return nil
}

func (a *actionsHandler) handleConnectToNetwork(nodeID int) error {
	if err := a.btcpn.Nodes()[nodeID].ConnectToNetwork(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	return nil
}

func (a *actionsHandler) handleDisconnectFromNetwork(nodeID int) error {
	if err := a.btcpn.Nodes()[nodeID].DisconnectFromNetwork(ctx); err != nil {
		return fmt.Errorf("disconnect: %w", err)
	}

	return nil
}

func (a *actionsHandler) replaceByFeeDrainToAddress(
	nodeID int,
	txID string,
	address string,
) (string, error) {
	node := a.btcpn.Nodes()[nodeID]

	txHash, err := privatebtc.ReplaceTransactionDrainToAddress(ctx, node.RPCClient(), txID, address)
	if err != nil {
		return "", fmt.Errorf("replace transaction drain to address: %w", err)
	}

	return txHash, nil
}
