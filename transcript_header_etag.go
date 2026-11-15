package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// TranscriptContentHashWord distinguishes a pending hash from a known one in
// the header's revalidation inputs.
type TranscriptContentHashWord string

const (
	TranscriptContentHashPending TranscriptContentHashWord = "pending"
	TranscriptContentHashKnown   TranscriptContentHashWord = "known"
)

var AllTranscriptContentHashWords = []TranscriptContentHashWord{
	TranscriptContentHashPending, TranscriptContentHashKnown,
}

func (v TranscriptContentHashWord) IsValid() bool {
	return inSet(v, AllTranscriptContentHashWords)
}

// TranscriptRetainedOwnerVisibility states whether evidence availability may
// be disclosed to this viewer.
type TranscriptRetainedOwnerVisibility string

const (
	TranscriptRetainedOwnerVisible  TranscriptRetainedOwnerVisibility = "visible"
	TranscriptRetainedOwnerWithheld TranscriptRetainedOwnerVisibility = "withheld"
)

var AllTranscriptRetainedOwnerVisibilities = []TranscriptRetainedOwnerVisibility{
	TranscriptRetainedOwnerVisible, TranscriptRetainedOwnerWithheld,
}

func (v TranscriptRetainedOwnerVisibility) IsValid() bool {
	return inSet(v, AllTranscriptRetainedOwnerVisibilities)
}

// HeaderETagInputs contains the facts that may change a header independently
// of the served turn representation. Declaration order is pinned by a test.
type HeaderETagInputs struct {
	Representation          string
	ContractVersion         PushContractVersion
	Kind                    ContentKind
	Canonical               bool
	OmittedImages           int
	FlaggedTurns            []TranscriptFlaggedTurn
	RetainedEvidence        *TranscriptRetainedState
	ContentHashWord         TranscriptContentHashWord
	RetainedOwnerVisibility TranscriptRetainedOwnerVisibility
}

// TranscriptHeaderETag returns a quoted HTTP entity tag for the complete
// header input set. Inputs are encoded as a JSON struct in declaration order,
// then hashed; no content or input values appear in the result.
func TranscriptHeaderETag(in HeaderETagInputs) string {
	// The declaration contains only JSON-supported primitives, pointers, and
	// slices of primitives, so encoding it cannot fail.
	encoded, _ := json.Marshal(in)
	digest := sha256.Sum256(encoded)
	return `"` + hex.EncodeToString(digest[:]) + `"`
}
