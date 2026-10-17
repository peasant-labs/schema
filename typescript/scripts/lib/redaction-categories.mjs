// Values and pairs come only from the isolated released-engine tool.
export function renderRedactionCategories(projection) {
  const fail = () => { throw new Error("Redaction projection validation failed during TypeScript generation: category membership or unique nonempty labels disagree with the released engine bindings; no browser binding was generated; restore the pinned engine projection and review contract changes before retrying."); };
  const { categories, labels, pairs, module, version, sum } = projection;
  if (module !== "github.com/peasant-labs/redact" || version !== "v0.1.6" || typeof sum !== "string" || !sum.startsWith("h1:")) fail();
  const categoryNames = new Set(["Secrets", "PII", "Paths", "Project"]);
  const labelNames = new Set(["Credential", "PII", "Path", "Internal"]);
  const validateMembers = (members, names) => {
    if (!Array.isArray(members)) fail();
    const values = new Set();
    const seen = new Set();
    for (const member of members) {
      if (!names.has(member.name) || seen.has(member.name) || typeof member.value !== "string" || member.value === "" || values.has(member.value)) fail();
      seen.add(member.name); values.add(member.value);
    }
    if (seen.size !== names.size) fail();
    return values;
  };
  const keys = validateMembers(categories, categoryNames);
  const values = validateMembers(labels, labelNames);
  if (!Array.isArray(pairs)) fail();
  const seenKeys = new Set(), seenLabels = new Set();
  for (const pair of pairs) {
    if (!Array.isArray(pair) || pair.length !== 2 || !keys.has(pair[0]) || !values.has(pair[1]) || seenKeys.has(pair[0]) || seenLabels.has(pair[1])) fail();
    seenKeys.add(pair[0]); seenLabels.add(pair[1]);
  }
  if (seenKeys.size !== keys.size || seenLabels.size !== values.size) fail();
  const binding = (name, members, inventory) => `export const ${name} = Object.freeze({\n${members.map(m => `  ${m.name}: ${JSON.stringify(m.value)},`).join("\n")}\n} as const);\nexport type ${name} = (typeof ${name})[keyof typeof ${name}];\nexport const ${inventory}: readonly ${name}[] = Object.freeze(Object.values(${name}));\n`;
  return `// Code generated from ${module}@${version} (${sum}). DO NOT EDIT.
// Authority: AllCategories(), Category.Validate(), Category.String().
${binding("RedactionCategory", categories, "AllRedactionCategories")}
${binding("RedactionCategoryLabel", labels, "AllRedactionCategoryLabels")}
const labels: Readonly<Record<RedactionCategory, RedactionCategoryLabel>> = Object.freeze(${JSON.stringify(Object.fromEntries(pairs), null, 2)});
const categories: Readonly<Record<RedactionCategoryLabel, RedactionCategory>> = Object.freeze(${JSON.stringify(Object.fromEntries(pairs.map(([key, label]) => [label, key])), null, 2)});

function invalid(operation: string, allowed: readonly string[]): never {
  throw new TypeError("Redaction category validation failed at @peasant-labs/schema " + operation + " during conversion: input is outside the exact case-sensitive closed set; review cannot safely identify sensitive content; pass a named binding or parse boundary data first. Allowed values: " + allowed.join(", ") + ".");
}
export function isRedactionCategory(value: unknown): value is RedactionCategory {
  return typeof value === "string" && Object.hasOwn(labels, value);
}
export function parseRedactionCategory(value: unknown): RedactionCategory {
  if (!isRedactionCategory(value)) invalid("parseRedactionCategory", AllRedactionCategories);
  return value;
}
export function parseRedactionCategoryLabel(value: unknown): RedactionCategoryLabel {
  if (typeof value !== "string" || !Object.hasOwn(categories, value)) invalid("parseRedactionCategoryLabel", AllRedactionCategoryLabels);
  return value as RedactionCategoryLabel;
}
export function redactionCategoryFromLabel(label: RedactionCategoryLabel): RedactionCategory {
  if (typeof label !== "string" || !Object.hasOwn(categories, label)) invalid("redactionCategoryFromLabel", AllRedactionCategoryLabels);
  return categories[label];
}
export function redactionCategoryLabel(category: RedactionCategory): RedactionCategoryLabel {
  if (!isRedactionCategory(category)) invalid("redactionCategoryLabel", AllRedactionCategories);
  return labels[category];
}
`;
}
