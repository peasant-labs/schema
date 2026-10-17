import {
  RedactionCategory, RedactionCategoryLabel, AllRedactionCategories, AllRedactionCategoryLabels,
  parseRedactionCategory, parseRedactionCategoryLabel, isRedactionCategory,
  redactionCategoryFromLabel, redactionCategoryLabel,
} from "@peasant-labs/schema";

const categories: readonly RedactionCategory[] = [RedactionCategory.Secrets, RedactionCategory.PII, RedactionCategory.Paths, RedactionCategory.Project, ...AllRedactionCategories];
const labels: readonly RedactionCategoryLabel[] = [RedactionCategoryLabel.Credential, RedactionCategoryLabel.PII, RedactionCategoryLabel.Path, RedactionCategoryLabel.Internal, ...AllRedactionCategoryLabels];
for (const category of categories) { const label: RedactionCategoryLabel = redactionCategoryLabel(category); redactionCategoryFromLabel(label); }
for (const label of labels) { const category: RedactionCategory = redactionCategoryFromLabel(label); redactionCategoryLabel(category); }
declare const raw: unknown;
declare const text: string;
redactionCategoryLabel(parseRedactionCategory(raw));
redactionCategoryFromLabel(parseRedactionCategoryLabel(raw));
if (isRedactionCategory(raw)) redactionCategoryLabel(raw);
// @ts-expect-error Unknown boundary input requires parsing.
redactionCategoryLabel(raw);
// @ts-expect-error Unknown label input requires parsing.
redactionCategoryFromLabel(raw);
// @ts-expect-error Arbitrary strings are not closed categories.
redactionCategoryLabel(text);
// @ts-expect-error Arbitrary strings are not closed labels.
redactionCategoryFromLabel(text);
// @ts-expect-error Labels cannot be used as categories.
redactionCategoryLabel(RedactionCategoryLabel.Credential);
// @ts-expect-error Categories cannot be used as labels.
redactionCategoryFromLabel(RedactionCategory.Secrets);
