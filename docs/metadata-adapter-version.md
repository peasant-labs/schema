# Managed metadata adapter provenance

`UnifiedMetadata.adapterVersion` identifies the Peasant adapter/parser that
successfully produced a managed metadata/transcript artifact. It is independent
of `version` (the native harness release), `schemaVersion` (the metadata layout),
and the indexer's revision or stored index format. A present value is a positive
integer. Omission means unknown historical provenance, never zero or the current
registry target. Explicit JSON null is invalid.

Metadata schema 10 adds only this optional field. Reading schema 9 does not
upgrade its stamp or invent its producer. A consumer may losslessly adopt the
new layout while leaving `adapterVersion` absent; the bookkeeping migration is
not itself a reason to re-extract native sources. Consumers must preserve future
metadata rather than rewrite it through an older struct. Actual refresh policy
and durable migration belong to the consuming application.

The existing `ComputeMetadataHash` includes a present adapter revision and retains
the exact historic hash when the revision is omitted and other fields are
unchanged. Changing `schemaVersion` still changes the hash. Invalid revisions are
refused at JSON read/write boundaries; the existing hash helper returns its
failure sentinel (an empty string) when metadata cannot be serialized.

This is local managed-artifact provenance. `UnifiedMetadata` is catalogued in
Types, not embedded in the Local or Village API envelopes. Their separately
projected publication types do not acquire this field, and this change makes no
remote preservation promise or capability advertisement. Village API examples
continue to use valid historical schema-9 identity without changing their
released bytes. A future remote provenance feature needs its own compatibility
decision under [the content-capability policy](content-capability-negotiation.md).

Delivery follows the Schema contract ceremony: reviewed Schema PR, maintainer
release/tag and generated TypeScript publication, then consumer re-pin. The Types
catalog advances to 0.15.0; the retired 0.14.0 artifacts remain byte-frozen.
