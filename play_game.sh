#!/bin/bash

# Script to play chess game as white with debug enabled and make e2e4 move

echo "Starting chess game script..."

# Create input file with the required responses
cat << 'EOF' > game_inputs.txt
2
1
3
2
e2e4
quit
yes
EOF

# Run the chess engine and pipe the inputs, capture output to txt file
go run main.go < game_inputs.txt > game_output.txt 2>&1

# Clean up temporary input file
rm game_inputs.txt

echo "Game completed! Output saved to game_output.txt"
echo "Contents of game_output.txt:"
echo "=========================="
cat game_output.txt
