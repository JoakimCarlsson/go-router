# Benchmark Scripts

This directory contains scripts for running comprehensive performance benchmarks on the go-router.

## benchmark.ps1 (PowerShell Script)

A comprehensive benchmarking script that runs various performance tests and generates detailed reports.

### Usage

```powershell
# Basic usage - runs all benchmarks (includes comparisons with standard library)
.\benchmark.ps1

# Quick benchmark - only runs static routes test
.\benchmark.ps1 -Quick

# Enable profiling - generates CPU and memory profiles
.\benchmark.ps1 -EnableProfiling
```

### Parameters

- **`-Quick`** - Runs only the `BenchmarkRouter_StaticRoutes` benchmark for fast performance checks
- **`-EnableProfiling`** - Generates CPU and memory profiles for detailed analysis

**Note**: Comparison benchmarks against Go's standard library are automatically included in all runs.

### Output Files

All results are saved to `benchmark_results/` directory with timestamps:

#### Standard Runs
- `basic_YYYYMMDD_HHMMSS.txt` - All benchmark results
- `extended_YYYYMMDD_HHMMSS.txt` - 5 iterations for statistical accuracy
- `concurrent_Nproc_YYYYMMDD_HHMMSS.txt` - Concurrent performance with different GOMAXPROCS

#### Quick Runs
- `quick_YYYYMMDD_HHMMSS.txt` - Single benchmark result

#### Profiling (with -EnableProfiling)
- `cpu_YYYYMMDD_HHMMSS.prof` - CPU profiling data
- `mem_YYYYMMDD_HHMMSS.prof` - Memory profiling data
- `cpu_bench_YYYYMMDD_HHMMSS.txt` - CPU benchmark results
- `mem_bench_YYYYMMDD_HHMMSS.txt` - Memory benchmark results


### Analyzing Profile Data

After running with `-EnableProfiling`, analyze the profiles using:

```bash
# CPU profile analysis
go tool pprof benchmark_results/cpu_YYYYMMDD_HHMMSS.prof

# Memory profile analysis  
go tool pprof benchmark_results/mem_YYYYMMDD_HHMMSS.prof
```

### Examples

```powershell
# Quick performance check
.\benchmark.ps1 -Quick

# Full benchmark suite with profiling (includes standard library comparison)
.\benchmark.ps1 -EnableProfiling

# Basic comprehensive benchmark suite
.\benchmark.ps1
```

### Concurrent Testing

The script automatically tests concurrent performance with different GOMAXPROCS values:
- 1 core (single-threaded)
- 2 cores 
- 4 cores
- 8 cores

This helps identify optimal concurrency settings and scaling characteristics.

### Performance Metrics

The benchmarks measure:
- **Throughput** - Operations per second (ops/sec)
- **Latency** - Nanoseconds per operation (ns/op) 
- **Memory** - Bytes allocated per operation (B/op)
- **Allocations** - Number of allocations per operation (allocs/op)

### Benchmark Categories

#### Router Benchmarks
- Static route matching
- Parameter extraction (single and multiple)
- Middleware chain execution
- Large routing tables (1000+ routes)
- JSON response generation
- Route groups and nesting
- Concurrent request handling

#### Comparison Benchmarks  
- Direct performance comparison with Go's `net/http.ServeMux`
- Same operations implemented with standard library
- Highlights performance improvements and memory efficiency

### Prerequisites

- Go 1.22+ (for path parameter support)
- PowerShell (Windows) or PowerShell Core (cross-platform)
- Write access to create `benchmark_results/` directory

### Troubleshooting

If you encounter path issues, ensure you're running from the `scripts/` directory:

```powershell
cd scripts
.\benchmark.ps1
```

The script automatically handles directory navigation and creates output directories as needed.