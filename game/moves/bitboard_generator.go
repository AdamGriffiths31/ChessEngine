// Package moves provides high-performance chess move generation using bitboard operations.
package moves

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
	"strings"
)

// Square constants for castling
const (
	E1 = 4
	G1 = 6
	C1 = 2
	H1 = 7
	A1 = 0
	F1 = 5
	D1 = 3
	B1 = 1
	E8 = 60
	G8 = 62
	C8 = 58
	H8 = 63
	A8 = 56
	F8 = 61
	D8 = 59
	B8 = 57
)

// BitboardMoveGenerator provides high-performance move generation using bitboard operations
// This implementation aims for 3-5x performance improvement over array-based generation
type BitboardMoveGenerator struct {
	// No fields needed - all operations use pooled MoveList objects
}

// NewBitboardMoveGenerator creates a new bitboard-based move generator
func NewBitboardMoveGenerator() *BitboardMoveGenerator {
	return &BitboardMoveGenerator{}
}

// GenerateAllMovesBitboard generates all legal moves using bitboard operations
// This is the main entry point that coordinates all piece-specific generators
func (bmg *BitboardMoveGenerator) GenerateAllMovesBitboard(b *board.Board, player Player) *MoveList {

	moveList := GetMoveList()

	// Generate moves for each piece type using bitboard operations
	bmg.generatePawnMovesBitboard(b, player, moveList)
	bmg.GenerateKnightMovesBitboard(b, player, moveList)
	bmg.generateBishopMovesBitboard(b, player, moveList)
	bmg.generateRookMovesBitboard(b, player, moveList)
	bmg.generateQueenMovesBitboard(b, player, moveList)
	bmg.generateKingMovesBitboard(b, player, moveList)

	// Filter out illegal moves in-place (moves that leave king in check)
	bmg.filterLegalMovesInPlace(b, player, moveList)

	return moveList
}

// generatePawnMovesBitboard generates pawn moves using bitboard shifts and attack patterns
func (bmg *BitboardMoveGenerator) generatePawnMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var pawnPiece board.Piece
	var bitboardColor board.BitboardColor
	var startRank, promotionRank int

	if player == White {
		pawnPiece = board.WhitePawn
		bitboardColor = board.BitboardWhite
		startRank = 1     // 2nd rank (rank index 1)
		promotionRank = 7 // 8th rank (rank index 7) - where pawn arrives to promote
	} else {
		pawnPiece = board.BlackPawn
		bitboardColor = board.BitboardBlack
		startRank = 6     // 7th rank (rank index 6)
		promotionRank = 0 // 1st rank (rank index 0) - where pawn arrives to promote
	}

	pawns := b.GetPieceBitboard(pawnPiece)
	if pawns == 0 {
		return // No pawns to move
	}

	emptySquares := ^b.AllPieces // Invert to get empty squares
	enemyPieces := b.GetColorBitboard(board.OppositeBitboardColor(bitboardColor))

	// Single pawn pushes
	var singlePushes board.Bitboard
	if player == White {
		singlePushes = pawns.ShiftNorth() & emptySquares
	} else {
		singlePushes = pawns.ShiftSouth() & emptySquares
	}

	// Generate moves from single pushes
	for singlePushes != 0 {
		toSquare, newBitboard := singlePushes.PopLSB()
		singlePushes = newBitboard

		var fromSquare int
		if player == White {
			fromSquare = toSquare - 8 // Move came from south
		} else {
			fromSquare = toSquare + 8 // Move came from north
		}

		// Check for promotion
		toFile, toRank := board.SquareToFileRank(toSquare)
		if toRank == promotionRank {
			// Generate all promotion moves
			bmg.addPromotionMoves(moveList, fromSquare, toSquare, pawnPiece, board.Empty, player)
		} else {
			// Regular pawn push
			move := board.Move{
				From:        board.Square{File: fromSquare % 8, Rank: fromSquare / 8},
				To:          board.Square{File: toFile, Rank: toRank},
				Piece:       pawnPiece,
				Captured:    board.Empty,
				Promotion:   board.Empty,
				IsCapture:   false,
				IsCastling:  false,
				IsEnPassant: false,
			}
			moveList.AddMove(move)
		}
	}

	// Double pawn pushes (only from starting rank)
	// Important: both the single push square AND double push square must be empty
	var doublePushes board.Bitboard
	if player == White {
		// Only pawns on starting rank can double push
		startingPawns := pawns & board.RankMask(startRank)
		// Single push must be clear first
		singlePushClear := startingPawns.ShiftNorth() & emptySquares
		// Then double push must also be clear
		doublePushes = singlePushClear.ShiftNorth() & emptySquares
	} else {
		// Only pawns on starting rank can double push
		startingPawns := pawns & board.RankMask(startRank)
		// Single push must be clear first
		singlePushClear := startingPawns.ShiftSouth() & emptySquares
		// Then double push must also be clear
		doublePushes = singlePushClear.ShiftSouth() & emptySquares
	}

	// Generate moves from double pushes
	for doublePushes != 0 {
		toSquare, newBitboard := doublePushes.PopLSB()
		doublePushes = newBitboard

		var fromSquare int
		if player == White {
			fromSquare = toSquare - 16 // Move came from 2 squares south
		} else {
			fromSquare = toSquare + 16 // Move came from 2 squares north
		}

		toFile, toRank := board.SquareToFileRank(toSquare)
		move := board.Move{
			From:        board.Square{File: fromSquare % 8, Rank: fromSquare / 8},
			To:          board.Square{File: toFile, Rank: toRank},
			Piece:       pawnPiece,
			Captured:    board.Empty,
			Promotion:   board.Empty,
			IsCapture:   false,
			IsCastling:  false,
			IsEnPassant: false,
		}
		moveList.AddMove(move)
	}

	// Pawn captures
	bmg.generatePawnCapturesBitboard(b, player, pawns, enemyPieces, moveList, promotionRank)

	// En passant captures
	bmg.generateEnPassantCapturesBitboard(b, player, pawns, moveList)
}

// generatePawnCapturesBitboard generates pawn capture moves using attack patterns
func (bmg *BitboardMoveGenerator) generatePawnCapturesBitboard(b *board.Board, player Player, pawns board.Bitboard, enemyPieces board.Bitboard, moveList *MoveList, promotionRank int) {
	var pawnPiece board.Piece

	if player == White {
		pawnPiece = board.WhitePawn
	} else {
		pawnPiece = board.BlackPawn
	}

	// Iterate through each pawn and check its attack squares
	tempPawns := pawns
	for tempPawns != 0 {
		fromSquare, newBitboard := tempPawns.PopLSB()
		tempPawns = newBitboard

		// Get pawn attacks for this square
		var bitboardColor board.BitboardColor
		if player == White {
			bitboardColor = board.BitboardWhite
		} else {
			bitboardColor = board.BitboardBlack
		}
		attacks := board.GetPawnAttacks(fromSquare, bitboardColor)
		captures := attacks & enemyPieces

		// Generate capture moves
		for captures != 0 {
			toSquare, newCaptures := captures.PopLSB()
			captures = newCaptures

			capturedPiece := b.GetPieceOnSquare(toSquare)
			toFile, toRank := board.SquareToFileRank(toSquare)

			// Check for promotion
			if toRank == promotionRank {
				bmg.addPromotionMoves(moveList, fromSquare, toSquare, pawnPiece, capturedPiece, player)
			} else {
				// Regular capture
				move := board.Move{
					From:        board.Square{File: fromSquare % 8, Rank: fromSquare / 8},
					To:          board.Square{File: toFile, Rank: toRank},
					Piece:       pawnPiece,
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
}

// generateEnPassantCapturesBitboard generates en passant captures
func (bmg *BitboardMoveGenerator) generateEnPassantCapturesBitboard(b *board.Board, player Player, pawns board.Bitboard, moveList *MoveList) {
	enPassantTarget := b.GetEnPassantTarget()
	if enPassantTarget == nil {
		return
	}

	targetSquare := board.FileRankToSquare(enPassantTarget.File, enPassantTarget.Rank)
	var pawnPiece board.Piece
	var capturedPiece board.Piece

	if player == White {
		pawnPiece = board.WhitePawn
		capturedPiece = board.BlackPawn
	} else {
		pawnPiece = board.BlackPawn
		capturedPiece = board.WhitePawn
	}

	// Find pawns that can capture en passant
	// En passant is only valid for pawns on the correct rank:
	// - White pawns must be on 5th rank (rank index 4)
	// - Black pawns must be on 4th rank (rank index 3)
	var enPassantRank int
	if player == White {
		enPassantRank = 4 // 5th rank
	} else {
		enPassantRank = 3 // 4th rank
	}

	// Filter pawns to only those on the correct rank for en passant
	enPassantPawns := pawns & board.RankMask(enPassantRank)

	for enPassantPawns != 0 {
		fromSquare, newBitboard := enPassantPawns.PopLSB()
		enPassantPawns = newBitboard

		var bitboardColor board.BitboardColor
		if player == White {
			bitboardColor = board.BitboardWhite
		} else {
			bitboardColor = board.BitboardBlack
		}
		attacks := board.GetPawnAttacks(fromSquare, bitboardColor)
		if attacks.HasBit(targetSquare) {
			toFile, toRank := board.SquareToFileRank(targetSquare)
			move := board.Move{
				From:        board.Square{File: fromSquare % 8, Rank: fromSquare / 8},
				To:          board.Square{File: toFile, Rank: toRank},
				Piece:       pawnPiece,
				Captured:    capturedPiece,
				Promotion:   board.Empty,
				IsCapture:   true,
				IsCastling:  false,
				IsEnPassant: true,
			}
			moveList.AddMove(move)
		}
	}
}

// addPromotionMoves adds all four promotion moves for a pawn reaching the end rank
func (bmg *BitboardMoveGenerator) addPromotionMoves(moveList *MoveList, fromSquare, toSquare int, pawnPiece, capturedPiece board.Piece, player Player) {
	toFile, toRank := board.SquareToFileRank(toSquare)
	fromFile, fromRank := board.SquareToFileRank(fromSquare)
	isCapture := capturedPiece != board.Empty

	var promotionPieces []board.Piece
	if player == White {
		promotionPieces = []board.Piece{board.WhiteQueen, board.WhiteRook, board.WhiteBishop, board.WhiteKnight}
	} else {
		promotionPieces = []board.Piece{board.BlackQueen, board.BlackRook, board.BlackBishop, board.BlackKnight}
	}

	for _, promotion := range promotionPieces {
		move := board.Move{
			From:        board.Square{File: fromFile, Rank: fromRank},
			To:          board.Square{File: toFile, Rank: toRank},
			Piece:       pawnPiece,
			Captured:    capturedPiece,
			Promotion:   promotion,
			IsCapture:   isCapture,
			IsCastling:  false,
			IsEnPassant: false,
		}
		moveList.AddMove(move)
	}
}

// GenerateKnightMovesBitboard generates knight moves using precomputed attack patterns (exported for testing)
func (bmg *BitboardMoveGenerator) GenerateKnightMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var knightPiece board.Piece
	var bitboardColor board.BitboardColor

	if player == White {
		knightPiece = board.WhiteKnight
		bitboardColor = board.BitboardWhite
	} else {
		knightPiece = board.BlackKnight
		bitboardColor = board.BitboardBlack
	}

	knights := b.GetPieceBitboard(knightPiece)
	if knights == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)

	// Iterate through each knight
	for knights != 0 {
		fromSquare, newBitboard := knights.PopLSB()
		knights = newBitboard

		// Get knight attacks using precomputed table
		attacks := board.GetKnightAttacks(fromSquare)
		// Remove squares occupied by friendly pieces
		validMoves := attacks &^ friendlyPieces

		// Generate moves
		for validMoves != 0 {
			toSquare, newValidMoves := validMoves.PopLSB()
			validMoves = newValidMoves

			capturedPiece := b.GetPieceOnSquare(toSquare)
			isCapture := capturedPiece != board.Empty

			toFile, toRank := board.SquareToFileRank(toSquare)
			fromFile, fromRank := board.SquareToFileRank(fromSquare)

			move := board.Move{
				From:        board.Square{File: fromFile, Rank: fromRank},
				To:          board.Square{File: toFile, Rank: toRank},
				Piece:       knightPiece,
				Captured:    capturedPiece,
				Promotion:   board.Empty,
				IsCapture:   isCapture,
				IsCastling:  false,
				IsEnPassant: false,
			}
			moveList.AddMove(move)
		}
	}
}

// generateBishopMovesBitboard generates bishop moves using magic bitboards
func (bmg *BitboardMoveGenerator) generateBishopMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var bishopPiece board.Piece
	var bitboardColor board.BitboardColor

	if player == White {
		bishopPiece = board.WhiteBishop
		bitboardColor = board.BitboardWhite
	} else {
		bishopPiece = board.BlackBishop
		bitboardColor = board.BitboardBlack
	}

	bishops := b.GetPieceBitboard(bishopPiece)
	if bishops == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)
	occupancy := b.AllPieces

	// Iterate through each bishop
	for bishops != 0 {
		fromSquare, newBitboard := bishops.PopLSB()
		bishops = newBitboard

		// Get bishop attacks using magic bitboards
		attacks := board.GetBishopAttacks(fromSquare, occupancy)
		// Remove squares occupied by friendly pieces
		validMoves := attacks &^ friendlyPieces

		bmg.addSlidingPieceMoves(b, moveList, fromSquare, validMoves, bishopPiece)
	}
}

// generateRookMovesBitboard generates rook moves using magic bitboards
func (bmg *BitboardMoveGenerator) generateRookMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var rookPiece board.Piece
	var bitboardColor board.BitboardColor

	if player == White {
		rookPiece = board.WhiteRook
		bitboardColor = board.BitboardWhite
	} else {
		rookPiece = board.BlackRook
		bitboardColor = board.BitboardBlack
	}

	rooks := b.GetPieceBitboard(rookPiece)
	if rooks == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)
	occupancy := b.AllPieces

	// Iterate through each rook
	for rooks != 0 {
		fromSquare, newBitboard := rooks.PopLSB()
		rooks = newBitboard

		// Get rook attacks using magic bitboards
		attacks := board.GetRookAttacks(fromSquare, occupancy)
		// Remove squares occupied by friendly pieces
		validMoves := attacks &^ friendlyPieces

		bmg.addSlidingPieceMoves(b, moveList, fromSquare, validMoves, rookPiece)
	}
}

// generateQueenMovesBitboard generates queen moves using optimized separate rook/bishop processing
func (bmg *BitboardMoveGenerator) generateQueenMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var queenPiece board.Piece
	var bitboardColor board.BitboardColor

	if player == White {
		queenPiece = board.WhiteQueen
		bitboardColor = board.BitboardWhite
	} else {
		queenPiece = board.BlackQueen
		bitboardColor = board.BitboardBlack
	}

	queens := b.GetPieceBitboard(queenPiece)
	if queens == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)
	occupancy := b.AllPieces

	// Process each queen with optimized separate rook/bishop move generation
	for queens != 0 {
		fromSquare, newBitboard := queens.PopLSB()
		queens = newBitboard

		// Generate rook-like moves for this queen
		rookAttacks := board.GetRookAttacks(fromSquare, occupancy)
		rookValidMoves := rookAttacks &^ friendlyPieces
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, rookValidMoves, queenPiece)

		// Generate bishop-like moves for this queen
		bishopAttacks := board.GetBishopAttacks(fromSquare, occupancy)
		bishopValidMoves := bishopAttacks &^ friendlyPieces
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, bishopValidMoves, queenPiece)
	}
}

// addSlidingPieceMoves is a helper function to add moves for sliding pieces
func (bmg *BitboardMoveGenerator) addSlidingPieceMoves(b *board.Board, moveList *MoveList, fromSquare int, validMoves board.Bitboard, piece board.Piece) {
	fromFile, fromRank := board.SquareToFileRank(fromSquare)

	for validMoves != 0 {
		toSquare, newValidMoves := validMoves.PopLSB()
		validMoves = newValidMoves

		capturedPiece := b.GetPieceOnSquare(toSquare)
		isCapture := capturedPiece != board.Empty

		toFile, toRank := board.SquareToFileRank(toSquare)

		move := board.Move{
			From:        board.Square{File: fromFile, Rank: fromRank},
			To:          board.Square{File: toFile, Rank: toRank},
			Piece:       piece,
			Captured:    capturedPiece,
			Promotion:   board.Empty,
			IsCapture:   isCapture,
			IsCastling:  false,
			IsEnPassant: false,
		}
		moveList.AddMove(move)
	}
}

// generateKingMovesBitboard generates king moves including castling
func (bmg *BitboardMoveGenerator) generateKingMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var kingPiece board.Piece
	var bitboardColor board.BitboardColor

	if player == White {
		kingPiece = board.WhiteKing
		bitboardColor = board.BitboardWhite
	} else {
		kingPiece = board.BlackKing
		bitboardColor = board.BitboardBlack
	}

	kings := b.GetPieceBitboard(kingPiece)
	if kings == 0 {
		return
	}

	friendlyPieces := b.GetColorBitboard(bitboardColor)

	// Should only be one king
	kingSquare := kings.LSB()
	if kingSquare == -1 {
		return
	}

	// Get king attacks using precomputed table
	attacks := board.GetKingAttacks(kingSquare)
	// Remove squares occupied by friendly pieces
	validMoves := attacks &^ friendlyPieces

	fromFile, fromRank := board.SquareToFileRank(kingSquare)

	// Generate regular king moves
	for validMoves != 0 {
		toSquare, newValidMoves := validMoves.PopLSB()
		validMoves = newValidMoves

		capturedPiece := b.GetPieceOnSquare(toSquare)
		isCapture := capturedPiece != board.Empty

		toFile, toRank := board.SquareToFileRank(toSquare)

		move := board.Move{
			From:        board.Square{File: fromFile, Rank: fromRank},
			To:          board.Square{File: toFile, Rank: toRank},
			Piece:       kingPiece,
			Captured:    capturedPiece,
			Promotion:   board.Empty,
			IsCapture:   isCapture,
			IsCastling:  false,
			IsEnPassant: false,
		}
		moveList.AddMove(move)
	}

	// Generate castling moves
	bmg.generateCastlingMovesBitboard(b, player, kingSquare, moveList)
}

// generateCastlingMovesBitboard generates castling moves using bitboard path checking
func (bmg *BitboardMoveGenerator) generateCastlingMovesBitboard(b *board.Board, player Player, kingSquare int, moveList *MoveList) {
	castlingRights := b.GetCastlingRights()
	if castlingRights == "-" {
		return
	}

	var kingPiece board.Piece
	var kingside, queenside byte
	var kingStartSquare int
	var kingsideKingTarget, kingsideRookTarget int
	var queensideKingTarget, queensideRookTarget int

	if player == White {
		kingPiece = board.WhiteKing
		kingside = 'K'
		queenside = 'Q'
		kingStartSquare = E1
		kingsideKingTarget = G1
		kingsideRookTarget = F1
		queensideKingTarget = C1
		queensideRookTarget = D1
	} else {
		kingPiece = board.BlackKing
		kingside = 'k'
		queenside = 'q'
		kingStartSquare = E8
		kingsideKingTarget = G8
		kingsideRookTarget = F8
		queensideKingTarget = C8
		queensideRookTarget = D8
	}

	// Check if king is on starting square
	if kingSquare != kingStartSquare {
		return
	}

	var oppositeColor board.BitboardColor
	if player == White {
		oppositeColor = board.BitboardBlack
	} else {
		oppositeColor = board.BitboardWhite
	}

	// Kingside castling
	if strings.ContainsRune(castlingRights, rune(kingside)) {
		// Check if path is clear
		pathMask := board.Bitboard(0)
		pathMask = pathMask.SetBit(kingsideRookTarget).SetBit(kingsideKingTarget)

		if (b.AllPieces & pathMask) == 0 {
			// Check if king or path squares are under attack
			if !b.IsSquareAttackedByColor(kingSquare, oppositeColor) &&
				!b.IsSquareAttackedByColor(kingsideRookTarget, oppositeColor) &&
				!b.IsSquareAttackedByColor(kingsideKingTarget, oppositeColor) {

				kingFromFile, kingFromRank := board.SquareToFileRank(kingSquare)
				kingToFile, kingToRank := board.SquareToFileRank(kingsideKingTarget)

				move := board.Move{
					From:        board.Square{File: kingFromFile, Rank: kingFromRank},
					To:          board.Square{File: kingToFile, Rank: kingToRank},
					Piece:       kingPiece,
					Captured:    board.Empty,
					Promotion:   board.Empty,
					IsCapture:   false,
					IsCastling:  true,
					IsEnPassant: false,
				}
				moveList.AddMove(move)
			}
		}
	}

	// Queenside castling
	if strings.ContainsRune(castlingRights, rune(queenside)) {
		// Check if path is clear (queenside includes an extra square)
		pathMask := board.Bitboard(0)
		pathMask = pathMask.SetBit(queensideRookTarget).SetBit(queensideKingTarget).SetBit(B1)
		if player == Black {
			pathMask = pathMask.ClearBit(B1).SetBit(B8)
		}

		if (b.AllPieces & pathMask) == 0 {
			// Check if king or key path squares are under attack
			if !b.IsSquareAttackedByColor(kingSquare, oppositeColor) &&
				!b.IsSquareAttackedByColor(queensideRookTarget, oppositeColor) &&
				!b.IsSquareAttackedByColor(queensideKingTarget, oppositeColor) {

				kingFromFile, kingFromRank := board.SquareToFileRank(kingSquare)
				kingToFile, kingToRank := board.SquareToFileRank(queensideKingTarget)

				move := board.Move{
					From:        board.Square{File: kingFromFile, Rank: kingFromRank},
					To:          board.Square{File: kingToFile, Rank: kingToRank},
					Piece:       kingPiece,
					Captured:    board.Empty,
					Promotion:   board.Empty,
					IsCapture:   false,
					IsCastling:  true,
					IsEnPassant: false,
				}
				moveList.AddMove(move)
			}
		}
	}
}

// filterLegalMovesInPlace filters out moves that would leave the king in check
// Optimized implementation using pin analysis instead of make/unmake
func (bmg *BitboardMoveGenerator) filterLegalMovesInPlace(b *board.Board, player Player, moves *MoveList) {
	// Idiomatic chess engine approach: direct calculation when needed

	// Find our king position
	var ourKingPiece board.Piece
	var opponentColor board.BitboardColor
	if player == White {
		ourKingPiece = board.WhiteKing
		opponentColor = board.BitboardBlack
	} else {
		ourKingPiece = board.BlackKing
		opponentColor = board.BitboardWhite
	}

	ourKingBitboard := b.GetPieceBitboard(ourKingPiece)
	if ourKingBitboard == 0 {
		// No king - return empty move list
		moves.Count = 0
		return
	}
	kingSquare := ourKingBitboard.LSB()

	// Calculate pin information once (lightweight calculation)
	pinnedPieces := bmg.calculatePinnedPieces(b, kingSquare, opponentColor)

	// Check if we're currently in check
	inCheck := b.IsSquareAttackedByColor(kingSquare, opponentColor)

	// Filter moves in place using direct validation
	writeIndex := 0
	for readIndex := 0; readIndex < moves.Count; readIndex++ {
		move := moves.Moves[readIndex]

		if bmg.isMoveLegal(b, move, kingSquare, pinnedPieces, inCheck, opponentColor) {
			moves.Moves[writeIndex] = move
			writeIndex++
		}
	}

	// Update count and truncate slice
	moves.Count = writeIndex
	moves.Moves = moves.Moves[:writeIndex]
}

// calculatePinnedPieces finds all pieces pinned to the king by opponent sliding pieces
func (bmg *BitboardMoveGenerator) calculatePinnedPieces(b *board.Board, kingSquare int, opponentColor board.BitboardColor) board.Bitboard {
	var pinnedPieces board.Bitboard

	// Get opponent's sliding pieces (rooks, bishops, queens)
	var opponentRooks, opponentBishops, opponentQueens board.Bitboard
	if opponentColor == board.BitboardWhite {
		opponentRooks = b.GetPieceBitboard(board.WhiteRook)
		opponentBishops = b.GetPieceBitboard(board.WhiteBishop)
		opponentQueens = b.GetPieceBitboard(board.WhiteQueen)
	} else {
		opponentRooks = b.GetPieceBitboard(board.BlackRook)
		opponentBishops = b.GetPieceBitboard(board.BlackBishop)
		opponentQueens = b.GetPieceBitboard(board.BlackQueen)
	}

	// Check for pins by rooks/queens (rank and file attacks)
	rookAttackers := opponentRooks | opponentQueens
	for rookAttackers != 0 {
		attackerSquare, newBitboard := rookAttackers.PopLSB()
		rookAttackers = newBitboard

		// Check if attacker and king are on the same rank or file (rooks can't pin diagonally)
		if bmg.areOnSameRankOrFile(attackerSquare, kingSquare) {
			// They're on the same rank/file - check squares between them
			between := board.GetBetween(attackerSquare, kingSquare)
			blockers := between & b.AllPieces
			if blockers.PopCount() == 1 {
				// Exactly one piece between attacker and king - it's pinned
				pinnedPieces |= blockers
			}
		}
	}

	// Check for pins by bishops/queens (diagonal attacks)
	bishopAttackers := opponentBishops | opponentQueens
	for bishopAttackers != 0 {
		attackerSquare, newBitboard := bishopAttackers.PopLSB()
		bishopAttackers = newBitboard

		// Check if attacker and king are on the same diagonal line (potential pin)
		// Only consider diagonal lines for bishop pins (not rank/file lines)
		if bmg.areOnSameDiagonal(attackerSquare, kingSquare) {
			// They're on the same diagonal - check squares between them
			between := board.GetBetween(attackerSquare, kingSquare)
			blockers := between & b.AllPieces
			if blockers.PopCount() == 1 {
				// Exactly one piece between attacker and king - it's pinned
				pinnedPieces |= blockers
			}
		}
	}

	return pinnedPieces
}

// enPassantCapturesAttacker checks if an en passant move captures the attacking piece
func (bmg *BitboardMoveGenerator) enPassantCapturesAttacker(move board.Move, attackerSquare int) bool {
	toSquare := move.To.Rank*8 + move.To.File

	// Calculate where the captured pawn is located
	var capturedPawnSquare int
	if move.Piece == board.WhitePawn {
		capturedPawnSquare = toSquare - 8 // Black pawn is one rank below en passant destination
	} else {
		capturedPawnSquare = toSquare + 8 // White pawn is one rank above en passant destination
	}

	return capturedPawnSquare == attackerSquare
}

// areOnSameRankOrFile checks if two squares are on the same rank or file
func (bmg *BitboardMoveGenerator) areOnSameRankOrFile(square1, square2 int) bool {
	file1 := square1 % 8
	rank1 := square1 / 8
	file2 := square2 % 8
	rank2 := square2 / 8

	return file1 == file2 || rank1 == rank2
}

// areOnSameDiagonal checks if two squares are on the same diagonal
func (bmg *BitboardMoveGenerator) areOnSameDiagonal(square1, square2 int) bool {
	file1 := square1 % 8
	rank1 := square1 / 8
	file2 := square2 % 8
	rank2 := square2 / 8

	// Check if the absolute difference in files equals the absolute difference in ranks
	fileDiff := file1 - file2
	rankDiff := rank1 - rank2

	if fileDiff < 0 {
		fileDiff = -fileDiff
	}
	if rankDiff < 0 {
		rankDiff = -rankDiff
	}

	return fileDiff == rankDiff && fileDiff != 0
}

// isMoveLegal checks if a move is legal without making/unmaking it
func (bmg *BitboardMoveGenerator) isMoveLegal(b *board.Board, move board.Move, kingSquare int, pinnedPieces board.Bitboard, inCheck bool, opponentColor board.BitboardColor) bool {
	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	// King moves - check if destination square is attacked after king moves
	if move.Piece == board.WhiteKing || move.Piece == board.BlackKing {
		// For king moves, we need to check if the destination would be attacked
		// with the king on the new square (not the old one)
		return bmg.isKingMoveIntoSafety(b, move, kingSquare, toSquare, opponentColor)
	}

	// If we're in double check, only king moves are legal (already handled above)
	if inCheck {
		// Count attackers to determine if it's double check
		attackers := bmg.getAttackersToSquare(b, kingSquare, opponentColor)
		if attackers.PopCount() > 1 {
			return false // Only king moves allowed in double check
		}

		// Single check - move must block check or capture attacking piece
		attackerSquare := attackers.LSB()

		// Capturing the attacker (including en passant captures)
		if toSquare == attackerSquare || (move.IsEnPassant && bmg.enPassantCapturesAttacker(move, attackerSquare)) {
			// Still need to check if moving piece is pinned
			if pinnedPieces.HasBit(fromSquare) {
				pinRay := board.GetLine(kingSquare, bmg.findPinningPiece(b, fromSquare, kingSquare, opponentColor))
				return pinRay.HasBit(toSquare)
			}
			return true
		}

		// Blocking the check (only works against sliding pieces)
		between := board.GetBetween(attackerSquare, kingSquare)
		if between.HasBit(toSquare) {
			// Still need to check if moving piece is pinned
			if pinnedPieces.HasBit(fromSquare) {
				pinRay := board.GetLine(kingSquare, bmg.findPinningPiece(b, fromSquare, kingSquare, opponentColor))
				return pinRay.HasBit(toSquare)
			}
			return true
		}

		return false // Move doesn't address the check
	}

	// Handle pinned pieces - they can only move along the pin ray
	if pinnedPieces.HasBit(fromSquare) {
		pinningPieceSquare := bmg.findPinningPiece(b, fromSquare, kingSquare, opponentColor)
		if pinningPieceSquare == -1 {
			return false
		}

		pinRay := board.GetLine(kingSquare, pinningPieceSquare)
		return pinRay.HasBit(toSquare)
	}

	// Special case: En passant can expose king to rank attacks
	if move.IsEnPassant {
		return bmg.isEnPassantLegal(b, move, kingSquare, opponentColor)
	}

	// Regular move - legal if piece isn't pinned
	return true
}

// getAttackersToSquare returns a bitboard of pieces attacking the given square
func (bmg *BitboardMoveGenerator) getAttackersToSquare(b *board.Board, square int, attackerColor board.BitboardColor) board.Bitboard {
	var attackers board.Bitboard

	if attackerColor == board.BitboardWhite {
		// Check white pawn attacks
		whitePawnAttacks := board.GetPawnAttacks(square, board.BitboardBlack) // Reverse direction
		attackers |= whitePawnAttacks & b.GetPieceBitboard(board.WhitePawn)

		// Check knight attacks
		knightAttacks := board.GetKnightAttacks(square)
		attackers |= knightAttacks & b.GetPieceBitboard(board.WhiteKnight)

		// Check sliding piece attacks
		rookAttacks := board.GetRookAttacks(square, b.AllPieces)
		attackers |= rookAttacks & (b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteQueen))

		bishopAttacks := board.GetBishopAttacks(square, b.AllPieces)
		attackers |= bishopAttacks & (b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.WhiteQueen))

		// Check king attacks
		kingAttacks := board.GetKingAttacks(square)
		attackers |= kingAttacks & b.GetPieceBitboard(board.WhiteKing)
	} else {
		// Check black pawn attacks
		blackPawnAttacks := board.GetPawnAttacks(square, board.BitboardWhite) // Reverse direction
		attackers |= blackPawnAttacks & b.GetPieceBitboard(board.BlackPawn)

		// Check knight attacks
		knightAttacks := board.GetKnightAttacks(square)
		attackers |= knightAttacks & b.GetPieceBitboard(board.BlackKnight)

		// Check sliding piece attacks
		rookAttacks := board.GetRookAttacks(square, b.AllPieces)
		attackers |= rookAttacks & (b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackQueen))

		bishopAttacks := board.GetBishopAttacks(square, b.AllPieces)
		attackers |= bishopAttacks & (b.GetPieceBitboard(board.BlackBishop) | b.GetPieceBitboard(board.BlackQueen))

		// Check king attacks
		kingAttacks := board.GetKingAttacks(square)
		attackers |= kingAttacks & b.GetPieceBitboard(board.BlackKing)
	}

	return attackers
}

// findPinningPiece finds which piece is pinning the given piece to the king
func (bmg *BitboardMoveGenerator) findPinningPiece(b *board.Board, pinnedSquare, kingSquare int, opponentColor board.BitboardColor) int {
	// Get all opponent sliding pieces that could potentially pin
	var opponentSliders board.Bitboard
	if opponentColor == board.BitboardWhite {
		opponentSliders = b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.WhiteQueen)
	} else {
		opponentSliders = b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackBishop) | b.GetPieceBitboard(board.BlackQueen)
	}

	// Check each opponent sliding piece to see if it pins the given square to the king
	for opponentSliders != 0 {
		attackerSquare, newBitboard := opponentSliders.PopLSB()
		opponentSliders = newBitboard

		// Check if attacker and king are on the same line (potential pin)
		line := board.GetLine(attackerSquare, kingSquare)
		if line != 0 {
			// They're on the same line - check squares between them
			between := board.GetBetween(attackerSquare, kingSquare)
			blockers := between & b.AllPieces

			// If exactly one piece between attacker and king, and it's our pinned piece
			if blockers.PopCount() == 1 && blockers.HasBit(pinnedSquare) {
				return attackerSquare
			}
		}
	}

	return -1
}

// isEnPassantLegal checks if an en passant move is legal (doesn't expose king to rank attacks)
func (bmg *BitboardMoveGenerator) isEnPassantLegal(b *board.Board, move board.Move, kingSquare int, opponentColor board.BitboardColor) bool {
	// En passant captures remove a pawn from a different square than the destination
	// This can potentially expose the king to rank attacks

	fromSquare := move.From.Rank*8 + move.From.File
	toSquare := move.To.Rank*8 + move.To.File

	// The captured pawn square
	var capturedPawnSquare int
	if move.Piece == board.WhitePawn {
		capturedPawnSquare = toSquare - 8 // Black pawn is one rank below
	} else {
		capturedPawnSquare = toSquare + 8 // White pawn is one rank above
	}

	// Check if king and moving pawn are on the same rank
	kingRank := kingSquare / 8
	moveRank := fromSquare / 8
	if kingRank != moveRank {
		return true // No rank attack possible
	}

	// Check for opponent rooks/queens on the same rank
	var opponentRooksQueens board.Bitboard
	if opponentColor == board.BitboardWhite {
		opponentRooksQueens = b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteQueen)
	} else {
		opponentRooksQueens = b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackQueen)
	}

	rankMask := board.RankMask(kingRank)
	rankAttackers := opponentRooksQueens & rankMask

	if rankAttackers == 0 {
		return true // No rank attackers
	}

	// Simulate the en passant capture by temporarily removing both pawns
	occupancyAfterMove := b.AllPieces
	occupancyAfterMove = occupancyAfterMove.ClearBit(fromSquare)         // Remove moving pawn
	occupancyAfterMove = occupancyAfterMove.ClearBit(capturedPawnSquare) // Remove captured pawn
	occupancyAfterMove = occupancyAfterMove.SetBit(toSquare)             // Add pawn at destination

	// Check if any rank attacker can now attack the king
	for rankAttackers != 0 {
		attackerSquare, newBitboard := rankAttackers.PopLSB()
		rankAttackers = newBitboard

		attackRay := board.GetRookAttacks(attackerSquare, occupancyAfterMove)
		if attackRay.HasBit(kingSquare) {
			return false // En passant would expose king
		}
	}

	return true
}

// isKingMoveIntoSafety checks if a king move places the king in safety
func (bmg *BitboardMoveGenerator) isKingMoveIntoSafety(b *board.Board, _ board.Move, fromSquare, toSquare int, opponentColor board.BitboardColor) bool {
	// For king moves, we need to temporarily remove the king from its current square
	// and check if the destination square would be attacked

	// Create a modified occupancy bitboard with king removed from original square
	modifiedOccupancy := b.AllPieces
	modifiedOccupancy = modifiedOccupancy.ClearBit(fromSquare)

	// Check attacks to the destination square with the modified occupancy
	// This simulates the king no longer being on the original square
	return !bmg.isSquareAttackedByColorWithOccupancy(b, toSquare, opponentColor, modifiedOccupancy)
}

// isSquareAttackedByColorWithOccupancy checks square attacks with custom occupancy
func (bmg *BitboardMoveGenerator) isSquareAttackedByColorWithOccupancy(b *board.Board, square int, attackerColor board.BitboardColor, occupancy board.Bitboard) bool {
	if square < 0 || square > 63 {
		return false
	}

	if attackerColor == board.BitboardWhite {
		// Check white pawn attacks
		whitePawnAttacks := board.GetPawnAttacks(square, board.BitboardBlack) // Reverse direction
		if (whitePawnAttacks & b.GetPieceBitboard(board.WhitePawn)) != 0 {
			return true
		}

		// Check knight attacks
		knightAttacks := board.GetKnightAttacks(square)
		if (knightAttacks & b.GetPieceBitboard(board.WhiteKnight)) != 0 {
			return true
		}

		// Check sliding piece attacks with custom occupancy
		rookAttacks := board.GetRookAttacks(square, occupancy)
		if (rookAttacks & (b.GetPieceBitboard(board.WhiteRook) | b.GetPieceBitboard(board.WhiteQueen))) != 0 {
			return true
		}

		bishopAttacks := board.GetBishopAttacks(square, occupancy)
		if (bishopAttacks & (b.GetPieceBitboard(board.WhiteBishop) | b.GetPieceBitboard(board.WhiteQueen))) != 0 {
			return true
		}

		// Check king attacks (but not the moving king)
		kingAttacks := board.GetKingAttacks(square)
		if (kingAttacks & b.GetPieceBitboard(board.WhiteKing)) != 0 {
			return true
		}
	} else {
		// Check black pawn attacks
		blackPawnAttacks := board.GetPawnAttacks(square, board.BitboardWhite) // Reverse direction
		if (blackPawnAttacks & b.GetPieceBitboard(board.BlackPawn)) != 0 {
			return true
		}

		// Check knight attacks
		knightAttacks := board.GetKnightAttacks(square)
		if (knightAttacks & b.GetPieceBitboard(board.BlackKnight)) != 0 {
			return true
		}

		// Check sliding piece attacks with custom occupancy
		rookAttacks := board.GetRookAttacks(square, occupancy)
		if (rookAttacks & (b.GetPieceBitboard(board.BlackRook) | b.GetPieceBitboard(board.BlackQueen))) != 0 {
			return true
		}

		bishopAttacks := board.GetBishopAttacks(square, occupancy)
		if (bishopAttacks & (b.GetPieceBitboard(board.BlackBishop) | b.GetPieceBitboard(board.BlackQueen))) != 0 {
			return true
		}

		// Check king attacks (but not the moving king)
		kingAttacks := board.GetKingAttacks(square)
		if (kingAttacks & b.GetPieceBitboard(board.BlackKing)) != 0 {
			return true
		}
	}

	return false
}
