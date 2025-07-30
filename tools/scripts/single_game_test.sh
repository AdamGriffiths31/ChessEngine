#!/bin/bash

# Single game test to debug cutechess-cli setup
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
RESULTS_DIR="$PROJECT_ROOT/tools/results"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🎯 Single Game Test${NC}"
echo "Project root: $PROJECT_ROOT"
echo "Tools dir: $TOOLS_DIR"
echo "Results dir: $RESULTS_DIR"
echo ""

# Ensure directories exist
mkdir -p "$RESULTS_DIR"
mkdir -p "$TOOLS_DIR/bin"

# Build the engine
echo -e "${YELLOW}Building ChessEngine...${NC}"
cd "$PROJECT_ROOT"
if ! go build -o tools/bin/uci cmd/uci/main.go; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"

# Test paths
chess_cmd="$TOOLS_DIR/bin/uci"
opponent_cmd="$TOOLS_DIR/engines/stockfish"
cutechess_cmd="$TOOLS_DIR/engines/cutechess-cli"

echo -e "${YELLOW}Checking paths:${NC}"
echo "ChessEngine: $chess_cmd $(test -f "$chess_cmd" && echo "✅" || echo "❌")"
echo "Stockfish: $opponent_cmd $(test -f "$opponent_cmd" && echo "✅" || echo "❌")"
echo "Cutechess: $cutechess_cmd $(test -f "$cutechess_cmd" && echo "✅" || echo "❌")"
echo ""

# Test files
timestamp=$(date +%Y%m%d_%H%M%S)
pgn_file="$RESULTS_DIR/single_game_test_${timestamp}.pgn"
cutechess_log="$RESULTS_DIR/single_game_test_${timestamp}_cutechess.log"

echo -e "${YELLOW}Running single game:${NC}"
echo "PGN: $pgn_file"
echo "Log: $cutechess_log"
echo ""

# Run cutechess-cli with detailed output
echo -e "${YELLOW}Cutechess command:${NC}"
cmd="\"$cutechess_cmd\" -engine cmd=\"$chess_cmd\" name=\"ChessEngine\" proto=uci -engine cmd=\"$opponent_cmd\" name=\"WeakStockfish\" proto=uci option.Hash=1 option.\"Skill Level\"=0 -each tc=\"1+1\" -games 1 -concurrency 1 -event \"Single Game Test\" -site \"Local Testing\" -pgnout \"$pgn_file\""
echo "$cmd"
echo ""

echo -e "${YELLOW}Starting game...${NC}"
start_time=$(date +%s)

# Run the command and capture output
if eval "$cmd" 2>&1 | tee "$cutechess_log"; then
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    echo -e "${GREEN}✅ Game completed in ${duration}s${NC}"
else
    end_time=$(date +%s)
    duration=$((end_time - start_time)) 
    echo -e "${RED}❌ Game failed after ${duration}s${NC}"
fi

echo ""
echo -e "${YELLOW}=== CUTECHESS OUTPUT ===${NC}"
if [[ -f "$cutechess_log" ]]; then
    cat "$cutechess_log"
else
    echo "No cutechess log found"
fi

echo ""
echo -e "${YELLOW}=== PGN RESULT ===${NC}"
if [[ -f "$pgn_file" ]]; then
    cat "$pgn_file"
    echo ""
    echo -e "${BLUE}Game result analysis:${NC}"
    
    # Check for illegal moves
    if grep -q "illegal move" "$pgn_file" || grep -q "makes an illegal move" "$pgn_file"; then
        echo -e "${RED}🚨 ILLEGAL MOVE DETECTED!${NC}"
        grep -n "illegal move\|makes an illegal move" "$pgn_file"
    else
        echo -e "${GREEN}✅ No illegal moves detected${NC}"
    fi
    
    # Show game result
    if grep -q "Result" "$pgn_file"; then
        result=$(grep "Result" "$pgn_file" | head -1)
        echo "Game result: $result"
    fi
    
    # Count moves
    move_count=$(grep -o '[0-9]\+\.' "$pgn_file" | wc -l || echo "0")
    echo "Total moves: $move_count"
    
else
    echo "No PGN file found"
fi

echo ""
echo -e "${YELLOW}=== UCI DEBUG LOGS ===${NC}"
latest_uci_log=$(ls -t /tmp/uci_debug_*.log 2>/dev/null | head -1)
if [[ -n "$latest_uci_log" ]]; then
    echo "Latest UCI log: $latest_uci_log"
    echo "Last 20 lines:"
    tail -20 "$latest_uci_log"
else
    echo "No UCI debug logs found"
fi