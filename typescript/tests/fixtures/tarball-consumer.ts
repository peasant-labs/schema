import {
  Role,
  TypesVersion,
  isRole,
  type SchemaVersionResponse,
  type SessionDetailPayload,
} from "@peasant-labs/schema";
import type { DashboardPayload } from "@peasant-labs/schema/local-api";
import type { PublishRequest } from "@peasant-labs/schema/village-api";
import type { UnifiedMetadata } from "@peasant-labs/schema/types";
import { Classification, type Corpus } from "@peasant-labs/schema/testcase";
import { loadQualityFixtures } from "@peasant-labs/schema/fixtures";
import { qualitySessions } from "@peasant-labs/schema/fixtures/quality";

declare const detail: SessionDetailPayload;
declare const dashboard: DashboardPayload;
declare const publish: PublishRequest;
declare const metadata: UnifiedMetadata;
declare const versionResponse: SchemaVersionResponse;
const role: Role = Role.User;
const corpus: Corpus<string, boolean> = { cases: [] };
const fixtures = loadQualityFixtures();

void [TypesVersion, isRole(role), detail, dashboard, publish, metadata, versionResponse, corpus, Classification.MustPass, qualitySessions(fixtures)];
