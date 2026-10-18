package schema

// Deprecated: use VillageCollective. Retained for the complete alias release window.
type VillageGroup = VillageCollective

// Deprecated: use VillageCollectiveDetailRecord.
type VillageGroupDetailRecord = VillageCollectiveDetailRecord

// Deprecated: use VillagePublicCollective.
type VillagePublicGroup = VillagePublicCollective

// Deprecated: use VillageUserCollective.
type VillageUserGroup = VillageUserCollective

// Deprecated: use VillageVisibleCollective.
type VillageVisibleGroup = VillageVisibleCollective

// Deprecated: use VillageCreateCollectiveRequest.
type VillageCreateGroupRequest = VillageCreateCollectiveRequest

// Deprecated: use VillageUpdateCollectiveRequest.
type VillageUpdateGroupRequest = VillageUpdateCollectiveRequest

// Deprecated: use VillageCollectiveStatusRoleResponse.
type VillageGroupStatusRoleResponse = VillageCollectiveStatusRoleResponse

// Deprecated: use VillageRemoveCollectiveMemberResponse.
type VillageRemoveGroupMemberResponse = VillageRemoveCollectiveMemberResponse

// Deprecated: use VillageCollectiveMemberUsernameRequest.
type VillageGroupMemberUsernameRequest = VillageCollectiveMemberUsernameRequest

// Deprecated: use VillageCollectiveMemberRoleRequest.
type VillageGroupMemberRoleRequest = VillageCollectiveMemberRoleRequest

// Deprecated: use VillageCollectiveMember.
type VillageGroupMember = VillageCollectiveMember

// Deprecated: use VillageCollectiveTranscriptStats.
type VillageGroupTranscriptStats = VillageCollectiveTranscriptStats

// Deprecated: use VillageCollectiveModelBreakdown.
type VillageGroupModelBreakdown = VillageCollectiveModelBreakdown

// Deprecated: use VillageCollectiveContributor.
type VillageGroupContributor = VillageCollectiveContributor

// Deprecated: use VillageCollectiveDetailResponse.
type VillageGroupDetailResponse = VillageCollectiveDetailResponse

// Deprecated: use VillageCollectiveTranscript.
type VillageGroupTranscript = VillageCollectiveTranscript

// Deprecated: use VillageUserCollectiveShare.
type VillageUserGroupShare = VillageUserCollectiveShare

// Deprecated: use VillageGroupedCollectiveDetailResponse.
type VillageGroupedGroupDetailResponse = VillageGroupedCollectiveDetailResponse

// Deprecated: use VillageCollectiveAcceptanceMode.
type VillageGroupAcceptanceMode = VillageCollectiveAcceptanceMode

// Deprecated: use VillageCollectiveDataAccess.
type VillageGroupDataAccess = VillageCollectiveDataAccess

// Deprecated: use VillageCollectiveRole.
type VillageGroupRole = VillageCollectiveRole

// Deprecated: use VillageAssignableCollectiveRole.
type VillageAssignableGroupRole = VillageAssignableCollectiveRole

// Deprecated: use VillageCollectiveViewerRole.
type VillageGroupViewerRole = VillageCollectiveViewerRole

const (
	VillageCollectiveAcceptanceOpen            = VillageGroupAcceptanceOpen
	VillageCollectiveAcceptanceVerifiedOnly    = VillageGroupAcceptanceVerifiedOnly
	VillageCollectiveAcceptanceCurated         = VillageGroupAcceptanceCurated
	VillageCollectiveDataAccessMembersOnly     = VillageGroupDataAccessMembersOnly
	VillageCollectiveDataAccessContributors    = VillageGroupDataAccessContributors
	VillageCollectiveDataAccessPublic          = VillageGroupDataAccessPublic
	VillageCollectiveRoleOwner                 = VillageGroupRoleOwner
	VillageCollectiveRoleMember                = VillageGroupRoleMember
	VillageCollectiveRoleContributor           = VillageGroupRoleContributor
	VillageCollectiveRolePending               = VillageGroupRolePending
	VillageAssignableCollectiveRoleContributor = VillageAssignableGroupRoleContributor
	VillageAssignableCollectiveRoleMember      = VillageAssignableGroupRoleMember
	VillageCollectiveViewerRoleNone            = VillageGroupViewerRoleNone
	VillageCollectiveViewerRoleOwner           = VillageGroupViewerRoleOwner
	VillageCollectiveViewerRoleMember          = VillageGroupViewerRoleMember
	VillageCollectiveViewerRoleContributor     = VillageGroupViewerRoleContributor
	VillageCollectiveViewerRolePending         = VillageGroupViewerRolePending
)

var AllVillageCollectiveAcceptanceModes = AllVillageGroupAcceptanceModes
var AllVillageCollectiveDataAccessPolicies = AllVillageGroupDataAccessPolicies
var AllVillageCollectiveRoles = AllVillageGroupRoles
var AllVillageAssignableCollectiveRoles = AllVillageAssignableGroupRoles
var AllVillageCollectiveViewerRoles = AllVillageGroupViewerRoles
