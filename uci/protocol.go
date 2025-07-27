package uci

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Command represents a UCI command
type Command struct {
	Name string
	Args []string
}

// SearchParams represents parameters for the 'go' command
type SearchParams struct {
	Depth    int
	MoveTime time.Duration
	Infinite bool
	WTime    time.Duration // White's remaining time
	BTime    time.Duration // Black's remaining time
	WInc     time.Duration // White's increment per move
	BInc     time.Duration // Black's increment per move
	MovesToGo int          // Moves to next time control
}

// ProtocolHandler handles UCI protocol parsing and response formatting
type ProtocolHandler struct{}

// NewProtocolHandler creates a new protocol handler
func NewProtocolHandler() *ProtocolHandler {
	return &ProtocolHandler{}
}

// ParseCommand parses a UCI command string into a Command struct
func (ph *ProtocolHandler) ParseCommand(input string) Command {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return Command{Name: "", Args: []string{}}
	}
	
	return Command{
		Name: parts[0],
		Args: parts[1:],
	}
}

// ParsePosition parses a 'position' command
func (ph *ProtocolHandler) ParsePosition(args []string) (fen string, moves []string, err error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("position command requires arguments")
	}
	
	if args[0] == "startpos" {
		fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
		args = args[1:]
	} else if args[0] == "fen" {
		if len(args) < 7 {
			return "", nil, fmt.Errorf("fen position requires 6 parts")
		}
		// Reconstruct FEN from parts
		fen = strings.Join(args[1:7], " ")
		args = args[7:]
	} else {
		return "", nil, fmt.Errorf("position must start with 'startpos' or 'fen'")
	}
	
	// Check for moves
	if len(args) > 0 && args[0] == "moves" {
		moves = args[1:]
	}
	
	return fen, moves, nil
}

// ParseGo parses a 'go' command into SearchParams
func (ph *ProtocolHandler) ParseGo(args []string) SearchParams {
	params := SearchParams{}
	
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "depth":
			if i+1 < len(args) {
				if depth, err := strconv.Atoi(args[i+1]); err == nil {
					params.Depth = depth
				}
				i++
			}
		case "movetime":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					params.MoveTime = time.Duration(ms) * time.Millisecond
				}
				i++
			}
		case "infinite":
			params.Infinite = true
		case "wtime":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					params.WTime = time.Duration(ms) * time.Millisecond
				}
				i++
			}
		case "btime":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					params.BTime = time.Duration(ms) * time.Millisecond
				}
				i++
			}
		case "winc":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					params.WInc = time.Duration(ms) * time.Millisecond
				}
				i++
			}
		case "binc":
			if i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					params.BInc = time.Duration(ms) * time.Millisecond
				}
				i++
			}
		case "movestogo":
			if i+1 < len(args) {
				if moves, err := strconv.Atoi(args[i+1]); err == nil {
					params.MovesToGo = moves
				}
				i++
			}
		}
	}
	
	return params
}

// FormatUCIResponse formats the initial UCI response
func (ph *ProtocolHandler) FormatUCIResponse(engineName, author string) string {
	return fmt.Sprintf("id name %s\nid author %s\nuciok", engineName, author)
}

// FormatReadyOK formats the readyok response
func (ph *ProtocolHandler) FormatReadyOK() string {
	return "readyok"
}

// FormatBestMove formats the bestmove response
func (ph *ProtocolHandler) FormatBestMove(move string) string {
	return fmt.Sprintf("bestmove %s", move)
}

// FormatInfo formats search info during thinking
func (ph *ProtocolHandler) FormatInfo(depth int, score int, nodes int64, time time.Duration, pv string) string {
	timeMs := int(time.Milliseconds())
	nps := int64(0)
	if timeMs > 0 {
		nps = (nodes * 1000) / int64(timeMs)
	}
	
	info := fmt.Sprintf("info depth %d score cp %d nodes %d time %d nps %d",
		depth, score, nodes, timeMs, nps)
	
	if pv != "" {
		info += fmt.Sprintf(" pv %s", pv)
	}
	
	return info
}

// FormatOption formats a UCI option declaration
func (ph *ProtocolHandler) FormatOption(name, optionType, defaultValue string) string {
	return fmt.Sprintf("option name %s type %s default %s", name, optionType, defaultValue)
}

// FormatSetOption parses a setoption command
func (ph *ProtocolHandler) ParseSetOption(args []string) (name, value string, err error) {
	// Format: setoption name <name> value <value>
	nameIndex := -1
	valueIndex := -1
	
	for i, arg := range args {
		if arg == "name" && i+1 < len(args) {
			nameIndex = i + 1
		}
		if arg == "value" && i+1 < len(args) {
			valueIndex = i + 1
		}
	}
	
	if nameIndex == -1 {
		return "", "", fmt.Errorf("setoption command missing name")
	}
	
	// Collect name parts (until "value" or end)
	var nameParts []string
	for i := nameIndex; i < len(args) && args[i] != "value"; i++ {
		nameParts = append(nameParts, args[i])
	}
	name = strings.Join(nameParts, " ")
	
	// Collect value parts (from valueIndex to end)
	if valueIndex != -1 {
		var valueParts []string
		for i := valueIndex; i < len(args); i++ {
			valueParts = append(valueParts, args[i])
		}
		value = strings.Join(valueParts, " ")
	}
	
	return name, value, nil
}