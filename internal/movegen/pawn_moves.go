package movegen

import (
	"github.com/AdamGriffiths31/ChessEngine/internal/board"
)

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
		return
	}

	emptySquares := ^b.AllPieces // Invert to get empty squares
	enemyPieces := b.GetColorBitboard(board.OppositeBitboardColor(bitboardColor))

	var singlePushes board.Bitboard
	if player == White {
		singlePushes = pawns.ShiftNorth() & emptySquares
	} else {
		singlePushes = pawns.ShiftSouth() & emptySquares
	}

	for singlePushes != 0 {
		toSquare, newBitboard := singlePushes.PopLSB()
		singlePushes = newBitboard

		var fromSquare int
		if player == White {
			fromSquare = toSquare - 8 // Move came from south
		} else {
			fromSquare = toSquare + 8 // Move came from north
		}

		toFile, toRank := board.SquareToFileRank(toSquare)
		if toRank == promotionRank {
			bmg.addPromotionMoves(moveList, fromSquare, toSquare, pawnPiece, board.Empty, player)
		} else {
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

	// Important: both the single push square AND double push square must be empty
	var doublePushes board.Bitboard
	if player == White {
		startingPawns := pawns & board.RankMask(startRank)
		singlePushClear := startingPawns.ShiftNorth() & emptySquares
		doublePushes = singlePushClear.ShiftNorth() & emptySquares
	} else {
		startingPawns := pawns & board.RankMask(startRank)
		singlePushClear := startingPawns.ShiftSouth() & emptySquares
		doublePushes = singlePushClear.ShiftSouth() & emptySquares
	}

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

	bmg.generatePawnCapturesBitboard(b, player, pawns, enemyPieces, moveList, promotionRank)
	bmg.generateEnPassantCapturesBitboard(b, player, pawns, moveList)
}

func (bmg *BitboardMoveGenerator) generatePawnCapturesBitboard(b *board.Board, player Player, pawns board.Bitboard, enemyPieces board.Bitboard, moveList *MoveList, promotionRank int) {
	var pawnPiece board.Piece

	if player == White {
		pawnPiece = board.WhitePawn
	} else {
		pawnPiece = board.BlackPawn
	}

	tempPawns := pawns
	for tempPawns != 0 {
		fromSquare, newBitboard := tempPawns.PopLSB()
		tempPawns = newBitboard

		var bitboardColor board.BitboardColor
		if player == White {
			bitboardColor = board.BitboardWhite
		} else {
			bitboardColor = board.BitboardBlack
		}
		attacks := board.GetPawnAttacks(fromSquare, bitboardColor)
		captures := attacks & enemyPieces

		for captures != 0 {
			toSquare, newCaptures := captures.PopLSB()
			captures = newCaptures

			capturedPiece := b.GetPieceOnSquare(toSquare)
			toFile, toRank := board.SquareToFileRank(toSquare)

			if toRank == promotionRank {
				bmg.addPromotionMoves(moveList, fromSquare, toSquare, pawnPiece, capturedPiece, player)
			} else {
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

func (bmg *BitboardMoveGenerator) generateEnPassantCapturesBitboard(b *board.Board, player Player, pawns board.Bitboard, moveList *MoveList) {
	enPassantTarget, hasEnPassant := b.GetEnPassantTarget()
	if !hasEnPassant {
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

	// En passant is only valid for pawns on the correct rank:
	// - White pawns must be on 5th rank (rank index 4)
	// - Black pawns must be on 4th rank (rank index 3)
	var enPassantRank int
	if player == White {
		enPassantRank = 4 // 5th rank
	} else {
		enPassantRank = 3 // 4th rank
	}

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
