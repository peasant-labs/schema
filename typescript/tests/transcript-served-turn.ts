import { readTranscriptServedTurn, type TranscriptServedTurn, type TurnDetail } from "@peasant-labs/schema";

declare const source: unknown;
const turn = readTranscriptServedTurn(source);
if (turn !== undefined) {
  const strictView: Omit<TurnDetail, "index"> = turn;
  const maybeIndex: unknown = turn.index;
  // The recorded index must come from the segment row, never from this view.
  // @ts-expect-error the raw turn does not promise a numeric index
  const recordedIndex: number = turn.index;
  void [strictView, maybeIndex, recordedIndex];
}

const read: (value: unknown) => TranscriptServedTurn | undefined = readTranscriptServedTurn;
void read;
