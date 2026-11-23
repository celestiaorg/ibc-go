# End-to-end tests

This module tests the local IBC v8 checkout using Docker chains and relayers.
The Go module replaces `github.com/cosmos/ibc-go/v8` with `../` and uses the same
Celestia SDK, consensus and upgrade-module replacements as the root module.

## Running locally

Prerequisites: Go 1.24.6 or later, Docker with a running daemon, Bash and `jq`.
`fzf` is optional for interactive test selection. The relayer image must be
available locally or pullable from its registry.

From the repository root:

```sh
make -C e2e build-image
export E2E_CONFIG_PATH="$PWD/e2e/sample.config.yaml"
make -C e2e e2e-test test=TestMsgTransfer_Succeeds_Nonincentivized
```

The image is built from the current working tree and tagged
`ghcr.io/celestiaorg/ibc-go-simd:local`. Nothing is pushed to a registry.
Rebuild it after changing chain code. The sample config runs two chains with
one validator each and Hermes 1.13.1.

To run a suite:

```sh
make -C e2e e2e-suite entrypoint=TestTransferTestSuite
```

To select a test interactively, omit `test` and install `fzf`. To use a different
config, set `E2E_CONFIG_PATH` to its absolute path. Without this variable, the
runner uses `~/.ibc-go-e2e-config.yaml`, creating it from the sample if absent.

`CHAIN_IMAGE`, `CHAIN_A_TAG`, `CHAIN_B_TAG`, `CHAIN_BINARY` and `RELAYER_ID`
override the corresponding config values. When using `IMAGE` or `TAG` with
`build-image`, set these environment variables to the same values before testing.

The Docker tests require the `e2e` build tag. Ordinary `go test ./...` and root
unit-test targets do not run them. IDE invocations must enable this build tag
and use the same config path. To compile and link all suites without executing
any tests or starting containers:

```sh
make -C e2e compile
```

The runner detects Docker daemons whose minimum API version exceeds the older
interchaintest client's default (1.43) and sets `DOCKER_API_VERSION` accordingly.
Explicit overrides are preserved. For direct Go/IDE runs against Docker 29,
set `DOCKER_API_VERSION=1.44` as well.

## CI

[The e2e workflow](../.github/workflows/e2e.yaml) runs on relevant pull requests,
pushes to `main`, and manual dispatch. It builds the checked-out code into a
`simd` image and transfers that image as an artifact to each test job. It uses
read-only repository permissions and needs no registry login or package write
access, including for fork pull requests.

The default matrix includes 29 test methods across 12 suites. It runs each
method separately, with at most eight concurrent jobs. Failed tests upload diagnostics from `e2e/diagnostics`.

Inspect the matrix locally:

```sh
TEST_EXCLUSIONS=TestUpgradeTestSuite go run ./cmd/build_test_matrix | jq
```

Add tests under `e2e/tests`, with `//go:build e2e`. Keep each suite entrypoint and
its test methods in the same file, as required by the matrix generator.
Helper unit tests outside that directory are not included in the Docker matrix.

## Scope and provenance

The previous release-derived `main` retained only stale `e2e/go.mod` and
`e2e/go.sum` files. The harness and tests are restored from upstream history at
`08342291`, before transfer APIs diverged from v8. The current `upstream-main`
suite targets IBC v10 and SDK v0.53, including modules absent from this branch,
so it cannot be copied directly. The interchaintest v8 fork is pinned to the
revision used by that historical harness.

Coverage includes clients, connections, transfers, interchain accounts and
genesis export/import. Fee middleware (ICS-29) and channel-upgrade tests are
excluded from this restored suite. The standalone Wasm suites and
Wasm genesis modifications are excluded because this branch does not contain
the required Wasm simapp.

Chain software-upgrade tests are retained for explicit local runs, but excluded
from the default CI matrix. They require compatible source/target images and a
registered upgrade handler, configured through `upgrade` in the config file or
`CHAIN_UPGRADE_TAG` and `CHAIN_UPGRADE_PLAN`. Historical upstream upgrade paths
are not a validation of Celestia production-chain upgrades.
