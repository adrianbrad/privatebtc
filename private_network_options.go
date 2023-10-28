package privatebtc

import (
	"io"
	"log/slog"
	"time"
)

type options struct {
	rpcUser              string
	rpcPass              string
	fallbackFee          float64
	walletName           *string
	bitcoinClientVersion string
	nodeNamePrefix       string
	timeout              *time.Duration
	handler              slog.Handler
}

func defaultOptions() *options {
	return &options{
		rpcUser:              "foo",
		rpcPass:              "DnDEgWFCM2K0aZW3KMwswb7PyKdGr_cTJy9u9tIEypU=",
		fallbackFee:          0.01,
		walletName:           nil,
		bitcoinClientVersion: "latest",
		nodeNamePrefix:       "node_",
		timeout:              nil,
		handler:              slog.NewTextHandler(io.Discard, nil),
	}
}

// An Option configures a parameter for the Bitcoin Private Network.
// The functional options are implemented following uber guidelines.
// https://github.com/uber-go/guide/blob/master/style.md#functional-options
type Option interface {
	apply(*options)
}

type bitcoinClientVersion string

func (v bitcoinClientVersion) apply(opts *options) {
	opts.bitcoinClientVersion = string(v)
}

// WithBitcoinClientVersion configures the version of the bitcoin client to use.
func WithBitcoinClientVersion(version string) Option {
	return bitcoinClientVersion(version)
}

type withWallet string

func (w withWallet) apply(opts *options) {
	s := string(w)

	opts.walletName = &s
}

// WithWallet configures and creates a wallet for the nodes.
func WithWallet(walletName string) Option {
	return withWallet(walletName)
}

type withTimeout time.Duration

func (w withTimeout) apply(opts *options) {
	opts.timeout = (*time.Duration)(&w)
}

// WithTimeout configures the timeout for the docker operations.
func WithTimeout(timeout time.Duration) Option {
	return withTimeout(timeout)
}

type withNodeNamePrefix string

func (w withNodeNamePrefix) apply(opts *options) {
	opts.nodeNamePrefix = string(w)
}

// WithNodeNamePrefix configures the prefix for the node names.
func WithNodeNamePrefix(nodeNamePrefix string) Option {
	return withNodeNamePrefix(nodeNamePrefix)
}

type withRPCAuth struct {
	rpcUser     string
	rpcPassword string
}

func (w withRPCAuth) apply(opts *options) {
	opts.rpcUser = w.rpcUser
	opts.rpcPass = w.rpcPassword
}

// WithRPCAuth configures the RPC authentication for the Bitcoin Private Network.
func WithRPCAuth(rpcUser, rpcPassword string) Option {
	return withRPCAuth{
		rpcUser:     rpcUser,
		rpcPassword: rpcPassword,
	}
}

type withSlogHandler struct {
	handler slog.Handler
}

func (w withSlogHandler) apply(opts *options) {
	opts.handler = w.handler
}

// WithSlogHandler configures the handler for the Bitcoin Private Network.
func WithSlogHandler(handler slog.Handler) Option {
	return withSlogHandler{handler: handler}
}
