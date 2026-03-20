//go:build benchmark

package benchmark

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/initia-labs/miniwasm/integration-tests/e2e/cluster"
	"github.com/stretchr/testify/require"
)

const (
	clusterReadyTimeout = 120 * time.Second
	mempoolDrainTimeout = 180 * time.Second
	mempoolPollInterval = 500 * time.Millisecond
	warmupSettleTime    = 5 * time.Second
)

func resultsDir(t *testing.T) string {
	t.Helper()
	if d := os.Getenv("BENCHMARK_RESULTS_DIR"); d != "" {
		return d
	}
	return filepath.Join("results")
}

func setupCluster(t *testing.T, ctx context.Context, cfg BenchConfig) *cluster.Cluster {
	t.Helper()

	cl, err := cluster.NewCluster(ctx, t, cluster.ClusterOptions{
		NodeCount:      cfg.NodeCount,
		AccountCount:   cfg.AccountCount,
		ChainID:        "bench-miniwasm",
		BinaryPath:     os.Getenv("E2E_MINITIAD_BIN"),
		MemIAVL:        cfg.MemIAVL,
		ValidatorCount: cfg.ValidatorCount,
		MaxBlockGas:    cfg.MaxBlockGas,
		NoAllowQueued:  cfg.NoAllowQueued,
	})
	require.NoError(t, err)

	require.NoError(t, cl.Start(ctx))
	t.Cleanup(cl.Close)
	require.NoError(t, cl.WaitForReady(ctx, clusterReadyTimeout))

	return cl
}

func runBenchmarkWithCluster(t *testing.T, ctx context.Context, cl *cluster.Cluster, cfg BenchConfig, loadFn func(ctx context.Context, cl *cluster.Cluster, cfg BenchConfig, metas map[string]cluster.AccountMeta) LoadResult) BenchResult {
	t.Helper()

	metas, err := CollectInitialMetas(ctx, cl)
	require.NoError(t, err)

	Warmup(ctx, cl, metas)
	require.NoError(t, cl.WaitForMempoolEmpty(ctx, 30*time.Second))
	time.Sleep(warmupSettleTime)

	metas, err = CollectInitialMetas(ctx, cl)
	require.NoError(t, err)

	startHeight, err := cl.LatestHeight(ctx, 0)
	require.NoError(t, err)

	poller := NewMempoolPoller(ctx, cl, mempoolPollInterval)

	t.Logf("Starting load: %d accounts x %d txs = %d total", cfg.AccountCount, cfg.TxPerAccount, cfg.TotalTx())
	loadResult := loadFn(ctx, cl, cfg, metas)
	t.Logf("Load complete: %d submitted, %d errors, duration=%.1fs",
		len(loadResult.Submissions), len(loadResult.Errors),
		loadResult.EndTime.Sub(loadResult.StartTime).Seconds())

	drainTimeout := mempoolDrainTimeout + time.Duration(cfg.TotalTx()/20)*time.Second
	endHeight, err := WaitForAllIncluded(ctx, cl, drainTimeout)
	require.NoError(t, err)

	peakMempool := poller.Stop()

	result, err := CollectResults(ctx, cl, cfg, loadResult, startHeight, endHeight, peakMempool)
	require.NoError(t, err)

	t.Logf("Results: TPS=%.1f, P50=%.0fms, P95=%.0fms, P99=%.0fms, included=%d/%d, peak_mempool=%d",
		result.TxPerSecond, result.P50LatencyMs, result.P95LatencyMs, result.P99LatencyMs,
		result.TotalIncluded, result.TotalSubmitted, result.PeakMempoolSize)

	require.NoError(t, WriteResult(t, result, resultsDir(t)))

	return result
}

func runBenchmark(t *testing.T, cfg BenchConfig, loadFn func(ctx context.Context, cl *cluster.Cluster, cfg BenchConfig, metas map[string]cluster.AccountMeta) LoadResult) BenchResult {
	t.Helper()
	ctx := context.Background()

	cl := setupCluster(t, ctx, cfg)
	defer cl.Close()

	return runBenchmarkWithCluster(t, ctx, cl, cfg, loadFn)
}

// ---------------------------------------------------------------------------
// Pre-signed HTTP broadcast benchmarks
// ---------------------------------------------------------------------------

func runPreSignedBenchmark(
	t *testing.T, ctx context.Context, cl *cluster.Cluster, cfg BenchConfig,
	preSignFn func(metas map[string]cluster.AccountMeta) []cluster.SignedTx,
	loadFnFactory func([]cluster.SignedTx) func(ctx context.Context, cl *cluster.Cluster, cfg BenchConfig, metas map[string]cluster.AccountMeta) LoadResult,
) BenchResult {
	t.Helper()

	metas, err := CollectInitialMetas(ctx, cl)
	require.NoError(t, err)

	Warmup(ctx, cl, metas)
	require.NoError(t, cl.WaitForMempoolEmpty(ctx, 30*time.Second))
	time.Sleep(warmupSettleTime)

	metas, err = CollectInitialMetas(ctx, cl)
	require.NoError(t, err)

	signedTxs := preSignFn(metas)

	startHeight, err := cl.LatestHeight(ctx, 0)
	require.NoError(t, err)

	poller := NewMempoolPoller(ctx, cl, mempoolPollInterval)

	t.Logf("Starting load: %d accounts x %d txs = %d total (pre-signed HTTP)", cfg.AccountCount, cfg.TxPerAccount, cfg.TotalTx())
	loadFn := loadFnFactory(signedTxs)
	loadResult := loadFn(ctx, cl, cfg, metas)
	t.Logf("Load complete: %d submitted, %d errors, duration=%.1fs",
		len(loadResult.Submissions), len(loadResult.Errors),
		loadResult.EndTime.Sub(loadResult.StartTime).Seconds())

	drainTimeout := mempoolDrainTimeout + time.Duration(cfg.TotalTx()/20)*time.Second
	if cfg.DrainTimeoutOverride > 0 {
		drainTimeout = cfg.DrainTimeoutOverride
	}
	endHeight, err := WaitForAllIncluded(ctx, cl, drainTimeout)
	if err != nil {
		t.Logf("Warning: mempool drain incomplete: %v (collecting partial results)", err)
		endHeight, _ = cl.LatestHeight(ctx, 0)
	}

	peakMempool := poller.Stop()

	result, err := CollectResults(ctx, cl, cfg, loadResult, startHeight, endHeight, peakMempool)
	require.NoError(t, err)

	t.Logf("Results: TPS=%.1f, P50=%.0fms, P95=%.0fms, P99=%.0fms, included=%d/%d, peak_mempool=%d",
		result.TxPerSecond, result.P50LatencyMs, result.P95LatencyMs, result.P99LatencyMs,
		result.TotalIncluded, result.TotalSubmitted, result.PeakMempoolSize)

	require.NoError(t, WriteResult(t, result, resultsDir(t)))
	return result
}

// ---------------------------------------------------------------------------
// Wasm exec setup
// ---------------------------------------------------------------------------

// setupWasmExecCluster stores and instantiates BenchHeavyState, returns params for load generation.
// The contract wasm must be pre-built at bench_heavy_state/contract.wasm.
func setupWasmExecCluster(t *testing.T, ctx context.Context, cl *cluster.Cluster, sharedWrites, localWrites int64) (contractAddr, execMsg string, estimatedGas uint64) {
	t.Helper()

	wasmPath := filepath.Join("bench_heavy_state", "contract.wasm")
	if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
		t.Fatalf("contract.wasm not found at %s. build with: cd bench_heavy_state && cargo build --release --target wasm32-unknown-unknown && cp target/wasm32-unknown-unknown/release/bench_heavy_state.wasm contract.wasm", wasmPath)
	}

	absWasmPath, err := filepath.Abs(wasmPath)
	require.NoError(t, err)

	deployerName := cl.AccountNames()[0]

	// store contract
	storeRes := cl.StoreWasmContract(ctx, deployerName, absWasmPath, 0)
	require.NoError(t, storeRes.Err)
	require.Equal(t, int64(0), storeRes.Code, "store failed: %s", storeRes.RawLog)
	require.NoError(t, cl.WaitForMempoolEmpty(ctx, 30*time.Second))
	time.Sleep(3 * time.Second)

	// extract code ID
	txResult, err := cl.QueryTxResult(ctx, 0, storeRes.TxHash)
	require.NoError(t, err)
	codeID, err := cluster.ExtractCodeID(txResult)
	require.NoError(t, err)
	t.Logf("BenchHeavyState stored with code_id=%s", codeID)

	// instantiate contract
	initMsg := `{}`
	instRes := cl.InstantiateWasmContract(ctx, deployerName, codeID, initMsg, "bench_heavy_state", 0)
	require.NoError(t, instRes.Err)
	require.Equal(t, int64(0), instRes.Code, "instantiate failed: %s", instRes.RawLog)
	require.NoError(t, cl.WaitForMempoolEmpty(ctx, 30*time.Second))
	time.Sleep(3 * time.Second)

	// extract contract address
	instTxResult, err := cl.QueryTxResult(ctx, 0, instRes.TxHash)
	require.NoError(t, err)
	contractAddr, err = cluster.ExtractContractAddress(instTxResult)
	require.NoError(t, err)
	t.Logf("BenchHeavyState instantiated at: %s", contractAddr)

	// build execute message
	execMsg = fmt.Sprintf(`{"write_mixed":{"shared_count":%d,"local_count":%d}}`, sharedWrites, localWrites)

	estimatedGas, err = cl.EstimateWasmGas(ctx, deployerName, contractAddr, execMsg, 0)
	require.NoError(t, err)
	estimatedGas = estimatedGas * 3 / 2
	t.Logf("Estimated gas for write_mixed(%d shared, %d local): %d (with 1.5x adjustment)", sharedWrites, localWrites, estimatedGas)

	return contractAddr, execMsg, estimatedGas
}

// setupWasmExecLoad deploys BenchHeavyState and returns a LoadFn closure.
func setupWasmExecLoad(t *testing.T, ctx context.Context, cl *cluster.Cluster) func(ctx context.Context, cl *cluster.Cluster, cfg BenchConfig, metas map[string]cluster.AccountMeta) LoadResult {
	t.Helper()

	const (
		sharedWrites int64 = 5
		localWrites  int64 = 25
	)

	contractAddr, execMsg, gas := setupWasmExecCluster(t, ctx, cl, sharedWrites, localWrites)
	return WasmExecSequentialLoad(contractAddr, execMsg, gas)
}

// ---------------------------------------------------------------------------
// Mempool comparison: CList vs. Proxy+Priority (bank send)
// ---------------------------------------------------------------------------

func TestBenchmarkBaselineSeq(t *testing.T) {
	cfg := BaselineConfig()
	cfg.Label = "clist/iavl/seq"
	runBenchmark(t, cfg, SequentialLoad)
}

func TestBenchmarkBaselineBurst(t *testing.T) {
	cfg := BaselineConfig()
	cfg.Label = "clist/iavl/burst"
	runBenchmark(t, cfg, BurstLoad)
}

func TestBenchmarkSeqComparison(t *testing.T) {
	var results []BenchResult

	baselines := LoadBaselineResultsByLabel(resultsDir(t), "clist/iavl/seq")
	if len(baselines) > 0 {
		t.Logf("Loaded baseline result: %s", baselines[0].Config.Label)
		results = append(results, baselines[0])
	} else {
		t.Log("No baseline results found. Run TestBenchmarkBaselineSeq with pre-proxy binary for full comparison.")
	}

	t.Run("MempoolOnly", func(t *testing.T) {
		cfg := MempoolOnlyConfig()
		cfg.Label = "proxy+priority/iavl/seq"
		result := runBenchmark(t, cfg, SequentialLoad)
		results = append(results, result)
	})

	t.Run("Combined", func(t *testing.T) {
		cfg := CombinedConfig()
		cfg.Label = "proxy+priority/memiavl/seq"
		result := runBenchmark(t, cfg, SequentialLoad)
		results = append(results, result)
	})

	if len(results) >= 2 {
		PrintComparisonTable(t, results)
		PrintImprovementTable(t, results)
	}
}

func TestBenchmarkBurstComparison(t *testing.T) {
	var results []BenchResult

	baselines := LoadBaselineResultsByLabel(resultsDir(t), "clist/iavl/burst")
	if len(baselines) > 0 {
		t.Logf("Loaded baseline result: %s", baselines[0].Config.Label)
		results = append(results, baselines[0])
	} else {
		t.Log("No baseline results found. Run TestBenchmarkBaselineBurst with pre-proxy binary for full comparison.")
	}

	t.Run("MempoolOnly", func(t *testing.T) {
		cfg := MempoolOnlyConfig()
		cfg.Label = "proxy+priority/iavl/burst"
		result := runBenchmark(t, cfg, BurstLoad)
		results = append(results, result)
	})

	t.Run("Combined", func(t *testing.T) {
		cfg := CombinedConfig()
		cfg.Label = "proxy+priority/memiavl/burst"
		result := runBenchmark(t, cfg, BurstLoad)
		results = append(results, result)
	})

	if len(results) >= 2 {
		PrintComparisonTable(t, results)
		PrintImprovementTable(t, results)
	}
}

// ---------------------------------------------------------------------------
// Pre-signed HTTP broadcast benchmarks (bank send)
// ---------------------------------------------------------------------------

func TestBenchmarkPreSignedSeqComparison(t *testing.T) {
	var results []BenchResult

	t.Run("IAVL", func(t *testing.T) {
		cfg := MempoolOnlyConfig()
		cfg.AccountCount = 20
		cfg.TxPerAccount = 100
		cfg.Label = "presigned/iavl/seq"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		result := runPreSignedBenchmark(t, ctx, cl, cfg,
			func(metas map[string]cluster.AccountMeta) []cluster.SignedTx {
				return PreSignBankTxs(ctx, t, cl, cfg, metas)
			}, PreSignedSequentialLoad)
		results = append(results, result)
	})

	t.Run("MemIAVL", func(t *testing.T) {
		cfg := CombinedConfig()
		cfg.AccountCount = 20
		cfg.TxPerAccount = 100
		cfg.Label = "presigned/memiavl/seq"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		result := runPreSignedBenchmark(t, ctx, cl, cfg,
			func(metas map[string]cluster.AccountMeta) []cluster.SignedTx {
				return PreSignBankTxs(ctx, t, cl, cfg, metas)
			}, PreSignedSequentialLoad)
		results = append(results, result)
	})

	if len(results) == 2 {
		PrintComparisonTable(t, results)
	}
}

func TestBenchmarkPreSignedBurstComparison(t *testing.T) {
	var results []BenchResult

	t.Run("IAVL", func(t *testing.T) {
		cfg := MempoolOnlyConfig()
		cfg.AccountCount = 20
		cfg.TxPerAccount = 100
		cfg.Label = "presigned/iavl/burst"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		result := runPreSignedBenchmark(t, ctx, cl, cfg,
			func(metas map[string]cluster.AccountMeta) []cluster.SignedTx {
				return PreSignBankTxs(ctx, t, cl, cfg, metas)
			}, PreSignedBurstLoad)
		results = append(results, result)
	})

	t.Run("MemIAVL", func(t *testing.T) {
		cfg := CombinedConfig()
		cfg.AccountCount = 20
		cfg.TxPerAccount = 100
		cfg.Label = "presigned/memiavl/burst"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		result := runPreSignedBenchmark(t, ctx, cl, cfg,
			func(metas map[string]cluster.AccountMeta) []cluster.SignedTx {
				return PreSignBankTxs(ctx, t, cl, cfg, metas)
			}, PreSignedBurstLoad)
		results = append(results, result)
	})

	if len(results) == 2 {
		PrintComparisonTable(t, results)
	}
}

// ---------------------------------------------------------------------------
// State DB comparison: IAVL vs MemIAVL (bank send)
// ---------------------------------------------------------------------------

func TestBenchmarkMemIAVLBankSend(t *testing.T) {
	var results []BenchResult

	t.Run("IAVL", func(t *testing.T) {
		cfg := MempoolOnlyConfig()
		cfg.AccountCount = 100
		cfg.TxPerAccount = 200
		cfg.Label = "memiavl-compare/iavl/bank-send"
		result := runBenchmark(t, cfg, BurstLoad)
		results = append(results, result)
	})

	t.Run("MemIAVL", func(t *testing.T) {
		cfg := CombinedConfig()
		cfg.AccountCount = 100
		cfg.TxPerAccount = 200
		cfg.Label = "memiavl-compare/memiavl/bank-send"
		result := runBenchmark(t, cfg, BurstLoad)
		results = append(results, result)
	})

	if len(results) == 2 {
		PrintComparisonTable(t, results)
	}
}

// ---------------------------------------------------------------------------
// Capability demos
// ---------------------------------------------------------------------------

func TestBenchmarkQueuePromotion(t *testing.T) {
	cfg := MempoolOnlyConfig()
	cfg.TxPerAccount = 50
	cfg.Label = "queue-promotion/mempool-only"
	result := runBenchmark(t, cfg, OutOfOrderLoad)

	require.Equal(t, result.TotalSubmitted, result.TotalIncluded,
		"not all out-of-order transactions were included: submitted=%d included=%d",
		result.TotalSubmitted, result.TotalIncluded)
}

func TestBenchmarkQueuedFlood(t *testing.T) {
	cfg := MempoolOnlyConfig()
	cfg.TxPerAccount = 50
	cfg.Label = "queued-flood/mempool-only"
	result := runBenchmark(t, cfg, QueuedFloodLoad)

	require.Equal(t, result.TotalSubmitted, result.TotalIncluded,
		"not all queued-flood transactions were included: submitted=%d included=%d",
		result.TotalSubmitted, result.TotalIncluded)
}

func TestBenchmarkQueuedGapEviction(t *testing.T) {
	cfg := MempoolOnlyConfig()
	cfg.TxPerAccount = 50
	cfg.Label = "queued-gap-eviction/mempool-only"

	ctx := context.Background()
	cl := setupCluster(t, ctx, cfg)
	defer cl.Close()

	metas, err := CollectInitialMetas(ctx, cl)
	require.NoError(t, err)

	Warmup(ctx, cl, metas)
	require.NoError(t, cl.WaitForMempoolEmpty(ctx, 30*time.Second))
	time.Sleep(warmupSettleTime)

	metas, err = CollectInitialMetas(ctx, cl)
	require.NoError(t, err)

	loadResult := QueuedGapLoad(ctx, cl, cfg, metas)
	t.Logf("Submitted %d future-nonce txs (no gap fill), %d errors",
		len(loadResult.Submissions), len(loadResult.Errors))

	poller := NewMempoolPoller(ctx, cl, mempoolPollInterval)

	t.Log("Waiting for gap TTL eviction (60s + 30s buffer)...")
	time.Sleep(90 * time.Second)

	err = cl.WaitForMempoolEmpty(ctx, 30*time.Second)
	peakMempool := poller.Stop()

	t.Logf("Gap eviction test: peak_mempool=%d, mempool_drained=%v",
		peakMempool, err == nil)

	require.NoError(t, err, "mempool should be empty after gap TTL eviction")
	require.Greater(t, peakMempool, 0, "should have observed queued txs in mempool")
}

func TestBenchmarkGossipPropagation(t *testing.T) {
	cfg := MempoolOnlyConfig()
	cfg.AccountCount = 5
	cfg.TxPerAccount = 50
	cfg.Label = "gossip/mempool-only"

	ctx := context.Background()
	cl := setupCluster(t, ctx, cfg)
	defer cl.Close()

	metas, err := CollectInitialMetas(ctx, cl)
	require.NoError(t, err)

	Warmup(ctx, cl, metas)
	require.NoError(t, cl.WaitForMempoolEmpty(ctx, 30*time.Second))
	time.Sleep(warmupSettleTime)

	metas, err = CollectInitialMetas(ctx, cl)
	require.NoError(t, err)

	startHeight, err := cl.LatestHeight(ctx, 0)
	require.NoError(t, err)

	poller := NewMempoolPoller(ctx, cl, mempoolPollInterval)

	targetNode := cfg.ValidatorCount // first edge fullnode
	loadResult := SingleNodeLoad(ctx, cl, cfg, metas, targetNode)
	t.Logf("Submitted %d txs to node %d (edge)", len(loadResult.Submissions), targetNode)

	endHeight, err := WaitForAllIncluded(ctx, cl, mempoolDrainTimeout)
	require.NoError(t, err)

	peakMempool := poller.Stop()
	t.Logf("Cluster peak mempool size: %d", peakMempool)

	result, err := CollectResults(ctx, cl, cfg, loadResult, startHeight, endHeight, peakMempool)
	require.NoError(t, err)

	t.Logf("Gossip test: TPS=%.1f, included=%d/%d",
		result.TxPerSecond, result.TotalIncluded, result.TotalSubmitted)
	require.NoError(t, WriteResult(t, result, resultsDir(t)))
}

// ---------------------------------------------------------------------------
// Wasm exec tests
// ---------------------------------------------------------------------------

func TestBenchmarkBaselineSeqWasmExec(t *testing.T) {
	cfg := BaselineConfig()
	cfg.AccountCount = 100
	cfg.TxPerAccount = 50
	cfg.Label = "clist/iavl/seq-wasm-exec"

	ctx := context.Background()
	cl := setupCluster(t, ctx, cfg)
	defer cl.Close()

	wasmLoadFn := setupWasmExecLoad(t, ctx, cl)
	runBenchmarkWithCluster(t, ctx, cl, cfg, wasmLoadFn)
}

func TestBenchmarkSeqComparisonWasmExec(t *testing.T) {
	var results []BenchResult

	baselines := LoadBaselineResultsByLabel(resultsDir(t), "clist/iavl/seq-wasm-exec")
	if len(baselines) > 0 {
		t.Logf("Loaded baseline result: %s", baselines[0].Config.Label)
		results = append(results, baselines[0])
	} else {
		t.Log("No baseline results found. Run TestBenchmarkBaselineSeqWasmExec with pre-proxy binary for full comparison.")
	}

	t.Run("MempoolOnly", func(t *testing.T) {
		cfg := MempoolOnlyConfig()
		cfg.AccountCount = 100
		cfg.TxPerAccount = 50
		cfg.Label = "proxy+priority/iavl/seq-wasm-exec"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		wasmLoadFn := setupWasmExecLoad(t, ctx, cl)
		result := runBenchmarkWithCluster(t, ctx, cl, cfg, wasmLoadFn)
		results = append(results, result)
	})

	t.Run("Combined", func(t *testing.T) {
		cfg := CombinedConfig()
		cfg.AccountCount = 100
		cfg.TxPerAccount = 50
		cfg.Label = "proxy+priority/memiavl/seq-wasm-exec"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		wasmLoadFn := setupWasmExecLoad(t, ctx, cl)
		result := runBenchmarkWithCluster(t, ctx, cl, cfg, wasmLoadFn)
		results = append(results, result)
	})

	if len(results) >= 2 {
		PrintComparisonTable(t, results)
		PrintImprovementTable(t, results)
	}
}

func TestBenchmarkPreSignedSeqWasmExec(t *testing.T) {
	var results []BenchResult

	const (
		sharedWrites int64 = 20
		localWrites  int64 = 80
	)

	t.Run("IAVL", func(t *testing.T) {
		cfg := MempoolOnlyConfig()
		cfg.AccountCount = 20
		cfg.TxPerAccount = 100
		cfg.Label = "presigned/iavl/seq-wasm-exec"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		contractAddr, execMsg, gas := setupWasmExecCluster(t, ctx, cl, sharedWrites, localWrites)

		result := runPreSignedBenchmark(t, ctx, cl, cfg,
			func(metas map[string]cluster.AccountMeta) []cluster.SignedTx {
				return PreSignWasmExecTxs(ctx, t, cl, cfg, metas, contractAddr, execMsg, gas)
			}, PreSignedSequentialLoad)
		results = append(results, result)
	})

	t.Run("MemIAVL", func(t *testing.T) {
		cfg := CombinedConfig()
		cfg.AccountCount = 20
		cfg.TxPerAccount = 100
		cfg.Label = "presigned/memiavl/seq-wasm-exec"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		contractAddr, execMsg, gas := setupWasmExecCluster(t, ctx, cl, sharedWrites, localWrites)

		result := runPreSignedBenchmark(t, ctx, cl, cfg,
			func(metas map[string]cluster.AccountMeta) []cluster.SignedTx {
				return PreSignWasmExecTxs(ctx, t, cl, cfg, metas, contractAddr, execMsg, gas)
			}, PreSignedSequentialLoad)
		results = append(results, result)
	})

	if len(results) == 2 {
		PrintComparisonTable(t, results)
	}
}

func TestBenchmarkPreSignedSeqWasmExecStress(t *testing.T) {
	var results []BenchResult

	const (
		sharedWrites int64 = 20
		localWrites  int64 = 80
	)

	t.Run("IAVL", func(t *testing.T) {
		cfg := MempoolOnlyConfig()
		cfg.AccountCount = 20
		cfg.TxPerAccount = 200
		cfg.DrainTimeoutOverride = 10 * time.Minute
		cfg.Label = "presigned-stress/iavl/seq-wasm-exec"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		contractAddr, execMsg, gas := setupWasmExecCluster(t, ctx, cl, sharedWrites, localWrites)

		result := runPreSignedBenchmark(t, ctx, cl, cfg,
			func(metas map[string]cluster.AccountMeta) []cluster.SignedTx {
				return PreSignWasmExecTxs(ctx, t, cl, cfg, metas, contractAddr, execMsg, gas)
			}, PreSignedSequentialLoad)
		results = append(results, result)
	})

	t.Run("MemIAVL", func(t *testing.T) {
		cfg := CombinedConfig()
		cfg.AccountCount = 20
		cfg.TxPerAccount = 200
		cfg.DrainTimeoutOverride = 10 * time.Minute
		cfg.Label = "presigned-stress/memiavl/seq-wasm-exec"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		contractAddr, execMsg, gas := setupWasmExecCluster(t, ctx, cl, sharedWrites, localWrites)

		result := runPreSignedBenchmark(t, ctx, cl, cfg,
			func(metas map[string]cluster.AccountMeta) []cluster.SignedTx {
				return PreSignWasmExecTxs(ctx, t, cl, cfg, metas, contractAddr, execMsg, gas)
			}, PreSignedSequentialLoad)
		results = append(results, result)
	})

	if len(results) == 2 {
		PrintComparisonTable(t, results)
	}
}

func TestBenchmarkMemIAVLWasmExec(t *testing.T) {
	var results []BenchResult

	t.Run("IAVL", func(t *testing.T) {
		cfg := MempoolOnlyConfig()
		cfg.AccountCount = 100
		cfg.TxPerAccount = 50
		cfg.Label = "memiavl-compare/iavl/wasm-exec"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		wasmLoadFn := setupWasmExecLoad(t, ctx, cl)
		result := runBenchmarkWithCluster(t, ctx, cl, cfg, wasmLoadFn)
		results = append(results, result)
	})

	t.Run("MemIAVL", func(t *testing.T) {
		cfg := CombinedConfig()
		cfg.AccountCount = 100
		cfg.TxPerAccount = 50
		cfg.Label = "memiavl-compare/memiavl/wasm-exec"

		ctx := context.Background()
		cl := setupCluster(t, ctx, cfg)
		defer cl.Close()

		wasmLoadFn := setupWasmExecLoad(t, ctx, cl)
		result := runBenchmarkWithCluster(t, ctx, cl, cfg, wasmLoadFn)
		results = append(results, result)
	})

	if len(results) == 2 {
		PrintComparisonTable(t, results)
	}
}
