// The safe-integer source positions are JSON numbers, not coercing bigint
// values. Preserve the Go/OpenAPI bounds without accepting strings or booleans.
export function applyRetainedUnknownZodRefinements(source) {
  const start = source.indexOf("export const zRetainedUnknownRecord = z.object({");
  const end = source.indexOf("\n});", start);
  if (start < 0 || end < start) throw new Error("retained unknown generation failed: record declaration missing; update the pinned generator integration before publishing");
  let record = source.slice(start, end);
  for (const field of ["position", "recordIndex"]) {
    const before = `${field}: z.coerce.bigint().gte(BigInt(0)).lte(BigInt(9007199254740991))`;
    if (!record.includes(before)) throw new Error(`retained unknown generation failed: ${field} no longer has the expected safe-integer bounds; inspect the generated schema before publishing`);
    record = record.replace(before, `${field}: z.number().int().min(0).max(9007199254740991)`);
  }
  return source.slice(0, start) + record + source.slice(end);
}
