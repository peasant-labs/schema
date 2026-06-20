package release_test

import (
	"testing"

	"github.com/peasant-labs/schema/internal/release"
)

func TestNewVersion(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		want    release.Version
		wantErr bool
	}{
		{name: "final", raw: "v0.1.0", want: "v0.1.0"},
		{name: "final multi-digit", raw: "v12.34.56", want: "v12.34.56"},
		{name: "rc1", raw: "v0.1.0-rc1", want: "v0.1.0-rc1"},
		{name: "rc multi-digit", raw: "v2.0.0-rc12", want: "v2.0.0-rc12"},

		// must-fail grammar set
		{name: "missing leading v", raw: "0.1.0", wantErr: true},
		{name: "rc without number", raw: "v0.1.0-rc", wantErr: true},
		{name: "pkg/schema namespaced", raw: "pkg/schema/v1.2.3", wantErr: true},
		{name: "two-component", raw: "v0.1", wantErr: true},
		{name: "trailing junk", raw: "v0.1.0-rc1-foo", wantErr: true},
		{name: "prerelease not rc", raw: "v0.1.0-beta1", wantErr: true},
		{name: "build metadata", raw: "v0.1.0+build", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
		{name: "garbage", raw: "not-a-version", wantErr: true},
		{name: "leading space", raw: " v0.1.0", wantErr: true},
		{name: "uppercase V", raw: "V0.1.0", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := release.NewVersion(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NewVersion(%q): expected error, got %q", tc.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewVersion(%q): unexpected error: %v", tc.raw, err)
			}
			if got != tc.want {
				t.Errorf("NewVersion(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestVersionKindBaseIsRC(t *testing.T) {
	cases := []struct {
		raw      string
		wantKind release.ReleaseKind
		wantBase release.Version
		wantRC   bool
	}{
		{raw: "v0.1.0", wantKind: release.KindFinal, wantBase: "v0.1.0", wantRC: false},
		{raw: "v0.1.0-rc1", wantKind: release.KindRC, wantBase: "v0.1.0", wantRC: true},
		{raw: "v3.2.1-rc9", wantKind: release.KindRC, wantBase: "v3.2.1", wantRC: true},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			v, err := release.NewVersion(tc.raw)
			if err != nil {
				t.Fatalf("NewVersion(%q): %v", tc.raw, err)
			}
			if got := v.Kind(); got != tc.wantKind {
				t.Errorf("Kind() = %q, want %q", got, tc.wantKind)
			}
			if got := v.Base(); got != tc.wantBase {
				t.Errorf("Base() = %q, want %q", got, tc.wantBase)
			}
			if got := v.IsRC(); got != tc.wantRC {
				t.Errorf("IsRC() = %v, want %v", got, tc.wantRC)
			}
		})
	}
}

func TestParseReleaseTitle(t *testing.T) {
	cases := []struct {
		name     string
		title    string
		wantVer  release.Version
		wantKind release.ReleaseKind
		wantErr  bool
	}{
		{name: "rc", title: "release(v0.1.0-rc1): first release candidate", wantVer: "v0.1.0-rc1", wantKind: release.KindRC},
		{name: "final", title: "release(v0.1.0): first stable release", wantVer: "v0.1.0", wantKind: release.KindFinal},
		{name: "final multi-digit", title: "release(v10.20.30): big one", wantVer: "v10.20.30", wantKind: release.KindFinal},

		// must-fail grammar set
		{name: "bad prefix released", title: "released(v0.1.0): typo", wantErr: true},
		{name: "no parens", title: "release v0.1.0: x", wantErr: true},
		{name: "missing v", title: "release(0.1.0): x", wantErr: true},
		{name: "rc without number", title: "release(v0.1.0-rc): x", wantErr: true},
		{name: "missing colon-space separator", title: "release(v0.1.0):x", wantErr: true},
		{name: "missing subject and separator", title: "release(v0.1.0)", wantErr: true},
		{name: "pkg schema in scope", title: "release(pkg/schema/v1.2.3): x", wantErr: true},
		{name: "garbage", title: "chore: bump deps", wantErr: true},
		{name: "empty", title: "", wantErr: true},
		{name: "leading space", title: " release(v0.1.0): x", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, k, err := release.ParseReleaseTitle(tc.title)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseReleaseTitle(%q): expected error, got (%q, %q)", tc.title, v, k)
				}
				if k != release.KindInvalid {
					t.Errorf("ParseReleaseTitle(%q): want KindInvalid on error, got %q", tc.title, k)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseReleaseTitle(%q): unexpected error: %v", tc.title, err)
			}
			if v != tc.wantVer || k != tc.wantKind {
				t.Errorf("ParseReleaseTitle(%q) = (%q, %q), want (%q, %q)", tc.title, v, k, tc.wantVer, tc.wantKind)
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	cases := []struct {
		name     string
		tag      string
		wantVer  release.Version
		wantKind release.ReleaseKind
		wantErr  bool
	}{
		{name: "final", tag: "v0.1.0", wantVer: "v0.1.0", wantKind: release.KindFinal},
		{name: "rc", tag: "v0.1.0-rc1", wantVer: "v0.1.0-rc1", wantKind: release.KindRC},

		// must-fail grammar set — pkg/schema/v* must NEVER parse as a release tag
		{name: "pkg/schema namespaced", tag: "pkg/schema/v1.2.3", wantErr: true},
		{name: "rc without number", tag: "v0.1.0-rc", wantErr: true},
		{name: "missing v", tag: "0.1.0", wantErr: true},
		{name: "garbage", tag: "latest", wantErr: true},
		{name: "empty", tag: "", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, k, err := release.ParseTag(tc.tag)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseTag(%q): expected error, got (%q, %q)", tc.tag, v, k)
				}
				if k != release.KindInvalid {
					t.Errorf("ParseTag(%q): want KindInvalid on error, got %q", tc.tag, k)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTag(%q): unexpected error: %v", tc.tag, err)
			}
			if v != tc.wantVer || k != tc.wantKind {
				t.Errorf("ParseTag(%q) = (%q, %q), want (%q, %q)", tc.tag, v, k, tc.wantVer, tc.wantKind)
			}
		})
	}
}
