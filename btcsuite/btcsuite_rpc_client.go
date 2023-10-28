package btcsuite

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/adrianbrad/privatebtc"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/rpcclient"
	"golang.org/x/sync/errgroup"
)

var _ privatebtc.RPCClient = (*RPCClient)(nil)

// RPCClient is an RPC client for a BTC node.
type RPCClient struct {
	client *rpcclient.Client
}

// GetNewAddress generates a new BTC address.
func (c RPCClient) GetNewAddress(_ context.Context, label string) (string, error) {
	addr, err := c.client.GetNewAddress(label)
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}

// GetConnectionCount returns the number of connections to other nodes.
func (c RPCClient) GetConnectionCount(context.Context) (int, error) {
	count, err := c.client.GetConnectionCount()
	if err != nil {
		return 0, fmt.Errorf("get connection count: %w", err)
	}

	return int(count), nil
}

// GetRawMempool returns the hashes of all transactions in the mempool.
func (c RPCClient) GetRawMempool(context.Context) ([]string, error) {
	hashes, err := c.client.GetRawMempool()
	if err != nil {
		return nil, err
	}

	hs := make([]string, len(hashes))
	for i := range hashes {
		hs[i] = hashes[i].String()
	}

	return hs, nil
}

// GetBlockCount returns the current block count.
func (c RPCClient) GetBlockCount(context.Context) (int, error) {
	bc, err := c.client.GetBlockCount()
	if err != nil {
		return 0, err
	}

	return int(bc), nil
}

// WalletWarningError is an error returned by the RPC client when a wallet warning is encountered.
type WalletWarningError string

func (e WalletWarningError) Error() string {
	return fmt.Sprintf("wallet warning: %s", string(e))
}

// CreateWallet creates a wallet with the given name.
func (c RPCClient) CreateWallet(_ context.Context, walletName string) error {
	res, err := c.client.CreateWallet(walletName)
	if err != nil {
		return fmt.Errorf("create wallet: %w", err)
	}

	if res.Warning != "" {
		return WalletWarningError(res.Warning)
	}

	return nil
}

// SendToAddress sends the given amount to the given address.
func (c RPCClient) SendToAddress(
	_ context.Context,
	address string,
	amount float64,
) (string, error) {
	am, err := btcutil.NewAmount(amount)
	if err != nil {
		return "", fmt.Errorf("new amount: %w", err)
	}

	addr, err := btcutil.DecodeAddress(address, nil)
	if err != nil {
		return "", fmt.Errorf("decode address: %w", err)
	}

	h, err := c.client.SendToAddress(addr, am)
	if err != nil {
		return "", fmt.Errorf("send to address: %w", err)
	}

	return h.String(), nil
}

// SendCustomTransaction sends a custom transaction with the given inputs and amounts.
func (c RPCClient) SendCustomTransaction(
	_ context.Context,
	inputs []privatebtc.TransactionVin,
	amounts map[string]float64,
) (string, error) {
	jsonInputs := make([]btcjson.TransactionInput, len(inputs))

	for i := range inputs {
		jsonInputs[i] = btcjson.TransactionInput{
			Txid: inputs[i].TxID,
			Vout: inputs[i].Vout,
		}
	}

	btcAmounts := make(map[btcutil.Address]btcutil.Amount, len(amounts))

	for addr, amnt := range amounts {
		btcAddr, err := btcutil.DecodeAddress(addr, nil)
		if err != nil {
			return "", fmt.Errorf("decode address %q: %w", addr, err)
		}

		am, err := btcutil.NewAmount(amnt)
		if err != nil {
			return "", fmt.Errorf("new amount %f: %w", amnt, err)
		}

		btcAmounts[btcAddr] = am
	}

	rawTx, err := c.client.CreateRawTransaction(jsonInputs, btcAmounts, nil)
	if err != nil {
		return "", fmt.Errorf("create raw transaction: %w", err)
	}

	signedTx, _, err := c.client.SignRawTransactionWithWallet(rawTx)
	if err != nil {
		return "", fmt.Errorf("sign raw transaction: %w", err)
	}

	hash, err := c.client.SendRawTransaction(signedTx, true)
	if err != nil {
		return "", fmt.Errorf("send raw transaction: %w", err)
	}

	return hash.String(), nil
}

// GenerateToAddress generates numBlocks blocks and sends the coinbase to the given address.
func (c RPCClient) GenerateToAddress(
	_ context.Context,
	numBlocks int64,
	address string,
) ([]string, error) {
	addr, err := btcutil.DecodeAddress(address, nil)
	if err != nil {
		return nil, fmt.Errorf("decode address: %w", err)
	}

	hashes, err := c.client.GenerateToAddress(numBlocks, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("generate to address: %w", err)
	}

	hs := make([]string, len(hashes))
	for i := range hashes {
		hs[i] = hashes[i].String()
	}

	return hs, nil
}

// AddPeer adds a peer to the node.
func (c RPCClient) AddPeer(_ context.Context, peer privatebtc.Node) error {
	addr := fmt.Sprintf(
		"%s:%s",
		peer.NodeHandler().InternalIP(),
		privatebtc.P2PRegtestDefaultPort,
	)

	return c.client.AddNode(addr, rpcclient.ANOneTry)
}

// RemovePeer removes a peer from the node.
func (c RPCClient) RemovePeer(ctx context.Context, peer privatebtc.Node) error {
	peerInfo, err := c.client.GetPeerInfo()
	if err != nil {
		return fmt.Errorf("get peer info: %w", err)
	}

	var addr string

	for i := range peerInfo {
		if strings.Contains(peerInfo[i].Addr, peer.NodeHandler().InternalIP()) {
			addr = peerInfo[i].Addr
		}
	}

	if addr == "" {
		return privatebtc.ErrPeerNotFound
	}

	eg, _ := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if _, err := c.client.RawRequest(
			"disconnectnode",
			[]json.RawMessage{json.RawMessage(strconv.Quote(addr))},
		); err != nil {
			return fmt.Errorf("disconnect node: %w", err)
		}

		return nil
	})

	return eg.Wait()
}

// GetBalance returns the balance of the wallet.
func (c RPCClient) GetBalance(context.Context) (privatebtc.Balance, error) {
	balances, err := c.client.GetBalances()
	if err != nil {
		return privatebtc.Balance{}, err
	}

	return privatebtc.Balance{
		Trusted:  balances.Mine.Trusted,
		Pending:  balances.Mine.UntrustedPending,
		Immature: balances.Mine.Immature,
	}, nil
}

// GetPendingBalance returns the pending balance of the wallet.
func (c RPCClient) GetPendingBalance() (float64, error) {
	balances, err := c.client.GetBalances()
	if err != nil {
		return 0, fmt.Errorf("get balances: %w", err)
	}

	return balances.Mine.UntrustedPending, nil
}

// GetTransaction returns a transaction by its hash.
func (c RPCClient) GetTransaction(
	_ context.Context,
	txHash string,
) (*privatebtc.Transaction, error) {
	resp, err := c.client.RawRequest("getrawtransaction",
		[]json.RawMessage{
			json.RawMessage(strconv.Quote(txHash)),
			json.RawMessage("true"),
		})
	if err != nil {
		return nil, fmt.Errorf("get tx request: %w", err)
	}

	// nolint: tagliatelle
	var tx struct {
		TxID      string `json:"txid"`
		Hash      string `json:"hash"`
		BlockHash string `json:"blockhash"`
		Vin       []struct {
			TxID string `json:"txid"`
			Vout uint32 `json:"vout"`
		} `json:"vin"`
		Vouts []struct {
			Value        float64 `json:"value"`
			N            uint32  `json:"n"`
			ScriptPubKey struct {
				Address string `json:"address"`
			} `json:"scriptPubKey"`
		} `json:"vout"`
	}

	if err := json.Unmarshal(resp, &tx); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	vouts := make([]privatebtc.TransactionVout, len(tx.Vouts))

	for i, v := range tx.Vouts {
		vouts[i] = privatebtc.TransactionVout{
			Value:        v.Value,
			ScriptPubKey: struct{ Address string }{Address: v.ScriptPubKey.Address},
			N:            v.N,
		}
	}

	vins := make([]privatebtc.TransactionVin, len(tx.Vin))

	for i, v := range tx.Vin {
		vins[i] = privatebtc.TransactionVin{
			TxID: v.TxID,
			Vout: v.Vout,
		}
	}

	return &privatebtc.Transaction{
		TxID:      tx.TxID,
		Hash:      tx.Hash,
		BlockHash: tx.BlockHash,
		Vout:      vouts,
		Vin:       vins,
	}, nil
}

// ListAddresses returns all addresses in the wallet.
func (c RPCClient) ListAddresses(context.Context) ([]string, error) {
	resp, err := c.client.ListReceivedByAddressIncludeEmpty(1, true)
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}

	addresses := make([]string, len(resp))

	for i := range resp {
		addresses[i] = resp[i].Address
	}

	return addresses, nil
}

// GetBestBlockHash returns the hash of the best (tip) block in the longest block chain.
func (c RPCClient) GetBestBlockHash(context.Context) (string, error) {
	h, err := c.client.GetBestBlockHash()
	if err != nil {
		return "", err
	}

	return h.String(), nil
}

// GetCoinbaseValue returns the coinbase for the next block.
func (c RPCClient) GetCoinbaseValue(context.Context) (int64, error) {
	res, err := c.client.GetBlockTemplate(&btcjson.TemplateRequest{
		Mode:         "template",
		Capabilities: []string{"coinbasevalue"},
		Rules:        []string{"segwit"},
	})
	if err != nil {
		return 0, fmt.Errorf("get block template: %w", err)
	}

	var v int64

	if res.CoinbaseValue != nil {
		v = *res.CoinbaseValue
	}

	return v, nil
}

var _ privatebtc.RPCClientFactory = (*RPCClientFactory)(nil)

// RPCClientFactory is a factory for RPC clients.
type RPCClientFactory struct {
	NoPing bool
}

// NewRPCClient creates a new RPC client.
func (f RPCClientFactory) NewRPCClient(hostPort,
	rpcUser,
	rpcPass string,
) (privatebtc.RPCClient, error) {
	rpcClient, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         "localhost:" + hostPort,
		User:         rpcUser,
		Pass:         rpcPass,
		DisableTLS:   true,
		HTTPPostMode: true,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("create rpc client: %w", err)
	}

	c := RPCClient{
		client: rpcClient,
	}

	if f.NoPing {
		return c, nil
	}

	if err := rpcClient.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return c, nil
}
