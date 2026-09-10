// Read-only display items retain route rows and immediate-owner groups. A member
// page reuses those items but can contain only saved transcript identities.
export function applyGroupedReadZodRefinements(source) {
  let output = source;
  output = append(output, "zHelperGroupSummary", `if (value.groupId === "" || value.memberScope === "" || value.purpose !== "helper_review") fail("groupId/memberScope must identify an immediate helper_review group; refresh the original scoped list");`);
  output = append(output, "zHelperContextSummary", `if (value.groupId === "") fail("context groupId must identify its scoped group; refresh the original list");`);
  output = append(output, "zLocalSessionRow", `
    if (value.session.id === "") fail("session.id is empty; provide the saved transcript identity");
    if (value.sync && value.matches?.length) fail("sync and matches select incompatible route arms; retain the originating route only");
    if (value.sync) {
      const s = value.session, v = value.sync;
      if (v.id !== s.id || v.harness !== s.harness || v.projectName !== s.project || v.projectHash !== s.projectHash || v.totalTokens !== s.totalTokens || v.turnCount !== s.turnCount || v.inputSubmissionCount !== s.inputSubmissionCount) fail("sync identity or count mirrors disagree; derive both from the same row");
    }
    if (value.matches?.some(match => match.sessionId !== value.session.id)) fail("search match belongs to another session; retain only this transcript's evidence");`);
  output = append(output, "zVillageSessionRow", `
    const s = value.session;
    const arms = [value.collective, value.pending, value.myShare, value.contributable].filter(v => v != null);
    if (arms.length > 1) fail("multiple route arms supplied; retain the originating route only");
    const equal = (a: unknown, b: unknown): boolean => {
      if (a === b) return true;
      if (a === null || b === null || typeof a !== "object" || typeof b !== "object") return false;
      const left = a as Record<string, unknown>, right = b as Record<string, unknown>;
      return Object.keys(left).length === Object.keys(right).length && Object.keys(left).every(k => Object.hasOwn(right, k) && equal(left[k], right[k]));
    };
    const check = (v: unknown, keys: string[], aliases: Record<string, string> = {}) => {
      if (v == null) return;
      const row = v as Record<string, unknown>, session = s as unknown as Record<string, unknown>;
      for (const key of keys) if (!equal(row[key], session[aliases[key] ?? key])) fail("route identity or metric mirrors disagree; derive both from the same row");
    };
    const graph = ["parent_session_id", "root_session_id", "purpose", "relationships", "input_submission_count"];
    check(value.collective, [...graph, "id", "owner_id", "local_id", "title", "description", "visibility", "model_provider", "model_name", "harness_version", "turn_count", "token_count", "tokens_in", "tokens_out", "duration_ms", "project_hash", "project_name", "project_display_name", "project_name_source", "project_remote_label", "git_branch", "session_origin"]);
    check(value.pending, [...graph, "transcript_id", "owner_id", "local_id", "title", "model_provider", "project_hash", "project_name", "branch"], { transcript_id: "id", branch: "git_branch" });
    check(value.myShare, [...graph, "id", "owner_id", "local_id", "title", "visibility", "published_at", "model_provider", "model_name", "turn_count", "tokens_in", "tokens_out"]);
    check(value.contributable, [...graph, "id", "local_id", "title", "visibility", "model_provider", "project_hash", "project_display_name", "project_name_source", "git_branch", "published_at", "session_origin"]);`);
  for (const prefix of ["Local", "Village"]) {
    output = append(output, `z${prefix}SessionListItem`, `
      if ((value.kind === "transcript") !== (value.transcript != null) || (value.kind === "context_container") !== (value.context != null)) fail("kind must select exactly one matching transcript/context arm; omit the other arm");
      const groups = new Set<string>();
      for (const group of value.helperGroups ?? []) {
        if (groups.has(group.groupId)) fail("duplicate groupId prevents independent paging; emit each immediate group once");
        groups.add(group.groupId);
      }`);
    output = append(output, `z${prefix}HelperMembersPayload`, `
      if (value.members.length > value.limit || value.members.length > value.total) fail("member count exceeds page limit or direct total; count direct saved helpers before paging");
      const ids = new Set<string>();
      const groups = new Set<string>();
      for (const member of value.members) {
        if (member.kind !== "transcript" || member.transcript == null || member.context != null) { fail("members must be transcript items, never context containers; preserve nested groups on the item"); continue; }
        const id = member.transcript.session.id;
        if (ids.has(id)) fail("duplicate member identity would double count a saved helper; emit each transcript once");
        ids.add(id);
        for (const group of member.helperGroups ?? []) {
          if (groups.has(group.groupId)) fail("groupId appears under multiple members; emit each immediate-owner group once");
          groups.add(group.groupId);
        }
      }`);
  }
  return output;
}

function append(source, name, body) {
  const start = source.indexOf(`export const ${name} = z.object({`);
  const end = source.indexOf(";", source.indexOf("\n})", start));
  if (start < 0 || end < start) throw new Error(`grouped read generation failed at ${name}: expected object declaration not found; no validator was produced; update the generator for the pinned Zod output`);
  return source.slice(0, end) + `.superRefine((value, context) => {
    const fail = (reason: string) => context.addIssue({ code: "custom", message: "schema.${name} during grouped read validation: " + reason });
    ${body.trim()}
  })` + source.slice(end);
}
