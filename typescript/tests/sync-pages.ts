import type { LocalSessionPage, LocalSyncSessionPage, LocalSyncSessionRow, LocalSyncSessionTotals, LocalSyncProjectOptions, PageSize, ProjectHash } from "@peasant-labs/schema";
import type { operations } from "@peasant-labs/schema/local-api";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type SyncResponse = operations["listSyncSessionsGrouped"]["responses"][200]["content"]["application/json"];

const operationPage: Same<Extract<SyncResponse, { totals: unknown }>, LocalSyncSessionPage> = true;
const operationOptions: Same<Extract<SyncResponse, { projects: unknown[] }>, LocalSyncProjectOptions> = true;
const pageRow: Same<LocalSyncSessionPage["items"][number], LocalSyncSessionRow> = true;
const pageSize: Same<LocalSyncSessionPage["perPage"], PageSize> = true;
const totals: Same<LocalSyncSessionPage["totals"], LocalSyncSessionTotals> = true;
const flatPageSize: Same<LocalSessionPage["perPage"], PageSize> = true;
const optionProjectHash: Same<LocalSyncProjectOptions["projects"][number]["projectHash"], ProjectHash> = true;
void [operationPage, operationOptions, pageRow, pageSize, totals, flatPageSize, optionProjectHash];
