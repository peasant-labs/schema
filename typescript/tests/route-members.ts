import {
  ALL_LOCAL_ROUTES,
  HttpMethod as LocalHttpMethod,
  LocalRoute,
  type paths as LocalPaths,
} from "@peasant-labs/schema/local-api";
import {
  ALL_HTTP_METHODS,
  ALL_VILLAGE_ROUTES,
  HttpMethod,
  VillageRoute,
  isHttpMethod,
  type PathArgs,
  type RouteMember,
  type paths as VillagePaths,
} from "@peasant-labs/schema/village-api";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;

// Drift, both directions: the members of each API are exactly the operations of
// its document, as "<METHOD> <pattern>" keys. A document operation with no
// member, or a member with no document operation, makes these false.
type MemberOf<T> = T[keyof T];
type OperationKey<R> = R extends RouteMember<infer P, infer M> ? `${M} ${P}` : never;
type DocumentMethods<Item> = Exclude<keyof Item, "parameters"> & string;
type DocumentKeys<Paths> = {
  [P in keyof Paths & string]: {
    [M in DocumentMethods<Paths[P]>]-?: [Exclude<Paths[P][M], undefined>] extends [never] ? never : `${Uppercase<M>} ${P}`;
  }[DocumentMethods<Paths[P]>];
}[keyof Paths & string];

const villageMembersAreTheDocumentOperations: Same<OperationKey<MemberOf<typeof VillageRoute>>, DocumentKeys<VillagePaths>> = true;
const localMembersAreTheDocumentOperations: Same<OperationKey<MemberOf<typeof LocalRoute>>, DocumentKeys<LocalPaths>> = true;
const villagePatternsAreTheDocumentPaths: Same<MemberOf<typeof VillageRoute>["pattern"], keyof VillagePaths> = true;
const localPatternsAreTheDocumentPaths: Same<MemberOf<typeof LocalRoute>["pattern"], keyof LocalPaths> = true;
const allVillageRoutesCoverEveryMember: Same<(typeof ALL_VILLAGE_ROUTES)[number], MemberOf<typeof VillageRoute>> = true;
const allLocalRoutesCoverEveryMember: Same<(typeof ALL_LOCAL_ROUTES)[number], MemberOf<typeof LocalRoute>> = true;
const allHttpMethodsCoverHttpMethod: Same<(typeof ALL_HTTP_METHODS)[number], HttpMethod> = true;
const bothSubpathsShareOneMethodSet: Same<LocalHttpMethod, HttpMethod> = true;
void [
  villageMembersAreTheDocumentOperations,
  localMembersAreTheDocumentOperations,
  villagePatternsAreTheDocumentPaths,
  localPatternsAreTheDocumentPaths,
  allVillageRoutesCoverEveryMember,
  allLocalRoutesCoverEveryMember,
  allHttpMethodsCoverHttpMethod,
  bothSubpathsShareOneMethodSet,
];

// A member carries its literal method and pattern.
const attachment = VillageRoute.GetPullRequestAttachment;
const attachmentMethod: "GET" = attachment.method;
const attachmentPattern: "/api/v1/pulls/{owner}/{name}/{number}" = attachment.pattern;
// @ts-expect-error The attachment read is a GET, not a POST.
const attachmentIsNotAPost: "POST" = attachment.method;
const attachmentUsesTheGetMember: typeof HttpMethod.Get = attachment.method;
void [attachmentMethod, attachmentPattern, attachmentIsNotAPost, attachmentUsesTheGetMember];

// The path builder takes exactly the template's parameters.
const attachmentPath: string = attachment.path({ owner: "acme", name: "tool", number: 12 });
// @ts-expect-error number is a parameter of /api/v1/pulls/{owner}/{name}/{number}.
attachment.path({ owner: "acme", name: "tool" });
// @ts-expect-error branch is not a parameter of /api/v1/pulls/{owner}/{name}/{number}.
attachment.path({ owner: "acme", name: "tool", number: 12, branch: "main" });
// @ts-expect-error A route with parameters needs its parameter object.
VillageRoute.GetTranscriptSegments.path();
// @ts-expect-error A parameter value is a string or a number.
VillageRoute.GetTranscriptSegments.path({ id: true });
const segmentsPath: string = VillageRoute.GetTranscriptSegments.path({ id: "transcript" });
const versionPath: string = VillageRoute.GetSchemaVersion.path();
// @ts-expect-error A route without parameters takes no argument.
VillageRoute.GetSchemaVersion.path({ id: "transcript" });
const localPath: string = LocalRoute.GetSessionTranscriptSegment.path({ id: "session", index: 0 });
const noArgs: Same<PathArgs<"/api/v1/health">, []> = true;
const oneObject: Same<PathArgs<"/a/{x}/b/{y}">, [params: { readonly x: string | number; readonly y: string | number }]> = true;
void [attachmentPath, segmentsPath, versionPath, localPath, noArgs, oneObject];

// Members are read-only.
// @ts-expect-error A member's pattern cannot be reassigned.
VillageRoute.GetSchemaVersion.pattern = "/api/v1/other";
// @ts-expect-error A member cannot be replaced.
VillageRoute.GetSchemaVersion = VillageRoute.GetTranscriptSegments;

declare const unknownMethod: unknown;
if (isHttpMethod(unknownMethod)) {
  const narrowed: HttpMethod = unknownMethod;
  void narrowed;
}
