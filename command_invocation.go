package schema

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// CommandInvocation records that a turn invoked a skill or a user-defined
// slash command rather than typing a prompt. Name is slash-prefixed,
// for example "/superpowers:brainstorming". Args is the text that followed the
// command on the same line and may be empty. A harness that records a bare
// command name gets the slash added by the producer before construction; the
// constructor does not add it.
//
// Built-in harness commands (see BuiltinCommand) are structural signals and are
// never emitted as a CommandInvocation.
//
// Compatibility: the field is optional and safely ignorable. The invocation name
// is already present in the turn's content text, so a consumer that drops the
// field loses a structured duplicate, not meaning, provenance, or safety. Under
// docs/content-capability-negotiation.md that is an additive minor bump with no
// content-capability token.
type CommandInvocation struct {
	Name string `json:"name"`
	Args string `json:"args,omitempty"`
}

// NewCommandInvocation validates and constructs a CommandInvocation. Name must
// be a single slash-prefixed token that is not a built-in harness command. Args
// is carried as given.
func NewCommandInvocation(name, args string) (CommandInvocation, error) {
	const where = "command invocation validation failed at schema.NewCommandInvocation while recording a skill invocation: "
	if name == "" {
		return CommandInvocation{}, fmt.Errorf(where + "the name is empty, so the turn cannot be attributed to a command; omit the invocation or supply the slash-prefixed command name")
	}
	if !utf8.ValidString(name) {
		return CommandInvocation{}, fmt.Errorf(where+"name %q is not valid UTF-8, so it cannot be emitted on the wire; supply the harness-recorded name as valid UTF-8", name)
	}
	if name[0] != '/' {
		return CommandInvocation{}, fmt.Errorf(where+"name %q does not start with a slash, but the wire keeps the harness form; prefix the name with '/' at the producing boundary", name)
	}
	if len(name) == 1 {
		return CommandInvocation{}, fmt.Errorf(where + "the name is a bare slash and names no command; supply the command token after the slash")
	}
	for _, r := range name {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return CommandInvocation{}, fmt.Errorf(where+"name %q contains whitespace or a control character, so it cannot be a single command token; put trailing text in args", name)
		}
	}
	if IsClaudeBuiltinCommand(name) {
		return CommandInvocation{}, fmt.Errorf(where+"name %q is a built-in harness command, which is a structural signal and not a user command; do not emit it as a CommandInvocation", name)
	}
	return CommandInvocation{Name: name, Args: args}, nil
}

// Validate reports whether c would be accepted by NewCommandInvocation.
func (c CommandInvocation) Validate() error {
	_, err := NewCommandInvocation(c.Name, c.Args)
	return err
}
