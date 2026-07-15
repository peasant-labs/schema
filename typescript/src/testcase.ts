import { parseAllDocuments } from "yaml";
import {
  Classification,
  ProvenanceSource,
  isClassification,
  isProvenanceSource,
  validateCorpus,
  type Case,
  type Corpus,
  type Mutation,
  type Provenance,
} from "./internal/generated/testcase.js";

export {
  AllClassifications,
  AllProvenanceSources,
  Classification,
  ProvenanceSource,
  checkMin,
  isClassification,
  isProvenanceSource,
  validateCase,
  validateCorpus,
} from "./internal/generated/testcase.js";
export type { Case, Corpus, Mutation, Provenance } from "./internal/generated/testcase.js";

export interface CorpusDecoders<I, E> {
  decodeInput(value: unknown, path: string): I;
  decodeExpected(value: unknown, path: string): E;
}

export class CorpusError extends Error {
  readonly path: string;

  constructor(path: string, message: string, options?: ErrorOptions) {
    super(`${path}: ${message}`, options);
    this.name = "CorpusError";
    this.path = path;
  }
}

export function loadCorpus<I, E>(source: string, decoders: CorpusDecoders<I, E>): Corpus<I, E> {
  let documents;
  try {
    documents = parseAllDocuments(source, { strict: true, uniqueKeys: true });
  } catch (error) {
    throw new CorpusError("$", "YAML parsing failed", { cause: error });
  }
  if (documents.length !== 1) {
    throw new CorpusError("$", `expected exactly one YAML document, got ${documents.length}`);
  }
  const document = documents[0];
  if (document === undefined) {
    throw new CorpusError("$", "YAML parser returned no document");
  }
  if (document.errors.length > 0) {
    throw new CorpusError("$", `YAML parsing failed: ${document.errors.map((error) => error.message).join("; ")}`);
  }

  let value: unknown;
  try {
    value = document.toJS({ maxAliasCount: 0 });
  } catch (error) {
    throw new CorpusError("$", "YAML conversion failed", { cause: error });
  }
  const root = requireRecord(value, "$", ["cases"]);
  requireKeys(root, "$", ["cases"]);
  if (!Array.isArray(root.cases)) {
    throw new CorpusError("$.cases", "must be a sequence");
  }

  const cases = root.cases.map((candidate, index): Case<I, E> => {
    const path = `$.cases[${index}]`;
    const row = requireRecord(candidate, path, ["name", "input", "expected", "classification", "provenance", "mutation"]);
    requireKeys(row, path, ["name", "input", "expected", "classification", "provenance", "mutation"]);
    const name = requireString(row.name, `${path}.name`);
    const classification = row.classification;
    if (!isClassification(classification)) {
      throw new CorpusError(`${path}.classification`, `${JSON.stringify(classification)} is not one of must-pass/must-fail`);
    }

    const provenanceValue = requireRecord(row.provenance, `${path}.provenance`, ["source", "ref"]);
    requireKeys(provenanceValue, `${path}.provenance`, ["source", "ref"]);
    const sourceValue = provenanceValue.source;
    if (!isProvenanceSource(sourceValue)) {
      throw new CorpusError(`${path}.provenance.source`, `${JSON.stringify(sourceValue)} is not a known source`);
    }
    const provenance: Provenance = {
      source: sourceValue,
      ref: requireString(provenanceValue.ref, `${path}.provenance.ref`),
    };

    const mutationValue = requireRecord(row.mutation, `${path}.mutation`, ["description"]);
    requireKeys(mutationValue, `${path}.mutation`, ["description"]);
    const mutation: Mutation = {
      description: requireString(mutationValue.description, `${path}.mutation.description`),
    };

    return {
      name,
      input: decoders.decodeInput(row.input, `${path}.input`),
      expected: decoders.decodeExpected(row.expected, `${path}.expected`),
      classification,
      provenance,
      mutation,
    };
  });

  const corpus: Corpus<I, E> = { cases };
  const err = validateCorpus(corpus);
  if (err !== undefined) {
    throw new CorpusError("$", err.message, { cause: err });
  }
  return corpus;
}

function requireRecord(value: unknown, path: string, allowedKeys: readonly string[]): Record<string, unknown> {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new CorpusError(path, "must be a mapping");
  }
  const record = Object.fromEntries(Object.entries(value));
  for (const key of Object.keys(record)) {
    if (!allowedKeys.includes(key)) {
      throw new CorpusError(`${path}.${key}`, "unknown key");
    }
  }
  return record;
}

function requireKeys(record: Record<string, unknown>, path: string, requiredKeys: readonly string[]): void {
  for (const key of requiredKeys) {
    if (!Object.hasOwn(record, key)) {
      throw new CorpusError(`${path}.${key}`, "required key is missing");
    }
  }
}

function requireString(value: unknown, path: string): string {
  if (typeof value !== "string") {
    throw new CorpusError(path, "must be a string");
  }
  return value;
}
