// Pure functions behind the Local/Village API operation generator's
// canonical-root-type aliasing. Split out of generate-contract-support.mjs (which
// has file-I/O side effects at import time) so typescript/tests/*.test.mjs can
// exercise the collision guard directly instead of only observing that it does
// not throw on today's already-consistent specs.

// canonicalOperationAliases compares every Local/Village API OpenAPI component
// schema against the canonical root Types schema its name normalizes to, and
// throws an actionable error when a same-named component's schema is not
// structurally equal to the canonical definition (aliasing it would misrepresent
// the HTTP contract). It returns a Map of apiComponentName -> canonicalRootName
// for every component that DOES alias cleanly.
export function canonicalOperationAliases(surface, apiSpec, rootSpec) {
  const apiSchemas = apiSpec?.components?.schemas;
  const rootSchemas = rootSpec?.components?.schemas;
  if (typeof apiSchemas !== "object" || apiSchemas === null || typeof rootSchemas !== "object" || rootSchemas === null) {
    throw new Error(`TypeScript ${surface} operation generation could not compare components.schemas with the canonical Types catalog; operation payload identities cannot be trusted; regenerate both OpenAPI documents from Go.`);
  }
  const rootCanonical = new Map(Object.entries(rootSchemas).map(([name, schema]) => [name, stableSchema("types", schema)]));
  const aliases = new Map();
  for (const [rawName, schema] of Object.entries(apiSchemas)) {
    const rootName = canonicalTypeScriptName(surface, rawName);
    const rootSchema = rootCanonical.get(rootName);
    if (rootSchema === undefined) continue;
    const apiSchema = stableSchema(surface, schema);
    if (apiSchema !== rootSchema) {
      throw new Error(`TypeScript ${surface} operation generation found API component ${rawName} normalizing to canonical root type ${rootName} with an unequal schema; aliasing it would misrepresent the HTTP contract; rename the operation-specific component or make it exactly equal to the canonical Go/OpenAPI definition.`);
    }
    aliases.set(rawName, rootName);
  }
  return aliases;
}

// canonicalTypeScriptName maps a raw OpenAPI component name to the public
// TypeScript identity it should share with the canonical root Types catalog:
// strip the surface's "Schema" prefix (Local/Village API components are
// namespaced that way; the root Types catalog is not), then fold the two
// historical Harness leak names onto the canonical Harness identity.
export function canonicalTypeScriptName(surface, name) {
  const unprefixed = surface === "types" ? name : name.replace(/^Schema/, "");
  return unprefixed === "BestiaryHarness" || unprefixed === "Provider" ? "Harness" : unprefixed;
}

// stableSchema renders a schema (with its $refs canonicalized through
// canonicalTypeScriptName) into a normalized JSON string suitable for exact
// structural-equality comparison between two documents' independent copies of
// what should be the same component.
export function stableSchema(surface, schema) {
  return JSON.stringify(normalizeSchemaRefs(surface, structuredClone(schema)));
}

// normalizeSchemaRefs deep-sorts object keys with a plain ordinal comparator
// (not localeCompare: JSON Schema keys are always ASCII, and an ICU-collation
// sort is an unnecessary and potentially non-portable dependency for what is
// effectively an identifier sort) and canonicalizes every $ref target through
// canonicalTypeScriptName, so two documents that alias the same component under
// different surface-prefixed names normalize to byte-identical JSON.
export function normalizeSchemaRefs(surface, value) {
  if (Array.isArray(value)) return value.map((item) => normalizeSchemaRefs(surface, item));
  if (typeof value !== "object" || value === null) return value;
  return Object.fromEntries(Object.entries(value).sort(([left], [right]) => ordinalCompare(left, right)).map(([key, item]) => {
    if (key === "$ref" && typeof item === "string") {
      const prefix = "#/components/schemas/";
      return [key, item.startsWith(prefix) ? prefix + canonicalTypeScriptName(surface, item.slice(prefix.length)) : item];
    }
    return [key, normalizeSchemaRefs(surface, item)];
  }));
}

function ordinalCompare(left, right) {
  if (left < right) return -1;
  if (left > right) return 1;
  return 0;
}
