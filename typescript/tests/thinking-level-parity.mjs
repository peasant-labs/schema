// Shared loader for the thinking-level parity replays: reads the Go corpora and
// the TypeScript replay markers, and proves every Go case is classified once.
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { parse } from "yaml";

const root = new URL("../../", import.meta.url);
const readYAML = async (path) => parse(await readFile(new URL(path, root), "utf8"), { strict: true, uniqueKeys: true });

export const parity = await readYAML("testdata/typescript/thinking_level_parity.yaml");
export const metadataCorpus = await readYAML("testdata/metadata/thinking_level.yaml");
export const rawCorpus = await readYAML("testdata/metadata/thinking_level_raw.yaml");
export const turnCorpus = await readYAML("testdata/session-detail/transcripts/turn_thinking_levels.yaml");

// Returns the replayed Go cases after proving replay and goOnly partition the
// corpus names exactly.
export function replayedCases(label, corpus, markers) {
  const replay = markers?.replay;
  const goOnly = markers?.goOnly;
  assert.ok(Array.isArray(replay) && replay.length > 0, `${label}: testdata/typescript/thinking_level_parity.yaml must list a non-empty replay sequence`);
  assert.ok(goOnly !== null && typeof goOnly === "object" && !Array.isArray(goOnly), `${label}: goOnly must be a mapping of case name to reason`);
  for (const [name, reason] of Object.entries(goOnly)) {
    assert.ok(typeof reason === "string" && reason.trim().length > 0, `${label}: goOnly case ${JSON.stringify(name)} must state why it cannot replay in TypeScript`);
    assert.ok(!replay.includes(name), `${label}: case ${JSON.stringify(name)} is both replayed and Go-only; classify it once`);
  }
  const classified = new Set([...replay, ...Object.keys(goOnly)]);
  assert.equal(classified.size, replay.length + Object.keys(goOnly).length, `${label}: replay lists a case name twice`);
  const corpusNames = corpus.cases.map((fixtureCase) => fixtureCase.name);
  for (const name of corpusNames) assert.ok(classified.has(name), `${label}: Go case ${JSON.stringify(name)} is not classified in testdata/typescript/thinking_level_parity.yaml; add it to replay or goOnly with a reason`);
  for (const name of classified) assert.ok(corpusNames.includes(name), `${label}: marker ${JSON.stringify(name)} names no Go case; remove or rename it`);
  return corpus.cases.filter((fixtureCase) => replay.includes(fixtureCase.name));
}

export function issuePaths(result) {
  return result.success ? [] : result.error.issues.map((issue) => issue.path.join("."));
}
