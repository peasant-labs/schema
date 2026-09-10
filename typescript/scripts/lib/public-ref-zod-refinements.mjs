const declarations = ["PublicRevisionRef", "SourceEntryRef", "SubmissionRef"];

export function applyPublicRefZodRefinements(source) {
  let output = source;
  for (const name of declarations) {
    const original = `export const z${name} = z.string().min(1).max(96);`;
    const replacement = `export const z${name} = z.string().min(1).max(96).refine((value) => !/[\\uD800-\\uDFFF]/u.test(value) && new TextEncoder().encode(value).length <= 96, { error: "${name} is invalid UTF-8 or exceeds 96 bytes" });`;
    if (!output.includes(original)) {
      throw new Error(`TypeScript contract generation could not apply the UTF-8 byte limit to z${name}; the generated declaration changed and public references would accept values above 96 bytes; inspect the pinned generator output and update public-ref-zod-refinements.mjs.`);
    }
    output = output.replace(original, replacement);
  }
  const observedOriginal = source.match(/export const zObservedModelID = z\.string\(\)\.min\(1\)\.regex\([^;]+;/)?.[0];
  if (observedOriginal === undefined) {
    throw new Error("TypeScript contract generation could not strengthen zObservedModelID at the JavaScript regex boundary; edge whitespace handling would differ from the canonical validator; inspect the pinned generator output and update public-ref-zod-refinements.mjs.");
  }
  const observedReplacement = `${observedOriginal.slice(0, -1)}.refine((value) => { const edge = (text: string) => text === "\\uFEFF" || (!/\\s/u.test(text) && text !== "\\u0085"); return edge(value.charAt(0)) && edge(value.charAt(value.length - 1)); }, { error: "ObservedModelID must match pattern without edge whitespace" });`;
  output = output.replace(observedOriginal, observedReplacement);
  return output;
}
