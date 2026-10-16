import type {
  CollectiveObservationScope, LocalContributionEvidenceKind, CollectiveFetchDomainResult,
  CollectiveFetchReport, LocalConfirmedContributionReceipt, LocalOperationDiagnostic,
  PullAssociationMetadataResult, LocalContributionHistoryResponse, OperationIssue,
  CollectiveObservationResponse, CollectiveObservation, CollectiveFetchDomainSummary,
  VillageOrganizationCollectivesResponse,
} from "./internal/generated/contract/zod.gen.js";

export interface ContributionTraversalIdentity {
  scope: CollectiveObservationScope;
  incarnation: string;
  kind: LocalContributionEvidenceKind;
  generation: string;
}
export interface LocalContributionHistoryQuery {
  kind: LocalContributionEvidenceKind;
  collectiveId?: string;
  page: number;
  limit: number;
  revision?: string;
}
export function validateContributionHistoryQuery(query: LocalContributionHistoryQuery, current: string): number {
  if (!kinds.includes(query.kind) || !Number.isSafeInteger(query.page) || query.page < 1 || !Number.isSafeInteger(query.limit) || query.limit < 1 || query.limit > 200 || (query.collectiveId !== undefined && !uuid.test(query.collectiveId))) return 400;
  if (query.revision === undefined) return query.page > 1 ? 400 : 200;
  try { parseContributionTraversalRevision(query.revision); } catch { return 400; }
  return query.revision === current ? 200 : 409;
}
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const kinds = ["remote-attempts", "receipts", "operation-diagnostics"];
const decimal = (v: string): boolean => /^(0|[1-9][0-9]{0,19})$/.test(v) && BigInt(v) <= 18446744073709551615n;
const utc = (v: string | null | undefined): boolean => v == null || (!/^0001-01-01T00:00:00(?:\.0+)?Z$/.test(v) && v.endsWith("Z"));
function invalid(reason: string): TypeError { return new TypeError(`Collective observation contract validation failed during metadata processing: ${reason}; retain prior evidence and correct the request or fetch validated metadata before retrying.`); }

export function normalizeCollectiveObservationScope(raw: string, userId: string): CollectiveObservationScope {
  if (new TextEncoder().encode(raw).length > 1024 || !uuid.test(userId) || raw.includes("\\") || raw.trim() !== raw) throw invalid("scope URL or account identity is invalid");
  let url: URL;
  try { url = new URL(raw); } catch { throw invalid("scope requires an absolute HTTP(S) Village URL"); }
  if (!["http:", "https:"].includes(url.protocol) || !url.hostname || /[^\x00-\x7f]/.test(raw.split("/")[2] ?? "") || url.pathname.includes("%")) throw invalid("scope has unsupported protocol, host or ambiguous escaped path");
  url.username = ""; url.password = ""; url.search = ""; url.hash = "";
  const pathname = url.pathname.split("/").filter(Boolean).join("/");
  return { villageUrl: url.origin + (pathname === "" ? "" : "/" + pathname), userId };
}
function scopeValid(scope: CollectiveObservationScope): boolean { try { const v = normalizeCollectiveObservationScope(scope.villageUrl, scope.userId); return v.villageUrl === scope.villageUrl && v.userId === scope.userId; } catch { return false; } }
export function encodeContributionTraversalRevision(value: ContributionTraversalIdentity): string {
  if (!scopeValid(value.scope) || !uuid.test(value.incarnation) || !kinds.includes(value.kind) || !decimal(value.generation)) throw invalid("traversal identity fields are invalid");
  const json = JSON.stringify(["1", value.scope.villageUrl, value.scope.userId, value.incarnation, value.kind, value.generation]);
  const token = btoa(String.fromCharCode(...new TextEncoder().encode(json))).replaceAll("+", "-").replaceAll("/", "_").replace(/=+$/, "");
  if (token.length > 4096) throw invalid("traversal token exceeds its byte bound");
  return token;
}
export function parseContributionTraversalRevision(token: string): ContributionTraversalIdentity {
  if (token.length === 0 || token.length > 4096 || !/^[A-Za-z0-9_-]+$/.test(token)) throw invalid("traversal token is missing or not bounded base64url");
  let fields: unknown;
  try { fields = JSON.parse(new TextDecoder("utf-8", { fatal: true }).decode(Uint8Array.from(atob(token.replaceAll("-", "+").replaceAll("_", "/")), c => c.charCodeAt(0)))); } catch { throw invalid("traversal token framing is invalid"); }
  if (!Array.isArray(fields) || fields.length !== 6 || fields.some(v => typeof v !== "string") || fields[0] !== "1") throw invalid("traversal token version or field framing is invalid");
  const value: ContributionTraversalIdentity = { scope: { villageUrl: fields[1], userId: fields[2] }, incarnation: fields[3], kind: fields[4] as LocalContributionEvidenceKind, generation: fields[5] };
  if (encodeContributionTraversalRevision(value) !== token) throw invalid("traversal token is not canonical");
  return value;
}

// Called only after the generated Zod shape parser succeeds. This function
// supplies cross-field semantics; it does not replace generated wire parsing.
export function collectiveContractIssue(name: string, value: unknown): string | undefined {
  if (name === "VillageOrganizationCollectivesResponse") { const v=value as VillageOrganizationCollectivesResponse; if(new Set(v.collectives.map(c=>c.id)).size!==v.collectives.length) return "organization projection repeats a collective identity"; return; }
  if (name === "CollectiveObservationScope") return scopeValid(value as CollectiveObservationScope) ? undefined : "scope is not normalized";
  if (name === "ContributionTraversalRevision") { try { parseContributionTraversalRevision(value as string); return; } catch { return "traversal token is malformed"; } }
  if (name === "OperationIssue") { const v = value as OperationIssue; if (!v.message.trim() || !v.retry.trim() || new TextEncoder().encode(v.message).length > 2048 || new TextEncoder().encode(v.retry).length > 2048) return "issue needs bounded actionable message/retry"; return; }
  if (name === "PullAssociationMetadataResult") {
    const v = value as PullAssociationMetadataResult, needsTime = v.state !== "unknown", needsWarning = v.state === "stale" || v.state === "unknown";
    if ((v.observedAt != null) !== needsTime || !utc(v.observedAt) || (v.warning != null) !== needsWarning || (v.warning != null && v.warning.operation !== "pull")) return "metadata state/time/warning pairing is invalid";
    return;
  }
  if (name === "LocalConfirmedContributionReceipt") {
    const v = value as LocalConfirmedContributionReceipt;
    if (new TextEncoder().encode(v.receiptId).length > 256 || (v.sourceReceiptId != null && new TextEncoder().encode(v.sourceReceiptId).length > 256)) return "receipt identities exceed their encoded byte bounds";
    if (!utc(v.confirmedAt) || (v.sourceReceiptId != null && v.sourceReceiptId === "") || (v.operation === "publish" ? v.outcome !== "published" : v.operation !== "share" || v.outcome !== "submitted")) return "receipt must confirm its actual publish/share effect and optional known target";
    return;
  }
  if (name === "LocalOperationDiagnostic") { const v = value as LocalOperationDiagnostic; if (!utc(v.attemptedAt) || v.operation !== v.issue.operation) return "diagnostic time/issue operation mismatch"; return; }
  if (name === "CollectiveObservation") { const v = value as CollectiveObservation; if (!utc(v.observedAt) || new Set(v.reasons).size !== v.reasons.length) return "observation requires UTC time and unique reasons"; return; }
  if (name === "CollectiveFetchDomainSummary") { const v = value as CollectiveFetchDomainSummary; if (![v.lastAttemptAt, v.lastCompletedAt].every(utc) || (v.lastInstalledRevision != null && !decimal(v.lastInstalledRevision))) return "summary dates/revision are invalid"; return; }
  if (name === "CollectiveFetchDomainResult") return domainIssue(value as CollectiveFetchDomainResult);
  if (name === "CollectiveFetchReport") {
    const v = value as CollectiveFetchReport;
    if (!decimal(v.cacheRevision) || new Set(v.domains.map(d => d.domain)).size !== 3) return "report requires all independent domains and decimal revision";
    const requested = v.domains.filter(d => d.requested), complete = requested.filter(d => d.state === "complete").length;
    const progress = requested.some(d => d.state === "complete" || d.commit !== "retained");
    const outcome = requested.length > 0 && complete === requested.length ? "complete" : progress ? "partial" : "failed";
    if (v.outcome !== outcome || (outcome === "complete" && v.issues.length !== 0)) return "overall outcome disagrees with independently completed effects";
    return;
  }
  if (name === "CollectiveObservationResponse") { const v = value as CollectiveObservationResponse; if (!decimal(v.cacheRevision) || new Set(v.collectives.map(c => c.id)).size !== v.collectives.length || (v.fetch.latestReport != null && (v.fetch.latestReport.scope.villageUrl !== v.scope.villageUrl || v.fetch.latestReport.scope.userId !== v.scope.userId))) return "catalog revision/identity or latest report scope mismatch"; return; }
  if (name === "CollectiveFetchSummary") {
    const v = value as {lastCompletedFetchAt?: string | null;lastAttemptAt?: string | null}; if (!utc(v.lastCompletedFetchAt) || !utc(v.lastAttemptAt)) return "fetch summary dates must be UTC observations"; return;
  }
  if (name === "LocalContributionHistoryResponse") {
    const v = value as LocalContributionHistoryResponse;
    let identity: ContributionTraversalIdentity;
    try { identity = parseContributionTraversalRevision(v.traversalRevision); } catch { return "history token is malformed"; }
    if (identity.scope.villageUrl !== v.scope.villageUrl || identity.scope.userId !== v.scope.userId || identity.kind !== v.kind) return "history token scope/kind mismatch";
    const rows = v.kind === "remote-attempts" ? v.remoteAttempts : v.kind === "receipts" ? v.receipts : v.operationDiagnostics;
    if (rows.length > v.limit || rows.length > v.total) return "history rows exceed page limit or filtered total";
    return;
  }
}

function domainIssue(v: CollectiveFetchDomainResult): string | undefined {
  if (v.requested === (v.state === "not_requested") || ![v.startedAt, v.lastAttemptAt, v.completedAt, v.lastCompletedAt].every(utc) || (v.lastInstalledRevision != null && !decimal(v.lastInstalledRevision))) return "domain requested/state, dates or revision mismatch";
  if (v.state === "complete") { if (v.completedAt == null || v.lastCompletedAt == null || v.lastInstalledRevision == null || v.commit === "retained" || v.issue != null) return "complete domain needs installed completion evidence"; }
  else if (v.completedAt != null || (v.domain !== "associations" && v.commit !== "retained") || (v.state !== "not_requested" && v.issue == null)) return "incomplete domain must retain its previous complete dataset with an issue";
  for (const s of v.sourceOutcomes) if ((v.state === "complete" && !s.complete) || !utc(s.observedAt) || (s.outcome === "failed") === s.complete || (s.complete && s.observedAt == null) || ((s.outcome === "failed") !== (s.issue != null))) return "source completeness/time/issue mismatch";
  if (v.domain !== "associations" && v.associations != null) return "association scopes belong only to associations";
  if (v.domain === "associations" && v.requested && v.associations == null) return "requested associations require scoped results";
  if (!v.requested && (v.commit !== "retained" || v.issue != null || v.sourceOutcomes.length !== 0 || v.associations != null)) return "not-requested domain cannot claim effects/issues";
  if (v.associations != null) {
    const a = v.associations;
    if (a.scopes.length !== a.requested || a.completed + a.failed !== a.requested || new Set(a.scopes.map(s => s.transcriptId)).size !== a.scopes.length) return "association counts or identities disagree";
    let n = 0;
    for (const s of a.scopes) {
      if (s.state === "not_requested" || !utc(s.observedAt) || !utc(s.completedAt)) return "association state/date invalid";
      if (s.state === "complete") { n++; if (s.completedAt == null || s.observedAt == null || s.commit === "retained" || s.issue != null) return "completed association set needs actual committed observation"; }
      else if (s.issue == null || s.commit !== "retained" || s.completedAt != null) return "incomplete association set must retain prior evidence";
    }
    if (n !== a.completed || (v.state === "complete" && a.failed !== 0)) return "association completion count disagrees";
    const commit = a.scopes.some(s => s.commit === "updated") ? "updated" : n > 0 ? "unchanged" : "retained";
    if (v.state !== "complete" && v.commit !== commit) return "partial association domain lost actual committed effects";
  }
}

export function collectiveFetchHTTPStatus(report: CollectiveFetchReport): number {
  try {
  if (!scopeValid(report.scope) || !uuid.test(report.invocationId)) return 500;
  if (collectiveContractIssue("CollectiveFetchReport", report) || report.domains.some(d => domainIssue(d))) return 500;
  if (report.outcome === "complete") return 200;
  const issues = [...report.issues];
  for (const d of report.domains) { if (d.issue != null) issues.push(d.issue); for (const s of d.sourceOutcomes) if (s.issue != null) issues.push(s.issue); for (const s of d.associations?.scopes ?? []) if (s.issue != null) issues.push(s.issue); }
  const ranks = { persistence: 5, auth: 4, deadline: 3, remote: 2, contract: 2, superseded: 1 };
  const rank = issues.reduce((n, i) => Math.max(n, ranks[i.code]), 0);
  return [500, 409, 502, 504, 401, 500][rank] ?? 500;
  } catch { return 500; }
}
