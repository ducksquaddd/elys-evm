# Elys EVM

An experimental attempt to add an EVM compatibility layer on top of the existing [Elys](https://github.com/elys-network/elys) Cosmos SDK chain. Unlike traditional solutions like Ethermint/Evmos that run EVM as a sidechain, this approach integrates EVM execution directly on top of the Cosmos runtime, allowing EVM contracts to interact with native Cosmos modules (AMM, perpetuals, staking, etc.) via precompiles.

Built on top of a fork of [cosmos/evm](https://github.com/ducksquaddd/evm) (Interchain Labs' modular EVM framework for Cosmos SDK chains). The fork contains upstream modifications to the JSON-RPC layer to route balance queries through the Cosmos bank module instead of EVM state.

**Status:** Unfinished proof-of-concept. Does not work end-to-end.

## Approach

The goal was to embed a full EVM into Elys as a first-class execution environment rather than running a separate EVM sidechain. The key idea: Ethereum users connect via MetaMask/JSON-RPC while the chain remains a single Cosmos appchain under the hood, with the Cosmos bank module as the single source of truth for all balances.

### Changes in this repo (Elys side)

**Dependency upgrades** — Bumped ibc-go v8 → v10, interchain-security v6 → v7, and integrated `cosmos/evm` as a Go module dependency. This pulled in the EVM and feemarket keepers, JSON-RPC server, and EIP-1559 fee market.

**EVM & feemarket keepers** — Wired `evmkeeper` and `feemarketkeeper` into the Elys app alongside the existing 20+ Cosmos keepers. Added EVM-specific store keys and module registration in `app.go`.

**Native token configuration** — Configured the EVM to use `uelys` (6 decimals) as the native gas token instead of the typical 18-decimal ETH. This required custom `EvmCoinInfo` mapping and base denom registration so the EVM reads balances directly from the Cosmos bank module (`app/config.go`).

**EVM-aware fee checker** — Replaced the default ante handler fee checker with one that routes EVM transactions through gas price calculation (`gasPrice * gasLimit` in `uelys`) while falling back to standard Cosmos fee checking for native transactions (`app/ante/evm_fee_checker.go`).

**Account keeper wrapper** — Wrapped the Cosmos SDK `auth` keeper to satisfy the `cosmos/evm` `AccountKeeper` interface, stubbing out unordered-nonce methods that Elys doesn't need (`app/keepers/evm_wrappers.go`).

**Custom precompiles** — Built two precompiles to bridge EVM <> Cosmos state:
- **Bank precompile** (`0x...0804`) — `getBalance`, `getAllBalances`, `transfer`, `getSupply`, `getAllSupply` operating directly on the Cosmos bank keeper. Handles Ethereum <> Cosmos address conversion so EVM callers can query and move any Cosmos-native denom.
- **Wrapper factory precompile** (`0x...0806`) — `createWrapper` to deploy deterministic (CREATE2) ERC-20 wrapper contracts for Cosmos denoms and register them in the Elys asset profile module.

**Asset profile extensions** — Extended the `x/assetprofile` module with new message types (`MsgCreateWrapper`, `MsgRegisterWrapper`) and keeper methods to store ERC-20 wrapper contract addresses alongside existing denom metadata.

### Changes in the cosmos/evm fork (JSON-RPC side)

The core problem: `cosmos/evm`'s JSON-RPC `eth_getBalance` reads from EVM state, but Elys balances live in the Cosmos bank module. The fork modifies the RPC backend to use the bank keeper as the source of truth.

**Bank keeper injection into RPC backend** — Added a `bankKeeper` field to the RPC `Backend` struct and threaded it through `NewBackend`. The `GetBalance` RPC method was rewritten to call `bankKeeper.GetBalance()` on the Cosmos bank module instead of querying EVM state, with a fallback to the original EVM query path.

**Query context factory** — The RPC server runs outside the normal ABCI request lifecycle, so it doesn't have an SDK context. Added a `QueryContextFactory` abstraction and a `SetupEVMRPC` integration point so the app can provide a context factory at startup, giving the RPC backend access to the bank keeper at any block height.

**6-to-18 decimal scaling** — MetaMask expects 18-decimal native token balances. The fork multiplies bank module balances by 10^12 before returning them over JSON-RPC, converting 6-decimal `uelys` amounts to 18-decimal wei-equivalent values that wallets display correctly.

### What didn't work / was incomplete

Pretty much everything

- The full EVM transaction lifecycle (signing, broadcasting via JSON-RPC, executing in the EVM, committing state) was never completed end-to-end.
- The 6-to-18 decimal scaling was applied at the RPC read layer but not consistently across writes, gas accounting, and precompile interactions, creating potential precision mismatches.
- The `cosmos/evm` framework expects `x/precisebank` for 18-decimal abstraction at the module level, which was not integrated.
- The wrapper factory precompile uses `context.Background()` instead of extracting the actual SDK context from the EVM, so it would panic in production.
- IBC hooks were removed during the dependency upgrade and not re-added.
- No tests were written for the EVM integration layer.
