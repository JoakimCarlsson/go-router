# Go Router Performance Benchmark Script (PowerShell)
# This script runs comprehensive benchmarks and generates performance reports

param(
    [switch]$Quick,
    [switch]$EnableProfiling
)

$RootDir = Split-Path $PSScriptRoot -Parent
$BenchmarkDir = Join-Path $RootDir "benchmarks"
$ResultsDir = Join-Path $RootDir "benchmark_results"
$Timestamp = Get-Date -Format "yyyyMMdd_HHmmss"

# Create results directory
New-Item -ItemType Directory -Force -Path $ResultsDir | Out-Null

Write-Host "Running Go Router Performance Benchmarks..." -ForegroundColor Green
Write-Host "Results will be saved to: $ResultsDir/" -ForegroundColor Yellow

# Change to benchmark directory
Set-Location $BenchmarkDir

if ($Quick) {
    Write-Host "Running quick benchmarks..." -ForegroundColor Cyan
    go test -bench=BenchmarkRouter_StaticRoutes -benchmem | Tee-Object -FilePath "$ResultsDir/quick_$Timestamp.txt"
} else {
    # Basic benchmark run
    Write-Host "Running basic benchmarks..." -ForegroundColor Cyan
    go test -bench="Benchmark.*" -benchmem | Tee-Object -FilePath "$ResultsDir/basic_$Timestamp.txt"

    # Extended benchmark run for statistical significance
    Write-Host "Running extended benchmarks (5 iterations)..." -ForegroundColor Cyan
    go test -bench="Benchmark.*" -benchmem -count=5 | Tee-Object -FilePath "$ResultsDir/extended_$Timestamp.txt"
}

if ($EnableProfiling) {
    # CPU profiling
    Write-Host "Running CPU profiling benchmark..." -ForegroundColor Cyan
    go test -bench=BenchmarkRouter_StaticRoutes -benchmem -cpuprofile="$ResultsDir/cpu_$Timestamp.prof" | Tee-Object -FilePath "$ResultsDir/cpu_bench_$Timestamp.txt"

    # Memory profiling
    Write-Host "Running memory profiling benchmark..." -ForegroundColor Cyan
    go test -bench=BenchmarkRouter_MemoryAllocation -benchmem -memprofile="$ResultsDir/mem_$Timestamp.prof" | Tee-Object -FilePath "$ResultsDir/mem_bench_$Timestamp.txt"
}


# Concurrent benchmarks with different GOMAXPROCS
if (-not $Quick) {
    Write-Host "Running concurrent benchmarks..." -ForegroundColor Cyan
    foreach ($procs in @(1, 2, 4, 8)) {
        Write-Host "  Testing with GOMAXPROCS=$procs" -ForegroundColor Gray
        $env:GOMAXPROCS = $procs
        go test -bench=BenchmarkRouter_Concurrent -benchmem | Tee-Object -FilePath "$ResultsDir/concurrent_${procs}proc_$Timestamp.txt"
    }
    Remove-Item Env:GOMAXPROCS -ErrorAction SilentlyContinue
}

Write-Host "Benchmark suite completed!" -ForegroundColor Green
Write-Host ""
Write-Host "Results saved to:" -ForegroundColor Yellow

if ($Quick) {
    Write-Host "   Quick:       benchmark_results/quick_$Timestamp.txt" -ForegroundColor White
} else {
    Write-Host "   Basic:       benchmark_results/basic_$Timestamp.txt" -ForegroundColor White
    Write-Host "   Extended:    benchmark_results/extended_$Timestamp.txt" -ForegroundColor White
    Write-Host "   Concurrent:  benchmark_results/concurrent_*proc_$Timestamp.txt" -ForegroundColor White
}

if ($EnableProfiling) {
    Write-Host "   CPU Profile: benchmark_results/cpu_$Timestamp.prof" -ForegroundColor White
    Write-Host "   Mem Profile: benchmark_results/mem_$Timestamp.prof" -ForegroundColor White
}

if ($EnableProfiling) {
    Write-Host ""
    Write-Host "To analyze profiles:" -ForegroundColor Yellow
    Write-Host "   go tool pprof benchmark_results/cpu_$Timestamp.prof" -ForegroundColor White
    Write-Host "   go tool pprof benchmark_results/mem_$Timestamp.prof" -ForegroundColor White
}

Write-Host ""
Write-Host "Performance Summary:" -ForegroundColor Yellow
Write-Host "   Check basic_$Timestamp.txt for detailed router performance metrics" -ForegroundColor White
Write-Host "   Comparison benchmarks are included in the basic results" -ForegroundColor White