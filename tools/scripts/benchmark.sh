#!/bin/bash

# Interactive benchmark tool for ChessEngine
# Allows selecting opponents, time controls, and tracks results

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$(dirname "$SCRIPT_DIR")"
ENGINE_DIR="$TOOLS_DIR"
RESULTS_DIR="$TOOLS_DIR/results"
ENGINES_JSON="$TOOLS_DIR/engines.json"
BENCHMARK_LOG="$ENGINE_DIR/../history.md"

# Ensure directories exist
mkdir -p "$RESULTS_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== ChessEngine Benchmark Tool ===${NC}"
echo

# Parse available engines from engines.json
get_engines() {
    if [ ! -f "$ENGINES_JSON" ]; then
        echo -e "${RED}Error: engines.json not found at $ENGINES_JSON${NC}"
        exit 1
    fi
    
    # Extract engine names (simple parsing, assumes standard format)
    grep '"name"' "$ENGINES_JSON" | sed 's/.*"name": *"\([^"]*\)".*/\1/'
}

# Display available engines
show_engines() {
    echo -e "${YELLOW}Available engines:${NC}"
    local i=1
    while IFS= read -r engine; do
        echo "  $i) $engine"
        ((i++))
    done < <(get_engines)
    echo
}

# Select opponent engine
select_opponent() {
    local engines=()
    while IFS= read -r engine; do
        engines+=("$engine")
    done < <(get_engines)
    
    local chess_engine_index=-1
    
    # Find ChessEngine index
    for i in "${!engines[@]}"; do
        if [ "${engines[$i]}" = "ChessEngine" ]; then
            chess_engine_index=$i
            break
        fi
    done
    
    show_engines
    
    while true; do
        echo -n "Select opponent engine (number): "
        read selection
        
        if [[ "$selection" =~ ^[0-9]+$ ]] && [ "$selection" -ge 1 ] && [ "$selection" -le "${#engines[@]}" ]; then
            local selected_index=$((selection - 1))
            
            # Don't allow ChessEngine vs ChessEngine
            if [ "$selected_index" -eq "$chess_engine_index" ]; then
                echo -e "${RED}Cannot benchmark ChessEngine against itself. Please select a different opponent.${NC}"
                continue
            fi
            
            OPPONENT="${engines[$selected_index]}"
            break
        else
            echo -e "${RED}Invalid selection. Please enter a number between 1 and ${#engines[@]}.${NC}"
        fi
    done
}

# Select time control
select_time_control() {
    echo -e "${YELLOW}Available time controls:${NC}"
    echo "  1) Bullet (2+2) - 2 minutes + 2 second increment"
    echo "  2) Blitz (3+0) - 3 minutes per game"
    echo "  3) Rapid (10+0) - 10 minutes per game"
    echo "  4) Long (30+0) - 30 minutes per game"
    echo "  5) Fixed time (30s per move)"
    echo "  6) Fixed time (60s per move)"
    echo "  7) Custom"
    echo
    
    while true; do
        echo -n "Select time control (number): "
        read tc_selection
        
        case $tc_selection in
            1)
                TIME_CONTROL="2:00+2"
                TC_DESC="Bullet (2+2)"
                break
                ;;
            2)
                TIME_CONTROL="3:00+0"
                TC_DESC="Blitz (3+0)"
                break
                ;;
            3)
                TIME_CONTROL="10:00+0"
                TC_DESC="Rapid (10+0)"
                break
                ;;
            4)
                TIME_CONTROL="30:00+0"
                TC_DESC="Long (30+0)"
                break
                ;;
            5)
                TIME_CONTROL="st=30"
                TC_DESC="Fixed 30s/move"
                break
                ;;
            6)
                TIME_CONTROL="st=60"
                TC_DESC="Fixed 60s/move"
                break
                ;;
            7)
                echo -n "Enter custom time control (e.g., 5+3, 120, etc.): "
                read TIME_CONTROL
                TC_DESC="Custom ($TIME_CONTROL)"
                break
                ;;
            *)
                echo -e "${RED}Invalid selection. Please enter a number between 1 and 7.${NC}"
                ;;
        esac
    done
}

# Select thread count for ChessEngine
select_thread_count() {
    echo -e "${YELLOW}ChessEngine thread count:${NC}"
    echo "  1) Single-threaded (1 thread) - Strongest tactical play"
    echo "  2) Light parallel (2 threads) - Good balance"
    echo "  3) Medium parallel (4 threads) - Balanced speed/strength"
    echo "  4) High parallel (8 threads) - Maximum speed"
    echo "  5) Custom"
    echo
    
    while true; do
        echo -n "Select thread count (number): "
        read thread_selection
        
        case $thread_selection in
            1)
                THREAD_COUNT=1
                THREAD_DESC="1 thread"
                break
                ;;
            2)
                THREAD_COUNT=2
                THREAD_DESC="2 threads"
                break
                ;;
            3)
                THREAD_COUNT=4
                THREAD_DESC="4 threads"
                break
                ;;
            4)
                THREAD_COUNT=8
                THREAD_DESC="8 threads"
                break
                ;;
            5)
                while true; do
                    echo -n "Enter number of threads (1-32): "
                    read THREAD_COUNT
                    if [[ "$THREAD_COUNT" =~ ^[0-9]+$ ]] && [ "$THREAD_COUNT" -gt 0 ] && [ "$THREAD_COUNT" -le 32 ]; then
                        THREAD_DESC="$THREAD_COUNT threads"
                        break
                    else
                        echo -e "${RED}Please enter a number between 1 and 32.${NC}"
                    fi
                done
                break
                ;;
            *)
                echo -e "${RED}Invalid selection. Please enter a number between 1 and 5.${NC}"
                ;;
        esac
    done
}

# Select number of games
select_game_count() {
    echo -e "${YELLOW}Number of games:${NC}"
    echo "  1) Quick test (3 games)"
    echo "  2) Short session (10 games)"
    echo "  3) Medium session (25 games)"
    echo "  4) Long session (50 games)"
    echo "  5) Extended session (100 games)"
    echo "  6) Custom"
    echo
    
    while true; do
        echo -n "Select number of games (number): "
        read game_selection
        
        case $game_selection in
            1)
                GAME_COUNT=3
                break
                ;;
            2)
                GAME_COUNT=10
                break
                ;;
            3)
                GAME_COUNT=25
                break
                ;;
            4)
                GAME_COUNT=50
                break
                ;;
            5)
                GAME_COUNT=100
                break
                ;;
            6)
                while true; do
                    echo -n "Enter number of games: "
                    read GAME_COUNT
                    if [[ "$GAME_COUNT" =~ ^[0-9]+$ ]] && [ "$GAME_COUNT" -gt 0 ]; then
                        break
                    else
                        echo -e "${RED}Please enter a positive number.${NC}"
                    fi
                done
                break
                ;;
            *)
                echo -e "${RED}Invalid selection. Please enter a number between 1 and 6.${NC}"
                ;;
        esac
    done
}

# Get notes from user
get_notes() {
    echo -e "${YELLOW}Notes (optional):${NC}"
    echo "Enter notes about this benchmark (e.g., 'Fixed move validation', 'Added opening book', etc.)"
    echo "Press Enter for no notes, or type your notes:"
    echo -n "> "
    read NOTES
    
    # If empty, set default
    if [ -z "$NOTES" ]; then
        NOTES="-"
    fi
}

# Ask if user wants to record data
ask_record_data() {
    echo -e "${YELLOW}Record data to markdown file?${NC}"
    echo -n "Save results to benchmark history? (Y/n): "
    read record_choice
    
    if [[ "$record_choice" =~ ^[Nn]$ ]]; then
        RECORD_DATA=false
    else
        RECORD_DATA=true
    fi
}

# Get engine command from engines.json
get_engine_command() {
    local engine_name="$1"
    # Simple extraction - assumes standard JSON format
    grep -A 10 "\"name\": \"$engine_name\"" "$ENGINES_JSON" | grep '"command"' | sed 's/.*"command": *"\([^"]*\)".*/\1/' | head -1
}

# Get engine options from engines.json  
get_engine_options() {
    local engine_name="$1"
    local options=""
    local in_engine=false
    local in_options=false
    
    while IFS= read -r line; do
        if [[ "$line" =~ \"name\":[[:space:]]*\"$engine_name\" ]]; then
            in_engine=true
        elif [ "$in_engine" = true ] && [[ "$line" =~ \"options\" ]]; then
            in_options=true
        elif [ "$in_options" = true ] && [[ "$line" =~ ^[[:space:]]*\} ]]; then
            break
        elif [ "$in_options" = true ] && [[ "$line" =~ \"([^\"]*)\": ]]; then
            local key=$(echo "$line" | sed 's/.*"\([^"]*\)": *"\([^"]*\)".*/\1/')
            local value=$(echo "$line" | sed 's/.*"\([^"]*\)": *"\([^"]*\)".*/\2/')
            
            # Handle options with spaces in the name
            if [[ "$key" == *" "* ]]; then
                options="${options}option.\"$key\"=$value "
            else
                options="${options}option.$key=$value "
            fi
        fi
    done < "$ENGINES_JSON"
    
    echo "$options"
}

# Run benchmark
run_benchmark() {
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local pgn_file="$RESULTS_DIR/benchmark_${timestamp}.pgn"
    
    echo -e "${BLUE}=== Starting Benchmark ===${NC}"
    echo "ChessEngine ($THREAD_DESC) vs $OPPONENT"
    echo "Time Control: $TC_DESC"
    echo "Games: $GAME_COUNT"
    echo "Results will be saved to: $pgn_file"
    echo
    
    # Build ChessEngine
    echo -e "${YELLOW}Building ChessEngine...${NC}"
    cd "$ENGINE_DIR/.."
    go build -o tools/bin/uci cmd/uci/main.go
    cd "$SCRIPT_DIR"
    
    # Get engine commands and options
    local chess_cmd=$(get_engine_command "ChessEngine")
    local opponent_cmd=$(get_engine_command "$OPPONENT")
    local opponent_options=$(get_engine_options "$OPPONENT")
    
    echo -e "${YELLOW}Starting games...${NC}"
    
    # Build and execute cutechess-cli command using eval for proper option handling
    local base_cmd="\"$TOOLS_DIR/engines/cutechess-cli\""
    base_cmd="$base_cmd -engine cmd=\"$chess_cmd\" name=\"ChessEngine\" proto=uci option.Threads=$THREAD_COUNT"
    base_cmd="$base_cmd -engine cmd=\"$opponent_cmd\" name=\"$OPPONENT\" proto=uci"
    
    # Add opponent options if they exist
    if [ -n "$opponent_options" ]; then
        base_cmd="$base_cmd $opponent_options"
    fi
    
    # Handle different time control formats
    if [[ "$TIME_CONTROL" =~ ^st= ]]; then
        # Fixed seconds per move - use st parameter globally, not per engine
        base_cmd="$base_cmd -$TIME_CONTROL"
    else
        # Traditional time control - use tc parameter per engine  
        base_cmd="$base_cmd -each tc=\"$TIME_CONTROL\""
    fi
    base_cmd="$base_cmd -games \"$GAME_COUNT\""
    base_cmd="$base_cmd -concurrency 1"
    base_cmd="$base_cmd -ratinginterval 1"
    base_cmd="$base_cmd -outcomeinterval 1"
    base_cmd="$base_cmd -event \"ChessEngine Benchmark\""
    base_cmd="$base_cmd -site \"Local Testing\""
    base_cmd="$base_cmd -pgnout \"$pgn_file\""
    
    # Capture cutechess-cli output to check for illegal moves
    local cutechess_output_file="$RESULTS_DIR/cutechess_output_${timestamp}.log"
    
    echo -e "${YELLOW}Running cutechess-cli (output logged to: $cutechess_output_file)${NC}"
    
    # Run cutechess-cli and capture both stdout and stderr
    if ! eval "$base_cmd" 2>&1 | tee "$cutechess_output_file"; then
        echo -e "${RED}cutechess-cli failed to run successfully${NC}"
        exit 1
    fi
    
    # Check for illegal moves in the output
    check_for_illegal_moves "$cutechess_output_file" "$pgn_file" "$timestamp"
    
    # Analyze results
    analyze_results "$pgn_file" "$timestamp"
}


# Check for illegal moves and dump debug logs if found
check_for_illegal_moves() {
    local cutechess_output="$1"
    local pgn_file="$2" 
    local timestamp="$3"
    
    echo -e "${YELLOW}Checking for illegal moves...${NC}"
    
    # Check cutechess-cli output for illegal move messages
    if grep -q "illegal move" "$cutechess_output" || grep -q "makes an illegal move" "$pgn_file"; then
        echo -e "${RED}🚨 ILLEGAL MOVE DETECTED! 🚨${NC}"
        echo
        
        # Extract illegal move details from cutechess output
        echo -e "${RED}=== CUTECHESS-CLI OUTPUT ===\n"
        grep -A 5 -B 5 "illegal move" "$cutechess_output" || echo "No detailed illegal move info in cutechess output"
        echo -e "${NC}"
        
        # Extract illegal move details from PGN file
        echo -e "${RED}=== PGN GAME DETAILS ===\n"
        grep -A 10 -B 5 "illegal move" "$pgn_file" || echo "No illegal move info in PGN"
        echo -e "${NC}"
        
        # Find the most recent UCI debug log file
        local uci_debug_log=""
        local game_engine_log=""
        
        # Look for the most recent debug logs (they contain timestamp in filename)
        uci_debug_log=$(ls -t /tmp/uci_debug_* 2>/dev/null | head -1)
        game_engine_log=$(ls -t "$ENGINE_DIR"/../uci/logs/game_engine_* 2>/dev/null | head -1)
        
        echo -e "${RED}=== DEBUG LOG ANALYSIS ===\n"
        echo "Most recent UCI debug log: $uci_debug_log"
        echo "Most recent game engine log: $game_engine_log"
        echo -e "${NC}"
        
        # Dump the critical sections of the UCI debug log
        if [ -n "$uci_debug_log" ] && [ -f "$uci_debug_log" ]; then
            echo -e "${RED}=== UCI DEBUG LOG (Last 100 lines) ===${NC}"
            echo "File: $uci_debug_log"
            echo "---"
            tail -100 "$uci_debug_log"
            echo -e "${RED}=========================================${NC}"
            echo
            
            # Look for the final bestmove that caused the issue
            echo -e "${RED}=== FINAL BESTMOVE ANALYSIS ===${NC}"
            grep -A 10 -B 5 "SENDING FINAL BESTMOVE" "$uci_debug_log" | tail -20
            echo -e "${RED}=================================${NC}"
            echo
        else
            echo -e "${YELLOW}No UCI debug log found at $uci_debug_log${NC}"
        fi
        
        # Dump critical sections of game engine log  
        if [ -n "$game_engine_log" ] && [ -f "$game_engine_log" ]; then
            echo -e "${RED}=== GAME ENGINE LOG (Last 50 lines) ===${NC}"
            echo "File: $game_engine_log"
            echo "---"
            tail -50 "$game_engine_log"
            echo -e "${RED}=======================================${NC}"
            echo
        else
            echo -e "${YELLOW}No game engine log found at $game_engine_log${NC}"
        fi
        
        # Stop the script here for investigation
        echo -e "${RED}🛑 BENCHMARK STOPPED DUE TO ILLEGAL MOVE${NC}"
        echo -e "${YELLOW}Debug information has been printed above.${NC}"
        echo -e "${YELLOW}Investigate the logs to understand why the illegal move was generated.${NC}"
        echo
        echo -e "${BLUE}Files for investigation:${NC}"
        echo "  - PGN file: $pgn_file"
        echo "  - Cutechess output: $cutechess_output"
        [ -n "$uci_debug_log" ] && echo "  - UCI debug log: $uci_debug_log"
        [ -n "$game_engine_log" ] && echo "  - Game engine log: $game_engine_log"
        echo
        
        exit 1
    else
        echo -e "${GREEN}✅ No illegal moves detected${NC}"
    fi
}

# Analyze and display results
analyze_results() {
    local pgn_file="$1"
    local timestamp="$2"
    
    echo
    echo -e "${BLUE}=== Benchmark Results ===${NC}"
    
    if [ ! -f "$pgn_file" ]; then
        echo -e "${RED}Results file not found: $pgn_file${NC}"
        return 1
    fi
    
    local total_games=$(grep -c "Result" "$pgn_file" || echo "0")
    local wins=0
    local losses=0
    local draws=0
    
    # Count results properly - White and Result are on separate lines
    
    # Parse the file to match games correctly
    wins=0
    losses=0
    draws=0
    
    # Read line by line to track game context
    local in_chessengine_white_game=false
    local in_chessengine_black_game=false
    
    while IFS= read -r line; do
        if [[ "$line" =~ ^\[White\ \"ChessEngine\"\] ]]; then
            in_chessengine_white_game=true
            in_chessengine_black_game=false
        elif [[ "$line" =~ ^\[Black\ \"ChessEngine\"\] ]]; then
            in_chessengine_white_game=false
            in_chessengine_black_game=true
        elif [[ "$line" =~ ^\[White\ \".*\"\] ]] && [[ ! "$line" =~ ^\[White\ \"ChessEngine\"\] ]]; then
            in_chessengine_white_game=false
        elif [[ "$line" =~ ^\[Black\ \".*\"\] ]] && [[ ! "$line" =~ ^\[Black\ \"ChessEngine\"\] ]]; then
            in_chessengine_black_game=false
        elif [[ "$line" =~ ^\[Result\ \"([^\"]*)\"\] ]]; then
            result="${BASH_REMATCH[1]}"
            
            if [ "$in_chessengine_white_game" = true ]; then
                if [ "$result" = "1-0" ]; then
                    wins=$((wins + 1))
                elif [ "$result" = "0-1" ]; then
                    losses=$((losses + 1))
                elif [ "$result" = "1/2-1/2" ]; then
                    draws=$((draws + 1))
                fi
            elif [ "$in_chessengine_black_game" = true ]; then
                if [ "$result" = "0-1" ]; then
                    wins=$((wins + 1))
                elif [ "$result" = "1-0" ]; then
                    losses=$((losses + 1))
                elif [ "$result" = "1/2-1/2" ]; then
                    draws=$((draws + 1))
                fi
            fi
            
            # Reset flags after processing result
            in_chessengine_white_game=false
            in_chessengine_black_game=false
        fi
    done < "$pgn_file"
    
    # Calculate score percentage without bc
    local score=0
    if [ $total_games -gt 0 ]; then
        # Calculate score as (wins + draws*0.5) / total * 100, using integer arithmetic
        local points=$((wins * 2 + draws))  # wins worth 2 points, draws worth 1 point
        local max_points=$((total_games * 2))  # max possible points
        score=$((points * 100 / max_points))
    fi
    
    echo "Total Games: $total_games"
    echo -e "${GREEN}Wins: $wins${NC}"
    echo -e "${RED}Losses: $losses${NC}"
    echo "Draws: $draws"
    echo "Score: ${score}%"
    echo
    
    # Log to markdown file if requested
    if [ "$RECORD_DATA" = true ]; then
        log_to_markdown "$timestamp" "$total_games" "$wins" "$losses" "$draws" "$score"
        echo -e "${GREEN}Results logged to: $BENCHMARK_LOG${NC}"
    else
        echo -e "${YELLOW}Results not logged to markdown file (as requested)${NC}"
    fi
    
    echo -e "${GREEN}Full PGN saved to: $pgn_file${NC}"
}

# Log results to markdown
log_to_markdown() {
    local timestamp="$1"
    local total="$2"
    local wins="$3"
    local losses="$4"
    local draws="$5"
    local score="$6"
    
    # Create markdown file if it doesn't exist
    if [ ! -f "$BENCHMARK_LOG" ]; then
        cat > "$BENCHMARK_LOG" << 'EOF'
# ChessEngine Benchmark History

This file tracks the performance of ChessEngine against various opponents over time.

## Results Summary

| Date | Opponent | Time Control | Threads | Games | Wins | Losses | Draws | Score | Notes |
|------|----------|--------------|---------|-------|------|--------|-------|-------|-------|
EOF
    fi
    
    # Append new result
    # timestamp format is YYYYMMDD_HHMMSS, extract date and time parts
    local date_part="${timestamp:0:8}"      # YYYYMMDD
    local time_part="${timestamp:9:6}"      # HHMMSS
    
    # Format date as YYYY-MM-DD
    local date="${date_part:0:4}-${date_part:4:2}-${date_part:6:2}"
    
    # Format time as HH:MM
    local time="${time_part:0:2}:${time_part:2:2}"
    
    echo "| $date $time | $OPPONENT | $TC_DESC | $THREAD_DESC | $total | $wins | $losses | $draws | ${score}% | $NOTES |" >> "$BENCHMARK_LOG"
}

# Main execution
main() {
    select_opponent
    echo -e "${GREEN}Selected opponent: $OPPONENT${NC}"
    echo
    
    select_time_control
    echo -e "${GREEN}Selected time control: $TC_DESC${NC}"
    echo
    
    select_thread_count
    echo -e "${GREEN}Selected threads: $THREAD_DESC${NC}"
    echo
    
    select_game_count
    echo -e "${GREEN}Selected games: $GAME_COUNT${NC}"
    echo
    
    get_notes
    echo -e "${GREEN}Notes: $NOTES${NC}"
    echo
    
    ask_record_data
    echo -e "${GREEN}Record data: $([ "$RECORD_DATA" = true ] && echo "Yes" || echo "No")${NC}"
    echo
    
    echo -e "${YELLOW}Summary:${NC}"
    echo "  ChessEngine ($THREAD_DESC) vs $OPPONENT"
    echo "  Time Control: $TC_DESC"
    echo "  Games: $GAME_COUNT"
    echo "  Notes: $NOTES"
    echo "  Record data: $([ "$RECORD_DATA" = true ] && echo "Yes" || echo "No")"
    echo
    
    echo -n "Proceed with benchmark? (y/N): "
    read confirm
    
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        run_benchmark
    else
        echo "Benchmark cancelled."
        exit 0
    fi
}

# Check dependencies
if [ ! -f "$TOOLS_DIR/engines/cutechess-cli" ]; then
    echo -e "${RED}Error: cutechess-cli not found at $TOOLS_DIR/engines/cutechess-cli${NC}"
    exit 1
fi

if ! command -v bc &> /dev/null; then
    echo -e "${YELLOW}Warning: bc not found. Score percentages may not be calculated.${NC}"
fi

# Run main function
main