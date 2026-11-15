import { readTranscriptV2Turn, type TranscriptV2Turn, type TurnDetail } from "@peasant-labs/schema";

declare const source: unknown;
const turn = readTranscriptV2Turn(source);
if (turn !== undefined) {
  const strictView: Omit<TurnDetail, "index"> = turn;
  const maybeIndex: unknown = turn.index;
  // The recorded index must come from the segment row, never from this view.
  // @ts-expect-error the raw turn does not promise a numeric index
  const recordedIndex: number = turn.index;
  void [strictView, maybeIndex, recordedIndex];
}

const read: (value: unknown) => TranscriptV2Turn | undefined = readTranscriptV2Turn;
void read;
