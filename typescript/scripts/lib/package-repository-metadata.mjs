import assert from "node:assert/strict";

export const CANONICAL_REPOSITORY = Object.freeze({
  type: "git",
  url: "https://github.com/peasant-labs/schema",
});

export function assertPackageRepositoryMetadata(packageManifest, expectedRepository) {
  assert.deepEqual(
    packageManifest.repository,
    expectedRepository,
    "package/package.json must declare the canonical peasant-labs/schema git repository metadata for npm provenance",
  );
}
