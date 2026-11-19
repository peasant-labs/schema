import {
  AllRoles,
  EntryType,
  PageSize,
  Role,
  VillageCollectiveRole,
  type AuthoritativeQualityMetrics,
  type EntryType as EntryTypeValue,
  type License as LicenseValue,
  type PageSize as PageSizeValue,
  type PublishLicenseOperation,
  type SessionOutcome as SessionOutcomeValue,
  type TurnDetail,
  type VillageCollectiveRole as VillageCollectiveRoleValue,
  type Role as RoleValue,
  zRole,
} from "@peasant-labs/schema";

type ExactType<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;

// A member keeps its own literal type; it is not widened to the whole set.
const roleUserIsLiteral: ExactType<typeof Role.User, "user"> = true;
const roleAssistantIsLiteral: ExactType<typeof Role.Assistant, "assistant"> = true;
const userMember: "user" = Role.User;
const pageSizeMemberIsNumberLiteral: ExactType<typeof PageSize.P25, 25> = true;
const entryTypeMemberIsNotTheSet: ExactType<typeof EntryType.Text, EntryTypeValue> = false;
const collectiveRoleOwnerIsLiteral: ExactType<typeof VillageCollectiveRole.Owner, "owner"> = true;

// The named type is the union of the table's values, and the inventory uses it.
const roleIsTableValues: ExactType<RoleValue, "user" | "assistant" | "tool" | "system"> = true;
const pageSizeIsTableValues: ExactType<PageSizeValue, 25 | 50 | 100> = true;
const inventoryIsNamed: ExactType<(typeof AllRoles)[number], RoleValue> = true;
const collectiveRoleIsTableValues: ExactType<VillageCollectiveRoleValue, (typeof VillageCollectiveRole)[keyof typeof VillageCollectiveRole]> = true;

// Generated fields and schemas carry the named type.
const turnRoleIsRole: ExactType<TurnDetail["role"], RoleValue> = true;
declare const turn: TurnDetail;
const turnRole: RoleValue = turn.role;
const parsedRole: RoleValue = zRole.parse("user");
const schemaOptions: readonly RoleValue[] = zRole.options;
const memberIsAssignableToSet: RoleValue = Role.Tool;

// Optional, nullable and nullish fields keep the named set too.
const turnEntryTypeIsEntryType: ExactType<TurnDetail["entryType"], EntryTypeValue | undefined> = true;
const turnEntryType: EntryTypeValue | undefined = turn.entryType;
const turnIsThinking: boolean = turn.entryType === EntryType.Thinking;
const nullableLicenseIsLicense: ExactType<PublishLicenseOperation["license"], LicenseValue | null> = true;
const nullishOutcomeIsSessionOutcome: ExactType<AuthoritativeQualityMetrics["outcome"], SessionOutcomeValue | null | undefined> = true;

// A wrong literal is rejected through the set, the schema, and a field.
// @ts-expect-error "operator" is not a Role.
const wrongSetLiteral: RoleValue = "operator";
// @ts-expect-error a Role member is not another set's member.
const crossSetMember: RoleValue = VillageCollectiveRole.Owner;
// @ts-expect-error a field typed Role refuses a literal outside the set.
const wrongFieldLiteral: TurnDetail["role"] = "operator";
// @ts-expect-error the member is the literal "user", never "assistant".
const wrongMemberLiteral: typeof Role.User = "assistant";
// @ts-expect-error an optional field typed EntryType refuses a literal outside the set.
const wrongOptionalFieldLiteral: TurnDetail["entryType"] = "reasoning";

void [roleUserIsLiteral, roleAssistantIsLiteral, userMember, pageSizeMemberIsNumberLiteral, entryTypeMemberIsNotTheSet, collectiveRoleOwnerIsLiteral, roleIsTableValues, pageSizeIsTableValues, inventoryIsNamed, collectiveRoleIsTableValues, turnRoleIsRole, turnRole, parsedRole, schemaOptions, memberIsAssignableToSet, turnEntryTypeIsEntryType, turnEntryType, turnIsThinking, nullableLicenseIsLicense, nullishOutcomeIsSessionOutcome, wrongSetLiteral, crossSetMember, wrongFieldLiteral, wrongMemberLiteral, wrongOptionalFieldLiteral];
