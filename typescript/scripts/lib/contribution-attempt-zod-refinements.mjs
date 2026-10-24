// Restore the historical attempt's cross-field contract after shape generation.
// Keep this paired with VillageShareEvent.Validate and the shared YAML corpus.
export function applyContributionAttemptZodRefinements(source) {
  let output = append(source, "VillageShareEvent", `
    const actors = { pending: "", approved: "moderator", rejected: "moderator", retracted: "owner", revoked: "collective" } as const;
    if (value.event_num < 1) fail("event_num must be a positive int32 ordinal");
    if (value.recorded_at === "0001-01-01T00:00:00Z" || /^0001-01-01T00:00:00\\.0+Z$/.test(value.recorded_at)) fail("recorded_at must be a nonzero UTC timestamp");
    if (value.decided_by_actor !== actors[value.status]) fail("decided_by_actor does not match the status's actor capacity");
    if (value.status === "pending" && value.decided_at !== null) fail("pending attempts must have null decided_at");
    if (value.status !== "pending" && value.status !== "approved" && value.status !== "rejected" && value.decided_at === null) fail("this terminal status requires decided_at");
    if (value.decided_at !== null && (value.decided_at === "0001-01-01T00:00:00Z" || /^0001-01-01T00:00:00\\.0+Z$/.test(value.decided_at))) fail("decided_at must be a nonzero UTC timestamp when present");`);
  output = append(output, "LocalRemoteContributionAttempt", `
    if (value.firstSeenAt === "0001-01-01T00:00:00Z" || /^0001-01-01T00:00:00\\.0+Z$/.test(value.firstSeenAt)) fail("firstSeenAt must be a nonzero UTC timestamp");`);
  return output;
}

function append(source, name, body) {
  const startMarker = `export const z${name} = z.object({`;
  const endMarker = `\n\nexport type ${name} =`;
  const start = source.indexOf(startMarker);
  const end = source.indexOf(endMarker, start);
  if (start === -1 || end === -1 || source.indexOf(startMarker, start + startMarker.length) !== -1) {
    throw new Error(`Contribution attempt Zod refinement failed in scripts/lib/contribution-attempt-zod-refinements.mjs during generation: expected one ${name} object declaration and following type; generated validation is incomplete; inspect the pinned generator output and update this postprocessor before publishing.`);
  }
  const declaration = source.slice(start, end);
  const refinement = `.superRefine((value, ctx) => {\n  const fail = (message: string) => ctx.addIssue({ code: "custom", message: "Collective contribution attempt validation failed: " + message + "; retain prior evidence and fetch validated owner history before retrying" });\n${body}\n})`;
  if (declaration.endsWith(`${refinement};`)) return source;
  if (!declaration.endsWith("});") || declaration.includes(".superRefine(")) {
    throw new Error(`Contribution attempt Zod refinement failed for ${name} during generation: the object declaration changed or carries an unexpected refinement; output cannot be trusted; inspect scripts/lib/contribution-attempt-zod-refinements.mjs and regenerate.`);
  }
  return source.slice(0, start) + declaration.slice(0, -1) + refinement + ";" + source.slice(end);
}
