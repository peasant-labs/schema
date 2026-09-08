import type { SessionDetailPayload, ToolCallDetail, TranscriptContent } from "../src/index.js";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
type Turn = NonNullable<SessionDetailPayload["turns"]>[number];
type Tool = NonNullable<Turn["toolCalls"]>[number];
type ContentDetail = NonNullable<TranscriptContent["sessionDetail"]>;

// The public roots share one optional, non-null string, not a qualified name.
type Namespace = Assert<Same<ToolCallDetail["namespace"], string | undefined>>;
type NestedNamespace = Assert<Same<Tool["namespace"], ToolCallDetail["namespace"]>>;
type Envelope = Assert<Same<ContentDetail, SessionDetailPayload>>;

export type NamespaceContract = Namespace | NestedNamespace | Envelope;
