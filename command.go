package schema

import (
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// BuiltinCommand represents a built-in slash command supported by Claude Code.
// These are first-party commands that ship with the tool (e.g. /exit, /compact).
// Slash-command entries in transcripts whose name matches a BuiltinCommand are
// structural signals rather than user-initiated tool calls.
type BuiltinCommand string

const (
	ClaudeBuiltinCmdExit             BuiltinCommand = "exit"
	ClaudeBuiltinCmdCompact          BuiltinCommand = "compact"
	ClaudeBuiltinCmdClear            BuiltinCommand = "clear"
	ClaudeBuiltinCmdNew              BuiltinCommand = "new"
	ClaudeBuiltinCmdModel            BuiltinCommand = "model"
	ClaudeBuiltinCmdUsage            BuiltinCommand = "usage"
	ClaudeBuiltinCmdCost             BuiltinCommand = "cost"
	ClaudeBuiltinCmdContext          BuiltinCommand = "context"
	ClaudeBuiltinCmdPlugin           BuiltinCommand = "plugin"
	ClaudeBuiltinCmdPermissions      BuiltinCommand = "permissions"
	ClaudeBuiltinCmdLogin            BuiltinCommand = "login"
	ClaudeBuiltinCmdResume           BuiltinCommand = "resume"
	ClaudeBuiltinCmdPlan             BuiltinCommand = "plan"
	ClaudeBuiltinCmdFast             BuiltinCommand = "fast"
	ClaudeBuiltinCmdVoice            BuiltinCommand = "voice"
	ClaudeBuiltinCmdTodos            BuiltinCommand = "todos"
	ClaudeBuiltinCmdReloadPlugins    BuiltinCommand = "reload-plugins"
	ClaudeBuiltinCmdSandbox          BuiltinCommand = "sandbox"
	ClaudeBuiltinCmdConfig           BuiltinCommand = "config"
	ClaudeBuiltinCmdStatusline       BuiltinCommand = "statusline"
	ClaudeBuiltinCmdUpgrade          BuiltinCommand = "upgrade"
	ClaudeBuiltinCmdExtraUsage       BuiltinCommand = "extra-usage"
	ClaudeBuiltinCmdRateLimitOptions BuiltinCommand = "rate-limit-options"
	ClaudeBuiltinCmdPrivacySettings  BuiltinCommand = "privacy-settings"
	ClaudeBuiltinCmdHelp             BuiltinCommand = "help"
	ClaudeBuiltinCmdCommands         BuiltinCommand = "commands"
)

// IsValid returns true if the command is one of the known built-in variants.
// Derived from AllClaudeBuiltinCmds (single source of truth).
func (c BuiltinCommand) IsValid() bool {
	_, ok := builtinSet[c]
	return ok
}

// builtinSet is a lookup table derived from AllClaudeBuiltinCmds at init time.
var builtinSet map[BuiltinCommand]struct{}

func init() {
	builtinSet = make(map[BuiltinCommand]struct{}, len(AllClaudeBuiltinCmds))
	for _, cmd := range AllClaudeBuiltinCmds {
		builtinSet[cmd] = struct{}{}
	}
}

func (c BuiltinCommand) String() string { return string(c) }

// AllClaudeBuiltinCmds is the canonical list of all known Claude Code built-in commands.
var AllClaudeBuiltinCmds = [...]BuiltinCommand{
	ClaudeBuiltinCmdExit,
	ClaudeBuiltinCmdCompact,
	ClaudeBuiltinCmdClear,
	ClaudeBuiltinCmdNew,
	ClaudeBuiltinCmdModel,
	ClaudeBuiltinCmdUsage,
	ClaudeBuiltinCmdCost,
	ClaudeBuiltinCmdContext,
	ClaudeBuiltinCmdPlugin,
	ClaudeBuiltinCmdPermissions,
	ClaudeBuiltinCmdLogin,
	ClaudeBuiltinCmdResume,
	ClaudeBuiltinCmdPlan,
	ClaudeBuiltinCmdFast,
	ClaudeBuiltinCmdVoice,
	ClaudeBuiltinCmdTodos,
	ClaudeBuiltinCmdReloadPlugins,
	ClaudeBuiltinCmdSandbox,
	ClaudeBuiltinCmdConfig,
	ClaudeBuiltinCmdStatusline,
	ClaudeBuiltinCmdUpgrade,
	ClaudeBuiltinCmdExtraUsage,
	ClaudeBuiltinCmdRateLimitOptions,
	ClaudeBuiltinCmdPrivacySettings,
	ClaudeBuiltinCmdHelp,
	ClaudeBuiltinCmdCommands,
}

// JSONSchema implements jsonschema.Exposer.
func (BuiltinCommand) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema(
		"Built-in Command",
		"Claude Code built-in slash command name without the leading slash",
		AllClaudeBuiltinCmds[:],
	), nil
}

// IsClaudeBuiltinCommand reports whether name (with any leading slash stripped)
// matches a known Claude Code built-in command. Skill names like "aura:epoch" or
// "user-defined-command" will return false.
//
// Examples:
//
//	IsClaudeBuiltinCommand("exit")   → true
//	IsClaudeBuiltinCommand("/exit")  → true
//	IsClaudeBuiltinCommand("aura:epoch") → false
func IsClaudeBuiltinCommand(name string) bool {
	stripped := strings.TrimPrefix(name, "/")
	return BuiltinCommand(stripped).IsValid()
}
