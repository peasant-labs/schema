import { zConversationTurnSpan, type ConversationTurnSpan } from "./internal/generated/contract/zod.gen.js";

// Where a decoded segment-directory response places each segment among the
// session's conversation turns. This is the browser mirror of Go's four
// validateDirectoryConversationTurns rules: both counts of a stated span are
// non-negative, which the generated span shape already enforces; the total is a
// count; a span ends at or before it; and immediate neighbours that both state
// a span must meet exactly (an omitted span breaks the chain).
//
// Go's validateDirectoryConversationTurns remains the semantic reference: when
// the shared position rule changes, both implementations change in the same PR,
// and the shared fixtures detect disagreement only for the cases they cover.
//
// The browser is deliberately stricter than Go's struct validation in three
// places, following the generated wire shape: an explicit null, a number
// outside the safe-integer domain, and a span missing before or count are all
// unusable here where Go's non-pointer int fields would read zero. The helper
// also sees JavaScript numbers after JSON parsing, so it cannot detect a number
// lexeme that already rounded to a safe integer; this is not raw-token parity.
//
// false means the optional positions are unusable and a consumer should drop
// them. It is not a rejection of the directory or its turns. true is not an
// attestation of required fields, authorization, other pages, or a future
// negotiated representation.

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isNonNegativeSafeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

// The generated ConversationTurnSpan shape is the only accepted span: both
// counts present, integral, and inside 0..9007199254740991.
function parseSpan(value: unknown): ConversationTurnSpan | undefined {
  const parsed = zConversationTurnSpan.safeParse(value);
  return parsed.success ? parsed.data : undefined;
}

export function isTranscriptDirectoryPositionsValid(directory: unknown): boolean {
  if (!isRecord(directory)) return false;
  const segments = directory.segments;
  if (!Array.isArray(segments)) return false;

  const statedTotal = directory.totalConversationTurns;
  if (statedTotal !== undefined && !isNonNegativeSafeInteger(statedTotal)) return false;

  // Undefined means the previous ref stated no span, so the chain of
  // comparisons stops there instead of bridging to the next stated span.
  let previous: ConversationTurnSpan | undefined;
  for (const ref of segments) {
    if (!isRecord(ref)) return false;
    const stated = ref.conversationTurns;
    let span: ConversationTurnSpan | undefined;
    if (stated !== undefined) {
      span = parseSpan(stated);
      if (span === undefined) return false;
      if (statedTotal !== undefined && !endsAtOrBefore(span, statedTotal)) return false;
    }
    if (span !== undefined && previous !== undefined && !continues(previous, span)) return false;
    previous = span;
  }
  return true;
}

// before + count can pass Number.MAX_SAFE_INTEGER and round, so compare by an
// exact difference instead of the sum. Both operands are safe integers, so
// total - count is exact.
function endsAtOrBefore(span: ConversationTurnSpan, total: number): boolean {
  return span.count <= total && span.before <= total - span.count;
}

// The exact difference between two safe integers is exact, so this is the
// strict next.before === previous.before + previous.count without the sum.
function continues(previous: ConversationTurnSpan, next: ConversationTurnSpan): boolean {
  return next.before - previous.before === previous.count;
}
