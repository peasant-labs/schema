import type { PublishRequest, TimestampInfo } from "@peasant-labs/schema";

declare const timestamp: TimestampInfo;
declare const publication: PublishRequest;
const timestampStart: bigint = timestamp.start;
const timestampEnd: bigint = timestamp.end;
const publicationStart: bigint = publication.timestamp.start;
const publicationEnd: bigint = publication.timestamp.end;
// @ts-expect-error formatted int64 output is not a number
const incompatible: number = timestamp.start;
void [timestampStart, timestampEnd, publicationStart, publicationEnd, incompatible];
