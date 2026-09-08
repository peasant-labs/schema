# Content capability negotiation

This document is the authoritative specification and contributor policy for
optional transcript content whose loss would change meaning, provenance, or
safety. The protocol prevents a server from accepting a payload while silently
discarding data the user was told would be preserved.

## Normative wire behavior

### Discovery

`GET /api/v1/schema/version` may return `contentCapabilities`, a flat array of
opaque revision-token strings. The first known token is
`observed_model_v1`.

```json
{
  "contentCapabilities": ["observed_model_v1"]
}
```

OpenAPI array items are intentionally open `type: string`, not an `enum`. A
newer server may advertise a future token and an older generated client must
still parse the response and ignore that token. Go and TypeScript usage APIs
separately expose a strongly typed, closed inventory of tokens known to their
pinned schema release.

Matching is exact set membership. The `_v1` suffix is opaque revision identity,
not Semantic Versioning. Clients MUST NOT parse suffixes, compare revisions, or
infer version ranges.

- Omission and `[]` both mean no capabilities. JSON `null` is invalid.
- Clients ignore unknown tokens.
- Servers emit only tokens in their pinned schema inventory.
- Duplicate server emission is invalid. Clients tolerate and deduplicate it.
- Ordering has no meaning. Canonical server serialization deduplicates and sorts
  tokens lexicographically.
- Advertisement is deployment-specific. It describes behavior proven by the
  currently deployed code and migrations, not behavior implied merely by an API
  version.

### `observed_model_v1`

A publication requires `observed_model_v1` if any emitted turn has
`observedModel`. This includes nested or subagent assistant turns. The
session-level seed `model` alone does not require the capability.

A client MUST negotiate immediately before remote publication. If the remote
deployment does not advertise every required token, the client MUST refuse
before upload and MUST NOT silently strip or downgrade evidence. A dry-run does
local payload validation and requirement derivation without network negotiation
or upload. Explicit, user-invoked stripping is a possible future feature and is
outside this protocol.

A server advertising `observed_model_v1` guarantees that it:

1. validates that every `observedModel` is valid and belongs only to an
   assistant-role turn before persistence;
2. creates no database, blob, or other persistence side effects for invalid
   evidence; and
3. preserves every accepted `observedModel` string byte-exactly through storage,
   typed migration, rewrite, serving, and pull.

Nested and subagent assistant turns use the assistant role and receive the same
validation. Byte-exact preservation applies to string values, not JSON envelope
whitespace or object-key order.

### `tool_namespace_v1`

A publication requires `tool_namespace_v1` whenever any tool call has a
`namespace` member, including the empty string and pending tool calls. Namespace
is independent of `name`: omission means not recorded, while a present string
is exact source evidence after normal producer redaction. Producers MUST NOT
trim, lowercase, split, concatenate it into `name`, or infer an extension identity.
Null, non-string values, invalid Unicode, and duplicate JSON members (including
escaped-equivalent keys) are invalid at the public raw boundaries. Typed value
validators also reject invalid Unicode. No namespace-specific size or vocabulary
limit applies; existing whole-document safety limits still apply, not the
native-metadata-only budgets.

A deployment advertising this token guarantees validation before persistence and
exact namespace presence and string-value preservation, independently of name,
through storage, typed migration, rewrite, serving, and pull. Invalid evidence
creates no database, blob, or other persistence effects. Losing this field changes
tool identity: a deployment MUST withhold the token until its production-path
preservation proof passes, including a field-loss mutation. An API or package
version alone does not establish support. Missing advertisement requires refusal
before upload, never stripping or name qualification as a downgrade.

Requirements accumulate across the entire payload: `tool_namespace_v1` does not
replace `observed_model_v1`, `detailed_usage_v1`, or `native_metadata_v1`. Discovery
remains forward-open with exact set matching and sorted, unique producer output.
Dry-run derives these requirements locally without negotiation or upload.

### Native metadata byte budgets

Selected `nativeMetadata[*].data` subtrees permit decoded strings of at most
65,536 UTF-8 bytes. This is not a transcript-wide string limit. The independent
data budgets remain 65,536 raw JSON bytes per record and 1,048,576 bytes in
aggregate per session, both enforced before lossy decoding and again for typed
values. Object syntax, JSON quotes, escaping and raw whitespace consume record
bytes. Thus even a scalar string of exactly 65,536 decoded bytes cannot fit a
record: its two JSON quotes exceed that record budget. A scalar ASCII string of
65,534 bytes fits exactly; nested strings leave less room for other data.

All other metadata bounds remain unchanged: depth 32, array length 4096, object
members 256, keys 512 UTF-8 bytes, 256 records, and `customType` 128 UTF-8 bytes.
The existing numeric, Unicode, attribution and redaction rules still apply.
The Go metadata-string constant is the source for the generated internal
TypeScript limit; both raw scanning and typed validation use that same limit.

### Token evolution

A token's meaning is immutable. Incompatible behavior mints a new token, such
as `observed_model_v2`; a transitional deployment may advertise both revisions.
The inventory is append-only and tokens may be deprecated, but an existing
token is never silently redefined.

Each new token MUST extend the fixture coverage for accumulated requirements,
exact discovery matching, and canonical server output. Unsorted, unique
multi-token server advertisements must be rejected (or canonicalized by sorting
and deduplication) exactly as designed; duplicate and unknown producer tokens
must also remain covered.

### Worked examples

Compatible discovery and payload:

```json
{"contentCapabilities":["future_feature_v1","observed_model_v1"]}
```

```json
{"model":"anthropic/claude-opus","turns":[{"role":"assistant","observedModel":"anthropic/claude-fable-5"}]}
```

The client ignores `future_feature_v1`, finds the exact required token, and may
upload after local validation.

Incompatible discovery and payload:

```json
{"contentCapabilities":[]}
```

```json
{"turns":[{"role":"assistant","observedModel":"anthropic/claude-fable-5"}]}
```

The client refuses before upload. It does not remove `observedModel`. By
contrast, `{"model":"anthropic/claude-opus","turns":[]}` requires no content
capability because `model` is only the session seed.

## Contributor design policy

Use this decision policy when changing a field:

| Field behavior | Compatibility mechanism |
|---|---|
| Required | New API or contract version. |
| Optional and safely ignorable | Additive API bump only. |
| Optional, but loss changes meaning, provenance, or safety | Capability token, or a specifically justified minimum compatible API version. |
| Optional with an explicit, tested downgrade | A capability may be avoided only when conversion is deliberate and user-visible. |
| Inside a truly opaque, byte-preserved payload | No field-level capability when end-to-end preservation is guaranteed. |

Litmus test: **If an old server accepts this payload, does the user lose
something they were told was preserved?** If yes, acceptance needs negotiation
or an explicit, tested downgrade.

### Schema-first delivery

Land the schema PR first, then create the module tag and TypeScript package.
Village and Peasant re-pin only that published artifact. Cross-repo verification
then proves that the advertised deployment behavior matches validation,
persistence, rewrite, serving, pull, and pre-upload client refusal.

### Test ownership

- **Schema** owns fixture-backed wire shape, open-string OpenAPI and Zod parsing,
  closed known-token APIs, requirement derivation, and producer validation.
- **Village** owns fixture-backed pre-persistence validation, no-side-effect
  rejection, canonical advertisement, and storage/migration/rewrite/serve/pull
  preservation.
- **Peasant** owns fixture-backed requirement derivation and refusal before
  upload, including dry-run behavior and nested assistant evidence.

Combinatorial cases live in `testdata/*.yaml`, never inline test tables. Each
repository verifies its mounted production boundary rather than replacing the
system under test with a mock.
