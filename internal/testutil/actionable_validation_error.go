package testutil

import (
	"fmt"
	"strings"
	"testing"
)

type ActionableValidationErrorDimension string

const (
	ActionableValidationErrorWhat        ActionableValidationErrorDimension = "what"
	ActionableValidationErrorWhy         ActionableValidationErrorDimension = "why"
	ActionableValidationErrorWhere       ActionableValidationErrorDimension = "where"
	ActionableValidationErrorWhen        ActionableValidationErrorDimension = "when"
	ActionableValidationErrorMeaning     ActionableValidationErrorDimension = "meaning"
	ActionableValidationErrorRemediation ActionableValidationErrorDimension = "remediation"
)

type ActionableValidationErrorParts struct {
	What        string
	Why         string
	Where       string
	When        string
	Meaning     string
	Remediation string
}

func (p ActionableValidationErrorParts) Fragments() []string {
	return []string{p.What, p.Why, p.Where, p.When, p.Meaning, p.Remediation}
}

func (p ActionableValidationErrorParts) Fragment(d ActionableValidationErrorDimension) string {
	switch d {
	case ActionableValidationErrorWhat:
		return p.What
	case ActionableValidationErrorWhy:
		return p.Why
	case ActionableValidationErrorWhere:
		return p.Where
	case ActionableValidationErrorWhen:
		return p.When
	case ActionableValidationErrorMeaning:
		return p.Meaning
	case ActionableValidationErrorRemediation:
		return p.Remediation
	default:
		return ""
	}
}

func ParseActionableValidationError(msg string) (ActionableValidationErrorParts, error) {
	const (
		whereMarker = " at schema."
		whenMarker  = " during wire-boundary validation"
	)

	what, afterWhere, ok := strings.Cut(msg, whereMarker)
	if !ok {
		return ActionableValidationErrorParts{}, fmt.Errorf("missing actionable validation where marker %q", whereMarker)
	}
	where, afterWhen, ok := strings.Cut(afterWhere, whenMarker)
	if !ok {
		return ActionableValidationErrorParts{}, fmt.Errorf("missing actionable validation when marker %q", whenMarker)
	}
	afterWhen = strings.TrimSpace(afterWhen)
	if !strings.HasPrefix(afterWhen, ": ") {
		return ActionableValidationErrorParts{}, fmt.Errorf("missing actionable validation cause clause")
	}
	causeAndActions := strings.TrimSpace(strings.TrimPrefix(afterWhen, ": "))
	why, actionClause, ok := strings.Cut(causeAndActions, ";")
	if !ok {
		return ActionableValidationErrorParts{}, fmt.Errorf("missing actionable validation follow-up clause")
	}
	meaning := strings.TrimSpace(actionClause)
	remediation := ""
	if before, after, ok := strings.Cut(meaning, ";"); ok {
		meaning = strings.TrimSpace(before)
		remediation = strings.TrimSpace(after)
	}
	parts := ActionableValidationErrorParts{
		What:        strings.TrimSpace(what),
		Why:         strings.TrimSpace(why),
		Where:       strings.TrimSpace(where),
		When:        strings.TrimSpace(whenMarker),
		Meaning:     strings.TrimSpace(meaning),
		Remediation: strings.TrimSpace(remediation),
	}
	if parts.What == "" || parts.Why == "" || parts.Where == "" || parts.When == "" || parts.Meaning == "" {
		return ActionableValidationErrorParts{}, fmt.Errorf("incomplete actionable validation error")
	}
	if parts.Remediation == "" {
		return ActionableValidationErrorParts{}, fmt.Errorf("missing actionable validation remediation clause")
	}
	return parts, nil
}

func StripActionableValidationDimension(msg string, d ActionableValidationErrorDimension) (string, error) {
	parts, err := ParseActionableValidationError(msg)
	if err != nil {
		return "", err
	}
	switch d {
	case ActionableValidationErrorWhat:
		return strings.Replace(msg, parts.What, "opaque validation", 1), nil
	case ActionableValidationErrorWhy:
		return strings.Replace(msg, parts.Why, "opaque reason", 1), nil
	case ActionableValidationErrorWhere:
		return strings.Replace(msg, parts.Where, "Opaque.Validate", 1), nil
	case ActionableValidationErrorWhen:
		return strings.Replace(msg, parts.When, "during opaque validation", 1), nil
	case ActionableValidationErrorMeaning:
		return strings.Replace(msg, parts.Meaning, "opaque caller impact", 1), nil
	case ActionableValidationErrorRemediation:
		if parts.Remediation == "" {
			return "", fmt.Errorf("baseline actionable validation error has no remediation clause to strip")
		}
		remediationClause := "; " + parts.Remediation
		if idx := strings.LastIndex(msg, remediationClause); idx >= 0 {
			return strings.TrimSpace(msg[:idx]), nil
		}
		return "", fmt.Errorf("baseline actionable validation error has no remediation clause to strip")
	default:
		return "", fmt.Errorf("unknown actionable validation dimension %q", d)
	}
}

func ActionableValidationErrorViolations(err error, wantContains ...string) []string {
	if err == nil {
		return []string{"error"}
	}
	msg := err.Error()
	parts, parseErr := ParseActionableValidationError(msg)
	if parseErr != nil {
		if strings.Contains(parseErr.Error(), string(ActionableValidationErrorRemediation)) {
			return []string{string(ActionableValidationErrorRemediation)}
		}
		return []string{parseErr.Error()}
	}

	missing := make([]string, 0, 8)
	if !strings.Contains(parts.What, "validation failed") {
		missing = append(missing, string(ActionableValidationErrorWhat))
	}
	if parts.Why == "" {
		missing = append(missing, string(ActionableValidationErrorWhy))
	}
	if parts.Where == "" {
		missing = append(missing, string(ActionableValidationErrorWhere))
	}
	if parts.When == "" {
		missing = append(missing, string(ActionableValidationErrorWhen))
	}
	if parts.Meaning == "" {
		missing = append(missing, string(ActionableValidationErrorMeaning))
	}
	if parts.Remediation == "" {
		missing = append(missing, string(ActionableValidationErrorRemediation))
	}
	for _, want := range wantContains {
		if strings.TrimSpace(want) == "" {
			continue
		}
		if !strings.Contains(msg, want) {
			missing = append(missing, want)
		}
	}
	return missing
}

func RequireValidationErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("validation succeeded, want an error containing %q", want)
	}
	if strings.TrimSpace(want) == "" {
		t.Fatal("validation error expectation must not be empty")
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("validation error %q is missing %q", err, want)
	}
}

func RequireActionableValidationError(t *testing.T, err error, wantContains ...string) {
	t.Helper()
	if missing := ActionableValidationErrorViolations(err, wantContains...); len(missing) != 0 {
		t.Fatalf("validation error %q is missing actionable dimensions: %s", err, strings.Join(missing, ", "))
	}
}
