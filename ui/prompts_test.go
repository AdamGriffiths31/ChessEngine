package ui

import (
	"testing"

	"github.com/AdamGriffiths31/ChessEngine/game"
)

func TestPrompterCreation(t *testing.T) {
	prompter := NewPrompter()
	if prompter == nil {
		t.Fatal("Expected prompter to be non-nil")
	}
	if prompter.scanner == nil {
		t.Fatal("Expected scanner to be non-nil")
	}
}

func TestPrompterMessages(t *testing.T) {
	prompter := NewPrompter()
	
	// Test that these methods don't panic (they just print to stdout)
	prompter.ShowWelcome()
	prompter.ShowMessage("Test message")
	prompter.ShowFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	prompter.ShowGoodbye()
}

func TestPrompterShowGameState(t *testing.T) {
	prompter := NewPrompter()
	engine := game.NewEngine()
	state := engine.GetState()
	
	// Test that ShowGameState doesn't panic
	prompter.ShowGameState(state)
}