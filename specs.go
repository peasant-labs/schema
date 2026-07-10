package schema

import (
	"embed"
	"fmt"
)

// generatedFS embeds the committed OpenAPI spec artifacts (JSON + YAML) produced
// by `go run ./cmd/schema-gen`. It is the single in-binary source consumers read
// the published specs from — village serves GET /openapi.json from
// VillageAPISpecJSON() rather than vendoring its own copy, so the served spec
// follows the go.mod pin automatically.
//
//go:embed generated
var generatedFS embed.FS

// VillageAPISpecJSON returns the JSON bytes of the CURRENT Village API OpenAPI
// spec — the version named by VillageAPIVersion. It is version-aware: the filename
// is derived from the single-source VillageAPIVersion constant, never a literal, so
// a version bump in versions.go re-points this accessor in lockstep and consumers
// (e.g. village's GET /openapi.json) need no change.
//
// The bytes are embedded at compile time, so a failure here means the committed
// generated/ artifact for VillageAPIVersion is missing — a build/generation bug,
// not a runtime condition. It panics with an actionable message in that case; the
// codegen-freshness gate and TestVillageAPISpecJSON_MatchesCurrentVersion make this
// unreachable in a healthy tree.
func VillageAPISpecJSON() []byte {
	name := "generated/village-api-" + VillageAPIVersion + ".json"
	b, err := generatedFS.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf(
			"schema.VillageAPISpecJSON: embedded spec %q is missing: %v.\n"+
				"  what: the generated Village API spec for VillageAPIVersion=%s is not committed under generated/.\n"+
				"  why:  the version const and the committed artifact drifted, or generation was skipped.\n"+
				"  fix:  run `go run ./cmd/schema-gen` and commit generated/, or correct VillageAPIVersion in versions.go.",
			name, err, VillageAPIVersion))
	}
	return b
}

// PublishRequestSchemaJSON returns the standalone JSON-Schema (draft 2020-12)
// bytes for the PublishRequest wire format at the CURRENT VillageAPIVersion — the
// schema extracted from the Village API spec by openapi.BuildPublishRequestSchema
// and committed under generated/publish-request-<VillageAPIVersion>.schema.json.
//
// This is the SINGLE BYTE-SOURCE the publish-enforce path validates against:
// schema.ValidatePublishRequest (root publish_validate.go) compiles exactly these
// bytes, and village's publish handler enforces through it rather
// than vendoring its own schema copy. Routing both the documented spec and the
// enforced schema through this one accessor means they can never drift (it retired
// the hand-maintained validate/schema.json, which had diverged from the generated
// artifact, and — rc2 #118 — folded the standalone `validate` subpackage into the
// root schema package).
//
// It is version-aware: the filename is derived from VillageAPIVersion, never a
// literal, so a version bump in versions.go re-points it in lockstep. The bytes
// are embedded at compile time, so a miss here is a build/generation bug, not a
// runtime condition — it panics with an actionable message. The codegen-freshness
// gate and TestPublishRequestSchemaJSON_MatchesGenerated keep this unreachable in
// a healthy tree.
func PublishRequestSchemaJSON() []byte {
	name := "generated/publish-request-" + VillageAPIVersion + ".schema.json"
	b, err := generatedFS.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf(
			"schema.PublishRequestSchemaJSON: embedded schema %q is missing: %v.\n"+
				"  what: the generated PublishRequest JSON-Schema for VillageAPIVersion=%s is not committed under generated/.\n"+
				"  why:  the version const and the committed artifact drifted, or generation was skipped.\n"+
				"  fix:  run `go run ./cmd/schema-gen` and commit generated/, or correct VillageAPIVersion in versions.go.",
			name, err, VillageAPIVersion))
	}
	return b
}
