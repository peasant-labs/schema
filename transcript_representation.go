package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// RepresentationVersion names the representation derivation. Changing its
// inputs or encoding requires a coordinated version change.
const RepresentationVersion = "r1"

// TranscriptViewerProjection is the viewer class used for a projection, not an
// access-control decision.
type TranscriptViewerProjection string

const (
	TranscriptViewerProjectionAnonymous     TranscriptViewerProjection = "anonymous"
	TranscriptViewerProjectionAuthenticated TranscriptViewerProjection = "authenticated"
)

var AllTranscriptViewerProjections = []TranscriptViewerProjection{
	TranscriptViewerProjectionAnonymous, TranscriptViewerProjectionAuthenticated,
}

func (v TranscriptViewerProjection) IsValid() bool {
	return inSet(v, AllTranscriptViewerProjections)
}

func (TranscriptViewerProjection) JSONSchema() (jsonschema.Schema, error) {
	return enumSchema("Transcript Viewer Projection", AllTranscriptViewerProjections)
}

// TranscriptRepresentationInputs are the facts determining a served
// projection. Harness is the spelling in the served detail, or empty when
// absent. Declaration order is part of the derivation.
type TranscriptRepresentationInputs struct {
	RepresentationVersion string
	CutVersion            string
	Revision              string
	Validation            TranscriptReadValidation
	Harness               string
	ModelProvider         string
	ContractVersion       PushContractVersion
	ViewerProjection      TranscriptViewerProjection
	RepositoryWithheld    bool
	NativePresence        TranscriptNativePresence
}

// TranscriptRepresentation derives an opaque generation-scoped identity. It
// is neither a secret nor an authorization token.
func TranscriptRepresentation(in TranscriptRepresentationInputs) string {
	var data strings.Builder
	for _, field := range [...]string{
		in.RepresentationVersion,
		in.CutVersion,
		in.Revision,
		string(in.Validation),
		in.Harness,
		in.ModelProvider,
		string(in.ContractVersion),
		string(in.ViewerProjection),
		strconv.FormatBool(in.RepositoryWithheld),
		string(in.NativePresence),
	} {
		data.WriteString(strconv.Itoa(len(field)))
		data.WriteByte(':')
		data.WriteString(field)
	}
	digest := sha256.Sum256([]byte(data.String()))
	return RepresentationVersion + "." + hex.EncodeToString(digest[:])
}
