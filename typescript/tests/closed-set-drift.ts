import * as z from "zod";

// Keep this helper identical to the generated one. The alias test checks that
// the copy stays in sync, so these negative cases exercise the production check.
type ClosedSet<S extends z.ZodType, T> = Omit<S, "_zod" | "optional" | "nullable" | "nullish"> & { _zod: Omit<S["_zod"], "input" | "output"> & { input: T; output: T }; optional(): z.ZodOptional<Extract<ClosedSet<S, T>, z.ZodType>>; nullable(): z.ZodNullable<Extract<ClosedSet<S, T>, z.ZodType>>; nullish(): z.ZodOptional<z.ZodNullable<Extract<ClosedSet<S, T>, z.ZodType>>> };
const closedSet = <T>() => <S extends z.ZodType<T>>(schema: S & ([T] extends [z.output<S>] ? unknown : never)): ClosedSet<S, T> => schema as unknown as ClosedSet<S, T>;

type Example = "first" | "second";
closedSet<Example>()(z.enum(["first", "second"]));
// @ts-expect-error A schema missing a table member must not compile.
closedSet<Example>()(z.enum(["first"]));
// @ts-expect-error A schema with an extra member must not compile.
closedSet<Example>()(z.enum(["first", "second", "third"]));

// Optional, nullable and nullish fields keep the named set: the wrapper wraps
// the ClosedSet, not the unwrapped enum schema behind it.
type ExactType<A, B> = (<U>() => U extends A ? 1 : 2) extends (<U>() => U extends B ? 1 : 2) ? true : false;
const example = closedSet<Example>()(z.enum(["first", "second"]));
const optionalKeepsSet: ExactType<z.output<ReturnType<typeof example.optional>>, Example | undefined> = true;
const nullableKeepsSet: ExactType<z.output<ReturnType<typeof example.nullable>>, Example | null> = true;
const nullishKeepsSet: ExactType<z.output<ReturnType<typeof example.nullish>>, Example | null | undefined> = true;
const optionalWrapsClosedSet: ExactType<ReturnType<typeof example.optional>, z.ZodOptional<typeof example>> = true;
void [optionalKeepsSet, nullableKeepsSet, nullishKeepsSet, optionalWrapsClosedSet];
