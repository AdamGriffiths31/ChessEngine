#!/bin/bash

# Chess Engine Debug Tool
# Runs a game with 5 seconds per move and enhanced output

set -e

echo "=== Chess Engine Game ==="
echo "Starting game with enhanced output (5s per move)..."

# Build the engine
echo "Building UCI engine..."
cd /home/adam/Documents/git/ChessEngine
go build -o tools/bin/uci cmd/uci/main.go
cd tools/scripts

# Clean previous games
rm -f ../results/game.pgn

echo "Running with enhanced output options:"
echo "- Live ratings and outcomes"
echo "- Detailed game information"
echo

../engines/cutechess-cli \
  -engine cmd=../bin/uci name="White-Engine" proto=uci \
  -engine cmd=../bin/uci name="Black-Engine" proto=uci \
  -each tc=5+0 \
  -games 1 \
  -concurrency 1 \
  -ratinginterval 1 \
  -outcomeinterval 1 \
  -event "Debug Game" \
  -site "Local Testing" \
  -pgnout ../results/game.pgn

echo
echo "=== Game Results ==="
if [ -f ../results/game.pgn ]; then
    cat ../results/game.pgn
else
    echo "No game.pgn file found"
fi

echo
echo "Game completed!"