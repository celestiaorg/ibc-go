module github.com/cosmos/ibc-go/modules/light-clients/08-wasm

go 1.22.2

toolchain go1.22.3

replace github.com/cosmos/ibc-go/v9 => ../../../

replace (
	cosmossdk.io/x/upgrade => github.com/celestiaorg/cosmos-sdk/x/upgrade v0.2.0
	github.com/cometbft/cometbft => github.com/celestiaorg/celestia-core v0.39.0
	github.com/cosmos/cosmos-sdk => github.com/celestiaorg/cosmos-sdk v1.29.6-sdk-v0.50.14.0.20250812210754-fcb7d156613b // v1.29.6-sdk-v0.50.14
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
)
