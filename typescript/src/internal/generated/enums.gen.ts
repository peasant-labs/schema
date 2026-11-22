// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
import { zVillageAttestationType } from "./contract/zod.gen.js";
import { zTranscriptEntryReferenceKind } from "./contract/zod.gen.js";
import { zTranscriptReadTransportKind } from "./contract/zod.gen.js";
import { zTranscriptRecordedIndexState } from "./contract/zod.gen.js";
import { zTranscriptRetainedState } from "./contract/zod.gen.js";
import { zTranscriptViewerProjection } from "./contract/zod.gen.js";
import { zTranscriptReadValidation } from "./contract/zod.gen.js";
import { zTranscriptNativePresence } from "./contract/zod.gen.js";
import { zTranscriptUsagePresence } from "./contract/zod.gen.js";
import { zTranscriptSegmentPartitionKind } from "./contract/zod.gen.js";
import { zCollectiveObservationReason } from "./contract/zod.gen.js";
import { zCollectiveObservationState } from "./contract/zod.gen.js";
import { zCollectiveOperation } from "./contract/zod.gen.js";
import { zOperationIssueCode } from "./contract/zod.gen.js";
import { zLocalContributionEvidenceKind } from "./contract/zod.gen.js";
import { zCollectiveEvidenceCoverage } from "./contract/zod.gen.js";
import { zLocalConfirmedContributionOutcome } from "./contract/zod.gen.js";
import { zLocalOperationDiagnosticOutcome } from "./contract/zod.gen.js";
import { zPullAssociationMetadataState } from "./contract/zod.gen.js";
import { zCollectiveFetchDomain } from "./contract/zod.gen.js";
import { zCollectiveFetchState } from "./contract/zod.gen.js";
import { zCollectiveFetchCommit } from "./contract/zod.gen.js";
import { zCollectiveFetchOutcome } from "./contract/zod.gen.js";
import { zCollectiveObservationSource } from "./contract/zod.gen.js";
import { zCollectiveSourceOutcome } from "./contract/zod.gen.js";
import { zLocalSyncPublicationFilter } from "./contract/zod.gen.js";
import { zPageSize } from "./contract/zod.gen.js";
import { zPublishOperationKind } from "./contract/zod.gen.js";
import { zVisibilityIntent } from "./contract/zod.gen.js";
import { zAnnotationAxis } from "./contract/zod.gen.js";
import { zAnnotationDatatype } from "./contract/zod.gen.js";
import { zAnnotationPushStatus } from "./contract/zod.gen.js";
import { zAnnotationStatus } from "./contract/zod.gen.js";
import { zAnnotatorKind } from "./contract/zod.gen.js";
import { zAssociationConclusion } from "./contract/zod.gen.js";
import { zAssociationEvidenceKind } from "./contract/zod.gen.js";
import { zBuiltinCommand } from "./contract/zod.gen.js";
import { zChangeBinding } from "./contract/zod.gen.js";
import { zChannelTopic } from "./contract/zod.gen.js";
import { zConfidence } from "./contract/zod.gen.js";
import { zContentKind } from "./contract/zod.gen.js";
import { zDecayLevel } from "./contract/zod.gen.js";
import { zDiffLineKind } from "./contract/zod.gen.js";
import { zEdgeViolationKind } from "./contract/zod.gen.js";
import { zEntryType } from "./contract/zod.gen.js";
import { zTurnPartType } from "./contract/zod.gen.js";
import { zImageMediaType } from "./contract/zod.gen.js";
import { zOmittedReason } from "./contract/zod.gen.js";
import { zFileChangeStatus } from "./contract/zod.gen.js";
import { zHarness } from "./contract/zod.gen.js";
import { zInsightKind } from "./contract/zod.gen.js";
import { zInsightProvenance } from "./contract/zod.gen.js";
import { zInteractionType } from "./contract/zod.gen.js";
import { zLicense } from "./contract/zod.gen.js";
import { zMapNodeKind } from "./contract/zod.gen.js";
import { zMessageType } from "./contract/zod.gen.js";
import { zReadAttributionState } from "./contract/zod.gen.js";
import { zReadStateGrade } from "./contract/zod.gen.js";
import { zRewriteMethod } from "./contract/zod.gen.js";
import { zRewriteResolution } from "./contract/zod.gen.js";
import { zRole } from "./contract/zod.gen.js";
import { zScaleKind } from "./contract/zod.gen.js";
import { zSessionOrigin } from "./contract/zod.gen.js";
import { zSessionOutcome } from "./contract/zod.gen.js";
import { zSourceFormat } from "./contract/zod.gen.js";
import { zStopReason } from "./contract/zod.gen.js";
import { zTargetKind } from "./contract/zod.gen.js";
import { zToolCallKind } from "./contract/zod.gen.js";
import { zTypeOrigin } from "./contract/zod.gen.js";
import { zValueDomainKind } from "./contract/zod.gen.js";
import { zTranscriptUpdateLicense } from "./contract/zod.gen.js";
import { zTranscriptUpdateVisibility } from "./contract/zod.gen.js";
import { zVillageAssignableCollectiveRole } from "./contract/zod.gen.js";
import { zVillageContributionStatus } from "./contract/zod.gen.js";
import { zVillageCollectiveAcceptanceMode } from "./contract/zod.gen.js";
import { zVillageCollectiveDataAccess } from "./contract/zod.gen.js";
import { zVillageCollectiveRole } from "./contract/zod.gen.js";
import { zVillageCollectiveViewerRole } from "./contract/zod.gen.js";
import { zVillageContributableView } from "./contract/zod.gen.js";
import { zVillageContributionErrorCode } from "./contract/zod.gen.js";
import { zVillageProjectNameSource } from "./contract/zod.gen.js";
import { zVillageReviewDecision } from "./contract/zod.gen.js";
import { zVillageShareEventActor } from "./contract/zod.gen.js";
import { zVillageShareStatus } from "./contract/zod.gen.js";
import { zVillageTranscriptDeletionPolicy } from "./contract/zod.gen.js";
import { zVillageTranscriptVisibility } from "./contract/zod.gen.js";
import { zVisibility } from "./contract/zod.gen.js";
import { zUsageScope } from "./contract/zod.gen.js";
import { zUsageCompleteness } from "./contract/zod.gen.js";
import { zNativeMetadataKind } from "./contract/zod.gen.js";
import { zNativeMetadataSourceType } from "./contract/zod.gen.js";
import { zNativePiMessageRole } from "./contract/zod.gen.js";
import { zDigestItemKind } from "./contract/zod.gen.js";
import { zVillagePullRequestAttachmentState } from "./contract/zod.gen.js";
import { zVillagePromptsCheckMode } from "./contract/zod.gen.js";
import { zVillagePublicOwnerKind } from "./contract/zod.gen.js";
import { zSessionRelationshipKind } from "./contract/zod.gen.js";
import { zRelationshipTargetState } from "./contract/zod.gen.js";
import { zEvidenceKind } from "./contract/zod.gen.js";
import { zSessionTeammateEvidence } from "./contract/zod.gen.js";
import { zSessionPurpose } from "./contract/zod.gen.js";
import { zContentOrigin } from "./contract/zod.gen.js";
import { zActorOrigin } from "./contract/zod.gen.js";
import { zDeliveryOrigin } from "./contract/zod.gen.js";
import { zContentOwnership } from "./contract/zod.gen.js";
import { zInputModality } from "./contract/zod.gen.js";
import { zPublicSourceAnchorKind } from "./contract/zod.gen.js";
import { zEarlierHistoryState } from "./contract/zod.gen.js";
import { zRelationshipNavigationStatus } from "./contract/zod.gen.js";
import { zSessionListItemKind } from "./contract/zod.gen.js";
import { zSyncStatus } from "./contract/zod.gen.js";
import { zSyncHoldReason } from "./contract/zod.gen.js";
import { zSyncPushSessionStatus } from "./contract/zod.gen.js";
import { zSyncPushStep } from "./contract/zod.gen.js";
import { zSyncPushStepOutcome } from "./contract/zod.gen.js";
import { zSyncLoginStatus } from "./contract/zod.gen.js";
import { zSyncLogoutStatus } from "./contract/zod.gen.js";
import { zLocalPublicationState } from "./contract/zod.gen.js";
import { zLocalCollectiveSuggestionReason } from "./contract/zod.gen.js";
import { zLocalSettingKind } from "./contract/zod.gen.js";
import { zLocalKickstartScope } from "./contract/zod.gen.js";
import { zLocalKickstartStopKind } from "./contract/zod.gen.js";
import { zLocalKickstartFieldKind } from "./contract/zod.gen.js";
import { zLocalKickstartLineKind } from "./contract/zod.gen.js";
import { zLocalKickstartNodeKind } from "./contract/zod.gen.js";
import { zLocalKickstartNodeState } from "./contract/zod.gen.js";
import { zLocalKickstartScanState } from "./contract/zod.gen.js";
import { zLocalKickstartIngestState } from "./contract/zod.gen.js";
import { zLocalKickstartResolution } from "./contract/zod.gen.js";
import { zLocalKickstartNextStep } from "./contract/zod.gen.js";
import { zLocalKickstartDiffLineType } from "./contract/zod.gen.js";
import { zLocalKickstartEventKind } from "./contract/zod.gen.js";
import { zLocalKickstartPreviewOp } from "./contract/zod.gen.js";
import { zLocalKickstartPreviewContinuation } from "./contract/zod.gen.js";
import { zLocalKickstartPreviewEventKind } from "./contract/zod.gen.js";
import { zLocalKickstartPreviewCutReason } from "./contract/zod.gen.js";
import { zLocalKickstartPreviewCloseReason } from "./contract/zod.gen.js";
import { zLocalKickstartCloseReason } from "./contract/zod.gen.js";
import { zLocalKickstartConnectAnswer } from "./contract/zod.gen.js";
import { zLocalKickstartErrorCode } from "./contract/zod.gen.js";
import { zAutoPublishEvent } from "./contract/zod.gen.js";
import { zAutoPublishHookStatus } from "./contract/zod.gen.js";
import { zAutoPublishRuleKind } from "./contract/zod.gen.js";
import { zThinkingLevel } from "./contract/zod.gen.js";
import { zScryVerdict } from "./contract/zod.gen.js";

export const VillageAttestationType = Object.freeze({
  UsedInTraining: "used_in_training",
  Referenced: "referenced",
  Evaluated: "evaluated",
  Deployed: "deployed",
} as const);
export type VillageAttestationType = (typeof VillageAttestationType)[keyof typeof VillageAttestationType];
export const AllVillageAttestationTypes = Object.freeze([VillageAttestationType.UsedInTraining, VillageAttestationType.Referenced, VillageAttestationType.Evaluated, VillageAttestationType.Deployed]) as readonly VillageAttestationType[];
export function isVillageAttestationType(value: unknown): value is VillageAttestationType {
  return zVillageAttestationType.safeParse(value).success;
}

export const TranscriptEntryReferenceKind = Object.freeze({
  Source: "source",
  ToolCall: "tool_call",
  ToolResult: "tool_result",
} as const);
export type TranscriptEntryReferenceKind = (typeof TranscriptEntryReferenceKind)[keyof typeof TranscriptEntryReferenceKind];
export const AllTranscriptEntryReferenceKinds = Object.freeze([TranscriptEntryReferenceKind.Source, TranscriptEntryReferenceKind.ToolCall, TranscriptEntryReferenceKind.ToolResult]) as readonly TranscriptEntryReferenceKind[];
export function isTranscriptEntryReferenceKind(value: unknown): value is TranscriptEntryReferenceKind {
  return zTranscriptEntryReferenceKind.safeParse(value).success;
}

export const TranscriptReadTransportKind = Object.freeze({
  LocalSession: "local_session",
  VillageTranscript: "village_transcript",
} as const);
export type TranscriptReadTransportKind = (typeof TranscriptReadTransportKind)[keyof typeof TranscriptReadTransportKind];
export const AllTranscriptReadTransportKinds = Object.freeze([TranscriptReadTransportKind.LocalSession, TranscriptReadTransportKind.VillageTranscript]) as readonly TranscriptReadTransportKind[];
export function isTranscriptReadTransportKind(value: unknown): value is TranscriptReadTransportKind {
  return zTranscriptReadTransportKind.safeParse(value).success;
}

export const TranscriptRecordedIndexState = Object.freeze({
  Present: "present",
  Absent: "absent",
  Invalid: "invalid",
} as const);
export type TranscriptRecordedIndexState = (typeof TranscriptRecordedIndexState)[keyof typeof TranscriptRecordedIndexState];
export const AllTranscriptRecordedIndexStates = Object.freeze([TranscriptRecordedIndexState.Present, TranscriptRecordedIndexState.Absent, TranscriptRecordedIndexState.Invalid]) as readonly TranscriptRecordedIndexState[];
export function isTranscriptRecordedIndexState(value: unknown): value is TranscriptRecordedIndexState {
  return zTranscriptRecordedIndexState.safeParse(value).success;
}

export const TranscriptRetainedState = Object.freeze({
  Absent: "absent",
  Deferred: "deferred",
} as const);
export type TranscriptRetainedState = (typeof TranscriptRetainedState)[keyof typeof TranscriptRetainedState];
export const AllTranscriptRetainedStates = Object.freeze([TranscriptRetainedState.Absent, TranscriptRetainedState.Deferred]) as readonly TranscriptRetainedState[];
export function isTranscriptRetainedState(value: unknown): value is TranscriptRetainedState {
  return zTranscriptRetainedState.safeParse(value).success;
}

export const TranscriptViewerProjection = Object.freeze({
  Anonymous: "anonymous",
  Authenticated: "authenticated",
} as const);
export type TranscriptViewerProjection = (typeof TranscriptViewerProjection)[keyof typeof TranscriptViewerProjection];
export const AllTranscriptViewerProjections = Object.freeze([TranscriptViewerProjection.Anonymous, TranscriptViewerProjection.Authenticated]) as readonly TranscriptViewerProjection[];
export function isTranscriptViewerProjection(value: unknown): value is TranscriptViewerProjection {
  return zTranscriptViewerProjection.safeParse(value).success;
}

export const TranscriptReadValidation = Object.freeze({
  PublishedStrict: "published_strict",
  ServedLegacy: "served_legacy",
  Migrated: "migrated",
} as const);
export type TranscriptReadValidation = (typeof TranscriptReadValidation)[keyof typeof TranscriptReadValidation];
export const AllTranscriptReadValidations = Object.freeze([TranscriptReadValidation.PublishedStrict, TranscriptReadValidation.ServedLegacy, TranscriptReadValidation.Migrated]) as readonly TranscriptReadValidation[];
export function isTranscriptReadValidation(value: unknown): value is TranscriptReadValidation {
  return zTranscriptReadValidation.safeParse(value).success;
}

export const TranscriptNativePresence = Object.freeze({
  Absent: "absent",
  Present: "present",
  Withheld: "withheld",
} as const);
export type TranscriptNativePresence = (typeof TranscriptNativePresence)[keyof typeof TranscriptNativePresence];
export const AllTranscriptNativePresences = Object.freeze([TranscriptNativePresence.Absent, TranscriptNativePresence.Present, TranscriptNativePresence.Withheld]) as readonly TranscriptNativePresence[];
export function isTranscriptNativePresence(value: unknown): value is TranscriptNativePresence {
  return zTranscriptNativePresence.safeParse(value).success;
}

export const TranscriptUsagePresence = Object.freeze({
  Present: "present",
  Empty: "empty",
  Unknown: "unknown",
} as const);
export type TranscriptUsagePresence = (typeof TranscriptUsagePresence)[keyof typeof TranscriptUsagePresence];
export const AllTranscriptUsagePresences = Object.freeze([TranscriptUsagePresence.Present, TranscriptUsagePresence.Empty, TranscriptUsagePresence.Unknown]) as readonly TranscriptUsagePresence[];
export function isTranscriptUsagePresence(value: unknown): value is TranscriptUsagePresence {
  return zTranscriptUsagePresence.safeParse(value).success;
}

export const TranscriptSegmentPartitionKind = Object.freeze({
  Main: "main",
  History: "history",
} as const);
export type TranscriptSegmentPartitionKind = (typeof TranscriptSegmentPartitionKind)[keyof typeof TranscriptSegmentPartitionKind];
export const AllTranscriptSegmentPartitionKinds = Object.freeze([TranscriptSegmentPartitionKind.Main, TranscriptSegmentPartitionKind.History]) as readonly TranscriptSegmentPartitionKind[];
export function isTranscriptSegmentPartitionKind(value: unknown): value is TranscriptSegmentPartitionKind {
  return zTranscriptSegmentPartitionKind.safeParse(value).success;
}

export const CollectiveObservationReason = Object.freeze({
  Membership: "membership",
  Contribution: "contribution",
  Organization: "organization",
  Explicit: "explicit",
} as const);
export type CollectiveObservationReason = (typeof CollectiveObservationReason)[keyof typeof CollectiveObservationReason];
export const AllCollectiveObservationReasons = Object.freeze([CollectiveObservationReason.Membership, CollectiveObservationReason.Contribution, CollectiveObservationReason.Organization, CollectiveObservationReason.Explicit]) as readonly CollectiveObservationReason[];
export function isCollectiveObservationReason(value: unknown): value is CollectiveObservationReason {
  return zCollectiveObservationReason.safeParse(value).success;
}

export const CollectiveObservationState = Object.freeze({
  Observed: "observed",
  Unavailable: "unavailable",
  Invalidated: "invalidated",
} as const);
export type CollectiveObservationState = (typeof CollectiveObservationState)[keyof typeof CollectiveObservationState];
export const AllCollectiveObservationStates = Object.freeze([CollectiveObservationState.Observed, CollectiveObservationState.Unavailable, CollectiveObservationState.Invalidated]) as readonly CollectiveObservationState[];
export function isCollectiveObservationState(value: unknown): value is CollectiveObservationState {
  return zCollectiveObservationState.safeParse(value).success;
}

export const CollectiveOperation = Object.freeze({
  Publish: "publish",
  Share: "share",
  Fetch: "fetch",
  Pull: "pull",
} as const);
export type CollectiveOperation = (typeof CollectiveOperation)[keyof typeof CollectiveOperation];
export const AllCollectiveOperations = Object.freeze([CollectiveOperation.Publish, CollectiveOperation.Share, CollectiveOperation.Fetch, CollectiveOperation.Pull]) as readonly CollectiveOperation[];
export function isCollectiveOperation(value: unknown): value is CollectiveOperation {
  return zCollectiveOperation.safeParse(value).success;
}

export const OperationIssueCode = Object.freeze({
  Auth: "auth",
  Remote: "remote",
  Contract: "contract",
  Deadline: "deadline",
  Persistence: "persistence",
  Superseded: "superseded",
} as const);
export type OperationIssueCode = (typeof OperationIssueCode)[keyof typeof OperationIssueCode];
export const AllOperationIssueCodes = Object.freeze([OperationIssueCode.Auth, OperationIssueCode.Remote, OperationIssueCode.Contract, OperationIssueCode.Deadline, OperationIssueCode.Persistence, OperationIssueCode.Superseded]) as readonly OperationIssueCode[];
export function isOperationIssueCode(value: unknown): value is OperationIssueCode {
  return zOperationIssueCode.safeParse(value).success;
}

export const LocalContributionEvidenceKind = Object.freeze({
  RemoteAttempts: "remote-attempts",
  Receipts: "receipts",
  OperationDiagnostics: "operation-diagnostics",
} as const);
export type LocalContributionEvidenceKind = (typeof LocalContributionEvidenceKind)[keyof typeof LocalContributionEvidenceKind];
export const AllLocalContributionEvidenceKinds = Object.freeze([LocalContributionEvidenceKind.RemoteAttempts, LocalContributionEvidenceKind.Receipts, LocalContributionEvidenceKind.OperationDiagnostics]) as readonly LocalContributionEvidenceKind[];
export function isLocalContributionEvidenceKind(value: unknown): value is LocalContributionEvidenceKind {
  return zLocalContributionEvidenceKind.safeParse(value).success;
}

export const CollectiveEvidenceCoverage = Object.freeze({
  Available: "available",
  Unknown: "unknown",
} as const);
export type CollectiveEvidenceCoverage = (typeof CollectiveEvidenceCoverage)[keyof typeof CollectiveEvidenceCoverage];
export const AllCollectiveEvidenceCoverages = Object.freeze([CollectiveEvidenceCoverage.Available, CollectiveEvidenceCoverage.Unknown]) as readonly CollectiveEvidenceCoverage[];
export function isCollectiveEvidenceCoverage(value: unknown): value is CollectiveEvidenceCoverage {
  return zCollectiveEvidenceCoverage.safeParse(value).success;
}

export const LocalConfirmedContributionOutcome = Object.freeze({
  Published: "published",
  Submitted: "submitted",
} as const);
export type LocalConfirmedContributionOutcome = (typeof LocalConfirmedContributionOutcome)[keyof typeof LocalConfirmedContributionOutcome];
export const AllLocalConfirmedContributionOutcomes = Object.freeze([LocalConfirmedContributionOutcome.Published, LocalConfirmedContributionOutcome.Submitted]) as readonly LocalConfirmedContributionOutcome[];
export function isLocalConfirmedContributionOutcome(value: unknown): value is LocalConfirmedContributionOutcome {
  return zLocalConfirmedContributionOutcome.safeParse(value).success;
}

export const LocalOperationDiagnosticOutcome = Object.freeze({
  Failed: "failed",
  Uncertain: "uncertain",
} as const);
export type LocalOperationDiagnosticOutcome = (typeof LocalOperationDiagnosticOutcome)[keyof typeof LocalOperationDiagnosticOutcome];
export const AllLocalOperationDiagnosticOutcomes = Object.freeze([LocalOperationDiagnosticOutcome.Failed, LocalOperationDiagnosticOutcome.Uncertain]) as readonly LocalOperationDiagnosticOutcome[];
export function isLocalOperationDiagnosticOutcome(value: unknown): value is LocalOperationDiagnosticOutcome {
  return zLocalOperationDiagnosticOutcome.safeParse(value).success;
}

export const PullAssociationMetadataState = Object.freeze({
  Observed: "observed",
  Stale: "stale",
  Unknown: "unknown",
  Invalidated: "invalidated",
} as const);
export type PullAssociationMetadataState = (typeof PullAssociationMetadataState)[keyof typeof PullAssociationMetadataState];
export const AllPullAssociationMetadataStates = Object.freeze([PullAssociationMetadataState.Observed, PullAssociationMetadataState.Stale, PullAssociationMetadataState.Unknown, PullAssociationMetadataState.Invalidated]) as readonly PullAssociationMetadataState[];
export function isPullAssociationMetadataState(value: unknown): value is PullAssociationMetadataState {
  return zPullAssociationMetadataState.safeParse(value).success;
}

export const CollectiveFetchDomain = Object.freeze({
  Catalog: "catalog",
  History: "history",
  Associations: "associations",
} as const);
export type CollectiveFetchDomain = (typeof CollectiveFetchDomain)[keyof typeof CollectiveFetchDomain];
export const AllCollectiveFetchDomains = Object.freeze([CollectiveFetchDomain.Catalog, CollectiveFetchDomain.History, CollectiveFetchDomain.Associations]) as readonly CollectiveFetchDomain[];
export function isCollectiveFetchDomain(value: unknown): value is CollectiveFetchDomain {
  return zCollectiveFetchDomain.safeParse(value).success;
}

export const CollectiveFetchState = Object.freeze({
  Complete: "complete",
  Failed: "failed",
  Cancelled: "cancelled",
  Superseded: "superseded",
  NotRequested: "not_requested",
} as const);
export type CollectiveFetchState = (typeof CollectiveFetchState)[keyof typeof CollectiveFetchState];
export const AllCollectiveFetchStates = Object.freeze([CollectiveFetchState.Complete, CollectiveFetchState.Failed, CollectiveFetchState.Cancelled, CollectiveFetchState.Superseded, CollectiveFetchState.NotRequested]) as readonly CollectiveFetchState[];
export function isCollectiveFetchState(value: unknown): value is CollectiveFetchState {
  return zCollectiveFetchState.safeParse(value).success;
}

export const CollectiveFetchCommit = Object.freeze({
  Updated: "updated",
  Unchanged: "unchanged",
  Retained: "retained",
} as const);
export type CollectiveFetchCommit = (typeof CollectiveFetchCommit)[keyof typeof CollectiveFetchCommit];
export const AllCollectiveFetchCommits = Object.freeze([CollectiveFetchCommit.Updated, CollectiveFetchCommit.Unchanged, CollectiveFetchCommit.Retained]) as readonly CollectiveFetchCommit[];
export function isCollectiveFetchCommit(value: unknown): value is CollectiveFetchCommit {
  return zCollectiveFetchCommit.safeParse(value).success;
}

export const CollectiveFetchOutcome = Object.freeze({
  Complete: "complete",
  Partial: "partial",
  Failed: "failed",
} as const);
export type CollectiveFetchOutcome = (typeof CollectiveFetchOutcome)[keyof typeof CollectiveFetchOutcome];
export const AllCollectiveFetchOutcomes = Object.freeze([CollectiveFetchOutcome.Complete, CollectiveFetchOutcome.Partial, CollectiveFetchOutcome.Failed]) as readonly CollectiveFetchOutcome[];
export function isCollectiveFetchOutcome(value: unknown): value is CollectiveFetchOutcome {
  return zCollectiveFetchOutcome.safeParse(value).success;
}

export const CollectiveObservationSource = Object.freeze({
  Membership: "membership",
  ContributionDiscovery: "contribution-discovery",
  Organization: "organization",
  Explicit: "explicit",
  Submissions: "submissions",
  Attempts: "attempts",
  Associations: "associations",
} as const);
export type CollectiveObservationSource = (typeof CollectiveObservationSource)[keyof typeof CollectiveObservationSource];
export const AllCollectiveObservationSources = Object.freeze([CollectiveObservationSource.Membership, CollectiveObservationSource.ContributionDiscovery, CollectiveObservationSource.Organization, CollectiveObservationSource.Explicit, CollectiveObservationSource.Submissions, CollectiveObservationSource.Attempts, CollectiveObservationSource.Associations]) as readonly CollectiveObservationSource[];
export function isCollectiveObservationSource(value: unknown): value is CollectiveObservationSource {
  return zCollectiveObservationSource.safeParse(value).success;
}

export const CollectiveSourceOutcome = Object.freeze({
  Disclosed: "disclosed",
  NotDisclosed: "not_disclosed",
  Failed: "failed",
} as const);
export type CollectiveSourceOutcome = (typeof CollectiveSourceOutcome)[keyof typeof CollectiveSourceOutcome];
export const AllCollectiveSourceOutcomes = Object.freeze([CollectiveSourceOutcome.Disclosed, CollectiveSourceOutcome.NotDisclosed, CollectiveSourceOutcome.Failed]) as readonly CollectiveSourceOutcome[];
export function isCollectiveSourceOutcome(value: unknown): value is CollectiveSourceOutcome {
  return zCollectiveSourceOutcome.safeParse(value).success;
}

export const LocalSyncPublicationFilter = Object.freeze({
  All: "all",
  NotPublished: "not-published",
  Published: "published",
  Auto: "auto",
} as const);
export type LocalSyncPublicationFilter = (typeof LocalSyncPublicationFilter)[keyof typeof LocalSyncPublicationFilter];
export const AllLocalSyncPublicationFilters = Object.freeze([LocalSyncPublicationFilter.All, LocalSyncPublicationFilter.NotPublished, LocalSyncPublicationFilter.Published, LocalSyncPublicationFilter.Auto]) as readonly LocalSyncPublicationFilter[];
export function isLocalSyncPublicationFilter(value: unknown): value is LocalSyncPublicationFilter {
  return zLocalSyncPublicationFilter.safeParse(value).success;
}

export const PageSize = Object.freeze({
  P25: 25,
  P50: 50,
  P100: 100,
} as const);
export type PageSize = (typeof PageSize)[keyof typeof PageSize];
export const AllPageSizes = Object.freeze([PageSize.P25, PageSize.P50, PageSize.P100]) as readonly PageSize[];
export function isPageSize(value: unknown): value is PageSize {
  return zPageSize.safeParse(value).success;
}

export const PublishOperationKind = Object.freeze({
  Preserve: "preserve",
  Replace: "replace",
  Append: "append",
} as const);
export type PublishOperationKind = (typeof PublishOperationKind)[keyof typeof PublishOperationKind];
export const AllPublishOperationKinds = Object.freeze([PublishOperationKind.Preserve, PublishOperationKind.Replace, PublishOperationKind.Append]) as readonly PublishOperationKind[];
export function isPublishOperationKind(value: unknown): value is PublishOperationKind {
  return zPublishOperationKind.safeParse(value).success;
}

export const VisibilityIntent = Object.freeze({
  Private: "private",
  Public: "public",
} as const);
export type VisibilityIntent = (typeof VisibilityIntent)[keyof typeof VisibilityIntent];
export const AllVisibilityIntents = Object.freeze([VisibilityIntent.Private, VisibilityIntent.Public]) as readonly VisibilityIntent[];
export function isVisibilityIntent(value: unknown): value is VisibilityIntent {
  return zVisibilityIntent.safeParse(value).success;
}

export const AnnotationAxis = Object.freeze({
  Type: "type",
  Session: "session",
  Project: "project",
} as const);
export type AnnotationAxis = (typeof AnnotationAxis)[keyof typeof AnnotationAxis];
export const AllAnnotationAxes = Object.freeze([AnnotationAxis.Type, AnnotationAxis.Session, AnnotationAxis.Project]) as readonly AnnotationAxis[];
export function isAnnotationAxis(value: unknown): value is AnnotationAxis {
  return zAnnotationAxis.safeParse(value).success;
}

export const AnnotationDatatype = Object.freeze({
  Text: "text",
  Integer: "integer",
  Real: "real",
  Boolean: "boolean",
} as const);
export type AnnotationDatatype = (typeof AnnotationDatatype)[keyof typeof AnnotationDatatype];
export const AllAnnotationDatatypes = Object.freeze([AnnotationDatatype.Text, AnnotationDatatype.Integer, AnnotationDatatype.Real, AnnotationDatatype.Boolean]) as readonly AnnotationDatatype[];
export function isAnnotationDatatype(value: unknown): value is AnnotationDatatype {
  return zAnnotationDatatype.safeParse(value).success;
}

export const AnnotationPushStatus = Object.freeze({
  Created: "created",
  Updated: "updated",
  Skipped: "skipped",
  Error: "error",
} as const);
export type AnnotationPushStatus = (typeof AnnotationPushStatus)[keyof typeof AnnotationPushStatus];
export const AllAnnotationPushStatuses = Object.freeze([AnnotationPushStatus.Created, AnnotationPushStatus.Updated, AnnotationPushStatus.Skipped, AnnotationPushStatus.Error]) as readonly AnnotationPushStatus[];
export function isAnnotationPushStatus(value: unknown): value is AnnotationPushStatus {
  return zAnnotationPushStatus.safeParse(value).success;
}

export const AnnotationStatus = Object.freeze({
  Proposed: "proposed",
  Active: "active",
  Deprecated: "deprecated",
  Retired: "retired",
} as const);
export type AnnotationStatus = (typeof AnnotationStatus)[keyof typeof AnnotationStatus];
export const AllAnnotationStatuses = Object.freeze([AnnotationStatus.Proposed, AnnotationStatus.Active, AnnotationStatus.Deprecated, AnnotationStatus.Retired]) as readonly AnnotationStatus[];
export function isAnnotationStatus(value: unknown): value is AnnotationStatus {
  return zAnnotationStatus.safeParse(value).success;
}

export const AnnotatorKind = Object.freeze({
  Human: "human",
  Agent: "agent",
  Rule: "rule",
} as const);
export type AnnotatorKind = (typeof AnnotatorKind)[keyof typeof AnnotatorKind];
export const AllAnnotatorKinds = Object.freeze([AnnotatorKind.Human, AnnotatorKind.Agent, AnnotatorKind.Rule]) as readonly AnnotatorKind[];
export function isAnnotatorKind(value: unknown): value is AnnotatorKind {
  return zAnnotatorKind.safeParse(value).success;
}

export const AssociationConclusion = Object.freeze({
  Confirmed: "confirmed",
  Candidate: "candidate",
} as const);
export type AssociationConclusion = (typeof AssociationConclusion)[keyof typeof AssociationConclusion];
export const AllAssociationConclusions = Object.freeze([AssociationConclusion.Confirmed, AssociationConclusion.Candidate]) as readonly AssociationConclusion[];
export function isAssociationConclusion(value: unknown): value is AssociationConclusion {
  return zAssociationConclusion.safeParse(value).success;
}

export const AssociationEvidenceKind = Object.freeze({
  RecordedCommit: "recorded_commit",
  TouchedFile: "touched_file",
  BranchMembership: "branch_membership",
  TimeWindow: "time_window",
} as const);
export type AssociationEvidenceKind = (typeof AssociationEvidenceKind)[keyof typeof AssociationEvidenceKind];
export const AllAssociationEvidenceKinds = Object.freeze([AssociationEvidenceKind.RecordedCommit, AssociationEvidenceKind.TouchedFile, AssociationEvidenceKind.BranchMembership, AssociationEvidenceKind.TimeWindow]) as readonly AssociationEvidenceKind[];
export function isAssociationEvidenceKind(value: unknown): value is AssociationEvidenceKind {
  return zAssociationEvidenceKind.safeParse(value).success;
}

export const BuiltinCommand = Object.freeze({
  Exit: "exit",
  Compact: "compact",
  Clear: "clear",
  New: "new",
  Model: "model",
  Usage: "usage",
  Cost: "cost",
  Context: "context",
  Plugin: "plugin",
  Permissions: "permissions",
  Login: "login",
  Resume: "resume",
  Plan: "plan",
  Fast: "fast",
  Voice: "voice",
  Todos: "todos",
  ReloadPlugins: "reload-plugins",
  Sandbox: "sandbox",
  Config: "config",
  Statusline: "statusline",
  Upgrade: "upgrade",
  ExtraUsage: "extra-usage",
  RateLimitOptions: "rate-limit-options",
  PrivacySettings: "privacy-settings",
  Help: "help",
  Commands: "commands",
} as const);
export type BuiltinCommand = (typeof BuiltinCommand)[keyof typeof BuiltinCommand];
export const AllClaudeBuiltinCmds = Object.freeze([BuiltinCommand.Exit, BuiltinCommand.Compact, BuiltinCommand.Clear, BuiltinCommand.New, BuiltinCommand.Model, BuiltinCommand.Usage, BuiltinCommand.Cost, BuiltinCommand.Context, BuiltinCommand.Plugin, BuiltinCommand.Permissions, BuiltinCommand.Login, BuiltinCommand.Resume, BuiltinCommand.Plan, BuiltinCommand.Fast, BuiltinCommand.Voice, BuiltinCommand.Todos, BuiltinCommand.ReloadPlugins, BuiltinCommand.Sandbox, BuiltinCommand.Config, BuiltinCommand.Statusline, BuiltinCommand.Upgrade, BuiltinCommand.ExtraUsage, BuiltinCommand.RateLimitOptions, BuiltinCommand.PrivacySettings, BuiltinCommand.Help, BuiltinCommand.Commands]) as readonly BuiltinCommand[];
export function isBuiltinCommand(value: unknown): value is BuiltinCommand {
  return zBuiltinCommand.safeParse(value).success;
}

export const ChangeBinding = Object.freeze({
  Bound: "bound",
  Candidate: "candidate",
} as const);
export type ChangeBinding = (typeof ChangeBinding)[keyof typeof ChangeBinding];
export const AllChangeBindings = Object.freeze([ChangeBinding.Bound, ChangeBinding.Candidate]) as readonly ChangeBinding[];
export function isChangeBinding(value: unknown): value is ChangeBinding {
  return zChangeBinding.safeParse(value).success;
}

export const ChannelTopic = Object.freeze({
  Dashboard: "dashboard",
  Sessions: "sessions",
  SessionDetail: "session_detail",
  Trends: "trends",
  Quality: "quality",
  Annotations: "annotations",
  ProjectFamiliarity: "project_familiarity",
  Kickstart: "kickstart",
} as const);
export type ChannelTopic = (typeof ChannelTopic)[keyof typeof ChannelTopic];
export const AllChannelTopics = Object.freeze([ChannelTopic.Dashboard, ChannelTopic.Sessions, ChannelTopic.SessionDetail, ChannelTopic.Trends, ChannelTopic.Quality, ChannelTopic.Annotations, ChannelTopic.ProjectFamiliarity, ChannelTopic.Kickstart]) as readonly ChannelTopic[];
export function isChannelTopic(value: unknown): value is ChannelTopic {
  return zChannelTopic.safeParse(value).success;
}

export const Confidence = Object.freeze({
  High: "high",
  Medium: "medium",
  Low: "low",
} as const);
export type Confidence = (typeof Confidence)[keyof typeof Confidence];
export const AllConfidences = Object.freeze([Confidence.High, Confidence.Medium, Confidence.Low]) as readonly Confidence[];
export function isConfidence(value: unknown): value is Confidence {
  return zConfidence.safeParse(value).success;
}

export const ContentKind = Object.freeze({
  SessionDetail: "session_detail",
} as const);
export type ContentKind = (typeof ContentKind)[keyof typeof ContentKind];
export const AllContentKinds = Object.freeze([ContentKind.SessionDetail]) as readonly ContentKind[];
export function isContentKind(value: unknown): value is ContentKind {
  return zContentKind.safeParse(value).success;
}

export const DecayLevel = Object.freeze({
  Fresh: "fresh",
  Fading: "fading",
  Stale: "stale",
  Unexplored: "unexplored",
} as const);
export type DecayLevel = (typeof DecayLevel)[keyof typeof DecayLevel];
export const AllDecayLevels = Object.freeze([DecayLevel.Fresh, DecayLevel.Fading, DecayLevel.Stale, DecayLevel.Unexplored]) as readonly DecayLevel[];
export function isDecayLevel(value: unknown): value is DecayLevel {
  return zDecayLevel.safeParse(value).success;
}

export const DiffLineKind = Object.freeze({
  Context: "context",
  Add: "add",
  Delete: "del",
} as const);
export type DiffLineKind = (typeof DiffLineKind)[keyof typeof DiffLineKind];
export const AllDiffLineKinds = Object.freeze([DiffLineKind.Context, DiffLineKind.Add, DiffLineKind.Delete]) as readonly DiffLineKind[];
export function isDiffLineKind(value: unknown): value is DiffLineKind {
  return zDiffLineKind.safeParse(value).success;
}

export const EdgeViolationKind = Object.freeze({
  Cycle: "cycle",
  WrongWay: "wrong_way",
} as const);
export type EdgeViolationKind = (typeof EdgeViolationKind)[keyof typeof EdgeViolationKind];
export const AllEdgeViolationKinds = Object.freeze([EdgeViolationKind.Cycle, EdgeViolationKind.WrongWay]) as readonly EdgeViolationKind[];
export function isEdgeViolationKind(value: unknown): value is EdgeViolationKind {
  return zEdgeViolationKind.safeParse(value).success;
}

export const EntryType = Object.freeze({
  Text: "text",
  ToolUse: "tool_use",
  ToolResult: "tool_result",
  Thinking: "thinking",
  System: "system",
  Error: "error",
  Result: "result",
  Image: "image",
} as const);
export type EntryType = (typeof EntryType)[keyof typeof EntryType];
export const AllEntryTypes = Object.freeze([EntryType.Text, EntryType.ToolUse, EntryType.ToolResult, EntryType.Thinking, EntryType.System, EntryType.Error, EntryType.Result, EntryType.Image]) as readonly EntryType[];
export function isEntryType(value: unknown): value is EntryType {
  return zEntryType.safeParse(value).success;
}

export const TurnPartType = Object.freeze({
  Text: "text",
  Image: "image",
} as const);
export type TurnPartType = (typeof TurnPartType)[keyof typeof TurnPartType];
export const AllTurnPartTypes = Object.freeze([TurnPartType.Text, TurnPartType.Image]) as readonly TurnPartType[];
export function isTurnPartType(value: unknown): value is TurnPartType {
  return zTurnPartType.safeParse(value).success;
}

export const ImageMediaType = Object.freeze({
  Png: "image/png",
  Jpeg: "image/jpeg",
} as const);
export type ImageMediaType = (typeof ImageMediaType)[keyof typeof ImageMediaType];
export const AllImageMediaTypes = Object.freeze([ImageMediaType.Png, ImageMediaType.Jpeg]) as readonly ImageMediaType[];
export function isImageMediaType(value: unknown): value is ImageMediaType {
  return zImageMediaType.safeParse(value).success;
}

export const OmittedReason = Object.freeze({
  TooLarge: "too_large",
  Unsupported: "unsupported",
  Missing: "missing",
  NotPublished: "not_published",
} as const);
export type OmittedReason = (typeof OmittedReason)[keyof typeof OmittedReason];
export const AllOmittedReasons = Object.freeze([OmittedReason.TooLarge, OmittedReason.Unsupported, OmittedReason.Missing, OmittedReason.NotPublished]) as readonly OmittedReason[];
export function isOmittedReason(value: unknown): value is OmittedReason {
  return zOmittedReason.safeParse(value).success;
}

export const FileChangeStatus = Object.freeze({
  Modified: "M",
  Added: "A",
  Deleted: "D",
  Renamed: "R",
} as const);
export type FileChangeStatus = (typeof FileChangeStatus)[keyof typeof FileChangeStatus];
export const AllFileChangeStatuses = Object.freeze([FileChangeStatus.Modified, FileChangeStatus.Added, FileChangeStatus.Deleted, FileChangeStatus.Renamed]) as readonly FileChangeStatus[];
export function isFileChangeStatus(value: unknown): value is FileChangeStatus {
  return zFileChangeStatus.safeParse(value).success;
}

export const Harness = Object.freeze({
  ClaudeCode: "claude-code",
  GeminiCLI: "gemini-cli",
  Codex: "codex",
  OpenCode: "opencode",
  Cursor: "cursor",
  Antigravity: "antigravity",
  Strike: "strike",
  Pi: "pi",
  GrokBuild: "grok-build",
  DevinDesktop: "devin-desktop",
  Cline: "cline",
  Kiro: "kiro",
  KiroCLI: "kiro-cli",
  KiroCrew: "kiro-crew",
  AntigravityCLI: "antigravity-cli",
  AntigravityIDE: "antigravity-ide",
} as const);
export type Harness = (typeof Harness)[keyof typeof Harness];
/**
 * AllHarnesses is the declared subset of Harness, not the full
 * known set. It includes archived Gemini CLI, which has no harvester. Each
 * Peasant build's adapter registry owns which harnesses it can harvest. Every
 * Harness member remains individually valid and accepted by isHarness.
 */
export const AllHarnesses = Object.freeze([Harness.ClaudeCode, Harness.GeminiCLI, Harness.Codex, Harness.OpenCode, Harness.Cursor, Harness.Strike, Harness.Pi]) as readonly Harness[];
export function isHarness(value: unknown): value is Harness {
  return zHarness.safeParse(value).success;
}

export const InsightKind = Object.freeze({
  Decision: "decision",
  Friction: "friction",
  Unusual: "unusual",
  RetryLoop: "retry_loop",
} as const);
export type InsightKind = (typeof InsightKind)[keyof typeof InsightKind];
export const AllInsightKinds = Object.freeze([InsightKind.Decision, InsightKind.Friction, InsightKind.Unusual, InsightKind.RetryLoop]) as readonly InsightKind[];
export function isInsightKind(value: unknown): value is InsightKind {
  return zInsightKind.safeParse(value).success;
}

export const InsightProvenance = Object.freeze({
  Mechanical: "mechanical",
  Mined: "mined",
} as const);
export type InsightProvenance = (typeof InsightProvenance)[keyof typeof InsightProvenance];
export const AllInsightProvenances = Object.freeze([InsightProvenance.Mechanical, InsightProvenance.Mined]) as readonly InsightProvenance[];
export function isInsightProvenance(value: unknown): value is InsightProvenance {
  return zInsightProvenance.safeParse(value).success;
}

export const InteractionType = Object.freeze({
  Mentioned: "mentioned",
  Read: "read",
  Discussed: "discussed",
  Questioned: "questioned",
} as const);
export type InteractionType = (typeof InteractionType)[keyof typeof InteractionType];
export const AllInteractionTypes = Object.freeze([InteractionType.Mentioned, InteractionType.Read, InteractionType.Discussed, InteractionType.Questioned]) as readonly InteractionType[];
export function isInteractionType(value: unknown): value is InteractionType {
  return zInteractionType.safeParse(value).success;
}

export const License = Object.freeze({
  CC0: "CC0-1.0",
  CCBY: "CC-BY-4.0",
  CCBYSA: "CC-BY-SA-4.0",
} as const);
export type License = (typeof License)[keyof typeof License];
export const AllLicenses = Object.freeze([License.CC0, License.CCBY, License.CCBYSA]) as readonly License[];
export function isLicense(value: unknown): value is License {
  return zLicense.safeParse(value).success;
}

export const MapNodeKind = Object.freeze({
  Module: "module",
  Package: "package",
  File: "file",
} as const);
export type MapNodeKind = (typeof MapNodeKind)[keyof typeof MapNodeKind];
export const AllMapNodeKinds = Object.freeze([MapNodeKind.Module, MapNodeKind.Package, MapNodeKind.File]) as readonly MapNodeKind[];
export function isMapNodeKind(value: unknown): value is MapNodeKind {
  return zMapNodeKind.safeParse(value).success;
}

export const MessageType = Object.freeze({
  Subscribe: "subscribe",
  Unsubscribe: "unsubscribe",
  Dashboard: "dashboard",
  Sessions: "sessions",
  SessionDetail: "session_detail",
  Trends: "trends",
  Quality: "quality",
  Annotations: "annotations",
  ProjectFamiliarity: "project_familiarity",
  Connected: "connected",
  Error: "error",
  Kickstart: "kickstart",
  KickstartPreview: "kickstart_preview",
} as const);
export type MessageType = (typeof MessageType)[keyof typeof MessageType];
export const AllMessageTypes = Object.freeze([MessageType.Subscribe, MessageType.Unsubscribe, MessageType.Dashboard, MessageType.Sessions, MessageType.SessionDetail, MessageType.Trends, MessageType.Quality, MessageType.Annotations, MessageType.ProjectFamiliarity, MessageType.Connected, MessageType.Error, MessageType.Kickstart, MessageType.KickstartPreview]) as readonly MessageType[];
export function isMessageType(value: unknown): value is MessageType {
  return zMessageType.safeParse(value).success;
}

export const ReadAttributionState = Object.freeze({
  Complete: "complete",
  Partial: "partial",
  Unavailable: "unavailable",
} as const);
export type ReadAttributionState = (typeof ReadAttributionState)[keyof typeof ReadAttributionState];
export const AllReadAttributionStates = Object.freeze([ReadAttributionState.Complete, ReadAttributionState.Partial, ReadAttributionState.Unavailable]) as readonly ReadAttributionState[];
export function isReadAttributionState(value: unknown): value is ReadAttributionState {
  return zReadAttributionState.safeParse(value).success;
}

export const ReadStateGrade = Object.freeze({
  None: "none",
  Viewed: "viewed",
  Reviewed: "reviewed",
  ReviewedInDetail: "reviewed_in_detail",
} as const);
export type ReadStateGrade = (typeof ReadStateGrade)[keyof typeof ReadStateGrade];
export const AllReadStateGrades = Object.freeze([ReadStateGrade.None, ReadStateGrade.Viewed, ReadStateGrade.Reviewed, ReadStateGrade.ReviewedInDetail]) as readonly ReadStateGrade[];
export function isReadStateGrade(value: unknown): value is ReadStateGrade {
  return zReadStateGrade.safeParse(value).success;
}

export const RewriteMethod = Object.freeze({
  Hash: "hash",
  PatchID: "patch_id",
  AuthorIdentity: "author_identity",
  MessageEmbedded: "message_embedded",
  Temporal: "temporal",
  None: "none",
} as const);
export type RewriteMethod = (typeof RewriteMethod)[keyof typeof RewriteMethod];
export const AllRewriteMethods = Object.freeze([RewriteMethod.Hash, RewriteMethod.PatchID, RewriteMethod.AuthorIdentity, RewriteMethod.MessageEmbedded, RewriteMethod.Temporal, RewriteMethod.None]) as readonly RewriteMethod[];
export function isRewriteMethod(value: unknown): value is RewriteMethod {
  return zRewriteMethod.safeParse(value).success;
}

export const RewriteResolution = Object.freeze({
  Live: "live",
  Rewritten: "rewritten",
  Unresolved: "unresolved",
} as const);
export type RewriteResolution = (typeof RewriteResolution)[keyof typeof RewriteResolution];
export const AllRewriteResolutions = Object.freeze([RewriteResolution.Live, RewriteResolution.Rewritten, RewriteResolution.Unresolved]) as readonly RewriteResolution[];
export function isRewriteResolution(value: unknown): value is RewriteResolution {
  return zRewriteResolution.safeParse(value).success;
}

export const Role = Object.freeze({
  User: "user",
  Assistant: "assistant",
  Tool: "tool",
  System: "system",
} as const);
export type Role = (typeof Role)[keyof typeof Role];
export const AllRoles = Object.freeze([Role.User, Role.Assistant, Role.Tool, Role.System]) as readonly Role[];
export function isRole(value: unknown): value is Role {
  return zRole.safeParse(value).success;
}

export const ScaleKind = Object.freeze({
  Nominal: "nominal",
  Ordinal: "ordinal",
  Continuous: "continuous",
} as const);
export type ScaleKind = (typeof ScaleKind)[keyof typeof ScaleKind];
export const AllScaleKinds = Object.freeze([ScaleKind.Nominal, ScaleKind.Ordinal, ScaleKind.Continuous]) as readonly ScaleKind[];
export function isScaleKind(value: unknown): value is ScaleKind {
  return zScaleKind.safeParse(value).success;
}

export const SessionOrigin = Object.freeze({
  User: "user",
  Agent: "agent",
  Unknown: "unknown",
} as const);
export type SessionOrigin = (typeof SessionOrigin)[keyof typeof SessionOrigin];
export const AllSessionOrigins = Object.freeze([SessionOrigin.User, SessionOrigin.Agent, SessionOrigin.Unknown]) as readonly SessionOrigin[];
export function isSessionOrigin(value: unknown): value is SessionOrigin {
  return zSessionOrigin.safeParse(value).success;
}

export const SessionOutcome = Object.freeze({
  Resolved: "resolved",
  Partial: "partial",
  Failed: "failed",
} as const);
export type SessionOutcome = (typeof SessionOutcome)[keyof typeof SessionOutcome];
export const AllOutcomes = Object.freeze([SessionOutcome.Resolved, SessionOutcome.Partial, SessionOutcome.Failed]) as readonly SessionOutcome[];
export function isSessionOutcome(value: unknown): value is SessionOutcome {
  return zSessionOutcome.safeParse(value).success;
}

export const SourceFormat = Object.freeze({
  JSONL: "jsonl",
  JSON: "json",
} as const);
export type SourceFormat = (typeof SourceFormat)[keyof typeof SourceFormat];
export const AllSourceFormats = Object.freeze([SourceFormat.JSONL, SourceFormat.JSON]) as readonly SourceFormat[];
export function isSourceFormat(value: unknown): value is SourceFormat {
  return zSourceFormat.safeParse(value).success;
}

export const StopReason = Object.freeze({
  EndTurn: "end_turn",
  Cancelled: "cancelled",
  MaxTokens: "max_tokens",
  MaxTurnRequests: "max_turn_requests",
  Refusal: "refusal",
} as const);
export type StopReason = (typeof StopReason)[keyof typeof StopReason];
export const AllStopReasons = Object.freeze([StopReason.EndTurn, StopReason.Cancelled, StopReason.MaxTokens, StopReason.MaxTurnRequests, StopReason.Refusal]) as readonly StopReason[];
export function isStopReason(value: unknown): value is StopReason {
  return zStopReason.safeParse(value).success;
}

export const TargetKind = Object.freeze({
  Session: "session",
  Entry: "entry",
  Annotation: "annotation",
  Project: "project",
  FileVersion: "file_version",
  Association: "association",
} as const);
export type TargetKind = (typeof TargetKind)[keyof typeof TargetKind];
export const AllTargetKinds = Object.freeze([TargetKind.Session, TargetKind.Entry, TargetKind.Annotation, TargetKind.Project, TargetKind.FileVersion, TargetKind.Association]) as readonly TargetKind[];
export function isTargetKind(value: unknown): value is TargetKind {
  return zTargetKind.safeParse(value).success;
}

export const ToolCallKind = Object.freeze({
  Read: "read",
  Edit: "edit",
  Delete: "delete",
  Move: "move",
  Search: "search",
  Execute: "execute",
  Think: "think",
  Fetch: "fetch",
  Other: "other",
} as const);
export type ToolCallKind = (typeof ToolCallKind)[keyof typeof ToolCallKind];
export const AllToolCallKinds = Object.freeze([ToolCallKind.Read, ToolCallKind.Edit, ToolCallKind.Delete, ToolCallKind.Move, ToolCallKind.Search, ToolCallKind.Execute, ToolCallKind.Think, ToolCallKind.Fetch, ToolCallKind.Other]) as readonly ToolCallKind[];
export function isToolCallKind(value: unknown): value is ToolCallKind {
  return zToolCallKind.safeParse(value).success;
}

export const TypeOrigin = Object.freeze({
  System: "system",
  User: "user",
  Group: "group",
} as const);
export type TypeOrigin = (typeof TypeOrigin)[keyof typeof TypeOrigin];
export const AllTypeOrigins = Object.freeze([TypeOrigin.System, TypeOrigin.User, TypeOrigin.Group]) as readonly TypeOrigin[];
export function isTypeOrigin(value: unknown): value is TypeOrigin {
  return zTypeOrigin.safeParse(value).success;
}

export const ValueDomainKind = Object.freeze({
  Enumerated: "enumerated",
  Described: "described",
} as const);
export type ValueDomainKind = (typeof ValueDomainKind)[keyof typeof ValueDomainKind];
export const AllValueDomainKinds = Object.freeze([ValueDomainKind.Enumerated, ValueDomainKind.Described]) as readonly ValueDomainKind[];
export function isValueDomainKind(value: unknown): value is ValueDomainKind {
  return zValueDomainKind.safeParse(value).success;
}

export const TranscriptUpdateLicense = Object.freeze({
  Clear: "",
  CC0: "CC0-1.0",
  CCBY: "CC-BY-4.0",
  CCBYSA: "CC-BY-SA-4.0",
} as const);
export type TranscriptUpdateLicense = (typeof TranscriptUpdateLicense)[keyof typeof TranscriptUpdateLicense];
export const AllTranscriptUpdateLicenses = Object.freeze([TranscriptUpdateLicense.Clear, TranscriptUpdateLicense.CC0, TranscriptUpdateLicense.CCBY, TranscriptUpdateLicense.CCBYSA]) as readonly TranscriptUpdateLicense[];
export function isTranscriptUpdateLicense(value: unknown): value is TranscriptUpdateLicense {
  return zTranscriptUpdateLicense.safeParse(value).success;
}

export const TranscriptUpdateVisibility = Object.freeze({
  Private: "private",
  Public: "public",
} as const);
export type TranscriptUpdateVisibility = (typeof TranscriptUpdateVisibility)[keyof typeof TranscriptUpdateVisibility];
export const AllTranscriptUpdateVisibilities = Object.freeze([TranscriptUpdateVisibility.Private, TranscriptUpdateVisibility.Public]) as readonly TranscriptUpdateVisibility[];
export function isTranscriptUpdateVisibility(value: unknown): value is TranscriptUpdateVisibility {
  return zTranscriptUpdateVisibility.safeParse(value).success;
}

export const VillageAssignableCollectiveRole = Object.freeze({
  Contributor: "contributor",
  Member: "member",
} as const);
export type VillageAssignableCollectiveRole = (typeof VillageAssignableCollectiveRole)[keyof typeof VillageAssignableCollectiveRole];
export const AllVillageAssignableCollectiveRoles = Object.freeze([VillageAssignableCollectiveRole.Contributor, VillageAssignableCollectiveRole.Member]) as readonly VillageAssignableCollectiveRole[];
export function isVillageAssignableCollectiveRole(value: unknown): value is VillageAssignableCollectiveRole {
  return zVillageAssignableCollectiveRole.safeParse(value).success;
}

export const VillageContributionStatus = Object.freeze({
  Approved: "approved",
  Pending: "pending",
} as const);
export type VillageContributionStatus = (typeof VillageContributionStatus)[keyof typeof VillageContributionStatus];
export const AllVillageContributionStatuses = Object.freeze([VillageContributionStatus.Approved, VillageContributionStatus.Pending]) as readonly VillageContributionStatus[];
export function isVillageContributionStatus(value: unknown): value is VillageContributionStatus {
  return zVillageContributionStatus.safeParse(value).success;
}

export const VillageCollectiveAcceptanceMode = Object.freeze({
  Open: "open",
  VerifiedOnly: "verified_only",
  Curated: "curated",
} as const);
export type VillageCollectiveAcceptanceMode = (typeof VillageCollectiveAcceptanceMode)[keyof typeof VillageCollectiveAcceptanceMode];
export const AllVillageCollectiveAcceptanceModes = Object.freeze([VillageCollectiveAcceptanceMode.Open, VillageCollectiveAcceptanceMode.VerifiedOnly, VillageCollectiveAcceptanceMode.Curated]) as readonly VillageCollectiveAcceptanceMode[];
export function isVillageCollectiveAcceptanceMode(value: unknown): value is VillageCollectiveAcceptanceMode {
  return zVillageCollectiveAcceptanceMode.safeParse(value).success;
}

export const VillageCollectiveDataAccess = Object.freeze({
  MembersOnly: "members_only",
  Contributors: "contributors",
  Public: "public",
} as const);
export type VillageCollectiveDataAccess = (typeof VillageCollectiveDataAccess)[keyof typeof VillageCollectiveDataAccess];
export const AllVillageCollectiveDataAccessPolicies = Object.freeze([VillageCollectiveDataAccess.MembersOnly, VillageCollectiveDataAccess.Contributors, VillageCollectiveDataAccess.Public]) as readonly VillageCollectiveDataAccess[];
export function isVillageCollectiveDataAccess(value: unknown): value is VillageCollectiveDataAccess {
  return zVillageCollectiveDataAccess.safeParse(value).success;
}

export const VillageCollectiveRole = Object.freeze({
  Owner: "owner",
  Member: "member",
  Contributor: "contributor",
  Pending: "pending",
} as const);
export type VillageCollectiveRole = (typeof VillageCollectiveRole)[keyof typeof VillageCollectiveRole];
export const AllVillageCollectiveRoles = Object.freeze([VillageCollectiveRole.Owner, VillageCollectiveRole.Member, VillageCollectiveRole.Contributor, VillageCollectiveRole.Pending]) as readonly VillageCollectiveRole[];
export function isVillageCollectiveRole(value: unknown): value is VillageCollectiveRole {
  return zVillageCollectiveRole.safeParse(value).success;
}

export const VillageCollectiveViewerRole = Object.freeze({
  None: "",
  Owner: "owner",
  Member: "member",
  Contributor: "contributor",
  Pending: "pending",
} as const);
export type VillageCollectiveViewerRole = (typeof VillageCollectiveViewerRole)[keyof typeof VillageCollectiveViewerRole];
export const AllVillageCollectiveViewerRoles = Object.freeze([VillageCollectiveViewerRole.None, VillageCollectiveViewerRole.Owner, VillageCollectiveViewerRole.Member, VillageCollectiveViewerRole.Contributor, VillageCollectiveViewerRole.Pending]) as readonly VillageCollectiveViewerRole[];
export function isVillageCollectiveViewerRole(value: unknown): value is VillageCollectiveViewerRole {
  return zVillageCollectiveViewerRole.safeParse(value).success;
}

export const VillageContributableView = Object.freeze({
  Grouped: "grouped",
  Paged: "paged",
  Snapshot: "snapshot",
} as const);
export type VillageContributableView = (typeof VillageContributableView)[keyof typeof VillageContributableView];
export const AllVillageContributableViews = Object.freeze([VillageContributableView.Grouped, VillageContributableView.Paged, VillageContributableView.Snapshot]) as readonly VillageContributableView[];
export function isVillageContributableView(value: unknown): value is VillageContributableView {
  return zVillageContributableView.safeParse(value).success;
}

export const VillageContributionErrorCode = Object.freeze({
  Moved: "moved",
  Missing: "missing",
  PrivateUnapproved: "private_unapproved",
  Conflict: "conflict",
  Policy: "policy",
  Membership: "membership",
  StaleSelection: "stale_selection",
  SnapshotTimeout: "snapshot_timeout",
  SnapshotBusy: "snapshot_busy",
  SnapshotTooLarge: "snapshot_too_large",
} as const);
export type VillageContributionErrorCode = (typeof VillageContributionErrorCode)[keyof typeof VillageContributionErrorCode];
export const AllVillageContributionErrorCodes = Object.freeze([VillageContributionErrorCode.Moved, VillageContributionErrorCode.Missing, VillageContributionErrorCode.PrivateUnapproved, VillageContributionErrorCode.Conflict, VillageContributionErrorCode.Policy, VillageContributionErrorCode.Membership, VillageContributionErrorCode.StaleSelection, VillageContributionErrorCode.SnapshotTimeout, VillageContributionErrorCode.SnapshotBusy, VillageContributionErrorCode.SnapshotTooLarge]) as readonly VillageContributionErrorCode[];
export function isVillageContributionErrorCode(value: unknown): value is VillageContributionErrorCode {
  return zVillageContributionErrorCode.safeParse(value).success;
}

export const VillageProjectNameSource = Object.freeze({
  Override: "override",
  Consented: "consented",
  Remote: "remote",
  Path: "path",
  Privacy: "privacy",
} as const);
export type VillageProjectNameSource = (typeof VillageProjectNameSource)[keyof typeof VillageProjectNameSource];
export const AllVillageProjectNameSources = Object.freeze([VillageProjectNameSource.Override, VillageProjectNameSource.Consented, VillageProjectNameSource.Remote, VillageProjectNameSource.Path, VillageProjectNameSource.Privacy]) as readonly VillageProjectNameSource[];
export function isVillageProjectNameSource(value: unknown): value is VillageProjectNameSource {
  return zVillageProjectNameSource.safeParse(value).success;
}

export const VillageReviewDecision = Object.freeze({
  Approved: "approved",
  Rejected: "rejected",
} as const);
export type VillageReviewDecision = (typeof VillageReviewDecision)[keyof typeof VillageReviewDecision];
export const AllVillageReviewDecisions = Object.freeze([VillageReviewDecision.Approved, VillageReviewDecision.Rejected]) as readonly VillageReviewDecision[];
export function isVillageReviewDecision(value: unknown): value is VillageReviewDecision {
  return zVillageReviewDecision.safeParse(value).success;
}

export const VillageShareEventActor = Object.freeze({
  None: "",
  Owner: "owner",
  Collective: "collective",
  Moderator: "moderator",
} as const);
export type VillageShareEventActor = (typeof VillageShareEventActor)[keyof typeof VillageShareEventActor];
export const AllVillageShareEventActors = Object.freeze([VillageShareEventActor.None, VillageShareEventActor.Owner, VillageShareEventActor.Collective, VillageShareEventActor.Moderator]) as readonly VillageShareEventActor[];
export function isVillageShareEventActor(value: unknown): value is VillageShareEventActor {
  return zVillageShareEventActor.safeParse(value).success;
}

export const VillageShareStatus = Object.freeze({
  Pending: "pending",
  Approved: "approved",
  Rejected: "rejected",
  Retracted: "retracted",
  Revoked: "revoked",
} as const);
export type VillageShareStatus = (typeof VillageShareStatus)[keyof typeof VillageShareStatus];
export const AllVillageShareStatuses = Object.freeze([VillageShareStatus.Pending, VillageShareStatus.Approved, VillageShareStatus.Rejected, VillageShareStatus.Retracted, VillageShareStatus.Revoked]) as readonly VillageShareStatus[];
export function isVillageShareStatus(value: unknown): value is VillageShareStatus {
  return zVillageShareStatus.safeParse(value).success;
}

export const VillageTranscriptDeletionPolicy = Object.freeze({
  UserChoice: "user_choice",
  Mandatory: "mandatory",
} as const);
export type VillageTranscriptDeletionPolicy = (typeof VillageTranscriptDeletionPolicy)[keyof typeof VillageTranscriptDeletionPolicy];
export const AllVillageTranscriptDeletionPolicies = Object.freeze([VillageTranscriptDeletionPolicy.UserChoice, VillageTranscriptDeletionPolicy.Mandatory]) as readonly VillageTranscriptDeletionPolicy[];
export function isVillageTranscriptDeletionPolicy(value: unknown): value is VillageTranscriptDeletionPolicy {
  return zVillageTranscriptDeletionPolicy.safeParse(value).success;
}

export const VillageTranscriptVisibility = Object.freeze({
  Private: "private",
  Shared: "shared",
  Public: "public",
} as const);
export type VillageTranscriptVisibility = (typeof VillageTranscriptVisibility)[keyof typeof VillageTranscriptVisibility];
export const AllVillageTranscriptVisibilities = Object.freeze([VillageTranscriptVisibility.Private, VillageTranscriptVisibility.Shared, VillageTranscriptVisibility.Public]) as readonly VillageTranscriptVisibility[];
export function isVillageTranscriptVisibility(value: unknown): value is VillageTranscriptVisibility {
  return zVillageTranscriptVisibility.safeParse(value).success;
}

export const Visibility = Object.freeze({
  Private: "private",
  Group: "group",
  Public: "public",
} as const);
export type Visibility = (typeof Visibility)[keyof typeof Visibility];
export const AllVisibilities = Object.freeze([Visibility.Private, Visibility.Group, Visibility.Public]) as readonly Visibility[];
export function isVisibility(value: unknown): value is Visibility {
  return zVisibility.safeParse(value).success;
}

export const UsageScope = Object.freeze({
  Assistant: "assistant",
  Tool: "tool",
  Summary: "summary",
} as const);
export type UsageScope = (typeof UsageScope)[keyof typeof UsageScope];
export const AllUsageScopes = Object.freeze([UsageScope.Assistant, UsageScope.Tool, UsageScope.Summary]) as readonly UsageScope[];
export function isUsageScope(value: unknown): value is UsageScope {
  return zUsageScope.safeParse(value).success;
}

export const UsageCompleteness = Object.freeze({
  Complete: "complete",
  Partial: "partial",
  Unknown: "unknown",
} as const);
export type UsageCompleteness = (typeof UsageCompleteness)[keyof typeof UsageCompleteness];
export const AllUsageCompleteness = Object.freeze([UsageCompleteness.Complete, UsageCompleteness.Partial, UsageCompleteness.Unknown]) as readonly UsageCompleteness[];
export function isUsageCompleteness(value: unknown): value is UsageCompleteness {
  return zUsageCompleteness.safeParse(value).success;
}

export const NativeMetadataKind = Object.freeze({
  PiCustomData: "pi.custom.data",
  PiCustomMessageDetails: "pi.custommessage.details",
  PiToolResultDetails: "pi.toolresult.details",
  PiCompactionDetails: "pi.compaction.details",
  PiBranchSummaryDetails: "pi.branchsummary.details",
} as const);
export type NativeMetadataKind = (typeof NativeMetadataKind)[keyof typeof NativeMetadataKind];
export const AllNativeMetadataKinds = Object.freeze([NativeMetadataKind.PiCustomData, NativeMetadataKind.PiCustomMessageDetails, NativeMetadataKind.PiToolResultDetails, NativeMetadataKind.PiCompactionDetails, NativeMetadataKind.PiBranchSummaryDetails]) as readonly NativeMetadataKind[];
export function isNativeMetadataKind(value: unknown): value is NativeMetadataKind {
  return zNativeMetadataKind.safeParse(value).success;
}

export const NativeMetadataSourceType = Object.freeze({
  PiCustom: "pi.custom",
  PiCustomMessage: "pi.custom_message",
  PiMessage: "pi.message",
  PiCompaction: "pi.compaction",
  PiBranchSummary: "pi.branch_summary",
} as const);
export type NativeMetadataSourceType = (typeof NativeMetadataSourceType)[keyof typeof NativeMetadataSourceType];
export const AllNativeMetadataSourceTypes = Object.freeze([NativeMetadataSourceType.PiCustom, NativeMetadataSourceType.PiCustomMessage, NativeMetadataSourceType.PiMessage, NativeMetadataSourceType.PiCompaction, NativeMetadataSourceType.PiBranchSummary]) as readonly NativeMetadataSourceType[];
export function isNativeMetadataSourceType(value: unknown): value is NativeMetadataSourceType {
  return zNativeMetadataSourceType.safeParse(value).success;
}

export const NativePiMessageRole = Object.freeze({
  ToolResult: "toolResult",
} as const);
export type NativePiMessageRole = (typeof NativePiMessageRole)[keyof typeof NativePiMessageRole];
export const AllNativePiMessageRoles = Object.freeze([NativePiMessageRole.ToolResult]) as readonly NativePiMessageRole[];
export function isNativePiMessageRole(value: unknown): value is NativePiMessageRole {
  return zNativePiMessageRole.safeParse(value).success;
}

export const DigestItemKind = Object.freeze({
  Session: "session",
  Prompt: "prompt",
  Skill: "skill",
  Commit: "commit",
} as const);
export type DigestItemKind = (typeof DigestItemKind)[keyof typeof DigestItemKind];
export const AllDigestItemKinds = Object.freeze([DigestItemKind.Session, DigestItemKind.Prompt, DigestItemKind.Skill, DigestItemKind.Commit]) as readonly DigestItemKind[];
export function isDigestItemKind(value: unknown): value is DigestItemKind {
  return zDigestItemKind.safeParse(value).success;
}

export const VillagePullRequestAttachmentState = Object.freeze({
  Requested: "requested",
  Waiting: "waiting",
  Preview: "preview",
  Attached: "attached",
  Detached: "detached",
} as const);
export type VillagePullRequestAttachmentState = (typeof VillagePullRequestAttachmentState)[keyof typeof VillagePullRequestAttachmentState];
export const AllVillagePullRequestAttachmentStates = Object.freeze([VillagePullRequestAttachmentState.Requested, VillagePullRequestAttachmentState.Waiting, VillagePullRequestAttachmentState.Preview, VillagePullRequestAttachmentState.Attached, VillagePullRequestAttachmentState.Detached]) as readonly VillagePullRequestAttachmentState[];
export function isVillagePullRequestAttachmentState(value: unknown): value is VillagePullRequestAttachmentState {
  return zVillagePullRequestAttachmentState.safeParse(value).success;
}

export const VillagePromptsCheckMode = Object.freeze({
  Informational: "informational",
  Required: "required",
} as const);
export type VillagePromptsCheckMode = (typeof VillagePromptsCheckMode)[keyof typeof VillagePromptsCheckMode];
export const AllVillagePromptsCheckModes = Object.freeze([VillagePromptsCheckMode.Informational, VillagePromptsCheckMode.Required]) as readonly VillagePromptsCheckMode[];
export function isVillagePromptsCheckMode(value: unknown): value is VillagePromptsCheckMode {
  return zVillagePromptsCheckMode.safeParse(value).success;
}

export const VillagePublicOwnerKind = Object.freeze({
  Profile: "profile",
  Anonymous: "anonymous",
} as const);
export type VillagePublicOwnerKind = (typeof VillagePublicOwnerKind)[keyof typeof VillagePublicOwnerKind];
export const AllVillagePublicOwnerKinds = Object.freeze([VillagePublicOwnerKind.Profile, VillagePublicOwnerKind.Anonymous]) as readonly VillagePublicOwnerKind[];
export function isVillagePublicOwnerKind(value: unknown): value is VillagePublicOwnerKind {
  return zVillagePublicOwnerKind.safeParse(value).success;
}

export const SessionRelationshipKind = Object.freeze({
  StartedBy: "started_by",
  ContextFrom: "context_from",
} as const);
export type SessionRelationshipKind = (typeof SessionRelationshipKind)[keyof typeof SessionRelationshipKind];
export const AllSessionRelationshipKinds = Object.freeze([SessionRelationshipKind.StartedBy, SessionRelationshipKind.ContextFrom]) as readonly SessionRelationshipKind[];
export function isSessionRelationshipKind(value: unknown): value is SessionRelationshipKind {
  return zSessionRelationshipKind.safeParse(value).success;
}

export const RelationshipTargetState = Object.freeze({
  Known: "target_known",
  KnownRetained: "target_known_retained",
  ExplicitNone: "explicit_none",
  Unknown: "unknown",
  ConflictingCurrentNativeEvidence: "conflicting_current_native_evidence",
} as const);
export type RelationshipTargetState = (typeof RelationshipTargetState)[keyof typeof RelationshipTargetState];
export const AllRelationshipTargetStates = Object.freeze([RelationshipTargetState.Known, RelationshipTargetState.KnownRetained, RelationshipTargetState.ExplicitNone, RelationshipTargetState.Unknown, RelationshipTargetState.ConflictingCurrentNativeEvidence]) as readonly RelationshipTargetState[];
export function isRelationshipTargetState(value: unknown): value is RelationshipTargetState {
  return zRelationshipTargetState.safeParse(value).success;
}

export const EvidenceKind = Object.freeze({
  NativeTyped: "native_typed",
  LifecycleTyped: "lifecycle_typed",
  ExistingAdapter: "existing_adapter",
  RetainedLastGood: "retained_last_good",
  Unknown: "unknown",
  Conflict: "conflict",
} as const);
export type EvidenceKind = (typeof EvidenceKind)[keyof typeof EvidenceKind];
export const AllEvidenceKinds = Object.freeze([EvidenceKind.NativeTyped, EvidenceKind.LifecycleTyped, EvidenceKind.ExistingAdapter, EvidenceKind.RetainedLastGood, EvidenceKind.Unknown, EvidenceKind.Conflict]) as readonly EvidenceKind[];
export function isEvidenceKind(value: unknown): value is EvidenceKind {
  return zEvidenceKind.safeParse(value).success;
}

export const SessionTeammateEvidence = Object.freeze({
  NativeDynamicWorkflow: "native_dynamic_workflow",
} as const);
export type SessionTeammateEvidence = (typeof SessionTeammateEvidence)[keyof typeof SessionTeammateEvidence];
export const AllSessionTeammateEvidences = Object.freeze([SessionTeammateEvidence.NativeDynamicWorkflow]) as readonly SessionTeammateEvidence[];
export function isSessionTeammateEvidence(value: unknown): value is SessionTeammateEvidence {
  return zSessionTeammateEvidence.safeParse(value).success;
}

export const SessionPurpose = Object.freeze({
  Interaction: "interaction",
  DelegatedWork: "delegated_work",
  HelperReview: "helper_review",
  Unknown: "unknown",
} as const);
export type SessionPurpose = (typeof SessionPurpose)[keyof typeof SessionPurpose];
export const AllSessionPurposes = Object.freeze([SessionPurpose.Interaction, SessionPurpose.DelegatedWork, SessionPurpose.HelperReview, SessionPurpose.Unknown]) as readonly SessionPurpose[];
export function isSessionPurpose(value: unknown): value is SessionPurpose {
  return zSessionPurpose.safeParse(value).success;
}

export const ContentOrigin = Object.freeze({
  SubmittedInput: "submitted_input",
  HarnessContext: "harness_context",
  AgentOutput: "agent_output",
  AgentCommunication: "agent_communication",
  ToolActivity: "tool_activity",
  SystemControl: "system_control",
  GeneratedSummary: "generated_summary",
  Unknown: "unknown",
} as const);
export type ContentOrigin = (typeof ContentOrigin)[keyof typeof ContentOrigin];
export const AllContentOrigins = Object.freeze([ContentOrigin.SubmittedInput, ContentOrigin.HarnessContext, ContentOrigin.AgentOutput, ContentOrigin.AgentCommunication, ContentOrigin.ToolActivity, ContentOrigin.SystemControl, ContentOrigin.GeneratedSummary, ContentOrigin.Unknown]) as readonly ContentOrigin[];
export function isContentOrigin(value: unknown): value is ContentOrigin {
  return zContentOrigin.safeParse(value).success;
}

export const ActorOrigin = Object.freeze({
  Operator: "operator",
  AgentDelegate: "agent_delegate",
  Harness: "harness",
  Unknown: "unknown",
} as const);
export type ActorOrigin = (typeof ActorOrigin)[keyof typeof ActorOrigin];
export const AllActorOrigins = Object.freeze([ActorOrigin.Operator, ActorOrigin.AgentDelegate, ActorOrigin.Harness, ActorOrigin.Unknown]) as readonly ActorOrigin[];
export function isActorOrigin(value: unknown): value is ActorOrigin {
  return zActorOrigin.safeParse(value).success;
}

export const DeliveryOrigin = Object.freeze({
  SessionAdmission: "session_admission",
  GuardianReview: "guardian_review",
  SubagentDelivery: "subagent_delivery",
  InheritedContext: "inherited_context",
  ToolDelivery: "tool_delivery",
  SystemLifecycle: "system_lifecycle",
  Unknown: "unknown",
} as const);
export type DeliveryOrigin = (typeof DeliveryOrigin)[keyof typeof DeliveryOrigin];
export const AllDeliveryOrigins = Object.freeze([DeliveryOrigin.SessionAdmission, DeliveryOrigin.GuardianReview, DeliveryOrigin.SubagentDelivery, DeliveryOrigin.InheritedContext, DeliveryOrigin.ToolDelivery, DeliveryOrigin.SystemLifecycle, DeliveryOrigin.Unknown]) as readonly DeliveryOrigin[];
export function isDeliveryOrigin(value: unknown): value is DeliveryOrigin {
  return zDeliveryOrigin.safeParse(value).success;
}

export const ContentOwnership = Object.freeze({
  Local: "local",
  Inherited: "inherited",
  Uncertain: "uncertain",
} as const);
export type ContentOwnership = (typeof ContentOwnership)[keyof typeof ContentOwnership];
export const AllContentOwnerships = Object.freeze([ContentOwnership.Local, ContentOwnership.Inherited, ContentOwnership.Uncertain]) as readonly ContentOwnership[];
export function isContentOwnership(value: unknown): value is ContentOwnership {
  return zContentOwnership.safeParse(value).success;
}

export const InputModality = Object.freeze({
  None: "none",
  Text: "text",
  Media: "media",
  UserAction: "user_action",
  Mixed: "mixed",
  Unknown: "unknown",
} as const);
export type InputModality = (typeof InputModality)[keyof typeof InputModality];
export const AllInputModalities = Object.freeze([InputModality.None, InputModality.Text, InputModality.Media, InputModality.UserAction, InputModality.Mixed, InputModality.Unknown]) as readonly InputModality[];
export function isInputModality(value: unknown): value is InputModality {
  return zInputModality.safeParse(value).success;
}

export const PublicSourceAnchorKind = Object.freeze({
  General: "general_source_session",
  Before: "before_redacted_entry",
  Through: "through_redacted_entry",
} as const);
export type PublicSourceAnchorKind = (typeof PublicSourceAnchorKind)[keyof typeof PublicSourceAnchorKind];
export const AllPublicSourceAnchorKinds = Object.freeze([PublicSourceAnchorKind.General, PublicSourceAnchorKind.Before, PublicSourceAnchorKind.Through]) as readonly PublicSourceAnchorKind[];
export function isPublicSourceAnchorKind(value: unknown): value is PublicSourceAnchorKind {
  return zPublicSourceAnchorKind.safeParse(value).success;
}

export const EarlierHistoryState = Object.freeze({
  UncertainMigrated: "uncertain_migrated",
  UncertainUnresolved: "uncertain_unresolved",
} as const);
export type EarlierHistoryState = (typeof EarlierHistoryState)[keyof typeof EarlierHistoryState];
export const AllEarlierHistoryStates = Object.freeze([EarlierHistoryState.UncertainMigrated, EarlierHistoryState.UncertainUnresolved]) as readonly EarlierHistoryState[];
export function isEarlierHistoryState(value: unknown): value is EarlierHistoryState {
  return zEarlierHistoryState.safeParse(value).success;
}

export const RelationshipNavigationStatus = Object.freeze({
  Resolved: "resolved",
  GeneralLinkOnly: "general_link_only",
  KnownUnavailable: "known_unavailable",
  Inaccessible: "inaccessible",
  Unknown: "unknown",
  Conflicting: "conflicting",
} as const);
export type RelationshipNavigationStatus = (typeof RelationshipNavigationStatus)[keyof typeof RelationshipNavigationStatus];
export const AllRelationshipNavigationStatuses = Object.freeze([RelationshipNavigationStatus.Resolved, RelationshipNavigationStatus.GeneralLinkOnly, RelationshipNavigationStatus.KnownUnavailable, RelationshipNavigationStatus.Inaccessible, RelationshipNavigationStatus.Unknown, RelationshipNavigationStatus.Conflicting]) as readonly RelationshipNavigationStatus[];
export function isRelationshipNavigationStatus(value: unknown): value is RelationshipNavigationStatus {
  return zRelationshipNavigationStatus.safeParse(value).success;
}

export const SessionListItemKind = Object.freeze({
  Transcript: "transcript",
  ContextContainer: "context_container",
} as const);
export type SessionListItemKind = (typeof SessionListItemKind)[keyof typeof SessionListItemKind];
export const AllSessionListItemKinds = Object.freeze([SessionListItemKind.Transcript, SessionListItemKind.ContextContainer]) as readonly SessionListItemKind[];
export function isSessionListItemKind(value: unknown): value is SessionListItemKind {
  return zSessionListItemKind.safeParse(value).success;
}

export const SyncStatus = Object.freeze({
  New: "new",
  Updated: "updated",
  Synced: "synced",
  Held: "held",
} as const);
export type SyncStatus = (typeof SyncStatus)[keyof typeof SyncStatus];
export const AllSyncStatuses = Object.freeze([SyncStatus.New, SyncStatus.Updated, SyncStatus.Synced, SyncStatus.Held]) as readonly SyncStatus[];
export function isSyncStatus(value: unknown): value is SyncStatus {
  return zSyncStatus.safeParse(value).success;
}

export const SyncHoldReason = Object.freeze({
  MetricsMissing: "metrics_missing",
  MetadataMissing: "metadata_missing",
} as const);
export type SyncHoldReason = (typeof SyncHoldReason)[keyof typeof SyncHoldReason];
export const AllSyncHoldReasons = Object.freeze([SyncHoldReason.MetricsMissing, SyncHoldReason.MetadataMissing]) as readonly SyncHoldReason[];
export function isSyncHoldReason(value: unknown): value is SyncHoldReason {
  return zSyncHoldReason.safeParse(value).success;
}

export const SyncPushSessionStatus = Object.freeze({
  New: "new",
  Updated: "updated",
  Skipped: "skipped",
  Error: "error",
  Held: "held",
} as const);
export type SyncPushSessionStatus = (typeof SyncPushSessionStatus)[keyof typeof SyncPushSessionStatus];
export const AllSyncPushSessionStatuses = Object.freeze([SyncPushSessionStatus.New, SyncPushSessionStatus.Updated, SyncPushSessionStatus.Skipped, SyncPushSessionStatus.Error, SyncPushSessionStatus.Held]) as readonly SyncPushSessionStatus[];
export function isSyncPushSessionStatus(value: unknown): value is SyncPushSessionStatus {
  return zSyncPushSessionStatus.safeParse(value).success;
}

export const SyncPushStep = Object.freeze({
  Content: "content",
  AddCollective: "add_collective",
  RemoveCollective: "remove_collective",
} as const);
export type SyncPushStep = (typeof SyncPushStep)[keyof typeof SyncPushStep];
export const AllSyncPushSteps = Object.freeze([SyncPushStep.Content, SyncPushStep.AddCollective, SyncPushStep.RemoveCollective]) as readonly SyncPushStep[];
export function isSyncPushStep(value: unknown): value is SyncPushStep {
  return zSyncPushStep.safeParse(value).success;
}

export const SyncPushStepOutcome = Object.freeze({
  Succeeded: "succeeded",
  PendingApproval: "pending_approval",
  Skipped: "skipped",
  Failed: "failed",
  NotAttempted: "not_attempted",
} as const);
export type SyncPushStepOutcome = (typeof SyncPushStepOutcome)[keyof typeof SyncPushStepOutcome];
export const AllSyncPushStepOutcomes = Object.freeze([SyncPushStepOutcome.Succeeded, SyncPushStepOutcome.PendingApproval, SyncPushStepOutcome.Skipped, SyncPushStepOutcome.Failed, SyncPushStepOutcome.NotAttempted]) as readonly SyncPushStepOutcome[];
export function isSyncPushStepOutcome(value: unknown): value is SyncPushStepOutcome {
  return zSyncPushStepOutcome.safeParse(value).success;
}

export const SyncLoginStatus = Object.freeze({
  Pending: "pending",
  AlreadyAuthenticated: "already_authenticated",
} as const);
export type SyncLoginStatus = (typeof SyncLoginStatus)[keyof typeof SyncLoginStatus];
export const AllSyncLoginStatuses = Object.freeze([SyncLoginStatus.Pending, SyncLoginStatus.AlreadyAuthenticated]) as readonly SyncLoginStatus[];
export function isSyncLoginStatus(value: unknown): value is SyncLoginStatus {
  return zSyncLoginStatus.safeParse(value).success;
}

export const SyncLogoutStatus = Object.freeze({
  LoggedOut: "logged_out",
  AlreadyLoggedOut: "already_logged_out",
} as const);
export type SyncLogoutStatus = (typeof SyncLogoutStatus)[keyof typeof SyncLogoutStatus];
export const AllSyncLogoutStatuses = Object.freeze([SyncLogoutStatus.LoggedOut, SyncLogoutStatus.AlreadyLoggedOut]) as readonly SyncLogoutStatus[];
export function isSyncLogoutStatus(value: unknown): value is SyncLogoutStatus {
  return zSyncLogoutStatus.safeParse(value).success;
}

export const LocalPublicationState = Object.freeze({
  Unpublished: "unpublished",
  Published: "published",
} as const);
export type LocalPublicationState = (typeof LocalPublicationState)[keyof typeof LocalPublicationState];
export const AllLocalPublicationStates = Object.freeze([LocalPublicationState.Unpublished, LocalPublicationState.Published]) as readonly LocalPublicationState[];
export function isLocalPublicationState(value: unknown): value is LocalPublicationState {
  return zLocalPublicationState.safeParse(value).success;
}

export const LocalCollectiveSuggestionReason = Object.freeze({
  LinkedRepository: "linked_repository",
  LinkedGithubOrg: "linked_github_org",
} as const);
export type LocalCollectiveSuggestionReason = (typeof LocalCollectiveSuggestionReason)[keyof typeof LocalCollectiveSuggestionReason];
export const AllLocalCollectiveSuggestionReasons = Object.freeze([LocalCollectiveSuggestionReason.LinkedRepository, LocalCollectiveSuggestionReason.LinkedGithubOrg]) as readonly LocalCollectiveSuggestionReason[];
export function isLocalCollectiveSuggestionReason(value: unknown): value is LocalCollectiveSuggestionReason {
  return zLocalCollectiveSuggestionReason.safeParse(value).success;
}

export const LocalSettingKind = Object.freeze({
  Boolean: "boolean",
  Integer: "integer",
  String: "string",
  StringList: "string_list",
  Choice: "choice",
  Structured: "structured",
} as const);
export type LocalSettingKind = (typeof LocalSettingKind)[keyof typeof LocalSettingKind];
export const AllLocalSettingKinds = Object.freeze([LocalSettingKind.Boolean, LocalSettingKind.Integer, LocalSettingKind.String, LocalSettingKind.StringList, LocalSettingKind.Choice, LocalSettingKind.Structured]) as readonly LocalSettingKind[];
export function isLocalSettingKind(value: unknown): value is LocalSettingKind {
  return zLocalSettingKind.safeParse(value).success;
}

export const LocalKickstartScope = Object.freeze({
  Full: "full",
  Selection: "selection",
} as const);
export type LocalKickstartScope = (typeof LocalKickstartScope)[keyof typeof LocalKickstartScope];
export const AllLocalKickstartScopes = Object.freeze([LocalKickstartScope.Full, LocalKickstartScope.Selection]) as readonly LocalKickstartScope[];
export function isLocalKickstartScope(value: unknown): value is LocalKickstartScope {
  return zLocalKickstartScope.safeParse(value).success;
}

export const LocalKickstartStopKind = Object.freeze({
  Connection: "connection",
  Config: "config",
  Review: "review",
  Ingest: "ingest",
} as const);
export type LocalKickstartStopKind = (typeof LocalKickstartStopKind)[keyof typeof LocalKickstartStopKind];
export const AllLocalKickstartStopKinds = Object.freeze([LocalKickstartStopKind.Connection, LocalKickstartStopKind.Config, LocalKickstartStopKind.Review, LocalKickstartStopKind.Ingest]) as readonly LocalKickstartStopKind[];
export function isLocalKickstartStopKind(value: unknown): value is LocalKickstartStopKind {
  return zLocalKickstartStopKind.safeParse(value).success;
}

export const LocalKickstartFieldKind = Object.freeze({
  Toggle: "toggle",
  Choice: "choice",
  Action: "action",
  Tree: "tree",
  Notice: "notice",
} as const);
export type LocalKickstartFieldKind = (typeof LocalKickstartFieldKind)[keyof typeof LocalKickstartFieldKind];
export const AllLocalKickstartFieldKinds = Object.freeze([LocalKickstartFieldKind.Toggle, LocalKickstartFieldKind.Choice, LocalKickstartFieldKind.Action, LocalKickstartFieldKind.Tree, LocalKickstartFieldKind.Notice]) as readonly LocalKickstartFieldKind[];
export function isLocalKickstartFieldKind(value: unknown): value is LocalKickstartFieldKind {
  return zLocalKickstartFieldKind.safeParse(value).success;
}

export const LocalKickstartLineKind = Object.freeze({
  Text: "text",
  Label: "label",
  Before: "before",
  After: "after",
  Spacer: "spacer",
} as const);
export type LocalKickstartLineKind = (typeof LocalKickstartLineKind)[keyof typeof LocalKickstartLineKind];
export const AllLocalKickstartLineKinds = Object.freeze([LocalKickstartLineKind.Text, LocalKickstartLineKind.Label, LocalKickstartLineKind.Before, LocalKickstartLineKind.After, LocalKickstartLineKind.Spacer]) as readonly LocalKickstartLineKind[];
export function isLocalKickstartLineKind(value: unknown): value is LocalKickstartLineKind {
  return zLocalKickstartLineKind.safeParse(value).success;
}

export const LocalKickstartNodeKind = Object.freeze({
  Project: "project",
  Group: "group",
  Branch: "branch",
  Session: "session",
} as const);
export type LocalKickstartNodeKind = (typeof LocalKickstartNodeKind)[keyof typeof LocalKickstartNodeKind];
export const AllLocalKickstartNodeKinds = Object.freeze([LocalKickstartNodeKind.Project, LocalKickstartNodeKind.Group, LocalKickstartNodeKind.Branch, LocalKickstartNodeKind.Session]) as readonly LocalKickstartNodeKind[];
export function isLocalKickstartNodeKind(value: unknown): value is LocalKickstartNodeKind {
  return zLocalKickstartNodeKind.safeParse(value).success;
}

export const LocalKickstartNodeState = Object.freeze({
  None: "none",
  Some: "some",
  All: "all",
} as const);
export type LocalKickstartNodeState = (typeof LocalKickstartNodeState)[keyof typeof LocalKickstartNodeState];
export const AllLocalKickstartNodeStates = Object.freeze([LocalKickstartNodeState.None, LocalKickstartNodeState.Some, LocalKickstartNodeState.All]) as readonly LocalKickstartNodeState[];
export function isLocalKickstartNodeState(value: unknown): value is LocalKickstartNodeState {
  return zLocalKickstartNodeState.safeParse(value).success;
}

export const LocalKickstartScanState = Object.freeze({
  Idle: "idle",
  Running: "running",
  Done: "done",
  Error: "error",
} as const);
export type LocalKickstartScanState = (typeof LocalKickstartScanState)[keyof typeof LocalKickstartScanState];
export const AllLocalKickstartScanStates = Object.freeze([LocalKickstartScanState.Idle, LocalKickstartScanState.Running, LocalKickstartScanState.Done, LocalKickstartScanState.Error]) as readonly LocalKickstartScanState[];
export function isLocalKickstartScanState(value: unknown): value is LocalKickstartScanState {
  return zLocalKickstartScanState.safeParse(value).success;
}

export const LocalKickstartIngestState = Object.freeze({
  Idle: "idle",
  Running: "running",
  Busy: "busy",
  Failed: "failed",
} as const);
export type LocalKickstartIngestState = (typeof LocalKickstartIngestState)[keyof typeof LocalKickstartIngestState];
export const AllLocalKickstartIngestStates = Object.freeze([LocalKickstartIngestState.Idle, LocalKickstartIngestState.Running, LocalKickstartIngestState.Busy, LocalKickstartIngestState.Failed]) as readonly LocalKickstartIngestState[];
export function isLocalKickstartIngestState(value: unknown): value is LocalKickstartIngestState {
  return zLocalKickstartIngestState.safeParse(value).success;
}

export const LocalKickstartResolution = Object.freeze({
  Mine: "mine",
  Theirs: "theirs",
} as const);
export type LocalKickstartResolution = (typeof LocalKickstartResolution)[keyof typeof LocalKickstartResolution];
export const AllLocalKickstartResolutions = Object.freeze([LocalKickstartResolution.Mine, LocalKickstartResolution.Theirs]) as readonly LocalKickstartResolution[];
export function isLocalKickstartResolution(value: unknown): value is LocalKickstartResolution {
  return zLocalKickstartResolution.safeParse(value).success;
}

export const LocalKickstartNextStep = Object.freeze({
  Config: "config",
  WebStart: "web-start",
  VillageLogin: "village-login",
  VillagePush: "village-push",
  AutoPublish: "auto-publish-settings",
} as const);
export type LocalKickstartNextStep = (typeof LocalKickstartNextStep)[keyof typeof LocalKickstartNextStep];
export const AllLocalKickstartNextSteps = Object.freeze([LocalKickstartNextStep.Config, LocalKickstartNextStep.WebStart, LocalKickstartNextStep.VillageLogin, LocalKickstartNextStep.VillagePush, LocalKickstartNextStep.AutoPublish]) as readonly LocalKickstartNextStep[];
export function isLocalKickstartNextStep(value: unknown): value is LocalKickstartNextStep {
  return zLocalKickstartNextStep.safeParse(value).success;
}

export const LocalKickstartDiffLineType = Object.freeze({
  Add: "add",
  Del: "del",
  Ctx: "ctx",
} as const);
export type LocalKickstartDiffLineType = (typeof LocalKickstartDiffLineType)[keyof typeof LocalKickstartDiffLineType];
export const AllLocalKickstartDiffLineTypes = Object.freeze([LocalKickstartDiffLineType.Add, LocalKickstartDiffLineType.Del, LocalKickstartDiffLineType.Ctx]) as readonly LocalKickstartDiffLineType[];
export function isLocalKickstartDiffLineType(value: unknown): value is LocalKickstartDiffLineType {
  return zLocalKickstartDiffLineType.safeParse(value).success;
}

export const LocalKickstartEventKind = Object.freeze({
  Session: "session",
  Answers: "answers",
  Conflicts: "conflicts",
  Scan: "scan",
  Ingest: "ingest",
  Closed: "closed",
} as const);
export type LocalKickstartEventKind = (typeof LocalKickstartEventKind)[keyof typeof LocalKickstartEventKind];
export const AllLocalKickstartEventKinds = Object.freeze([LocalKickstartEventKind.Session, LocalKickstartEventKind.Answers, LocalKickstartEventKind.Conflicts, LocalKickstartEventKind.Scan, LocalKickstartEventKind.Ingest, LocalKickstartEventKind.Closed]) as readonly LocalKickstartEventKind[];
export function isLocalKickstartEventKind(value: unknown): value is LocalKickstartEventKind {
  return zLocalKickstartEventKind.safeParse(value).success;
}

export const LocalKickstartPreviewOp = Object.freeze({
  Open: "open",
  More: "more",
  Close: "close",
} as const);
export type LocalKickstartPreviewOp = (typeof LocalKickstartPreviewOp)[keyof typeof LocalKickstartPreviewOp];
export const AllLocalKickstartPreviewOps = Object.freeze([LocalKickstartPreviewOp.Open, LocalKickstartPreviewOp.More, LocalKickstartPreviewOp.Close]) as readonly LocalKickstartPreviewOp[];
export function isLocalKickstartPreviewOp(value: unknown): value is LocalKickstartPreviewOp {
  return zLocalKickstartPreviewOp.safeParse(value).success;
}

export const LocalKickstartPreviewContinuation = Object.freeze({
  More: "more",
  End: "end",
  Cut: "cut",
} as const);
export type LocalKickstartPreviewContinuation = (typeof LocalKickstartPreviewContinuation)[keyof typeof LocalKickstartPreviewContinuation];
export const AllLocalKickstartPreviewContinuations = Object.freeze([LocalKickstartPreviewContinuation.More, LocalKickstartPreviewContinuation.End, LocalKickstartPreviewContinuation.Cut]) as readonly LocalKickstartPreviewContinuation[];
export function isLocalKickstartPreviewContinuation(value: unknown): value is LocalKickstartPreviewContinuation {
  return zLocalKickstartPreviewContinuation.safeParse(value).success;
}

export const LocalKickstartPreviewEventKind = Object.freeze({
  Opened: "opened",
  Segment: "segment",
  Closed: "closed",
} as const);
export type LocalKickstartPreviewEventKind = (typeof LocalKickstartPreviewEventKind)[keyof typeof LocalKickstartPreviewEventKind];
export const AllLocalKickstartPreviewEventKinds = Object.freeze([LocalKickstartPreviewEventKind.Opened, LocalKickstartPreviewEventKind.Segment, LocalKickstartPreviewEventKind.Closed]) as readonly LocalKickstartPreviewEventKind[];
export function isLocalKickstartPreviewEventKind(value: unknown): value is LocalKickstartPreviewEventKind {
  return zLocalKickstartPreviewEventKind.safeParse(value).success;
}

export const LocalKickstartPreviewCutReason = Object.freeze({
  SourceBudget: "source_budget",
  SerializedBudget: "serialized_budget",
  TurnBudget: "turn_budget",
  OversizedRecord: "oversized_record",
  OriginNotResumable: "origin_not_resumable",
} as const);
export type LocalKickstartPreviewCutReason = (typeof LocalKickstartPreviewCutReason)[keyof typeof LocalKickstartPreviewCutReason];
export const AllLocalKickstartPreviewCutReasons = Object.freeze([LocalKickstartPreviewCutReason.SourceBudget, LocalKickstartPreviewCutReason.SerializedBudget, LocalKickstartPreviewCutReason.TurnBudget, LocalKickstartPreviewCutReason.OversizedRecord, LocalKickstartPreviewCutReason.OriginNotResumable]) as readonly LocalKickstartPreviewCutReason[];
export function isLocalKickstartPreviewCutReason(value: unknown): value is LocalKickstartPreviewCutReason {
  return zLocalKickstartPreviewCutReason.safeParse(value).success;
}

export const LocalKickstartPreviewCloseReason = Object.freeze({
  ClosedByClient: "closed_by_client",
  Replaced: "replaced",
  SessionClosed: "session_closed",
  Rescanned: "rescanned",
  Idle: "idle",
  Evicted: "evicted",
  Shutdown: "shutdown",
  SourceChanged: "source_changed",
  LateResult: "late_result",
  Unreadable: "unreadable",
  OutOfOrder: "out_of_order",
  SlowClient: "slow_client",
  RowNotFound: "row_not_found",
  NotASessionRow: "not_a_session_row",
  Unavailable: "unavailable",
  NoTurns: "no_turns",
  Capacity: "capacity",
} as const);
export type LocalKickstartPreviewCloseReason = (typeof LocalKickstartPreviewCloseReason)[keyof typeof LocalKickstartPreviewCloseReason];
export const AllLocalKickstartPreviewCloseReasons = Object.freeze([LocalKickstartPreviewCloseReason.ClosedByClient, LocalKickstartPreviewCloseReason.Replaced, LocalKickstartPreviewCloseReason.SessionClosed, LocalKickstartPreviewCloseReason.Rescanned, LocalKickstartPreviewCloseReason.Idle, LocalKickstartPreviewCloseReason.Evicted, LocalKickstartPreviewCloseReason.Shutdown, LocalKickstartPreviewCloseReason.SourceChanged, LocalKickstartPreviewCloseReason.LateResult, LocalKickstartPreviewCloseReason.Unreadable, LocalKickstartPreviewCloseReason.OutOfOrder, LocalKickstartPreviewCloseReason.SlowClient, LocalKickstartPreviewCloseReason.RowNotFound, LocalKickstartPreviewCloseReason.NotASessionRow, LocalKickstartPreviewCloseReason.Unavailable, LocalKickstartPreviewCloseReason.NoTurns, LocalKickstartPreviewCloseReason.Capacity]) as readonly LocalKickstartPreviewCloseReason[];
export function isLocalKickstartPreviewCloseReason(value: unknown): value is LocalKickstartPreviewCloseReason {
  return zLocalKickstartPreviewCloseReason.safeParse(value).success;
}

export const LocalKickstartCloseReason = Object.freeze({
  Committed: "committed",
  Expired: "expired",
} as const);
export type LocalKickstartCloseReason = (typeof LocalKickstartCloseReason)[keyof typeof LocalKickstartCloseReason];
export const AllLocalKickstartCloseReasons = Object.freeze([LocalKickstartCloseReason.Committed, LocalKickstartCloseReason.Expired]) as readonly LocalKickstartCloseReason[];
export function isLocalKickstartCloseReason(value: unknown): value is LocalKickstartCloseReason {
  return zLocalKickstartCloseReason.safeParse(value).success;
}

export const LocalKickstartConnectAnswer = Object.freeze({
  Accept: "accept",
  Skip: "skip",
} as const);
export type LocalKickstartConnectAnswer = (typeof LocalKickstartConnectAnswer)[keyof typeof LocalKickstartConnectAnswer];
export const AllLocalKickstartConnectAnswers = Object.freeze([LocalKickstartConnectAnswer.Accept, LocalKickstartConnectAnswer.Skip]) as readonly LocalKickstartConnectAnswer[];
export function isLocalKickstartConnectAnswer(value: unknown): value is LocalKickstartConnectAnswer {
  return zLocalKickstartConnectAnswer.safeParse(value).success;
}

export const LocalKickstartErrorCode = Object.freeze({
  BodyMissing: "kickstart_body_missing",
  BodyInvalid: "kickstart_body_invalid",
  ScopeInvalid: "kickstart_scope_invalid",
  AnswersInvalid: "kickstart_answers_invalid",
  SessionCommitting: "kickstart_session_committing",
  RevisionAhead: "kickstart_revision_ahead",
  AnswerRefused: "kickstart_answer_refused",
  ScanNotDone: "kickstart_scan_not_done",
  FieldUnknown: "kickstart_field_unknown",
  FieldOutOfScope: "kickstart_field_out_of_scope",
  FieldHidden: "kickstart_field_hidden",
  TreeAnswerShape: "kickstart_tree_answer_shape",
  ConnectOutOfScope: "kickstart_connect_out_of_scope",
  ConnectNotConnected: "kickstart_connect_not_connected",
  ConnectUnknown: "kickstart_connect_unknown",
  CommitRefused: "kickstart_commit_refused",
  SessionNotFound: "kickstart_session_not_found",
  PreviewRowNotFound: "kickstart_preview_row_not_found",
  PreviewCaptureNotFound: "kickstart_preview_capture_not_found",
  PreviewSourceChanged: "kickstart_preview_source_changed",
  PreviewLateResult: "kickstart_preview_late_result",
  PreviewCapacity: "kickstart_preview_capacity",
  PreviewOutOfOrder: "kickstart_preview_out_of_order",
  Failed: "kickstart_failed",
  StatusUnreadable: "kickstart_status_unreadable",
  ConfigUnreadable: "kickstart_config_unreadable",
  ConfigUnwritable: "kickstart_config_unwritable",
  SessionUnbuildable: "kickstart_session_unbuildable",
  AnswerUnencodable: "kickstart_answer_unencodable",
  PreviewUnreadable: "kickstart_preview_unreadable",
  ContractBreach: "publishing_contract_breach",
  Unavailable: "kickstart_unavailable",
  ConfigBusy: "kickstart_config_busy",
  ConfigChanging: "kickstart_config_changing",
} as const);
export type LocalKickstartErrorCode = (typeof LocalKickstartErrorCode)[keyof typeof LocalKickstartErrorCode];
export const AllLocalKickstartErrorCodes = Object.freeze([LocalKickstartErrorCode.BodyMissing, LocalKickstartErrorCode.BodyInvalid, LocalKickstartErrorCode.ScopeInvalid, LocalKickstartErrorCode.AnswersInvalid, LocalKickstartErrorCode.SessionCommitting, LocalKickstartErrorCode.RevisionAhead, LocalKickstartErrorCode.AnswerRefused, LocalKickstartErrorCode.ScanNotDone, LocalKickstartErrorCode.FieldUnknown, LocalKickstartErrorCode.FieldOutOfScope, LocalKickstartErrorCode.FieldHidden, LocalKickstartErrorCode.TreeAnswerShape, LocalKickstartErrorCode.ConnectOutOfScope, LocalKickstartErrorCode.ConnectNotConnected, LocalKickstartErrorCode.ConnectUnknown, LocalKickstartErrorCode.CommitRefused, LocalKickstartErrorCode.SessionNotFound, LocalKickstartErrorCode.PreviewRowNotFound, LocalKickstartErrorCode.PreviewCaptureNotFound, LocalKickstartErrorCode.PreviewSourceChanged, LocalKickstartErrorCode.PreviewLateResult, LocalKickstartErrorCode.PreviewCapacity, LocalKickstartErrorCode.PreviewOutOfOrder, LocalKickstartErrorCode.Failed, LocalKickstartErrorCode.StatusUnreadable, LocalKickstartErrorCode.ConfigUnreadable, LocalKickstartErrorCode.ConfigUnwritable, LocalKickstartErrorCode.SessionUnbuildable, LocalKickstartErrorCode.AnswerUnencodable, LocalKickstartErrorCode.PreviewUnreadable, LocalKickstartErrorCode.ContractBreach, LocalKickstartErrorCode.Unavailable, LocalKickstartErrorCode.ConfigBusy, LocalKickstartErrorCode.ConfigChanging]) as readonly LocalKickstartErrorCode[];
export function isLocalKickstartErrorCode(value: unknown): value is LocalKickstartErrorCode {
  return zLocalKickstartErrorCode.safeParse(value).success;
}

export const AutoPublishEvent = Object.freeze({
  PrePush: "pre-push",
  PostCommit: "post-commit",
} as const);
export type AutoPublishEvent = (typeof AutoPublishEvent)[keyof typeof AutoPublishEvent];
export const AllAutoPublishEvents = Object.freeze([AutoPublishEvent.PrePush, AutoPublishEvent.PostCommit]) as readonly AutoPublishEvent[];
export function isAutoPublishEvent(value: unknown): value is AutoPublishEvent {
  return zAutoPublishEvent.safeParse(value).success;
}

export const AutoPublishHookStatus = Object.freeze({
  Absent: "absent",
  Installed: "installed",
  Blocked: "blocked",
  Failed: "failed",
} as const);
export type AutoPublishHookStatus = (typeof AutoPublishHookStatus)[keyof typeof AutoPublishHookStatus];
export const AllAutoPublishHookStatuses = Object.freeze([AutoPublishHookStatus.Absent, AutoPublishHookStatus.Installed, AutoPublishHookStatus.Blocked, AutoPublishHookStatus.Failed]) as readonly AutoPublishHookStatus[];
export function isAutoPublishHookStatus(value: unknown): value is AutoPublishHookStatus {
  return zAutoPublishHookStatus.safeParse(value).success;
}

export const AutoPublishRuleKind = Object.freeze({
  Folder: "folder",
  Remote: "remote",
} as const);
export type AutoPublishRuleKind = (typeof AutoPublishRuleKind)[keyof typeof AutoPublishRuleKind];
export const AllAutoPublishRuleKinds = Object.freeze([AutoPublishRuleKind.Folder, AutoPublishRuleKind.Remote]) as readonly AutoPublishRuleKind[];
export function isAutoPublishRuleKind(value: unknown): value is AutoPublishRuleKind {
  return zAutoPublishRuleKind.safeParse(value).success;
}

export const ThinkingLevel = Object.freeze({
  Off: "off",
  Minimal: "minimal",
  Low: "low",
  Medium: "medium",
  High: "high",
  XHigh: "xhigh",
  Max: "max",
  Ultra: "ultra",
  Ultracode: "ultracode",
} as const);
export type ThinkingLevel = (typeof ThinkingLevel)[keyof typeof ThinkingLevel];
export const AllThinkingLevels = Object.freeze([ThinkingLevel.Off, ThinkingLevel.Minimal, ThinkingLevel.Low, ThinkingLevel.Medium, ThinkingLevel.High, ThinkingLevel.XHigh, ThinkingLevel.Max, ThinkingLevel.Ultra, ThinkingLevel.Ultracode]) as readonly ThinkingLevel[];
export function isThinkingLevel(value: unknown): value is ThinkingLevel {
  return zThinkingLevel.safeParse(value).success;
}

export const ScryVerdict = Object.freeze({
  Checked: "checked",
  Unchecked: "unchecked",
} as const);
export type ScryVerdict = (typeof ScryVerdict)[keyof typeof ScryVerdict];
export const AllScryVerdicts = Object.freeze([ScryVerdict.Checked, ScryVerdict.Unchecked]) as readonly ScryVerdict[];
export function isScryVerdict(value: unknown): value is ScryVerdict {
  return zScryVerdict.safeParse(value).success;
}
