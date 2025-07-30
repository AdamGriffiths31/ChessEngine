package uci

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// UCICommunicationLogger captures all UCI input/output for debugging
type UCICommunicationLogger struct {
	logFile  *os.File
	logger   *log.Logger
	mutex    sync.Mutex
	gameID   string
	moveNum  int
}

// NewUCICommunicationLogger creates a new UCI communication logger
func NewUCICommunicationLogger() *UCICommunicationLogger {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("/tmp/uci_communication_%s.log", timestamp)
	
	logFile, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create UCI communication log: %v", err)
		return nil
	}
	
	logger := log.New(logFile, "", log.LstdFlags|log.Lmicroseconds)
	
	commLogger := &UCICommunicationLogger{
		logFile: logFile,
		logger:  logger,
		gameID:  timestamp,
		moveNum: 0,
	}
	
	// Log session start
	commLogger.LogSessionStart()
	
	return commLogger
}

// LogSessionStart logs the beginning of a UCI session
func (ucl *UCICommunicationLogger) LogSessionStart() {
	if ucl == nil {
		return
	}
	
	ucl.mutex.Lock()
	defer ucl.mutex.Unlock()
	
	ucl.logger.Println("=====================================")
	ucl.logger.Println("UCI COMMUNICATION SESSION STARTED")
	ucl.logger.Printf("Game ID: %s", ucl.gameID)
	ucl.logger.Println("=====================================")
}

// LogIncoming logs incoming UCI commands from the GUI
func (ucl *UCICommunicationLogger) LogIncoming(command string) {
	if ucl == nil {
		return
	}
	
	ucl.mutex.Lock()
	defer ucl.mutex.Unlock()
	
	ucl.logger.Printf(">>> IN:  %s", command)
	
	// Track game state changes
	if strings.HasPrefix(command, "position") {
		ucl.LogPositionCommand(command)
	} else if strings.HasPrefix(command, "go") {
		ucl.moveNum++
		ucl.logger.Printf("    [MOVE %d STARTING]", ucl.moveNum)
	}
}

// LogOutgoing logs outgoing UCI responses to the GUI
func (ucl *UCICommunicationLogger) LogOutgoing(response string) {
	if ucl == nil {
		return
	}
	
	ucl.mutex.Lock()
	defer ucl.mutex.Unlock()
	
	ucl.logger.Printf("<<< OUT: %s", response)
	
	// Track move outputs
	if strings.HasPrefix(response, "bestmove") {
		ucl.LogBestMove(response)
	}
}

// LogPositionCommand analyzes and logs position commands in detail
func (ucl *UCICommunicationLogger) LogPositionCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) < 2 {
		return
	}
	
	if parts[1] == "startpos" {
		ucl.logger.Printf("    [POSITION: Starting position]")
		if len(parts) > 3 && parts[2] == "moves" {
			moves := parts[3:]
			ucl.logger.Printf("    [MOVES: %v (%d moves)]", moves, len(moves))
		}
	} else if parts[1] == "fen" {
		// Extract FEN
		fenEnd := 2
		for i := 2; i < len(parts) && parts[i] != "moves"; i++ {
			fenEnd = i + 1
		}
		
		fen := strings.Join(parts[2:fenEnd], " ")
		ucl.logger.Printf("    [POSITION: FEN = %s]", fen)
		
		if fenEnd < len(parts) && parts[fenEnd] == "moves" {
			moves := parts[fenEnd+1:]
			ucl.logger.Printf("    [MOVES: %v (%d moves)]", moves, len(moves))
		}
	}
}

// LogBestMove analyzes and logs bestmove responses
func (ucl *UCICommunicationLogger) LogBestMove(response string) {
	parts := strings.Fields(response)
	if len(parts) >= 2 {
		move := parts[1]
		ucl.logger.Printf("    [BESTMOVE %d: %s]", ucl.moveNum, move)
		
		// Flag potentially problematic moves
		if ucl.isPotentiallyIllegal(move) {
			ucl.logger.Printf("    [WARNING: Move %s might be problematic]", move)
		}
	}
}

// LogError logs errors in UCI communication
func (ucl *UCICommunicationLogger) LogError(context string, err error) {
	if ucl == nil {
		return
	}
	
	ucl.mutex.Lock()
	defer ucl.mutex.Unlock()
	
	ucl.logger.Printf("!!! ERROR in %s: %v", context, err)
}

// LogGameTermination logs when a game ends
func (ucl *UCICommunicationLogger) LogGameTermination(reason string) {
	if ucl == nil {
		return
	}
	
	ucl.mutex.Lock()
	defer ucl.mutex.Unlock()
	
	ucl.logger.Printf("*** GAME TERMINATED: %s ***", reason)
	ucl.logger.Printf("*** Total moves played: %d ***", ucl.moveNum)
}

// LogDebugInfo logs additional debug information
func (ucl *UCICommunicationLogger) LogDebugInfo(info string) {
	if ucl == nil {
		return
	}
	
	ucl.mutex.Lock()
	defer ucl.mutex.Unlock()
	
	ucl.logger.Printf("DEBUG: %s", info)
}

// Close closes the communication logger
func (ucl *UCICommunicationLogger) Close() {
	if ucl == nil {
		return
	}
	
	ucl.mutex.Lock()
	defer ucl.mutex.Unlock()
	
	if ucl.logFile != nil {
		ucl.logger.Println("=====================================")
		ucl.logger.Println("UCI COMMUNICATION SESSION ENDED")
		ucl.logger.Println("=====================================")
		ucl.logFile.Close()
	}
}

// isPotentiallyIllegal checks if a move might be problematic
func (ucl *UCICommunicationLogger) isPotentiallyIllegal(move string) bool {
	// Check for known problematic moves
	problematicMoves := []string{"d4e3", "f6f7"}
	
	for _, problem := range problematicMoves {
		if move == problem {
			return true
		}
	}
	
	return false
}

// WrapWriter wraps an io.Writer to log all output
func (ucl *UCICommunicationLogger) WrapWriter(original io.Writer) io.Writer {
	if ucl == nil {
		return original
	}
	
	return &LoggingWriter{
		original: original,
		logger:   ucl,
	}
}

// LoggingWriter wraps an io.Writer to capture output
type LoggingWriter struct {
	original io.Writer
	logger   *UCICommunicationLogger
}

// Write implements io.Writer interface
func (lw *LoggingWriter) Write(p []byte) (n int, err error) {
	// Log the output
	output := strings.TrimSpace(string(p))
	if output != "" {
		lw.logger.LogOutgoing(output)
	}
	
	// Write to original
	return lw.original.Write(p)
}