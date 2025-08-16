// Package epd provides functionality for parsing and handling Extended Position Description (EPD) format.
package epd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/board"
)

// MoveScore represents a move and its point value
type MoveScore struct {
	Move   string
	Points int
}

// Position represents a single EPD position with annotations
type Position struct {
	Board      *board.Board
	BestMove   string      // Best move in algebraic notation (e.g., "Nf3", "e1g1")
	AvoidMove  string      // Move to avoid (if any)
	Comment    string      // Position comment/description
	ID         string      // Position ID
	MoveScores []MoveScore // All moves with their point values from c0 annotation
}

// ParseEPD parses a single EPD line and returns a Position
func ParseEPD(epdLine string) (*Position, error) {
	epdLine = strings.TrimSpace(epdLine)
	if epdLine == "" || strings.HasPrefix(epdLine, "#") {
		return nil, fmt.Errorf("empty or comment line")
	}

	// EPD format can be:
	// 1. position side castling enpassant [annotations] (STS format)
	// 2. position side castling enpassant halfmove fullmove [annotations] (full FEN format)
	parts := strings.Fields(epdLine)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid EPD format: insufficient fields")
	}

	var fenParts []string
	var annotationStart int

	// Check if parts[4] and parts[5] look like halfmove/fullmove counters
	if len(parts) >= 6 {
		// Check if parts[4] and parts[5] are numeric (halfmove/fullmove)
		if isNumeric(parts[4]) && isNumeric(parts[5]) {
			// Full FEN format: position side castling enpassant halfmove fullmove
			fenParts = parts[:6]
			annotationStart = 6
		} else {
			// EPD format without counters: position side castling enpassant
			fenParts = parts[:4]
			annotationStart = 4
		}
	} else {
		// Not enough parts for full FEN, assume EPD format
		fenParts = parts[:4]
		annotationStart = 4
	}

	// If we only have 4 parts (no halfmove/fullmove), add defaults
	var fen string
	if len(fenParts) == 4 {
		fen = strings.Join(fenParts, " ") + " 0 1"
	} else {
		fen = strings.Join(fenParts, " ")
	}

	// Parse board from FEN
	board, err := board.FromFEN(fen)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FEN: %w", err)
	}

	position := &Position{
		Board: board,
	}

	// The remaining parts are annotations
	if len(parts) > annotationStart {
		annotationParts := parts[annotationStart:]
		annotationStr := strings.Join(annotationParts, " ")

		// Parse the annotations
		position = parseAnnotations(position, annotationStr)
	}

	return position, nil
}

// parseAnnotations handles the annotation parsing separately
func parseAnnotations(position *Position, annotationStr string) *Position {
	// Split by semicolon to handle EPD annotations properly
	annotations := strings.Split(annotationStr, ";")

	for _, annotation := range annotations {
		annotation = strings.TrimSpace(annotation)
		if annotation == "" {
			continue
		}

		// Split each annotation into parts
		annotParts := strings.Fields(annotation)
		if len(annotParts) < 2 {
			continue
		}

		switch annotParts[0] {
		case "bm":
			// Best move annotation: bm Ba7
			position.BestMove = annotParts[1]

		case "am":
			// Avoid move annotation: am Nf3
			position.AvoidMove = annotParts[1]

		case "c0":
			// Comment annotation: c0 "Ba7=10, Qf6+=3, a5=3, h5=5"
			comment := strings.Join(annotParts[1:], " ")
			// Remove quotes if present
			comment = strings.Trim(comment, "\"")
			position.Comment = comment

			// Parse move scores from the comment
			position.MoveScores = parseMoveScores(comment)

		case "id":
			// ID annotation: id "STS(v2.2) Open Files and Diagonals.001"
			id := strings.Join(annotParts[1:], " ")
			// Remove quotes if present
			id = strings.Trim(id, "\"")
			position.ID = id
		}
	}

	return position
}

// ParseEPDFile parses an entire EPD file and returns a slice of Position structs
func ParseEPDFile(content string) ([]*Position, error) {
	lines := strings.Split(content, "\n")
	positions := make([]*Position, 0, len(lines))

	for lineNum, line := range lines {
		position, err := ParseEPD(line)
		if err != nil {
			// Skip empty lines and comments, but report other parsing errors
			if !strings.Contains(err.Error(), "empty or comment line") {
				return nil, fmt.Errorf("error parsing line %d: %w", lineNum+1, err)
			}
			continue
		}
		positions = append(positions, position)
	}

	return positions, nil
}

// String returns a string representation of the EPD position
func (pos *Position) String() string {
	var parts []string

	if pos.ID != "" {
		parts = append(parts, fmt.Sprintf("ID: %s", pos.ID))
	}
	if pos.Comment != "" {
		parts = append(parts, fmt.Sprintf("Comment: %s", pos.Comment))
	}
	if pos.BestMove != "" {
		parts = append(parts, fmt.Sprintf("Best Move: %s", pos.BestMove))
	}
	if pos.AvoidMove != "" {
		parts = append(parts, fmt.Sprintf("Avoid Move: %s", pos.AvoidMove))
	}

	if len(parts) == 0 {
		return "EPD Position (no annotations)"
	}

	return strings.Join(parts, ", ")
}

// parseMoveScores parses move scores from STS c0 comments like "Ba7=10, Qf6+=3, a5=3, h5=5"
func parseMoveScores(comment string) []MoveScore {
	var moveScores []MoveScore

	// Pattern to match move=score pairs like "Ba7=10" or "Qf6+=3"
	re := regexp.MustCompile(`([A-Za-z0-9+=-]+)=(\d+)`)
	matches := re.FindAllStringSubmatch(comment, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			move := strings.TrimSuffix(match[1], "+") // Remove + suffix if present
			if points, err := strconv.Atoi(match[2]); err == nil {
				moveScores = append(moveScores, MoveScore{
					Move:   move,
					Points: points,
				})
			}
		}
	}

	return moveScores
}

// isNumeric checks if a string represents a number
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
