# Grouped read contracts

Grouped REST responses are opt-in with `view=grouped`. Omission retains the
legacy list shape. WebSocket session lists remain unchanged; hosts use updates
to invalidate grouped REST queries. Durable transcript content never includes
helper groups or authorized navigation.

## Helper members can own helper groups

`LocalHelperMembersPayload.members` and `VillageHelperMembersPayload.members`
use the existing `LocalSessionListItem` and `VillageSessionListItem` envelopes.
Every member must have `kind: transcript`, a non-null `transcript`, and no
populated `context`. Context containers exist only at the grouped list boundary.
The member's `helperGroups` describes its immediate helper children. There is
no duplicate group field on the session row or durable session.

The Go JSON decoders and generated public Zod validators reject context members,
missing transcript arms, duplicate member identities, duplicate group IDs,
inconsistent route-row mirrors, and invalid pagination or counts. Members and
items are non-null arrays. Page and limit are positive safe integers; totals
and helper thread counts are nonnegative safe integers. A member page cannot
contain more rows than its limit or direct total. A zero total is measured zero,
not a replacement for missing or null evidence. Optional input submission counts
remain independent of turn and helper counts; null is invalid.

For an admitted ordinary P, helper G1 owned by P, and helper G2 owned by G1:

- Root list: P with one direct helper, ordinary total 1, helper total 2, item
  total 1. Do not add a second top-level container for G1's represented group.
- P's group expansion: one transcript item for G1, direct total 1, with G1's
  own helper group and its distinct member scope.
- G1's group expansion: one transcript item for G2, direct total 1.
- A search matching only G2: one context container for its missing/excluded
  immediate owner, whose expansion returns G2 only, never P or G1.

Hosts construct the owner/group index over the full currently authorized and
eligible original-query candidate set before member paging. Nested scopes retain
the same viewer and parsed route/search/profile/project/collective/status/
selection predicates, but name the nested group's own ID. The outer group's
final membership restriction is not the nested candidate universe. Scope misses,
expiration, or changed selection require an actionable list refresh, never a
fallback to all members. Every expansion rechecks current rights and eligibility.

Each disclosure has independent paging and callback state. A member key and
selection callback identify exactly `item.transcript.session.id`; group IDs and
scope tokens are never transcript selections or access grants. Cyclic/conflicting
ownership stays in independent unresolved groups, not recursive expansion loops.

## Collective detail compatibility

Both `VillageGroupDetailResponse.group` and
`VillageGroupedGroupDetailResponse.group` are `VillageGroupDetailRecord`. This
read projection retains the eleven released base fields, including nullable
description and linked organization, and adds optional `post_prompts_check` and
`prompts_check_mode` pointers. Producers without these stored facts omit them;
absence is unknown, not false or a guessed mode. Explicit false is retained and
present modes use the canonical closed menu. Canonical `VillageGroup` and its
independent required prompts configuration contract are unchanged.

## Content and publication

`GET /api/v1/transcripts/{id}/content` returns `TranscriptContent`, containing
`contractVersion`, `kind`, and `sessionDetail`, not a bare detail or read DTO.
Authorized navigation comes from the separate metadata read. The existing
unimplemented batch-upload route remains a refusal; this contract does not add
batch upload DTOs or a successful operation. Existing successful single publishes
and explicit collective batch share/review actions keep their preservation and
individual-transcript selection requirements.
