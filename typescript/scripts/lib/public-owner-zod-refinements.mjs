// Opens the generated person validator to kinds a later contract version adds.
//
// VillagePublicOwner is closed for producers (the OpenAPI oneOf lists exactly
// profile and anonymous) and open for readers: the contract says a reader that
// meets a kind it does not know treats the person as anonymous. Hey API renders
// the oneOf as a closed z.union, so an older package would refuse a whole
// response the moment the server sends a new kind, which would make adding a
// kind a breaking change. This pass appends one reader arm that accepts any
// other kind and reads it as anonymous, carrying no field from the unknown
// shape. The known kinds come from the generated literals, never a second list.
const declarationPattern = /export const zVillagePublicOwner = z\.union\(\[\n[\s\S]*?\n\]\);/;

export function applyPublicOwnerZodRefinements(source, withheld) {
  const declaration = source.match(declarationPattern)?.[0];
  if (declaration === undefined) {
    throw new Error("TypeScript contract generation could not open zVillagePublicOwner to unknown kinds in src/internal/generated/contract/zod.gen.ts; the generated declaration no longer matches `z.union([...])`, so readers would refuse a response carrying a kind added later instead of showing the person as anonymous; inspect the pinned generator output and update scripts/lib/public-owner-zod-refinements.mjs.");
  }
  const kinds = [...declaration.matchAll(/kind: z\.literal\('([^']+)'\)/g)].map((match) => match[1]);
  if (kinds.length !== 2 || !kinds.includes("profile") || !kinds.includes("anonymous")) {
    throw new Error(`TypeScript contract generation found kinds ${JSON.stringify(kinds)} on zVillagePublicOwner; expected the declared profile and anonymous arms; inspect the generated union and update scripts/lib/public-owner-zod-refinements.mjs.`);
  }
  const reader = `,\n    z.object({\n        kind: z.string().min(1).refine((kind) => !${JSON.stringify(kinds)}.includes(kind))\n    }).transform((): { kind: 'anonymous' } => ({ kind: 'anonymous' }))\n]);`;
  let output = source.replace(declaration, `${declaration.slice(0, -"\n]);".length)}${reader}`);
  // An organization login can identify a person, so none may sit beside one
  // shown as anonymous. These mirror the Go Validate rules on the same shapes.
  output = refine(output, "zVillageTranscriptListRow", `if (value.owner.kind !== "profile" && (value.owner_orgs?.length ?? 0) > 0) fail("owner_orgs lists organizations beside an anonymous owner; emit none for an anonymous owner");`);
  output = refine(output, "zVillageTranscriptMetadataResponse", `if (value.owner.kind !== "profile" && value.owner_orgs.length > 0) fail("owner_orgs lists organizations beside an anonymous owner; emit none for an anonymous owner");`);
  output = refine(output, "zVillageCollectiveMember", `if (value.member.kind !== "profile" && value.github_orgs.length > 0) fail("github_orgs lists organizations beside an anonymous member; emit none for an anonymous member");`);
  // Repository fields identify the owner's repository, so none may sit beside
  // an anonymous owner. Withheld means null or absent: an empty string is a
  // value and is refused, exactly as the Go Validate refuses it. The field lists come from
  // testdata/village_public_owner.yaml, the list the Go test also checks.
  const fields = (holder) => {
    const list = withheld?.[holder];
    if (!Array.isArray(list) || list.length === 0) throw new Error(`TypeScript contract generation found no withheld repository fields for ${holder} in testdata/village_public_owner.yaml; the package would accept a repository beside an anonymous owner; restore the list.`);
    return JSON.stringify(list);
  };
  const withheldCheck = (owner, record, holder) => `if (${owner}.kind !== "profile") { const present = ${fields(holder)}.filter((field) => { const v = (${record} as Record<string, unknown>)[field]; return v !== null && v !== undefined; }); if (present.length > 0) fail(present.join(", ") + " identify the repository of an anonymous owner; withhold every repository field (null, or omitted on the pull envelope) when the owner is anonymous"); }`;
  output = refine(output, "zVillageTranscriptListRow", withheldCheck("value.owner", "value.transcript", "VillageTranscript"));
  output = refine(output, "zVillageTranscriptMetadataResponse", withheldCheck("value.owner", "value.transcript", "VillageTranscript"));
  output = refine(output, "zVillageSessionRow", withheldCheck("value.owner", "value.session", "VillageTranscript"));
  output = refine(output, "zVillageCollectiveTranscript", withheldCheck("value.owner", "value", "VillageCollectiveTranscript"));
  output = refine(output, "zVillagePendingShare", withheldCheck("value.owner", "value", "VillagePendingShare"));
  output = refine(output, "zPullTranscriptInfo", withheldCheck("value.owner", "value", "PullTranscriptInfo"));
  return output;
}

function refine(source, name, body) {
  const start = source.indexOf(`export const ${name} = z.object({`);
  const typeDeclaration = start < 0 ? -1 : source.indexOf(`\n\nexport type ${name.slice(1)} =`, start);
  const end = typeDeclaration < 0 ? -1 : source.lastIndexOf(";", typeDeclaration);
  if (start < 0 || end < start) {
    throw new Error(`TypeScript contract generation could not keep anonymous people anonymous on ${name} in src/internal/generated/contract/zod.gen.ts; the generated object declaration was not found, so the package would accept identifying data beside an anonymous person; update scripts/lib/public-owner-zod-refinements.mjs.`);
  }
  return source.slice(0, end) + `.superRefine((value, context) => {
    const fail = (reason: string) => context.addIssue({ code: "custom", message: "schema.${name}: " + reason });
    ${body}
  })` + source.slice(end);
}

// Pull envelope 0.2.0 blanks a person anonymous to the caller to an empty id
// beside the anon username, and omits the project name beside an anonymous
// owner. These checks mirror the Go Validate on the Legacy* types; the
// sentinel comes from testdata/pull/legacy_envelope.yaml, the value the Go
// constant is tested against.
export function applyLegacyPullZodRefinements(source, anonymousUsername) {
  if (typeof anonymousUsername !== "string" || anonymousUsername === "") {
    throw new Error("TypeScript contract generation found no anonymous_username in testdata/pull/legacy_envelope.yaml; the legacy pull validators would accept a hidden person's name; restore the sentinel.");
  }
  const sentinel = JSON.stringify(anonymousUsername);
  const person = (id, username) => `if (${id} === "" && ${username} !== ${sentinel}) fail("an empty id carries a username; a hidden person would be named; send ${anonymousUsername} beside an empty id"); if (${id} !== "" && ${username} === "") fail("an id carries no username; send the profile's username, or an empty id with ${anonymousUsername}");`;
  let output = refine(source, "zLegacyPullTranscriptInfo", `${person("value.ownerUserId", "value.ownerUsername")} if (value.ownerUserId === "" && (value.projectName ?? "") !== "") fail("projectName sits beside an anonymous owner; omit it for an anonymous owner");`);
  output = refine(output, "zLegacyPullAnnotation", person("value.authorUserId", "value.authorUsername"));
  return output;
}
