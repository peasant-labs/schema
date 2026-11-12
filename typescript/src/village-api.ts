import type { operations as GeneratedOperations, paths as GeneratedPaths } from "./internal/generated/village-api.js";

export { VillageAPIVersion } from "./index.js";
export { ALL_HTTP_METHODS, ALL_VILLAGE_ROUTES, HttpMethod, VillageRoute, isHttpMethod } from "./internal/generated/routes.gen.js";
export type { PathArgs, PathParams, RouteMember } from "./route-member.js";
export type paths = GeneratedPaths;
export type operations = GeneratedOperations;
