// Code generated from github.com/peasant-labs/redact@v0.1.6 (h1:Kk03w/5aGmagtuKrV2NLeRVeclA90UBghSXKjTbU4ls=). DO NOT EDIT.
// Authority: AllCategories(), Category.Validate(), Category.String().
export const RedactionCategory = Object.freeze({
  Secrets: "secrets",
  PII: "pii",
  Paths: "paths",
  Project: "project",
} as const);
export type RedactionCategory = (typeof RedactionCategory)[keyof typeof RedactionCategory];
export const AllRedactionCategories: readonly RedactionCategory[] = Object.freeze(Object.values(RedactionCategory));

export const RedactionCategoryLabel = Object.freeze({
  Credential: "CREDENTIAL",
  PII: "PII",
  Path: "PATH",
  Internal: "INTERNAL",
} as const);
export type RedactionCategoryLabel = (typeof RedactionCategoryLabel)[keyof typeof RedactionCategoryLabel];
export const AllRedactionCategoryLabels: readonly RedactionCategoryLabel[] = Object.freeze(Object.values(RedactionCategoryLabel));

const labels: Readonly<Record<RedactionCategory, RedactionCategoryLabel>> = Object.freeze({
  "secrets": "CREDENTIAL",
  "pii": "PII",
  "paths": "PATH",
  "project": "INTERNAL"
});
const categories: Readonly<Record<RedactionCategoryLabel, RedactionCategory>> = Object.freeze({
  "CREDENTIAL": "secrets",
  "PII": "pii",
  "PATH": "paths",
  "INTERNAL": "project"
});

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
