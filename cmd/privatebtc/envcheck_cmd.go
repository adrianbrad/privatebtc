package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/adrianbrad/privatebtc/docker"
	"github.com/spf13/cobra"
)

// envcheckCMD is Cobra command for the envcheck subcommand.
// It checks the environment dependencies.
var envcheckCMD = &cobra.Command{
	Use:   "envcheck",
	Short: "run a environment dependency check",
	Run: func(cmd *cobra.Command, args []string) {
		loggerHandler := slog.NewTextHandler(os.Stdout, nil)

		_ = envCheck(loggerHandler)
	},
}

func envCheck(loggerHandler slog.Handler) bool {
	logger := slog.New(loggerHandler)

	dockerClient, err := docker.NewClient()
	if err != nil {
		logger.Error("❌ new docker client error", slog.String("error", err.Error()))
		return false
	}

	var (
		wg       sync.WaitGroup
		errCount atomic.Int64
	)

	const totalChecks = 2

	wg.Add(totalChecks)

	ctx := context.Background()

	if runtime.GOOS == "windows" {
		logger.Error("❌ windows system detected", slog.String(
			"details",
			fmt.Sprintf("due to the fact that windows is not supported by the "+
				"%s docker image, "+
				"this program is not able to run on windows", docker.BitcoinImage),
		))

		return false
	}

	go func() {
		defer wg.Done()

		if _, err := dockerClient.Ping(ctx); err != nil {
			logger.Error(
				"❌ docker daemon check failed",
				slog.String("error", err.Error()),
			)

			errCount.Add(1)

			return
		}

		logger.Info(
			"✅ docker daemon is found and running",
			slog.String("host", dockerClient.DaemonHost()),
		)
	}()

	go func() {
		defer wg.Done()

		logger := logger.With(slog.String("image", docker.BitcoinImage))

		if err := docker.CheckImageExistsInLocalCache(
			ctx,
			dockerClient,
			docker.BitcoinImage,
		); err != nil {
			if !errors.Is(err, &docker.ImageNotFoundError{Image: docker.BitcoinImage}) {
				logger.Error(
					"❌ bitcoin image exists check failed",
					slog.String("error", err.Error()),
				)

				errCount.Add(1)

				return
			}

			logger.Info(
				"⚠️ bitcoin image not found in local cache, pulling it. Please wait...",
			)

			if err := docker.PullImage(ctx, dockerClient, docker.BitcoinImage); err != nil {
				logger.Error(
					"❌ bitcoin image pull failed",
					slog.String("error", err.Error()),
				)

				errCount.Add(1)

				return
			}

			logger.Info("✅ bitcoin image pulled successfully")

			return
		}

		logger.Info("✅ bitcoin image exists in local cache")
	}()

	wg.Wait()

	if c := errCount.Load(); c > 0 {
		logger.Error(
			"❌ environment check failed",
			slog.Int64("error count", c),
		)

		return false
	}

	logger.Info("✅ environment check passed")

	return true
}
