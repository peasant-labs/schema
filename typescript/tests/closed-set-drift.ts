import * as z from "zod";

// Keep this helper identical to the generated one. The alias test checks that
// the copy stays in sync, so these negative cases exercise the production check.
type ClosedSet<S extends z.ZodType, T> = Omit<S, "_zod"> & { _zod: Omit<S["_zod"], "input" | "output"> & { input: T; output: T } };
const closedSet = <T>() => <S extends z.ZodType<T>>(schema: S & ([T] extends [z.output<S>] ? unknown : never)): ClosedSet<S, T> => schema as unknown as ClosedSet<S, T>;

type Example = "first" | "second";
closedSet<Example>()(z.enum(["first", "second"]));
// @ts-expect-error A schema missing a table member must not compile.
closedSet<Example>()(z.enum(["first"]));
// @ts-expect-error A schema with an extra member must not compile.
closedSet<Example>()(z.enum(["first", "second", "third"]));
