// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
export const Classification = Object.freeze({
  MustPass: "must-pass",
  MustFail: "must-fail",
} as const);
export type Classification = (typeof Classification)[keyof typeof Classification];
export const AllClassifications = Object.freeze([Classification.MustPass, Classification.MustFail]) as readonly Classification[];
export function isClassification(value: unknown): value is Classification {
  return typeof value === "string" && (AllClassifications as readonly string[]).includes(value);
}

export const ProvenanceSource = Object.freeze({
  Requirement: "requirement",
  Bug: "bug",
  Enum: "enum",
  Boundary: "boundary",
  Manual: "manual",
} as const);
export type ProvenanceSource = (typeof ProvenanceSource)[keyof typeof ProvenanceSource];
export const AllProvenanceSources = Object.freeze([ProvenanceSource.Requirement, ProvenanceSource.Bug, ProvenanceSource.Enum, ProvenanceSource.Boundary, ProvenanceSource.Manual]) as readonly ProvenanceSource[];
export function isProvenanceSource(value: unknown): value is ProvenanceSource {
  return typeof value === "string" && (AllProvenanceSources as readonly string[]).includes(value);
}
