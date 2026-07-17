# Deep Dive: armchain-client — Explained Simply
**For personal understanding** | July 2026

This document explains what `armchain-client` does, how it connects consensus (Lachesis DAG) with execution (`fantom-geth`), and how Account Abstraction (AA) and Post-Quantum Cryptography (PQC) are wired at the consensus node level.

---

## Part 1: The Consensus Layer vs. The Execution Layer

To understand `armchain-client`, you must understand that it is a **consensus node**, not an execution client.

### Standard Ethereum (geth)
In vanilla Ethereum, `geth` contains both the execution engine (EVM) and consensus logic (historically PoW, now engine API for PoS).

### ArmChain Architecture
ArmChain uses the **Lachesis consensus algorithm** (from Fantom), which is a leaderless, asynchronous, DAG-based Proof-of-Stake consensus.
- **Consensus (DAG, Epochs, Blocks)** is managed by `armchain-client` (specifically the `gossip/` packages).
- **Execution (EVM, State transitions, Transactions)** is offloaded to the imported `fantom-geth` library.

```
+-------------------------------------------------------------+
|                     armchain-client                         |
|  - Gossip Protocol & Event Blocks (DAG consensus)           |
|  - TxPool & Mempool management                              |
|  - Node admin (RPC interface, P2P network peers)            |
|  - valkeystore (Validator key management)                   |
|                                                             |
|          Go module dependency (replace via go.mod)          |
|                              ↓                              |
|                       fantom-geth                           |
|  - EVM interpreter (Runs smart contracts & custom opcodes)  |
|  - Transactions (Legacy Tx, PQC Tx, AA Tx) definitions      |
|  - Receipt root serialization                               |
+-------------------------------------------------------------+
```

---

## Part 2: Activation of Network Upgrades (Forks)

### `opera/rules.go`
Consensus rules, block time configuration, validator limits, and network upgrades are managed in `opera/rules.go`.

Upgrades on a DAG consensus blockchain cannot simply be defined by block numbers (like Ethereum forks). Instead, they are defined by **epoch heights** or **consensus configurations**.

```go
type Upgrades struct {
    Berlin bool
    London bool // EIP-1559 support
    Llr    bool // Lachesis Light-weight Records
    AA     bool // Native Account Abstraction (EIP-7701)
}
```

### Mapping to `Ethparams.ChainConfig`
When the consensus engine triggers an upgrade, the node must inform the execution layer (the EVM). This is done in the `EvmChainConfig` method:

```go
func (r Rules) EvmChainConfig(hh []UpgradeHeight) *ethparams.ChainConfig {
    cfg := &ethparams.ChainConfig{
        ChainID: r.ChainID,
        ...
    }
    // If the AA upgrade is active in our consensus rules, we enable EIP-7701
    if h.Upgrades.AA && cfg.ArmChain == nil {
        cfg.ArmChain = &ethparams.ArmChainConfig{
            EnableAA: true,
        }
    }
    return cfg
}
```
This maps our consensus upgrade rule (`Upgrades.AA`) to the execution engine's custom field `ArmChain.EnableAA`. Once this flag is active, `fantom-geth` turns on Phase-based execution and custom opcodes.

---

## Part 3: Memory Pool (TxPool) and the AA-Pool

### The Standard TxPool Challenge
A standard EVM TxPool:
1. Receives a transaction.
2. Recovers the sender's address from the signature.
3. Checks the sender's account balance to ensure they can afford the gas fee.
4. Rejects the transaction immediately if the signature is invalid or if the sender's balance is 0.

For Native AA transactions (`0x04` type), this standard behavior breaks because:
- **No ECDSA Signature**: An `AATx` does not have standard ECDSA signature parameters ($V, R, S$).
- **Zero Balance**: The smart account (`Sender`) might not have any funds yet (e.g. if it is a sponsored deployment). Instead, the gas will be sponsored by a `Paymaster`.

### The `AAPool` Sidecar
To resolve this, we added the `AAPool` component (integrated in `evmcore/tx_pool.go`).

When a transaction arrives in the TxPool, `validateTx` checks its type:
```go
if tx.IsAATx() {
    return pool.validateAATx(tx)
}
```

Instead of signature checks, the pool executes `validateAATx`:
1. It ignores standard ECDSA checks.
2. It runs a **lightweight EVM simulation** of the AA validation phase (runs the account's validation phase or checks paymaster balance).
3. If the simulation succeeds, the transaction is admitted into the pool.
4. AA transactions are stored inside `pool.aaPool` (managed by `core/aa_pool.go`).

### Bundle Building (Sorting and Merging)
When the miner builds a new block, it doesn't just pull sequential nonces. It calls `BuildBundle`:
1. It simulates applying AA transactions from the `aaPool` against the current state.
2. It matches nonces for AA senders.
3. It merges AA transactions with standard legacy transactions into a single candidate block list.

---

## Part 4: Validator Keystore & PQC Test Failures

### `valkeystore/`
The consensus node needs to keep validator private keys secure to sign consensus blocks (epochs, events). The `valkeystore` directory manages this key storage.

Unlike standard Ethereum accounts which are managed by users via wallets, validator keys are managed inside the node software.

### Why do some `valkeystore` unit tests fail?
In our codebase:
1. Validator keys have been upgraded to support **ML-DSA** (lattice-based post-quantum cryptography) to secure the consensus engine against quantum threats.
2. The keystore implementation (`valkeystore/files.go` and `valkeystore/mem.go`) explicitly checks:
   ```go
   if pubkey.Type != validatorpk.Types.MLDSA {
       return encryption.ErrNotSupportedType
   }
   ```
   This means the node now **only allows ML-DSA keys** for validators.
3. However, legacy test fixtures inside `valkeystore/common_test.go` still define standard ECDSA public keys (`pubkey1`) and ECDSA private keys (`key1`).
4. When `TestMemKeystoreAdd` and `TestFileKeystoreAdd` try to load these ECDSA key fixtures, the keystore returns `"not supported key type"`, causing the tests to fail.
5. This is a pre-existing testing gap: the test code was not updated when the validator keystore was locked down to only accept PQC/ML-DSA keys. This is unrelated to Account Abstraction.

---

## Part 5: Ingesting System Contracts via Genesis

### How System Contracts get into Block 0
In `armchain-client`, the block 0 genesis configuration is built programmatically.

The package `integration/makegenesis/genesis.go` is responsible for setting up the initial state:
1. It reads the compiled bytecode of our Solidity and Yul contracts from `fantom-geth/core/contracts/` (via generated hex files).
2. It assigns this bytecode to the dedicated system contract addresses (`0x0000...00007700` through `0x77B4`).
3. It sets up the initial balance and storage variables for `AAStakingRegistry` and `FactoryRegistry`.
4. It compiles this entire state into the embedded genesis JSON, ensuring that the moment the blockchain starts, the system contracts exist and are ready for use.
