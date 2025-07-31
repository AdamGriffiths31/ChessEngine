package moves

import (
	"strings"
	"github.com/AdamGriffiths31/ChessEngine/board"
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
	// Reuse move list from pool for memory efficiency
	tempMoveList *MoveList
}

// NewBitboardMoveGenerator creates a new bitboard-based move generator
func NewBitboardMoveGenerator() *BitboardMoveGenerator {
	return &BitboardMoveGenerator{
		tempMoveList: GetMoveList(),
	}
}

// GenerateAllMovesBitboard generates all legal moves using bitboard operations
// This is the main entry point that coordinates all piece-specific generators
func (bmg *BitboardMoveGenerator) GenerateAllMovesBitboard(b *board.Board, player Player) *MoveList {
	// Initialize tables if not already done
	board.InitializeTables()
	board.InitializeMagicBitboards()
	
	moveList := GetMoveList()
	
	// Generate moves for each piece type using bitboard operations
	bmg.generatePawnMovesBitboard(b, player, moveList)
	bmg.GenerateKnightMovesBitboard(b, player, moveList)
	bmg.generateBishopMovesBitboard(b, player, moveList)
	bmg.generateRookMovesBitboard(b, player, moveList)
	bmg.generateQueenMovesBitboard(b, player, moveList)
	bmg.generateKingMovesBitboard(b, player, moveList)
	
	// Filter out illegal moves (moves that leave king in check)
	filteredMoves := bmg.filterLegalMoves(b, player, moveList)
	ReleaseMoveList(moveList)
	
	return filteredMoves
}

// generatePawnMovesBitboard generates pawn moves using bitboard shifts and attack patterns
func (bmg *BitboardMoveGenerator) generatePawnMovesBitboard(b *board.Board, player Player, moveList *MoveList) {
	var pawnPiece board.Piece
	var bitboardColor board.BitboardColor
	var startRank, promotionRank int
	
	if player == White {
		pawnPiece = board.WhitePawn
		bitboardColor = board.BitboardWhite
		startRank = 1  // 2nd rank (rank index 1)
		promotionRank = 7  // 8th rank (rank index 7) - where pawn arrives to promote
	} else {
		pawnPiece = board.BlackPawn
		bitboardColor = board.BitboardBlack
		startRank = 6  // 7th rank (rank index 6)
		promotionRank = 0  // 1st rank (rank index 0) - where pawn arrives to promote
	}
	
	pawns := b.GetPieceBitboard(pawnPiece)
	if pawns == 0 {
		return // No pawns to move
	}
	
	emptySquares := ^b.AllPieces  // Invert to get empty squares
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
			fromSquare = toSquare - 8  // Move came from south
		} else {
			fromSquare = toSquare + 8  // Move came from north
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
			fromSquare = toSquare - 16  // Move came from 2 squares south
		} else {
			fromSquare = toSquare + 16  // Move came from 2 squares north
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

// generateQueenMovesBitboard generates queen moves using magic bitboards
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
	
	// Iterate through each queen
	for queens != 0 {
		fromSquare, newBitboard := queens.PopLSB()
		queens = newBitboard
		
		// Get queen attacks using magic bitboards
		attacks := board.GetQueenAttacks(fromSquare, occupancy)
		// Remove squares occupied by friendly pieces
		validMoves := attacks &^ friendlyPieces
		
		bmg.addSlidingPieceMoves(b, moveList, fromSquare, validMoves, queenPiece)
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

// filterLegalMoves filters out moves that would leave the king in check
func (bmg *BitboardMoveGenerator) filterLegalMoves(b *board.Board, player Player, moves *MoveList) *MoveList {
	legalMoves := GetMoveList()
	
	// Create a proper move executor for make/unmake
	moveExecutor := &MoveExecutor{}
	
	// Simple update function that doesn't require a generator instance
	updateBoardState := func(b *board.Board, move board.Move) {
		// Update castling rights based on the move
		castlingRights := b.GetCastlingRights()
		piece := b.GetPiece(move.To.Rank, move.To.File)
		
		// King moves remove all castling rights for that side
		if piece == board.WhiteKing {
			castlingRights = removeCastlingRights(castlingRights, "KQ")
		} else if piece == board.BlackKing {
			castlingRights = removeCastlingRights(castlingRights, "kq")
		}
		
		// Update other board state as needed...
		b.SetCastlingRights(castlingRights)
		// ... rest of state updates
	}
	
	for i := 0; i < moves.Count; i++ {
		move := moves.Moves[i]
		
		// Make the move
		history := moveExecutor.MakeMove(b, move, updateBoardState)
		
		// Check if our king is in check after this move
		var kingPiece board.Piece
		var oppositeColor board.BitboardColor
		if player == White {
			kingPiece = board.WhiteKing
			oppositeColor = board.BitboardBlack
		} else {
			kingPiece = board.BlackKing
			oppositeColor = board.BitboardWhite
		}
		
		kingBitboard := b.GetPieceBitboard(kingPiece)
		isLegal := true
		if kingBitboard != 0 {
			kingSquare := kingBitboard.LSB()
			if kingSquare != -1 && b.IsSquareAttackedByColor(kingSquare, oppositeColor) {
				isLegal = false
			}
		}
		
		// Unmake the move
		moveExecutor.UnmakeMove(b, history)
		
		if isLegal {
			legalMoves.AddMove(move)
		}
	}
	
	return legalMoves
}

// Helper function to remove castling rights
func removeCastlingRights(rights, toRemove string) string {
	result := ""
	for _, r := range rights {
		remove := false
		for _, remove_r := range toRemove {
			if r == remove_r {
				remove = true
				break
			}
		}
		if !remove {
			result += string(r)
		}
	}
	if result == "" {
		return "-"
	}
	return result
}

// Release releases the temporary move list back to the pool
func (bmg *BitboardMoveGenerator) Release() {
	if bmg.tempMoveList != nil {
		ReleaseMoveList(bmg.tempMoveList)
		bmg.tempMoveList = nil
	}
}