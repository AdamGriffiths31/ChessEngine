#!/bin/bash

# Quick benchmark: ChessEngine vs Very Weak Stockfish
# No user input required - runs with Skill Level 0

set -e

echo "=== Quick Benchmark: ChessEngine vs Very Weak Stockfish ==="
echo "Configuration: Stockfish Skill Level 0, 1MB Hash, 3 games"
echo

# Build our engine
echo "Building ChessEngine..."
cd /home/adam/Documents/git/ChessEngine
go build -o tools/bin/uci cmd/uci/main.go
cd tools/scripts

# Clean previous results
rm -f ../results/quick_benchmark.pgn

echo "Starting benchmark..."

# Run 3 games against very weak Stockfish
../engines/cutechess-cli \
  -engine cmd=../bin/uci name="ChessEngine" proto=uci \
  -engine cmd=../engines/stockfish name="Stockfish-Weak" proto=uci option.Hash=1 option."Skill Level"=0 \
  -each tc=3+0 \
  -games 3 \
  -concurrency 1 \
  -ratinginterval 1 \
  -outcomeinterval 1 \
  -event "Quick Benchmark" \
  -site "Local Testing" \
  -pgnout ../results/quick_benchmark.pgn

echo
echo "=== Benchmark Results ==="
if [ -f ../results/quick_benchmark.pgn ]; then
    echo "Win/Loss/Draw breakdown:"
    grep "Result" ../results/quick_benchmark.pgn | cut -d'"' -f2 | sort | uniq -c
    
    echo
    echo "ChessEngine results:"
    total_games=$(grep -c "Result" ../results/quick_benchmark.pgn || echo "0")
    wins=0
    losses=0 
    draws=0
    
    # Count wins where ChessEngine wins
    if grep -q 'White "ChessEngine".*Result "1-0"' ../results/quick_benchmark.pgn; then
        wins=$((wins + $(grep -c 'White "ChessEngine".*Result "1-0"' ../results/quick_benchmark.pgn)))
    fi
    if grep -q 'Black "ChessEngine".*Result "0-1"' ../results/quick_benchmark.pgn; then
        wins=$((wins + $(grep -c 'Black "ChessEngine".*Result "0-1"' ../results/quick_benchmark.pgn)))
    fi
    
    # Count losses where ChessEngine loses  
    if grep -q 'White "ChessEngine".*Result "0-1"' ../results/quick_benchmark.pgn; then
        losses=$((losses + $(grep -c 'White "ChessEngine".*Result "0-1"' ../results/quick_benchmark.pgn)))
    fi
    if grep -q 'Black "ChessEngine".*Result "1-0"' ../results/quick_benchmark.pgn; then
        losses=$((losses + $(grep -c 'Black "ChessEngine".*Result "1-0"' ../results/quick_benchmark.pgn)))
    fi
    
    # Count draws
    if grep -q 'Result "1/2-1/2"' ../results/quick_benchmark.pgn; then
        draws=$(grep -c 'Result "1/2-1/2"' ../results/quick_benchmark.pgn)
    fi
    
    echo "  Wins: $wins"
    echo "  Losses: $losses" 
    echo "  Draws: $draws"
    
    if [ $total_games -gt 0 ]; then
        win_rate=$(( (wins * 100) / total_games ))
        echo "  Win Rate: ${win_rate}%"
    else
        echo "  Win Rate: 0%"
    fi
    
    echo
    echo "Full results saved to ../results/quick_benchmark.pgn"
else
    echo "No results file found"
fi

echo
echo "=== Debug Log Analysis ==="
cd /home/adam/Documents/git/ChessEngine

# Check for UCI debug logs
uci_logs=$(find logs/ -name "uci_debug_*.log" -type f 2>/dev/null | sort -r | head -1)
if [ -n "$uci_logs" ]; then
    echo "Latest UCI debug log: $uci_logs"
    echo "Log file size: $(wc -l < "$uci_logs") lines"
    
    # Check for illegal move errors
    illegal_moves=$(grep -c "CRITICAL ERROR.*illegal move" "$uci_logs" || echo "0")
    if [ $illegal_moves -gt 0 ]; then
        echo "⚠️  FOUND $illegal_moves ILLEGAL MOVE ERRORS!"
        echo "Illegal move details:"
        grep -A 3 "CRITICAL ERROR.*illegal move" "$uci_logs" | head -20
    else
        echo "✅ No illegal move errors found"
    fi
    
    # Check for UCI command errors
    uci_errors=$(grep -c "ERROR:" "$uci_logs" || echo "0")
    echo "Total UCI errors: $uci_errors"
    if [ $uci_errors -gt 0 ]; then
        echo "Recent UCI errors:"
        grep "ERROR:" "$uci_logs" | tail -5
    fi
    
    echo
    echo "To view full UCI debug log:"
    echo "  cat $uci_logs"
else
    echo "No UCI debug logs found"
fi

# Check for game engine logs
engine_logs=$(find logs/ -name "game_engine_*.log" -type f 2>/dev/null | sort -r | head -1)
if [ -n "$engine_logs" ]; then
    echo
    echo "Latest game engine log: $engine_logs"
    echo "Log file size: $(wc -l < "$engine_logs") lines"
    
    # Check for move validation failures
    validation_failures=$(grep -c "Move validation FAILED" "$engine_logs" || echo "0")
    echo "Move validation failures: $validation_failures"
    if [ $validation_failures -gt 0 ]; then
        echo "Recent validation failures:"
        grep -A 5 "Move validation FAILED" "$engine_logs" | tail -20
    fi
    
    echo
    echo "To view full game engine log:"
    echo "  cat $engine_logs"
else
    echo "No game engine logs found"
fi

echo
echo "All debug logs preserved in: /home/adam/Documents/git/ChessEngine/logs/"