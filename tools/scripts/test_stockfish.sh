#!/bin/bash

# Test script for ChessEngine vs Weakened Stockfish
# Phase 2.2: Benchmarking against known strong engine

set -e

echo "=== ChessEngine vs Stockfish Benchmark ==="
echo

# Build our engine
echo "Building ChessEngine..."
cd /home/adam/Documents/git/ChessEngine
go build -o tools/bin/uci cmd/uci/main.go
cd tools

# Test different Stockfish weakness levels
declare -a CONFIGS=(
    "skill0:Skill Level=0,Hash=1:Very Weak (Skill 0, 1MB hash)"
    "skill3:Skill Level=3,Hash=4:Weak (Skill 3, 4MB hash)" 
    "skill5:Skill Level=5,Hash=8:Medium-Weak (Skill 5, 8MB hash)"
    "elo1400:UCI_LimitStrength=true,UCI_Elo=1400,Hash=4:ELO 1400 Limited"
)

echo "Available Stockfish configurations:"
for i in "${!CONFIGS[@]}"; do
    IFS=':' read -r name options desc <<< "${CONFIGS[$i]}"
    echo "$((i+1)). $desc"
done
echo

read -p "Choose configuration (1-${#CONFIGS[@]}): " choice

if [[ ! "$choice" =~ ^[1-${#CONFIGS[@]}]$ ]]; then
    echo "Invalid choice"
    exit 1
fi

# Parse selected configuration
config_index=$((choice-1))
IFS=':' read -r config_name config_options config_desc <<< "${CONFIGS[$config_index]}"

echo "Selected: $config_desc"
echo "Starting test games..."
echo

# Clean previous results
rm -f stockfish_test.pgn

# Create temporary UCI options for Stockfish
stockfish_options=""
IFS=',' read -ra OPTS <<< "$config_options"
for opt in "${OPTS[@]}"; do
    IFS='=' read -r name value <<< "$opt"
    stockfish_options+="-each option.$name=$value "
done

echo "Stockfish UCI options: $stockfish_options"
echo

# Run the match
./squashfs-root/usr/bin/cutechess-cli \
  -engine cmd=./bin/uci name="ChessEngine" proto=uci \
  -engine cmd=./stockfish name="Stockfish-$config_name" proto=uci $stockfish_options \
  -each tc=5+0 \
  -games 6 \
  -concurrency 1 \
  -ratinginterval 2 \
  -outcomeinterval 2 \
  -event "ChessEngine vs Stockfish Test" \
  -site "Local Benchmark" \
  -pgnout stockfish_test.pgn

echo
echo "=== Test Results ==="
if [ -f stockfish_test.pgn ]; then
    echo "Results saved to stockfish_test.pgn"
    echo
    echo "Game outcomes:"
    grep "Result" stockfish_test.pgn | cut -d'"' -f2 | sort | uniq -c
    echo
    echo "Sample games:"
    grep -E "^\[White|^\[Black|^\[Result|^1\." stockfish_test.pgn | head -20
else
    echo "No results file found"
fi

echo
echo "Benchmark completed!"