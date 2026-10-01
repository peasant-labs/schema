// Hey API emits an unknown intersection for the condition-only oneOf that
// selects a session target or an explicit pattern. Keep the reflected fields,
// close the object in the ordinary strict-object pass, then enforce that choice.
const name = "zAutoPublishRuleRequest";
const plain = `export const ${name} = z.object({`;
const intersection = `export const ${name} = z.intersection(z.unknown(), z.object({`;
const refinement = `.superRefine((value, context) => {
    const sessionTarget = value.sessionId !== undefined;
    const patternTarget = value.kind !== undefined && value.match !== undefined;
    const anyPatternField = value.kind !== undefined || value.match !== undefined;
    if (sessionTarget ? anyPatternField : !patternTarget) {
        context.addIssue({ code: "custom", message: "name exactly one session or explicit pattern target" });
    }
})`;

function declaration(source) {
  const starts = [...source.matchAll(/export const zAutoPublishRuleRequest = /g)];
  if (starts.length !== 1) throw new Error(`automatic-publish target refinement expected exactly one ${name} declaration; found ${starts.length}`);
  const start = starts[0].index;
  const next = source.indexOf("\nexport ", start + 1);
  if (next === -1) throw new Error("automatic-publish target refinement found no following declaration");
  return { start, block: source.slice(start, next) };
}

export function normalizeAutoPublishTargetZod(source) {
  const { start, block } = declaration(source);
  if (block.startsWith(plain)) return source;
  if (!block.startsWith(intersection) || !block.endsWith("\n}));\n")) {
    throw new Error("automatic-publish target refinement found an unexpected generated intersection; no validator was produced");
  }
  const normalized = block.replace(intersection, plain).replace(/\n\}\)\);\n$/, "\n});\n");
  return source.slice(0, start) + normalized + source.slice(start + block.length);
}

export function applyAutoPublishTargetZodRefinements(source) {
  const { start, block } = declaration(source);
  if (!block.startsWith(plain)) throw new Error("automatic-publish target refinement requires the normalized object declaration");
  const refinedTail = `\n}).strict()${refinement};\n`;
  if (block.endsWith(refinedTail)) return source;
  if (!block.endsWith("\n}).strict();\n")) throw new Error("automatic-publish target refinement requires exactly one strict object terminator");
  const refined = block.replace(/\n\}\)\.strict\(\);\n$/, refinedTail);
  return source.slice(0, start) + refined + source.slice(start + block.length);
}
