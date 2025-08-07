#!/bin/bash

# STS (Strategic Test Suite) Benchmark Runner
# Simple script to run the STS benchmark with sensible defaults

echo "Building STS benchmark..."
go build -o bin/sts ./cmd/sts

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Running STS benchmark..."
echo "======================="

# Run with time-based search instead of depth limit
# - No depth limit: Let engine search as deep as possible
# - Timeout 2 seconds per position: Use full time budget
# - Max 5 positions: Quick test run  
# - Verbose: Show detailed results

./bin/sts -depth 999 -timeout 30 -max 5 -verbose

echo ""
