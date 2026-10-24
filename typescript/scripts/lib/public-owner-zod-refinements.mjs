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

export function applyPublicOwnerZodRefinements(source) {
  const declaration = source.match(declarationPattern)?.[0];
  if (declaration === undefined) {
    throw new Error("TypeScript contract generation could not open zVillagePublicOwner to unknown kinds in src/internal/generated/contract/zod.gen.ts; the generated declaration no longer matches `z.union([...])`, so readers would refuse a response carrying a kind added later instead of showing the person as anonymous; inspect the pinned generator output and update scripts/lib/public-owner-zod-refinements.mjs.");
  }
  const kinds = [...declaration.matchAll(/kind: z\.literal\('([^']+)'\)/g)].map((match) => match[1]);
  if (kinds.length !== 2 || !kinds.includes("profile") || !kinds.includes("anonymous")) {
    throw new Error(`TypeScript contract generation found kinds ${JSON.stringify(kinds)} on zVillagePublicOwner; expected the declared profile and anonymous arms; inspect the generated union and update scripts/lib/public-owner-zod-refinements.mjs.`);
  }
  const reader = `,\n    z.object({\n        kind: z.string().min(1).refine((kind) => !${JSON.stringify(kinds)}.includes(kind))\n    }).transform((): { kind: 'anonymous' } => ({ kind: 'anonymous' }))\n]);`;
  return source.replace(declaration, `${declaration.slice(0, -"\n]);".length)}${reader}`);
}
