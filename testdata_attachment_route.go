package schema

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/peasant-labs/schema/testcase"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/local-api/attachment_route.yaml
var attachmentRouteYAML []byte

//go:embed testdata/local-api/attachment_route_manifest.yaml
var attachmentRouteManifestYAML []byte

// AttachmentRouteOperationFixture is the declared local image read: the path
// parameter, the success media types, the immutable cache header, and the
// refusal statuses. The OpenAPI document and the Go validators must match it.
type AttachmentRouteOperationFixture struct {
	Path               string   `yaml:"path"`
	Method             string   `yaml:"method"`
	OperationID        string   `yaml:"operation_id"`
	Parameter          string   `yaml:"parameter"`
	Pattern            string   `yaml:"pattern"`
	MinLength          int      `yaml:"min_length"`
	MaxLength          int      `yaml:"max_length"`
	SuccessStatus      int      `yaml:"success_status"`
	SuccessMediaTypes  []string `yaml:"success_media_types"`
	CacheHeader        string   `yaml:"cache_header"`
	CacheControl       string   `yaml:"cache_control"`
	MalformedStatus    int      `yaml:"malformed_status"`
	MalformedCode      string   `yaml:"malformed_code"`
	MissingStatus      int      `yaml:"missing_status"`
	MissingCode        string   `yaml:"missing_code"`
	DescriptionClauses []string `yaml:"description_clauses"`
}

// AttachmentRouteRefInput is one path-parameter spelling.
type AttachmentRouteRefInput struct {
	Ref string `yaml:"ref"`
}

// AttachmentRouteRefExpected is whether that spelling is a lowercase 64-hex
// sha256, and the rule a rejection must name.
type AttachmentRouteRefExpected struct {
	Accepted      bool   `yaml:"accepted"`
	ErrorContains string `yaml:"error_contains,omitempty"`
}

// AttachmentRouteResponseInput names one closed route outcome.
type AttachmentRouteResponseInput struct {
	Outcome string `yaml:"outcome"`
}

// AttachmentRouteResponseExpected is the declared status, media type, cache
// header, and error code for that outcome.
type AttachmentRouteResponseExpected struct {
	Accepted      bool   `yaml:"accepted"`
	Status        int    `yaml:"status,omitempty"`
	MediaType     string `yaml:"media_type,omitempty"`
	CacheControl  string `yaml:"cache_control,omitempty"`
	Code          string `yaml:"code,omitempty"`
	ErrorContains string `yaml:"error_contains,omitempty"`
}

// AttachmentRouteFixtures is the segmented local image-read corpus: one arm
// for the path parameter, one arm for the declared responses, and the operation
// facts both arms and the OpenAPI document must match.
type AttachmentRouteFixtures struct {
	Operation AttachmentRouteOperationFixture                                                `yaml:"operation"`
	Refs      testcase.Corpus[AttachmentRouteRefInput, AttachmentRouteRefExpected]           `yaml:"refs"`
	Responses testcase.Corpus[AttachmentRouteResponseInput, AttachmentRouteResponseExpected] `yaml:"responses"`
}

type attachmentRouteManifest struct {
	Refs      []string `yaml:"refs"`
	Responses []string `yaml:"responses"`
}

// LoadAttachmentRouteFixtures loads the local image-read corpus and proves each
// arm's exact case-name inventory. The manifest is a separate document so a
// count-preserving rename still fails.
func LoadAttachmentRouteFixtures() (AttachmentRouteFixtures, error) {
	var fixtures AttachmentRouteFixtures
	decoder := yaml.NewDecoder(bytes.NewReader(attachmentRouteYAML))
	decoder.KnownFields(true)
	if err := decoder.Decode(&fixtures); err != nil {
		return AttachmentRouteFixtures{}, fmt.Errorf("load attachment route fixtures: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return AttachmentRouteFixtures{}, fmt.Errorf("load attachment route fixtures trailing document: %w", err)
		}
		return AttachmentRouteFixtures{}, fmt.Errorf("load attachment route fixtures: multiple YAML documents are not allowed")
	}
	if err := fixtures.Refs.Validate(); err != nil {
		return AttachmentRouteFixtures{}, fmt.Errorf("load attachment route ref fixtures: %w", err)
	}
	if err := fixtures.Responses.Validate(); err != nil {
		return AttachmentRouteFixtures{}, fmt.Errorf("load attachment route response fixtures: %w", err)
	}
	if err := fixtures.Operation.validate(); err != nil {
		return AttachmentRouteFixtures{}, err
	}
	manifest, err := loadAttachmentRouteManifest(attachmentRouteManifestYAML)
	if err != nil {
		return AttachmentRouteFixtures{}, err
	}
	if err := requireFixtureNames(fixtures.Refs.Cases, manifest.Refs); err != nil {
		return AttachmentRouteFixtures{}, fmt.Errorf("attachment route ref fixtures: %w", err)
	}
	if err := requireFixtureNames(fixtures.Responses.Cases, manifest.Responses); err != nil {
		return AttachmentRouteFixtures{}, fmt.Errorf("attachment route response fixtures: %w", err)
	}
	return fixtures, nil
}

func (o AttachmentRouteOperationFixture) validate() error {
	if o.Path == "" || o.Method == "" || o.OperationID == "" || o.Parameter == "" || o.Pattern == "" {
		return fmt.Errorf("attachment route operation fixture is missing its path, method, operation id, parameter, or pattern")
	}
	if o.MinLength != 64 || o.MaxLength != 64 || o.SuccessStatus == 0 || o.MalformedStatus == 0 || o.MissingStatus == 0 {
		return fmt.Errorf("attachment route operation fixture is missing its length or status facts")
	}
	if len(o.SuccessMediaTypes) != 2 || o.CacheHeader == "" || o.CacheControl == "" || o.MalformedCode == "" || o.MissingCode == "" {
		return fmt.Errorf("attachment route operation fixture is missing its media types, cache header, or error codes")
	}
	if len(o.DescriptionClauses) == 0 {
		return fmt.Errorf("attachment route operation fixture has no description clauses")
	}
	return nil
}

func loadAttachmentRouteManifest(data []byte) (attachmentRouteManifest, error) {
	var manifest attachmentRouteManifest
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return attachmentRouteManifest{}, fmt.Errorf("load attachment route manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return attachmentRouteManifest{}, fmt.Errorf("load attachment route manifest: expected exactly one YAML document: %v", err)
	}
	if len(manifest.Refs) == 0 || len(manifest.Responses) == 0 {
		return attachmentRouteManifest{}, fmt.Errorf("attachment route manifest is missing a ref or response inventory")
	}
	return manifest, nil
}
