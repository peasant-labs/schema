package schema

// VillageOrganizationCollective is the existing viewer-permitted collective
// projection returned by the organization read. It deliberately carries neither
// rosters nor a locally reconstructed visibility decision.
type VillageOrganizationCollective struct {
	ID              VillageUUID `json:"id"`
	Name            string      `json:"name"`
	Description     *string     `json:"description"`
	MemberCount     int32       `json:"member_count" minimum:"0"`
	TranscriptCount int32       `json:"transcript_count" minimum:"0"`
}

// VillageOrganizationCollectivesResponse is the narrow metadata consumer view
// of GET /orgs/{login}. The existing endpoint also returns organization stats and
// members; additive fields are permitted but are not retained by this contract
// projection or used to derive permissions.
type VillageOrganizationCollectivesResponse struct {
	Collectives []VillageOrganizationCollective `json:"collectives" nullable:"false"`
}

func (v VillageOrganizationCollectivesResponse) Validate() error {
	if v.Collectives == nil {
		return observationError("organization collective projection requires a complete nonnull disclosed set")
	}
	seen := map[VillageUUID]bool{}
	for _, c := range v.Collectives {
		if !transcriptIDUUID.MatchString(string(c.ID)) || seen[c.ID] || c.MemberCount < 0 || c.TranscriptCount < 0 {
			return observationError("organization collective projection has invalid/repeated identity or negative counts")
		}
		seen[c.ID] = true
	}
	return nil
}
