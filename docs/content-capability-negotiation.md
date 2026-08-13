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

### Evolution

A token's meaning is immutable. Incompatible behavior mints a new token, such
as `observed_model_v2`; a transitional deployment may advertise both revisions.
The inventory is append-only and tokens may be deprecated, but an existing
token is never silently redefined.

With only one known token, the canonical-order and duplicate-rejection producer
rules cannot yet be exercised by a distinct-token permutation. When a second
token is introduced, that same change MUST add a fixture case proving an
unsorted, unique multi-token server advertisement is rejected (or canonicalized
by sorting and deduplication) exactly as designed. Do not add a fabricated
one-token stand-in for this now; add the real permutation with the token that
makes it testable.

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
