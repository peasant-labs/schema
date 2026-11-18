// Closed sets keep their runtime Zod schema but report the named set type.
//
// Hey API emits `export const zRole = z.enum([...])`, so every object field that
// uses it is inferred as an anonymous literal union and the declared alias
// `Role` is lost. This pass wraps each catalogued closed-set schema in
// `closedSet<Role>()`, which returns the same runtime object typed so that
// `_zod.output` / `_zod.input` are exactly the facade alias from enums.gen.ts.
// Object fields, arrays and `z.infer` then carry `Role` instead of a widened
// union, and the schema keeps its enum API (`options`, `extract`, ...).
//
// The helper fails compilation if the schema's accepted values and the facade
// table ever differ, which replaces the old load-time `zRole.parse(value)`
// check on every member.
const HELPER = `import type * as ClosedSets from "../enums.gen.js";

type ClosedSet<S extends z.ZodType, T> = Omit<S, "_zod"> & { _zod: Omit<S["_zod"], "input" | "output"> & { input: T; output: T } };
const closedSet = <T>() => <S extends z.ZodType<T>>(schema: S & ([T] extends [z.output<S>] ? unknown : never)): ClosedSet<S, T> => schema as unknown as ClosedSet<S, T>;
`;

export function applyClosedSetZodTypes(source, catalog) {
  const importMatch = /^import \* as z from 'zod';\n/m.exec(source);
  if (importMatch === null) {
    throw new Error("TypeScript closed-set generation could not find \"import * as z from 'zod';\" in src/internal/generated/contract/zod.gen.ts; the named set types cannot be attached to the Zod schemas; keep the Hey API zod import in that form or update typescript/scripts/lib/closed-set-zod-types.mjs.");
  }
  let rendered = source;
  for (const { name } of catalog.enums) {
    const pattern = new RegExp(`^export const z${name} = ((?:(?!\\nexport )[\\s\\S])*?);\\n(\\n?export type ${name} = z\\.infer<typeof z${name}>;)$`, "m");
    const match = pattern.exec(rendered);
    if (match === null) {
      throw new Error(`TypeScript closed-set generation could not find "export const z${name} = ...;" followed by "export type ${name} = z.infer<typeof z${name}>;" in src/internal/generated/contract/zod.gen.ts; z${name} would keep an anonymous literal union instead of the named ${name} set; keep the Hey API enum output in that form or update typescript/scripts/lib/closed-set-zod-types.mjs.`);
    }
    rendered = rendered.replace(pattern, () => `export const z${name} = closedSet<ClosedSets.${name}>()(${match[1]});\n${match[2]}`);
  }
  const at = importMatch.index + importMatch[0].length;
  return `${rendered.slice(0, at)}${HELPER}${rendered.slice(at)}`;
}
