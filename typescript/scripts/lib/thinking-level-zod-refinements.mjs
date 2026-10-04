// The generated zThinkingLevelRaw inherits code-point maxLength and a regex
// whose \s is narrower than Go's Unicode White_Space rule once ported to
// JavaScript. The canonical constructor bounds the raw spelling by encoded
// UTF-8 bytes and rejects edge Unicode whitespace (including U+0085) while
// permitting an edge U+FEFF, so the TypeScript validator gains two refines
// that restore exact parity with NewThinkingLevelRaw.
const declarationPattern = /export const zThinkingLevelRaw = z\.string\(\)\.min\(1\)\.max\(128\)\.regex\([^;\n]+\);/;

export function applyThinkingLevelZodRefinements(source) {
  const original = source.match(declarationPattern)?.[0];
  if (original === undefined) {
    throw new Error("TypeScript contract generation could not apply the UTF-8 byte limit and edge-whitespace rule to zThinkingLevelRaw in src/internal/generated/contract/zod.gen.ts; the generated declaration no longer matches `z.string().min(1).max(128).regex(...)`, so the package would accept raw thinking levels the Go contract rejects (over 128 bytes or with edge U+0085); inspect the pinned generator output and update scripts/lib/thinking-level-zod-refinements.mjs.");
  }
  const bytes = `.refine((value) => !/[\\uD800-\\uDFFF]/u.test(value) && new TextEncoder().encode(value).length <= 128, { error: "ThinkingLevelRaw is invalid UTF-8 or exceeds 128 bytes" })`;
  const edge = `.refine((value) => { const edge = (text: string) => text === "\\uFEFF" || (!/\\s/u.test(text) && text !== "\\u0085"); return edge(value.charAt(0)) && edge(value.charAt(value.length - 1)); }, { error: "ThinkingLevelRaw must not begin or end with whitespace" })`;
  return source.replace(original, `${original.slice(0, -1)}${bytes}${edge};`);
}
