package release_test

import (
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

func TestIsMaintainer(t *testing.T) {
	cases := []struct {
		perm release.CollaboratorPermission
		want bool
	}{
		{release.PermAdmin, true},
		{release.PermMaintain, true},
		{release.PermWrite, false},
		{release.PermTriage, false},
		{release.PermRead, false},
		{release.PermNone, false},
		{"", false},
		{"Admin", false}, // case-sensitive: the API returns lowercase
		{"superadmin", false},
	}
	for _, tc := range cases {
		t.Run(string(tc.perm), func(t *testing.T) {
			if got := release.IsMaintainer(tc.perm); got != tc.want {
				t.Errorf("IsMaintainer(%q) = %v, want %v", tc.perm, got, tc.want)
			}
		})
	}
}
