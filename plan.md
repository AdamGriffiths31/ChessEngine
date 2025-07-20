# Chess Engine Move Generation Performance Optimization Guide

## Overview

This document provides detailed implementation instructions for three major performance optimizations to the chess engine's move generation system. These optimizations target the most significant bottlenecks identified in the current implementation.

**Current Performance**: ~4.3M nodes/second (Kiwipete depth 5)  
**Target Performance**: 10M+ nodes/second  
**Expected Total Improvement**: 2-3x performance increase

## Optimization 1: Eliminate Make/Unmake Overhead with Pin Detection

### Problem Statement

The `filterLegalMoves()` function in `bitboard_generator.go` performs make/unmake for every pseudo-legal move to check if it leaves the king in check. This is the single largest performance bottleneck.

### Solution: Implement Pin Detection and Smart Legal Move Filtering

#### Step 1: Add Pin Detection Data Structures

**File**: `game/moves/bitboard_generator.go`

```go
// Update the BitboardMoveGenerator struct
type BitboardMoveGenerator struct {
    tempMoveList *MoveList
    
    // ADD THESE NEW FIELDS:
    pinnedPieces   board.Bitboard  // Bitmap of pinned pieces
    pinRays        [64]board.Bitboard // Pin ray for each square (0 if not pinned)
    checkers       board.Bitboard  // Pieces giving check
    checkMask      board.Bitboard  // Squares to block check or capture checker
    inCheck        bool
    doubleCheck    bool
}
```

#### Step 2: Implement Pin Detection Algorithm

**File**: `game/moves/pin_detection.go` (new file)

```go
package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// PinDetector handles detection of pinned pieces and check information
type PinDetector struct {
    pinnedPieces board.Bitboard
    pinRays      [64]board.Bitboard
    checkers     board.Bitboard
    checkMask    board.Bitboard
}

// DetectPinsAndChecks finds all pinned pieces and checking pieces
func (pd *PinDetector) DetectPinsAndChecks(b *board.Board, kingSquare int, color board.BitboardColor) {
    // Reset state
    pd.pinnedPieces = 0
    pd.checkers = 0
    pd.checkMask = board.Bitboard(0xFFFFFFFFFFFFFFFF) // All squares initially
    for i := 0; i < 64; i++ {
        pd.pinRays[i] = 0
    }
    
    if kingSquare < 0 || kingSquare > 63 {
        return
    }
    
    oppositeColor := board.OppositeBitboardColor(color)
    occupied := b.AllPieces
    friendlyPieces := b.GetColorBitboard(color)
    
    // Get enemy pieces that can pin/check
    enemyRooksQueens := b.GetPieceBitboard(getRookPiece(oppositeColor)) | 
                       b.GetPieceBitboard(getQueenPiece(oppositeColor))
    enemyBishopsQueens := b.GetPieceBitboard(getBishopPiece(oppositeColor)) | 
                          b.GetPieceBitboard(getQueenPiece(oppositeColor))
    
    // Check straight lines (rook/queen attacks)
    pd.checkStraightLines(b, kingSquare, friendlyPieces, enemyRooksQueens, occupied)
    
    // Check diagonals (bishop/queen attacks)
    pd.checkDiagonals(b, kingSquare, friendlyPieces, enemyBishopsQueens, occupied)
    
    // Check knight attacks
    pd.checkKnightAttacks(b, kingSquare, oppositeColor)
    
    // Check pawn attacks
    pd.checkPawnAttacks(b, kingSquare, color, oppositeColor)
    
    // If in check, compute check mask
    if pd.checkers != 0 {
        pd.computeCheckMask(kingSquare)
    }
}

// checkStraightLines checks for pins and checks along ranks and files
func (pd *PinDetector) checkStraightLines(b *board.Board, kingSquare int, 
    friendlyPieces, enemySliders, occupied board.Bitboard) {
    
    kingFile, kingRank := board.SquareToFileRank(kingSquare)
    
    // Check each of the 4 straight directions
    directions := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
    
    for _, dir := range directions {
        fileStep, rankStep := dir[0], dir[1]
        currentFile, currentRank := kingFile + fileStep, kingRank + rankStep
        
        var ray board.Bitboard
        var firstPiece board.Bitboard
        friendlyCount := 0
        
        // Slide along the direction
        for currentFile >= 0 && currentFile <= 7 && currentRank >= 0 && currentRank <= 7 {
            square := board.FileRankToSquare(currentFile, currentRank)
            ray = ray.SetBit(square)
            
            if occupied.HasBit(square) {
                if friendlyPieces.HasBit(square) {
                    if friendlyCount == 0 {
                        firstPiece = board.Bitboard(1) << square
                    }
                    friendlyCount++
                } else if enemySliders.HasBit(square) {
                    // Found enemy slider
                    if friendlyCount == 0 {
                        // Direct check
                        pd.checkers = pd.checkers.SetBit(square)
                    } else if friendlyCount == 1 {
                        // Pin
                        pd.pinnedPieces |= firstPiece
                        pd.pinRays[firstPiece.LSB()] = ray | (board.Bitboard(1) << kingSquare)
                    }
                    break
                } else {
                    // Non-sliding piece blocks ray
                    break
                }
            }
            
            currentFile += fileStep
            currentRank += rankStep
        }
    }
}

// checkDiagonals checks for pins and checks along diagonals
func (pd *PinDetector) checkDiagonals(b *board.Board, kingSquare int, 
    friendlyPieces, enemySliders, occupied board.Bitboard) {
    
    kingFile, kingRank := board.SquareToFileRank(kingSquare)
    
    // Check each of the 4 diagonal directions
    directions := [][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
    
    for _, dir := range directions {
        fileStep, rankStep := dir[0], dir[1]
        currentFile, currentRank := kingFile + fileStep, kingRank + rankStep
        
        var ray board.Bitboard
        var firstPiece board.Bitboard
        friendlyCount := 0
        
        // Slide along the direction
        for currentFile >= 0 && currentFile <= 7 && currentRank >= 0 && currentRank <= 7 {
            square := board.FileRankToSquare(currentFile, currentRank)
            ray = ray.SetBit(square)
            
            if occupied.HasBit(square) {
                if friendlyPieces.HasBit(square) {
                    if friendlyCount == 0 {
                        firstPiece = board.Bitboard(1) << square
                    }
                    friendlyCount++
                } else if enemySliders.HasBit(square) {
                    // Found enemy slider
                    if friendlyCount == 0 {
                        // Direct check
                        pd.checkers = pd.checkers.SetBit(square)
                    } else if friendlyCount == 1 {
                        // Pin
                        pd.pinnedPieces |= firstPiece
                        pd.pinRays[firstPiece.LSB()] = ray | (board.Bitboard(1) << kingSquare)
                    }
                    break
                } else {
                    // Non-sliding piece blocks ray
                    break
                }
            }
            
            currentFile += fileStep
            currentRank += rankStep
        }
    }
}

// checkKnightAttacks checks for knight checks
func (pd *PinDetector) checkKnightAttacks(b *board.Board, kingSquare int, oppositeColor board.BitboardColor) {
    enemyKnights := b.GetPieceBitboard(getKnightPiece(oppositeColor))
    knightAttacks := board.GetKnightAttacks(kingSquare)
    
    pd.checkers |= knightAttacks & enemyKnights
}

// checkPawnAttacks checks for pawn checks
func (pd *PinDetector) checkPawnAttacks(b *board.Board, kingSquare int, 
    color, oppositeColor board.BitboardColor) {
    
    enemyPawns := b.GetPieceBitboard(getPawnPiece(oppositeColor))
    pawnAttacks := board.GetPawnAttacks(kingSquare, color)
    
    pd.checkers |= pawnAttacks & enemyPawns
}

// computeCheckMask computes the mask of squares that block check or capture checker
func (pd *PinDetector) computeCheckMask(kingSquare int) {
    numCheckers := pd.checkers.PopCount()
    
    if numCheckers == 0 {
        pd.checkMask = board.Bitboard(0xFFFFFFFFFFFFFFFF) // Not in check
    } else if numCheckers == 1 {
        // Single check - can block or capture
        checkerSquare := pd.checkers.LSB()
        pd.checkMask = board.GetBetween(kingSquare, checkerSquare) | pd.checkers
    } else {
        // Double check - only king moves allowed
        pd.checkMask = 0
    }
}

// Helper functions to get piece types
func getRookPiece(color board.BitboardColor) board.Piece {
    if color == board.BitboardWhite {
        return board.WhiteRook
    }
    return board.BlackRook
}

func getBishopPiece(color board.BitboardColor) board.Piece {
    if color == board.BitboardWhite {
        return board.WhiteBishop
    }
    return board.BlackBishop
}

func getQueenPiece(color board.BitboardColor) board.Piece {
    if color == board.BitboardWhite {
        return board.WhiteQueen
    }
    return board.BlackQueen
}

func getKnightPiece(color board.BitboardColor) board.Piece {
    if color == board.BitboardWhite {
        return board.WhiteKnight
    }
    return board.BlackKnight
}

func getPawnPiece(color board.BitboardColor) board.Piece {
    if color == board.BitboardWhite {
        return board.WhitePawn
    }
    return board.BlackPawn
}

func getKingPiece(player Player) board.Piece {
    if player == White {
        return board.WhiteKing
    }
    return board.BlackKing
}

func playerToColor(player Player) board.PieceColor {
    if player == White {
        return board.WhiteColor
    }
    return board.BlackColor
}
```

#### Step 3: Update filterLegalMoves to Use Pin Detection

**File**: `game/moves/bitboard_generator.go`

Replace the existing `filterLegalMoves` method:

```go
// filterLegalMoves filters out moves that would leave the king in check
func (bmg *BitboardMoveGenerator) filterLegalMoves(b *board.Board, player Player, moves *MoveList) *MoveList {
    legalMoves := GetMoveList()
    
    // Find king position
    kingPiece := getKingPiece(player)
    kingBitboard := b.GetPieceBitboard(kingPiece)
    if kingBitboard == 0 {
        return legalMoves // No king found
    }
    kingSquare := kingBitboard.LSB()
    
    // Detect pins and checks
    color := board.ConvertToBitboardColor(playerToColor(player))
    pinDetector := &PinDetector{}
    pinDetector.DetectPinsAndChecks(b, kingSquare, color)
    
    // Store pin information in generator
    bmg.pinnedPieces = pinDetector.pinnedPieces
    bmg.checkers = pinDetector.checkers
    bmg.checkMask = pinDetector.checkMask
    bmg.inCheck = pinDetector.checkers != 0
    bmg.doubleCheck = pinDetector.checkers.PopCount() > 1
    
    // If double check, only king moves are legal
    if bmg.doubleCheck {
        for i := 0; i < moves.Count; i++ {
            move := moves.Moves[i]
            if board.IsKing(move.Piece) {
                // Still need to verify king doesn't move into check
                if bmg.isKingMoveLegal(b, move, player) {
                    legalMoves.AddMove(move)
                }
            }
        }
        return legalMoves
    }
    
    // Process each move
    for i := 0; i < moves.Count; i++ {
        move := moves.Moves[i]
        fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
        toSquare := board.FileRankToSquare(move.To.File, move.To.Rank)
        
        // King moves always need special handling
        if board.IsKing(move.Piece) {
            if bmg.isKingMoveLegal(b, move, player) {
                legalMoves.AddMove(move)
            }
            continue
        }
        
        // If in check, move must block or capture checker
        if bmg.inCheck {
            if !bmg.checkMask.HasBit(toSquare) {
                continue // Move doesn't address check
            }
        }
        
        // Handle pinned pieces
        if bmg.pinnedPieces.HasBit(fromSquare) {
            // Pinned piece can only move along pin ray
            pinRay := pinDetector.pinRays[fromSquare]
            if !pinRay.HasBit(toSquare) {
                continue // Move leaves pin ray
            }
        }
        
        // Handle en passant special case
        if move.IsEnPassant {
            if bmg.isEnPassantLegal(b, move, player, kingSquare) {
                legalMoves.AddMove(move)
            }
            continue
        }
        
        // Move is legal
        legalMoves.AddMove(move)
    }
    
    return legalMoves
}

// isKingMoveLegal checks if a king move is legal
func (bmg *BitboardMoveGenerator) isKingMoveLegal(b *board.Board, move board.Move, player Player) bool {
    // Temporarily move king to check if destination is safe
    toSquare := board.FileRankToSquare(move.To.File, move.To.Rank)
    oppositeColor := board.OppositeBitboardColor(board.ConvertToBitboardColor(playerToColor(player)))
    
    // For castling, check intermediate squares
    if move.IsCastling {
        return bmg.isCastlingPathSafe(b, move, oppositeColor)
    }
    
    // Regular king move - check if destination is attacked
    // Need to consider that king is not on original square
    return !bmg.isSquareAttackedAfterKingMove(b, toSquare, move.From, oppositeColor)
}

// isSquareAttackedAfterKingMove checks if a square is attacked after king moves away
func (bmg *BitboardMoveGenerator) isSquareAttackedAfterKingMove(b *board.Board, targetSquare int, 
    kingFrom board.Square, byColor board.BitboardColor) bool {
    
    // Temporarily remove king from board for attack calculation
    kingFromSquare := board.FileRankToSquare(kingFrom.File, kingFrom.Rank)
    
    // Create modified occupancy without king
    modifiedOccupancy := b.AllPieces.ClearBit(kingFromSquare)
    
    // Check attacks with modified occupancy
    // ... implement using modified occupancy for sliding piece attacks
    
    return false // Simplified - implement full logic
}

// isEnPassantLegal checks if en passant doesn't expose king to check
func (bmg *BitboardMoveGenerator) isEnPassantLegal(b *board.Board, move board.Move, 
    player Player, kingSquare int) bool {
    
    // En passant can expose king to check on the rank
    // This needs special handling with make/unmake
    moveExecutor := &MoveExecutor{}
    history := moveExecutor.MakeMove(b, move, func(*board.Board, board.Move) {})
    
    color := board.ConvertToBitboardColor(playerToColor(player))
    oppositeColor := board.OppositeBitboardColor(color)
    legal := !b.IsSquareAttackedByColor(kingSquare, oppositeColor)
    
    moveExecutor.UnmakeMove(b, history)
    
    return legal
}

// isCastlingPathSafe verifies castling path is not under attack
func (bmg *BitboardMoveGenerator) isCastlingPathSafe(b *board.Board, move board.Move, 
    oppositeColor board.BitboardColor) bool {
    
    // King must not be in check, pass through check, or end in check
    if move.To.File == 6 { // Kingside
        return !b.IsSquareAttackedByColor(board.FileRankToSquare(4, move.From.Rank), oppositeColor) &&
               !b.IsSquareAttackedByColor(board.FileRankToSquare(5, move.From.Rank), oppositeColor) &&
               !b.IsSquareAttackedByColor(board.FileRankToSquare(6, move.From.Rank), oppositeColor)
    } else { // Queenside
        return !b.IsSquareAttackedByColor(board.FileRankToSquare(4, move.From.Rank), oppositeColor) &&
               !b.IsSquareAttackedByColor(board.FileRankToSquare(3, move.From.Rank), oppositeColor) &&
               !b.IsSquareAttackedByColor(board.FileRankToSquare(2, move.From.Rank), oppositeColor)
    }
}
```

#### Step 4: Add Tests for Pin Detection

**File**: `game/moves/pin_detection_test.go` (new file)

```go
package moves

import (
    "testing"
    "github.com/AdamGriffiths31/ChessEngine/board"
)

func TestPinDetection(t *testing.T) {
    testCases := []struct {
        name         string
        fen          string
        kingSquare   string
        pinnedPieces []string
        checkers     []string
    }{
        {
            name:         "bishop_pin",
            fen:          "8/8/8/b7/2n5/8/4K3/8 w - - 0 1",
            kingSquare:   "e2",
            pinnedPieces: []string{"c4"},
            checkers:     []string{},
        },
        {
            name:         "rook_pin",
            fen:          "8/8/8/8/r1n1K3/8/8/8 w - - 0 1",
            kingSquare:   "e4",
            pinnedPieces: []string{"c4"},
            checkers:     []string{},
        },
        {
            name:         "check_from_queen",
            fen:          "8/8/8/8/4K3/8/8/3q4 w - - 0 1",
            kingSquare:   "e4",
            pinnedPieces: []string{},
            checkers:     []string{"d1"},
        },
        {
            name:         "double_check",
            fen:          "8/8/8/2n5/4K3/8/8/3r4 w - - 0 1",
            kingSquare:   "e4",
            pinnedPieces: []string{},
            checkers:     []string{"c5", "d1"},
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            b, err := board.FromFEN(tc.fen)
            if err != nil {
                t.Fatalf("Failed to parse FEN: %v", err)
            }
            
            kingSquare := board.StringToSquare(tc.kingSquare)
            pd := &PinDetector{}
            pd.DetectPinsAndChecks(b, kingSquare, board.BitboardWhite)
            
            // Check pinned pieces
            for _, pinnedSq := range tc.pinnedPieces {
                square := board.StringToSquare(pinnedSq)
                if !pd.pinnedPieces.HasBit(square) {
                    t.Errorf("Expected %s to be pinned", pinnedSq)
                }
            }
            
            // Check checkers
            for _, checkerSq := range tc.checkers {
                square := board.StringToSquare(checkerSq)
                if !pd.checkers.HasBit(square) {
                    t.Errorf("Expected %s to be giving check", checkerSq)
                }
            }
        })
    }
}
```

## Optimization 2: Implement Staged Move Generation

### Problem Statement

Generating all moves at once is wasteful when only a subset might be needed (e.g., in alpha-beta search with good move ordering).

### Solution: Lazy/Staged Move Generation

#### Step 1: Create Staged Move Generator

**File**: `game/moves/staged_generator.go` (new file)

```go
package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// MoveStage represents different stages of move generation
type MoveStage int

const (
    StageHashMove MoveStage = iota
    StageCapturesWinning
    StageKillerMoves
    StageCapturesLosing
    StageQuietMoves
    StageDone
)

// StagedMoveGenerator generates moves in stages for better move ordering
type StagedMoveGenerator struct {
    board          *board.Board
    player         Player
    generator      *BitboardMoveGenerator
    pinDetector    *PinDetector
    
    // State
    stage          MoveStage
    currentBatch   *MoveList
    batchIndex     int
    
    // Special moves
    hashMove       *board.Move
    killerMoves    [2]board.Move
    
    // Flags
    generated      [StageDone]bool
    capturesOnly   bool
    
    // Legal move filtering state
    kingSquare     int
    pinnedPieces   board.Bitboard
    checkers       board.Bitboard
    checkMask      board.Bitboard
    inCheck        bool
    doubleCheck    bool
}

// NewStagedMoveGenerator creates a new staged move generator
func NewStagedMoveGenerator(b *board.Board, player Player) *StagedMoveGenerator {
    smg := &StagedMoveGenerator{
        board:        b,
        player:       player,
        generator:    NewBitboardMoveGenerator(),
        pinDetector:  &PinDetector{},
        currentBatch: GetMoveList(),
        stage:        StageHashMove,
    }
    
    // Pre-calculate pin and check information
    smg.initializeLegalMoveInfo()
    
    return smg
}

// initializeLegalMoveInfo sets up pin and check detection
func (smg *StagedMoveGenerator) initializeLegalMoveInfo() {
    kingPiece := getKingPiece(smg.player)
    kingBitboard := smg.board.GetPieceBitboard(kingPiece)
    if kingBitboard == 0 {
        return
    }
    
    smg.kingSquare = kingBitboard.LSB()
    color := board.ConvertToBitboardColor(playerToColor(smg.player))
    
    smg.pinDetector.DetectPinsAndChecks(smg.board, smg.kingSquare, color)
    smg.pinnedPieces = smg.pinDetector.pinnedPieces
    smg.checkers = smg.pinDetector.checkers
    smg.checkMask = smg.pinDetector.checkMask
    smg.inCheck = smg.checkers != 0
    smg.doubleCheck = smg.checkers.PopCount() > 1
}

// SetHashMove sets the hash move for priority generation
func (smg *StagedMoveGenerator) SetHashMove(move board.Move) {
    smg.hashMove = &move
}

// SetKillerMoves sets killer moves for move ordering
func (smg *StagedMoveGenerator) SetKillerMoves(killer1, killer2 board.Move) {
    smg.killerMoves[0] = killer1
    smg.killerMoves[1] = killer2
}

// NextMove returns the next move in the staged generation
func (smg *StagedMoveGenerator) NextMove() (*board.Move, bool) {
    for smg.stage < StageDone {
        // Return moves from current batch if available
        if smg.batchIndex < smg.currentBatch.Count {
            move := smg.currentBatch.Moves[smg.batchIndex]
            smg.batchIndex++
            
            // Verify move is legal
            if smg.isMoveLegal(move) {
                return &move, true
            }
            continue
        }
        
        // Generate next batch
        smg.advanceStage()
    }
    
    return nil, false
}

// advanceStage moves to the next generation stage
func (smg *StagedMoveGenerator) advanceStage() {
    smg.batchIndex = 0
    smg.currentBatch.Clear()
    
    switch smg.stage {
    case StageHashMove:
        if smg.hashMove != nil && smg.isMovePseudoLegal(*smg.hashMove) {
            smg.currentBatch.AddMove(*smg.hashMove)
        }
        
    case StageCapturesWinning:
        smg.generateCaptures(true)
        
    case StageKillerMoves:
        for _, killer := range smg.killerMoves {
            if !MovesEqual(killer, board.Move{}) && smg.isMovePseudoLegal(killer) {
                smg.currentBatch.AddMove(killer)
            }
        }
        
    case StageCapturesLosing:
        smg.generateCaptures(false)
        
    case StageQuietMoves:
        smg.generateQuietMoves()
    }
    
    smg.stage++
}

// generateCaptures generates capture moves
func (smg *StagedMoveGenerator) generateCaptures(winningOnly bool) {
    // Generate all captures
    captures := GetMoveList()
    defer ReleaseMoveList(captures)
    
    smg.generator.generateCapturesBitboard(smg.board, smg.player, captures)
    
    // Filter and order captures
    for i := 0; i < captures.Count; i++ {
        move := captures.Moves[i]
        
        if winningOnly {
            // Use Static Exchange Evaluation (SEE) to filter
            if smg.seeScore(move) >= 0 {
                smg.currentBatch.AddMove(move)
            }
        } else {
            // Losing captures
            if smg.seeScore(move) < 0 {
                smg.currentBatch.AddMove(move)
            }
        }
    }
}

// generateQuietMoves generates non-capture moves
func (smg *StagedMoveGenerator) generateQuietMoves() {
    allMoves := GetMoveList()
    defer ReleaseMoveList(allMoves)
    
    smg.generator.GenerateAllMovesBitboard(smg.board, smg.player, allMoves)
    
    // Filter out captures and special moves already generated
    for i := 0; i < allMoves.Count; i++ {
        move := allMoves.Moves[i]
        
        if !move.IsCapture && !smg.isSpecialMoveAlreadyGenerated(move) {
            smg.currentBatch.AddMove(move)
        }
    }
}

// isMoveLegal checks if a move is legal given pins and checks
func (smg *StagedMoveGenerator) isMoveLegal(move board.Move) bool {
    // Double check - only king moves allowed
    if smg.doubleCheck {
        return board.IsKing(move.Piece) && smg.isKingMoveSafe(move)
    }
    
    fromSquare := board.FileRankToSquare(move.From.File, move.From.Rank)
    toSquare := board.FileRankToSquare(move.To.File, move.To.Rank)
    
    // King moves need special handling
    if board.IsKing(move.Piece) {
        return smg.isKingMoveSafe(move)
    }
    
    // If in check, move must block or capture checker
    if smg.inCheck && !smg.checkMask.HasBit(toSquare) {
        return false
    }
    
    // Handle pinned pieces
    if smg.pinnedPieces.HasBit(fromSquare) {
        pinRay := smg.pinDetector.pinRays[fromSquare]
        if !pinRay.HasBit(toSquare) {
            return false
        }
    }
    
    // En passant needs special handling
    if move.IsEnPassant {
        return smg.isEnPassantSafe(move)
    }
    
    return true
}

// isMovePseudoLegal checks if a move is pseudo-legal (piece can make the move)
func (smg *StagedMoveGenerator) isMovePseudoLegal(move board.Move) bool {
    // Verify piece is on from square
    piece := smg.board.GetPiece(move.From.Rank, move.From.File)
    if piece != move.Piece {
        return false
    }
    
    // Basic validation
    return true
}

// isSpecialMoveAlreadyGenerated checks if move was generated in earlier stages
func (smg *StagedMoveGenerator) isSpecialMoveAlreadyGenerated(move board.Move) bool {
    if smg.hashMove != nil && MovesEqual(move, *smg.hashMove) {
        return true
    }
    
    for _, killer := range smg.killerMoves {
        if MovesEqual(move, killer) {
            return true
        }
    }
    
    return false
}

// isKingMoveSafe checks if king move is safe
func (smg *StagedMoveGenerator) isKingMoveSafe(move board.Move) bool {
    // Implement king safety check
    return true
}

// isEnPassantSafe checks if en passant is safe
func (smg *StagedMoveGenerator) isEnPassantSafe(move board.Move) bool {
    // En passant can expose king to check, needs special handling
    // TODO: Implement en passant safety check
    return true
}

// seeScore calculates Static Exchange Evaluation score
func (smg *StagedMoveGenerator) seeScore(move board.Move) int {
    // Simplified SEE - positive for winning captures
    // TODO: Implement full SEE
    victimValue := getPieceValue(move.Captured)
    attackerValue := getPieceValue(move.Piece)
    
    // Simple approximation
    return victimValue - attackerValue
}

// getPieceValue returns material value for SEE
func getPieceValue(piece board.Piece) int {
    switch piece {
    case board.WhitePawn, board.BlackPawn:
        return 100
    case board.WhiteKnight, board.BlackKnight:
        return 320
    case board.WhiteBishop, board.BlackBishop:
        return 330
    case board.WhiteRook, board.BlackRook:
        return 500
    case board.WhiteQueen, board.BlackQueen:
        return 900
    case board.WhiteKing, board.BlackKing:
        return 20000
    default:
        return 0
    }
}

// Release releases resources held by the generator
func (smg *StagedMoveGenerator) Release() {
    if smg.currentBatch != nil {
        ReleaseMoveList(smg.currentBatch)
    }
    if smg.generator != nil {
        smg.generator.Release()
    }
}
```

#### Step 2: Add Capture-Only Generation to Bitboard Generator

**File**: `game/moves/bitboard_generator.go`

Add this method:

```go
// generateCapturesBitboard generates only capture moves
func (bmg *BitboardMoveGenerator) generateCapturesBitboard(b *board.Board, player Player, moveList *MoveList) {
    var bitboardColor board.BitboardColor
    if player == White {
        bitboardColor = board.BitboardWhite
    } else {
        bitboardColor = board.BitboardBlack
    }
    
    enemyPieces := b.GetColorBitboard(board.OppositeBitboardColor(bitboardColor))
    
    // Generate pawn captures
    bmg.generatePawnCapturesBitboard(b, player, b.GetPieceBitboard(getPawnPiece(bitboardColor)), 
        enemyPieces, moveList, getPromotionRank(player))
    
    // Generate knight captures
    bmg.generatePieceCapturesBitboard(b, player, board.WhiteKnight, enemyPieces, moveList)
    
    // Generate bishop captures
    bmg.generateSlidingCapturesBitboard(b, player, board.WhiteBishop, enemyPieces, moveList)
    
    // Generate rook captures
    bmg.generateSlidingCapturesBitboard(b, player, board.WhiteRook, enemyPieces, moveList)
    
    // Generate queen captures
    bmg.generateSlidingCapturesBitboard(b, player, board.WhiteQueen, enemyPieces, moveList)
    
    // Generate king captures
    bmg.generatePieceCapturesBitboard(b, player, board.WhiteKing, enemyPieces, moveList)
}

// generatePieceCapturesBitboard generates captures for non-sliding pieces
func (bmg *BitboardMoveGenerator) generatePieceCapturesBitboard(b *board.Board, player Player, 
    pieceType board.Piece, enemyPieces board.Bitboard, moveList *MoveList) {
    
    var piece board.Piece
    if player == White {
        piece = pieceType
    } else {
        // Convert to black piece
        piece = board.Piece(byte(pieceType) + 32) // lowercase
    }
    
    pieces := b.GetPieceBitboard(piece)
    
    for pieces != 0 {
        fromSquare, newBitboard := pieces.PopLSB()
        pieces = newBitboard
        
        var attacks board.Bitboard
        switch pieceType {
        case board.WhiteKnight:
            attacks = board.GetKnightAttacks(fromSquare)
        case board.WhiteKing:
            attacks = board.GetKingAttacks(fromSquare)
        }
        
        captures := attacks & enemyPieces
        
        for captures != 0 {
            toSquare, newCaptures := captures.PopLSB()
            captures = newCaptures
            
            capturedPiece := b.GetPieceOnSquare(toSquare)
            toFile, toRank := board.SquareToFileRank(toSquare)
            fromFile, fromRank := board.SquareToFileRank(fromSquare)
            
            move := board.Move{
                From:        board.Square{File: fromFile, Rank: fromRank},
                To:          board.Square{File: toFile, Rank: toRank},
                Piece:       piece,
                Captured:    capturedPiece,
                Promotion:   board.Empty,
                IsCapture:   true,
                IsCastling:  false,
                IsEnPassant: false,
            }
            moveList.AddMove(move)
        }
    }
}

// generateSlidingCapturesBitboard generates captures for sliding pieces
func (bmg *BitboardMoveGenerator) generateSlidingCapturesBitboard(b *board.Board, player Player, 
    pieceType board.Piece, enemyPieces board.Bitboard, moveList *MoveList) {
    
    var piece board.Piece
    if player == White {
        piece = pieceType
    } else {
        piece = board.Piece(byte(pieceType) + 32)
    }
    
    pieces := b.GetPieceBitboard(piece)
    occupancy := b.AllPieces
    
    for pieces != 0 {
        fromSquare, newBitboard := pieces.PopLSB()
        pieces = newBitboard
        
        var attacks board.Bitboard
        switch pieceType {
        case board.WhiteBishop:
            attacks = board.GetBishopAttacks(fromSquare, occupancy)
        case board.WhiteRook:
            attacks = board.GetRookAttacks(fromSquare, occupancy)
        case board.WhiteQueen:
            attacks = board.GetQueenAttacks(fromSquare, occupancy)
        }
        
        captures := attacks & enemyPieces
        
        for captures != 0 {
            toSquare, newCaptures := captures.PopLSB()
            captures = newCaptures
            
            capturedPiece := b.GetPieceOnSquare(toSquare)
            toFile, toRank := board.SquareToFileRank(toSquare)
            fromFile, fromRank := board.SquareToFileRank(fromSquare)
            
            move := board.Move{
                From:        board.Square{File: fromFile, Rank: fromRank},
                To:          board.Square{File: toFile, Rank: toRank},
                Piece:       piece,
                Captured:    capturedPiece,
                Promotion:   board.Empty,
                IsCapture:   true,
                IsCastling:  false,
                IsEnPassant: false,
            }
            moveList.AddMove(move)
        }
    }
}

// Helper function to get promotion rank
func getPromotionRank(player Player) int {
    if player == White {
        return 7
    }
    return 0
}
```

#### Step 3: Integration with Search

**File**: `game/moves/search_integration.go` (new file)

```go
package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// MoveIterator provides an interface for move generation in search
type MoveIterator interface {
    NextMove() (*board.Move, bool)
    Release()
}

// CreateMoveIterator creates appropriate move iterator based on context
func CreateMoveIterator(b *board.Board, player Player, hashMove *board.Move, 
    killers [2]board.Move, capturesOnly bool) MoveIterator {
    
    if capturesOnly {
        // For quiescence search
        return &CaptureIterator{
            board:     b,
            player:    player,
            generator: NewBitboardMoveGenerator(),
        }
    }
    
    // Full staged generation
    smg := NewStagedMoveGenerator(b, player)
    if hashMove != nil {
        smg.SetHashMove(*hashMove)
    }
    smg.SetKillerMoves(killers[0], killers[1])
    
    return smg
}

// CaptureIterator generates only captures for quiescence search
type CaptureIterator struct {
    board     *board.Board
    player    Player
    generator *BitboardMoveGenerator
    moves     *MoveList
    index     int
    generated bool
}

func (ci *CaptureIterator) NextMove() (*board.Move, bool) {
    if !ci.generated {
        ci.moves = GetMoveList()
        ci.generator.generateCapturesBitboard(ci.board, ci.player, ci.moves)
        ci.generated = true
    }
    
    if ci.index < ci.moves.Count {
        move := ci.moves.Moves[ci.index]
        ci.index++
        return &move, true
    }
    
    return nil, false
}

func (ci *CaptureIterator) Release() {
    if ci.moves != nil {
        ReleaseMoveList(ci.moves)
    }
    if ci.generator != nil {
        ci.generator.Release()
    }
}
```

## Optimization 3: Platform-Specific Optimizations

### Problem Statement

Bitboard operations can be further optimized using SIMD instructions and better memory layout.

### Solution: Assembly and Memory Optimizations

#### Step 1: Add Assembly Functions for Critical Operations

**File**: `board/bitboard_amd64.s` (new file)

```assembly
// +build amd64

#include "textflag.h"

// func findFirstBitASM(bitboard uint64) int
TEXT ·findFirstBitASM(SB), NOSPLIT, $0-16
    MOVQ bitboard+0(FP), AX
    BSFQ AX, AX
    JZ   notfound
    MOVQ AX, ret+8(FP)
    RET
notfound:
    MOVQ $-1, ret+8(FP)
    RET

// func popCountASM(bitboard uint64) int
TEXT ·popCountASM(SB), NOSPLIT, $0-16
    MOVQ bitboard+0(FP), AX
    POPCNTQ AX, AX
    MOVQ AX, ret+8(FP)
    RET

// func pextASM(src, mask uint64) uint64
TEXT ·pextASM(SB), NOSPLIT, $0-24
    MOVQ src+0(FP), AX
    MOVQ mask+8(FP), BX
    PEXTQ BX, AX, AX
    MOVQ AX, ret+16(FP)
    RET

// func pdepASM(src, mask uint64) uint64
TEXT ·pdepASM(SB), NOSPLIT, $0-24
    MOVQ src+0(FP), AX
    MOVQ mask+8(FP), BX
    PDEPQ BX, AX, AX
    MOVQ AX, ret+16(FP)
    RET
```

#### Step 2: Add Go Wrapper Functions

**File**: `board/bitboard_amd64.go` (new file)

```go
// +build amd64

package board

import "runtime"

// Assembly function declarations
func findFirstBitASM(bitboard uint64) int
func popCountASM(bitboard uint64) int
func pextASM(src, mask uint64) uint64
func pdepASM(src, mask uint64) uint64

// hasBMI2 checks if the CPU supports BMI2 instructions
var hasBMI2 = checkBMI2Support()

func checkBMI2Support() bool {
    // Check CPU features
    // This is simplified - real implementation would check CPUID
    return runtime.GOARCH == "amd64"
}

// Optimized LSB using assembly
func (b Bitboard) LSBAsm() int {
    if b == 0 {
        return -1
    }
    return findFirstBitASM(uint64(b))
}

// Optimized PopCount using assembly
func (b Bitboard) PopCountAsm() int {
    return popCountASM(uint64(b))
}

// Parallel bit extraction for magic bitboards
func extractBits(src Bitboard, mask Bitboard) Bitboard {
    if hasBMI2 {
        return Bitboard(pextASM(uint64(src), uint64(mask)))
    }
    // Fallback to regular implementation
    return extractBitsSlow(src, mask)
}

func extractBitsSlow(src Bitboard, mask Bitboard) Bitboard {
    var result Bitboard
    var bitPos uint
    
    for mask != 0 {
        lsb := mask.LSB()
        mask = mask.ClearBit(lsb)
        
        if src.HasBit(lsb) {
            result |= (1 << bitPos)
        }
        bitPos++
    }
    
    return result
}
```

#### Step 3: Optimize Memory Layout

**File**: `game/moves/memory_layout.go` (new file)

```go
package moves

import (
    "unsafe"
    "github.com/AdamGriffiths31/ChessEngine/board"
)

// CacheLineSize is the typical CPU cache line size
const CacheLineSize = 64

// AlignedBitboardGenerator uses cache-aligned data structures
type AlignedBitboardGenerator struct {
    _ [0]byte // Force alignment
    
    // Hot data - frequently accessed together (64 bytes)
    occupied      board.Bitboard // 8 bytes
    friendlyMask  board.Bitboard // 8 bytes
    enemyMask     board.Bitboard // 8 bytes
    pinnedPieces  board.Bitboard // 8 bytes
    checkers      board.Bitboard // 8 bytes
    checkMask     board.Bitboard // 8 bytes
    kingSquare    int32          // 4 bytes
    inCheck       bool           // 1 byte
    doubleCheck   bool           // 1 byte
    _ [10]byte                   // Padding to 64 bytes
    
    // Move generation state (separate cache line)
    _ [0]byte
    moveList      *MoveList
    tempList      *MoveList
    pinDetector   *PinDetector
    _ [40]byte    // Padding
    
    // Cold data (rarely accessed)
    _ [0]byte
    stats         GeneratorStats
}

// GeneratorStats tracks performance statistics
type GeneratorStats struct {
    MovesGenerated   uint64
    CapturesGenerated uint64
    ChecksEvaluated  uint64
    PinsDetected     uint64
}

// Ensure proper alignment
var _ = unsafe.Sizeof(AlignedBitboardGenerator{})

// BatchBitboardOps performs multiple bitboard operations in parallel
func BatchBitboardOps(ops []BitboardOp) []board.Bitboard {
    results := make([]board.Bitboard, len(ops))
    
    // Process in batches of 4 for better CPU utilization
    for i := 0; i < len(ops); i += 4 {
        end := i + 4
        if end > len(ops) {
            end = len(ops)
        }
        
        // Unroll loop for batch processing
        for j := i; j < end; j++ {
            switch ops[j].Type {
            case OpAnd:
                results[j] = ops[j].A & ops[j].B
            case OpOr:
                results[j] = ops[j].A | ops[j].B
            case OpXor:
                results[j] = ops[j].A ^ ops[j].B
            case OpAndNot:
                results[j] = ops[j].A &^ ops[j].B
            }
        }
    }
    
    return results
}

// BitboardOp represents a bitboard operation
type BitboardOp struct {
    Type OpType
    A, B board.Bitboard
}

type OpType int

const (
    OpAnd OpType = iota
    OpOr
    OpXor
    OpAndNot
)

// Prefetch hints for better cache utilization
func prefetchBitboards(b *board.Board) {
    // This would use compiler intrinsics in a real implementation
    // Prefetch commonly accessed bitboards
    _ = b.WhitePieces
    _ = b.BlackPieces
    _ = b.AllPieces
}
```

#### Step 4: SIMD Pawn Move Generation

**File**: `game/moves/simd_pawns.go` (new file)

```go
// +build amd64

package moves

import "github.com/AdamGriffiths31/ChessEngine/board"

// generatePawnMovesSIMD uses SIMD-like operations for parallel pawn move generation
func generatePawnMovesSIMD(pawns, empty, enemyPieces board.Bitboard, player Player) (pushes, captures board.Bitboard) {
    if player == White {
        // Single pushes - all pawns move forward one rank in parallel
        singlePushes := (pawns << 8) & empty
        
        // Double pushes - use mask to select only 2nd rank pawns
        rank2Mask := board.Bitboard(0x0000FF00)
        doublePushCandidates := pawns & rank2Mask
        doublePushes := ((doublePushCandidates << 8) & empty) << 8 & empty
        
        pushes = singlePushes | doublePushes
        
        // Captures - parallel shift for both directions
        leftCaptures := ((pawns &^ board.FileA) << 7) & enemyPieces
        rightCaptures := ((pawns &^ board.FileH) << 9) & enemyPieces
        
        captures = leftCaptures | rightCaptures
    } else {
        // Black pawns - mirror operations
        singlePushes := (pawns >> 8) & empty
        
        rank7Mask := board.Bitboard(0x00FF000000000000)
        doublePushCandidates := pawns & rank7Mask
        doublePushes := ((doublePushCandidates >> 8) & empty) >> 8 & empty
        
        pushes = singlePushes | doublePushes
        
        leftCaptures := ((pawns &^ board.FileH) >> 7) & enemyPieces
        rightCaptures := ((pawns &^ board.FileA) >> 9) & enemyPieces
        
        captures = leftCaptures | rightCaptures
    }
    
    return pushes, captures
}

// generateAllPawnMovesSIMD generates all pawn moves using SIMD-like operations
func (bmg *BitboardMoveGenerator) generateAllPawnMovesSIMD(b *board.Board, player Player, moveList *MoveList) {
    var pawnPiece board.Piece
    var bitboardColor board.BitboardColor
    var promotionRank board.Bitboard
    
    if player == White {
        pawnPiece = board.WhitePawn
        bitboardColor = board.BitboardWhite
        promotionRank = board.Rank8
    } else {
        pawnPiece = board.BlackPawn
        bitboardColor = board.BitboardBlack
        promotionRank = board.Rank1
    }
    
    pawns := b.GetPieceBitboard(pawnPiece)
    empty := ^b.AllPieces
    enemyPieces := b.GetColorBitboard(board.OppositeBitboardColor(bitboardColor))
    
    // Generate all pawn moves in parallel
    pushes, captures := generatePawnMovesSIMD(pawns, empty, enemyPieces, player)
    
    // Convert bitboards to move list
    bmg.extractPawnMoves(pushes, captures, promotionRank, player, moveList)
}

// extractPawnMoves converts pawn move bitboards to move list
func (bmg *BitboardMoveGenerator) extractPawnMoves(pushes, captures, promotionRank board.Bitboard, 
    player Player, moveList *MoveList) {
    
    // Extract push moves
    for pushes != 0 {
        to, newPushes := pushes.PopLSB()
        pushes = newPushes
        
        var from int
        if player == White {
            // Check for double push
            if board.GetRank(to) == 3 && !b.AllPieces.HasBit(to-8) {
                from = to - 16
            } else {
                from = to - 8
            }
        } else {
            if board.GetRank(to) == 4 && !b.AllPieces.HasBit(to+8) {
                from = to + 16
            } else {
                from = to + 8
            }
        }
        
        // Add move (handle promotions separately)
        // ... implementation
    }
    
    // Extract capture moves
    // ... similar implementation
}
```

### Implementation Testing

#### Test File for Pin Detection

**File**: `game/moves/optimization_test.go` (new file)

```go
package moves

import (
    "testing"
    "time"
    "github.com/AdamGriffiths31/ChessEngine/board"
)

func BenchmarkMoveGenerationOptimized(b *testing.B) {
    positions := []struct {
        name string
        fen  string
    }{
        {"StartingPosition", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"},
        {"Kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - "},
        {"ComplexMiddlegame", "r2q1rk1/1ppnbppp/p1np4/3Pp3/B3P3/2N1BN2/PPP2PPP/R2Q1RK1 w - - 0 1"},
    }
    
    for _, pos := range positions {
        board, _ := board.FromFEN(pos.fen)
        
        b.Run(pos.name+"_Original", func(b *testing.B) {
            gen := NewBitboardMoveGenerator()
            defer gen.Release()
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                moves := gen.GenerateAllMovesBitboard(board, White)
                ReleaseMoveList(moves)
            }
        })
        
        b.Run(pos.name+"_WithPinDetection", func(b *testing.B) {
            gen := NewBitboardMoveGenerator()
            defer gen.Release()
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                moves := gen.GenerateAllMovesBitboard(board, White)
                ReleaseMoveList(moves)
            }
        })
        
        b.Run(pos.name+"_Staged", func(b *testing.B) {
            b.ResetTimer()
            
            for i := 0; i < b.N; i++ {
                smg := NewStagedMoveGenerator(board, White)
                count := 0
                for {
                    move, ok := smg.NextMove()
                    if !ok {
                        break
                    }
                    count++
                    _ = move
                }
                smg.Release()
            }
        })
    }
}

func TestPinDetectionCorrectness(t *testing.T) {
    testCases := []struct {
        name     string
        fen      string
        expected int // Expected legal move count
    }{
        {
            name:     "PinnedBishop",
            fen:      "8/8/8/b7/2n5/8/4K3/8 w - - 0 1",
            expected: 8, // Only king moves
        },
        {
            name:     "DoubleCheck",
            fen:      "8/8/8/2n5/4K3/8/8/3r4 w - - 0 1",
            expected: 4, // Only king moves to escape
        },
        {
            name:     "PinnedPawn",
            fen:      "8/8/8/r7/4P3/8/4K3/8 w - - 0 1",
            expected: 8, // King moves only, pawn is pinned
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            b, err := board.FromFEN(tc.fen)
            if err != nil {
                t.Fatalf("Failed to parse FEN: %v", err)
            }
            
            gen := NewBitboardMoveGenerator()
            defer gen.Release()
            
            moves := gen.GenerateAllMovesBitboard(b, White)
            defer ReleaseMoveList(moves)
            
            if moves.Count != tc.expected {
                t.Errorf("Expected %d legal moves, got %d", tc.expected, moves.Count)
                for i := 0; i < moves.Count; i++ {
                    move := moves.Moves[i]
                    t.Logf("Move: %s%s", 
                        board.SquareToString(board.FileRankToSquare(move.From.File, move.From.Rank)),
                        board.SquareToString(board.FileRankToSquare(move.To.File, move.To.Rank)))
                }
            }
        })
    }
}

func TestStagedGenerationCorrectness(t *testing.T) {
    fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - "
    b, _ := board.FromFEN(fen)
    
    // Generate all moves at once
    gen := NewBitboardMoveGenerator()
    allMoves := gen.GenerateAllMovesBitboard(b, White)
    gen.Release()
    
    // Generate moves staged
    smg := NewStagedMoveGenerator(b, White)
    stagedMoves := make([]board.Move, 0, 100)
    
    for {
        move, ok := smg.NextMove()
        if !ok {
            break
        }
        stagedMoves = append(stagedMoves, *move)
    }
    smg.Release()
    
    // Compare counts
    if len(stagedMoves) != allMoves.Count {
        t.Errorf("Move count mismatch: all=%d, staged=%d", allMoves.Count, len(stagedMoves))
    }
    
    ReleaseMoveList(allMoves)
}

func BenchmarkPerftWithOptimizations(b *testing.B) {
    fen := "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - "
    board, _ := board.FromFEN(fen)
    
    b.Run("Depth4", func(b *testing.B) {
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            gen := NewGenerator()
            result := PerftWithGenerator(board, 4, White, gen)
            _ = result
            gen.Release()
        }
    })
}
```

## Implementation Checklist

### Phase 1: Pin Detection (Week 1)

- [ ] Implement `PinDetector` struct and methods
- [ ] Add pin ray calculation
- [ ] Implement check detection and check mask
- [ ] Update `filterLegalMoves` to use pin detection
- [ ] Add comprehensive tests for pin detection
- [ ] Benchmark improvement over make/unmake

### Phase 2: Staged Generation (Week 2)

- [ ] Implement `StagedMoveGenerator`
- [ ] Add capture-only generation
- [ ] Implement SEE for capture ordering
- [ ] Add killer move support
- [ ] Create search integration interface
- [ ] Test correctness vs all-at-once generation

### Phase 3: Platform Optimizations (Week 3)

- [ ] Add assembly functions for AMD64
- [ ] Implement cache-aligned data structures
- [ ] Add SIMD-like pawn generation
- [ ] Optimize memory layout
- [ ] Add BMI2 support detection
- [ ] Benchmark on different platforms

## Expected Results

### Performance Targets

- **Pin Detection**: 40-60% reduction in legal move generation time
- **Staged Generation**: 20-30% improvement in search performance
- **Platform Optimizations**: 15-25% improvement in bitboard operations
- **Combined**: 2-3x overall performance improvement

### Verification Commands

```bash
# Run optimization benchmarks
go test -bench=BenchmarkMoveGenerationOptimized -benchtime=10s ./game/moves

# Test correctness
go test -run=TestPinDetection ./game/moves
go test -run=TestStagedGeneration ./game/moves

# Run perft with optimizations
go test -run=TestKiwipeteDepth6Performance -v ./game/moves
```

## Notes for Implementation

1. **Start with Pin Detection** - This provides the biggest immediate gain
2. **Test Thoroughly** - Each optimization must maintain 100% correctness
3. **Profile Before and After** - Use pprof to verify improvements
4. **Consider Memory Usage** - Don't sacrifice memory for minor speed gains
5. **Keep Fallbacks** - Maintain non-optimized paths for debugging
