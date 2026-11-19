package schema_test

import (
	"bytes"
	_ "embed"
	"fmt"
	"slices"
	"sync"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/validator_cache_concurrency.yaml
var validatorCacheConcurrencyYAML []byte

// validatorCacheCallers is how many goroutines call one validator at once.
const validatorCacheCallers = 32

type validatorCacheInput struct {
	Validator      string                    `yaml:"validator"`
	Body           string                    `yaml:"body"`
	Datatype       schema.AnnotationDatatype `yaml:"datatype"`
	ConstraintSpec string                    `yaml:"constraint_spec"`
	Value          string                    `yaml:"value"`
}

// call runs the validator the case names and reports its verdict.
func (in validatorCacheInput) call() (bool, error) {
	var err error
	switch in.Validator {
	case "publish-request":
		err = schema.ValidatePublishRequest([]byte(in.Body))
	case "annotation-push-request":
		err = schema.ValidateAnnotationPushRequest([]byte(in.Body))
	case "described-value":
		err = schema.ValidateDescribedValue(schema.ValueDomain{Kind: schema.DomainDescribed, Datatype: in.Datatype, ConstraintSpec: in.ConstraintSpec}, in.Value)
	default:
		return false, fmt.Errorf("fixture names unknown validator %q; use publish-request, annotation-push-request or described-value", in.Validator)
	}
	return err == nil, nil
}

// TestValidatorCachesConcurrently calls each cached validator from many
// goroutines released together, so the first schema compilation and every
// cache read contend. `make race-selective` runs this test alone under the
// race detector, which makes the first use in the process the contended one.
func TestValidatorCachesConcurrently(t *testing.T) {
	t.Parallel()
	var fixture struct {
		RequiredNames []string                                   `yaml:"required_names"`
		Cases         []testcase.Case[validatorCacheInput, bool] `yaml:"cases"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(validatorCacheConcurrencyYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode the validator cache corpus: %v", err)
	}
	corpus := testcase.Corpus[validatorCacheInput, bool]{Cases: fixture.Cases}
	caseassert.RequireValid(t, corpus)
	var names []string
	for _, c := range corpus.Cases {
		names = append(names, c.Name)
	}
	if !slices.Equal(slices.Sorted(slices.Values(names)), slices.Sorted(slices.Values(fixture.RequiredNames))) {
		t.Fatalf("validator cache corpus names %v, required_names %v: reconcile them", names, fixture.RequiredNames)
	}

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			if (c.Classification == testcase.MustPass) != c.Expected {
				t.Fatalf("classification %s disagrees with expected verdict %t", c.Classification, c.Expected)
			}
			start := make(chan struct{})
			verdicts := make([]bool, validatorCacheCallers)
			errs := make([]error, validatorCacheCallers)
			var wg sync.WaitGroup
			for i := range validatorCacheCallers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					verdicts[i], errs[i] = c.Input.call()
				}()
			}
			close(start)
			wg.Wait()
			for i := range validatorCacheCallers {
				if errs[i] != nil {
					t.Fatal(errs[i])
				}
				if verdicts[i] != c.Expected {
					t.Fatalf("caller %d: valid=%t, want %t", i, verdicts[i], c.Expected)
				}
			}
		})
	}
}
