// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
/**
 * KnownContentCapability is the closed inventory of content-capability tokens known to this
 * pinned schema release. It is intentionally distinct from the OPEN discovery
 * wire alias ContentCapability (an arbitrary string): SchemaVersionResponse's
 * contentCapabilities parses unknown future tokens so an older client keeps
 * working, while these constants give strongly typed access to known tokens.
 * isContentCapability narrows an arbitrary string to a KnownContentCapability, so consumers can filter a
 * discovered list down to the tokens they understand without stringly typing.
 */
export const KnownContentCapability = Object.freeze({
  DetailedUsageV1: "detailed_usage_v1",
  NativeMetadataV1: "native_metadata_v1",
  ObservedModelV1: "observed_model_v1",
  SessionGraphProvenanceV1: "session_graph_provenance_v1",
} as const);
export type KnownContentCapability = (typeof KnownContentCapability)[keyof typeof KnownContentCapability];
export const AllContentCapabilities = Object.freeze([KnownContentCapability.DetailedUsageV1, KnownContentCapability.NativeMetadataV1, KnownContentCapability.ObservedModelV1, KnownContentCapability.SessionGraphProvenanceV1]) as readonly KnownContentCapability[];
export function isContentCapability(value: unknown): value is KnownContentCapability {
  return typeof value === "string" && (AllContentCapabilities as readonly string[]).includes(value);
}
