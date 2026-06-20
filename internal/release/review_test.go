package release_test

import (
	"slices"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

func TestLatestApprovers(t *testing.T) {
	cases := []struct {
		name    string
		reviews []release.Review
		want    []string
	}{
		{
			name: "zero reviews",
		},
		{
			name: "approved reviewers",
			reviews: []release.Review{
				review("alice", release.ReviewStateApproved),
				review("bob", release.ReviewStateApproved),
			},
			want: []string{"alice", "bob"},
		},
		{
			name: "dismissed withdraws approval",
			reviews: []release.Review{
				review("alice", release.ReviewStateApproved),
				review("alice", release.ReviewStateDismissed),
			},
		},
		{
			name: "changes requested after approved withdraws approval",
			reviews: []release.Review{
				review("alice", release.ReviewStateApproved),
				review("alice", release.ReviewStateChangesRequested),
			},
		},
		{
			name: "later plain comment does not shadow standing approval",
			reviews: []release.Review{
				review("alice", release.ReviewStateApproved),
				review("alice", release.ReviewStateCommented),
			},
			want: []string{"alice"},
		},
		{
			name: "bot reviews are included for downstream maintainer filtering",
			reviews: []release.Review{
				review("dependabot[bot]", release.ReviewStateApproved),
			},
			want: []string{"dependabot[bot]"},
		},
		{
			name: "page straddle approve then withdraw from complete merged input",
			reviews: []release.Review{
				// Page 1.
				review("alice", release.ReviewStateApproved),
				review("bob", release.ReviewStateApproved),
				// Page 2.
				review("alice", release.ReviewStateChangesRequested),
			},
			want: []string{"bob"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := release.LatestApprovers(tc.reviews)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("LatestApprovers() = %v, want %v", got, tc.want)
			}
		})
	}
}

func review(login string, state release.ReviewState) release.Review {
	return release.Review{
		User:  &release.ReviewUser{Login: login},
		State: state,
	}
}
