package schema_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	schema "github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/openapi"
	"github.com/peasant-labs/schema/testcase"
	caseassert "github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/transcript_content_header.yaml
var contentHeaderYAML []byte

//go:embed testdata/transcript_content_header_manifest.yaml
var contentHeaderManifestYAML []byte

type headerProjectionInput struct {
	Detail   string `yaml:"detail"`
	Other    string `yaml:"other"`
	Withhold bool   `yaml:"withhold"`
}

type headerExpected struct {
	JSON          string `yaml:"json"`
	Other         string `yaml:"other"`
	ErrorContains string `yaml:"errorContains"`
	Contains      string `yaml:"contains"`
	NotContains   string `yaml:"notContains"`
}

type headerMarkInput struct {
	Strict     bool     `yaml:"strict"`
	Turns      []string `yaml:"turns"`
	Flags      []int    `yaml:"flags"`
	DropPrefix int      `yaml:"dropPrefix"`
	KeepPrefix int      `yaml:"keepPrefix"`
}

type headerETagMutation struct {
	Field  string `yaml:"field"`
	Change string `yaml:"change"`
}

type headerETagInput struct {
	Guard     string               `yaml:"guard"`
	Left      string               `yaml:"left"`
	Right     string               `yaml:"right"`
	Mutations []headerETagMutation `yaml:"mutations"`
}

type headerETagExpected struct {
	Left  string `yaml:"left"`
	Right string `yaml:"right"`
}

type contentHeaderCases struct {
	Projection   testcase.Corpus[headerProjectionInput, headerExpected] `yaml:"projection"`
	Envelopes    testcase.Corpus[string, headerExpected]                `yaml:"envelopes"`
	ETagBaseline string                                                 `yaml:"etagBaseline"`
	ETagFields   []string                                               `yaml:"etagFields"`
	ETags        testcase.Corpus[headerETagInput, headerETagExpected]   `yaml:"etags"`
	Marks        testcase.Corpus[headerMarkInput, headerExpected]       `yaml:"marks"`
	LocalHeaders testcase.Corpus[string, headerExpected]                `yaml:"localHeaders"`
	Guards       testcase.Corpus[string, string]                        `yaml:"guards"`
}

func loadContentHeaderCases(t *testing.T) contentHeaderCases {
	t.Helper()
	fixtures := decodeStrictYAML[contentHeaderCases](t, contentHeaderYAML, "transcript content header fixture")
	caseassert.RequireValid(t, fixtures.Projection)
	caseassert.RequireValid(t, fixtures.Envelopes)
	caseassert.RequireValid(t, fixtures.ETags)
	caseassert.RequireValid(t, fixtures.Marks)
	caseassert.RequireValid(t, fixtures.LocalHeaders)
	caseassert.RequireValid(t, fixtures.Guards)
	var names []string
	for _, c := range fixtures.Projection.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixtures.Envelopes.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixtures.ETags.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixtures.Marks.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixtures.LocalHeaders.Cases {
		names = append(names, c.Name)
	}
	for _, c := range fixtures.Guards.Cases {
		names = append(names, c.Name)
	}
	caseassert.RequireManifestNames(t, "transcript content header", contentHeaderManifestYAML, names)
	return fixtures
}

func TestSessionDetailPayloadHeaderProjection(t *testing.T) {
	for _, c := range loadContentHeaderCases(t).LocalHeaders.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var payload schema.SessionDetailPayload
			if err := json.Unmarshal([]byte(c.Input), &payload); err != nil {
				t.Fatal(err)
			}
			header := payload.Header()
			source, projected := reflect.ValueOf(payload), reflect.ValueOf(header)
			typ := projected.Type()
			for i := range typ.NumField() {
				field := typ.Field(i)
				value := source.FieldByName(field.Name)
				if !value.IsValid() {
					t.Fatalf("header field %s has no typed payload source", field.Name)
				}
				if value.IsZero() {
					t.Fatalf("fixture leaves header field %s at zero; the projection guard would be vacuous", field.Name)
				}
				if field.Name == "EarlierHistory" {
					got, err := json.Marshal(header.EarlierHistory)
					if err != nil || string(got) != c.Expected.JSON {
						t.Fatalf("history = %s, %v; want %s", got, err, c.Expected.JSON)
					}
					continue
				}
				if !reflect.DeepEqual(value.Interface(), projected.Field(i).Interface()) {
					t.Fatalf("header field %s differs from its typed payload", field.Name)
				}
			}
		})
	}
}

func TestTranscriptHeaderETag(t *testing.T) {
	fixture := loadContentHeaderCases(t)
	baseline := func(t *testing.T, change string) schema.HeaderETagInputs {
		t.Helper()
		var in schema.HeaderETagInputs
		if err := json.Unmarshal([]byte(fixture.ETagBaseline), &in); err != nil {
			t.Fatal(err)
		}
		if change != "" {
			if err := json.Unmarshal([]byte(change), &in); err != nil {
				t.Fatal(err)
			}
		}
		return in
	}
	typ := reflect.TypeFor[schema.HeaderETagInputs]()
	for _, c := range fixture.ETags.Cases {
		t.Run(c.Name, func(t *testing.T) {
			switch c.Input.Guard {
			case "fields":
				fields := make([]string, typ.NumField())
				for i := range fields {
					fields[i] = typ.Field(i).Name
				}
				if !slices.Equal(fields, fixture.ETagFields) {
					t.Fatalf("etag fields = %v, want %v", fields, fixture.ETagFields)
				}
				return
			case "closed-sets":
				hashWord, _ := typ.FieldByName("ContentHashWord")
				visibility, _ := typ.FieldByName("RetainedOwnerVisibility")
				if hashWord.Type != reflect.TypeFor[schema.TranscriptContentHashWord]() ||
					visibility.Type != reflect.TypeFor[schema.TranscriptRetainedOwnerVisibility]() ||
					!slices.Equal(schema.AllTranscriptContentHashWords, []schema.TranscriptContentHashWord{schema.TranscriptContentHashPending, schema.TranscriptContentHashKnown}) ||
					!slices.Equal(schema.AllTranscriptRetainedOwnerVisibilities, []schema.TranscriptRetainedOwnerVisibility{schema.TranscriptRetainedOwnerVisible, schema.TranscriptRetainedOwnerWithheld}) ||
					schema.TranscriptContentHashWord("other").IsValid() || schema.TranscriptRetainedOwnerVisibility("other").IsValid() {
					t.Fatal("etag closed sets are not the declared named types and exact values")
				}
				return
			case "":
			default:
				t.Fatalf("unknown etag guard %q", c.Input.Guard)
			}
			if len(c.Input.Mutations) != 0 {
				base := baseline(t, "")
				baseTag := schema.TranscriptHeaderETag(base)
				var fields []string
				for _, mutation := range c.Input.Mutations {
					changed := baseline(t, mutation.Change)
					fields = append(fields, mutation.Field)
					baseValue, changedValue := reflect.ValueOf(base), reflect.ValueOf(changed)
					for i := 0; i < typ.NumField(); i++ {
						changedField := !reflect.DeepEqual(baseValue.Field(i).Interface(), changedValue.Field(i).Interface())
						if changedField != (typ.Field(i).Name == mutation.Field) {
							t.Fatalf("mutation %s unexpectedly changed field %s: %t", mutation.Field, typ.Field(i).Name, changedField)
						}
					}
					if got := schema.TranscriptHeaderETag(changed); got == baseTag {
						t.Fatalf("mutation of %s left etag unchanged", mutation.Field)
					}
				}
				if !slices.Equal(fields, fixture.ETagFields) {
					t.Fatalf("etag mutation fields = %v, want %v", fields, fixture.ETagFields)
				}
				return
			}
			left := schema.TranscriptHeaderETag(baseline(t, c.Input.Left))
			right := schema.TranscriptHeaderETag(baseline(t, c.Input.Right))
			if left != c.Expected.Left || right != c.Expected.Right || left == right {
				t.Fatalf("etag pair = %q, %q; want distinct %q, %q", left, right, c.Expected.Left, c.Expected.Right)
			}
		})
	}
}

// This checks the shared Schema capture, cook, recorded-index decoder and wire
// against the host's specified remap. The actual host builder has its own
// production-path tests and is not exercised by this Schema corpus.
func TestTranscriptFlaggedMarkRemapReadsTheRecordedIndexPair(t *testing.T) {
	for _, c := range loadContentHeaderCases(t).Marks.Cases {
		t.Run(c.Name, func(t *testing.T) {
			turns := make([]json.RawMessage, 0, c.Input.DropPrefix+c.Input.KeepPrefix+len(c.Input.Turns))
			for range c.Input.DropPrefix {
				turns = append(turns, json.RawMessage(`{"role":"assistant","content":""}`))
			}
			for i := range c.Input.KeepPrefix {
				turns = append(turns, json.RawMessage(fmt.Sprintf(`{"role":"assistant","content":"prefix-%d"}`, i)))
			}
			for _, turn := range c.Input.Turns {
				turns = append(turns, json.RawMessage(turn))
			}
			var raw json.RawMessage
			var err error
			if c.Input.Strict {
				base, marshalErr := json.Marshal(schema.SessionDetailPayload{
					ID: "synthetic", Harness: schema.HarnessCodex,
					TurnCount: len(turns), Turns: []schema.TurnDetail{},
				})
				if marshalErr != nil {
					t.Fatal(marshalErr)
				}
				var members map[string]json.RawMessage
				if err := json.Unmarshal(base, &members); err != nil {
					t.Fatal(err)
				}
				members["turns"], err = json.Marshal(turns)
				if err != nil {
					t.Fatal(err)
				}
				raw, err = json.Marshal(members)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := schema.DecodeSessionDetailPayloadRaw(raw); err != nil {
					t.Fatalf("strict fixture is not accepted: %v", err)
				}
			} else {
				raw, err = json.Marshal(struct {
					Turns []json.RawMessage `json:"turns"`
				}{Turns: turns})
				if err != nil {
					t.Fatal(err)
				}
			}
			captured, err := schema.NewTranscriptCookTurns(raw)
			if err != nil {
				t.Fatal(err)
			}
			harness := schema.Harness("")
			if c.Input.Strict {
				harness = schema.HarnessCodex
			}
			verdict, err := schema.CookTranscriptRows(captured, harness)
			if err != nil {
				t.Fatal(err)
			}
			type indexPair struct {
				state schema.TranscriptRecordedIndexState
				index *int
			}
			indexes := make([]indexPair, len(captured))
			for i, turn := range captured {
				indexes[i].state, indexes[i].index = schema.DecodeTranscriptRecordedIndex(turn.Raw())
				if (indexes[i].state == schema.TranscriptRecordedIndexPresent) != (indexes[i].index != nil) {
					t.Fatalf("stored row %d: state %q disagrees with recorded-index presence", i, indexes[i].state)
				}
			}
			marks := make([]schema.TranscriptFlaggedTurn, 0, len(c.Input.Flags))
			for _, position := range c.Input.Flags {
				if position < 0 || position >= len(captured) || !verdict[position].Keep {
					continue
				}
				marks = append(marks, schema.TranscriptFlaggedTurn{
					Position: verdict[position].ServedPosition,
					Turn:     indexes[position].index,
				})
			}
			encoded, err := json.Marshal(marks)
			if err != nil || string(encoded) != c.Expected.JSON {
				t.Fatalf("marks = %s, %v; want %s", encoded, err, c.Expected.JSON)
			}
			header := schema.TranscriptContentHeader{
				Kind: schema.ContentKindSessionDetail, Revision: "1",
				Representation: "r1.test", FlaggedTurns: marks,
			}
			if err := header.Validate(); err != nil {
				t.Fatalf("mark wire is invalid: %v", err)
			}
		})
	}
}

func TestTranscriptSessionHeaderProjectionIsTotalForServedMetadata(t *testing.T) {
	for _, c := range loadContentHeaderCases(t).Projection.Cases {
		t.Run(c.Name, func(t *testing.T) {
			header, err := schema.NewTranscriptSessionHeader(json.RawMessage(c.Input.Detail))
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("projection error = %v, want %q", err, c.Expected.ErrorContains)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if c.Input.Withhold {
				header = schema.WithholdSessionHeaderRepository(header)
			}
			raw, err := json.Marshal(header)
			if err != nil || string(raw) != c.Expected.JSON {
				t.Fatalf("projection = %s, %v; want %s", raw, err, c.Expected.JSON)
			}
			if c.Input.Other != "" {
				other, err := schema.NewTranscriptSessionHeader(json.RawMessage(c.Input.Other))
				if err != nil {
					t.Fatal(err)
				}
				raw, err := json.Marshal(other)
				if err != nil || string(raw) != c.Expected.Other {
					t.Fatalf("other projection = %s, %v; want %s", raw, err, c.Expected.Other)
				}
			}
		})
	}
}

func TestTranscriptContentHeaderValidation(t *testing.T) {
	for _, c := range loadContentHeaderCases(t).Envelopes.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var header schema.TranscriptContentHeader
			if err := json.Unmarshal([]byte(c.Input), &header); err != nil {
				t.Fatal(err)
			}
			err := header.Validate()
			if c.Classification == testcase.MustFail {
				if err == nil || !strings.Contains(err.Error(), c.Expected.ErrorContains) {
					t.Fatalf("validation = %v, want %q", err, c.Expected.ErrorContains)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(header)
			if err != nil {
				t.Fatal(err)
			}
			if c.Expected.NotContains != "" && strings.Contains(string(raw), c.Expected.NotContains) {
				t.Fatalf("header unexpectedly contains %s: %s", c.Expected.NotContains, raw)
			}
			if c.Expected.Contains != "" && !strings.Contains(string(raw), c.Expected.Contains) {
				t.Fatalf("header lacks %s: %s", c.Expected.Contains, raw)
			}
		})
	}
}

func TestTranscriptSessionHeaderDeclaration(t *testing.T) {
	for _, c := range loadContentHeaderCases(t).Guards.Cases {
		t.Run(c.Name, func(t *testing.T) {
			switch c.Input {
			case "mark-schema":
				assertFlaggedMarkSchema(t)
			case "presence", "omissions":
				assertHeaderDeclaration(t)
			case "records":
				value := reflect.TypeFor[schema.TranscriptSessionHeader]()
				for i := range value.NumField() {
					field := value.Field(i)
					if field.Type == reflect.TypeFor[[]schema.RetainedUnknownRecord]() || field.Type == reflect.TypeFor[*schema.InterpretationDiagnostics]() {
						t.Fatalf("%s carries retained records", field.Name)
					}
				}
				raw, err := json.Marshal(schema.TranscriptContentHeader{})
				if err != nil || strings.Contains(string(raw), "retainedUnknown") || strings.Contains(string(raw), "diagnostics") {
					t.Fatalf("header carries retained records: %s, %v", raw, err)
				}
			case "withholding":
				want := []string{"project", "workingDirectory", "gitBranch", "gitRemote"}
				if !slices.Equal(schema.TranscriptSessionHeaderRepositoryMembers, want) {
					t.Fatalf("withheld members = %v, want %v", schema.TranscriptSessionHeaderRepositoryMembers, want)
				}
			default:
				t.Fatalf("unknown guard %q", c.Input)
			}
		})
	}
}

func assertFlaggedMarkSchema(t *testing.T) {
	t.Helper()
	spec, err := openapi.BuildTypesSpec()
	if err != nil {
		t.Fatal(err)
	}
	component := spec.Components.Schemas["TranscriptFlaggedTurn"]
	if component == nil {
		t.Fatal("flagged-turn schema is absent from the Types catalog")
	}
	raw, err := json.Marshal(component)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Required   []string `json:"required"`
		Properties map[string]struct {
			Type    string `json:"type"`
			Minimum *int64 `json:"minimum"`
			Maximum *int64 `json:"maximum"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("decode flagged-turn schema %s: %v", raw, err)
	}
	turn, ok := wire.Properties["turn"]
	if !ok || turn.Type != "integer" || turn.Minimum == nil || *turn.Minimum != 0 ||
		turn.Maximum == nil || *turn.Maximum != 9007199254740991 ||
		slices.Contains(wire.Required, "turn") || !slices.Contains(wire.Required, "position") {
		t.Fatalf("flagged-turn schema does not require position and permit an optional non-null safe integer turn: %s", raw)
	}
}

func assertHeaderDeclaration(t *testing.T) {
	t.Helper()
	base := reflect.TypeFor[schema.SessionDetailHeader]()
	projected := reflect.TypeFor[schema.TranscriptSessionHeader]()
	omitted := []string{"childSessions", "relationships", "scorecard", "retainedUnknown", "diagnostics"}
	presence := []string{"id", "harness", "startTime", "endTime", "durationMins", "totalTokens", "tokensIn", "tokensOut", "turnCount", "toolCallCount"}
	if !slices.Equal(schema.TranscriptSessionHeaderPresenceMembers, presence) {
		t.Fatalf("presence members = %v, want %v", schema.TranscriptSessionHeaderPresenceMembers, presence)
	}
	seenOmitted := map[string]bool{}
	j := 0
	for i := range base.NumField() {
		b := base.Field(i)
		name := strings.Split(b.Tag.Get("json"), ",")[0]
		if slices.Contains(omitted, name) {
			seenOmitted[name] = true
			continue
		}
		if j >= projected.NumField() {
			t.Fatalf("missing header field %s", name)
		}
		h := projected.Field(j)
		j++
		hName := strings.Split(h.Tag.Get("json"), ",")[0]
		if hName != name || h.Name != b.Name {
			t.Fatalf("field %d = %s/%s, want %s/%s", j-1, h.Name, hName, b.Name, name)
		}
		if slices.Contains(presence, name) {
			want := reflect.PointerTo(b.Type)
			if name == "harness" || name == "startTime" || name == "endTime" {
				want = reflect.TypeFor[*string]()
			}
			if h.Type != want || !strings.Contains(h.Tag.Get("json"), ",omitempty") || h.Tag.Get("nullable") != "false" || strings.Contains(b.Tag.Get("json"), ",omitempty") {
				t.Fatalf("presence field %s has type/tags %s %q; want %s and nonnullable omitempty", name, h.Type, h.Tag, want)
			}
			if name == "turnCount" && h.Tag.Get("description") != b.Tag.Get("description") {
				t.Fatal("turnCount description changed")
			}
			continue
		}
		switch name {
		case "parentSessionId", "rootSessionId":
			if h.Type != reflect.TypeFor[*string]() || h.Tag.Get("nullable") != "false" {
				t.Fatalf("%s must be an open nonnullable pointer string", name)
			}
		case "purpose", "outcome", "sessionOrigin":
			if h.Type != reflect.TypeFor[string]() {
				t.Fatalf("%s must be an open string", name)
			}
		case "inputSubmissionCount", "conversationTurnCount":
			if h.Type != reflect.TypeFor[*int64]() || h.Tag.Get("minimum") != "" || h.Tag.Get("maximum") != "" || h.Tag.Get("nullable") != "false" {
				t.Fatalf("%s must be an unconstrained optional int64", name)
			}
		case "earlierHistory":
			if h.Type != reflect.TypeFor[[]string]() || h.Tag.Get("nullable") != "false" {
				t.Fatal("earlierHistory must be an open nonnullable string array")
			}
		default:
			if h.Type != b.Type || h.Tag.Get("json") != b.Tag.Get("json") {
				t.Fatalf("%s = %s %q, want %s %q", name, h.Type, h.Tag.Get("json"), b.Type, b.Tag.Get("json"))
			}
		}
	}
	if j != projected.NumField() || len(seenOmitted) != len(omitted) {
		t.Fatalf("unexpected header fields or omissions: fields=%d/%d omitted=%v", j, projected.NumField(), seenOmitted)
	}
}
