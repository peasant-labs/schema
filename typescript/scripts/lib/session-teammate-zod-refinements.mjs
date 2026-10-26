// Go and the published JSON Schema both bound the teammate identity by Unicode
// code points, while the generated Zod chain emits `.max(...)`, which counts
// UTF-16 units. A 200-astral-character name is 200 code points but 400 UTF-16
// units, so the generated validator would refuse a name the wire accepts.
//
// This refinement rewrites ONLY zSessionTeammate: it drops the generated
// `.max(...)` from team and name (keeping `.min(1)`) and adds one code-point
// check on the object, with the bound read from the Go constant so the two
// languages cannot drift apart silently.
export function applySessionTeammateZodRefinements(source, bound) {
  if (!Number.isSafeInteger(bound) || bound <= 0) {
    throw new Error(`session teammate generation failed: the code-point bound ${bound} read from SessionTeammateMaxIdentityRunes is not a positive integer; the generated validator would bound nothing`);
  }
  const declaration = "export const zSessionTeammate = z.object({";
  const start = source.indexOf(declaration);
  if (start < 0) {
    throw new Error(`session teammate generation failed at zSessionTeammate: expected ${JSON.stringify(declaration)} not found; no validator was produced; update the generator for the pinned Zod output`);
  }
  const closing = "\n});";
  const end = source.indexOf(closing, start);
  if (end < start) {
    throw new Error("session teammate generation failed at zSessionTeammate: expected object closing not found; no validator was produced; update the generator for the pinned Zod output");
  }
  const block = source.slice(start, end);
  const emittedBound = `z.string().min(1).max(${bound})`;
  if (!block.includes(emittedBound)) {
    throw new Error(`session teammate generation failed at zSessionTeammate: expected ${JSON.stringify(emittedBound)} on team and name; the generated bound changed shape, so the code-point rewrite would silently do nothing; update the generator for the pinned Zod output`);
  }
  const rewritten = block.split(emittedBound).join("z.string().min(1)");
  const body = `
    const fail = (reason: string) => context.addIssue({ code: "custom", message: "schema.SessionTeammate during local teammate identity validation: " + reason });
    for (const entry of [["team", value.team], ["name", value.name]] as const) {
      const codePoints = [...entry[1]].length;
      if (codePoints > ${bound}) fail(entry[0] + " is " + codePoints + " Unicode code points, over the ${bound}-code-point bound that Go and the published JSON Schema enforce; shorten the recorded " + entry[0] + " at the producing boundary rather than truncating here");
    }`;
  return source.slice(0, start) + rewritten + "\n}).superRefine((value, context) => {" + body + "\n});" + source.slice(end + closing.length);
}
