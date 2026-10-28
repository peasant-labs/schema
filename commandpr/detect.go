// Package commandpr detects pull requests created by a recorded shell
// command. It reads stored tool calls; it never contacts GitHub and never
// decides who may read a transcript.
//
// Detection is display-only. A detected pull request is named in the local
// transcript beside the session that produced it; it widens no access and
// posts nothing.
package commandpr

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/peasant-labs/schema"
	"mvdan.cc/sh/v3/syntax"
)

// DetectSession returns first-seen recorded PR evidence with exact call coordinates.
func DetectSession(detail schema.SessionDetailPayload) []schema.UnverifiedPullRequest {
	detected := make([]schema.UnverifiedPullRequest, 0)
	seen := make(map[string]struct{})
	for _, turn := range detail.Turns {
		for index, call := range turn.ToolCalls {
			if !isShellCall(call) || call.IsError || (call.ExitCode != nil && *call.ExitCode != 0) {
				continue
			}
			command := commandFromArguments(call.Arguments)
			if command == "" || !invokesGHPRCreate(command) {
				continue
			}
			ref, ok := lastPullRequestReference(call.Result)
			if !ok {
				continue
			}
			key := strings.ToLower(ref.Owner) + "/" + strings.ToLower(ref.Repo) + "/" + strconv.Itoa(ref.Number)
			if _, ok := seen[key]; ok {
				continue
			}
			row := schema.UnverifiedPullRequest{Owner: ref.Owner, Repo: ref.Repo, Number: ref.Number, URL: ref.URL, Source: schema.UnverifiedPullRequestSource{TurnIndex: turn.Index, ToolCallIndex: index, ToolCallID: call.ID}}
			if row.Validate() != nil {
				continue
			}
			seen[key] = struct{}{}
			detected = append(detected, row)
		}
	}
	return detected
}

// isShellCall requires recorded execution identity; arguments alone are not evidence.
func isShellCall(call schema.ToolCallDetail) bool {
	name := strings.ToLower(strings.TrimSpace(call.Name))
	switch name {
	case "read", "read_file", "readfile", "write", "write_file", "edit", "grep", "glob", "search", "fetch", "webfetch":
		return false
	}
	if call.ToolKind != "" {
		return call.ToolKind == schema.ToolCallKindExecute
	}
	switch name {
	case "bash", "shell", "terminal", "execute", "run", "command", "exec_command", "exec", "runterminalcommand", "run_terminal_cmd", "runcommand", "run_command", "executecommand":
		return true
	}
	return false
}

// commandFromArguments extracts the shell command from a stored tool call's
// arguments. Harnesses record arguments either as a JSON object with a
// "command" key (some use "cmd") or, for a harness that stores the invocation
// directly, as the command text itself. Empty input carries no command.
func commandFromArguments(arguments string) string {
	trimmed := strings.TrimSpace(arguments)
	if trimmed == "" {
		return ""
	}
	if trimmed[0] == '{' {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(trimmed), &fields); err == nil {
			for _, key := range []string{"command", "cmd"} {
				raw, ok := fields[key]
				if !ok {
					continue
				}
				var value string
				if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
					return value
				}
			}
			// A JSON object without a command field describes a tool that did
			// not run a shell command.
			return ""
		}
	}
	return arguments
}

// maxWrapperDepth bounds how many nested `-c` re-parses detection follows
// before it gives up. The bound counts only the shell scripts re-parsed; the
// `sudo`, `command`, `env`, and `nohup` wrappers around them cost no depth
// because they are read from the same parse. Four levels cover the shapes a
// recorded command takes (`bash -c 'bash -c "…"'` and friends); a deeper
// command is not a shell script a person writes, and the bound keeps a
// pathological input finite.
const maxWrapperDepth = 4

// invokesGHPRCreate reports whether the command runs the `gh pr create`
// subcommand, directly or through one of the shell wrappers a person may put
// in front of it.
//
// The command is parsed as shell rather than searched as text, so an
// invocation inside a heredoc body or a quoted argument is not a call. A
// leading `cd … &&`, an environment assignment (`GH_TOKEN=x gh pr create`),
// and a pipeline all keep working because the walk visits every simple
// command in the syntax tree.
func invokesGHPRCreate(command string) bool {
	return commandInvokesGHPRCreate(command, maxWrapperDepth)
}

// commandInvokesGHPRCreate walks one parsed shell program for a simple command
// that runs `gh pr create`. depth is the recursion budget: each shell wrapper
// whose script is re-parsed spends one level, so a script that nests wrappers
// terminates.
func commandInvokesGHPRCreate(command string, depth int) bool {
	if depth < 0 {
		return false
	}
	file, err := syntax.NewParser().Parse(strings.NewReader(command), "")
	if err != nil {
		return false
	}
	found := false
	var visit func(syntax.Node) bool
	visit = func(node syntax.Node) bool {
		if found {
			return false
		}
		switch n := node.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.IfClause:
			if len(n.Cond) == 1 {
				switch staticTruth(n.Cond[0]) {
				case -1:
					if n.Else != nil {
						syntax.Walk(n.Else, visit)
					}
					return false
				case 1:
					for _, stmt := range n.Then {
						syntax.Walk(stmt, visit)
					}
					return false
				}
			}
		case *syntax.ForClause:
			if iter, ok := n.Loop.(*syntax.WordIter); ok && iter.InPos.IsValid() && len(iter.Items) == 0 {
				return false
			}
		case *syntax.CaseClause:
			word := n.Word.Lit()
			if word == "" {
				return false
			}
			for _, item := range n.Items {
				for _, pattern := range item.Patterns {
					lit := pattern.Lit()
					if lit == "" || strings.ContainsAny(lit, "*?[\\") {
						return false
					}
					if lit == word {
						for _, stmt := range item.Stmts {
							syntax.Walk(stmt, visit)
						}
						return false
					}
				}
			}
			return false
		case *syntax.WhileClause:
			if len(n.Cond) == 1 && ((!n.Until && staticTruth(n.Cond[0]) == -1) || (n.Until && staticTruth(n.Cond[0]) == 1)) {
				syntax.Walk(n.Cond[0], visit)
				return false
			}
		case *syntax.BinaryCmd:
			truth := staticTruth(n.X)
			if (n.Op == syntax.AndStmt && truth == -1) || (n.Op == syntax.OrStmt && truth == 1) {
				syntax.Walk(n.X, visit)
				return false
			}
		case *syntax.CallExpr:
			if wordsInvoke(commandWords(n), depth) {
				found = true
				return false
			}
		}
		return true
	}
	syntax.Walk(file, visit)
	return found
}

// staticTruth infers only literal statuses and bounded boolean composition.
// Unknown commands remain unknown; no recorded command is executed here.
func staticTruth(stmt *syntax.Stmt) int { return staticTruthBounded(stmt, 64) }

func staticTruthBounded(stmt *syntax.Stmt, budget int) int {
	if stmt == nil || budget <= 0 {
		return 0
	}
	truth := 0
	switch cmd := stmt.Cmd.(type) {
	case *syntax.CallExpr:
		if len(cmd.Assigns) == 0 && len(cmd.Args) == 1 {
			switch cmd.Args[0].Lit() {
			case "true":
				truth = 1
			case "false":
				truth = -1
			}
		}
	case *syntax.Subshell:
		if len(cmd.Stmts) == 1 {
			truth = staticTruthBounded(cmd.Stmts[0], budget-1)
		}
	case *syntax.BinaryCmd:
		left := staticTruthBounded(cmd.X, budget-1)
		switch cmd.Op {
		case syntax.AndStmt:
			if left == -1 {
				truth = -1
			} else {
				right := staticTruthBounded(cmd.Y, budget-1)
				if left == 1 {
					truth = right
				} else if right == -1 {
					truth = -1
				}
			}
		case syntax.OrStmt:
			if left == 1 {
				truth = 1
			} else {
				right := staticTruthBounded(cmd.Y, budget-1)
				if left == -1 {
					truth = right
				} else if right == 1 {
					truth = 1
				}
			}
		}
	}
	if stmt.Negated {
		truth = -truth
	}
	return truth
}

// commandWord is one word of a simple command, kept in both literal forms the
// detector needs.
type commandWord struct {
	// value is syntax.Word.Lit: the word's text when the whole word is one
	// unquoted literal part. A quoted, expanded, or composite word has no
	// value, and no command, wrapper, flag, or option name matches an empty
	// value, so such a word fails closed.
	value string
	// text is the word's text when the whole word is a plain literal, including
	// one wrapped in quotes, and literal reports whether it is. Only a shell's
	// -c script reads text: the script a person records is normally quoted, and
	// Word.Lit does not unwrap quotes.
	text    string
	literal bool
}

// commandWords returns one simple command's words.
func commandWords(call *syntax.CallExpr) []commandWord {
	words := make([]commandWord, len(call.Args))
	for index, arg := range call.Args {
		words[index] = commandWord{value: arg.Lit()}
		words[index].text, words[index].literal = literalWordText(arg)
	}
	return words
}

// literalWordText returns the text a word expands to when no part of it can
// change at run time. An expansion, a substitution, or an ANSI-C quoted word is
// not literal, and false tells the caller to reject it.
//
// The text is the source text a shell has not unescaped: the script of
// `bash -c "echo \"hi\""` reads back with its backslashes still in place. A
// caller that hands the text to a shell again must decide what to do about
// that; shellCommandScript refuses such a script outright.
func literalWordText(word *syntax.Word) (string, bool) {
	var builder strings.Builder
	for _, part := range word.Parts {
		switch part := part.(type) {
		case *syntax.Lit:
			builder.WriteString(part.Value)
		case *syntax.SglQuoted:
			if part.Dollar {
				return "", false
			}
			builder.WriteString(part.Value)
		case *syntax.DblQuoted:
			for _, inner := range part.Parts {
				literal, ok := inner.(*syntax.Lit)
				if !ok {
					return "", false
				}
				builder.WriteString(literal.Value)
			}
		default:
			return "", false
		}
	}
	return builder.String(), true
}

// ghPRCreateWords are the words that open a direct invocation, after the
// command name.
var ghPRCreateWords = [...]string{"pr", "create"}

// ghCommandNames are the spellings of the GitHub CLI. The Windows executable
// name is accepted because a recorded command may come from either platform.
var ghCommandNames = [...]string{"gh", "gh.exe"}

// shellWrappers are the shells whose `-c` script is itself a shell program
// that detection re-parses.
var shellWrappers = [...]string{"bash", "sh", "zsh", "dash", "ksh"}

// wordsInvoke reports whether one simple command's words run `gh pr create`,
// unwrapping the closed set of shell wrappers detection understands: `sudo`,
// `command`, `env`, `nohup`, and a shell with a `-c` script.
//
// A wrapper's own options are skipped only where their arity is known, so a
// word this code does not recognize becomes the wrapped command and is
// rejected. That is what keeps `echo "gh pr create"` and
// `bash -c 'echo gh pr create'` false: the words after the wrapper are not the
// gh command, and no text search ever runs.
//
// depth is always non-negative here: commandInvokesGHPRCreate rejects a spent
// budget before it walks, and only that function decrements.
func wordsInvoke(words []commandWord, depth int) bool {
	if len(words) == 0 {
		return false
	}
	for _, name := range ghCommandNames {
		if words[0].value != name {
			continue
		}
		if len(words) < 1+len(ghPRCreateWords) {
			return false
		}
		for index, want := range ghPRCreateWords {
			if words[1+index].value != want {
				return false
			}
		}
		return true
	}
	switch words[0].value {
	case "sudo":
		return wordsInvoke(skipWrapperOptions(words[1:], sudoFlags, sudoValueFlags), depth)
	case "command":
		return wordsInvoke(skipWrapperOptions(words[1:], commandFlags, nil), depth)
	case "env":
		rest := skipWrapperOptions(words[1:], envFlags, envValueFlags)
		for len(rest) > 0 && isEnvironmentAssignment(rest[0]) {
			rest = rest[1:]
		}
		return wordsInvoke(rest, depth)
	case "nohup":
		return wordsInvoke(skipWrapperOptions(words[1:], nohupFlags, nil), depth)
	default:
		for _, name := range shellWrappers {
			if words[0].value != name {
				continue
			}
			script, ok := shellCommandScript(words[1:])
			if !ok {
				return false
			}
			return commandInvokesGHPRCreate(script, depth-1)
		}
	}
	return false
}

// Wrapper options that take no value, and the options that consume the next
// word. Each set is deliberately closed: an option this code does not
// recognize is treated as the wrapped command, so a wrapper form that was not
// considered fails to detect rather than guessing. A bare `--` is not listed
// because skipWrapperOptions handles it for every wrapper.
//
// A flag belongs here only when the wrapper still runs the command after
// skipping it. `command -v`/`-V` and `nohup --help`/`--version` report on the
// command instead of running it, so they are absent and those words become the
// wrapped command, which fails the check.
var (
	sudoFlags      = stringSet("-E", "-H", "-n", "-S", "-b")
	sudoValueFlags = stringSet("-u", "-g", "-h", "-p", "-C", "-T", "-r", "-t")
	commandFlags   = stringSet("-p")
	envFlags       = stringSet("-i", "-0")
	envValueFlags  = stringSet("-u", "-C", "-S")
	nohupFlags     = stringSet()
)

// stringSet builds a lookup set from a closed list of words.
func stringSet(words ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(words))
	for _, word := range words {
		set[word] = struct{}{}
	}
	return set
}

// skipWrapperOptions drops a wrapper's leading options and returns the words
// that remain: the wrapped command and its arguments. valueFlags names the
// options that consume the following word. A bare `--` ends the options. A
// word with no literal value is returned unchanged so the caller's command
// check rejects it.
func skipWrapperOptions(args []commandWord, flags, valueFlags map[string]struct{}) []commandWord {
	for len(args) > 0 {
		word := args[0].value
		switch {
		case word == "":
			return args
		case word == "--":
			return args[1:]
		case hasWord(valueFlags, word):
			if len(args) < 2 {
				return nil
			}
			args = args[2:]
		case hasWord(flags, word):
			args = args[1:]
		default:
			return args
		}
	}
	return nil
}

// hasWord reports whether a closed set contains word.
func hasWord(set map[string]struct{}, word string) bool {
	_, ok := set[word]
	return ok
}

// isEnvironmentAssignment reports whether an env argument is the `NAME=VALUE`
// form env accepts before the command it runs.
func isEnvironmentAssignment(word commandWord) bool {
	name, _, ok := strings.Cut(word.value, "=")
	return ok && name != ""
}

// shellValueFlags are the short shell options that consume the following word,
// so `bash -o pipefail -c '…'` still finds its script.
var shellValueFlags = stringSet("-o", "-O")

// shellCommandScript returns the script a shell runs for its `-c` option. The
// option may be written alone (`-c`) or in a combined flag word that contains
// `c` (`-lc`, `-xec`). A shell followed by a script file instead of `-c` runs
// no inline command and reports no script, and so does a script that is not a
// literal word.
//
// A literal script that still contains a backslash also reports no script.
// The stored text comes from the syntax tree before a shell applies its own
// quoting rules, so `bash -c "echo \"; gh pr create; echo \""` reads as
// `echo \"; gh pr create; echo \"`. Re-parsing that text would treat the
// escaped quote as a real quote and turn an echoed string into an invocation,
// and unescaping it here would mean re-implementing the shell's rules at every
// quoting form. Escaped scripts are therefore deliberately not unwrapped: the
// call detects nothing rather than guessing. The cost is accepted knowingly: a
// script that would really run the command and happens to keep a backslash is
// missed, which is cheaper than reporting a pull request nobody created.
func shellCommandScript(args []commandWord) (string, bool) {
	for index := 0; index < len(args); index++ {
		word := args[index].value
		switch {
		case word == "":
			return "", false
		case word == "--":
			return "", false
		case strings.HasPrefix(word, "--"):
			continue
		case len(word) > 1 && word[0] == '-':
			if hasWord(shellValueFlags, word) {
				index++
				continue
			}
			if !strings.Contains(word[1:], "c") {
				continue
			}
			if index+1 >= len(args) || !args[index+1].literal {
				return "", false
			}
			script := args[index+1].text
			if strings.ContainsRune(script, '\\') {
				return "", false
			}
			return script, true
		default:
			// A non-flag word before -c is a script file the shell reads; the
			// words after it are that script's arguments, not a shell command.
			return "", false
		}
	}
	return "", false
}

// pullRequestReference is the identity parsed out of one GitHub URL.
type pullRequestReference struct {
	Owner  string
	Repo   string
	Number int
	URL    string
}

// Scan entire URL tokens, so URLs inside a query or fragment are never new candidates.
var resultURLToken = regexp.MustCompile(`[A-Za-z][A-Za-z0-9+.-]*:[^\s<>"']+`)

func lastPullRequestReference(result string) (pullRequestReference, bool) {
	tokens := resultURLToken.FindAllStringIndex(result, -1)
	for i := len(tokens) - 1; i >= 0; i-- {
		m := tokens[i]
		if m[0] > 0 && strings.ContainsAny(result[m[0]-1:m[0]], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_@./:-") {
			continue
		}
		token := strings.TrimRight(result[m[0]:m[1]], ".,;!)]}")
		parsed, err := url.Parse(token)
		if err != nil || parsed.Scheme != "https" || parsed.Host != "github.com" || parsed.User != nil || parsed.Opaque != "" {
			continue
		}
		parts := strings.Split(strings.TrimPrefix(parsed.EscapedPath(), "/"), "/")
		if len(parts) < 4 || parts[2] != "pull" || (len(parts) > 4 && (len(parts) != 5 || !validPullSubpage(parts[4]))) {
			continue
		}
		digits := parts[3]
		if digits == "" || strings.Trim(digits, "0123456789") != "" {
			continue
		}
		number, err := strconv.Atoi(digits)
		if err != nil || number < 1 {
			continue
		}
		ref := pullRequestReference{Owner: parts[0], Repo: parts[1], Number: number, URL: "https://github.com/" + parts[0] + "/" + parts[1] + "/pull/" + strconv.Itoa(number)}
		if (schema.UnverifiedPullRequest{Owner: ref.Owner, Repo: ref.Repo, Number: ref.Number, URL: ref.URL}).Validate() != nil {
			continue
		}
		return ref, true
	}
	return pullRequestReference{}, false
}

func validPullSubpage(part string) bool {
	switch part {
	case "files", "commits", "checks", "conversation":
		return true
	}
	return false
}
