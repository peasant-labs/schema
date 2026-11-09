import type {
  SessionDetailHeader,
  SessionDetailHeaderPayload,
  SessionDetailReadPayload,
  TranscriptSegment,
  TranscriptSegmentDirectory,
  TranscriptSegmentRef,
  TranscriptTurn,
  TranscriptWindow,
  TurnDetail,
} from "@peasant-labs/schema";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends <T>() => T extends B ? 1 : 2 ? true : false;
type HasKey<T, K extends PropertyKey> = K extends keyof T ? true : false;

// The window is a descriptor: it names where the window is and carries no turn
// content. A viewer loads the covering segments through the segment read.
const windowHasNoTurns: HasKey<TranscriptWindow, "turns"> = false;
const windowHasNoContent: HasKey<TranscriptWindow, "content"> = false;
const windowHasNoToolCalls: HasKey<TranscriptWindow, "toolCalls"> = false;
const windowHasNoPosition: HasKey<TranscriptWindow, "position"> = false;
const windowHasNoPage: HasKey<TranscriptWindow, "page"> = false;
const windowHasNoPerPage: HasKey<TranscriptWindow, "perPage"> = false;
const windowSegmentsAreOrdinals: Same<TranscriptWindow["segments"], number[]> = true;

// The directory is the only global index: one ref per segment plus the cut's
// scalars. It carries no per-turn row and no page coordinate for the transcript.
const directoryHasNoTurns: HasKey<TranscriptSegmentDirectory, "turns"> = false;
const directoryHasNoPosition: HasKey<TranscriptSegmentDirectory, "position"> = false;
const directoryHasNoPage: HasKey<TranscriptSegmentDirectory, "page"> = false;
const directoryRefsAreRows: Same<TranscriptSegmentDirectory["segments"][number], TranscriptSegmentRef> = true;
const directoryRefHasNoPosition: HasKey<TranscriptSegmentRef, "position"> = false;
const directoryPageIsOptional: Same<TranscriptSegmentDirectory["directoryPage"], number | undefined> = true;

// One turn row: the single TurnDetail definition plus the read's own facts,
// never a flat copy and never a second projection.
type TranscriptTurnKeys = keyof TranscriptTurn;
type ExpectedTurnKeys = keyof TurnDetail | "truncated" | "byteLength";
const turnRowIsTurnDetailPlusFacts: Same<TranscriptTurnKeys, ExpectedTurnKeys> = true;
const turnIndexHasOneDefinition: Same<TranscriptTurn["index"], TurnDetail["index"]> = true;
const turnContentHasOneDefinition: Same<TranscriptTurn["content"], TurnDetail["content"]> = true;
const segmentTurnsAreRows: Same<TranscriptSegment["turns"][number], TranscriptTurn> = true;
const segmentHasNoPosition: HasKey<TranscriptSegment, "position"> = false;

// The header is a metadata-only projection and a distinct root type. The
// durable read keeps its turns and earlier-history sections.
const headerHasNoTurns: HasKey<SessionDetailHeader, "turns"> = false;
const headerHasNoNativeMetadata: HasKey<SessionDetailHeader, "nativeMetadata"> = false;
const headerPayloadHasNoTurns: HasKey<SessionDetailHeaderPayload, "turns"> = false;
const durableReadHasTurns: HasKey<SessionDetailReadPayload, "turns"> = true;
const durableReadHasEarlierHistory: HasKey<SessionDetailReadPayload, "earlierHistory"> = true;
const headerIsNotTheDurableRead: Same<SessionDetailHeader, SessionDetailReadPayload> = false;

declare const window: TranscriptWindow;
declare const directory: TranscriptSegmentDirectory;
declare const segment: TranscriptSegment;
declare const header: SessionDetailHeader;
declare const headerPayload: SessionDetailHeaderPayload;

void [
  window,
  directory,
  segment,
  header,
  headerPayload,
  windowHasNoTurns,
  windowHasNoContent,
  windowHasNoToolCalls,
  windowHasNoPosition,
  windowHasNoPage,
  windowHasNoPerPage,
  windowSegmentsAreOrdinals,
  directoryHasNoTurns,
  directoryHasNoPosition,
  directoryHasNoPage,
  directoryRefsAreRows,
  directoryRefHasNoPosition,
  directoryPageIsOptional,
  turnRowIsTurnDetailPlusFacts,
  turnIndexHasOneDefinition,
  turnContentHasOneDefinition,
  segmentTurnsAreRows,
  segmentHasNoPosition,
  headerHasNoTurns,
  headerHasNoNativeMetadata,
  headerPayloadHasNoTurns,
  durableReadHasTurns,
  durableReadHasEarlierHistory,
  headerIsNotTheDurableRead,
];
