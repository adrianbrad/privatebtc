package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// go build -ldflags "-X main.version=x.y.z".
var (
	// version is the version of the project.
	version = "0.0.0"
	// revision is the git short commit revision number.
	revision = "-"
	// time is the build time of the project.
	time = "-"
)

var rootCMD = &cobra.Command{
	Use:   "privatebtc",
	Short: "Start a bitcoin private network with a terminal user interface",
	Run: func(*cobra.Command, []string) {
		const nodes = 3

		loggerHandler := slog.NewTextHandler(os.Stdout, nil)

		if !envCheck(loggerHandler) {
			return
		}

		logger := slog.New(loggerHandler)

		logger.Info(
			"starting PrivateBTC TUI",
			slog.Int("nodes", nodes),
			slog.String("version", version),
			slog.String("revision", revision),
			slog.String("build_time", time),
		)

		if err := runTUI(nodes, loggerHandler); err != nil {
			logger.Error("run error", "err", err)
		}
	},
}

func init() {
	rootCMD.AddCommand(envcheckCMD)
}

func main() {
	if err := rootCMD.Execute(); err != nil {
		log.Fatalf("execute command error: %s", err)
	}
}
