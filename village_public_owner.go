package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

// VillagePublicOwnerKind selects how a person is shown to the viewer of a
// response: their public profile, or anonymous.
//
// The set is closed for producers and open for readers. A later contract
// version may add a kind (for example an organization credit); a reader that
// meets a kind it does not know MUST treat the person as anonymous: it renders
// no identity and never infers one. VillagePublicOwner.UnmarshalJSON decodes
// such a kind to AnonymousVillageOwner() for Go readers.
type VillagePublicOwnerKind string

const (
	// VillagePublicOwnerProfile carries the person's public profile: id,
	// username, display name, and avatar.
	VillagePublicOwnerProfile VillagePublicOwnerKind = "profile"
	// VillagePublicOwnerAnonymous carries no identifying field at all, not
	// even an id: a stable id would link one anonymous person's transcripts.
	VillagePublicOwnerAnonymous VillagePublicOwnerKind = "anonymous"
)

// AllVillagePublicOwnerKinds is the closed set a producer may emit.
var AllVillagePublicOwnerKinds = []VillagePublicOwnerKind{
	VillagePublicOwnerProfile,
	VillagePublicOwnerAnonymous,
}

// IsValid reports whether k is a kind this contract version defines.
func (k VillagePublicOwnerKind) IsValid() bool {
	for _, known := range AllVillagePublicOwnerKinds {
		if k == known {
			return true
		}
	}
	return false
}

func (k VillagePublicOwnerKind) String() string { return string(k) }

// JSONSchema implements jsonschema.Exposer.
func (VillagePublicOwnerKind) JSONSchema() (jsonschema.Schema, error) {
	return closedStringEnumSchema("Village Public Owner Kind", "How a person is shown to the viewer: their public profile, or anonymous. Readers treat a kind they do not know as anonymous.", AllVillagePublicOwnerKinds), nil
}

// VillagePublicOwner is a person as another viewer sees them: the owner of a
// transcript, a collective member or contributor, or an annotation author.
//
// The server resolves anonymity for each viewer. Consumers render what they
// receive and never re-derive the rule: they do not compare a discoverable
// flag, a membership, or a username to decide whether to show a person. The
// wire carries one of two shapes, selected by kind:
//
//   - profile: {"kind":"profile","id","username","display_name","avatar_url"}.
//     id and username are required; display_name and avatar_url are always
//     present and may be null.
//   - anonymous: {"kind":"anonymous"} and nothing else.
//
// A viewer always receives their own profile, so "this is mine" is exactly
// IsAccount(viewerID): a profile whose id is the viewer's own. An anonymous
// value never matches.
//
// Adding a kind is an additive contract change. Readers treat an unknown kind
// as anonymous (see VillagePublicOwnerKind).
//
// The Go fields carry no JSON tags of their own: MarshalJSON and UnmarshalJSON
// own the wire shape, and JSONSchema declares it.
type VillagePublicOwner struct {
	Kind        VillagePublicOwnerKind `json:"-"`
	ID          VillageUUID            `json:"-"`
	Username    string                 `json:"-"`
	DisplayName *string                `json:"-"`
	AvatarURL   *string                `json:"-"`
}

// NewVillagePublicProfile returns the profile shape of a person.
func NewVillagePublicProfile(id VillageUUID, username string, displayName, avatarURL *string) VillagePublicOwner {
	return VillagePublicOwner{Kind: VillagePublicOwnerProfile, ID: id, Username: username, DisplayName: displayName, AvatarURL: avatarURL}
}

// AnonymousVillageOwner returns the anonymous shape.
func AnonymousVillageOwner() VillagePublicOwner {
	return VillagePublicOwner{Kind: VillagePublicOwnerAnonymous}
}

// IsProfile reports whether the value shows a person's profile. Every other
// kind, including one this contract version does not know, is anonymous.
func (o VillagePublicOwner) IsProfile() bool {
	return o.Kind == VillagePublicOwnerProfile && transcriptIDUUID.MatchString(string(o.ID)) && strings.TrimSpace(o.Username) != ""
}

// IsAccount reports whether the value is the profile of the given Village
// account. Readers use it for "is this mine": an anonymous value is never
// the viewer's, because the server always shows viewers their own profile.
func (o VillagePublicOwner) IsAccount(account VillageUUID) bool {
	return account != "" && o.IsProfile() && o.ID == account
}

// Validate refuses a value a producer must not emit: an unknown kind, a
// profile without a canonical id or a username, or an anonymous value that
// carries any identifying field.
func (o VillagePublicOwner) Validate() error {
	switch o.Kind {
	case VillagePublicOwnerProfile:
		if !transcriptIDUUID.MatchString(string(o.ID)) || strings.TrimSpace(o.Username) == "" {
			return fmt.Errorf("Village public owner validation failed at schema.VillagePublicOwner.Validate: a profile needs a canonical lowercase id and a nonempty username; readers cannot show or link this person; emit the anonymous shape when the viewer may not see the profile")
		}
	case VillagePublicOwnerAnonymous:
		if o.ID != "" || o.Username != "" || o.DisplayName != nil || o.AvatarURL != nil {
			return fmt.Errorf("Village public owner validation failed at schema.VillagePublicOwner.Validate: an anonymous person carries an identifying field; the viewer would learn who it is; emit only the kind")
		}
	default:
		return fmt.Errorf("Village public owner validation failed at schema.VillagePublicOwner.Validate: kind %q is not defined by this contract version; readers on this version would show the person as anonymous; emit profile or anonymous", o.Kind)
	}
	return nil
}

type villagePublicProfileWire struct {
	Kind        VillagePublicOwnerKind `json:"kind"`
	ID          VillageUUID            `json:"id"`
	Username    string                 `json:"username"`
	DisplayName *string                `json:"display_name"`
	AvatarURL   *string                `json:"avatar_url"`
}

type villagePublicAnonymousWire struct {
	Kind VillagePublicOwnerKind `json:"kind"`
}

var villagePublicProfileKeys = []string{"kind", "id", "username", "display_name", "avatar_url"}

// MarshalJSON emits exactly the field set of the value's kind. It refuses a
// value Validate refuses, so a producer cannot send identity on an anonymous
// value or an undefined kind.
func (o VillagePublicOwner) MarshalJSON() ([]byte, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	if o.Kind == VillagePublicOwnerProfile {
		return json.Marshal(villagePublicProfileWire{Kind: o.Kind, ID: o.ID, Username: o.Username, DisplayName: o.DisplayName, AvatarURL: o.AvatarURL})
	}
	return json.Marshal(villagePublicAnonymousWire{Kind: o.Kind})
}

// UnmarshalJSON reads a person the way every reader must:
//
//   - profile requires every declared field and keeps only those;
//   - anonymous keeps nothing but its kind, whatever else arrives;
//   - a kind this contract version does not know decodes to
//     AnonymousVillageOwner(), exactly as the generated TypeScript reader
//     does, so no identity from the unknown shape is kept and re-encoding
//     the decoded value always succeeds.
//
// Fields a later version adds to a shape are ignored, as the generated
// TypeScript reader ignores them, so adding a field or a kind stays additive.
func (o *VillagePublicOwner) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("Village public owner decoding failed at schema.VillagePublicOwner.UnmarshalJSON: the value is null; a person is always an object with a kind; emit the anonymous shape instead of null")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("Village public owner decoding failed at schema.VillagePublicOwner.UnmarshalJSON: %w", err)
	}
	var kind VillagePublicOwnerKind
	raw, ok := fields["kind"]
	if !ok || json.Unmarshal(raw, &kind) != nil || kind == "" {
		return fmt.Errorf("Village public owner decoding failed at schema.VillagePublicOwner.UnmarshalJSON: kind is missing or not a nonempty string; the shape cannot be selected; emit kind profile or anonymous")
	}
	switch kind {
	case VillagePublicOwnerProfile:
		for _, key := range villagePublicProfileKeys {
			if _, present := fields[key]; !present {
				return fmt.Errorf("Village public owner decoding failed at schema.VillagePublicOwner.UnmarshalJSON: profile is missing %q; emit every profile field, with null for an absent display name or avatar", key)
			}
		}
		var wire villagePublicProfileWire
		if err := json.Unmarshal(data, &wire); err != nil {
			return fmt.Errorf("Village public owner decoding failed at schema.VillagePublicOwner.UnmarshalJSON: %w", err)
		}
		decoded := NewVillagePublicProfile(wire.ID, wire.Username, wire.DisplayName, wire.AvatarURL)
		if err := decoded.Validate(); err != nil {
			return err
		}
		*o = decoded
	case VillagePublicOwnerAnonymous:
		*o = AnonymousVillageOwner()
	default:
		*o = AnonymousVillageOwner()
	}
	return nil
}

// JSONSchema implements jsonschema.Exposer: exactly one closed object per
// kind, discriminated by a constant kind.
func (VillagePublicOwner) JSONSchema() (jsonschema.Schema, error) {
	id, err := VillageUUID("").JSONSchema()
	if err != nil {
		return jsonschema.Schema{}, err
	}
	nullableString := func(description string) jsonschema.SchemaOrBool {
		s := jsonschema.Schema{}
		s.AddType(jsonschema.String)
		s.AddType(jsonschema.Null)
		s.WithDescription(description)
		return s.ToSchemaOrBool()
	}
	kindConst := func(kind VillagePublicOwnerKind) jsonschema.SchemaOrBool {
		s := jsonschema.Schema{}
		s.AddType(jsonschema.String)
		s.WithConst(string(kind))
		return s.ToSchemaOrBool()
	}
	username := jsonschema.Schema{}
	username.AddType(jsonschema.String)
	username.WithMinLength(1)
	username.WithDescription("The person's public username")

	profile := jsonschema.Schema{}
	profile.AddType(jsonschema.Object)
	profile.WithTitle("Village Public Owner Profile")
	profile.WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: new(bool)})
	profile.WithProperties(map[string]jsonschema.SchemaOrBool{
		"kind":         kindConst(VillagePublicOwnerProfile),
		"id":           id.ToSchemaOrBool(),
		"username":     username.ToSchemaOrBool(),
		"display_name": nullableString("The person's display name, or null when they have none"),
		"avatar_url":   nullableString("The person's avatar URL, or null when they have none"),
	})
	profile.WithRequired(villagePublicProfileKeys...)

	anonymous := jsonschema.Schema{}
	anonymous.AddType(jsonschema.Object)
	anonymous.WithTitle("Village Public Owner Anonymous")
	anonymous.WithAdditionalProperties(jsonschema.SchemaOrBool{TypeBoolean: new(bool)})
	anonymous.WithProperties(map[string]jsonschema.SchemaOrBool{"kind": kindConst(VillagePublicOwnerAnonymous)})
	anonymous.WithRequired("kind")

	s := jsonschema.Schema{}
	s.WithTitle("Village Public Owner")
	s.WithDescription("A person as another viewer sees them. The server resolves anonymity per viewer; consumers render what they receive and never re-derive the rule. profile carries id, username, display_name, and avatar_url; anonymous carries only its kind, with no id. Readers treat an unknown kind as anonymous.")
	s.WithOneOf(profile.ToSchemaOrBool(), anonymous.ToSchemaOrBool())
	return s, nil
}
