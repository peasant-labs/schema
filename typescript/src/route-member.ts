import type { HttpMethod } from "./internal/generated/routes.gen.js";

// PathParamNames is the union of the {name} parameters in a path template.
type PathParamNames<P extends string> = P extends `${string}{${infer Name}}${infer Rest}` ? Name | PathParamNames<Rest> : never;

/** PathParams is the parameter object of a path template: one key per {name}. */
export type PathParams<P extends string> = { readonly [Name in PathParamNames<P>]: string | number };

/**
 * PathArgs is the argument list of a route member's path builder: no argument
 * when the template has no parameter, otherwise one object carrying exactly the
 * template's parameters. A missing or extra key is a compile error.
 */
export type PathArgs<P extends string> = [PathParamNames<P>] extends [never] ? [] : [params: PathParams<P>];

/** RouteMember is one operation a Village API or Local API document declares. */
export interface RouteMember<P extends string, M extends HttpMethod> {
  /** The HTTP method, as sent on the wire. */
  readonly method: M;
  /** The document's path template, for example "/api/v1/pulls/{owner}/{name}/{number}". */
  readonly pattern: P;
  /** The document's operationId, for example "getPullRequestAttachment". */
  readonly operationId: string;
  /** Builds the request path. Every parameter value is escaped with encodeURIComponent. */
  path(...args: PathArgs<P>): string;
}

type PathPart = { readonly literal: string } | { readonly parameter: string };

const parameterPattern = /\{([^{}]+)\}/g;

/**
 * routeMember builds one frozen member. The generated route modules are its
 * only callers; the template is split once, when the module loads.
 */
export function routeMember<P extends string, M extends HttpMethod>(method: M, pattern: P, operationId: string): RouteMember<P, M> {
  const parts: PathPart[] = [];
  let offset = 0;
  for (const match of pattern.matchAll(parameterPattern)) {
    parts.push({ literal: pattern.slice(offset, match.index) });
    parts.push({ parameter: match[1] ?? "" });
    offset = match.index + match[0].length;
  }
  parts.push({ literal: pattern.slice(offset) });
  const frozenParts = Object.freeze(parts.map((part) => Object.freeze(part)));

  function path(...args: PathArgs<P>): string {
    // A JavaScript caller can pass anything; a value that is not an object
    // carries no parameter, so each parameter is then refused as missing.
    const given: unknown = args[0];
    const params: Readonly<Record<string, unknown>> | undefined = typeof given === "object" && given !== null ? args[0] : undefined;
    let out = "";
    for (const part of frozenParts) {
      if ("literal" in part) {
        out += part.literal;
        continue;
      }
      const value = params === undefined || !Object.hasOwn(params, part.parameter) ? undefined : params[part.parameter];
      out += encodeURIComponent(pathValue(operationId, pattern, part.parameter, value));
    }
    return out;
  }

  return Object.freeze({ method, pattern, operationId, path });
}

function pathValue(operationId: string, pattern: string, parameter: string, value: unknown): string {
  if (typeof value === "string") return value;
  if (typeof value === "number" && Number.isFinite(value)) return String(value);
  const received = value === undefined ? "no value" : typeof value === "number" ? String(value) : `a ${value === null ? "null" : typeof value}`;
  throw new TypeError(
    `route path for ${operationId} (${pattern}) could not be built: parameter ${JSON.stringify(parameter)} received ${received}; ` +
      "a path parameter must be a string or a finite number, so the request would address the wrong resource; " +
      `pass ${JSON.stringify(parameter)} in the object given to path().`,
  );
}
