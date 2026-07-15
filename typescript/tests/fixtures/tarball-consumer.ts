import {
  MetadataSchemaVersion,
  Role,
  TypesVersion,
  isRole,
  type SchemaVersionResponse,
  type PushContractVersion,
  type SessionDetailPayload,
} from "@peasant-labs/schema";
import type { DashboardPayload } from "@peasant-labs/schema/local-api";
import type { PublishRequest } from "@peasant-labs/schema/village-api";
import type { UnifiedMetadata } from "@peasant-labs/schema/types";
import { Classification, type Corpus } from "@peasant-labs/schema/testcase";
import { loadQualityFixtures } from "@peasant-labs/schema/fixtures";
import { qualitySessions } from "@peasant-labs/schema/fixtures/quality";
import { loadTimelineFixtures } from "@peasant-labs/schema/fixtures/timeline";

declare const detail: SessionDetailPayload;
declare const dashboard: DashboardPayload;
declare const publish: PublishRequest;
declare const metadata: UnifiedMetadata;
declare const versionResponse: SchemaVersionResponse;
declare const pushContractVersion: PushContractVersion;
const role: Role = Role.User;
const corpus: Corpus<string, boolean> = { cases: [] };
const fixtures = loadQualityFixtures();
const timelineFixtures = loadTimelineFixtures();

void [MetadataSchemaVersion, TypesVersion, isRole(role), detail, dashboard, publish, metadata, versionResponse, pushContractVersion, corpus, Classification.MustPass, qualitySessions(fixtures), timelineFixtures];
