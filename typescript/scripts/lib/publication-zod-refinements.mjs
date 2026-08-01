const marker = "publication-contract-semantic-refinement";

export function applyPublicationZodRefinements(source) {
  if (source.includes(marker)) {
    requireInDeclaration(source, "zAuthoritativePublishRequest", "    visibilityIntent: zVisibilityIntent.optional()");
    requireInDeclaration(source, "zAuthoritativePublishRequest", "    model: zAuthoritativeModelInfo,");
    return source;
  }
  let refined = source;
  requireInDeclaration(refined, "zAuthoritativePublishRequest", "    visibilityIntent: zVisibilityIntent.optional()");
  requireInDeclaration(refined, "zAuthoritativePublishRequest", "    model: zAuthoritativeModelInfo,");
  refined = append(refined, "zPublishLicenseOperation", `(value, context) => {
    const valid = (value.kind === "preserve" && value.license === null) || (value.kind === "replace" && value.license !== null);
    if (!valid) context.addIssue({ code: "custom", message: "license operation must be preserve+null or replace+license" });
  }`);
  refined = append(refined, "zPublishAssociationOperation", `(value, context) => {
    if (value.kind !== "append") context.addIssue({ code: "custom", path: ["kind"], message: "association operation must append" });
    const ids = new Set(); const hashes = new Set(); let prior = "";
    for (const [index, association] of value.associations.entries()) {
      const key = association.id + "\\0" + association.observedCommitHash;
      if (key < prior) context.addIssue({ code: "custom", path: ["associations", index], message: "associations must use canonical ID and binding order" });
      if (ids.has(association.id) || hashes.has(association.observedCommitHash) || association.observedCommitHash.trim() === "") context.addIssue({ code: "custom", path: ["associations", index], message: "associations require unique nonempty IDs and bindings" });
      ids.add(association.id); hashes.add(association.observedCommitHash); prior = key;
    }
  }`);
  refined = append(refined, "zPublishNormalizedValues", `(value, context) => {
    if (!/^[1-9][0-9]*$/.test(value.schemaVersion)) context.addIssue({ code: "custom", path: ["schemaVersion"], message: "schemaVersion must be positive canonical decimal text" });
  }`);
  for (const name of ["zOwnerTranscriptUpdateResponse", "zAuthoritativePublishResponse"]) {
    refined = append(refined, name, `(value, context) => {
      let parsed; try { parsed = new URL(value.transcriptUrl); } catch { context.addIssue({ code: "custom", path: ["transcriptUrl"], message: "transcriptUrl must be absolute HTTPS and bind transcriptId" }); return; }
      if (parsed.protocol !== "https:" || parsed.search || parsed.hash || parsed.pathname.replace(/\\/$/, "") !== "/transcripts/" + value.transcriptId) context.addIssue({ code: "custom", path: ["transcriptUrl"], message: "transcriptUrl must be exact HTTPS /transcripts/{transcriptId}" });
      ${name === "zAuthoritativePublishResponse" ? `if (value.blobKey.trim() === "" || value.blobSizeBytes < 0n || value.publishedAt <= 0n || value.updatedAt < value.publishedAt) context.addIssue({ code: "custom", message: "blob facts and publication chronology must be complete and ordered" });
      if (value.visibility !== value.applied.normalizedValues.visibility) context.addIssue({ code: "custom", path: ["applied", "normalizedValues", "visibility"], message: "top-level and applied visibility must agree" });` : `if (value.updatedAt <= 0n) context.addIssue({ code: "custom", path: ["updatedAt"], message: "updatedAt must be positive" });`}
${name === "zOwnerTranscriptUpdateResponse" ? `      const tags = new Set(); for (const [index, tag] of value.tags.entries()) { if (tag.trim() === "" || tag !== tag.trim() || tags.has(tag)) context.addIssue({ code: "custom", path: ["tags", index], message: "tags must be unique nonempty trimmed strings" }); tags.add(tag); }\n` : ""}
    }`);
  }
  return `// ${marker}\n${refined}`;
}

function requireInDeclaration(source, name, expected) {
  const start = source.indexOf(`export const ${name} = z.object({`);
  const end = source.indexOf(";", start);
  if (start < 0 || end < 0) throw new Error(`publication refinement cannot find ${name}`);
  const declaration = source.slice(start, end + 1);
  if (!declaration.includes(expected)) {
    throw new Error(`publication refinement requires ${name}.${expected.trim()}; visibilityIntent must remain optional for legacy compatibility`);
  }
}

function append(source, name, refinement) {
  const start = source.indexOf(`export const ${name} = z.object({`);
  if (start < 0) throw new Error(`publication refinement cannot find ${name}; semantic validation was not generated`);
  const end = source.indexOf(";", source.indexOf("\n})", start));
  if (end < 0) throw new Error(`publication refinement cannot find ${name} terminator; semantic validation was not generated`);
  const declaration = source.slice(start, end + 1);
  if (!declaration.endsWith(".strict();")) throw new Error(`publication refinement requires strict ${name}; unknown fields would otherwise be stripped`);
  return source.slice(0, end) + `.superRefine(${refinement})` + source.slice(end);
}
