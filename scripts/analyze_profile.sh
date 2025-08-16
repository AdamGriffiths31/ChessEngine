#!/bin/bash

PROFILE="${1:-profiles/cpu_*.prof}"

# Find the most recent profile if not specified
if [[ "$PROFILE" == "profiles/cpu_*.prof" ]]; then
    PROFILE=$(ls -t profiles/cpu_*.prof 2>/dev/null | head -1)
fi

if [[ ! -f "$PROFILE" ]]; then
    echo "No profile file found. Run profile_position.sh first."
    exit 1
fi

echo "Analyzing profile: $PROFILE"
echo ""

# Top functions by CPU time
echo "=== Top 10 Functions by CPU Time ==="
go tool pprof -text "$PROFILE" | head -20

echo ""
echo "=== Functions consuming >5% CPU ==="
go tool pprof -text "$PROFILE" | awk '$2 > 5.0 {print $0}'

echo ""
echo "=== Move Ordering Analysis ==="
go tool pprof -list "orderMoves" "$PROFILE" | head -50

echo ""
echo "=== Transposition Table Analysis ==="
go tool pprof -list "Probe|Store" "$PROFILE" | grep -A5 -B5 "MinimaxEngine"

echo ""
echo "To view interactive profile:"
echo "  go tool pprof -http=:8080 $PROFILE"