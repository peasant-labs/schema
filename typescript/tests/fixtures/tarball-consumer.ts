import {
  MetadataSchemaVersion,
  newProjectHash,
  DiffLineKind,
  FileChangeStatus,
  Role,
  TypesVersion,
  isDiffLineKind,
  isFileChangeStatus,
  isRole,
  type SchemaVersionResponse,
  type PushContractVersion,
  type DashboardPayload,
  type PublishRequest,
  type SessionDetailPayload,
  type ProjectHash,
} from "@peasant-labs/schema";
import type { UnifiedMetadata } from "@peasant-labs/schema";
import { Classification, type Corpus } from "@peasant-labs/schema/testcase";
import { loadQualityFixtures } from "@peasant-labs/schema/fixtures";
import { qualitySessions } from "@peasant-labs/schema/fixtures/quality";
import {
  loadTimelineFixtures,
  type TimelineFixtureCase,
  type TimelineFixtureCorpus,
} from "@peasant-labs/schema/fixtures/timeline";

declare const detail: SessionDetailPayload;
declare const dashboard: DashboardPayload;
declare const publish: PublishRequest;
declare const metadata: UnifiedMetadata;
declare const versionResponse: SchemaVersionResponse;
declare const pushContractVersion: PushContractVersion;
const role: Role = Role.User;
const projectHash: ProjectHash = newProjectHash("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2");
const corpus: Corpus<string, boolean> = { cases: [] };
const fixtures = loadQualityFixtures();
const timelineFixtures: TimelineFixtureCorpus = loadTimelineFixtures();
const firstTimelineFixture: TimelineFixtureCase | undefined = timelineFixtures.cases[0];

void [MetadataSchemaVersion, TypesVersion, isRole(role), isFileChangeStatus(FileChangeStatus.Modified), isDiffLineKind(DiffLineKind.Add), projectHash, detail, dashboard, publish, metadata, versionResponse, pushContractVersion, corpus, Classification.MustPass, qualitySessions(fixtures), timelineFixtures, firstTimelineFixture];
