package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/adrianbrad/privatebtc"
	"github.com/adrianbrad/privatebtc/btcsuite"
	"github.com/adrianbrad/privatebtc/docker/testcontainers"
	"github.com/adrianbrad/privatebtc/tview"
)

func runTUI(nodes int, loggerHandler slog.Handler) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	btcpn, err := privatebtc.NewPrivateNetwork(
		&testcontainers.NodeService{
			SlogHandler: loggerHandler,
		},
		btcsuite.RPCClientFactory{},
		nodes,
		privatebtc.WithWallet("tui"),
		privatebtc.WithSlogHandler(loggerHandler),
	)
	if err != nil {
		return fmt.Errorf("create bitcoin private network error: %w", err)
	}

	if err := btcpn.Start(ctx); err != nil {
		return fmt.Errorf("start bitcoin private network error: %w", err)
	}

	// nolint: errcheck
	defer btcpn.Close()

	if err := tview.NewTUI(btcpn, version).Run(); err != nil {
		return fmt.Errorf("run tui error: %w", err)
	}

	if err := btcpn.Close(); err != nil {
		return fmt.Errorf("close bitcoin private network error: %w", err)
	}

	return nil
}
