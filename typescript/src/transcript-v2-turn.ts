import type { TurnDetail } from "./internal/generated/contract/zod.gen.js";

/**
 * The strict body's typed view. A served legacy body may omit any member, and
 * index is never an address: use the row's recordedIndex and
 * recordedIndexState instead. Read every other member defensively.
 *
 * This view preserves values, not JSON bytes. Re-encoding may change whitespace,
 * number lexemes, or escapes; the wire's byte guarantee belongs to the server.
 */
export type TranscriptV2Turn = Omit<TurnDetail, "index"> & { readonly index?: unknown };

/**
 * Read a server-validated v2 turn. This checks object-ness only, not the
 * members: legacy turns need not have the strict shape. Non-objects are refused.
 */
export function readTranscriptV2Turn(value: unknown): TranscriptV2Turn | undefined {
  if (typeof value !== "object" || value === null || Array.isArray(value)) return undefined;
  // The strict view describes the contract, not a guarantee about legacy
  // members. Consumers must inspect every member before using it.
  return value as TranscriptV2Turn;
}
