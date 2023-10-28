package privatebtc

import (
	"errors"
	"fmt"
)

// exported errors.
var (
	// ErrCannotConnectToDockerAPI is returned when the Docker Engine API cannot be reached.
	ErrCannotConnectToDockerAPI = errors.New("cannot connect to Docker Engine API")
	// ErrPeerNotFound is returned when a peer is not found in a node's peer info response.
	ErrPeerNotFound = errors.New("peer not found in node peer info response")
	// ErrChainReorgMustDisconnectNodeFirst is returned whenever a chain reorg action is attempted
	// without first disconnecting a node from the network.
	// nolint: revive // line too long
	ErrChainReorgMustDisconnectNodeFirst = errors.New("chain reorg: must disconnect node first")
	// ErrTimeoutAndChainsAreNotSynced is returned when a timeout occurs and the
	// chains are not synced.
	ErrTimeoutAndChainsAreNotSynced = errors.New("timeout and chains are not synced")
	// ErrChainsShouldNotBeSynced is returned when the chains should not be synced.
	ErrChainsShouldNotBeSynced = errors.New("chains should not be synced")
	// ErrTxNotFoundInMempool is returned when a transaction is not found in the mempool.
	ErrTxNotFoundInMempool = errors.New("tx not found in mempool")
	// ErrTxFoundInMempool is returned when a transaction is unexpectedly found in the mempool.
	ErrTxFoundInMempool = errors.New("tx found in mempool")
	// ErrNodeIndexOutOfRange is returned when a node index is out of range.
	ErrNodeIndexOutOfRange = errors.New("node index out of range")
)

type peerCountShouldBeZeroError struct {
	got int
}

func (e *peerCountShouldBeZeroError) Error() string {
	return fmt.Sprintf("node 0 should not have any peers, got: %d", e.got)
}

// UnexpectedPeerCountError is returned when a node has an unexpected number of peers.
type UnexpectedPeerCountError struct {
	nodeName string
	expected int
	got      int
}

func (e *UnexpectedPeerCountError) Error() string {
	return fmt.Sprintf(
		"unexpected peer count: node %q should have %d peers, got: %d",
		e.nodeName,
		e.expected,
		e.got,
	)
}
