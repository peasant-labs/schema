// The generated zThinkingLevelRaw inherits code-point maxLength and a regex
// whose \s is narrower than Go's Unicode White_Space rule once ported to
// JavaScript. The canonical constructor bounds the raw spelling by encoded
// UTF-8 bytes and rejects edge Unicode whitespace (including U+0085) while
// permitting an edge U+FEFF, so the TypeScript validator gains two refines
// that restore exact parity with NewThinkingLevelRaw.
//
// The generated zThinkingLevelChange inherits the published JSON Schema's
// anyOf required level/raw only as an intersection with z.unknown(), which the
// pinned generator cannot turn into a runtime check. A third refine restores
// the at-least-one-field rule so an empty history entry is refused exactly as
// the Go validator refuses it, without changing the inferred object type.
const declarationPattern = /export const zThinkingLevelRaw = z\.string\(\)\.min\(1\)\.max\(128\)\.regex\([^;\n]+\);/;
const changeDeclarationPattern = /export const zThinkingLevelChange = z\.intersection\(z\.unknown\(\), z\.object\(\{[\s\S]*?\}\)\);/;

export function applyThinkingLevelZodRefinements(source) {
  const original = source.match(declarationPattern)?.[0];
  if (original === undefined) {
    throw new Error("TypeScript contract generation could not apply the UTF-8 byte limit and edge-whitespace rule to zThinkingLevelRaw in src/internal/generated/contract/zod.gen.ts; the generated declaration no longer matches `z.string().min(1).max(128).regex(...)`, so the package would accept raw thinking levels the Go contract rejects (over 128 bytes or with edge U+0085); inspect the pinned generator output and update scripts/lib/thinking-level-zod-refinements.mjs.");
  }
  const changeDeclaration = source.match(changeDeclarationPattern)?.[0];
  if (changeDeclaration === undefined) {
    throw new Error("TypeScript contract generation could not apply the at-least-one-field rule to zThinkingLevelChange in src/internal/generated/contract/zod.gen.ts; the generated declaration no longer matches `z.intersection(z.unknown(), z.object({...}))`, so the package would accept an empty thinkingLevelHistory entry the Go contract rejects; inspect the pinned generator output and update scripts/lib/thinking-level-zod-refinements.mjs.");
  }
  const bytes = `.refine((value) => !/[\\uD800-\\uDFFF]/u.test(value) && new TextEncoder().encode(value).length <= 128, { error: "ThinkingLevelRaw is invalid UTF-8 or exceeds 128 bytes" })`;
  const edge = `.refine((value) => { const edge = (text: string) => text === "\\uFEFF" || (!/\\s/u.test(text) && text !== "\\u0085"); return edge(value.charAt(0)) && edge(value.charAt(value.length - 1)); }, { error: "ThinkingLevelRaw must not begin or end with whitespace" })`;
  const change = `.refine((value) => value.level !== undefined || value.raw !== undefined, { error: "ThinkingLevelChange must carry level or raw" })`;
  const withRawRefinements = source.replace(original, `${original.slice(0, -1)}${bytes}${edge};`);
  return withRawRefinements.replace(changeDeclaration, `${changeDeclaration.slice(0, -1)}${change};`);
}
