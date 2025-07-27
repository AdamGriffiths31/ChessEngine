#!/bin/bash

# Manual UCI Testing Script
# This script allows interactive testing of the UCI engine implementation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== UCI Engine Manual Testing Script ===${NC}"
echo

# Build the UCI engine
echo -e "${YELLOW}Building UCI engine...${NC}"
cd "$(dirname "$0")/.."
go build -o bin/uci cmd/uci/main.go

if [ ! -f "bin/uci" ]; then
    echo -e "${RED}Failed to build UCI engine${NC}"
    exit 1
fi

echo -e "${GREEN}UCI engine built successfully${NC}"
echo

# Function to send command and show response
send_uci_command() {
    local cmd="$1"
    echo -e "${BLUE}> $cmd${NC}"
    echo "$cmd" | timeout 5s ./bin/uci 2>/dev/null || true
    echo
}

# Function to run automated test sequence
run_automated_tests() {
    echo -e "${YELLOW}Running automated UCI test sequence...${NC}"
    echo
    
    # Test 1: Basic UCI handshake
    echo -e "${YELLOW}Test 1: UCI handshake${NC}"
    send_uci_command "uci"
    
    # Test 2: Ready check
    echo -e "${YELLOW}Test 2: Ready check${NC}"
    send_uci_command "isready"
    
    # Test 3: New game
    echo -e "${YELLOW}Test 3: New game${NC}"
    send_uci_command "ucinewgame"
    
    # Test 4: Set position to starting position
    echo -e "${YELLOW}Test 4: Set starting position${NC}"
    send_uci_command "position startpos"
    
    # Test 5: Search for best move from starting position
    echo -e "${YELLOW}Test 5: Search from starting position (depth 3)${NC}"
    echo "position startpos" | ./bin/uci &
    UCI_PID=$!
    sleep 0.1
    echo "go depth 3" | ./bin/uci &
    sleep 3
    echo "quit" | ./bin/uci
    wait $UCI_PID 2>/dev/null || true
    echo
    
    # Test 6: Set position with moves
    echo -e "${YELLOW}Test 6: Position with moves${NC}"
    send_uci_command "position startpos moves e2e4 e7e5"
    
    # Test 7: Search from position with moves
    echo -e "${YELLOW}Test 7: Search after e2e4 e7e5 (depth 2)${NC}"
    {
        echo "position startpos moves e2e4 e7e5"
        echo "go depth 2"
        sleep 2
        echo "quit"
    } | ./bin/uci
    echo
    
    # Test 8: FEN position
    echo -e "${YELLOW}Test 8: FEN position${NC}"
    send_uci_command "position fen rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"
    
    echo -e "${GREEN}Automated tests completed${NC}"
}

# Function to run interactive mode
run_interactive_mode() {
    echo -e "${YELLOW}Starting interactive UCI mode...${NC}"
    echo -e "${BLUE}Type UCI commands and press Enter. Type 'quit' to exit.${NC}"
    echo -e "${BLUE}Common commands: uci, isready, position startpos, go depth 3, quit${NC}"
    echo
    
    ./bin/uci
}

# Main menu
while true; do
    echo -e "${YELLOW}Choose test mode:${NC}"
    echo "1) Run automated test sequence"
    echo "2) Interactive UCI mode"
    echo "3) Quick unit tests"
    echo "4) Exit"
    echo
    read -p "Enter choice (1-4): " choice
    
    case $choice in
        1)
            run_automated_tests
            ;;
        2)
            run_interactive_mode
            ;;
        3)
            echo -e "${YELLOW}Running unit tests...${NC}"
            go test ./uci/... -v
            echo
            ;;
        4)
            echo -e "${GREEN}Exiting...${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid choice. Please enter 1-4.${NC}"
            ;;
    esac
    
    echo
    read -p "Press Enter to continue..."
    echo
done