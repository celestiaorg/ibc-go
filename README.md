# celestiaorg/ibc-go

This is Celestia's fork of [cosmos/ibc-go](https://github.com/cosmos/ibc-go). [celestia-app](https://github.com/celestiaorg/celestia-app) uses it through `replace` directives in its `go.mod`.

For general IBC and ibc-go documentation, see the upstream repository and [docs](https://ibc.cosmos.network).

## Why fork?

- celestia-app runs on [celestia-core](https://github.com/celestiaorg/celestia-core) and [celestiaorg/cosmos-sdk](https://github.com/celestiaorg/cosmos-sdk), not on upstream CometBFT and Cosmos SDK. ibc-go must build against those forks.
- celestia-app uses ibc-go v8. Celestia ships fixes to that release line here, on its own schedule.
- Some fixes change state transitions. Celestia releases them as new fork versions so that celestia-app can activate them at a coordinated network upgrade.

## Maintained branches

| Branch | Tags | Used by |
| ------ | ---- | ------- |
| [`release/v8.8.x`](https://github.com/celestiaorg/ibc-go/tree/release/v8.8.x) | `v8.8.0-celestia` | celestia-app `main` |
| [`release/v8.7.x`](https://github.com/celestiaorg/ibc-go/tree/release/v8.7.x) | `v8.7.1` to `v8.7.4` | celestia-app `v6.x` to `v11.x` |
| [`capability/release/v1.0.x`](https://github.com/celestiaorg/ibc-go/tree/capability/release/v1.0.x) | `modules/capability/v1.0.2` | celestia-app `v10.x` and later |
| [`v6.2.x`](https://github.com/celestiaorg/ibc-go/tree/v6.2.x) | `v6.3.0` | celestia-app `v3.x` (now only for sync from genesis) |
| `main` | none | Default branch. Created from `release/v8.8.x`. Fixes land here first, then go to the release branches. |
| `upstream-main` | none | The old `main`, which followed upstream `main`. Not maintained. |

Fork tags are not upstream tags. For example, `v8.8.0-celestia` is based on upstream `v8.7.0`, not upstream `v8.8.0`.

## Differences from upstream

### `release/v8.7.x` and `release/v8.8.x`

Based on upstream `v8.7.0`.

- **celestia-core and celestiaorg/cosmos-sdk**: `go.mod` uses the Celestia forks. Test code implements `SignRawBytes`, which the celestia-core `PrivValidator` interface requires ([#7](https://github.com/celestiaorg/ibc-go/pull/7)).
- **Genesis export and import**: `ExportGenesis` no longer writes the `09-localhost` client, which `InitGenesis` rejects. Channel upgrade state is now exported, imported and validated in genesis, so an exported state can be imported again ([#12](https://github.com/celestiaorg/ibc-go/pull/12), [#13](https://github.com/celestiaorg/ibc-go/pull/13), [#14](https://github.com/celestiaorg/ibc-go/pull/14), [#15](https://github.com/celestiaorg/ibc-go/pull/15)).
- **Transfer error text**: the non-positive amount error prints the amount, not a pointer address ([#21](https://github.com/celestiaorg/ibc-go/pull/21)).
- **Transfer unescrow** (`release/v8.8.x` only, consensus-breaking): a returning packet with an invalid base denom gets an error acknowledgement instead of a panic ([#22](https://github.com/celestiaorg/ibc-go/pull/22)).

### `capability/release/v1.0.x`

Based on upstream `modules/capability/v1.0.1`.

- The in-memory capability map is safe for concurrent use and for cached (branched) contexts ([#9](https://github.com/celestiaorg/ibc-go/pull/9)).

### `v6.2.x`

Based on upstream `release/v6.2.x`.

- `AcknowledgePacket` rejects a JSON acknowledgement whose bytes change when it is decoded and encoded again, so that only the canonical encoding is accepted ([#3](https://github.com/celestiaorg/ibc-go/pull/3)).
- `testing/simapp` does not panic when the user home directory is not available, so celestia-app can run in the Vercel Go runtime ([#4](https://github.com/celestiaorg/ibc-go/pull/4)).
- `go.mod` uses the Celestia forks of Tendermint and Cosmos SDK ([#6](https://github.com/celestiaorg/ibc-go/pull/6)).

## Contributing

- Open PRs against `main`, then backport them to the release branches that celestia-app uses (`release/v8.8.x`, `release/v8.7.x`).
- PR titles use [conventional commits](https://www.conventionalcommits.org/) with a review level tag (`[low]`, `[medium]` or `[high]`). Add `!` for a consensus-breaking change.
- A consensus-breaking change needs a new fork version. Do not add it to a tag that a released celestia-app version already uses.
