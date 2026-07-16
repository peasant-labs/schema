import type { QualitySession } from "../index.js";
import { canonicalQualityFixtures } from "../internal/generated/quality-fixtures.gen.js";

export const QualityFixtureName = Object.freeze({
  ResolvedTypical: "resolved_typical",
  ResolvedHighTokens: "resolved_high_tokens",
  PartialMedium: "partial_medium",
  FailedComplex: "failed_complex",
  ResolvedMinimal: "resolved_minimal",
} as const);
export type QualityFixtureName = (typeof QualityFixtureName)[keyof typeof QualityFixtureName];
export const AllQualityFixtureNames = Object.freeze(Object.values(QualityFixtureName)) as readonly QualityFixtureName[];

export const QualityFixtureSetName = Object.freeze({ ProjectMix: "project_mix" } as const);
export type QualityFixtureSetName = (typeof QualityFixtureSetName)[keyof typeof QualityFixtureSetName];
export const AllQualityFixtureSetNames = Object.freeze(Object.values(QualityFixtureSetName)) as readonly QualityFixtureSetName[];

export interface QualityStringVariation { value: string; }
export interface QualityRatioVariation { name: string; inputRatio: number; }
export interface QualityMetricVariation { name: string; value: number; }
export interface QualityMetricVariations {
  retryLoops: QualityMetricVariation[];
  signalDensity: QualityMetricVariation[];
  specQualityScore: QualityMetricVariation[];
  filesTouched: QualityMetricVariation[];
  linesChanged: QualityMetricVariation[];
}
export interface QualityVariations {
  outcomes: QualityStringVariation[];
  projects: QualityStringVariation[];
  scopes: QualityStringVariation[];
  taskTitles: QualityStringVariation[];
  tokenRatios: QualityRatioVariation[];
  metrics: QualityMetricVariations;
}
export interface QualitySessionFixture {
  name: QualityFixtureName;
  id: string;
  date: string;
  project: string;
  scope: string;
  title: string;
  totalTokens: number;
  inputTokens: number;
  outputTokens: number;
  turnCount: number;
  toolCalls: number;
  outcome: string;
  filesTouched: number;
  linesChanged: number;
  durationMinutes: number;
  retryLoops: number;
  retryTokensWasted: number;
  withinSessionReverts: number;
  signalDensity: number;
  specQualityScore: number;
  explorationRatio: number;
  scopeBreadth: number;
  discoveryTurns: number;
}
export interface QualityFixtureSet { name: QualityFixtureSetName; cases: QualityFixtureName[]; }
export interface QualityFixtures { sessions: QualitySessionFixture[]; sets: QualityFixtureSet[]; variations: QualityVariations; }

export function loadQualityFixtures(): QualityFixtures { return structuredClone(canonicalQualityFixtures); }

export function sessionByName(fixtures: QualityFixtures, name: QualityFixtureName): QualitySessionFixture | undefined {
  const fixture = fixtures.sessions.find((candidate) => candidate.name === name);
  return fixture === undefined ? undefined : structuredClone(fixture);
}

export function setByName(fixtures: QualityFixtures, name: QualityFixtureSetName): QualityFixtureSet | undefined {
  const fixture = fixtures.sets.find((candidate) => candidate.name === name);
  return fixture === undefined ? undefined : structuredClone(fixture);
}

export function toQualitySession(fixture: QualitySessionFixture): QualitySession {
  const { name: _name, ...session } = fixture;
  return structuredClone(session);
}

export function qualitySessions(fixtures: QualityFixtures): QualitySession[] { return fixtures.sessions.map(toQualitySession); }

export function qualitySessionsForSet(fixtures: QualityFixtures, name: QualityFixtureSetName): QualitySession[] {
  const fixtureSet = setByName(fixtures, name);
  if (fixtureSet === undefined) throw new Error(`unknown quality fixture set ${JSON.stringify(name)}`);
  return fixtureSet.cases.map((caseName) => {
    const fixture = sessionByName(fixtures, caseName);
    if (fixture === undefined) throw new Error(`quality fixture set ${JSON.stringify(name)} references unknown case ${JSON.stringify(caseName)}`);
    return toQualitySession(fixture);
  });
}
