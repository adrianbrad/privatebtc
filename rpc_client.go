package privatebtc

import (
	"context"
)

// RPCClient is an interface for RPC clients.
// The methods are closely mapped to the bitcoin RPC API.
// nolint: interfacebloat, revive
type RPCClient interface {
	// SendToAddress sends the given satoshi amount to the given address.
	SendToAddress(ctx context.Context, address string, amount float64) (txHash string, _ error)

	// SendCustomTransaction sends a custom transaction with the given inputs and amounts.
	SendCustomTransaction(ctx context.Context, inputs []TransactionVin, amounts map[string]float64) (txHash string, _ error)

	// GenerateToAddress generates the given number of blocks to the given address.
	GenerateToAddress(ctx context.Context, numBlocks int64, address string) (blockHashes []string, _ error)

	// GetConnectionCount returns the number of connections to other nodes.
	GetConnectionCount(ctx context.Context) (int, error)

	// AddPeer adds the given peer to the node.
	AddPeer(ctx context.Context, peer Node) error

	// RemovePeer removes the given peer from the node.
	RemovePeer(ctx context.Context, peer Node) error

	// CreateWallet creates a new wallet with the given name.
	CreateWallet(ctx context.Context, walletName string) error

	// GetRawMempool returns all transaction ids in memory pool
	GetRawMempool(ctx context.Context) ([]string, error)

	// GetBlockCount returns the number of blocks in the longest blockchain.
	GetBlockCount(ctx context.Context) (int, error)

	// GetNewAddress returns a new address for receiving payments.
	GetNewAddress(ctx context.Context, label string) (string, error)

	// GetBalance returns the total available balance.
	// From the Bitcoin Core RPC Docs:
	// The available balance is what the wallet considers currently spendable, and is
	// thus affected by options which limit spendability such as -spendzeroconfchange.
	GetBalance(ctx context.Context) (Balance, error)

	// GetTransaction returns the transaction with the given Hash.
	GetTransaction(ctx context.Context, txHash string) (*Transaction, error)

	ListAddresses(ctx context.Context) ([]string, error)

	GetBestBlockHash(ctx context.Context) (string, error)

	GetCoinbaseValue(ctx context.Context) (int64, error)

	GetTransactionOutputs(ctx context.Context, txHash string) ([]MempoolTransactionOutput, error)
}

// RPCClientFactory is an interface for RPC client factories.
// It is used to decouple the creation of the RPC Client from the actual implementation.
// We have to use the factory pattern because the RPC Clients are created dynamically for each
// node when the network is created.
type RPCClientFactory interface {
	NewRPCClient(hostRPCPort, rpcUser, rpcPass string) (RPCClient, error)
}
