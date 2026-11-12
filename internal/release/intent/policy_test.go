package intent_test

// Policy-violation suite: drives intent.PolicyViolation over the policy
// corpus. The constructor is the pinned rejection shape: must-fail rows
// assert the typed code with all six parts, the accept sibling asserts the
// report echoes the diagnosis verbatim.

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/policy_corpus.yaml
var policyCorpusYAML []byte

type policyInput struct {
	Rule    string `yaml:"rule"`
	Where   string `yaml:"where"`
	Message string `yaml:"message"`
}

type policyExpected struct {
	Code string `yaml:"code"`
}

func TestPolicyViolation(t *testing.T) {
	corpus := loadNamedCorpus[policyInput, policyExpected](t, policyCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got := intent.PolicyViolation(c.Input.Rule, c.Input.Where, c.Input.Message)
			if c.Classification == "must-fail" {
				requireSixParts(t, got, c.Expected.Code)
				return
			}
			if string(got.Code) != c.Expected.Code {
				t.Errorf("Code = %q, want %q", got.Code, c.Expected.Code)
			}
			if !strings.Contains(got.What, c.Input.Rule) || !strings.Contains(got.What, c.Input.Message) {
				t.Errorf("What = %q, want it to echo the rule and the message", got.What)
			}
			if got.Where != c.Input.Where {
				t.Errorf("Where = %q, want %q", got.Where, c.Input.Where)
			}
		})
	}
}
