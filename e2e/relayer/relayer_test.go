//go:build e2e

package relayer_test

import (
	"context"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/cosmos/ibc-go/e2e/relayer"
	"github.com/cosmos/ibc-go/e2e/testsuite"
)

// TestHermesConfiguration checks config generation and Docker volume permissions
// against the configured Hermes image without starting any chains.
func TestHermesConfiguration(t *testing.T) {
	t.Setenv(testsuite.E2EConfigFilePathEnv, "../sample.config.yaml")
	var cfg relayer.Config
	for _, candidate := range testsuite.LoadConfig().RelayerConfigs {
		if candidate.ID == relayer.Hermes {
			cfg = candidate
			break
		}
	}
	require.NotEmpty(t, cfg.Image)

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	client, network := interchaintest.DockerSetup(t)
	r := relayer.New(t, cfg, zaptest.NewLogger(t), client, network)
	reporter := testreporter.NewNopReporter().RelayerExecReporter(t)
	err := r.AddChainConfiguration(ctx, reporter, ibc.ChainConfig{
		ChainID:       "hermes-config-1",
		Bech32Prefix:  "cosmos",
		Denom:         "stake",
		GasPrices:     "0.01stake",
		GasAdjustment: 1.3,
	}, "relayer", "http://localhost:26657", "localhost:9090")
	require.NoError(t, err)
}
