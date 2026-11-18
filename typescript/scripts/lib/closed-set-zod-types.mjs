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
//
// `ClosedSet` is a mapped type (`Omit`), and a mapped type resolves Zod's
// `this`-typed wrappers against the unwrapped schema: `optional(): ZodOptional<this>`
// would return `ZodOptional<ZodEnum<...>>` and drop the alias from every optional
// field. So the helper re-declares the wrappers Hey API emits on a named schema
// (`optional`, `nullable`, `nullish`) to wrap the `ClosedSet` itself, and the
// pass fails generation if the output calls any other method on a closed-set
// schema, because that method would silently drop the alias the same way.
// `Extract<..., z.ZodType>` defers the wrapper's constraint check until the
// helper is instantiated; it resolves to the `ClosedSet` itself.
const HELPER = `import type * as ClosedSets from "../enums.gen.js";

type ClosedSet<S extends z.ZodType, T> = Omit<S, "_zod" | "optional" | "nullable" | "nullish"> & { _zod: Omit<S["_zod"], "input" | "output"> & { input: T; output: T }; optional(): z.ZodOptional<Extract<ClosedSet<S, T>, z.ZodType>>; nullable(): z.ZodNullable<Extract<ClosedSet<S, T>, z.ZodType>>; nullish(): z.ZodOptional<z.ZodNullable<Extract<ClosedSet<S, T>, z.ZodType>>> };
const closedSet = <T>() => <S extends z.ZodType<T>>(schema: S & ([T] extends [z.output<S>] ? unknown : never)): ClosedSet<S, T> => schema as unknown as ClosedSet<S, T>;
`;

// The Zod wrappers the ClosedSet helper re-declares so they keep the named set.
const RE_DECLARED_WRAPPERS = new Set(["optional", "nullable", "nullish"]);

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
  const names = catalog.enums.map(({ name }) => name).join("|");
  for (const call of rendered.matchAll(new RegExp(`\\bz(${names})\\.([A-Za-z_$][\\w$]*)\\(`, "g"))) {
    if (!RE_DECLARED_WRAPPERS.has(call[2])) {
      throw new Error(`TypeScript closed-set generation found "z${call[1]}.${call[2]}(" in src/internal/generated/contract/zod.gen.ts; the ClosedSet helper re-declares only ${[...RE_DECLARED_WRAPPERS].map((method) => `.${method}()`).join(", ")}, so .${call[2]}() would resolve against the unwrapped schema and the field would lose the named ${call[1]} set; add .${call[2]}() to the ClosedSet helper in typescript/scripts/lib/closed-set-zod-types.mjs (and to tests/closed-set-drift.ts) with a return type that wraps the ClosedSet itself.`);
    }
  }
  const at = importMatch.index + importMatch[0].length;
  return `${rendered.slice(0, at)}${HELPER}${rendered.slice(at)}`;
}
