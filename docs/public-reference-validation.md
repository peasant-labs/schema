# Public reference validation

`SourceEntryRef`, `SubmissionRef`, and `PublicRevisionRef` required values contain
1 through 96 valid UTF-8 bytes. Constructors, the production JSON Schema format
registry, and public `zSourceEntryRef`, `zSubmissionRef`, and `zPublicRevisionRef`
validators enforce that boundary. They do not trim, restrict Unicode to ASCII,
or accept an empty required alias.

In optional positions, an empty string represents absence, matching Go's empty
string plus `omitempty` encoding. The canonical TypeScript detail, content, and
WebSocket parsers validate the original input first, then normalize empty optional
turn/folded source references, provenance submission references, and general-anchor
references to omitted properties on a clone. Go's validated raw detail/content
decoders encode the same values as omitted fields. Required usage/native source
references remain required; before/through anchors still require both nonempty
source and revision references. Explicit null is not an empty optional reference.

Consumers, including Fairtrade, must use the canonical payload parsers for this
context-sensitive boundary. Do not apply a bare required `zSourceEntryRef` to an
optional empty field and treat its rejection as invalid evidence or legacy data.
For an array-only boundary, distinguish omitted/empty optional references before
using the required alias validator for a nonempty value; retain all other evidence
and reject malformed present values. Neither normalization nor legacy filtering
authorizes dropping nonempty references or provenance, including all-unknown
provenance. Successful parsing does not mutate the caller's input.
