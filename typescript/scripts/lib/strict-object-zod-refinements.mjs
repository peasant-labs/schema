// Closes the generated runtime validators for request bodies the contract
// declares as closed objects.
//
// Hey API renders every OpenAPI object as a plain `z.object({...})`, which in
// Zod STRIPS unknown keys and reports success. For most payloads that is the
// right permissiveness: a producer can add a field without breaking consumers.
// For a partial-update request body it is exactly wrong. The whole meaning of
// such a body is "change precisely these fields", so an unrecognized key means
// the caller asked for something that will not happen, and silently discarding
// it hands back a success the caller will misread as "applied".
//
// That is not hypothetical here: the village accepts a `tags` field on the owner
// update and drops it, returning 200 with the pre-existing tags, so a client
// sees its own input echoed and concludes the edit worked. The Go contract
// refuses unknown fields at its decode boundary; without this pass the
// TypeScript half of the same published package would still accept and strip
// them, and the two language bindings would disagree about what the contract is.
//
// Kept deliberately narrow: only components the Go source marks strict appear
// here, because closing a schema is a contract promise rather than a default.

// STRICT_OBJECTS lists the generated declarations to close. Each name must
// correspond to a component the Go source also closes, so this file cannot
// quietly close something the OpenAPI document leaves open.
export const STRICT_OBJECTS = [
  "zTranscriptUpdateRequest",
  "zAuthoritativePublishRequest",
  "zAuthoritativeSessionIdentity",
  "zAuthoritativeModelInfo",
  "zAuthoritativeTimestampInfo",
  "zAuthoritativeSourceInfo",
  "zAuthoritativeCommitInfo",
  "zAuthoritativeGitContext",
  "zAuthoritativeProjectContext",
  "zAuthoritativeSessionStats",
  "zAuthoritativeQualityMetrics",
  "zAuthoritativeSessionEntry",
  "zAuthoritativeSubagentRef",
  "zAuthoritativeDiagnosticEntry",
  "zAuthoritativeDiagnosticsInfo",
  "zCanonicalPublishGitContext",
  "zCanonicalPublishReplacement",
  "zPublishLicenseOperation",
  "zPublishAssociationOperation",
  "zCanonicalPublishOperation",
  "zPublishedAssociation",
  "zOwnerTranscriptUpdateRequest",
  "zOwnerTranscriptUpdateResponse",
  "zUICapabilitiesResponse",
  "zPublishNormalizedValues",
  "zPublishAppliedState",
  "zAuthoritativePublishResponse",
];

export function applyStrictObjectZodRefinements(source) {
  const states = STRICT_OBJECTS.map((name) => classify(source, name));

  if (states.every((state) => state.state === "raw")) {
    let refined = source;
    for (const state of states) {
      refined = replaceOnce(refined, state.raw, state.refined, state.name);
    }
    return refined;
  }
  if (states.every((state) => state.state === "refined")) return source;

  throw new Error(stateError(states));
}

// classify locates one generated object declaration and reports whether it is
// still raw or already closed. It fails loudly when a declaration is missing or
// duplicated rather than silently skipping, because a silent skip would ship an
// open validator while every gate stayed green.
function classify(source, name) {
  const declaration = `export const ${name} = z.object({`;
  const occurrences = countText(source, declaration);
  if (occurrences !== 1) {
    throw new Error(
      `strict-object refinement expected exactly one \`${declaration}\` declaration in the generated Zod contract but found ${occurrences}; ` +
        `the generator output shape changed, so unknown-field rejection was NOT applied and the published validator would accept and strip unknown keys; ` +
        `re-check the Hey API Zod output for ${name} and update this refinement before regenerating.`,
    );
  }

  const start = source.indexOf(declaration);
  // Find the declaration's own terminator. Scanning for the RAW terminator alone
  // is wrong on an already-refined file: once `.strict()` is appended the raw
  // form no longer matches here, the scan runs on to the NEXT declaration's
  // terminator, and the span silently swallows an unrelated schema — which then
  // gets `.strict()` appended to it. Match the shared prefix and decide from
  // what follows, so the span always ends at this declaration.
  const objectEnd = source.indexOf("\n})", start);
  const nextDeclaration = source.indexOf("\nexport const ", start + declaration.length);
  if (objectEnd === -1 || (nextDeclaration !== -1 && nextDeclaration < objectEnd)) {
    throw new Error(
      `strict-object refinement found the ${name} declaration but no terminating \`});\`; the generated Zod contract is not in the expected shape and no refinement was applied.`,
    );
  }

  const rawTail = "\n});";
  const refinedTail = "\n}).strict();";
  const followsRefined = source.startsWith(refinedTail, objectEnd) || source.startsWith("\n}).strict().superRefine(", objectEnd);
  const followsRaw = source.startsWith(rawTail, objectEnd);
  if (!followsRefined && !followsRaw) {
    throw new Error(
      `strict-object refinement found the ${name} declaration ending in an unrecognized form; expected either \`});\` or \`}).strict();\`, ` +
        `so neither the raw nor the refined state could be identified and no refinement was applied.`,
    );
  }

  const body = source.slice(start, objectEnd);
  const raw = `${body}${rawTail}`;
  const refined = `${body}${refinedTail}`;
  return { name, raw, refined, state: followsRefined ? "refined" : "raw" };
}

function replaceOnce(source, raw, refined, name) {
  if (countText(source, raw) !== 1) {
    throw new Error(`strict-object refinement could not uniquely replace the ${name} declaration; no refinement was applied.`);
  }
  return source.replace(raw, refined);
}

function countText(source, text) {
  let count = 0;
  let index = source.indexOf(text);
  while (index !== -1) {
    count += 1;
    index = source.indexOf(text, index + text.length);
  }
  return count;
}

// stateError refuses a partially-refined file. A mixed state means a previous
// run was interrupted or the generator changed under us, and continuing would
// leave some bodies closed and others silently open.
function stateError(states) {
  const detail = states.map((state) => `${state.name}=${state.state}`).join(", ");
  return (
    `strict-object refinement found a mixed raw/refined generated contract (${detail}); ` +
    `refusing to continue because some request bodies would be closed and others left silently open; ` +
    `regenerate the contract from a clean checkout so every strict object is refined in one pass.`
  );
}
