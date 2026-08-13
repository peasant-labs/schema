// Code generated from the canonical Go/OpenAPI contract. DO NOT EDIT.
import { zContentCapability, type ContentCapability as ContentCapabilityContract } from "./contract/zod.gen.js";
import { zContentCapabilityVersion, type ContentCapabilityVersion as ContentCapabilityVersionContract } from "./contract/zod.gen.js";
import { zPublishOperationKind, type PublishOperationKind as PublishOperationKindContract } from "./contract/zod.gen.js";
import { zVisibilityIntent, type VisibilityIntent as VisibilityIntentContract } from "./contract/zod.gen.js";
import { zAnnotationAxis, type AnnotationAxis as AnnotationAxisContract } from "./contract/zod.gen.js";
import { zAnnotationDatatype, type AnnotationDatatype as AnnotationDatatypeContract } from "./contract/zod.gen.js";
import { zAnnotationPushStatus, type AnnotationPushStatus as AnnotationPushStatusContract } from "./contract/zod.gen.js";
import { zAnnotationStatus, type AnnotationStatus as AnnotationStatusContract } from "./contract/zod.gen.js";
import { zAnnotatorKind, type AnnotatorKind as AnnotatorKindContract } from "./contract/zod.gen.js";
import { zAssociationConclusion, type AssociationConclusion as AssociationConclusionContract } from "./contract/zod.gen.js";
import { zAssociationEvidenceKind, type AssociationEvidenceKind as AssociationEvidenceKindContract } from "./contract/zod.gen.js";
import { zBuiltinCommand, type BuiltinCommand as BuiltinCommandContract } from "./contract/zod.gen.js";
import { zChangeBinding, type ChangeBinding as ChangeBindingContract } from "./contract/zod.gen.js";
import { zChannelTopic, type ChannelTopic as ChannelTopicContract } from "./contract/zod.gen.js";
import { zConfidence, type Confidence as ConfidenceContract } from "./contract/zod.gen.js";
import { zContentKind, type ContentKind as ContentKindContract } from "./contract/zod.gen.js";
import { zDecayLevel, type DecayLevel as DecayLevelContract } from "./contract/zod.gen.js";
import { zDiffLineKind, type DiffLineKind as DiffLineKindContract } from "./contract/zod.gen.js";
import { zEdgeViolationKind, type EdgeViolationKind as EdgeViolationKindContract } from "./contract/zod.gen.js";
import { zEntryType, type EntryType as EntryTypeContract } from "./contract/zod.gen.js";
import { zFileChangeStatus, type FileChangeStatus as FileChangeStatusContract } from "./contract/zod.gen.js";
import { zHarness, type Harness as HarnessContract } from "./contract/zod.gen.js";
import { zInsightKind, type InsightKind as InsightKindContract } from "./contract/zod.gen.js";
import { zInsightProvenance, type InsightProvenance as InsightProvenanceContract } from "./contract/zod.gen.js";
import { zInteractionType, type InteractionType as InteractionTypeContract } from "./contract/zod.gen.js";
import { zLicense, type License as LicenseContract } from "./contract/zod.gen.js";
import { zMapNodeKind, type MapNodeKind as MapNodeKindContract } from "./contract/zod.gen.js";
import { zMessageType, type MessageType as MessageTypeContract } from "./contract/zod.gen.js";
import { zReadAttributionState, type ReadAttributionState as ReadAttributionStateContract } from "./contract/zod.gen.js";
import { zReadStateGrade, type ReadStateGrade as ReadStateGradeContract } from "./contract/zod.gen.js";
import { zRewriteMethod, type RewriteMethod as RewriteMethodContract } from "./contract/zod.gen.js";
import { zRewriteResolution, type RewriteResolution as RewriteResolutionContract } from "./contract/zod.gen.js";
import { zRole, type Role as RoleContract } from "./contract/zod.gen.js";
import { zScaleKind, type ScaleKind as ScaleKindContract } from "./contract/zod.gen.js";
import { zSessionOutcome, type SessionOutcome as SessionOutcomeContract } from "./contract/zod.gen.js";
import { zSourceFormat, type SourceFormat as SourceFormatContract } from "./contract/zod.gen.js";
import { zStopReason, type StopReason as StopReasonContract } from "./contract/zod.gen.js";
import { zTargetKind, type TargetKind as TargetKindContract } from "./contract/zod.gen.js";
import { zToolCallKind, type ToolCallKind as ToolCallKindContract } from "./contract/zod.gen.js";
import { zTypeOrigin, type TypeOrigin as TypeOriginContract } from "./contract/zod.gen.js";
import { zValueDomainKind, type ValueDomainKind as ValueDomainKindContract } from "./contract/zod.gen.js";
import { zTranscriptUpdateLicense, type TranscriptUpdateLicense as TranscriptUpdateLicenseContract } from "./contract/zod.gen.js";
import { zTranscriptUpdateVisibility, type TranscriptUpdateVisibility as TranscriptUpdateVisibilityContract } from "./contract/zod.gen.js";
import { zVisibility, type Visibility as VisibilityContract } from "./contract/zod.gen.js";

export type ContentCapability = ContentCapabilityContract;
export const ContentCapability = Object.freeze({
  ObservedModel: zContentCapability.parse("observed_model"),
} as const);
export const AllContentCapabilities = Object.freeze([ContentCapability.ObservedModel]) as readonly ContentCapability[];
export function isContentCapability(value: unknown): value is ContentCapability {
  return zContentCapability.safeParse(value).success;
}

export type ContentCapabilityVersion = ContentCapabilityVersionContract;
export const ContentCapabilityVersion = Object.freeze({
  ObservedModelV1: zContentCapabilityVersion.parse("1.0.0"),
} as const);
export const AllContentCapabilityVersions = Object.freeze([ContentCapabilityVersion.ObservedModelV1]) as readonly ContentCapabilityVersion[];
export function isContentCapabilityVersion(value: unknown): value is ContentCapabilityVersion {
  return zContentCapabilityVersion.safeParse(value).success;
}

export type PublishOperationKind = PublishOperationKindContract;
export const PublishOperationKind = Object.freeze({
  Preserve: zPublishOperationKind.parse("preserve"),
  Replace: zPublishOperationKind.parse("replace"),
  Append: zPublishOperationKind.parse("append"),
} as const);
export const AllPublishOperationKinds = Object.freeze([PublishOperationKind.Preserve, PublishOperationKind.Replace, PublishOperationKind.Append]) as readonly PublishOperationKind[];
export function isPublishOperationKind(value: unknown): value is PublishOperationKind {
  return zPublishOperationKind.safeParse(value).success;
}

export type VisibilityIntent = VisibilityIntentContract;
export const VisibilityIntent = Object.freeze({
  Private: zVisibilityIntent.parse("private"),
  Public: zVisibilityIntent.parse("public"),
} as const);
export const AllVisibilityIntents = Object.freeze([VisibilityIntent.Private, VisibilityIntent.Public]) as readonly VisibilityIntent[];
export function isVisibilityIntent(value: unknown): value is VisibilityIntent {
  return zVisibilityIntent.safeParse(value).success;
}

export type AnnotationAxis = AnnotationAxisContract;
export const AnnotationAxis = Object.freeze({
  Type: zAnnotationAxis.parse("type"),
  Session: zAnnotationAxis.parse("session"),
  Project: zAnnotationAxis.parse("project"),
} as const);
export const AllAnnotationAxes = Object.freeze([AnnotationAxis.Type, AnnotationAxis.Session, AnnotationAxis.Project]) as readonly AnnotationAxis[];
export function isAnnotationAxis(value: unknown): value is AnnotationAxis {
  return zAnnotationAxis.safeParse(value).success;
}

export type AnnotationDatatype = AnnotationDatatypeContract;
export const AnnotationDatatype = Object.freeze({
  Text: zAnnotationDatatype.parse("text"),
  Integer: zAnnotationDatatype.parse("integer"),
  Real: zAnnotationDatatype.parse("real"),
  Boolean: zAnnotationDatatype.parse("boolean"),
} as const);
export const AllAnnotationDatatypes = Object.freeze([AnnotationDatatype.Text, AnnotationDatatype.Integer, AnnotationDatatype.Real, AnnotationDatatype.Boolean]) as readonly AnnotationDatatype[];
export function isAnnotationDatatype(value: unknown): value is AnnotationDatatype {
  return zAnnotationDatatype.safeParse(value).success;
}

export type AnnotationPushStatus = AnnotationPushStatusContract;
export const AnnotationPushStatus = Object.freeze({
  Created: zAnnotationPushStatus.parse("created"),
  Updated: zAnnotationPushStatus.parse("updated"),
  Skipped: zAnnotationPushStatus.parse("skipped"),
  Error: zAnnotationPushStatus.parse("error"),
} as const);
export const AllAnnotationPushStatuses = Object.freeze([AnnotationPushStatus.Created, AnnotationPushStatus.Updated, AnnotationPushStatus.Skipped, AnnotationPushStatus.Error]) as readonly AnnotationPushStatus[];
export function isAnnotationPushStatus(value: unknown): value is AnnotationPushStatus {
  return zAnnotationPushStatus.safeParse(value).success;
}

export type AnnotationStatus = AnnotationStatusContract;
export const AnnotationStatus = Object.freeze({
  Proposed: zAnnotationStatus.parse("proposed"),
  Active: zAnnotationStatus.parse("active"),
  Deprecated: zAnnotationStatus.parse("deprecated"),
  Retired: zAnnotationStatus.parse("retired"),
} as const);
export const AllAnnotationStatuses = Object.freeze([AnnotationStatus.Proposed, AnnotationStatus.Active, AnnotationStatus.Deprecated, AnnotationStatus.Retired]) as readonly AnnotationStatus[];
export function isAnnotationStatus(value: unknown): value is AnnotationStatus {
  return zAnnotationStatus.safeParse(value).success;
}

export type AnnotatorKind = AnnotatorKindContract;
export const AnnotatorKind = Object.freeze({
  Human: zAnnotatorKind.parse("human"),
  Agent: zAnnotatorKind.parse("agent"),
  Rule: zAnnotatorKind.parse("rule"),
} as const);
export const AllAnnotatorKinds = Object.freeze([AnnotatorKind.Human, AnnotatorKind.Agent, AnnotatorKind.Rule]) as readonly AnnotatorKind[];
export function isAnnotatorKind(value: unknown): value is AnnotatorKind {
  return zAnnotatorKind.safeParse(value).success;
}

export type AssociationConclusion = AssociationConclusionContract;
export const AssociationConclusion = Object.freeze({
  Confirmed: zAssociationConclusion.parse("confirmed"),
  Candidate: zAssociationConclusion.parse("candidate"),
} as const);
export const AllAssociationConclusions = Object.freeze([AssociationConclusion.Confirmed, AssociationConclusion.Candidate]) as readonly AssociationConclusion[];
export function isAssociationConclusion(value: unknown): value is AssociationConclusion {
  return zAssociationConclusion.safeParse(value).success;
}

export type AssociationEvidenceKind = AssociationEvidenceKindContract;
export const AssociationEvidenceKind = Object.freeze({
  RecordedCommit: zAssociationEvidenceKind.parse("recorded_commit"),
  TouchedFile: zAssociationEvidenceKind.parse("touched_file"),
  BranchMembership: zAssociationEvidenceKind.parse("branch_membership"),
  TimeWindow: zAssociationEvidenceKind.parse("time_window"),
} as const);
export const AllAssociationEvidenceKinds = Object.freeze([AssociationEvidenceKind.RecordedCommit, AssociationEvidenceKind.TouchedFile, AssociationEvidenceKind.BranchMembership, AssociationEvidenceKind.TimeWindow]) as readonly AssociationEvidenceKind[];
export function isAssociationEvidenceKind(value: unknown): value is AssociationEvidenceKind {
  return zAssociationEvidenceKind.safeParse(value).success;
}

export type BuiltinCommand = BuiltinCommandContract;
export const BuiltinCommand = Object.freeze({
  Exit: zBuiltinCommand.parse("exit"),
  Compact: zBuiltinCommand.parse("compact"),
  Clear: zBuiltinCommand.parse("clear"),
  New: zBuiltinCommand.parse("new"),
  Model: zBuiltinCommand.parse("model"),
  Usage: zBuiltinCommand.parse("usage"),
  Cost: zBuiltinCommand.parse("cost"),
  Context: zBuiltinCommand.parse("context"),
  Plugin: zBuiltinCommand.parse("plugin"),
  Permissions: zBuiltinCommand.parse("permissions"),
  Login: zBuiltinCommand.parse("login"),
  Resume: zBuiltinCommand.parse("resume"),
  Plan: zBuiltinCommand.parse("plan"),
  Fast: zBuiltinCommand.parse("fast"),
  Voice: zBuiltinCommand.parse("voice"),
  Todos: zBuiltinCommand.parse("todos"),
  ReloadPlugins: zBuiltinCommand.parse("reload-plugins"),
  Sandbox: zBuiltinCommand.parse("sandbox"),
  Config: zBuiltinCommand.parse("config"),
  Statusline: zBuiltinCommand.parse("statusline"),
  Upgrade: zBuiltinCommand.parse("upgrade"),
  ExtraUsage: zBuiltinCommand.parse("extra-usage"),
  RateLimitOptions: zBuiltinCommand.parse("rate-limit-options"),
  PrivacySettings: zBuiltinCommand.parse("privacy-settings"),
  Help: zBuiltinCommand.parse("help"),
  Commands: zBuiltinCommand.parse("commands"),
} as const);
export const AllClaudeBuiltinCmds = Object.freeze([BuiltinCommand.Exit, BuiltinCommand.Compact, BuiltinCommand.Clear, BuiltinCommand.New, BuiltinCommand.Model, BuiltinCommand.Usage, BuiltinCommand.Cost, BuiltinCommand.Context, BuiltinCommand.Plugin, BuiltinCommand.Permissions, BuiltinCommand.Login, BuiltinCommand.Resume, BuiltinCommand.Plan, BuiltinCommand.Fast, BuiltinCommand.Voice, BuiltinCommand.Todos, BuiltinCommand.ReloadPlugins, BuiltinCommand.Sandbox, BuiltinCommand.Config, BuiltinCommand.Statusline, BuiltinCommand.Upgrade, BuiltinCommand.ExtraUsage, BuiltinCommand.RateLimitOptions, BuiltinCommand.PrivacySettings, BuiltinCommand.Help, BuiltinCommand.Commands]) as readonly BuiltinCommand[];
export function isBuiltinCommand(value: unknown): value is BuiltinCommand {
  return zBuiltinCommand.safeParse(value).success;
}

export type ChangeBinding = ChangeBindingContract;
export const ChangeBinding = Object.freeze({
  Bound: zChangeBinding.parse("bound"),
  Candidate: zChangeBinding.parse("candidate"),
} as const);
export const AllChangeBindings = Object.freeze([ChangeBinding.Bound, ChangeBinding.Candidate]) as readonly ChangeBinding[];
export function isChangeBinding(value: unknown): value is ChangeBinding {
  return zChangeBinding.safeParse(value).success;
}

export type ChannelTopic = ChannelTopicContract;
export const ChannelTopic = Object.freeze({
  Dashboard: zChannelTopic.parse("dashboard"),
  Sessions: zChannelTopic.parse("sessions"),
  SessionDetail: zChannelTopic.parse("session_detail"),
  Trends: zChannelTopic.parse("trends"),
  Quality: zChannelTopic.parse("quality"),
  Annotations: zChannelTopic.parse("annotations"),
  ProjectFamiliarity: zChannelTopic.parse("project_familiarity"),
} as const);
export const AllChannelTopics = Object.freeze([ChannelTopic.Dashboard, ChannelTopic.Sessions, ChannelTopic.SessionDetail, ChannelTopic.Trends, ChannelTopic.Quality, ChannelTopic.Annotations, ChannelTopic.ProjectFamiliarity]) as readonly ChannelTopic[];
export function isChannelTopic(value: unknown): value is ChannelTopic {
  return zChannelTopic.safeParse(value).success;
}

export type Confidence = ConfidenceContract;
export const Confidence = Object.freeze({
  High: zConfidence.parse("high"),
  Medium: zConfidence.parse("medium"),
  Low: zConfidence.parse("low"),
} as const);
export const AllConfidences = Object.freeze([Confidence.High, Confidence.Medium, Confidence.Low]) as readonly Confidence[];
export function isConfidence(value: unknown): value is Confidence {
  return zConfidence.safeParse(value).success;
}

export type ContentKind = ContentKindContract;
export const ContentKind = Object.freeze({
  SessionDetail: zContentKind.parse("session_detail"),
} as const);
export const AllContentKinds = Object.freeze([ContentKind.SessionDetail]) as readonly ContentKind[];
export function isContentKind(value: unknown): value is ContentKind {
  return zContentKind.safeParse(value).success;
}

export type DecayLevel = DecayLevelContract;
export const DecayLevel = Object.freeze({
  Fresh: zDecayLevel.parse("fresh"),
  Fading: zDecayLevel.parse("fading"),
  Stale: zDecayLevel.parse("stale"),
  Unexplored: zDecayLevel.parse("unexplored"),
} as const);
export const AllDecayLevels = Object.freeze([DecayLevel.Fresh, DecayLevel.Fading, DecayLevel.Stale, DecayLevel.Unexplored]) as readonly DecayLevel[];
export function isDecayLevel(value: unknown): value is DecayLevel {
  return zDecayLevel.safeParse(value).success;
}

export type DiffLineKind = DiffLineKindContract;
export const DiffLineKind = Object.freeze({
  Context: zDiffLineKind.parse("context"),
  Add: zDiffLineKind.parse("add"),
  Delete: zDiffLineKind.parse("del"),
} as const);
export const AllDiffLineKinds = Object.freeze([DiffLineKind.Context, DiffLineKind.Add, DiffLineKind.Delete]) as readonly DiffLineKind[];
export function isDiffLineKind(value: unknown): value is DiffLineKind {
  return zDiffLineKind.safeParse(value).success;
}

export type EdgeViolationKind = EdgeViolationKindContract;
export const EdgeViolationKind = Object.freeze({
  Cycle: zEdgeViolationKind.parse("cycle"),
  WrongWay: zEdgeViolationKind.parse("wrong_way"),
} as const);
export const AllEdgeViolationKinds = Object.freeze([EdgeViolationKind.Cycle, EdgeViolationKind.WrongWay]) as readonly EdgeViolationKind[];
export function isEdgeViolationKind(value: unknown): value is EdgeViolationKind {
  return zEdgeViolationKind.safeParse(value).success;
}

export type EntryType = EntryTypeContract;
export const EntryType = Object.freeze({
  Text: zEntryType.parse("text"),
  ToolUse: zEntryType.parse("tool_use"),
  ToolResult: zEntryType.parse("tool_result"),
  Thinking: zEntryType.parse("thinking"),
  System: zEntryType.parse("system"),
  Error: zEntryType.parse("error"),
  Result: zEntryType.parse("result"),
} as const);
export const AllEntryTypes = Object.freeze([EntryType.Text, EntryType.ToolUse, EntryType.ToolResult, EntryType.Thinking, EntryType.System, EntryType.Error, EntryType.Result]) as readonly EntryType[];
export function isEntryType(value: unknown): value is EntryType {
  return zEntryType.safeParse(value).success;
}

export type FileChangeStatus = FileChangeStatusContract;
export const FileChangeStatus = Object.freeze({
  Modified: zFileChangeStatus.parse("M"),
  Added: zFileChangeStatus.parse("A"),
  Deleted: zFileChangeStatus.parse("D"),
  Renamed: zFileChangeStatus.parse("R"),
} as const);
export const AllFileChangeStatuses = Object.freeze([FileChangeStatus.Modified, FileChangeStatus.Added, FileChangeStatus.Deleted, FileChangeStatus.Renamed]) as readonly FileChangeStatus[];
export function isFileChangeStatus(value: unknown): value is FileChangeStatus {
  return zFileChangeStatus.safeParse(value).success;
}

export type Harness = HarnessContract;
export const Harness = Object.freeze({
  ClaudeCode: zHarness.parse("claude-code"),
  GeminiCLI: zHarness.parse("gemini-cli"),
  Codex: zHarness.parse("codex"),
  OpenCode: zHarness.parse("opencode"),
  Cursor: zHarness.parse("cursor"),
  Antigravity: zHarness.parse("antigravity"),
  Strike: zHarness.parse("strike"),
} as const);
/**
 * AllHarnesses is the ingestion-supported subset of Harness, not the full
 * canonical set (mirrors types.go's AllHarnesses doc comment). Every Harness
 * member remains individually valid and accepted by isHarness.
 */
export const AllHarnesses = Object.freeze([Harness.ClaudeCode, Harness.GeminiCLI, Harness.Codex, Harness.OpenCode, Harness.Cursor, Harness.Strike]) as readonly Harness[];
export function isHarness(value: unknown): value is Harness {
  return zHarness.safeParse(value).success;
}

export type InsightKind = InsightKindContract;
export const InsightKind = Object.freeze({
  Decision: zInsightKind.parse("decision"),
  Friction: zInsightKind.parse("friction"),
  Unusual: zInsightKind.parse("unusual"),
  RetryLoop: zInsightKind.parse("retry_loop"),
} as const);
export const AllInsightKinds = Object.freeze([InsightKind.Decision, InsightKind.Friction, InsightKind.Unusual, InsightKind.RetryLoop]) as readonly InsightKind[];
export function isInsightKind(value: unknown): value is InsightKind {
  return zInsightKind.safeParse(value).success;
}

export type InsightProvenance = InsightProvenanceContract;
export const InsightProvenance = Object.freeze({
  Mechanical: zInsightProvenance.parse("mechanical"),
  Mined: zInsightProvenance.parse("mined"),
} as const);
export const AllInsightProvenances = Object.freeze([InsightProvenance.Mechanical, InsightProvenance.Mined]) as readonly InsightProvenance[];
export function isInsightProvenance(value: unknown): value is InsightProvenance {
  return zInsightProvenance.safeParse(value).success;
}

export type InteractionType = InteractionTypeContract;
export const InteractionType = Object.freeze({
  Mentioned: zInteractionType.parse("mentioned"),
  Read: zInteractionType.parse("read"),
  Discussed: zInteractionType.parse("discussed"),
  Questioned: zInteractionType.parse("questioned"),
} as const);
export const AllInteractionTypes = Object.freeze([InteractionType.Mentioned, InteractionType.Read, InteractionType.Discussed, InteractionType.Questioned]) as readonly InteractionType[];
export function isInteractionType(value: unknown): value is InteractionType {
  return zInteractionType.safeParse(value).success;
}

export type License = LicenseContract;
export const License = Object.freeze({
  CC0: zLicense.parse("CC0-1.0"),
  CCBY: zLicense.parse("CC-BY-4.0"),
  CCBYSA: zLicense.parse("CC-BY-SA-4.0"),
} as const);
export const AllLicenses = Object.freeze([License.CC0, License.CCBY, License.CCBYSA]) as readonly License[];
export function isLicense(value: unknown): value is License {
  return zLicense.safeParse(value).success;
}

export type MapNodeKind = MapNodeKindContract;
export const MapNodeKind = Object.freeze({
  Module: zMapNodeKind.parse("module"),
  Package: zMapNodeKind.parse("package"),
  File: zMapNodeKind.parse("file"),
} as const);
export const AllMapNodeKinds = Object.freeze([MapNodeKind.Module, MapNodeKind.Package, MapNodeKind.File]) as readonly MapNodeKind[];
export function isMapNodeKind(value: unknown): value is MapNodeKind {
  return zMapNodeKind.safeParse(value).success;
}

export type MessageType = MessageTypeContract;
export const MessageType = Object.freeze({
  Subscribe: zMessageType.parse("subscribe"),
  Unsubscribe: zMessageType.parse("unsubscribe"),
  Dashboard: zMessageType.parse("dashboard"),
  Sessions: zMessageType.parse("sessions"),
  SessionDetail: zMessageType.parse("session_detail"),
  Trends: zMessageType.parse("trends"),
  Quality: zMessageType.parse("quality"),
  Annotations: zMessageType.parse("annotations"),
  ProjectFamiliarity: zMessageType.parse("project_familiarity"),
  Connected: zMessageType.parse("connected"),
  Error: zMessageType.parse("error"),
} as const);
export const AllMessageTypes = Object.freeze([MessageType.Subscribe, MessageType.Unsubscribe, MessageType.Dashboard, MessageType.Sessions, MessageType.SessionDetail, MessageType.Trends, MessageType.Quality, MessageType.Annotations, MessageType.ProjectFamiliarity, MessageType.Connected, MessageType.Error]) as readonly MessageType[];
export function isMessageType(value: unknown): value is MessageType {
  return zMessageType.safeParse(value).success;
}

export type ReadAttributionState = ReadAttributionStateContract;
export const ReadAttributionState = Object.freeze({
  Complete: zReadAttributionState.parse("complete"),
  Partial: zReadAttributionState.parse("partial"),
  Unavailable: zReadAttributionState.parse("unavailable"),
} as const);
export const AllReadAttributionStates = Object.freeze([ReadAttributionState.Complete, ReadAttributionState.Partial, ReadAttributionState.Unavailable]) as readonly ReadAttributionState[];
export function isReadAttributionState(value: unknown): value is ReadAttributionState {
  return zReadAttributionState.safeParse(value).success;
}

export type ReadStateGrade = ReadStateGradeContract;
export const ReadStateGrade = Object.freeze({
  None: zReadStateGrade.parse("none"),
  Viewed: zReadStateGrade.parse("viewed"),
  Reviewed: zReadStateGrade.parse("reviewed"),
  ReviewedInDetail: zReadStateGrade.parse("reviewed_in_detail"),
} as const);
export const AllReadStateGrades = Object.freeze([ReadStateGrade.None, ReadStateGrade.Viewed, ReadStateGrade.Reviewed, ReadStateGrade.ReviewedInDetail]) as readonly ReadStateGrade[];
export function isReadStateGrade(value: unknown): value is ReadStateGrade {
  return zReadStateGrade.safeParse(value).success;
}

export type RewriteMethod = RewriteMethodContract;
export const RewriteMethod = Object.freeze({
  Hash: zRewriteMethod.parse("hash"),
  PatchID: zRewriteMethod.parse("patch_id"),
  AuthorIdentity: zRewriteMethod.parse("author_identity"),
  MessageEmbedded: zRewriteMethod.parse("message_embedded"),
  Temporal: zRewriteMethod.parse("temporal"),
  None: zRewriteMethod.parse("none"),
} as const);
export const AllRewriteMethods = Object.freeze([RewriteMethod.Hash, RewriteMethod.PatchID, RewriteMethod.AuthorIdentity, RewriteMethod.MessageEmbedded, RewriteMethod.Temporal, RewriteMethod.None]) as readonly RewriteMethod[];
export function isRewriteMethod(value: unknown): value is RewriteMethod {
  return zRewriteMethod.safeParse(value).success;
}

export type RewriteResolution = RewriteResolutionContract;
export const RewriteResolution = Object.freeze({
  Live: zRewriteResolution.parse("live"),
  Rewritten: zRewriteResolution.parse("rewritten"),
  Unresolved: zRewriteResolution.parse("unresolved"),
} as const);
export const AllRewriteResolutions = Object.freeze([RewriteResolution.Live, RewriteResolution.Rewritten, RewriteResolution.Unresolved]) as readonly RewriteResolution[];
export function isRewriteResolution(value: unknown): value is RewriteResolution {
  return zRewriteResolution.safeParse(value).success;
}

export type Role = RoleContract;
export const Role = Object.freeze({
  User: zRole.parse("user"),
  Assistant: zRole.parse("assistant"),
  Tool: zRole.parse("tool"),
  System: zRole.parse("system"),
} as const);
export const AllRoles = Object.freeze([Role.User, Role.Assistant, Role.Tool, Role.System]) as readonly Role[];
export function isRole(value: unknown): value is Role {
  return zRole.safeParse(value).success;
}

export type ScaleKind = ScaleKindContract;
export const ScaleKind = Object.freeze({
  Nominal: zScaleKind.parse("nominal"),
  Ordinal: zScaleKind.parse("ordinal"),
  Continuous: zScaleKind.parse("continuous"),
} as const);
export const AllScaleKinds = Object.freeze([ScaleKind.Nominal, ScaleKind.Ordinal, ScaleKind.Continuous]) as readonly ScaleKind[];
export function isScaleKind(value: unknown): value is ScaleKind {
  return zScaleKind.safeParse(value).success;
}

export type SessionOutcome = SessionOutcomeContract;
export const SessionOutcome = Object.freeze({
  Resolved: zSessionOutcome.parse("resolved"),
  Partial: zSessionOutcome.parse("partial"),
  Failed: zSessionOutcome.parse("failed"),
} as const);
export const AllOutcomes = Object.freeze([SessionOutcome.Resolved, SessionOutcome.Partial, SessionOutcome.Failed]) as readonly SessionOutcome[];
export function isSessionOutcome(value: unknown): value is SessionOutcome {
  return zSessionOutcome.safeParse(value).success;
}

export type SourceFormat = SourceFormatContract;
export const SourceFormat = Object.freeze({
  JSONL: zSourceFormat.parse("jsonl"),
  JSON: zSourceFormat.parse("json"),
} as const);
export const AllSourceFormats = Object.freeze([SourceFormat.JSONL, SourceFormat.JSON]) as readonly SourceFormat[];
export function isSourceFormat(value: unknown): value is SourceFormat {
  return zSourceFormat.safeParse(value).success;
}

export type StopReason = StopReasonContract;
export const StopReason = Object.freeze({
  EndTurn: zStopReason.parse("end_turn"),
  Cancelled: zStopReason.parse("cancelled"),
  MaxTokens: zStopReason.parse("max_tokens"),
  MaxTurnRequests: zStopReason.parse("max_turn_requests"),
  Refusal: zStopReason.parse("refusal"),
} as const);
export const AllStopReasons = Object.freeze([StopReason.EndTurn, StopReason.Cancelled, StopReason.MaxTokens, StopReason.MaxTurnRequests, StopReason.Refusal]) as readonly StopReason[];
export function isStopReason(value: unknown): value is StopReason {
  return zStopReason.safeParse(value).success;
}

export type TargetKind = TargetKindContract;
export const TargetKind = Object.freeze({
  Session: zTargetKind.parse("session"),
  Entry: zTargetKind.parse("entry"),
  Annotation: zTargetKind.parse("annotation"),
  Project: zTargetKind.parse("project"),
  FileVersion: zTargetKind.parse("file_version"),
  Association: zTargetKind.parse("association"),
} as const);
export const AllTargetKinds = Object.freeze([TargetKind.Session, TargetKind.Entry, TargetKind.Annotation, TargetKind.Project, TargetKind.FileVersion, TargetKind.Association]) as readonly TargetKind[];
export function isTargetKind(value: unknown): value is TargetKind {
  return zTargetKind.safeParse(value).success;
}

export type ToolCallKind = ToolCallKindContract;
export const ToolCallKind = Object.freeze({
  Read: zToolCallKind.parse("read"),
  Edit: zToolCallKind.parse("edit"),
  Delete: zToolCallKind.parse("delete"),
  Move: zToolCallKind.parse("move"),
  Search: zToolCallKind.parse("search"),
  Execute: zToolCallKind.parse("execute"),
  Think: zToolCallKind.parse("think"),
  Fetch: zToolCallKind.parse("fetch"),
  Other: zToolCallKind.parse("other"),
} as const);
export const AllToolCallKinds = Object.freeze([ToolCallKind.Read, ToolCallKind.Edit, ToolCallKind.Delete, ToolCallKind.Move, ToolCallKind.Search, ToolCallKind.Execute, ToolCallKind.Think, ToolCallKind.Fetch, ToolCallKind.Other]) as readonly ToolCallKind[];
export function isToolCallKind(value: unknown): value is ToolCallKind {
  return zToolCallKind.safeParse(value).success;
}

export type TypeOrigin = TypeOriginContract;
export const TypeOrigin = Object.freeze({
  System: zTypeOrigin.parse("system"),
  User: zTypeOrigin.parse("user"),
  Group: zTypeOrigin.parse("group"),
} as const);
export const AllTypeOrigins = Object.freeze([TypeOrigin.System, TypeOrigin.User, TypeOrigin.Group]) as readonly TypeOrigin[];
export function isTypeOrigin(value: unknown): value is TypeOrigin {
  return zTypeOrigin.safeParse(value).success;
}

export type ValueDomainKind = ValueDomainKindContract;
export const ValueDomainKind = Object.freeze({
  Enumerated: zValueDomainKind.parse("enumerated"),
  Described: zValueDomainKind.parse("described"),
} as const);
export const AllValueDomainKinds = Object.freeze([ValueDomainKind.Enumerated, ValueDomainKind.Described]) as readonly ValueDomainKind[];
export function isValueDomainKind(value: unknown): value is ValueDomainKind {
  return zValueDomainKind.safeParse(value).success;
}

export type TranscriptUpdateLicense = TranscriptUpdateLicenseContract;
export const TranscriptUpdateLicense = Object.freeze({
  Clear: zTranscriptUpdateLicense.parse(""),
  CC0: zTranscriptUpdateLicense.parse("CC0-1.0"),
  CCBY: zTranscriptUpdateLicense.parse("CC-BY-4.0"),
  CCBYSA: zTranscriptUpdateLicense.parse("CC-BY-SA-4.0"),
} as const);
export const AllTranscriptUpdateLicenses = Object.freeze([TranscriptUpdateLicense.Clear, TranscriptUpdateLicense.CC0, TranscriptUpdateLicense.CCBY, TranscriptUpdateLicense.CCBYSA]) as readonly TranscriptUpdateLicense[];
export function isTranscriptUpdateLicense(value: unknown): value is TranscriptUpdateLicense {
  return zTranscriptUpdateLicense.safeParse(value).success;
}

export type TranscriptUpdateVisibility = TranscriptUpdateVisibilityContract;
export const TranscriptUpdateVisibility = Object.freeze({
  Private: zTranscriptUpdateVisibility.parse("private"),
  Public: zTranscriptUpdateVisibility.parse("public"),
} as const);
export const AllTranscriptUpdateVisibilities = Object.freeze([TranscriptUpdateVisibility.Private, TranscriptUpdateVisibility.Public]) as readonly TranscriptUpdateVisibility[];
export function isTranscriptUpdateVisibility(value: unknown): value is TranscriptUpdateVisibility {
  return zTranscriptUpdateVisibility.safeParse(value).success;
}

export type Visibility = VisibilityContract;
export const Visibility = Object.freeze({
  Private: zVisibility.parse("private"),
  Group: zVisibility.parse("group"),
  Public: zVisibility.parse("public"),
} as const);
export const AllVisibilities = Object.freeze([Visibility.Private, Visibility.Group, Visibility.Public]) as readonly Visibility[];
export function isVisibility(value: unknown): value is Visibility {
  return zVisibility.safeParse(value).success;
}
