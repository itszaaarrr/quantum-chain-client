# ArmChain Client — AA Devnet Guide
**Branch**: `kz/feat/devnet-aa-support` | **Last updated**: July 7, 2026

This is the single reference doc for the `armchain-client` AA + devnet branch. `armchain-client` is the **deployable node** — it embeds `fantom-geth` (the execution layer) as a Go module dependency.

---

## 1. Project Layout at a Glance

- **`evmcore/`** — EVM state transitions, tx pool, AA pool integration
- **`gossip/`** — Lachesis consensus engine, event propagation, E2E tests
- **`inter/`** — CSER serializer for PQC and AA transaction types
- **`opera/`** — Network rules, chain config, genesis allocation
- **`integration/makegenesis/`** — Genesis state builder (injects system contracts)

The node binary is `build/opera`, built by `make opera`.

---

## 2. Network Rules & AA Activation

Network rules live in [`opera/rules.go`](file:///home/khizar/armchain/armchain-client/opera/rules.go). The `Upgrades` struct has an `AA` boolean that gates EIP-7701 support. When `AA: true`, the EVM chain config gets populated with `ArmChainConfig{EnableAA: true}`, which turns on the AA opcode set and validates type `0x04` transactions.

Two environments are supported:

| Environment | Flag | AA Enabled | Notes |
|---|---|---|---|
| **FakeNet** | `--fakenet 1/1` | Yes | Local single-validator, accelerated params |
| **DevNet** | *(multi-node)* | Yes | 5-VM staging cluster, wire-compatible PQC |

---

## 3. AA Transaction Pool

Standard TxPool would reject AA transactions (no ECDSA signature). Instead, `evmcore/tx_pool.go` routes them through an `AAPool` side-car:

- **Validation**: `validateTx` calls `validateAATx` for any `tx.IsAATx()` — skips signature checks, runs simulation instead
- **Add**: AA txs bypass the standard pending queue and go into `pool.aaPool.Add(tx)`
- **Block building**: `BuildBundle(stateSnapshot)` returns ordered, simulation-verified AA txs to the block builder

---

## 4. AA Test Suite

```bash
cd /home/khizar/armchain/armchain-client
go test ./... -count=1
```

Key test files:

| File | What It Tests |
|---|---|
| `gossip/aa_e2e_test.go` | Full AA wallet deploy + execute end-to-end |
| `gossip/aa_pipeline_integration_test.go` | Paymaster validation + ingress pipeline |
| `evmcore/aa_txpool_test.go` | Pool simulation rules, banning, caps |

Key test cases:
- **`TestAA_SystemContractsInGenesis`** — verifies all system contracts (`0x7700–0x7704`, factories, impls) are in genesis
- **`TestAA_GuardianWallet_Deploy/Execute`** — deploys and executes via type `0x04` tx
- **`TestAA_FoundationPaymaster_SponsorsDeploy`** — sponsor pays for first-time wallet deployment

> **Note**: `valkeystore` tests fail on non-MLDSA key types — this is a pre-existing gap (keystore only supports MLDSA validator keys; test fixtures use ECDSA keys). Unrelated to AA changes.

---

## 5. Build & Run

```bash
# Build
make opera

# Run single-node fakenet (local dev)
./build/opera --fakenet 1/1 --http --http.port 4000 --http.addr "127.0.0.1" \
  --http.corsdomain "*" \
  --http.api "eth,debug,net,admin,web3,personal,txpool,ftm,dag" \
  --allow-insecure-unlock

# Multi-node local demo
cd demo && N=3 ./start.sh

# Clean all state
rm -rf ~/.opera /tmp/fakenet1 ~/keystore
```

---

## 6. Console Quick Reference

```bash
# Attach
./build/opera attach http://localhost:4000

# One-liner attach
./build/opera attach --exec "ftm.blockNumber" http://127.0.0.1:4000
```

```js
// Accounts & balances
ftm.accounts
ftm.getBalance(ftm.accounts[0])
eth.getTransactionCount(ftm.accounts[0])   // nonce

// Send
ftm.sendTransaction({from: ftm.accounts[0], to: ftm.accounts[1], value: '1000000000'})
eth.sendRawTransaction('0x<SIGNED_HEX>')

// Blocks & receipts
ftm.getBlock('latest')
eth.getTransaction('0x...')
eth.getTransactionReceipt('0x...')

// Txpool
txpool.status
txpool.inspect

// AA — check if smart account is deployed
ftm.getCode('0x<SENDER_ADDRESS>')

// AA — trace a failed AA tx
debug.traceTransaction('0x...')

// Peers
admin.peers
admin.addPeer('enr:...')
```

---

## 7. Dependency on fantom-geth

`go.mod` uses a local `replace` directive:
```
replace github.com/ethereum/go-ethereum => ../fantom-geth
```

Switch to `kz/feat/aa-support` in `fantom-geth` before building. No `go mod tidy` changes needed unless you add imports.

---

## 8. Next Steps

- Finalize devnet genesis config (align `DevNetRules()` with chain config in `fantom-geth`)
- Bootstrap devnet validation nodes (scripts in `demo/`)
- Run E2E AA test against live devnet via `armchain-ethersv6` test scripts
