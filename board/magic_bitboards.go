package board


// Magic bitboard implementation for sliding pieces (rooks, bishops, queens)

// MagicEntry represents a magic number entry for a square
type MagicEntry struct {
	Mask   Bitboard // Relevant occupancy mask (excludes edges for sliding pieces)
	Magic  uint64   // Magic number for multiplication
	Shift  int      // Right shift amount after magic multiplication
	Offset int      // Offset into the attack table
}

var (
	// Magic entries for each square
	RookMagics   [64]MagicEntry
	BishopMagics [64]MagicEntry

	// Attack lookup tables
	RookAttacks   []Bitboard // Global attack table for all rook positions
	BishopAttacks []Bitboard // Global attack table for all bishop positions

	// Relevant occupancy bits for each square
	RookRelevantBits   [64]int
	BishopRelevantBits [64]int

)

// Pre-computed magic numbers
var rookMagicNumbers = [64]uint64{
	0x8a80104000800020, 0x140002000100040, 0x2801880a0017001, 0x100081001000420, 0x200020010080420, 0x3001c0002010008, 0x8480008002000100, 0x2080088004402900,
	0x800098204000, 0x2024401000200040, 0x100802000801000, 0x120800800801000, 0x208808088000400, 0x2802200800400, 0x2200800100020080, 0x801000060821100,
	0x80044006422000, 0x100808020004000, 0x12108a0010204200, 0x140848010000802, 0x481828014002800, 0x8094004002004100, 0x4010040010010802, 0x20008806104,
	0x100400080208000, 0x2040002120081000, 0x21200680100081, 0x20100080080080, 0x2000a00200410, 0x20080800400, 0x80088400100102, 0x80004600042881,
	0x4040008040800020, 0x440003000200801, 0x4200011004500, 0x188020010100100, 0x14800401802800, 0x2080040080800200, 0x124080204001001, 0x200046502000484,
	0x480400080088020, 0x1000422010034000, 0x30200100110040, 0x100021010009, 0x2002080100110004, 0x202008004008002, 0x20020004010100, 0x2048440040820001,
	0x101002200408200, 0x40802000401080, 0x4008142004410100, 0x2060820c0120200, 0x1001004080100, 0x20c020080040080, 0x2935610830022400, 0x44440041009200,
	0x280001040802101, 0x2100190040002085, 0x80c0084100102001, 0x4024081001000421, 0x20030a0244872, 0x12001008414402, 0x2006104900a0804, 0x1004081002402,
}

var bishopMagicNumbers = [64]uint64{
	0x40040844404084, 0x2004208a004208, 0x10190041080202, 0x108060845042010, 0x581104180800210, 0x2112080446200010, 0x1080820820060210, 0x3c0808410220200,
	0x4050404440404, 0x21001420088, 0x24d0080801082102, 0x1020a0a020400, 0x40308200402, 0x4011002100800, 0x401484104104005, 0x801010402020200,
	0x400210c3880100, 0x404022024108200, 0x810018200204102, 0x4002801a02003, 0x85040820080400, 0x810102c808880400, 0x2002410088800, 0x2002410088800,
	0x8002100400820, 0x1010100200424202, 0x840050860000002, 0x840050860000002, 0x1040080020800080, 0x1040080020800080, 0x42044200040802, 0x42044200040802,
	0x2040820080400, 0x2040820080400, 0x412824080202000, 0x412824080202000, 0x80208410220100, 0x80208410220100, 0x40400000801a00, 0x40400000801a00,
	0x400000020080021, 0x400000020080021, 0x800828028020000, 0x800828028020000, 0x8080080020004, 0x8080080020004, 0x2000204100041004, 0x2000204100041004,
	0x204420081020400, 0x204420081020400, 0x482000904420000, 0x482000904420000, 0x40408000400080, 0x40408000400080, 0x8080202000841, 0x8080202000841,
	0x90200046800, 0x90200046800, 0x420208080100, 0x420208080100, 0x82001002001080, 0x82001002001080, 0xa00080410004100, 0xa00080410004100,
}

// init automatically initializes magic bitboard tables when the package is loaded
func init() {
	initializeRelevantOccupancy()
	initializeRookMagics()
	initializeBishopMagics()
}

// initializeRelevantOccupancy calculates relevant occupancy bits for each square
func initializeRelevantOccupancy() {
	for square := 0; square < 64; square++ {
		RookRelevantBits[square] = rookRelevantOccupancy(square).PopCount()
		BishopRelevantBits[square] = bishopRelevantOccupancy(square).PopCount()
	}
}

// rookRelevantOccupancy returns the relevant occupancy mask for a rook on a given square
func rookRelevantOccupancy(square int) Bitboard {
	var mask Bitboard
	file, rank := SquareToFileRank(square)

	// Horizontal (exclude board edges, but include squares next to edges)
	for f := file + 1; f <= 6; f++ {
		mask = mask.SetBit(FileRankToSquare(f, rank))
	}
	for f := file - 1; f >= 1; f-- {
		mask = mask.SetBit(FileRankToSquare(f, rank))
	}

	// Vertical (exclude board edges, but include squares next to edges)
	for r := rank + 1; r <= 6; r++ {
		mask = mask.SetBit(FileRankToSquare(file, r))
	}
	for r := rank - 1; r >= 1; r-- {
		mask = mask.SetBit(FileRankToSquare(file, r))
	}

	return mask
}

// bishopRelevantOccupancy returns the relevant occupancy mask for a bishop on a given square
func bishopRelevantOccupancy(square int) Bitboard {
	var mask Bitboard
	file, rank := SquareToFileRank(square)

	// Diagonal up-right (exclude board edges)
	for f, r := file+1, rank+1; f <= 6 && r <= 6; f, r = f+1, r+1 {
		mask = mask.SetBit(FileRankToSquare(f, r))
	}

	// Diagonal up-left (exclude board edges)
	for f, r := file-1, rank+1; f >= 1 && r <= 6; f, r = f-1, r+1 {
		mask = mask.SetBit(FileRankToSquare(f, r))
	}

	// Diagonal down-right (exclude board edges)
	for f, r := file+1, rank-1; f <= 6 && r >= 1; f, r = f+1, r-1 {
		mask = mask.SetBit(FileRankToSquare(f, r))
	}

	// Diagonal down-left (exclude board edges)
	for f, r := file-1, rank-1; f >= 1 && r >= 1; f, r = f-1, r-1 {
		mask = mask.SetBit(FileRankToSquare(f, r))
	}

	return mask
}

// initializeRookMagics initializes magic bitboards for rooks
func initializeRookMagics() {
	offset := 0

	for square := 0; square < 64; square++ {
		relevantOccupancy := rookRelevantOccupancy(square)
		relevantBits := relevantOccupancy.PopCount()

		RookMagics[square] = MagicEntry{
			Mask:   relevantOccupancy,
			Magic:  rookMagicNumbers[square],
			Shift:  64 - relevantBits,
			Offset: offset,
		}

		// Generate all possible occupancy variations
		occupancyVariations := 1 << relevantBits
		offset += occupancyVariations
	}

	// Allocate attack table
	RookAttacks = make([]Bitboard, offset)

	// Fill attack table
	for square := 0; square < 64; square++ {
		magic := RookMagics[square]
		relevantBits := RookRelevantBits[square]
		occupancyVariations := 1 << relevantBits

		for i := 0; i < occupancyVariations; i++ {
			occupancy := indexToOccupancy(i, magic.Mask)
			magicIndex := (occupancy * Bitboard(magic.Magic)) >> magic.Shift
			RookAttacks[magic.Offset+int(magicIndex)] = rookAttacksOnTheFly(square, occupancy)
		}
	}
}

// initializeBishopMagics initializes magic bitboards for bishops
func initializeBishopMagics() {
	offset := 0

	for square := 0; square < 64; square++ {
		relevantOccupancy := bishopRelevantOccupancy(square)
		relevantBits := relevantOccupancy.PopCount()

		BishopMagics[square] = MagicEntry{
			Mask:   relevantOccupancy,
			Magic:  bishopMagicNumbers[square],
			Shift:  64 - relevantBits,
			Offset: offset,
		}

		// Generate all possible occupancy variations
		occupancyVariations := 1 << relevantBits
		offset += occupancyVariations
	}

	// Allocate attack table
	BishopAttacks = make([]Bitboard, offset)

	// Fill attack table
	for square := 0; square < 64; square++ {
		magic := BishopMagics[square]
		relevantBits := BishopRelevantBits[square]
		occupancyVariations := 1 << relevantBits

		for i := 0; i < occupancyVariations; i++ {
			occupancy := indexToOccupancy(i, magic.Mask)
			magicIndex := (occupancy * Bitboard(magic.Magic)) >> magic.Shift
			BishopAttacks[magic.Offset+int(magicIndex)] = bishopAttacksOnTheFly(square, occupancy)
		}
	}
}

// indexToOccupancy converts an index to an occupancy bitboard based on a mask
func indexToOccupancy(index int, mask Bitboard) Bitboard {
	var occupancy Bitboard
	bits := mask.BitList()

	for i := 0; i < len(bits); i++ {
		if (index & (1 << i)) != 0 {
			occupancy = occupancy.SetBit(bits[i])
		}
	}

	return occupancy
}

// rookAttacksOnTheFly generates rook attacks for a given square and occupancy
func rookAttacksOnTheFly(square int, occupancy Bitboard) Bitboard {
	var attacks Bitboard
	file, rank := SquareToFileRank(square)

	// North
	for r := rank + 1; r <= 7; r++ {
		targetSquare := FileRankToSquare(file, r)
		attacks = attacks.SetBit(targetSquare)
		if occupancy.HasBit(targetSquare) {
			break
		}
	}

	// South
	for r := rank - 1; r >= 0; r-- {
		targetSquare := FileRankToSquare(file, r)
		attacks = attacks.SetBit(targetSquare)
		if occupancy.HasBit(targetSquare) {
			break
		}
	}

	// East
	for f := file + 1; f <= 7; f++ {
		targetSquare := FileRankToSquare(f, rank)
		attacks = attacks.SetBit(targetSquare)
		if occupancy.HasBit(targetSquare) {
			break
		}
	}

	// West
	for f := file - 1; f >= 0; f-- {
		targetSquare := FileRankToSquare(f, rank)
		attacks = attacks.SetBit(targetSquare)
		if occupancy.HasBit(targetSquare) {
			break
		}
	}

	return attacks
}

// bishopAttacksOnTheFly generates bishop attacks for a given square and occupancy
func bishopAttacksOnTheFly(square int, occupancy Bitboard) Bitboard {
	var attacks Bitboard
	file, rank := SquareToFileRank(square)

	// Northeast
	for f, r := file+1, rank+1; f <= 7 && r <= 7; f, r = f+1, r+1 {
		targetSquare := FileRankToSquare(f, r)
		attacks = attacks.SetBit(targetSquare)
		if occupancy.HasBit(targetSquare) {
			break
		}
	}

	// Northwest
	for f, r := file-1, rank+1; f >= 0 && r <= 7; f, r = f-1, r+1 {
		targetSquare := FileRankToSquare(f, r)
		attacks = attacks.SetBit(targetSquare)
		if occupancy.HasBit(targetSquare) {
			break
		}
	}

	// Southeast
	for f, r := file+1, rank-1; f <= 7 && r >= 0; f, r = f+1, r-1 {
		targetSquare := FileRankToSquare(f, r)
		attacks = attacks.SetBit(targetSquare)
		if occupancy.HasBit(targetSquare) {
			break
		}
	}

	// Southwest
	for f, r := file-1, rank-1; f >= 0 && r >= 0; f, r = f-1, r-1 {
		targetSquare := FileRankToSquare(f, r)
		attacks = attacks.SetBit(targetSquare)
		if occupancy.HasBit(targetSquare) {
			break
		}
	}

	return attacks
}

// Public API functions for getting sliding piece attacks

// GetRookAttacks returns rook attacks for a given square and occupancy
func GetRookAttacks(square int, occupancy Bitboard) Bitboard {
	if square < 0 || square > 63 {
		return 0
	}

	// For now, use on-the-fly calculation to ensure correctness
	// TODO: Fix magic bitboard implementation
	return rookAttacksOnTheFly(square, occupancy)

	// Original magic bitboard code (temporarily disabled)
	/*
		magic := RookMagics[square]
		relevantOccupancy := occupancy & magic.Mask
		magicIndex := (relevantOccupancy * Bitboard(magic.Magic)) >> magic.Shift
		return RookAttacks[magic.Offset+int(magicIndex)]
	*/
}

// GetBishopAttacks returns bishop attacks for a given square and occupancy
func GetBishopAttacks(square int, occupancy Bitboard) Bitboard {
	if square < 0 || square > 63 {
		return 0
	}

	// For now, use on-the-fly calculation to ensure correctness
	// TODO: Fix magic bitboard implementation
	return bishopAttacksOnTheFly(square, occupancy)

	// Original magic bitboard code (temporarily disabled)
	/*
		magic := BishopMagics[square]
		relevantOccupancy := occupancy & magic.Mask
		magicIndex := (relevantOccupancy * Bitboard(magic.Magic)) >> magic.Shift
		return BishopAttacks[magic.Offset+int(magicIndex)]
	*/
}

// GetQueenAttacks returns queen attacks for a given square and occupancy
func GetQueenAttacks(square int, occupancy Bitboard) Bitboard {
	return GetRookAttacks(square, occupancy) | GetBishopAttacks(square, occupancy)
}

