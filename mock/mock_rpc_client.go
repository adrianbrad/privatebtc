// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/adrianbrad/privatebtc"
	"sync"
)

// Ensure, that RPCClient does implement privatebtc.RPCClient.
// If this is not the case, regenerate this file with moq.
var _ privatebtc.RPCClient = &RPCClient{}

// RPCClient is a mock implementation of privatebtc.RPCClient.
//
//	func TestSomethingThatUsesRPCClient(t *testing.T) {
//
//		// make and configure a mocked privatebtc.RPCClient
//		mockedRPCClient := &RPCClient{
//			AddPeerFunc: func(ctx context.Context, peer privatebtc.Node) error {
//				panic("mock out the AddPeer method")
//			},
//			CreateWalletFunc: func(ctx context.Context, walletName string) error {
//				panic("mock out the CreateWallet method")
//			},
//			GenerateToAddressFunc: func(ctx context.Context, numBlocks int64, address string) ([]string, error) {
//				panic("mock out the GenerateToAddress method")
//			},
//			GetBalanceFunc: func(ctx context.Context) (privatebtc.Balance, error) {
//				panic("mock out the GetBalance method")
//			},
//			GetBestBlockHashFunc: func(ctx context.Context) (string, error) {
//				panic("mock out the GetBestBlockHash method")
//			},
//			GetBlockCountFunc: func(ctx context.Context) (int, error) {
//				panic("mock out the GetBlockCount method")
//			},
//			GetCoinbaseValueFunc: func(ctx context.Context) (int64, error) {
//				panic("mock out the GetCoinbaseValue method")
//			},
//			GetConnectionCountFunc: func(ctx context.Context) (int, error) {
//				panic("mock out the GetConnectionCount method")
//			},
//			GetNewAddressFunc: func(ctx context.Context, label string) (string, error) {
//				panic("mock out the GetNewAddress method")
//			},
//			GetRawMempoolFunc: func(ctx context.Context) ([]string, error) {
//				panic("mock out the GetRawMempool method")
//			},
//			GetTransactionFunc: func(ctx context.Context, txHash string) (*privatebtc.Transaction, error) {
//				panic("mock out the GetTransaction method")
//			},
//			ListAddressesFunc: func(ctx context.Context) ([]string, error) {
//				panic("mock out the ListAddresses method")
//			},
//			RemovePeerFunc: func(ctx context.Context, peer privatebtc.Node) error {
//				panic("mock out the RemovePeer method")
//			},
//			SendCustomTransactionFunc: func(ctx context.Context, inputs []privatebtc.TransactionVin, amounts map[string]float64) (string, error) {
//				panic("mock out the SendCustomTransaction method")
//			},
//			SendToAddressFunc: func(ctx context.Context, address string, amount float64) (string, error) {
//				panic("mock out the SendToAddress method")
//			},
//		}
//
//		// use mockedRPCClient in code that requires privatebtc.RPCClient
//		// and then make assertions.
//
//	}
type RPCClient struct {
	// AddPeerFunc mocks the AddPeer method.
	AddPeerFunc func(ctx context.Context, peer privatebtc.Node) error

	// CreateWalletFunc mocks the CreateWallet method.
	CreateWalletFunc func(ctx context.Context, walletName string) error

	// GenerateToAddressFunc mocks the GenerateToAddress method.
	GenerateToAddressFunc func(ctx context.Context, numBlocks int64, address string) ([]string, error)

	// GetBalanceFunc mocks the GetBalance method.
	GetBalanceFunc func(ctx context.Context) (privatebtc.Balance, error)

	// GetBestBlockHashFunc mocks the GetBestBlockHash method.
	GetBestBlockHashFunc func(ctx context.Context) (string, error)

	// GetBlockCountFunc mocks the GetBlockCount method.
	GetBlockCountFunc func(ctx context.Context) (int, error)

	// GetCoinbaseValueFunc mocks the GetCoinbaseValue method.
	GetCoinbaseValueFunc func(ctx context.Context) (int64, error)

	// GetConnectionCountFunc mocks the GetConnectionCount method.
	GetConnectionCountFunc func(ctx context.Context) (int, error)

	// GetNewAddressFunc mocks the GetNewAddress method.
	GetNewAddressFunc func(ctx context.Context, label string) (string, error)

	// GetRawMempoolFunc mocks the GetRawMempool method.
	GetRawMempoolFunc func(ctx context.Context) ([]string, error)

	// GetTransactionFunc mocks the GetTransaction method.
	GetTransactionFunc func(ctx context.Context, txHash string) (*privatebtc.Transaction, error)

	// ListAddressesFunc mocks the ListAddresses method.
	ListAddressesFunc func(ctx context.Context) ([]string, error)

	// RemovePeerFunc mocks the RemovePeer method.
	RemovePeerFunc func(ctx context.Context, peer privatebtc.Node) error

	// SendCustomTransactionFunc mocks the SendCustomTransaction method.
	SendCustomTransactionFunc func(ctx context.Context, inputs []privatebtc.TransactionVin, amounts map[string]float64) (string, error)

	// SendToAddressFunc mocks the SendToAddress method.
	SendToAddressFunc func(ctx context.Context, address string, amount float64) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// AddPeer holds details about calls to the AddPeer method.
		AddPeer []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Peer is the peer argument value.
			Peer privatebtc.Node
		}
		// CreateWallet holds details about calls to the CreateWallet method.
		CreateWallet []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// WalletName is the walletName argument value.
			WalletName string
		}
		// GenerateToAddress holds details about calls to the GenerateToAddress method.
		GenerateToAddress []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// NumBlocks is the numBlocks argument value.
			NumBlocks int64
			// Address is the address argument value.
			Address string
		}
		// GetBalance holds details about calls to the GetBalance method.
		GetBalance []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// GetBestBlockHash holds details about calls to the GetBestBlockHash method.
		GetBestBlockHash []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// GetBlockCount holds details about calls to the GetBlockCount method.
		GetBlockCount []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// GetCoinbaseValue holds details about calls to the GetCoinbaseValue method.
		GetCoinbaseValue []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// GetConnectionCount holds details about calls to the GetConnectionCount method.
		GetConnectionCount []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// GetNewAddress holds details about calls to the GetNewAddress method.
		GetNewAddress []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Label is the label argument value.
			Label string
		}
		// GetRawMempool holds details about calls to the GetRawMempool method.
		GetRawMempool []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// GetTransaction holds details about calls to the GetTransaction method.
		GetTransaction []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// TxHash is the txHash argument value.
			TxHash string
		}
		// ListAddresses holds details about calls to the ListAddresses method.
		ListAddresses []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// RemovePeer holds details about calls to the RemovePeer method.
		RemovePeer []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Peer is the peer argument value.
			Peer privatebtc.Node
		}
		// SendCustomTransaction holds details about calls to the SendCustomTransaction method.
		SendCustomTransaction []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Inputs is the inputs argument value.
			Inputs []privatebtc.TransactionVin
			// Amounts is the amounts argument value.
			Amounts map[string]float64
		}
		// SendToAddress holds details about calls to the SendToAddress method.
		SendToAddress []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Address is the address argument value.
			Address string
			// Amount is the amount argument value.
			Amount float64
		}
	}
	lockAddPeer               sync.RWMutex
	lockCreateWallet          sync.RWMutex
	lockGenerateToAddress     sync.RWMutex
	lockGetBalance            sync.RWMutex
	lockGetBestBlockHash      sync.RWMutex
	lockGetBlockCount         sync.RWMutex
	lockGetCoinbaseValue      sync.RWMutex
	lockGetConnectionCount    sync.RWMutex
	lockGetNewAddress         sync.RWMutex
	lockGetRawMempool         sync.RWMutex
	lockGetTransaction        sync.RWMutex
	lockListAddresses         sync.RWMutex
	lockRemovePeer            sync.RWMutex
	lockSendCustomTransaction sync.RWMutex
	lockSendToAddress         sync.RWMutex
}

// AddPeer calls AddPeerFunc.
func (mock *RPCClient) AddPeer(ctx context.Context, peer privatebtc.Node) error {
	if mock.AddPeerFunc == nil {
		panic("RPCClient.AddPeerFunc: method is nil but RPCClient.AddPeer was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Peer privatebtc.Node
	}{
		Ctx:  ctx,
		Peer: peer,
	}
	mock.lockAddPeer.Lock()
	mock.calls.AddPeer = append(mock.calls.AddPeer, callInfo)
	mock.lockAddPeer.Unlock()
	return mock.AddPeerFunc(ctx, peer)
}

// AddPeerCalls gets all the calls that were made to AddPeer.
// Check the length with:
//
//	len(mockedRPCClient.AddPeerCalls())
func (mock *RPCClient) AddPeerCalls() []struct {
	Ctx  context.Context
	Peer privatebtc.Node
} {
	var calls []struct {
		Ctx  context.Context
		Peer privatebtc.Node
	}
	mock.lockAddPeer.RLock()
	calls = mock.calls.AddPeer
	mock.lockAddPeer.RUnlock()
	return calls
}

// CreateWallet calls CreateWalletFunc.
func (mock *RPCClient) CreateWallet(ctx context.Context, walletName string) error {
	if mock.CreateWalletFunc == nil {
		panic("RPCClient.CreateWalletFunc: method is nil but RPCClient.CreateWallet was just called")
	}
	callInfo := struct {
		Ctx        context.Context
		WalletName string
	}{
		Ctx:        ctx,
		WalletName: walletName,
	}
	mock.lockCreateWallet.Lock()
	mock.calls.CreateWallet = append(mock.calls.CreateWallet, callInfo)
	mock.lockCreateWallet.Unlock()
	return mock.CreateWalletFunc(ctx, walletName)
}

// CreateWalletCalls gets all the calls that were made to CreateWallet.
// Check the length with:
//
//	len(mockedRPCClient.CreateWalletCalls())
func (mock *RPCClient) CreateWalletCalls() []struct {
	Ctx        context.Context
	WalletName string
} {
	var calls []struct {
		Ctx        context.Context
		WalletName string
	}
	mock.lockCreateWallet.RLock()
	calls = mock.calls.CreateWallet
	mock.lockCreateWallet.RUnlock()
	return calls
}

// GenerateToAddress calls GenerateToAddressFunc.
func (mock *RPCClient) GenerateToAddress(ctx context.Context, numBlocks int64, address string) ([]string, error) {
	if mock.GenerateToAddressFunc == nil {
		panic("RPCClient.GenerateToAddressFunc: method is nil but RPCClient.GenerateToAddress was just called")
	}
	callInfo := struct {
		Ctx       context.Context
		NumBlocks int64
		Address   string
	}{
		Ctx:       ctx,
		NumBlocks: numBlocks,
		Address:   address,
	}
	mock.lockGenerateToAddress.Lock()
	mock.calls.GenerateToAddress = append(mock.calls.GenerateToAddress, callInfo)
	mock.lockGenerateToAddress.Unlock()
	return mock.GenerateToAddressFunc(ctx, numBlocks, address)
}

// GenerateToAddressCalls gets all the calls that were made to GenerateToAddress.
// Check the length with:
//
//	len(mockedRPCClient.GenerateToAddressCalls())
func (mock *RPCClient) GenerateToAddressCalls() []struct {
	Ctx       context.Context
	NumBlocks int64
	Address   string
} {
	var calls []struct {
		Ctx       context.Context
		NumBlocks int64
		Address   string
	}
	mock.lockGenerateToAddress.RLock()
	calls = mock.calls.GenerateToAddress
	mock.lockGenerateToAddress.RUnlock()
	return calls
}

// GetBalance calls GetBalanceFunc.
func (mock *RPCClient) GetBalance(ctx context.Context) (privatebtc.Balance, error) {
	if mock.GetBalanceFunc == nil {
		panic("RPCClient.GetBalanceFunc: method is nil but RPCClient.GetBalance was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockGetBalance.Lock()
	mock.calls.GetBalance = append(mock.calls.GetBalance, callInfo)
	mock.lockGetBalance.Unlock()
	return mock.GetBalanceFunc(ctx)
}

// GetBalanceCalls gets all the calls that were made to GetBalance.
// Check the length with:
//
//	len(mockedRPCClient.GetBalanceCalls())
func (mock *RPCClient) GetBalanceCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockGetBalance.RLock()
	calls = mock.calls.GetBalance
	mock.lockGetBalance.RUnlock()
	return calls
}

// GetBestBlockHash calls GetBestBlockHashFunc.
func (mock *RPCClient) GetBestBlockHash(ctx context.Context) (string, error) {
	if mock.GetBestBlockHashFunc == nil {
		panic("RPCClient.GetBestBlockHashFunc: method is nil but RPCClient.GetBestBlockHash was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockGetBestBlockHash.Lock()
	mock.calls.GetBestBlockHash = append(mock.calls.GetBestBlockHash, callInfo)
	mock.lockGetBestBlockHash.Unlock()
	return mock.GetBestBlockHashFunc(ctx)
}

// GetBestBlockHashCalls gets all the calls that were made to GetBestBlockHash.
// Check the length with:
//
//	len(mockedRPCClient.GetBestBlockHashCalls())
func (mock *RPCClient) GetBestBlockHashCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockGetBestBlockHash.RLock()
	calls = mock.calls.GetBestBlockHash
	mock.lockGetBestBlockHash.RUnlock()
	return calls
}

// GetBlockCount calls GetBlockCountFunc.
func (mock *RPCClient) GetBlockCount(ctx context.Context) (int, error) {
	if mock.GetBlockCountFunc == nil {
		panic("RPCClient.GetBlockCountFunc: method is nil but RPCClient.GetBlockCount was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockGetBlockCount.Lock()
	mock.calls.GetBlockCount = append(mock.calls.GetBlockCount, callInfo)
	mock.lockGetBlockCount.Unlock()
	return mock.GetBlockCountFunc(ctx)
}

// GetBlockCountCalls gets all the calls that were made to GetBlockCount.
// Check the length with:
//
//	len(mockedRPCClient.GetBlockCountCalls())
func (mock *RPCClient) GetBlockCountCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockGetBlockCount.RLock()
	calls = mock.calls.GetBlockCount
	mock.lockGetBlockCount.RUnlock()
	return calls
}

// GetCoinbaseValue calls GetCoinbaseValueFunc.
func (mock *RPCClient) GetCoinbaseValue(ctx context.Context) (int64, error) {
	if mock.GetCoinbaseValueFunc == nil {
		panic("RPCClient.GetCoinbaseValueFunc: method is nil but RPCClient.GetCoinbaseValue was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockGetCoinbaseValue.Lock()
	mock.calls.GetCoinbaseValue = append(mock.calls.GetCoinbaseValue, callInfo)
	mock.lockGetCoinbaseValue.Unlock()
	return mock.GetCoinbaseValueFunc(ctx)
}

// GetCoinbaseValueCalls gets all the calls that were made to GetCoinbaseValue.
// Check the length with:
//
//	len(mockedRPCClient.GetCoinbaseValueCalls())
func (mock *RPCClient) GetCoinbaseValueCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockGetCoinbaseValue.RLock()
	calls = mock.calls.GetCoinbaseValue
	mock.lockGetCoinbaseValue.RUnlock()
	return calls
}

// GetConnectionCount calls GetConnectionCountFunc.
func (mock *RPCClient) GetConnectionCount(ctx context.Context) (int, error) {
	if mock.GetConnectionCountFunc == nil {
		panic("RPCClient.GetConnectionCountFunc: method is nil but RPCClient.GetConnectionCount was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockGetConnectionCount.Lock()
	mock.calls.GetConnectionCount = append(mock.calls.GetConnectionCount, callInfo)
	mock.lockGetConnectionCount.Unlock()
	return mock.GetConnectionCountFunc(ctx)
}

// GetConnectionCountCalls gets all the calls that were made to GetConnectionCount.
// Check the length with:
//
//	len(mockedRPCClient.GetConnectionCountCalls())
func (mock *RPCClient) GetConnectionCountCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockGetConnectionCount.RLock()
	calls = mock.calls.GetConnectionCount
	mock.lockGetConnectionCount.RUnlock()
	return calls
}

// GetNewAddress calls GetNewAddressFunc.
func (mock *RPCClient) GetNewAddress(ctx context.Context, label string) (string, error) {
	if mock.GetNewAddressFunc == nil {
		panic("RPCClient.GetNewAddressFunc: method is nil but RPCClient.GetNewAddress was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		Label string
	}{
		Ctx:   ctx,
		Label: label,
	}
	mock.lockGetNewAddress.Lock()
	mock.calls.GetNewAddress = append(mock.calls.GetNewAddress, callInfo)
	mock.lockGetNewAddress.Unlock()
	return mock.GetNewAddressFunc(ctx, label)
}

// GetNewAddressCalls gets all the calls that were made to GetNewAddress.
// Check the length with:
//
//	len(mockedRPCClient.GetNewAddressCalls())
func (mock *RPCClient) GetNewAddressCalls() []struct {
	Ctx   context.Context
	Label string
} {
	var calls []struct {
		Ctx   context.Context
		Label string
	}
	mock.lockGetNewAddress.RLock()
	calls = mock.calls.GetNewAddress
	mock.lockGetNewAddress.RUnlock()
	return calls
}

// GetRawMempool calls GetRawMempoolFunc.
func (mock *RPCClient) GetRawMempool(ctx context.Context) ([]string, error) {
	if mock.GetRawMempoolFunc == nil {
		panic("RPCClient.GetRawMempoolFunc: method is nil but RPCClient.GetRawMempool was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockGetRawMempool.Lock()
	mock.calls.GetRawMempool = append(mock.calls.GetRawMempool, callInfo)
	mock.lockGetRawMempool.Unlock()
	return mock.GetRawMempoolFunc(ctx)
}

// GetRawMempoolCalls gets all the calls that were made to GetRawMempool.
// Check the length with:
//
//	len(mockedRPCClient.GetRawMempoolCalls())
func (mock *RPCClient) GetRawMempoolCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockGetRawMempool.RLock()
	calls = mock.calls.GetRawMempool
	mock.lockGetRawMempool.RUnlock()
	return calls
}

// GetTransaction calls GetTransactionFunc.
func (mock *RPCClient) GetTransaction(ctx context.Context, txHash string) (*privatebtc.Transaction, error) {
	if mock.GetTransactionFunc == nil {
		panic("RPCClient.GetTransactionFunc: method is nil but RPCClient.GetTransaction was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		TxHash string
	}{
		Ctx:    ctx,
		TxHash: txHash,
	}
	mock.lockGetTransaction.Lock()
	mock.calls.GetTransaction = append(mock.calls.GetTransaction, callInfo)
	mock.lockGetTransaction.Unlock()
	return mock.GetTransactionFunc(ctx, txHash)
}

// GetTransactionCalls gets all the calls that were made to GetTransaction.
// Check the length with:
//
//	len(mockedRPCClient.GetTransactionCalls())
func (mock *RPCClient) GetTransactionCalls() []struct {
	Ctx    context.Context
	TxHash string
} {
	var calls []struct {
		Ctx    context.Context
		TxHash string
	}
	mock.lockGetTransaction.RLock()
	calls = mock.calls.GetTransaction
	mock.lockGetTransaction.RUnlock()
	return calls
}

// ListAddresses calls ListAddressesFunc.
func (mock *RPCClient) ListAddresses(ctx context.Context) ([]string, error) {
	if mock.ListAddressesFunc == nil {
		panic("RPCClient.ListAddressesFunc: method is nil but RPCClient.ListAddresses was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockListAddresses.Lock()
	mock.calls.ListAddresses = append(mock.calls.ListAddresses, callInfo)
	mock.lockListAddresses.Unlock()
	return mock.ListAddressesFunc(ctx)
}

// ListAddressesCalls gets all the calls that were made to ListAddresses.
// Check the length with:
//
//	len(mockedRPCClient.ListAddressesCalls())
func (mock *RPCClient) ListAddressesCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockListAddresses.RLock()
	calls = mock.calls.ListAddresses
	mock.lockListAddresses.RUnlock()
	return calls
}

// RemovePeer calls RemovePeerFunc.
func (mock *RPCClient) RemovePeer(ctx context.Context, peer privatebtc.Node) error {
	if mock.RemovePeerFunc == nil {
		panic("RPCClient.RemovePeerFunc: method is nil but RPCClient.RemovePeer was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Peer privatebtc.Node
	}{
		Ctx:  ctx,
		Peer: peer,
	}
	mock.lockRemovePeer.Lock()
	mock.calls.RemovePeer = append(mock.calls.RemovePeer, callInfo)
	mock.lockRemovePeer.Unlock()
	return mock.RemovePeerFunc(ctx, peer)
}

// RemovePeerCalls gets all the calls that were made to RemovePeer.
// Check the length with:
//
//	len(mockedRPCClient.RemovePeerCalls())
func (mock *RPCClient) RemovePeerCalls() []struct {
	Ctx  context.Context
	Peer privatebtc.Node
} {
	var calls []struct {
		Ctx  context.Context
		Peer privatebtc.Node
	}
	mock.lockRemovePeer.RLock()
	calls = mock.calls.RemovePeer
	mock.lockRemovePeer.RUnlock()
	return calls
}

// SendCustomTransaction calls SendCustomTransactionFunc.
func (mock *RPCClient) SendCustomTransaction(ctx context.Context, inputs []privatebtc.TransactionVin, amounts map[string]float64) (string, error) {
	if mock.SendCustomTransactionFunc == nil {
		panic("RPCClient.SendCustomTransactionFunc: method is nil but RPCClient.SendCustomTransaction was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Inputs  []privatebtc.TransactionVin
		Amounts map[string]float64
	}{
		Ctx:     ctx,
		Inputs:  inputs,
		Amounts: amounts,
	}
	mock.lockSendCustomTransaction.Lock()
	mock.calls.SendCustomTransaction = append(mock.calls.SendCustomTransaction, callInfo)
	mock.lockSendCustomTransaction.Unlock()
	return mock.SendCustomTransactionFunc(ctx, inputs, amounts)
}

// SendCustomTransactionCalls gets all the calls that were made to SendCustomTransaction.
// Check the length with:
//
//	len(mockedRPCClient.SendCustomTransactionCalls())
func (mock *RPCClient) SendCustomTransactionCalls() []struct {
	Ctx     context.Context
	Inputs  []privatebtc.TransactionVin
	Amounts map[string]float64
} {
	var calls []struct {
		Ctx     context.Context
		Inputs  []privatebtc.TransactionVin
		Amounts map[string]float64
	}
	mock.lockSendCustomTransaction.RLock()
	calls = mock.calls.SendCustomTransaction
	mock.lockSendCustomTransaction.RUnlock()
	return calls
}

// SendToAddress calls SendToAddressFunc.
func (mock *RPCClient) SendToAddress(ctx context.Context, address string, amount float64) (string, error) {
	if mock.SendToAddressFunc == nil {
		panic("RPCClient.SendToAddressFunc: method is nil but RPCClient.SendToAddress was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Address string
		Amount  float64
	}{
		Ctx:     ctx,
		Address: address,
		Amount:  amount,
	}
	mock.lockSendToAddress.Lock()
	mock.calls.SendToAddress = append(mock.calls.SendToAddress, callInfo)
	mock.lockSendToAddress.Unlock()
	return mock.SendToAddressFunc(ctx, address, amount)
}

// SendToAddressCalls gets all the calls that were made to SendToAddress.
// Check the length with:
//
//	len(mockedRPCClient.SendToAddressCalls())
func (mock *RPCClient) SendToAddressCalls() []struct {
	Ctx     context.Context
	Address string
	Amount  float64
} {
	var calls []struct {
		Ctx     context.Context
		Address string
		Amount  float64
	}
	mock.lockSendToAddress.RLock()
	calls = mock.calls.SendToAddress
	mock.lockSendToAddress.RUnlock()
	return calls
}