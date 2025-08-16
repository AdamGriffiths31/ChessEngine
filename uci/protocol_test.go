package uci

import (
	"reflect"
	"testing"
	"time"
)

func TestProtocolHandler_ParseCommand(t *testing.T) {
	handler := NewProtocolHandler()

	tests := []struct {
		name     string
		input    string
		expected Command
	}{
		{
			name:  "uci command",
			input: "uci",
			expected: Command{
				Name: "uci",
				Args: []string{},
			},
		},
		{
			name:  "position startpos",
			input: "position startpos",
			expected: Command{
				Name: "position",
				Args: []string{"startpos"},
			},
		},
		{
			name:  "position with moves",
			input: "position startpos moves e2e4 e7e5",
			expected: Command{
				Name: "position",
				Args: []string{"startpos", "moves", "e2e4", "e7e5"},
			},
		},
		{
			name:  "go depth",
			input: "go depth 6",
			expected: Command{
				Name: "go",
				Args: []string{"depth", "6"},
			},
		},
		{
			name:  "empty input",
			input: "",
			expected: Command{
				Name: "",
				Args: []string{},
			},
		},
		{
			name:  "whitespace only",
			input: "   ",
			expected: Command{
				Name: "",
				Args: []string{},
			},
		},
		{
			name:  "setoption command",
			input: "setoption name Hash value 256",
			expected: Command{
				Name: "setoption",
				Args: []string{"name", "Hash", "value", "256"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.ParseCommand(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseCommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProtocolHandler_ParsePosition(t *testing.T) {
	handler := NewProtocolHandler()

	tests := []struct {
		name          string
		args          []string
		expectedFEN   string
		expectedMoves []string
		expectedErr   bool
	}{
		{
			name:          "startpos only",
			args:          []string{"startpos"},
			expectedFEN:   "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedMoves: nil,
		},
		{
			name:          "startpos with moves",
			args:          []string{"startpos", "moves", "e2e4", "e7e5"},
			expectedFEN:   "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
			expectedMoves: []string{"e2e4", "e7e5"},
		},
		{
			name:          "fen position",
			args:          []string{"fen", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR", "b", "KQkq", "e3", "0", "1"},
			expectedFEN:   "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			expectedMoves: nil,
		},
		{
			name:          "fen position with moves",
			args:          []string{"fen", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR", "b", "KQkq", "e3", "0", "1", "moves", "e7e5"},
			expectedFEN:   "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
			expectedMoves: []string{"e7e5"},
		},
		{
			name:        "empty args",
			args:        []string{},
			expectedErr: true,
		},
		{
			name:        "invalid position type",
			args:        []string{"invalid"},
			expectedErr: true,
		},
		{
			name:        "incomplete fen",
			args:        []string{"fen", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR"},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fen, moves, err := handler.ParsePosition(tt.args)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("ParsePosition() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParsePosition() unexpected error: %v", err)
				return
			}

			if fen != tt.expectedFEN {
				t.Errorf("ParsePosition() fen = %v, want %v", fen, tt.expectedFEN)
			}

			if (moves == nil && tt.expectedMoves != nil) ||
				(moves != nil && tt.expectedMoves == nil) ||
				!reflect.DeepEqual(moves, tt.expectedMoves) {
				t.Errorf("ParsePosition() moves = %v, want %v", moves, tt.expectedMoves)
			}
		})
	}
}

func TestProtocolHandler_ParseGo(t *testing.T) {
	handler := NewProtocolHandler()

	tests := []struct {
		name     string
		args     []string
		expected SearchParams
	}{
		{
			name: "depth only",
			args: []string{"depth", "6"},
			expected: SearchParams{
				Depth: 6,
			},
		},
		{
			name: "movetime only",
			args: []string{"movetime", "5000"},
			expected: SearchParams{
				MoveTime: 5000 * time.Millisecond,
			},
		},
		{
			name: "infinite",
			args: []string{"infinite"},
			expected: SearchParams{
				Infinite: true,
			},
		},
		{
			name: "time control",
			args: []string{"wtime", "300000", "btime", "300000", "winc", "5000", "binc", "5000"},
			expected: SearchParams{
				WTime: 300000 * time.Millisecond,
				BTime: 300000 * time.Millisecond,
				WInc:  5000 * time.Millisecond,
				BInc:  5000 * time.Millisecond,
			},
		},
		{
			name: "moves to go",
			args: []string{"movestogo", "20"},
			expected: SearchParams{
				MovesToGo: 20,
			},
		},
		{
			name: "multiple parameters",
			args: []string{"depth", "8", "movetime", "10000", "wtime", "180000"},
			expected: SearchParams{
				Depth:    8,
				MoveTime: 10000 * time.Millisecond,
				WTime:    180000 * time.Millisecond,
			},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: SearchParams{},
		},
		{
			name:     "invalid depth value",
			args:     []string{"depth", "invalid"},
			expected: SearchParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.ParseGo(tt.args)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseGo() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProtocolHandler_ParseSetOption(t *testing.T) {
	handler := NewProtocolHandler()

	tests := []struct {
		name          string
		args          []string
		expectedName  string
		expectedValue string
		expectedErr   bool
	}{
		{
			name:          "simple option",
			args:          []string{"name", "Hash", "value", "256"},
			expectedName:  "Hash",
			expectedValue: "256",
		},
		{
			name:          "multi-word name",
			args:          []string{"name", "UCI_LimitStrength", "value", "true"},
			expectedName:  "UCI_LimitStrength",
			expectedValue: "true",
		},
		{
			name:          "multi-word value",
			args:          []string{"name", "SyzygyPath", "value", "/path/to/syzygy", "files"},
			expectedName:  "SyzygyPath",
			expectedValue: "/path/to/syzygy files",
		},
		{
			name:          "option without value",
			args:          []string{"name", "Clear", "Hash"},
			expectedName:  "Clear Hash",
			expectedValue: "",
		},
		{
			name:        "missing name",
			args:        []string{"value", "256"},
			expectedErr: true,
		},
		{
			name:        "empty args",
			args:        []string{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, value, err := handler.ParseSetOption(tt.args)

			if tt.expectedErr {
				if err == nil {
					t.Errorf("ParseSetOption() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSetOption() unexpected error: %v", err)
				return
			}

			if name != tt.expectedName {
				t.Errorf("ParseSetOption() name = %v, want %v", name, tt.expectedName)
			}

			if value != tt.expectedValue {
				t.Errorf("ParseSetOption() value = %v, want %v", value, tt.expectedValue)
			}
		})
	}
}

func TestProtocolHandler_FormatResponses(t *testing.T) {
	handler := NewProtocolHandler()

	t.Run("FormatUCIResponse", func(t *testing.T) {
		result := handler.FormatUCIResponse("TestEngine", "Test Author")
		expected := "id name TestEngine\nid author Test Author\nuciok"
		if result != expected {
			t.Errorf("FormatUCIResponse() = %v, want %v", result, expected)
		}
	})

	t.Run("FormatReadyOK", func(t *testing.T) {
		result := handler.FormatReadyOK()
		expected := "readyok"
		if result != expected {
			t.Errorf("FormatReadyOK() = %v, want %v", result, expected)
		}
	})

	t.Run("FormatBestMove", func(t *testing.T) {
		result := handler.FormatBestMove("e2e4")
		expected := "bestmove e2e4"
		if result != expected {
			t.Errorf("FormatBestMove() = %v, want %v", result, expected)
		}
	})

	t.Run("FormatInfo", func(t *testing.T) {
		result := handler.FormatInfo(6, 50, 1000000, 2*time.Second, "e2e4 e7e5")
		expected := "info depth 6 score cp 50 nodes 1000000 time 2000 nps 500000 pv e2e4 e7e5"
		if result != expected {
			t.Errorf("FormatInfo() = %v, want %v", result, expected)
		}
	})

	t.Run("FormatOption", func(t *testing.T) {
		result := handler.FormatOption("Hash", "spin", "128")
		expected := "option name Hash type spin default 128"
		if result != expected {
			t.Errorf("FormatOption() = %v, want %v", result, expected)
		}
	})
}
