// Relational canonical-address rule shared with UnverifiedPullRequest.Validate.
export function applyUnverifiedPullRequestZodRefinements(source) {
  const start = source.indexOf("export const zUnverifiedPullRequest = z.object({");
  const end = source.indexOf("\n\nexport type UnverifiedPullRequest =", start);
  if (start < 0 || end < 0) throw new Error("Unverified pull request generation cannot find the root validator; inspect generator output before release");
  const raw = source.slice(start, end);
  const refinement = `.superRefine((value, context) => {
    const segment = /^[A-Za-z0-9_-][A-Za-z0-9._-]*$/;
    const validSegment = (part: string) => segment.test(part) && part !== "." && part !== "..";
    if (!validSegment(value.owner) || !validSegment(value.repo) || value.url !== "https://github.com/" + value.owner + "/" + value.repo + "/pull/" + String(value.number)) {
      context.addIssue({ code: "custom", message: "Unverified pull request identity and canonical GitHub URL must agree" });
    }
    if (value.source.toolCallId !== undefined && /[\\uD800-\\uDFFF]/u.test(value.source.toolCallId)) {
      context.addIssue({ code: "custom", message: "Recorded tool call ID must be valid Unicode" });
    }
  })`;
  if (raw.endsWith(`${refinement};`)) return source;
  if (!raw.endsWith("});") || raw.includes(".superRefine(")) throw new Error("Unverified pull request validator changed shape; cannot safely enforce canonical identity");
  return source.slice(0,start) + raw.slice(0,-1) + refinement + ";" + source.slice(end);
}
