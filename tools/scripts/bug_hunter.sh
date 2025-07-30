#!/bin/bash

# Bug Hunter Script - Runs real games until f6f7 illegal move bug is reproduced
# This script captures EXACT UCI communication as sent by cutechess-cli

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
RESULTS_DIR="$PROJECT_ROOT/tools/results"
ENGINE_PATH="$PROJECT_ROOT/cmd/uci/main.go"
ENGINES_JSON="$TOOLS_DIR/engines.json"

echo -e "${BLUE}🔍 BUG HUNTER: Searching for f6f7 illegal move bug${NC}"
echo "Project root: $PROJECT_ROOT"
echo "Results dir: $RESULTS_DIR"
echo ""

# Ensure results directory exists
mkdir -p "$RESULTS_DIR"

# Build the engine (same way as benchmark.sh)
echo -e "${YELLOW}Building ChessEngine...${NC}"
cd "$PROJECT_ROOT"
if ! go build -o tools/bin/uci cmd/uci/main.go; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Build successful${NC}"

# Bug hunting function
hunt_for_bug() {
    local game_num=$1
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local pgn_file="$RESULTS_DIR/bug_hunt_${timestamp}_game${game_num}.pgn"
    local cutechess_log="$RESULTS_DIR/bug_hunt_${timestamp}_game${game_num}_cutechess.log"
    local engine_log="/tmp/uci_debug_$(date +%s).log"
    
    echo -e "${BLUE}🎯 Game $game_num: Hunting for bug...${NC}"
    echo "  PGN: $(basename "$pgn_file")"
    echo "  Cutechess log: $(basename "$cutechess_log")"
    echo "  Engine UCI log: $(basename "$engine_log")"
    
    # Run cutechess-cli with our engine vs weak opponent (same setup as benchmark.sh)
    # Use 1+1 time control (1 minute + 1 second increment) to match original bug conditions
    local chess_cmd="$TOOLS_DIR/bin/uci"
    local opponent_cmd="$TOOLS_DIR/engines/stockfish"
    local opponent_options="option.Hash=1 option.\"Skill Level\"=0"
    
    timeout 180s "$TOOLS_DIR/engines/cutechess-cli" \
        -engine cmd="$chess_cmd" name="ChessEngine" proto=uci \
        -engine cmd="$opponent_cmd" name="WeakStockfish" proto=uci $opponent_options \
        -each tc="1+1" \
        -games 1 \
        -concurrency 1 \
        -event "Bug Hunt Game $game_num" \
        -site "Local Testing" \
        -pgnout "$pgn_file" \
        &> "$cutechess_log" || true
    
    # Check if illegal move occurred
    if grep -q "illegal move" "$cutechess_log" || grep -q "makes an illegal move" "$pgn_file"; then
        echo -e "${RED}🚨 ILLEGAL MOVE DETECTED IN GAME $game_num! 🚨${NC}"
        
        # Check specifically for f6f7
        if grep -q "f6f7" "$cutechess_log" || grep -q "f6f7" "$pgn_file"; then
            echo -e "${RED}🎯 TARGET FOUND: f6f7 illegal move reproduced!${NC}"
            echo ""
            echo -e "${YELLOW}=== BUG REPRODUCTION SUCCESSFUL ===${NC}"
            echo "Game number: $game_num"
            echo "Timestamp: $timestamp"
            echo "PGN file: $pgn_file"
            echo "Cutechess log: $cutechess_log"
            echo "UCI debug log: $engine_log"
            echo ""
            
            # Show the critical parts
            echo -e "${YELLOW}=== ILLEGAL MOVE FROM PGN ===${NC}"
            grep -n "f6f7" "$pgn_file" || true
            echo ""
            
            echo -e "${YELLOW}=== ILLEGAL MOVE FROM CUTECHESS LOG ===${NC}"
            grep -n -A5 -B5 "f6f7" "$cutechess_log" || true
            echo ""
            
            echo -e "${YELLOW}=== FINAL UCI MESSAGES FROM ENGINE LOG ===${NC}"
            if [[ -f "$engine_log" ]]; then
                tail -50 "$engine_log" | grep -E "(FINAL-MOVE|bestmove|f6f7)" || true
            else
                echo "Engine log not found at $engine_log"
                # Try to find the most recent UCI debug log
                latest_log=$(ls -t /tmp/uci_debug_*.log 2>/dev/null | head -1)
                if [[ -n "$latest_log" ]]; then
                    echo "Found latest engine log: $latest_log"
                    tail -50 "$latest_log" | grep -E "(FINAL-MOVE|bestmove|f6f7)" || true
                fi
            fi
            echo ""
            
            return 0  # Bug found!
        else
            echo -e "${YELLOW}⚠️  Different illegal move detected (not f6f7)${NC}"
            grep -n "illegal move" "$cutechess_log" || true
            grep -n "makes an illegal move" "$pgn_file" || true
        fi
    fi
    
    # Show game result and clean up if no bug found
    if [[ -f "$pgn_file" ]]; then
        local result=$(grep "Result" "$pgn_file" 2>/dev/null | head -1 | sed 's/.*Result "\([^"]*\)".*/\1/' || echo "Unknown")
        local duration=$(grep "GameDuration" "$pgn_file" 2>/dev/null | head -1 | sed 's/.*GameDuration "\([^"]*\)".*/\1/' || echo "Unknown")
        echo -e "${GREEN}✅ Game $game_num: $result (${duration}) - Clean${NC}"
        
        # Clean up if no bug found (keep disk space manageable)
        if ! grep -q "illegal move" "$cutechess_log" && ! grep -q "makes an illegal move" "$pgn_file"; then
            rm -f "$pgn_file" "$cutechess_log"
        fi
    else
        echo -e "${RED}❌ Game $game_num: No PGN file created${NC}"
    fi
    
    return 1  # Bug not found yet
}

# Main hunting loop
echo -e "${BLUE}🏁 Starting bug hunt...${NC}"
echo "Will run games until f6f7 bug is reproduced"
echo "Press Ctrl+C to stop"
echo ""

game_count=0
start_time=$(date +%s)

while true; do
    game_count=$((game_count + 1))
    
    if hunt_for_bug $game_count; then
        echo -e "${GREEN}🎉 SUCCESS: Bug reproduced after $game_count games!${NC}"
        break
    fi
    
    # Show progress every 10 games
    if (( game_count % 10 == 0 )); then
        elapsed=$(($(date +%s) - start_time))
        echo -e "${BLUE}📊 Progress: $game_count games completed in ${elapsed}s (avg $(echo "scale=1; $elapsed/$game_count" | bc)s per game)${NC}"
    fi
    
    # Small delay to avoid overwhelming the system
    sleep 0.5
done

end_time=$(date +%s)
total_time=$((end_time - start_time))

echo ""
echo -e "${GREEN}=== BUG HUNT COMPLETED ===${NC}"
echo "Games played: $game_count"
echo "Total time: ${total_time}s"
echo "f6f7 bug successfully reproduced!"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Analyze the captured UCI logs to see exact message sequence"
echo "2. Compare with our test results to find the difference"
echo "3. Fix the root cause"