export * from "./internal/generated/public-contract.gen.js";
export * from "./internal/generated/enums.gen.js";
export * from "./internal/generated/content-capabilities.gen.js";
export * from "./internal/generated/versions.gen.js";

import { zNativeMetadataRecord, zTurnDetail, zProjectHash, zServerMessage, zSessionDetailPayload, zTranscriptContent, type ProjectHash } from "./internal/generated/contract/zod.gen.js";

function assertProjectHash(value: unknown, operation: "newProjectHash" | "validateProjectHash"): asserts value is ProjectHash {
  if (!isProjectHash(value)) {
    const rendered = renderProjectHashInput(value);
    throw new TypeError("ProjectHash validation failed for " + rendered + " at @peasant-labs/schema ProjectHash during " + operation + ": the value is not a 64-character lowercase hexadecimal string; callers cannot use it as a canonical project identity; pass the lowercase SHA-256 hex digest of the project origin URL or local path.");
  }
}

function renderProjectHashInput(value: unknown): string {
  try {
    const json = JSON.stringify(value);
    if (json !== undefined) return json;
  } catch {
    // Fall through to String for values whose JSON hooks throw.
  }
  try {
    return String(value);
  } catch {
    return "<unrenderable value>";
  }
}

export function isProjectHash(value: unknown): value is ProjectHash {
  return zProjectHash.safeParse(value).success;
}

export function validateProjectHash(value: unknown): asserts value is ProjectHash {
  assertProjectHash(value, "validateProjectHash");
}

export function newProjectHash(raw: string): ProjectHash {
  assertProjectHash(raw, "newProjectHash");
  return raw;
}

export type PushContractVersion = import("./internal/generated/contract/zod.gen.js").ContractVersion;

export interface RawJsonPathPolicy {
  maxDocumentBytes?: number;
  maxDocumentDepth?: number;
  opaqueMetadataPointers?: readonly string[];
}

export function scanRawJsonText(text: string, policy: RawJsonPathPolicy = {}): void {
  scanRawJson(text, policy);
}

function scanRawJson(text: string, policy: RawJsonPathPolicy): RawJsonScanner {
  if (policy.maxDocumentBytes !== undefined && new TextEncoder().encode(text).length > policy.maxDocumentBytes) {
    throw new TypeError("Raw JSON validation failed at @peasant-labs/schema scanRawJsonText during pre-decode scanning: the document exceeds the configured byte limit; parsing it would cross the caller's safety boundary; reduce the document before retrying.");
  }
  const scanner = new RawJsonScanner(text, policy);
  scanner.scan();
  return scanner;
}

export function parseSessionDetailPayloadValue(value: unknown): import("./internal/generated/contract/zod.gen.js").SessionDetailPayload {
  rejectExplicitNullEvidence(value);
  validateRawGraphValue(value);
  const payload = zSessionDetailPayload.parse(value);
  validateSessionDetail(payload);
  return payload;
}

export function parseSessionDetailPayloadText(text: string): import("./internal/generated/contract/zod.gen.js").SessionDetailPayload {
  scanRawJsonText(text, {maxDocumentBytes: 8 << 20, maxDocumentDepth: 64, opaqueMetadataPointers: ["/nativeMetadata/*/data"]});
  return parseSessionDetailPayloadValue(JSON.parse(text));
}

export function parseSchemaVersionAdvertisement(text: string): string[] {
  scanRawJsonText(text, {maxDocumentBytes: 1 << 20, maxDocumentDepth: 16});
  const value: unknown = JSON.parse(text);
  if (!isRecord(value)) throw new TypeError("Schema capability validation failed at @peasant-labs/schema parseSchemaVersionAdvertisement during discovery decoding: response is not an object; callers cannot negotiate preservation support; send a JSON object and retry.");
  if (value.contentCapabilities === null) throw new TypeError("Schema capability validation failed at @peasant-labs/schema parseSchemaVersionAdvertisement during discovery decoding: contentCapabilities is null; callers cannot distinguish omission from malformed support; omit it or send an array of strings.");
  if (value.contentCapabilities === undefined) return [];
  if (!Array.isArray(value.contentCapabilities) || value.contentCapabilities.some((token) => typeof token !== "string")) throw new TypeError("Schema capability validation failed at @peasant-labs/schema parseSchemaVersionAdvertisement during discovery decoding: contentCapabilities is not an array of strings; callers cannot negotiate support; send opaque string tokens and retry.");
  return [...value.contentCapabilities];
}

export function parseTranscriptContentText(text: string): import("./internal/generated/contract/zod.gen.js").TranscriptContent {
  scanRawJsonText(text, { maxDocumentBytes: 8 << 20, maxDocumentDepth: 64, opaqueMetadataPointers: ["/sessionDetail/nativeMetadata/*/data"] });
  const raw: unknown = JSON.parse(text);
  if (isRecord(raw)) rejectExplicitNullEvidence(raw.sessionDetail);
  const content = zTranscriptContent.parse(raw);
  if (content.sessionDetail === undefined || content.sessionDetail === null) failSemantic("sessionDetail is required");
  parseSessionDetailPayloadValue(content.sessionDetail);
  return content;
}

export function parseServerMessageRaw(text: string): import("./internal/generated/contract/zod.gen.js").ServerMessage {
  // Dispatch from the syntax scan, not a lossy full-document JSON.parse. The
  // type member may follow data, so a second selected scan precedes decoding.
  const dispatch = scanRawJson(text, {maxDocumentBytes: 8 << 20, maxDocumentDepth: 64});
  if (dispatch.rootMessageType === "session_detail") {
    scanRawJsonText(text, { maxDocumentBytes: 8 << 20, maxDocumentDepth: 64, opaqueMetadataPointers: ["/data/nativeMetadata/*/data"] });
  }
  const value: unknown = JSON.parse(text);
  if (isRecord(value) && value.type === "session_detail") rejectExplicitNullEvidence(value.data);
  const message = zServerMessage.parse(value);
  if (message.type === "session_detail") {
    parseSessionDetailPayloadValue(message.data);
  }
  return message;
}

class RawJsonScanner {
  private position = 0;
  private metadataBytes = 0;
  rootMessageType: string | undefined;
  constructor(private readonly text: string, private readonly policy: RawJsonPathPolicy) {}
  scan(): void { this.space(); this.value("", 1, 0); this.space(); if (this.position !== this.text.length) this.fail("trailing content"); }
  private value(path: string, depth: number, metadataDepth: number): void {
    if (metadataDepth === 0 && (this.policy.opaqueMetadataPointers ?? []).some(pattern => pointerMatches(pattern, path))) metadataDepth = 1;
    if (this.policy.maxDocumentDepth !== undefined && depth > this.policy.maxDocumentDepth) this.fail("document depth exceeds configured limit");
    if (metadataDepth > 32) this.fail("metadata depth exceeds 32");
    this.space();
    const start = this.position;
    this.valueBody(path, depth, metadataDepth);
    if (metadataDepth === 1) {
      const bytes = new TextEncoder().encode(this.text.slice(start, this.position)).length;
      this.metadataBytes += bytes;
      if (bytes > 65536 || this.metadataBytes > 1048576) this.fail("metadata data exceeds its byte budget (64 KiB per value, 1 MiB aggregate)");
    }
  }
  private valueBody(path: string, depth: number, metadataDepth: number): void {
    const opaque = metadataDepth > 0;
    this.space(); const c = this.text[this.position];
    if (c === "{") return this.object(path, depth, metadataDepth);
    if (c === "[") return this.array(path, depth, metadataDepth);
    if (c === '"') { const value = this.string(); if (path === "/type") this.rootMessageType = value; if (opaque && new TextEncoder().encode(value).length > 16384) this.fail("metadata string exceeds 16 KiB"); return; }
    if (c === "t") return this.literal("true"); if (c === "f") return this.literal("false"); if (c === "n") return this.literal("null");
    const match = this.text.slice(this.position).match(/^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?/); if (match === null) this.fail("invalid value"); this.position += match![0].length;
    if (opaque) { const n = Number(match![0]); if (n === 0 && !/^-?0(?:\.0*)?(?:[eE][+-]?\d+)?$/.test(match![0])) this.fail("metadata number underflows to zero"); if (!Number.isFinite(n) || (Number.isInteger(n) && !Number.isSafeInteger(n))) this.fail("metadata number is outside the finite JS-safe domain"); }
  }
  private object(path: string, depth: number, metadataDepth: number): void {
    this.position++; this.space(); const keys = new Set<string>();
    while (this.text[this.position] !== "}") {
      const key = this.string();
      if (metadataDepth > 0 && new TextEncoder().encode(key).length > 512) this.fail("metadata key exceeds 512 bytes");
      if (keys.has(key)) this.fail(`duplicate object key ${JSON.stringify(key)}`);
      keys.add(key);
      if (metadataDepth > 0 && keys.size > 256) this.fail("metadata object exceeds 256 members");
      this.space(); this.expect(":");
      const child = `${path}/${key.replaceAll("~", "~0").replaceAll("/", "~1")}`;
      this.value(child, depth + 1, metadataDepth > 0 ? metadataDepth + 1 : 0);
      this.space(); if (this.text[this.position] !== ",") break;
      this.position++; this.space(); if (this.text[this.position] === "}") this.fail("trailing comma");
    }
    this.expect("}");
  }
  private array(path: string, depth: number, metadataDepth: number): void {
    this.position++; this.space(); let count = 0;
    while (this.text[this.position] !== "]") {
      if (metadataDepth > 0 && count >= 4096) this.fail("metadata array exceeds 4096 elements");
      this.value(`${path}/${count++}`, depth + 1, metadataDepth > 0 ? metadataDepth + 1 : 0);
      this.space(); if (this.text[this.position] !== ",") break;
      this.position++; this.space(); if (this.text[this.position] === "]") this.fail("trailing comma");
    }
    this.expect("]");
  }
  private string(): string { const start = this.position; this.expect('"'); for (;;) { const c = this.text[this.position++]; if (c === undefined || c === "\n" || c === "\r") this.fail("unterminated string"); if (c === '"') break; if (c === "\\") { const escaped = this.text[this.position++]; if (escaped === "u") { const hex = this.text.slice(this.position, this.position + 4); if (!/^[0-9a-fA-F]{4}$/.test(hex)) this.fail("invalid Unicode escape"); const unit = Number.parseInt(hex, 16); this.position += 4; if (unit >= 0xd800 && unit <= 0xdbff) { if (this.text.slice(this.position, this.position + 2) !== "\\u") this.fail("unpaired Unicode surrogate"); const lowHex = this.text.slice(this.position + 2, this.position + 6); if (!/^[0-9a-fA-F]{4}$/.test(lowHex) || Number.parseInt(lowHex, 16) < 0xdc00 || Number.parseInt(lowHex, 16) > 0xdfff) this.fail("invalid Unicode surrogate pair"); this.position += 6; } else if (unit >= 0xdc00 && unit <= 0xdfff) this.fail("unpaired Unicode surrogate"); } else if (!'"\\/bfnrt'.includes(escaped ?? "")) this.fail("invalid escape"); } else { const unit = c.charCodeAt(0); if (unit < 0x20) this.fail("unescaped control character"); if (unit >= 0xd800 && unit <= 0xdbff) { const low = this.text.charCodeAt(this.position); if (low < 0xdc00 || low > 0xdfff) this.fail("unpaired Unicode surrogate"); this.position++; } else if (unit >= 0xdc00 && unit <= 0xdfff) this.fail("unpaired Unicode surrogate"); } } return JSON.parse(this.text.slice(start, this.position)) as string; }
  private literal(value: string): void { if (!this.text.startsWith(value, this.position)) this.fail("invalid literal"); this.position += value.length; }
  private expect(value: string): void { if (this.text[this.position] !== value) this.fail(`expected ${value}`); this.position++; }
  private space(): void { while (this.text[this.position] === " " || this.text[this.position] === "\t" || this.text[this.position] === "\n" || this.text[this.position] === "\r") this.position++; }
  private fail(reason: string): never { throw new TypeError(`Raw JSON validation failed at @peasant-labs/schema scanRawJsonText during pre-decode scanning near character ${this.position}: ${reason}; JSON.parse would otherwise lose or misinterpret evidence; correct the raw JSON and retry.`); }
}

function pointerMatches(pattern: string, path: string): boolean { const expected = pattern.split("/"); const actual = path.split("/"); return expected.length === actual.length && expected.every((part, index) => part === "*" || part === actual[index]); }

type Detail = import("./internal/generated/contract/zod.gen.js").SessionDetailPayload;
function validateSessionDetail(payload: Detail): void {
  const refs = new Set<string>();
  const turns = validateTurnEvidence(payload.turns ?? [], refs);
  if ((payload.nativeMetadata?.length ?? 0) > 0 && String(payload.harness) !== "pi") failSemantic("Pi native metadata requires harness pi");
  validateMetadataRecords(payload.nativeMetadata ?? [], turns, refs);
  for (const section of payload.earlierHistory ?? []) {
    const earlierTurns = validateTurnEvidence(section.turns ?? [], refs);
    validateMetadataRecords(section.nativeMetadata ?? [], earlierTurns, refs);
  }
}

// Producers can validate metadata before attaching it to a harness envelope.
// The full detail parsers reuse this same record and target validation.
export function parseNativeMetadataRecordsValue(value: unknown, targets: unknown = []): NonNullable<Detail["nativeMetadata"]> {
  rejectExplicitNullEvidence({nativeMetadata: value, turns: targets});
  const turns = zTurnDetail.array().parse(targets);
  const records = zNativeMetadataRecord.array().parse(value);
  validateMetadataRecords(records, validateTurnEvidence(turns));
  return records;
}

export function parseNativeMetadataRecordsText(text: string, targets: unknown = []): NonNullable<Detail["nativeMetadata"]> {
  scanRawJsonText(text, {maxDocumentBytes: 8 << 20, maxDocumentDepth: 64, opaqueMetadataPointers: ["/*/data"]});
  return parseNativeMetadataRecordsValue(JSON.parse(text), targets);
}

function validateTurnEvidence(targets: NonNullable<Detail["turns"]>, blockRefs = new Set<string>()): Map<number, NonNullable<Detail["turns"]>[number]> {
  const owners = new Set<string>(); const sources = new Set<string>();
  const turnIndexes = new Set<number>(); const toolIds = new Set<string>(); const turns = new Map<number, NonNullable<Detail["turns"]>[number]>();
  for (const turn of targets) {
    if (turn.observedModel !== undefined && (turn.role !== "assistant" || !validUnicode(turn.observedModel))) failSemantic("observedModel requires valid Unicode and an assistant role");
    if (turnIndexes.has(turn.index)) failSemantic("duplicate turn index makes metadata attachment ambiguous"); turnIndexes.add(turn.index); turns.set(turn.index, turn);
    if (turn.sourceEntryRef !== undefined && (!validOptionalRef(turn.sourceEntryRef) || blockRefs.has(turn.sourceEntryRef))) failSemantic("turn source reference is invalid, duplicated, or exceeds 96 bytes");
    if (turn.sourceEntryRef !== undefined) blockRefs.add(turn.sourceEntryRef);
    const projected = turn.sourceEntryRef !== undefined || turn.provenance !== undefined || (turn.toolCalls ?? []).some((tool) => tool.callEntryRef !== undefined || tool.resultEntryRef !== undefined || tool.callProvenance !== undefined || tool.resultProvenance !== undefined);
    if (projected && (((turn.depth ?? 0) === 0) !== (turn.parentIndex == null))) failSemantic("turn depth and parentIndex disagree");
    if (turn.usage != null) {
      if (!((turn.role === "assistant" && turn.usage.scope === "assistant") || (turn.role === "system" && turn.usage.scope === "summary"))) failSemantic("turn role and usage scope disagree");
      if (!turn.sourceEntryRef || turn.usage.sourceEntryRef !== turn.sourceEntryRef) failSemantic("turn usage source reference disagrees");
      validateUsage(turn.usage, owners, sources);
    }
    for (const tool of turn.toolCalls ?? []) { if (!validOptionalRef(tool.id)) failSemantic("tool id reference is invalid or exceeds 96 bytes"); if (tool.id && toolIds.has(tool.id)) failSemantic("duplicate tool call id makes metadata attachment ambiguous"); if (tool.id) toolIds.add(tool.id); for (const ref of [tool.callEntryRef, tool.resultEntryRef]) { if (ref !== undefined && (!validOptionalRef(ref) || blockRefs.has(ref))) failSemantic("tool reference is invalid, duplicated, or exceeds 96 bytes"); if (ref !== undefined) blockRefs.add(ref); } if (tool.usage != null) { if (tool.usage.scope !== "tool" || !tool.resultEntryRef || tool.usage.sourceEntryRef !== tool.resultEntryRef) failSemantic("tool usage attribution disagrees"); validateUsage(tool.usage, owners, sources); } }
  }
  for (const turn of targets) if (turn.parentIndex != null && !turns.has(turn.parentIndex)) failSemantic("parentIndex does not target a turn in the same partition");
  return turns;
}

function validateMetadataRecords(records: NonNullable<Detail["nativeMetadata"]>, turns: Map<number, NonNullable<Detail["turns"]>[number]>, blockRefs = new Set<string>()): void {
  const ids = new Set<string>();
  let metadataBytes = 0;
  if (records.length > 256) failSemantic("native metadata record count exceeds 256");
  for (const metadata of records) {
    if (!validRef(metadata.id) || !validRef(metadata.source.entryRef) || ids.has(metadata.id) || (metadata.attachment === undefined && blockRefs.has(metadata.source.entryRef))) failSemantic("metadata reference is invalid or duplicated"); ids.add(metadata.id); if (metadata.attachment === undefined) blockRefs.add(metadata.source.entryRef);
    if (metadata.customType !== undefined && (!validUnicode(metadata.customType) || new TextEncoder().encode(metadata.customType).length > 128)) failSemantic("customType is invalid Unicode or exceeds 128 bytes");
    validateMetadataValue(metadata.data, 1);
    const encoded = new TextEncoder().encode(JSON.stringify(metadata.data)).length; metadataBytes += encoded; if (encoded > 65536 || metadataBytes > 1048576) failSemantic("metadata data exceeds its byte budget");
    const attachment = metadata.attachment; const target = attachment?.turnIndex == null ? undefined : turns.get(attachment.turnIndex);
    if (attachment?.toolCallId !== undefined && !validRef(attachment.toolCallId)) failSemantic("attachment tool reference is invalid or exceeds 96 bytes");
    switch (metadata.kind) {
      case "pi.custom.data": if (metadata.source.sourceType !== "pi.custom" || metadata.source.messageRole !== undefined || attachment !== undefined || !metadata.customType) failSemantic("custom data metadata matrix disagrees"); break;
      case "pi.custommessage.details": if (metadata.source.sourceType !== "pi.custom_message" || metadata.source.messageRole !== undefined || attachment?.turnIndex === undefined || attachment.toolCallId !== undefined || !metadata.customType || target?.role !== "system" || target.sourceEntryRef !== metadata.source.entryRef) failSemantic("custom message metadata matrix disagrees"); break;
      case "pi.toolresult.details": { const tools = target?.toolCalls?.filter(tool => tool.id === attachment?.toolCallId) ?? []; const tool = tools[0]; if (metadata.source.sourceType !== "pi.message" || metadata.source.messageRole !== "toolResult" || attachment?.turnIndex === undefined || !attachment.toolCallId || metadata.customType !== undefined || tools.length !== 1 || tool?.resultEntryRef !== metadata.source.entryRef || (tool.usage != null && tool.usage.sourceEntryRef !== metadata.source.entryRef)) failSemantic("tool result metadata matrix disagrees"); break; }
      case "pi.compaction.details": case "pi.branchsummary.details": { const source = metadata.kind === "pi.compaction.details" ? "pi.compaction" : "pi.branch_summary"; if (metadata.source.sourceType !== source || metadata.source.messageRole !== undefined || attachment?.turnIndex === undefined || attachment.toolCallId !== undefined || metadata.customType !== undefined || target?.role !== "system" || target.sourceEntryRef !== metadata.source.entryRef || target.usage?.scope !== "summary" || target.usage.sourceEntryRef !== metadata.source.entryRef) failSemantic("summary metadata matrix disagrees"); break; }
      default: failSemantic("metadata kind is outside its closed set");
    }
  }
}
function validateUsage(usage: NonNullable<NonNullable<Detail["turns"]>[number]["usage"]>, owners: Set<string>, sources: Set<string>): void {
  if (!validRef(usage.ownerId) || !validRef(usage.sourceEntryRef) || owners.has(usage.ownerId) || sources.has(usage.sourceEntryRef)) failSemantic("usage owner/source reference is invalid or duplicated");
  owners.add(usage.ownerId); sources.add(usage.sourceEntryRef);
  const tokens = usage.tokens; const base = tokens == null ? [] : [tokens.input, tokens.output, tokens.cacheRead, tokens.cacheWrite, tokens.totalTokens]; const present = base.filter(v => v !== undefined && v !== null).length;
  for (const value of tokens == null ? [] : [...base, tokens.reasoning, tokens.cacheWrite1h]) if (value != null && (!Number.isSafeInteger(value) || value < 0)) failSemantic("token is outside JS-safe nonnegative integers");
  const expected = present === 5 ? "complete" : present === 0 ? "unknown" : "partial"; if (usage.completeness !== expected) failSemantic("usage completeness disagrees with token presence");
  if (tokens?.reasoning != null && tokens.output != null && tokens.reasoning > tokens.output) failSemantic("reasoning exceeds output"); if (tokens?.cacheWrite1h != null && tokens.cacheWrite != null && tokens.cacheWrite1h > tokens.cacheWrite) failSemantic("cacheWrite1h exceeds cacheWrite");
  if (present === 5 && tokens && tokens.input! + tokens.output! + tokens.cacheRead! + tokens.cacheWrite! !== tokens.totalTokens) failSemantic("totalTokens disagrees with base sum");
  if (usage.cost != null) { if (usage.cost.source !== "recorded_harness_estimate") failSemantic("recorded cost source is invalid"); for (const value of [usage.cost.input, usage.cost.output, usage.cost.cacheRead, usage.cost.cacheWrite, usage.cost.total]) if (value != null && (!/^-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?$/.test(value) || !Number.isFinite(Number(value)) || Number(value) < 0)) failSemantic("recorded cost is not a finite nonnegative JSON number"); }
}
function validRef(value: string | undefined): value is string { return value !== undefined && value.length > 0 && validOptionalRef(value); }
function validOptionalRef(value: string): boolean { return validUnicode(value) && new TextEncoder().encode(value).length <= 96; }
function validUnicode(value: string): boolean {
  for (let i = 0; i < value.length; i++) {
    const unit = value.charCodeAt(i);
    if (unit >= 0xd800 && unit <= 0xdbff) {
      const low = value.charCodeAt(++i);
      if (!(low >= 0xdc00 && low <= 0xdfff)) return false;
    } else if (unit >= 0xdc00 && unit <= 0xdfff) return false;
  }
  return true;
}
function validateMetadataValue(value: unknown, depth: number): void {
  if (depth > 32) failSemantic("metadata data exceeds depth 32");
  if (typeof value === "string" && (!validUnicode(value) || new TextEncoder().encode(value).length > 16384)) failSemantic("metadata string is invalid Unicode or exceeds 16 KiB");
  if (typeof value === "number" && (!Number.isFinite(value) || (Number.isInteger(value) && !Number.isSafeInteger(value)))) failSemantic("metadata number is outside the finite JS-safe domain");
  if (value !== null && !["string", "number", "boolean", "object"].includes(typeof value)) failSemantic("metadata data is not a JSON value");
  if (Array.isArray(value)) {
    if (value.length > 4096) failSemantic("metadata array exceeds 4096 elements");
    for (const item of value) validateMetadataValue(item, depth + 1);
  } else if (isRecord(value)) {
    if (Object.getPrototypeOf(value) !== Object.prototype && Object.getPrototypeOf(value) !== null) failSemantic("metadata data must use plain JSON objects");
    const entries = Object.entries(value);
    if (entries.length > 256) failSemantic("metadata object exceeds 256 members");
    for (const [key, item] of entries) {
      if (!validUnicode(key) || new TextEncoder().encode(key).length > 512) failSemantic("metadata key is invalid Unicode or exceeds 512 bytes");
      validateMetadataValue(item, depth + 1);
    }
  }
}
function isRecord(value: unknown): value is Record<string, unknown> { return typeof value === "object" && value !== null && !Array.isArray(value); }
function rejectExplicitNullEvidence(value: unknown): void {
  if (!isRecord(value)) return;
  if (value.inputSubmissionCount === null) failSemantic("inputSubmissionCount is explicitly null");
  if (value.inputSubmissionCount !== undefined && (typeof value.inputSubmissionCount !== "number" || !Number.isSafeInteger(value.inputSubmissionCount) || value.inputSubmissionCount < 0)) failSemantic("inputSubmissionCount is outside 0..9007199254740991");
  for (const turn of Array.isArray(value.turns) ? value.turns : []) {
    if (!isRecord(turn)) continue;
    if (turn.provenance === null) failSemantic("turn provenance is explicitly null");
    const tools = Array.isArray(turn.toolCalls) ? turn.toolCalls : [];
    for (const tool of tools) if (isRecord(tool) && (tool.callProvenance === null || tool.resultProvenance === null)) failSemantic("folded provenance is explicitly null");
    for (const usage of [turn.usage, ...tools.map(tool => isRecord(tool) ? tool.usage : undefined)]) {
      if (!isRecord(usage) || !isRecord(usage.tokens)) continue;
      for (const [name, token] of Object.entries(usage.tokens)) if (token === null) failSemantic(`token ${name} is explicitly null`);
    }
  }
  for (const section of Array.isArray(value.earlierHistory) ? value.earlierHistory : []) if (isRecord(section)) rejectExplicitNullEvidence({turns: section.turns, nativeMetadata: section.nativeMetadata});
  for (const record of Array.isArray(value.nativeMetadata) ? value.nativeMetadata : []) {
    if (!isRecord(record)) continue;
    for (const name of ["id", "kind", "source", "data"]) if (!(name in record) || record[name] === null) failSemantic(`metadata required field ${name} is missing or null`);
    if (record.attachment === null) failSemantic("metadata attachment is explicitly null");
    if (isRecord(record.attachment)) {
      if (record.attachment.turnIndex === null) failSemantic("metadata attachment.turnIndex is explicitly null");
      if ("toolCallId" in record.attachment && record.kind !== "pi.toolresult.details") failSemantic("metadata attachment.toolCallId is forbidden for this kind");
    }
    if (isRecord(record.source) && "messageRole" in record.source && record.kind !== "pi.toolresult.details") failSemantic("metadata source.messageRole is forbidden for this kind");
  }
}

const forbiddenGraphFields = new Set(["relationshipNavigation", "status", "transcriptId", "label", "explanation", "cooked", "collapsed", "url", "resolved", "detail"]);
function validateRawGraphValue(value: unknown): void {
  if (!isRecord(value)) return;
  for (const key of Object.keys(value)) if (forbiddenGraphFields.has(key)) failSemantic(`durable payload contains forbidden read-only or cooked field ${key}`);
  for (const relationship of Array.isArray(value.relationships) ? value.relationships : []) if (isRecord(relationship)) {
    for (const key of Object.keys(relationship)) if (!["kind", "targetState", "targetLocalId", "evidence", "anchor"].includes(key)) failSemantic(`relationship contains forbidden field ${key}`);
    if (isRecord(relationship.anchor)) for (const key of Object.keys(relationship.anchor)) if (!["kind", "sourceEntryRef", "sourceRevisionRef"].includes(key)) failSemantic(`anchor contains forbidden field ${key}`);
  }
  const visitTurns = (turns: unknown): void => {
    if (!Array.isArray(turns)) return;
    for (const turn of turns) {
      if (!isRecord(turn)) continue;
      for (const key of Object.keys(turn)) if (forbiddenGraphFields.has(key)) failSemantic(`turn contains forbidden read-only or cooked field ${key}`);
      for (const tool of Array.isArray(turn.toolCalls) ? turn.toolCalls : []) if (isRecord(tool)) for (const key of Object.keys(tool)) if (forbiddenGraphFields.has(key)) failSemantic(`tool contains forbidden read-only or cooked field ${key}`);
      for (const provenance of [turn.provenance, ...(Array.isArray(turn.toolCalls) ? turn.toolCalls.flatMap((tool) => isRecord(tool) ? [tool.callProvenance, tool.resultProvenance] : []) : [])]) if (isRecord(provenance)) for (const key of Object.keys(provenance)) if (!["origin", "actor", "delivery", "ownership", "evidence", "inputModality", "submissionRef"].includes(key)) failSemantic(`provenance contains forbidden field ${key}`);
    }
  };
  visitTurns(value.turns);
  for (const section of Array.isArray(value.earlierHistory) ? value.earlierHistory : []) if (isRecord(section)) {
    for (const key of Object.keys(section)) if (!["state", "turns", "nativeMetadata"].includes(key)) failSemantic(`earlier history contains forbidden field ${key}`);
    visitTurns(section.turns);
  }
}
function failSemantic(reason: string): never { throw new TypeError(`Session detail validation failed at @peasant-labs/schema public parser during semantic validation: ${reason}; consumers cannot safely attribute transcript evidence; correct the payload and retry.`); }
