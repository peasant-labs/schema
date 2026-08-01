// Compile-time proof that the ProjectHash nominal brand actually propagates to
// every wire location that carries a project identity, rather than degrading
// back to plain string at one specific field. Each assertion below is tagged
// with a "projectHashLocation:<tag>" comment matched by the companion runtime
// test project-hash-locations.test.mjs against
// testdata/typescript/project_hash_locations.yaml: dropping an assertion here
// typechecks fine on its own (removing a passing check is not a type error),
// so that coupling test is the only thing that fails if a location silently
// loses its brand or this file silently loses its assertion.
import type { AnnotationPushItem, ProjectHash } from "@peasant-labs/schema";
import type { operations as LocalOperations } from "@peasant-labs/schema/local-api";
import type { operations as VillageOperations } from "@peasant-labs/schema/village-api";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type VillagePublishRequest = NonNullable<VillageOperations["publishTranscript"]["requestBody"]>["content"]["multipart/form-data"]["metadata"];
type VillagePullAnnotations = VillageOperations["getPullTranscriptAnnotations"]["responses"][200]["content"]["application/json"];
type VillagePullAnnotation = NonNullable<VillagePullAnnotations>[number];
// A same-spelling, unconstrained interface field: proves property spelling
// alone must never confer nominal identity (the negative control).
interface SameSpellingUnconstrained { projectHash: string; targetProjectHash?: string }

// projectHashLocation:root-annotation-push
const rootAnnotationPushProjectHash: Same<Exclude<AnnotationPushItem["projectHash"], null | undefined>, ProjectHash> = true;
void rootAnnotationPushProjectHash;

// projectHashLocation:local-path
const localReviewChangesPathProjectHash: Same<LocalOperations["listReviewChanges"]["parameters"]["path"]["projectHash"], ProjectHash> = true;
void localReviewChangesPathProjectHash;

// projectHashLocation:local-response
const localResolveProjectResponseProjectHash: Same<LocalOperations["resolveProject"]["responses"][200]["content"]["application/json"]["projectHash"], ProjectHash> = true;
void localResolveProjectResponseProjectHash;

// projectHashLocation:village-request
const villagePublishRequestProjectHash: Same<NonNullable<VillagePublishRequest["project"]>["hash"], ProjectHash> = true;
void villagePublishRequestProjectHash;

// projectHashLocation:village-response
const villagePullAnnotationProjectHash: Same<Exclude<VillagePullAnnotation["targetProjectHash"], null | undefined>, ProjectHash> = true;
void villagePullAnnotationProjectHash;

// projectHashLocation:same-spelling-string
const sameSpellingUnconstrainedProjectHash: Same<SameSpellingUnconstrained["projectHash"], string> = true;
void sameSpellingUnconstrainedProjectHash;
