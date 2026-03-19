# E2E Benchmark

Performance benchmark for ProxyMempool + PriorityMempool + MemIAVL on MiniWasm. Measures throughput, latency, and mempool behavior
across optimization layers using a multi-node cluster with production-realistic settings.

## Cluster topology

4-node cluster: 1 sequencer + 3 fullnodes on localhost with deterministic port allocation.

- **Fullnode submission**: Benchmark load is submitted to edge (non-validator) nodes (indices 1-3), testing gossip propagation to the sequencer.
- **Block interval**: 100ms (sequencing default, `CreateEmptyBlocks = false` thus blocks only created when txs exist)
- **Gas price**: 0umin
- **Queued tx extension**: All tx submissions include `--allow-queued` flag (required for `ExtensionOptionQueuedTx`)

## Comparison matrix

### 1. Mempool comparison: CList vs Proxy+Priority

Three load patterns: sequential tests give a fair TPS comparison (both mempools handle in-order correctly), burst tests demonstrate CList's tx-drop problem.

**Baselines** (run with v1.2.10 binary):

| Test | Load | Config |
|---|---|---|
| `TestBenchmarkBaselineSeq` | Sequential, bank send | 10 accts x 200 txs |
| `TestBenchmarkBaselineBurst` | Burst, bank send | 10 accts x 200 txs |
| `TestBenchmarkBaselineSeqWasmExec` | Sequential, Wasm exec | 100 accts x 50 txs, 30 writes/tx |

**Comparisons** (3-way: CList vs Proxy+IAVL vs Proxy+MemIAVL):

| Test | Load | Purpose |
|---|---|---|
| `TestBenchmarkSeqComparison` | Sequential, bank send | Fair TPS comparison, lightweight workload |
| `TestBenchmarkSeqComparisonWasmExec` | Sequential, Wasm exec | Fair TPS comparison under heavy state pressure |
| `TestBenchmarkBurstComparison` | Burst, bank send | Inclusion rate (CList drops txs) |

### 2. State db comparison: IAVL vs MemIAVL

Both use Proxy+Priority mempool. Isolates state storage impact.

| Test | Workload                           | Config |
|---|------------------------------------|---|
| `TestBenchmarkMemIAVLBankSend` | Bank sends                         | 100 accts x 200 txs |
| `TestBenchmarkMemIAVLWasmExec` | Wasm exec (ExecuteMsg::WriteMixed) | 100 accts x 50 txs, 30 writes/tx |

### 3. Pre-signed HTTP broadcast (saturated chain)

Bypasses CLI bottleneck. Txs are generated+signed offline, then POSTed via HTTP to `/broadcast_tx_sync`.

| Test | Load | Config |
|---|---|---|
| `TestBenchmarkPreSignedSeqComparison` | Sequential, bank send, HTTP | 20 accts x 100 txs |
| `TestBenchmarkPreSignedBurstComparison` | Burst, bank send, HTTP | 20 accts x 100 txs |
| `TestBenchmarkPreSignedSeqWasmExec` | Sequential, Wasm exec, HTTP | 20 accts x 100 txs, 100 writes/tx |
| `TestBenchmarkPreSignedSeqWasmExecStress` | Sequential, Wasm exec, HTTP (stress) | 20 accts x 200 txs, 200 writes/tx |

### 4. Capability demos

| Test | What | Config |
|---|---|---|
| `TestBenchmarkQueuePromotion` | Out-of-order nonce handling, 100% inclusion | 10 accts x 50 txs |
| `TestBenchmarkGossipPropagation` | Gossip across nodes | 5 accts x 50 txs |
| `TestBenchmarkQueuedFlood` | Future-nonce burst (nonce gaps), queued pool stress + promotion cascade | 10 accts x 50 txs |
| `TestBenchmarkQueuedGapEviction` | Gap TTL eviction under sustained load | 10 accts x 50 txs |

## Expected outcomes

1. **Sequential (fair comparison)**: CList and Proxy+Priority both handle in-order nonces correctly, so sequential submission should show similar TPS. This is the control that proves Proxy+Priority doesn't regress on the happy path.
2. **Burst (stress test)**: Proxy+Priority >> CList. Under burst, CList's `reCheckTx` and cache-based dedup cause it to silently drop txs, while Proxy+Priority's queued pool absorbs out-of-order arrivals and achieves 100% inclusion.
3. **Heavy state writes**: MemIAVL > IAVL. Lightweight workloads (bank send) won't show a difference because the state db isn't the bottleneck. Heavy Wasm exec with many writes per tx is needed, and the chain must be saturated (pre-signed HTTP) so the state db becomes the limiting factor.
4. **Combined (Proxy+Priority+MemIAVL)**: Best overall throughput and latency, the mempool improvement eliminates tx drops, and MemIAVL reduces state commit time under heavy writes.

## Results

### 1. Mempool comparison: CList (v1.2.10) vs Proxy+Priority

#### Sequential bank send

| Config | Variant | TPS | vs base | P50ms | vs base | P95ms | vs base | Included | Peak MP |
|---|---|---:|--------:|---:|--------:|---:|--------:|---:|---:|
| clist/iavl/seq | baseline | 49.3 |       - | 1090 |       - | 1987 |       - | 2000/2000 | 38 |
| proxy+priority/iavl/seq | mempool-only | 47.9 |   -2.8% | 281 |  -74.2% | 363 |  -81.7% | 2000/2000 | 12 |
| proxy+priority/memiavl/seq | combined | 47.4 |   -3.9% | 283 |  -74.0% | 395 |  -80.1% | 2000/2000 | 10 |

#### Burst bank send

| Config | Variant | TPS | vs base | P50ms | vs base | P95ms | vs base | Included | Peak MP |
|---|---|---:|--------:|---:|--------:|---:|--------:|---:|---:|
| clist/iavl/burst | baseline | 13.0 |       - | 359 |       - | 2113 |       - | 33/2000 | 9 |
| proxy+priority/iavl/burst | mempool-only | 51.2 | +293.8% | 269 |  -25.1% | 370 |  -82.5% | 2000/2000 | 80 |
| proxy+priority/memiavl/burst | combined | 45.3 | +248.5% | 298 |  -17.0% | 400 |  -81.1% | 2000/2000 | 11 |

#### Sequential Wasm exec (BenchHeavyState, 30 unique-key writes/tx)

| Config | Variant | TPS | vs base | P50ms | vs base | P95ms | vs base | Included | Peak MP |
|---|---|---:|--------:|---:|--------:|---:|--------:|---:|---:|
| clist/iavl/seq-wasm-exec | baseline | 44.9 |       - | 2697 |       - | 3816 |       - | 5000/5000 | 738 |
| proxy+priority/iavl/seq-wasm-exec | mempool-only | 51.2 |  +14.0% | 2072 |  -23.2% | 2716 |  -28.8% | 5000/5000 | 80 |
| proxy+priority/memiavl/seq-wasm-exec | combined | 51.9 |  +15.6% | 2006 |  -25.6% | 2583 |  -32.3% | 5000/5000 | 41 |

### 2. State db comparison: IAVL vs MemIAVL (CLI-based, Proxy+Priority)

| Config | Workload | TPS | P50ms | P95ms | P99ms | Included | Peak MP |
|---|---|---:|---:|---:|---:|---:|---:|
| memiavl-compare/iavl/bank-send | bank send | 45.5 | 2287 | 2977 | 3256 | 20000/20000 | 44 |
| memiavl-compare/memiavl/bank-send | bank send | 43.7 | 2350 | 3021 | 3325 | 20000/20000 | 33 |
| memiavl-compare/iavl/wasm-exec | wasm exec | 50.7 | 2125 | 2731 | 3016 | 5000/5000 | 80 |
| memiavl-compare/memiavl/wasm-exec | wasm exec | 50.7 | 2039 | 2668 | 2930 | 5000/5000 | 45 |

CLI-based tests are bottlenecked by CLI overhead (~50 TPS ceiling), masking IAVL vs MemIAVL throughput differences. MemIAVL shows improvement in peak mempool size (-44% for wasm exec). See pre-signed HTTP results below for saturated-chain comparison.

### 3. Pre-signed HTTP broadcast (saturated chain)

#### Bank send (IAVL vs MemIAVL)

| Config | TPS | P50ms | P95ms | P99ms | Included | Peak MP |
|---|---:|---:|---:|---:|---:|---:|
| presigned/iavl/seq | 1759.6 | 197 | 284 | 308 | 2000/2000 | 415 |
| presigned/memiavl/seq | 2097.7 | 225 | 339 | 357 | 2000/2000 | 535 |
| presigned/iavl/burst | 2140.4 | 517 | 763 | 768 | 2000/2000 | 1198 |
| presigned/memiavl/burst | 2115.9 | 510 | 796 | 802 | 2000/2000 | 1436 |

#### Wasm exec (IAVL vs MemIAVL, 100 unique-key writes/tx)

| Config | TPS | P50ms | P95ms | P99ms | Included | Peak MP |
|---|---:|---:|---:|---:|---:|---:|
| presigned/iavl/seq-wasm-exec | 464.3 | 1486 | 2710 | 2766 | 2000/2000 | 1136 |
| presigned/memiavl/seq-wasm-exec | 815.2 | 844 | 1422 | 1440 | 2000/2000 | 1183 |

#### Wasm exec stress (IAVL vs MemIAVL, 200 unique-key writes/tx, 4000 txs)

| Config | TPS | P50ms | P95ms | P99ms | Included | Peak MP |
|---|---:|---:|---:|---:|---:|---:|
| presigned-stress/iavl/seq-wasm-exec | 239.8 | 3854 | 8507 | 8935 | 2949/4000 | 2627 |
| presigned-stress/memiavl/seq-wasm-exec | 535.2 | 1934 | 3583 | 3719 | 2842/4000 | 2682 |

Under saturated heavy state writes with continuously growing state tree, MemIAVL demonstrates decisive superiority.
At 2000 txs (100 writes/tx): **+75.6% TPS** with 100% inclusion for both.
At 4000 txs (200 writes/tx): **+123.2% TPS** (535 vs 240) and **-49.8% P50 latency** (1934 vs 3854ms).
IAVL degrades sharply as the state tree grows while MemIAVL maintains consistent throughput.

### 4. Capability demos

| Test |   TPS | P50ms | P95ms | P99ms | Included | Peak MP | Notes |
|---|------:|------:|------:|------:|---------:|--------:|---|
| queue-promotion |  48.9 |   281 |   371 |   734 |  500/500 |      19 | Out-of-order nonces, 100% inclusion |
| gossip |  33.4 |     - |     - |     - |  250/250 |       - | All txs to single node, gossip to validator |
| queued-flood | 481.6 |  7316 | 11854 | 12476 |  500/500 |     490 | Nonce gap burst + promotion cascade |
| queued-gap-eviction |     - |     - |     - |     - |        - |       - | Qualitative: gap TTL eviction confirmed, mempool drained |

## Run

All commands assume `cd integration-tests` first. The full workflow has 3 phases:
baselines first, then current-branch benchmarks, then the comparison tests that
load both result sets. Capability demos / queued tests are standalone and can run
any time.

### Prerequisites: Build the CosmWasm contract

The Wasm exec benchmarks require a compiled CosmWasm contract.

```bash
cd e2e/benchmark/bench_heavy_state
cargo build --target wasm32-unknown-unknown --release
```

### Phase 1: Collecting baselines (CList mempool)

Build the pre-proxy binary once, then run the baseline tests.
Results are written to `e2e/benchmark/results/` as JSON keyed by label.

```bash
# Build pre-proxy binary
git checkout tags/v1.2.10
go build -o build/minitiad-baseline ./cmd/minitiad
git checkout -   # return to current branch

cd integration-tests

# Sequential bank send baseline
E2E_MINITIAD_BIN="$(pwd)/../build/minitiad-baseline" \
  go test -v -tags benchmark -run TestBenchmarkBaselineSeq -timeout 30m -count=1 ./e2e/benchmark/

# Burst bank send baseline
E2E_MINITIAD_BIN="$(pwd)/../build/minitiad-baseline" \
  go test -v -tags benchmark -run TestBenchmarkBaselineBurst -timeout 30m -count=1 ./e2e/benchmark/

# Sequential Wasm exec baseline
E2E_MINITIAD_BIN="$(pwd)/../build/minitiad-baseline" \
  go test -v -tags benchmark -run TestBenchmarkBaselineSeqWasmExec -timeout 60m -count=1 ./e2e/benchmark/
```

### Phase 2: Running current-branch benchmarks

These use the current binary (auto-built or via `E2E_MINITIAD_BIN`).
Each test writes its own result JSON.

```bash
# Build current binary
go build -o ./minitiad ../cmd/minitiad

# State db comparison (IAVL vs MemIAVL)
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkMemIAVLBankSend -timeout 60m -count=1 ./e2e/benchmark/
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkMemIAVLWasmExec -timeout 60m -count=1 ./e2e/benchmark/

# Capability demos (standalone, no baselines needed)
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkQueuePromotion -timeout 30m -count=1 ./e2e/benchmark/
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkGossipPropagation -timeout 30m -count=1 ./e2e/benchmark/

# Queued mempool behavior (standalone, no baselines needed)
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkQueuedFlood -timeout 30m -count=1 ./e2e/benchmark/
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkQueuedGapEviction -timeout 30m -count=1 ./e2e/benchmark/
```

### Pre-signed HTTP broadcast tests (saturated chain)

These use pre-signed txs via HTTP to saturate the chain, bypassing the CLI bottleneck.

```bash
# Sequential bank send (IAVL vs MemIAVL)
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkPreSignedSeqComparison -timeout 20m -count=1 ./e2e/benchmark/

# Burst bank send (IAVL vs MemIAVL)
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkPreSignedBurstComparison -timeout 20m -count=1 ./e2e/benchmark/

# Sequential Wasm exec (IAVL vs MemIAVL)
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkPreSignedSeqWasmExec$ -timeout 30m -count=1 ./e2e/benchmark/

# Sequential Wasm exec stress (IAVL vs MemIAVL, 4000 txs)
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkPreSignedSeqWasmExecStress -timeout 30m -count=1 ./e2e/benchmark/
```

### Phase 3: Comparison tests (baseline vs current)

These load baseline JSONs from `e2e/benchmark/results/` by label and run Proxy+IAVL
and Proxy+MemIAVL variants, then print a side-by-side comparison table with deltas.

```bash
# Sequential bank send: CList vs Proxy+IAVL vs Proxy+MemIAVL
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkSeqComparison$ -timeout 30m -count=1 ./e2e/benchmark/

# Sequential Wasm exec: CList vs Proxy+IAVL vs Proxy+MemIAVL
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkSeqComparisonWasmExec -timeout 60m -count=1 ./e2e/benchmark/

# Burst bank send: CList vs Proxy+IAVL vs Proxy+MemIAVL
E2E_MINITIAD_BIN=./minitiad \
  go test -v -tags benchmark -run TestBenchmarkBurstComparison -timeout 30m -count=1 ./e2e/benchmark/
```

Each Phase 3 test prints a comparison table like:

```
Config                    | Variant      |     TPS | vs base |   P50ms | vs base |   P95ms | vs base | Peak Mempool
clist/iavl/seq            | baseline     |   120.5 |       - |    2500 |       - |    4800 |       - |         1950
proxy+priority/iavl/seq   | mempool-only |   245.3 | +103.6% |    1823 |  -27.1% |    3412 |  -28.9% |         1847
proxy+priority/memiavl/seq| combined     |   312.7 | +159.5% |    1401 |  -44.0% |    2845 |  -40.7% |         1823
```

## Configuration

### Ground Rules

1. Baseline requires a separate binary built from v1.2.10 (pre-proxy CometBFT, pre-ABCI++ changes).
2. Run baseline and current benchmarks on the same machine.
3. Warmup runs before every measured load (5 txs, metadata re-queried after).
4. TPS is derived from block timestamps, not submission wall clock.
5. Latency = `block_time - submit_time` (covers mempool wait, gossip, proposal, execution).
6. Load is submitted to edge nodes (non-validator) to test realistic gossip propagation.

### Configurable mempool limits

These can be tuned in `app.toml` under `[abcipp]` (defaults shown):

| Parameter | Default | Description |
|---|---|---|
| `max-queued-per-sender` | 64 | Max queued txs per sender |
| `max-queued-total` | 1024 | Max queued txs globally |
| `queued-gap-ttl` | 60s | TTL for stalled senders missing head nonce |

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `E2E_MINITIAD_BIN` | (auto-build) | Path to prebuilt `minitiad` binary |
| `BENCHMARK_RESULTS_DIR` | `results/` | Output directory for JSON results |

## Structure

```
benchmark/
  config.go          Variant definitions, BenchConfig, preset constructors
  load.go            Load generators
  collector.go       MempoolPoller, CollectResults, latency aggregation
  report.go          JSON output, comparison tables, delta calculations, LoadBaselineResultsByLabel
  benchmark_test.go  Test suite (build-tagged `benchmark`)
  bench_heavy_state/ CosmWasm contract (BenchHeavyState, write_mixed)
  results/           JSON output directory
```

### Load generators

All load generators route transactions to fullnodes when `ValidatorCount > 0`.

- **BurstLoad**: All accounts submit concurrently with sequential nonces, round-robin across fullnodes.
- **SequentialLoad**: Accounts run concurrently, but each account sends txs one-at-a-time. Each account pinned to a single fullnode.
- **OutOfOrderLoad**: First 3 txs per account use `[seq+2, seq+0, seq+1]` to test queue promotion.
- **SingleNodeLoad**: All txs to a single node for gossip propagation measurement.
- **WasmExecBurstLoad**: Like BurstLoad but calls `ExecuteWasmContract` (`write_mixed`) instead of bank sends.
- **WasmExecSequentialLoad**: Like SequentialLoad but calls `ExecuteWasmContract`. Each account pinned to a single fullnode.
- **QueuedFloodLoad**: Sends txs with nonces `[base+1..base+N]` (skipping `base+0`), then after all are submitted, sends the gap-filling `base+0` tx to trigger promotion cascade.
- **PreSignedBurstLoad**: Broadcasts pre-signed Cosmos txs via HTTP POST to `/broadcast_tx_sync`. All accounts concurrent, round-robin across fullnodes.
- **PreSignedSequentialLoad**: Broadcasts pre-signed Cosmos txs via HTTP POST. Each account pinned to a single fullnode, txs sent sequentially per account.

### Metrics

| Metric | Source |
|---|---|
| **TPS** | `included_tx_count / block_time_span` |
| **Latency** (avg, p50, p95, p99, max) | `block_timestamp - submit_timestamp` per tx |
| **Peak mempool size** | Goroutine polling `/num_unconfirmed_txs` every 500ms |
| **Per-block tx count** | CometBFT RPC `/block?height=N` |

## Wasm exec workload: BenchHeavyState

The Wasm exec tests deploy the `BenchHeavyState` CosmWasm contract at runtime. Each tx calls `ExecuteMsg::WriteMixed{shared_count, local_count}` which performs:

- **shared writes** to the `SHARED_STATE` mapping.
- **local writes** to the `LOCAL_STATE` mapping.

Each call writes to **unique keys** using a per-sender nonce, so the state tree grows continuously. This creates IAVL rebalancing pressure that MemIAVL handles more efficiently.

CLI-based tests use `write_mixed{5, 25}` = 30 writes/tx. Pre-signed HTTP tests use `write_mixed{20, 80}` = 100 writes/tx. Stress tests use `write_mixed{40, 160}` = 200 writes/tx.

### Building the contract

```bash
cd bench_heavy_state
cargo build --target wasm32-unknown-unknown --release
```
