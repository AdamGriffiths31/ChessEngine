package evaluation

import (
	"github.com/AdamGriffiths31/ChessEngine/board"
)

// Bishop evaluation constants
const (
	// Bishop pair bonus
	BishopPairBonus = 50

	// Bad bishop penalties (based on pawn obstruction)
	BadBishopPenalty     = -15 // Per blocked center pawn
	SemiBadBishopPenalty = -8  // Per blocked non-center pawn

	// Long diagonal control
	LongDiagonalControl    = 25 // Full control of long diagonal
	PartialDiagonalControl = 15 // Partial control

	// Color complex
	ColorComplexDominance = 30 // Opponent missing bishop of that color

	// X-ray bonus
	XRayAttackBonus = 20 // X-ray to valuable piece
	XRayThreatBonus = 10 // Potential x-ray

	// Mobility (per square)
	BishopMobilityUnit   = 3
	BishopTrappedPenalty = -50 // Less than 3 moves
)

// Long diagonal masks
var (
	LongDiagonalA1H8 board.Bitboard
	LongDiagonalH1A8 board.Bitboard
)

// Initialize diagonal masks
func init() {
	// A1-H8 diagonal
	for i := 0; i < 8; i++ {
		LongDiagonalA1H8 = LongDiagonalA1H8.SetBit(board.FileRankToSquare(i, i))
	}

	// H1-A8 diagonal
	for i := 0; i < 8; i++ {
		LongDiagonalH1A8 = LongDiagonalH1A8.SetBit(board.FileRankToSquare(7-i, i))
	}
}

// evaluateBishopPairBonus calculates the bishop pair bonus for both sides
func evaluateBishopPairBonus(b *board.Board) int {
	// Get bishop bitboards
	whiteLightBishops := b.GetPieceBitboard(board.WhiteBishop) & board.LightSquares
	whiteDarkBishops := b.GetPieceBitboard(board.WhiteBishop) & board.DarkSquares
	blackLightBishops := b.GetPieceBitboard(board.BlackBishop) & board.LightSquares
	blackDarkBishops := b.GetPieceBitboard(board.BlackBishop) & board.DarkSquares

	score := 0

	// Bishop pair bonus
	if whiteLightBishops != 0 && whiteDarkBishops != 0 {
		score += BishopPairBonus
	}
	if blackLightBishops != 0 && blackDarkBishops != 0 {
		score -= BishopPairBonus
	}

	return score
}

// evaluateBishops evaluates all bishop-specific features
func evaluateBishops(b *board.Board) int {
	score := 0

	// Get bishop bitboards
	whiteLightBishops := b.GetPieceBitboard(board.WhiteBishop) & board.LightSquares
	whiteDarkBishops := b.GetPieceBitboard(board.WhiteBishop) & board.DarkSquares
	blackLightBishops := b.GetPieceBitboard(board.BlackBishop) & board.LightSquares
	blackDarkBishops := b.GetPieceBitboard(board.BlackBishop) & board.DarkSquares

	// Bishop pair bonus
	score += evaluateBishopPairBonus(b)

	// Evaluate individual bishops
	score += evaluateBishopFeatures(b, b.GetPieceBitboard(board.WhiteBishop), board.BitboardWhite)
	score -= evaluateBishopFeatures(b, b.GetPieceBitboard(board.BlackBishop), board.BitboardBlack)

	// Color complex evaluation
	score += evaluateColorComplex(b, whiteLightBishops, whiteDarkBishops,
		blackLightBishops, blackDarkBishops)

	return score
}

// evaluateBishopFeatures evaluates features for bishops of one color
func evaluateBishopFeatures(b *board.Board, bishops board.Bitboard, color board.BitboardColor) int {
	if bishops == 0 {
		return 0
	}

	score := 0

	// Get relevant bitboards
	var friendlyPawns board.Bitboard
	if color == board.BitboardWhite {
		friendlyPawns = b.GetPieceBitboard(board.WhitePawn)
	} else {
		friendlyPawns = b.GetPieceBitboard(board.BlackPawn)
	}

	// Process each bishop
	for bishops != 0 {
		square, newBishops := bishops.PopLSB()
		bishops = newBishops

		// Evaluate individual bishop features
		score += evaluateBadBishop(b, square, friendlyPawns, color)
		score += evaluateLongDiagonalControl(b, square, color)
		score += evaluateBishopMobility(b, square, color)
		score += evaluateXRayAttacks(b, square, color)
	}

	return score
}

// evaluateBadBishop checks if bishop is blocked by own pawns
func evaluateBadBishop(b *board.Board, bishopSquare int, friendlyPawns board.Bitboard, color board.BitboardColor) int {
	// Determine bishop color (light or dark square)
	bishopOnLightSquare := (bishopSquare % 2) == ((bishopSquare / 8) % 2)

	penalty := 0

	// Check central pawns (more important)
	centralPawns := friendlyPawns & board.CenterFiles & board.CenterRanks

	// Count blocked central pawns on same color squares
	for centralPawns != 0 {
		pawnSquare, newPawns := centralPawns.PopLSB()
		centralPawns = newPawns

		pawnOnLightSquare := (pawnSquare % 2) == ((pawnSquare / 8) % 2)
		if pawnOnLightSquare == bishopOnLightSquare {
			// Check if pawn is blocked
			if isPawnBlocked(b, pawnSquare, color) {
				penalty += BadBishopPenalty
			}
		}
	}

	// Check non-central pawns (less important)
	nonCentralPawns := friendlyPawns &^ (board.CenterFiles & board.CenterRanks)

	for nonCentralPawns != 0 {
		pawnSquare, newPawns := nonCentralPawns.PopLSB()
		nonCentralPawns = newPawns

		pawnOnLightSquare := (pawnSquare % 2) == ((pawnSquare / 8) % 2)
		if pawnOnLightSquare == bishopOnLightSquare {
			if isPawnBlocked(b, pawnSquare, color) {
				penalty += SemiBadBishopPenalty
			}
		}
	}

	return penalty
}

// isPawnBlocked checks if a pawn is blocked
func isPawnBlocked(b *board.Board, pawnSquare int, color board.BitboardColor) bool {
	file, rank := board.SquareToFileRank(pawnSquare)

	// Check square in front of pawn
	var targetRank int
	if color == board.BitboardWhite {
		targetRank = rank + 1
		if targetRank > 7 {
			return false
		}
	} else {
		targetRank = rank - 1
		if targetRank < 0 {
			return false
		}
	}

	targetSquare := board.FileRankToSquare(file, targetRank)
	return !b.IsSquareEmptyBitboard(targetSquare)
}

var (
	CenterFiles = board.FileMask(3) | board.FileMask(4) // d and e files
	CenterRanks = board.RankMask(3) | board.RankMask(4) // 4th and 5th ranks
)

// evaluateLongDiagonalControl evaluates control of long diagonals
func evaluateLongDiagonalControl(b *board.Board, bishopSquare int, color board.BitboardColor) int {
	// Get bishop attacks
	attacks := board.GetBishopAttacks(bishopSquare, b.AllPieces)

	score := 0

	// Check A1-H8 diagonal control
	a1h8Control := attacks & LongDiagonalA1H8
	if a1h8Control != 0 {
		controlCount := a1h8Control.PopCount()
		if controlCount >= 5 {
			score += LongDiagonalControl
		} else if controlCount >= 3 {
			score += PartialDiagonalControl
		}
	}

	// Check H1-A8 diagonal control
	h1a8Control := attacks & LongDiagonalH1A8
	if h1a8Control != 0 {
		controlCount := h1a8Control.PopCount()
		if controlCount >= 5 {
			score += LongDiagonalControl
		} else if controlCount >= 3 {
			score += PartialDiagonalControl
		}
	}

	// Extra bonus if bishop is actually on a long diagonal
	bishopBit := board.Bitboard(1) << uint(bishopSquare)
	if (bishopBit & (LongDiagonalA1H8 | LongDiagonalH1A8)) != 0 {
		score += 5
	}

	return score
}

// evaluateColorComplex checks for color complex advantages
func evaluateColorComplex(b *board.Board, whiteLightBishops, whiteDarkBishops,
	blackLightBishops, blackDarkBishops board.Bitboard) int {

	score := 0

	// White has light-squared bishop, black doesn't
	if whiteLightBishops != 0 && blackLightBishops == 0 {
		score += evaluateColorDominance(b, board.LightSquares, board.BitboardWhite)
	}

	// White has dark-squared bishop, black doesn't
	if whiteDarkBishops != 0 && blackDarkBishops == 0 {
		score += evaluateColorDominance(b, board.DarkSquares, board.BitboardWhite)
	}

	// Black has light-squared bishop, white doesn't
	if blackLightBishops != 0 && whiteLightBishops == 0 {
		score -= evaluateColorDominance(b, board.LightSquares, board.BitboardBlack)
	}

	// Black has dark-squared bishop, white doesn't
	if blackDarkBishops != 0 && whiteDarkBishops == 0 {
		score -= evaluateColorDominance(b, board.DarkSquares, board.BitboardBlack)
	}

	return score
}

// evaluateColorDominance evaluates advantage on specific color squares
func evaluateColorDominance(b *board.Board, colorMask board.Bitboard, dominantColor board.BitboardColor) int {
	bonus := ColorComplexDominance

	// Extra bonus if enemy has many pawns on that color
	var enemyPawns board.Bitboard
	if dominantColor == board.BitboardWhite {
		enemyPawns = b.GetPieceBitboard(board.BlackPawn)
	} else {
		enemyPawns = b.GetPieceBitboard(board.WhitePawn)
	}

	enemyPawnsOnColor := (enemyPawns & colorMask).PopCount()
	bonus += enemyPawnsOnColor * 3 // 3 points per enemy pawn on that color

	return bonus
}

// evaluateXRayAttacks detects x-ray attacks through pieces
func evaluateXRayAttacks(b *board.Board, bishopSquare int, color board.BitboardColor) int {
	score := 0

	// Get direct bishop attacks
	directAttacks := board.GetBishopAttacks(bishopSquare, b.AllPieces)

	// Get theoretical attacks if board was empty
	emptyBoardAttacks := board.GetBishopAttacks(bishopSquare, 0)

	// X-ray squares are those not directly attacked but on bishop rays
	xraySquares := emptyBoardAttacks &^ directAttacks

	// Check each x-ray square for valuable targets
	enemyColor := board.OppositeBitboardColor(color)

	for xraySquares != 0 {
		xraySquare, newSquares := xraySquares.PopLSB()
		xraySquares = newSquares

		// Check if there's a valuable enemy piece on this square
		piece := b.GetPieceOnSquare(xraySquare)
		if piece != board.Empty {
			pieceColor := getPieceColor(piece)
			if pieceColor == enemyColor {
				// Check if there's exactly one piece between bishop and target
				between := board.GetBetween(bishopSquare, xraySquare)
				blockers := between & b.AllPieces

				if blockers.PopCount() == 1 {
					// Valid x-ray
					value := getXRayValue(piece)
					score += value
				}
			}
		}
	}

	return score
}

// getXRayValue returns the value of x-ray attack based on target piece
func getXRayValue(targetPiece board.Piece) int {
	switch targetPiece {
	case board.BlackQueen, board.WhiteQueen:
		return XRayAttackBonus + 10
	case board.BlackRook, board.WhiteRook:
		return XRayAttackBonus
	case board.BlackKing, board.WhiteKing:
		return XRayAttackBonus + 5
	default:
		return XRayThreatBonus
	}
}

// getPieceColor helper function
func getPieceColor(piece board.Piece) board.BitboardColor {
	if piece >= 'A' && piece <= 'Z' {
		return board.BitboardWhite
	}
	return board.BitboardBlack
}

// evaluateBishopMobility evaluates bishop mobility
func evaluateBishopMobility(b *board.Board, square int, color board.BitboardColor) int {
	// Get bishop attacks
	attacks := board.GetBishopAttacks(square, b.AllPieces)

	// Remove squares occupied by friendly pieces
	friendlyPieces := b.GetColorBitboard(color)
	validMoves := attacks &^ friendlyPieces

	// Count mobility
	mobility := validMoves.PopCount()

	// Check if bishop is trapped
	if mobility < 3 {
		return BishopTrappedPenalty
	}

	// Base mobility bonus
	score := mobility * BishopMobilityUnit

	// Extra bonus for forward mobility
	_, rank := board.SquareToFileRank(square)
	var forwardMoves board.Bitboard

	if color == board.BitboardWhite {
		// Count moves to higher ranks
		for r := rank + 1; r < 8; r++ {
			forwardMoves |= validMoves & board.RankMask(r)
		}
	} else {
		// Count moves to lower ranks
		for r := 0; r < rank; r++ {
			forwardMoves |= validMoves & board.RankMask(r)
		}
	}

	score += forwardMoves.PopCount() * 2 // Extra 2 points per forward move

	return score
}
