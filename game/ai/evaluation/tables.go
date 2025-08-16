package evaluation

// Piece-square tables from Chess Programming Wiki - Simplified Evaluation Function
// All values are from white's perspective (rank 0 = white back rank, rank 7 = black back rank)

// KnightTable contains positional bonuses/penalties for knights
// Knights prefer center squares and avoid edges and corners
var KnightTable = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50, // rank 1 (white back rank)
	-40, -20, 0, 5, 5, 0, -20, -40, // rank 2
	-30, 5, 10, 15, 15, 10, 5, -30, // rank 3
	-30, 0, 15, 20, 20, 15, 0, -30, // rank 4
	-30, 5, 15, 20, 20, 15, 5, -30, // rank 5
	-30, 0, 10, 15, 15, 10, 0, -30, // rank 6
	-40, -20, 0, 0, 0, 0, -20, -40, // rank 7
	-50, -40, -30, -30, -30, -30, -40, -50, // rank 8 (black back rank)
}

// BishopTable contains positional bonuses/penalties for bishops
// Bishops avoid corners and borders, prefer squares like b3, c4, b5, d3 and central ones
var BishopTable = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20, // rank 1 (white back rank)
	-10, 0, 0, 0, 0, 0, 0, -10, // rank 2
	-10, 0, 5, 10, 10, 5, 0, -10, // rank 3
	-10, 5, 5, 10, 10, 5, 5, -10, // rank 4
	-10, 0, 10, 10, 10, 10, 0, -10, // rank 5
	-10, 10, 10, 10, 10, 10, 10, -10, // rank 6
	-10, 5, 0, 0, 0, 0, 5, -10, // rank 7
	-20, -10, -10, -10, -10, -10, -10, -20, // rank 8 (black back rank)
}

// RookTable contains positional bonuses/penalties for rooks
// Rooks prefer to centralize, occupy the 7th rank, and avoid a, h columns
var RookTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0, // rank 1 (white back rank)
	5, 10, 10, 10, 10, 10, 10, 5, // rank 2
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 3
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 4
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 5
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 6
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 7
	0, 0, 0, 5, 5, 0, 0, 0, // rank 8 (black back rank)
}

// QueenTable contains positional bonuses/penalties for queens
// Queens avoid corners and prefer to stay safe while maintaining central influence
var QueenTable = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20, // rank 1 (white back rank)
	-10, 0, 0, 0, 0, 0, 0, -10, // rank 2
	-10, 0, 5, 5, 5, 5, 0, -10, // rank 3
	-5, 0, 5, 5, 5, 5, 0, -5, // rank 4
	0, 0, 5, 5, 5, 5, 0, -5, // rank 5
	-10, 5, 5, 5, 5, 5, 0, -10, // rank 6
	-10, 0, 5, 0, 0, 0, 0, -10, // rank 7
	-20, -10, -10, -5, -5, -10, -10, -20, // rank 8 (black back rank)
}

// KingTable contains positional bonuses/penalties for kings (middle game)
// Kings should stay behind pawn shelter and avoid center during middle game
var KingTable = [64]int{
	-30, -40, -40, -50, -50, -40, -40, -30, // rank 1 (white back rank)
	-30, -40, -40, -50, -50, -40, -40, -30, // rank 2
	-30, -40, -40, -50, -50, -40, -40, -30, // rank 3
	-30, -40, -40, -50, -50, -40, -40, -30, // rank 4
	-20, -30, -30, -40, -40, -30, -30, -20, // rank 5
	-10, -20, -20, -20, -20, -20, -20, -10, // rank 6
	20, 20, 0, 0, 0, 0, 20, 20, // rank 7
	20, 30, 10, 0, 0, 10, 30, 20, // rank 8 (black back rank)
}

// PawnTable contains positional bonuses/penalties for pawns
// Pawns are encouraged to advance, with bonuses for advanced pawns and penalties for certain positions
var PawnTable = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0, // rank 1 (white back rank - pawns can't be here)
	5, 10, 10, -20, -20, 10, 10, 5, // rank 2 (white's pawn starting rank - penalty for unmoved center pawns)
	5, -5, -10, 0, 0, -10, -5, 5, // rank 3
	0, 0, 0, 20, 20, 0, 0, 0, // rank 4
	5, 5, 10, 25, 25, 10, 5, 5, // rank 5
	10, 10, 20, 30, 30, 20, 10, 10, // rank 6
	50, 50, 50, 50, 50, 50, 50, 50, // rank 7 (near promotion - high bonus)
	0, 0, 0, 0, 0, 0, 0, 0, // rank 8 (black back rank - pawns can't be here)
}
