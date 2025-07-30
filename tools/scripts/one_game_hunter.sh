#!/bin/bash

# Run one game synchronously and show all details
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

echo -e "${BLUE}🎯 One Game Bug Hunter${NC}"
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

# Test files
timestamp=$(date +%Y%m%d_%H%M%S)
pgn_file="$RESULTS_DIR/one_game_hunt_${timestamp}.pgn"
cutechess_log="$RESULTS_DIR/one_game_hunt_${timestamp}_cutechess.log"

echo -e "${YELLOW}Running one game to test setup:${NC}"
echo "PGN: $pgn_file"
echo "Log: $cutechess_log"
echo ""

# Paths
chess_cmd="$TOOLS_DIR/bin/uci"
opponent_cmd="$TOOLS_DIR/engines/stockfish"
cutechess_cmd="$TOOLS_DIR/engines/cutechess-cli"

echo -e "${YELLOW}Checking paths:${NC}"
echo "ChessEngine: $chess_cmd $(test -f "$chess_cmd" && echo "✅" || echo "❌")"
echo "Stockfish: $opponent_cmd $(test -f "$opponent_cmd" && echo "✅" || echo "❌")"
echo "Cutechess: $cutechess_cmd $(test -f "$cutechess_cmd" && echo "✅" || echo "❌")"
echo ""

# Run cutechess-cli with the same setup as the working single game test
echo -e "${YELLOW}Cutechess command:${NC}"
cmd="\"$cutechess_cmd\" -engine cmd=\"$chess_cmd\" name=\"ChessEngine\" proto=uci -engine cmd=\"$opponent_cmd\" name=\"WeakStockfish\" proto=uci option.Hash=1 option.\"Skill Level\"=0 -each tc=\"1+1\" -games 1 -concurrency 1 -event \"One Game Hunt\" -site \"Local Testing\" -pgnout \"$pgn_file\""
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
    echo -e "${GREEN}✅ PGN file created successfully${NC}"
    cat "$pgn_file"
    echo ""
    
    # Check for illegal moves
    if grep -q "illegal move" "$pgn_file" || grep -q "makes an illegal move" "$pgn_file"; then
        echo -e "${RED}🚨 ILLEGAL MOVE DETECTED!${NC}"
        grep -n "illegal move\|makes an illegal move" "$pgn_file"
        
        # Check specifically for f6f7
        if grep -q "f6f7" "$pgn_file"; then
            echo -e "${RED}🎯 TARGET FOUND: f6f7 illegal move reproduced!${NC}"
            exit 0
        fi
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
    echo "Total moves played: $move_count"
    
else
    echo -e "${RED}❌ No PGN file created${NC}"
    echo "This indicates cutechess-cli failed to run properly"
fi

echo ""
echo -e "${YELLOW}=== FILES CREATED ===${NC}"
ls -la "$RESULTS_DIR"/one_game_hunt_${timestamp}*