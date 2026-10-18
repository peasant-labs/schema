import {
  zVillageDeprecatedGroupedContributableResponse,
  type VillageDeprecatedGroupedContributableResponse,
  type VillageGroupedContributableResponse,
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
declare const deprecatedGroupedContribution: VillageDeprecatedGroupedContributableResponse;
declare const canonicalGroupedContribution: VillageGroupedContributableResponse;
zVillageDeprecatedGroupedContributableResponse.parse(deprecatedGroupedContribution);
const retainedGroupID: string = deprecatedGroupedContribution.groupId;
const canonicalGroupID: string = canonicalGroupedContribution.group_id;
// @ts-expect-error The deprecated route retains only the released key spelling.
deprecatedGroupedContribution.group_id;
// @ts-expect-error The canonical route retains only the corrected key spelling.
canonicalGroupedContribution.groupId;
import { PeasantLocalAPIVersion, type operations as LocalOperations } from "@peasant-labs/schema/local-api";
import { VillageAPIVersion, type operations as VillageOperations } from "@peasant-labs/schema/village-api";
import type { UnifiedMetadata } from "@peasant-labs/schema/types";
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
declare const localOperations: LocalOperations;
declare const villageOperations: VillageOperations;
declare const metadata: UnifiedMetadata;
declare const versionResponse: SchemaVersionResponse;
declare const pushContractVersion: PushContractVersion;
const role: Role = Role.User;
const projectHash: ProjectHash = newProjectHash("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2");
const corpus: Corpus<string, boolean> = { cases: [] };
const fixtures = loadQualityFixtures();
const timelineFixtures: TimelineFixtureCorpus = loadTimelineFixtures();
const firstTimelineFixture: TimelineFixtureCase | undefined = timelineFixtures.cases[0];

void [MetadataSchemaVersion, TypesVersion, PeasantLocalAPIVersion, VillageAPIVersion, isRole(role), isFileChangeStatus(FileChangeStatus.Modified), isDiffLineKind(DiffLineKind.Add), projectHash, detail, dashboard, publish, localOperations, villageOperations, metadata, versionResponse, pushContractVersion, corpus, Classification.MustPass, qualitySessions(fixtures), timelineFixtures, firstTimelineFixture];
