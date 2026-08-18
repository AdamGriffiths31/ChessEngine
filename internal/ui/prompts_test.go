package ui

import (
	"testing"
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
