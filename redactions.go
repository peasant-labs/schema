package schema

import "gopkg.in/yaml.v3"

// RedactionFixtureLevel is the redaction level at which a fixture case's rule
// fires. The string values are byte-identical to peasant pkg/redact's
// RedactionLevel ("minimal"/"standard"/"maximum") so the peasant conformance test
// can construct a redactor at the exact firing level via
// redact.RedactionLevel(string(level)) with no translation table.
//
// Default firing semantics are category-based, but the engine may give an
// individual rule a stricter minimum:
//   - secrets, paths -> fire at Minimal and above (unconditional)
//   - pii, project   -> normally fire at Standard and above
//   - selected project rules may require Maximum
//
// The fixture stores the MINIMUM level at which each case's rule fires so the
// conformance test exercises the narrowest level that still triggers the rule
// (avoids entropy/AST over-redaction that only Maximum adds).
type RedactionFixtureLevel string

const (
	// RedactionLevelMinimal redacts only secrets and paths.
	RedactionLevelMinimal RedactionFixtureLevel = "minimal"
	// RedactionLevelStandard adds PII and most project-identity redaction.
	RedactionLevelStandard RedactionFixtureLevel = "standard"
	// RedactionLevelMaximum adds stricter project rules, AST anonymization, and entropy detection.
	RedactionLevelMaximum RedactionFixtureLevel = "maximum"
)

// String returns the wire form of the level.
func (l RedactionFixtureLevel) String() string { return string(l) }

// IsValid reports whether l is one of the three known levels.
func (l RedactionFixtureLevel) IsValid() bool {
	switch l {
	case RedactionLevelMinimal, RedactionLevelStandard, RedactionLevelMaximum:
		return true
	}
	return false
}

// RedactionExample is one entry in the redaction example corpus. It is the
// single source of truth for the redaction session-detail fixture: the leaf
// STORES this data and a format-only generator serialises it to
// testdata/session-detail/redactions.yaml. The leaf never recomputes
// RedactedReplacement — the redaction ENGINE lives only in peasant (pkg/redact),
// and a peasant-side behavioural conformance test binds RedactedReplacement to
// the real engine output (the no-drift guarantee).
type RedactionExample struct {
	// Name is the stable case key (snake_case). The web codegen joins its
	// presentation side-table (UI context snippets) to a case by this Name.
	Name string `yaml:"name"`
	// RuleID is the pkg/redact rule that fires on OriginalText (e.g. "github_pat").
	RuleID string `yaml:"ruleId"`
	// Category is the ENGINE category vocabulary: "secrets" | "pii" | "paths" |
	// "project". This is the ONE canonical category vocabulary; the web display
	// enum (CREDENTIAL|PII|PATH|INTERNAL) is a derived, audited projection.
	Category string `yaml:"category"`
	// Level is the minimum redaction level at which RuleID fires.
	Level RedactionFixtureLevel `yaml:"level"`
	// OriginalText is a realistic, public-safe secret FORMAT (non-functional
	// example value) that the engine genuinely detects.
	OriginalText string `yaml:"originalText"`
	// RedactedReplacement is the engine-APPLIED output, stored verbatim. For
	// back-reference rules this is the partially-redacted form
	// (e.g. "/Users/<USER>/Projects/internal-api"), not a bare label.
	RedactedReplacement string `yaml:"redactedReplacement"`
	// Confidence is a presentation-only 0-100 score shown in the mock UI.
	Confidence int `yaml:"confidence"`
	// LineNumber is a presentation-only source line shown in the mock UI.
	LineNumber int `yaml:"lineNumber"`
	// Description is a human-readable label shown in the mock UI.
	Description string `yaml:"description"`
}

// RedactionExamples is the canonical redaction example corpus — the CURRENT
// session-detail mock cases only (NOT the full pkg/redact corpus). Every
// RedactedReplacement here is the verbatim output of the real pkg/redact engine
// run on OriginalText at Level; the peasant conformance test
// (pkg/redact/redactconform_test.go) enforces that and fails on any drift.
var RedactionExamples = []RedactionExample{
	{
		Name:                "aws_access_key",
		RuleID:              "aws_access_key",
		Category:            "secrets",
		Level:               RedactionLevelMinimal,
		OriginalText:        "AKIAIOSFODNN7EXAMPLE",
		RedactedReplacement: "<AWS_ACCESS_KEY>",
		Confidence:          98,
		LineNumber:          142,
		Description:         "AWS access key ID (pkg/redact rule: aws_access_key)",
	},
	{
		Name:                "email_address",
		RuleID:              "email",
		Category:            "pii",
		Level:               RedactionLevelStandard,
		OriginalText:        "vitor.eduardo@company.com",
		RedactedReplacement: "<EMAIL>",
		Confidence:          94,
		LineNumber:          587,
		Description:         "Email address (pkg/redact rule: email)",
	},
	{
		Name:                "home_directory",
		RuleID:              "unix_home_path",
		Category:            "paths",
		Level:               RedactionLevelMinimal,
		OriginalText:        "/Users/acme-dev/Projects/internal-api",
		RedactedReplacement: "/Users/<USER>/Projects/internal-api",
		Confidence:          72,
		LineNumber:          23,
		Description:         "Unix home directory path with username (pkg/redact rule: unix_home_path)",
	},
	{
		Name:                "github_token",
		RuleID:              "github_pat",
		Category:            "secrets",
		Level:               RedactionLevelMinimal,
		OriginalText:        "ghp_xK9mN2pL4qR7sT8vW1yZ3aB5cD6eF0gH12ab",
		RedactedReplacement: "<GITHUB_PAT>",
		Confidence:          96,
		LineNumber:          310,
		Description:         "GitHub personal access token (pkg/redact rule: github_pat)",
	},
	{
		Name:                "git_remote_url",
		RuleID:              "git_remote_https",
		Category:            "project",
		Level:               RedactionLevelMaximum,
		OriginalText:        "https://github.com/acme-corp/internal-api",
		RedactedReplacement: "<PROJECT_URL>",
		Confidence:          61,
		LineNumber:          445,
		Description:         "Git remote URL exposing project identity (pkg/redact rule: git_remote_https)",
	},
	{
		Name:                "phone_number",
		RuleID:              "phone_us",
		Category:            "pii",
		Level:               RedactionLevelStandard,
		OriginalText:        "+1-555-867-5309",
		RedactedReplacement: "<PHONE>",
		Confidence:          89,
		LineNumber:          712,
		Description:         "US phone number (pkg/redact rule: phone_us)",
	},
	{
		Name:                "database_connection",
		RuleID:              "basic_auth_uri",
		Category:            "secrets",
		Level:               RedactionLevelMinimal,
		OriginalText:        "postgresql://admin:s3cretPass99@db.prod.internal:5432/maindb",
		RedactedReplacement: "postgresql://<BASIC_AUTH_URI>@db.prod.internal:5432/maindb",
		Confidence:          99,
		LineNumber:          18,
		Description:         "URL with embedded basic-auth credentials (pkg/redact rule: basic_auth_uri)",
	},
	{
		Name:                "workspace_path",
		RuleID:              "unix_home_path",
		Category:            "paths",
		Level:               RedactionLevelMinimal,
		OriginalText:        "/home/deploy/apps/acme-billing-service",
		RedactedReplacement: "/home/<USER>/apps/acme-billing-service",
		Confidence:          65,
		LineNumber:          891,
		Description:         "Unix home directory path with username (pkg/redact rule: unix_home_path)",
	},
	{
		Name:                "ip_address",
		RuleID:              "ip_address",
		Category:            "pii",
		Level:               RedactionLevelStandard,
		OriginalText:        "192.168.1.42",
		RedactedReplacement: "<IP_ADDRESS>",
		Confidence:          85,
		LineNumber:          156,
		Description:         "IPv4 address (pkg/redact rule: ip_address)",
	},
	{
		Name:                "slack_webhook",
		RuleID:              "slack_webhook_url",
		Category:            "secrets",
		Level:               RedactionLevelMinimal,
		OriginalText:        "https://hooks.slack.com/services/T024BE7LD/B0DAF2RC3/bJTqg8NS27FYm2pQL0jAiE",
		RedactedReplacement: "<SLACK_WEBHOOK_URL>",
		Confidence:          58,
		LineNumber:          234,
		Description:         "Slack incoming webhook URL (pkg/redact rule: slack_webhook_url)",
	},
	{
		Name:                "stripe_secret",
		RuleID:              "stripe_key",
		Category:            "secrets",
		Level:               RedactionLevelMinimal,
		OriginalText:        "sk_live_4eC39HqLyjWDarjtT1zdp7dc",
		RedactedReplacement: "<STRIPE_KEY>",
		Confidence:          92,
		LineNumber:          67,
		Description:         "Stripe API secret key (pkg/redact rule: stripe_key)",
	},
}

// redactionsDoc is the on-disk YAML shape for redactions.yaml: a single
// top-level `redactions:` list. It is BOTH the generator's serialisation source
// and the unmarshal target consumed by peasant's conformance test + the web
// codegen, so the shape round-trips through one struct.
type redactionsDoc struct {
	Redactions []RedactionExample `yaml:"redactions"`
}

// LoadRedactionExamples parses RedactionsYAML (the embedded, generated artifact)
// into typed cases. Consumers (peasant conformance test, web codegen) bind to the
// PUBLISHED bytes via this loader rather than the in-memory RedactionExamples
// slice, so they validate exactly what ships; the leaf freshness gate keeps the
// two in lockstep.
func LoadRedactionExamples() ([]RedactionExample, error) {
	var doc redactionsDoc
	if err := yaml.Unmarshal(RedactionsYAML, &doc); err != nil {
		return nil, err
	}
	return doc.Redactions, nil
}
