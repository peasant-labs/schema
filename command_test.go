package schema_test

import (
	"testing"

	"github.com/peasant-labs/schema"
)

func TestBuiltinCommand_IsValid(t *testing.T) {
	for _, c := range schema.AllClaudeBuiltinCmds {
		if !c.IsValid() {
			t.Errorf("AllClaudeBuiltinCmds[%s] should be valid", c)
		}
	}
	if schema.BuiltinCommand("unknown").IsValid() {
		t.Error("unknown command should be invalid")
	}
}

func TestIsClaudeBuiltinCommand(t *testing.T) {
	// With and without leading slash.
	if !schema.IsClaudeBuiltinCommand("exit") {
		t.Error("IsClaudeBuiltinCommand(\"exit\") should be true")
	}
	if !schema.IsClaudeBuiltinCommand("/exit") {
		t.Error("IsClaudeBuiltinCommand(\"/exit\") should be true")
	}
	// Skill names are not builtins.
	if schema.IsClaudeBuiltinCommand("aura:epoch") {
		t.Error("IsClaudeBuiltinCommand(\"aura:epoch\") should be false")
	}
	if schema.IsClaudeBuiltinCommand("/aura:epoch") {
		t.Error("IsClaudeBuiltinCommand(\"/aura:epoch\") should be false")
	}
}
