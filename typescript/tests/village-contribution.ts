// Compile-time proof that a consumer can switch exhaustively over every closed
// menu the contribution surface adds or leans on. Each switch ends in a
// `never` assignment, so adding a member to a menu without handling it here is
// a type error, and so is a handler for a token the menu does not have.
import type {
  VillageContributableView,
  VillageContributionErrorCode,
  VillageProjectNameSource,
  VillageShareStatus,
} from "@peasant-labs/schema";
import type { operations as VillageOperations } from "@peasant-labs/schema/village-api";

type Same<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;

export function describeView(view: VillageContributableView): string {
  switch (view) {
    case "grouped":
      return "helper groups";
    case "paged":
      return "one page of root sessions";
    case "snapshot":
      return "compact records of one project";
    default: {
      const unreachable: never = view;
      return unreachable;
    }
  }
}

export function describeErrorCode(code: VillageContributionErrorCode): string {
  switch (code) {
    case "moved":
    case "missing":
    case "live":
    case "private_unapproved":
    case "conflict":
      return "transcript";
    case "policy":
    case "membership":
    case "stale_selection":
      return "request";
    case "snapshot_timeout":
    case "snapshot_busy":
      return "snapshot";
    default: {
      const unreachable: never = code;
      return unreachable;
    }
  }
}

export function describeShareStatus(status: VillageShareStatus): string {
  switch (status) {
    case "pending":
    case "approved":
    case "rejected":
    case "retracted":
    case "revoked":
      return status;
    default: {
      const unreachable: never = status;
      return unreachable;
    }
  }
}

export function describeNameSource(source: VillageProjectNameSource): string {
  switch (source) {
    case "override":
    case "consented":
    case "remote":
    case "path":
    case "privacy":
      return source;
    default: {
      const unreachable: never = source;
      return unreachable;
    }
  }
}

// The operation contract uses the same menus the root package exports.
type ContributableQuery = NonNullable<VillageOperations["listContributableTranscripts"]["parameters"]["query"]>;
const viewParameterIsTheMenu: Same<NonNullable<ContributableQuery["view"]>, VillageContributableView> = true;
void viewParameterIsTheMenu;

type SelectionRefusal = VillageOperations["batchShareSelection"]["responses"][409]["content"]["application/json"];
const refusalCodeIsTheMenu: Same<NonNullable<SelectionRefusal["code"]>, VillageContributionErrorCode> = true;
void refusalCodeIsTheMenu;
