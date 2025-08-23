#!/bin/bash

# Default values
STS_FILE="testdata/STS1.epd"
POSITION=1
SEARCH_TIME="10s"
TT_SIZE=256

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--file)
            STS_FILE="$2"
            shift 2
            ;;
        -p|--position)
            POSITION="$2"
            shift 2
            ;;
        -t|--time)
            SEARCH_TIME="$2"
            shift 2
            ;;
        --tt)
            TT_SIZE="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -f, --file FILE       STS file to use (default: testdata/STS1.epd)"
            echo "  -p, --position NUM    Position number in file (default: 1)"
            echo "  -t, --time DURATION   Search time (default: 10s)"
            echo "  --tt SIZE            TT size in MB (default: 256)"
            echo "  -h, --help           Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create profiles directory if it doesn't exist
mkdir -p profiles

# Generate timestamp for unique filenames
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
CPU_PROFILE="profiles/cpu_${TIMESTAMP}.prof"
MEM_PROFILE="profiles/mem_${TIMESTAMP}.prof"

echo "Running profile for position $POSITION in $STS_FILE"
echo "Search time: $SEARCH_TIME, TT: ${TT_SIZE}MB"
echo "CPU profile: $CPU_PROFILE"
echo "Memory profile: $MEM_PROFILE"
echo ""

# Build and run the profiler
go build -o cmd/profile/profile cmd/profile/main.go || exit 1

./cmd/profile/profile \
    -file="$STS_FILE" \
    -pos="$POSITION" \
    -time="$SEARCH_TIME" \
    -tt="$TT_SIZE" \
    -cpuprofile="$CPU_PROFILE" \
    -memprofile="$MEM_PROFILE" \
    -details

echo ""
echo "To analyze CPU profile:"
echo "  go tool pprof -http=:8080 $CPU_PROFILE"
echo ""
echo "To analyze memory profile:"
echo "  go tool pprof -http=:8081 $MEM_PROFILE"
echo ""
echo "Common pprof commands:"
echo "  top10    - Show top 10 functions by CPU time"
echo "  list fn  - Show source code for function"
echo "  web      - Show call graph"