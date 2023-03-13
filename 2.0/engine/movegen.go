package engine

import (
	"fmt"
	"math/bits"

	"github.com/AdamGriffiths31/ChessEngine/data"
	"github.com/AdamGriffiths31/ChessEngine/io"
	"github.com/AdamGriffiths31/ChessEngine/validate"
)

var preCalculatedKnightMoves [64]uint64
var preCalculatedKingMoves [64]uint64

type MoveList struct {
	Moves [300]Move
	Count int
}

type Move struct {
	Score int
	Move  int
}

func (p *Position) GenerateAllMoves(ml *MoveList) {
	ml.Count = 0
	if p.Side == data.White {
		p.generateWhitePawnMoves(ml)
		p.generateSliderMoves(ml, WR, true)
		p.generateSliderMoves(ml, WB, true)
		p.generateSliderMoves(ml, WQ, true)
		p.generateSliderMoves(ml, WK, true)
		p.generateSliderMoves(ml, WN, true)
		p.generateWhiteCastleMoves(ml)
	} else {
		p.generateBlackPawnMoves(ml)
		p.generateSliderMoves(ml, BR, true)
		p.generateSliderMoves(ml, BB, true)
		p.generateSliderMoves(ml, BQ, true)
		p.generateSliderMoves(ml, BK, true)
		p.generateSliderMoves(ml, BN, true)
		p.generateBlackCastleMoves(ml)
	}
}

func (p *Position) generateWhitePawnMoves(ml *MoveList) {
	p.generateWhitePawnQuietMoves(ml)
	p.generateWhitePawnEnPassantMoves(ml)
	p.generateWhitePawnCaptureMoves(ml)
}
func (p *Position) generateBlackPawnMoves(ml *MoveList) {
	p.generateBlackPawnQuietMoves(ml)
	p.generateBlackPawnEnPassantMoves(ml)
	p.generateBlackPawnCaptureMoves(ml)
}

func (p *Position) generateWhitePawnQuietMoves(ml *MoveList) {
	pawns := p.Board.WhitePawn
	for pawns != 0 {
		sq := bits.TrailingZeros64(pawns)
		pawns ^= (1 << sq)
		sq = data.Square64ToSquare120[sq]

		// Check for single pawn push
		to := sq + 10
		if p.Board.PieceAt(data.Square120ToSquare64[to]) == data.Empty {
			p.addWhitePawnMove(sq, to, ml)
			// Check for double pawn push
			if data.RanksBoard[sq] == data.Rank2 && p.Board.PieceAt(data.Square120ToSquare64[to+10]) == data.Empty {
				p.addQuiteMove(MakeMoveInt(sq, to+10, data.Empty, data.Empty, data.MFLAGPS), ml)
			}
		}
	}
}

func (p *Position) generateBlackPawnQuietMoves(ml *MoveList) {
	pawns := p.Board.BlackPawn
	for pawns != 0 {
		sq := bits.TrailingZeros64(pawns)
		pawns ^= (1 << sq)
		sq = data.Square64ToSquare120[sq]
		// Check for single pawn push
		to := sq - 10
		if p.Board.PieceAt(data.Square120ToSquare64[to]) == data.Empty {
			p.addBlackPawnMove(sq, to, ml)
			// Check for double pawn push
			if data.RanksBoard[sq] == data.Rank7 && p.Board.PieceAt(data.Square120ToSquare64[to-10]) == data.Empty {
				p.addQuiteMove(MakeMoveInt(sq, to-10, data.Empty, data.Empty, data.MFLAGPS), ml)
			}
		}
	}
}

func (p *Position) generateWhitePawnEnPassantMoves(ml *MoveList) {

	if p.EnPassant != data.NoSquare && p.EnPassant != data.Empty {
		pawns := p.Board.WhitePawn &^ (data.RankBBMask[data.Rank8] | data.RankBBMask[data.Rank7])
		leftAttacks := (pawns << 7) &^ data.FileBBMask[data.FileH] & data.SquareBB[data.Square120ToSquare64[p.EnPassant]]
		rightAttacks := (pawns << 9) &^ data.FileBBMask[data.FileA] & data.SquareBB[data.Square120ToSquare64[p.EnPassant]]
		if leftAttacks != 0 {
			sq := bits.TrailingZeros64(leftAttacks)
			sq = data.Square64ToSquare120[sq]
			p.addEnPasMove(MakeMoveInt(sq-9, p.EnPassant, data.Empty, data.Empty, data.MFLAGEP), ml)
		}

		if rightAttacks != 0 {
			sq := bits.TrailingZeros64(rightAttacks)
			sq = data.Square64ToSquare120[sq]
			p.addEnPasMove(MakeMoveInt(sq-11, p.EnPassant, data.Empty, data.Empty, data.MFLAGEP), ml)
		}
	}
}

func (p *Position) generateBlackPawnEnPassantMoves(moveList *MoveList) {
	if p.EnPassant != data.NoSquare && p.EnPassant != data.Empty {
		pawns := p.Board.BlackPawn &^ (data.RankBBMask[data.Rank1] | data.RankBBMask[data.Rank2])
		leftAttacks := (pawns >> 9) &^ data.FileBBMask[data.FileH] & data.SquareBB[data.Square120ToSquare64[p.EnPassant]]
		rightAttacks := (pawns >> 7) &^ data.FileBBMask[data.FileA] & data.SquareBB[data.Square120ToSquare64[p.EnPassant]]
		if leftAttacks != 0 {
			sq := bits.TrailingZeros64(leftAttacks)
			sq = data.Square64ToSquare120[sq]
			p.addEnPasMove(MakeMoveInt(sq+11, p.EnPassant, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}

		if rightAttacks != 0 {
			sq := bits.TrailingZeros64(rightAttacks)
			sq = data.Square64ToSquare120[sq]
			p.addEnPasMove(MakeMoveInt(sq+9, p.EnPassant, data.Empty, data.Empty, data.MFLAGEP), moveList)
		}
	}
}

func (p *Position) generateWhitePawnCaptureMoves(ml *MoveList) {
	// Generate attacks to the left
	whitePawns := p.Board.WhitePawn
	blackPieces := p.Board.BlackPieces

	leftAttacks := (whitePawns << 7) &^ data.FileBBMask[data.FileH] & blackPieces

	for leftAttacks != 0 {
		sq := bits.TrailingZeros64(leftAttacks)
		leftAttacks &= leftAttacks - 1
		sq = data.Square64ToSquare120[sq]
		p.addWhitePawnCaptureMove(ml, sq-9, sq, p.Board.PieceAt(data.Square120ToSquare64[sq]))
	}

	// Generate attacks to the right
	rightAttacks := (whitePawns << 9) &^ data.FileBBMask[data.FileA] & blackPieces
	for rightAttacks != 0 {
		sq := bits.TrailingZeros64(rightAttacks)
		rightAttacks &= rightAttacks - 1
		sq = data.Square64ToSquare120[sq]
		p.addWhitePawnCaptureMove(ml, sq-11, sq, p.Board.PieceAt(data.Square120ToSquare64[sq]))
	}
}

func (p *Position) generateBlackPawnCaptureMoves(moveList *MoveList) {

	// Generate attacks to the left
	leftAttacks := (p.Board.BlackPawn >> 9) &^ data.FileBBMask[data.FileH] & p.Board.WhitePieces
	for leftAttacks != 0 {
		sq := bits.TrailingZeros64(leftAttacks)
		leftAttacks &= leftAttacks - 1
		sq = data.Square64ToSquare120[sq]
		p.addBlackPawnCaptureMove(moveList, sq+11, sq, p.Board.PieceAt(data.Square120ToSquare64[sq]))
	}

	// Generate attacks to the right
	rightAttacks := (p.Board.BlackPawn >> 7) &^ data.FileBBMask[data.FileA] & p.Board.WhitePieces
	for rightAttacks != 0 {
		sq := bits.TrailingZeros64(rightAttacks)
		rightAttacks &= rightAttacks - 1
		sq = data.Square64ToSquare120[sq]
		p.addBlackPawnCaptureMove(moveList, sq+9, sq, p.Board.PieceAt(data.Square120ToSquare64[sq]))
	}
}

func (p *Position) addEnPasMove(move int, moveList *MoveList) {
	moveList.Moves[moveList.Count].Move = move
	moveList.Moves[moveList.Count].Score = 105 + 1000000
	moveList.Count++
}

func (p *Position) addCaptureMove(move int, moveList *MoveList) {
	piece := p.Board.PieceAt(data.Square120ToSquare64[data.FromSquare(move)])
	moveList.Moves[moveList.Count].Move = move
	moveList.Moves[moveList.Count].Score = data.MvvLvaScores[data.Captured(move)][piece] + 1000000
	moveList.Count++
}

func (p *Position) addQuiteMove(move int, moveList *MoveList) {
	//moveList.Moves[moveList.Count].Move = move
	// piece := p.Board.PieceAt(data.Square120ToSquare64[data.FromSquare(move)])
	// switch move {
	// case pos.SearchKillers[0][pos.Play]:
	// 	moveList.Moves[moveList.Count].Score = 900000
	// case pos.SearchKillers[1][pos.Play]:
	// 	moveList.Moves[moveList.Count].Score = 800000
	// default:
	// 	moveList.Moves[moveList.Count].Score = pos.SearchHistory[piece][data.ToSquare(move)]
	// }
	moveList.Moves[moveList.Count].Move = move
	moveList.Moves[moveList.Count].Score = 0
	moveList.Count++
}

func (p *Position) addWhitePawnCaptureMove(moveList *MoveList, from, to, cap int) {
	if data.RanksBoard[from] == data.Rank7 {
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.WQ, 0), moveList)
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.WR, 0), moveList)
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.WB, 0), moveList)
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.WN, 0), moveList)
	} else {
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.Empty, 0), moveList)
	}
}

func (p *Position) addWhitePawnMove(from, to int, moveList *MoveList) {
	if data.RanksBoard[from] == data.Rank7 {
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.WQ, 0), moveList)
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.WR, 0), moveList)
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.WB, 0), moveList)
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.WN, 0), moveList)
	} else {
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.Empty, 0), moveList)
	}
}

func (p *Position) addBlackPawnMove(from, to int, moveList *MoveList) {
	if data.RanksBoard[from] == data.Rank2 {
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.BQ, 0), moveList)
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.BR, 0), moveList)
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.BB, 0), moveList)
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.BN, 0), moveList)
	} else {
		p.addQuiteMove(MakeMoveInt(from, to, data.Empty, data.Empty, 0), moveList)
	}
}

func (p *Position) addBlackPawnCaptureMove(moveList *MoveList, from, to, cap int) {
	if data.RanksBoard[from] == data.Rank2 {
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.BQ, 0), moveList)
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.BR, 0), moveList)
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.BB, 0), moveList)
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.BN, 0), moveList)
	} else {
		p.addCaptureMove(MakeMoveInt(from, to, cap, data.Empty, 0), moveList)
	}
}

func (p *Position) generateSliderMoves(moveList *MoveList, piece int, includeQuite bool) {
	bitboard := p.Board.GetBitboardForPiece(piece)
	for bitboard != 0 {
		sq := bits.TrailingZeros64(bitboard)
		sq120 := data.Square64ToSquare120[sq]
		if piece == data.WR || piece == data.WQ {
			attack := data.GetRookAttacks(p.Board.Pieces, sq)
			p.generateMovesForSlider(moveList, includeQuite, sq120, attack, p.Board.BlackPieces)
		}
		if piece == data.WB || piece == data.WQ {
			attack := data.GetBishopAttacks(p.Board.Pieces, sq)
			p.generateMovesForSlider(moveList, includeQuite, sq120, attack, p.Board.BlackPieces)
		}
		if piece == data.BR || piece == data.BQ {
			attack := data.GetRookAttacks(p.Board.Pieces, sq)
			p.generateMovesForSlider(moveList, includeQuite, sq120, attack, p.Board.WhitePieces)
		}
		if piece == data.BB || piece == data.BQ {
			attack := data.GetBishopAttacks(p.Board.Pieces, sq)
			p.generateMovesForSlider(moveList, includeQuite, sq120, attack, p.Board.WhitePieces)
		}
		if piece == WK {
			attack := preCalculatedKingMoves[sq]
			p.generateMovesForSlider(moveList, includeQuite, sq120, attack, p.Board.BlackPieces)
		}
		if piece == BK {
			attack := preCalculatedKingMoves[sq]
			p.generateMovesForSlider(moveList, includeQuite, sq120, attack, p.Board.WhitePieces)
		}
		if piece == WN {
			attack := preCalculatedKnightMoves[sq]
			p.generateMovesForSlider(moveList, includeQuite, sq120, attack, p.Board.BlackPieces)
		}
		if piece == BN {
			attack := preCalculatedKnightMoves[sq]
			p.generateMovesForSlider(moveList, includeQuite, sq120, attack, p.Board.WhitePieces)
		}
		bitboard &= bitboard - 1
	}
}

func (p *Position) generateMovesForSlider(moveList *MoveList, includeQuite bool, square int, attackBB, oppositeBB uint64) {
	captures := attackBB & oppositeBB
	for captures != 0 {
		sq := bits.TrailingZeros64(captures)
		captures &= captures - 1
		sq = data.Square64ToSquare120[sq]
		p.addCaptureMove(MakeMoveInt(square, sq, p.Board.PieceAt(data.Square120ToSquare64[sq]), data.Empty, 0), moveList)
	}

	if includeQuite {
		quite := attackBB &^ p.Board.Pieces
		for quite != 0 {
			sq := bits.TrailingZeros64(quite)
			quite &= quite - 1
			sq = data.Square64ToSquare120[sq]
			p.addQuiteMove(MakeMoveInt(square, sq, data.Empty, data.Empty, 0), moveList)
		}
	}
}

func (p *Position) generateWhiteCastleMoves(moveList *MoveList) {
	if (p.CastlePermission & data.WhiteKingCastle) != 0 {
		if p.Board.PieceAt(data.Square120ToSquare64[data.F1]) == data.Empty && p.Board.PieceAt(data.Square120ToSquare64[data.G1]) == data.Empty {
			if !p.SquaresUnderAttack(data.Black, data.Square120ToSquare64[data.E1]) && !p.SquaresUnderAttack(data.Black, data.Square120ToSquare64[data.F1]) {
				p.addQuiteMove(MakeMoveInt(data.E1, data.G1, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
	if (p.CastlePermission & data.WhiteQueenCastle) != 0 {
		if p.Board.PieceAt(data.Square120ToSquare64[data.D1]) == data.Empty && p.Board.PieceAt(data.Square120ToSquare64[data.C1]) == data.Empty && p.Board.PieceAt(data.Square120ToSquare64[data.B1]) == data.Empty {
			if !p.SquaresUnderAttack(data.Black, data.Square120ToSquare64[data.E1]) && !p.SquaresUnderAttack(data.Black, data.Square120ToSquare64[data.D1]) {
				p.addQuiteMove(MakeMoveInt(data.E1, data.C1, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
}

func (p *Position) generateBlackCastleMoves(moveList *MoveList) {
	if (p.CastlePermission & data.BlackKingCastle) != 0 {
		if p.Board.PieceAt(data.Square120ToSquare64[data.F8]) == data.Empty && p.Board.PieceAt(data.Square120ToSquare64[data.G8]) == data.Empty {
			if !p.SquaresUnderAttack(data.White, data.Square120ToSquare64[data.E8]) && !p.SquaresUnderAttack(data.White, data.Square120ToSquare64[data.F8]) {
				p.addQuiteMove(MakeMoveInt(data.E8, data.G8, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
	if (p.CastlePermission & data.BlackQueenCastle) != 0 {
		if p.Board.PieceAt(data.Square120ToSquare64[data.D8]) == data.Empty && p.Board.PieceAt(data.Square120ToSquare64[data.C8]) == data.Empty && p.Board.PieceAt(data.Square120ToSquare64[data.B8]) == data.Empty {
			if !p.SquaresUnderAttack(data.White, data.Square120ToSquare64[data.E8]) && !p.SquaresUnderAttack(data.White, data.Square120ToSquare64[data.D8]) {
				p.addQuiteMove(MakeMoveInt(data.E8, data.C8, data.Empty, data.Empty, data.MFLAGGCA), moveList)
			}
		}
	}
}

func (p *Position) SquaresUnderAttack(side int, sq64 int) bool {
	mask := uint64(1) << sq64
	var attacked uint64
	var enemyBishop, enemyQueen, enemyRook, enemyKnight, enemyKing, enemyPawn uint64
	var king int
	if side == data.Black {
		enemyBishop = p.Board.BlackBishop
		enemyQueen = p.Board.BlackQueen
		enemyRook = p.Board.BlackRook
		enemyKnight = p.Board.BlackKnight
		enemyKing = p.Board.BlackKing
		enemyPawn = p.Board.BlackPawn
		king = BK
	} else {
		enemyBishop = p.Board.WhiteBishop
		enemyQueen = p.Board.WhiteQueen
		enemyRook = p.Board.WhiteRook
		enemyKnight = p.Board.WhiteKnight
		enemyKing = p.Board.WhiteKing
		enemyPawn = p.Board.WhitePawn
		king = WK
	}

	attacked |= p.getBishopAttackedSquares(enemyBishop)
	if (attacked & mask) != 0 {
		return true
	}

	attacked |= p.getRookAttackedSquares(enemyRook)
	if (attacked & mask) != 0 {
		return true
	}

	attacked |= p.getQueenAttackedSquares(enemyQueen)
	if (attacked & mask) != 0 {
		return true
	}

	attacked |= p.getKnightAttackedSquares(enemyKnight)
	if (attacked & mask) != 0 {
		return true
	}

	if king == BK {
		attacked |= p.getBlackPawnAttackedSquares(enemyPawn)

	} else {
		attacked |= p.getWhitePawnAttackedSquares(enemyPawn)
	}
	if (attacked & mask) != 0 {
		return true
	}

	attacked |= p.getKingAttackedSquares(enemyKing)
	return (attacked & mask) != 0
}

func (p *Position) IsKingAttacked() bool {
	var king uint64
	if p.Side == data.White {
		king = p.Board.BlackKing
	} else {
		king = p.Board.WhiteKing
	}
	sq64 := bits.TrailingZeros64(king)
	return p.SquaresUnderAttack(p.Side, sq64)
}

func (p *Position) PrintMoveList() {
	moveList := &MoveList{}
	p.GenerateAllMoves(moveList)
	fmt.Println("\nPrinting move list:")
	for i := 0; i < moveList.Count; i++ {
		fmt.Printf("Move %v: %v (score: %v) %v\n", i+1, io.PrintMove(moveList.Moves[i].Move), moveList.Moves[i].Score, moveList.Moves[i].Move)
	}
	fmt.Printf("Printed %v total moves.\n", moveList.Count)
}

func init() {
	preCalculatedMoves()
}

func preCalculatedMoves() {
	for i := 0; i < 64; i++ {
		preCalculatedKnightMoves[i] = calculateKnightMoves(i)
		preCalculatedKingMoves[i] = calculateKingMoves(i)
	}
}

func calculateKingMoves(sq64 int) uint64 {
	var attacks uint64
	sq120 := data.Square64ToSquare120[sq64]

	for i := 0; i < data.NumDir[WK]; i++ {
		dir := data.PieceDir[WK][i]
		tempSq := sq120 + dir
		if !validate.SquareOnBoard(tempSq) {
			continue
		}
		SetBit(&attacks, data.Square120ToSquare64[tempSq])
	}
	return attacks
}

func calculateKnightMoves(sq64 int) uint64 {
	var attacks uint64
	sq120 := data.Square64ToSquare120[sq64]

	for i := 0; i < data.NumDir[WN]; i++ {
		dir := data.PieceDir[WN][i]
		tempSq := sq120 + dir
		if !validate.SquareOnBoard(tempSq) {
			continue
		}
		SetBit(&attacks, data.Square120ToSquare64[tempSq])
	}
	return attacks
}

func (p *Position) getBishopAttackedSquares(enemyBishop uint64) uint64 {
	var attacked uint64
	for enemyBishop != 0 {
		sq := bits.TrailingZeros64(enemyBishop)
		bishopAttacks := data.GetBishopAttacks(p.Board.Pieces, sq)
		attacked |= bishopAttacks
		enemyBishop &= enemyBishop - 1
	}
	return attacked
}

func (p *Position) getRookAttackedSquares(enemyRook uint64) uint64 {
	var attacked uint64
	for enemyRook != 0 {
		sq := bits.TrailingZeros64(enemyRook)
		rookAttacks := data.GetRookAttacks(p.Board.Pieces, sq)
		attacked |= rookAttacks
		enemyRook &= enemyRook - 1
	}
	return attacked
}

func (p *Position) getQueenAttackedSquares(enemyQueen uint64) uint64 {
	var attacked uint64
	for enemyQueen != 0 {
		sq := bits.TrailingZeros64(enemyQueen)
		bishopAttacks := data.GetBishopAttacks(p.Board.Pieces, sq)
		rookAttacks := data.GetRookAttacks(p.Board.Pieces, sq)
		attacked |= bishopAttacks
		attacked |= rookAttacks
		enemyQueen &= enemyQueen - 1
	}
	return attacked
}

func (p *Position) getKnightAttackedSquares(enemyKnight uint64) uint64 {
	attacked := uint64(0)
	for enemyKnight != 0 {
		sq := bits.TrailingZeros64(enemyKnight)
		attacked |= preCalculatedKnightMoves[sq]
		enemyKnight &= enemyKnight - 1
	}
	return attacked
}

func (p *Position) getKingAttackedSquares(enemyKing uint64) uint64 {
	attacked := uint64(0)
	for enemyKing != 0 {
		sq := bits.TrailingZeros64(enemyKing)
		attacked |= preCalculatedKingMoves[sq]
		enemyKing &= enemyKing - 1
	}
	return attacked
}

func (p *Position) getWhitePawnAttackedSquares(enemyPawn uint64) uint64 {
	var attacked uint64
	for enemyPawn != 0 {
		leftAttacks := (enemyPawn << 7) &^ data.FileBBMask[data.FileH]
		attacked |= leftAttacks

		// Generate attacks to the right
		rightAttacks := (enemyPawn << 9) &^ data.FileBBMask[data.FileA]
		attacked |= rightAttacks

		enemyPawn &= enemyPawn - 1
	}
	return attacked
}

func (p *Position) getBlackPawnAttackedSquares(enemyPawn uint64) uint64 {
	var attacked uint64
	for enemyPawn != 0 {
		leftAttacks := (enemyPawn >> 9) &^ data.FileBBMask[data.FileH]
		attacked |= leftAttacks

		rightAttacks := (enemyPawn >> 7) &^ data.FileBBMask[data.FileA]
		attacked |= rightAttacks

		enemyPawn &= enemyPawn - 1
	}
	return attacked
}
