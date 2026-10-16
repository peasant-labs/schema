import { zCollectiveFetchReport } from "./internal/generated/contract/zod.gen.js";

// Keep the public raw boundary above the generated parser, not inside its
// refinement module: Zod -> pure semantics, public status -> complete Zod.
// No partial parallel validator or runtime import cycle is needed.
export function collectiveFetchHTTPStatus(value: unknown): number {
  try {
    const parsed = zCollectiveFetchReport.safeParse(value);
    if (!parsed.success) return 500;
    const report = parsed.data;
    if (report.outcome === "complete") return 200;
    const issues = [...report.issues];
    for (const domain of report.domains) {
      if (domain.issue != null) issues.push(domain.issue);
      for (const source of domain.sourceOutcomes) if (source.issue != null) issues.push(source.issue);
      for (const scope of domain.associations?.scopes ?? []) if (scope.issue != null) issues.push(scope.issue);
    }
    const ranks = { persistence: 5, auth: 4, deadline: 3, remote: 2, contract: 2, superseded: 1 };
    const rank = issues.reduce((highest, issue) => Math.max(highest, ranks[issue.code]), 0);
    return [500, 409, 502, 504, 401, 500][rank] ?? 500;
  } catch { return 500; }
}
