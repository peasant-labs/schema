package openapi

import (
	"bytes"
	_ "embed"
	"sort"
	"strings"
	"testing"

	"github.com/swaggest/openapi-go/openapi31"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/operation_declarations.yaml
var operationDeclarationsYAML []byte

type operationDeclarationFixture struct {
	RequiredNames []string                   `yaml:"required_names"`
	Cases         []operationDeclarationCase `yaml:"cases"`
}

type operationDeclarationCase struct {
	Name          string          `yaml:"name"`
	Call          string          `yaml:"call"`
	Setup         []declaredProbe `yaml:"setup"`
	Operations    []declaredProbe `yaml:"operations"`
	ErrorContains string          `yaml:"error_contains"`
	OK            bool            `yaml:"ok"`
}

// declaredProbe is the part of a localOperationSpec a declaration mistake can
// touch. The probe operations carry no request or response, so the only thing
// under test is the declaration check.
type declaredProbe struct {
	Method        string           `yaml:"method"`
	Path          string           `yaml:"path"`
	ID            string           `yaml:"id"`
	HostRefusal   bool             `yaml:"host_refusal"`
	ErrorStatuses []int            `yaml:"error_statuses"`
	ErrorCodes    map[int][]string `yaml:"error_codes"`
}

func (p declaredProbe) spec() localOperationSpec {
	op := localOperationSpec{
		method: p.Method, path: p.Path, id: p.ID, tag: "probe",
		hostRefusal: p.HostRefusal, errorStatuses: p.ErrorStatuses,
	}
	if len(p.ErrorCodes) > 0 {
		op.errorCodes = p.ErrorCodes
	}
	if len(p.ErrorStatuses) > 0 {
		op.errorResponse = new(LocalErrorResponse)
	}
	return op
}

// The generator's own operation specs are valid, so the checks that refuse a
// malformed declaration never run during a build. This test builds each
// malformed declaration and requires the error the check must give, so a check
// that stops firing, or fires with another message, fails here.
func TestOperationDeclarationChecksRefuseMistakes(t *testing.T) {
	var fixture operationDeclarationFixture
	decoder := yaml.NewDecoder(bytes.NewReader(operationDeclarationsYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixture); err != nil {
		t.Fatalf("decode operation declaration fixture: %v", err)
	}
	names := make([]string, 0, len(fixture.Cases))
	for _, c := range fixture.Cases {
		names = append(names, c.Name)
	}
	sort.Strings(names)
	required := append([]string(nil), fixture.RequiredNames...)
	sort.Strings(required)
	if strings.Join(names, "\n") != strings.Join(required, "\n") {
		t.Fatalf("operation declaration cases = %v, required names = %v; a case was added or removed without its manifest entry", names, required)
	}

	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if len(c.Operations) == 0 {
				t.Fatal("case names no operation under test")
			}
			if !c.OK && c.ErrorContains == "" {
				t.Fatal("case neither expects success nor names the error it must produce")
			}
			reflector := openapi31.NewReflector()
			if len(c.Setup) > 0 {
				setup := make([]localOperationSpec, len(c.Setup))
				for i, probe := range c.Setup {
					setup[i] = probe.spec()
				}
				if err := addLocalOperations(reflector, setup); err != nil {
					t.Fatalf("setup operations did not declare: %v", err)
				}
			}
			subject := c.Operations[0].spec()
			var err error
			switch c.Call {
			case "add_operations":
				specs := make([]localOperationSpec, len(c.Operations))
				for i, probe := range c.Operations {
					specs[i] = probe.spec()
				}
				err = addLocalOperations(reflector, specs)
			case "check_error_codes":
				err = checkErrorCodes(subject)
			case "describe_refusal":
				err = describeRefusal(reflector.Spec, subject)
			case "describe_error_codes":
				err = describeErrorCodes(reflector.Spec, subject)
			default:
				t.Fatalf("case names the unknown call %q", c.Call)
			}
			switch {
			case c.OK && err != nil:
				t.Fatalf("%s refused a valid declaration: %v", c.Call, err)
			case !c.OK && err == nil:
				t.Fatalf("%s accepted a malformed declaration; it must fail with %q", c.Call, c.ErrorContains)
			case !c.OK && !strings.Contains(err.Error(), c.ErrorContains):
				t.Fatalf("%s error = %q, want it to contain %q", c.Call, err, c.ErrorContains)
			}
		})
	}
}
