package release_test

import (
	"strings"
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

func TestCheckFinal(t *testing.T) {
	cases := []struct {
		name      string
		final     release.Version
		rcs       []release.RCStatus
		wantErr   bool
		errSubstr string
	}{
		{
			name:  "green ancestor rc proceeds",
			final: "v0.1.0",
			rcs: []release.RCStatus{
				{Tag: "v0.1.0-rc1", RunGreen: true, IsAncestor: true},
			},
		},
		{
			name:  "one of several rcs is green+ancestor proceeds",
			final: "v0.1.0",
			rcs: []release.RCStatus{
				{Tag: "v0.1.0-rc1", RunGreen: false, IsAncestor: true},
				{Tag: "v0.1.0-rc2", RunGreen: true, IsAncestor: true},
			},
		},
		{
			name:      "no rc at all is blocked",
			final:     "v0.1.0",
			rcs:       nil,
			wantErr:   true,
			errSubstr: "no same-version release candidate",
		},
		{
			name:  "rc exists but not green is blocked",
			final: "v0.1.0",
			rcs: []release.RCStatus{
				{Tag: "v0.1.0-rc1", RunGreen: false, IsAncestor: true},
			},
			wantErr:   true,
			errSubstr: "not green",
		},
		{
			name:  "rc green but not ancestor is blocked",
			final: "v0.1.0",
			rcs: []release.RCStatus{
				{Tag: "v0.1.0-rc1", RunGreen: true, IsAncestor: false},
			},
			wantErr:   true,
			errSubstr: "not an ancestor",
		},
		{
			name:  "only different-base rc is blocked",
			final: "v0.2.0",
			rcs: []release.RCStatus{
				{Tag: "v0.1.0-rc1", RunGreen: true, IsAncestor: true},
			},
			wantErr:   true,
			errSubstr: "no same-version release candidate",
		},
		{
			name:      "non-final input is rejected",
			final:     "v0.1.0-rc1",
			rcs:       nil,
			wantErr:   true,
			errSubstr: "non-final version",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := release.CheckFinal(tc.final, tc.rcs)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("CheckFinal(%q, %+v): expected error, got nil", tc.final, tc.rcs)
				}
				if tc.errSubstr != "" && !strings.Contains(err.Error(), tc.errSubstr) {
					t.Errorf("CheckFinal(%q): error %q does not contain %q", tc.final, err.Error(), tc.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CheckFinal(%q, %+v): unexpected error: %v", tc.final, tc.rcs, err)
			}
		})
	}
}
