import { EntryType, Role } from "./internal/generated/enums.gen.js";
import type { TurnDetail } from "./internal/generated/contract/zod.gen.js";

const conversationTurnKinds = new Set<string>([
  `${Role.User}|${EntryType.Text}`,
  `${Role.User}|${EntryType.Image}`,
  `${Role.Assistant}|${EntryType.Text}`,
  `${Role.Assistant}|${EntryType.Image}`,
]);

/**
 * Whether a turn is a conversation turn: depth 0 and its (role, entryType) on
 * the allowed list. A missing or unknown role or entry type does not count.
 */
export function isConversationTurn(turn: Pick<TurnDetail, "role" | "entryType" | "depth">): boolean {
  if (turn.depth !== 0) return false;
  if (typeof turn.role !== "string" || typeof turn.entryType !== "string") return false;
  return conversationTurnKinds.has(`${turn.role}|${turn.entryType}`);
}
