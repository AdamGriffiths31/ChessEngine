#!/usr/bin/env python3
"""
UCI Position to FEN Converter

This script converts UCI position commands to FEN notation.
Usage: python uci_to_fen.py "position startpos moves d2d4 g8f6 ..."
"""

import sys

try:
    import chess
except ImportError:
    print("Error: python-chess library not found.")
    print("Install it with: pip install python-chess")
    sys.exit(1)


def parse_uci_position(uci_command):
    """
    Parse a UCI position command and return the resulting FEN.
    
    Args:
        uci_command (str): UCI command like "position startpos moves d2d4 g8f6 ..."
    
    Returns:
        str: FEN string of the resulting position
    """
    # Remove "position" from the beginning if present
    if uci_command.startswith("position "):
        uci_command = uci_command[9:]
    
    # Initialize board
    board = chess.Board()
    
    # Parse the command
    parts = uci_command.split()
    
    if len(parts) == 0:
        return board.fen()
    
    # Handle starting position
    if parts[0] == "startpos":
        # Already initialized to starting position
        start_idx = 1
    elif parts[0] == "fen":
        # Find where moves start (after the 6 FEN parts)
        fen_parts = parts[1:7]  # FEN has 6 parts
        fen = " ".join(fen_parts)
        board = chess.Board(fen)
        start_idx = 7
    else:
        raise ValueError("UCI command must start with 'startpos' or 'fen'")
    
    # Check if there are moves to apply
    if start_idx < len(parts) and parts[start_idx] == "moves":
        moves = parts[start_idx + 1:]
        
        # Apply each move
        for move_str in moves:
            try:
                move = chess.Move.from_uci(move_str)
                if move in board.legal_moves:
                    board.push(move)
                else:
                    print(f"Warning: Illegal move {move_str} at position {board.fen()}")
                    break
            except ValueError:
                print(f"Warning: Invalid move format {move_str}")
                break
    
    return board.fen()


def main():
    if len(sys.argv) != 2:
        print("Usage: python uci_to_fen.py \"position startpos moves d2d4 g8f6 ...\"")
        print("\nExample:")
        print("python uci_to_fen.py \"position startpos moves d2d4 g8f6 c2c4\"")
        sys.exit(1)
    
    uci_command = sys.argv[1]
    
    try:
        fen = parse_uci_position(uci_command)
        print("FEN:", fen)
        
        # Also show a simple ASCII board
        board = chess.Board(fen)
        print("\nBoard:")
        print(board)
        
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()

# python uci_to_fen.py "position startpos moves d2d4 g8f6 c2c4 g7g6 b1c3 d7d5 c4d5 f6d5 e2e4 d5c3 b2c3 f8g7 g1f3 c7c5 a1b1 e8g8 f1e2 d8a5 e1h1 c5d4 a2a3 d4c3 g2g3 c8h3 a3a4 b8a6 g3g4 h3f1 h2h3 a8d8 h3h4 f1e2 e4e5 d8d1 f3e1 d1e1"

# python uci_to_fen.py "position startpos moves e2e4 e7e5 g1f3 b8c6 f1b5 a7a6 b5a4 d7d6 e1h1 b7b5 a2a3 b5a4 b2b3 g7g6 c2c3 f8h6 d2d3 c8d7 g2g3 a8b8 h2h3 f7f5 b3b4 d8f6 c3c4 h6g7 d3d4 f5e4 g3g4 f6f3 d1f3 e4f3 h3h4 g8e7 b4b5 a6b5 c4c5 h8f8 d4d5 e7d5 g4g5 g7f6 h4h5 c6d4 c5c6 d7f5 h5h6 f6e7 b1d2 d5c3 d2b1 c3e2"
