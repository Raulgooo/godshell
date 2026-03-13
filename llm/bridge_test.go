package llm

import (
	"testing"
)

func TestTruncateString(t *testing.T) {
	s := "hello world"
	max := 5
	truncated := TruncateString(s, max)
	expected := "hello\n... [TRUNCATED DUE TO CONTEXT LIMITS] ..."
	if truncated != expected {
		t.Errorf("expected %q, got %q", expected, truncated)
	}

	short := "abc"
	if TruncateString(short, 10) != "abc" {
		t.Errorf("expected %q, got %q", "abc", TruncateString(short, 10))
	}
}

func TestManageHistory(t *testing.T) {
	systemPrompt := Message{Role: RoleSystem, Content: "System Prompt"}
	c := &Conversation{
		History:        []Message{systemPrompt},
		MaxTotalTokens: 10, // Very small limit for testing
	}

	// Add some messages
	c.History = append(c.History, Message{Role: RoleUser, Content: "1234567890"})      // 10 chars
	c.History = append(c.History, Message{Role: RoleAssistant, Content: "1234567890"}) // 10 chars
	c.History = append(c.History, Message{Role: RoleUser, Content: "1234567890"})      // 10 chars

	// Total chars: 13 + 10 + 10 + 10 = 43 chars
	// Max chars: 10 tokens * 4 = 40 chars

	c.ManageHistory()

	if len(c.History) != 3 {
		t.Errorf("expected history length 3, got %d", len(c.History))
	}

	if c.History[0].Content != "System Prompt" {
		t.Errorf("system prompt was lost or changed")
	}

	// Total chars should now be <= 40
	totalChars := 0
	for _, m := range c.History {
		totalChars += len(m.Content)
	}
	if totalChars > 40 {
		t.Errorf("total chars %d still above limit 40", totalChars)
	}
}
