package epd

import (
	"testing"
)

func TestParseEPD(t *testing.T) {
	tests := []struct {
		name            string
		epdLine         string
		expectError     bool
		expectedBM      string
		expectedComment string
	}{
		{
			name:        "Basic EPD with best move",
			epdLine:     "1R6/1brk2p1/4p2p/p1P1Pp2/P7/6P1/1P4P1/2R3K1 w - - 0 1 bm b8b7",
			expectError: false,
			expectedBM:  "b8b7",
		},
		{
			name:        "EPD with castling best move",
			epdLine:     "r1b2rk1/ppq1bppp/2p1pn2/8/2NP4/2N1P3/PP2BPPP/2RQK2R w K - 0 1 bm e1g1",
			expectError: false,
			expectedBM:  "e1g1",
		},
		{
			name:        "Empty line",
			epdLine:     "",
			expectError: true,
		},
		{
			name:        "Comment line",
			epdLine:     "# This is a comment",
			expectError: true,
		},
		{
			name:        "Invalid FEN",
			epdLine:     "invalid_fen w - - 0 1 bm Nf3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position, err := ParseEPD(tt.epdLine)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if position.BestMove != tt.expectedBM {
				t.Errorf("Expected best move %s, got %s", tt.expectedBM, position.BestMove)
			}

			if tt.expectedComment != "" && position.Comment != tt.expectedComment {
				t.Errorf("Expected comment %s, got %s", tt.expectedComment, position.Comment)
			}

			// Verify that board was parsed correctly
			if position.Board == nil {
				t.Error("Board should not be nil")
			}
		})
	}
}

func TestParseEPDFile(t *testing.T) {
	epdContent := `1R6/1brk2p1/4p2p/p1P1Pp2/P7/6P1/1P4P1/2R3K1 w - - 0 1 bm b8b7
4r1k1/p1qr1p2/2pb1Bp1/1p5p/3P1n1R/1B3P2/PP3PK1/2Q4R w - - 0 1 bm c1f4
# This is a comment line
r1b2rk1/ppq1bppp/2p1pn2/8/2NP4/2N1P3/PP2BPPP/2RQK2R w K - 0 1 bm e1g1
`

	positions, err := ParseEPDFile(epdContent)
	if err != nil {
		t.Fatalf("Unexpected error parsing EPD file: %v", err)
	}

	expectedPositions := 3 // Should skip the comment line
	if len(positions) != expectedPositions {
		t.Errorf("Expected %d positions, got %d", expectedPositions, len(positions))
	}

	// Test first position
	if positions[0].BestMove != "b8b7" {
		t.Errorf("Expected first position best move b8b7, got %s", positions[0].BestMove)
	}

	// Test second position
	if positions[1].BestMove != "c1f4" {
		t.Errorf("Expected second position best move c1f4, got %s", positions[1].BestMove)
	}

	// Test third position (castling)
	if positions[2].BestMove != "e1g1" {
		t.Errorf("Expected third position best move e1g1, got %s", positions[2].BestMove)
	}
}
