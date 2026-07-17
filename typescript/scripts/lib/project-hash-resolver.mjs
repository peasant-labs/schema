// Pure decision logic behind the Hey API Zod plugin's ProjectHash brand
// resolver (typescript/openapi-ts.config.mjs's $resolvers.string). Split out
// so typescript/tests/*.test.mjs can drive it directly against synthetic
// schema/path pairs, instead of only observing that it does not misfire
// against today's already-single-match Types catalog.

// PROJECT_HASH_PATTERN is the exact regex source the canonical Go ProjectHash
// newtype's JSONSchema emits (see cmd/schema-gen/typescript_project_hash_test.go's
// predecessor and generated/types-*.json's ProjectHash component).
export const PROJECT_HASH_PATTERN = "^[0-9a-f]{64}$";

// shouldBrandProjectHash decides whether a string schema encountered during
// root Types generation should receive the nominal ProjectHash brand. Both
// conditions are required: the schema's pattern must match, AND the schema
// must live at the canonical #/components/schemas/ProjectHash location.
// Pattern alone is not sufficient -- a future unrelated 64-lowercase-hex field
// (for example a SHA-256 content hash) must not silently receive the brand
// merely for sharing the same shape.
export function shouldBrandProjectHash(schema, path) {
  if (schema?.pattern !== PROJECT_HASH_PATTERN) return false;
  return isCanonicalProjectHashSchema(path);
}

// isCanonicalProjectHashSchema anchors the ProjectHash brand to the schema's
// canonical $ref location (#/components/schemas/ProjectHash) rather than its
// pattern string alone. A pattern-only match would silently brand any future
// unrelated 64-lowercase-hex field (for example a SHA-256 content hash) as
// ProjectHash; matching on JSON-pointer identity keeps the nominal brand tied
// to the one real ProjectHash schema component, mirroring the $ref-anchored
// mechanism this generator's predecessor used before the Hey API rewrite.
// `path` is a Hey API SchemaVisitorContext.path Ref: its resolved segment
// array lives under the bracket-only `~ref` property (see
// @hey-api/codegen-core's Ref<T> = { '~ref': T }).
export function isCanonicalProjectHashSchema(path) {
  const segments = path?.["~ref"];
  return Array.isArray(segments) && segments.length === 3 && segments[0] === "components" && segments[1] === "schemas" && segments[2] === "ProjectHash";
}
