package contractgates

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/aggregate.yaml
var aggregateFixtures []byte

func TestAggregateKeepsIndependentAPIDiff(t *testing.T) {
	skipIfMissing(t, "bash")
	skipIfMissing(t, "git")
	var fixture struct {
		Required []string `yaml:"required"`
		Cases    []struct {
			Name         string `yaml:"name"`
			VacuumExit   int    `yaml:"vacuum_exit"`
			APIExit      int    `yaml:"api_exit"`
			APIOutput    string `yaml:"api_output"`
			ReportExit   int    `yaml:"report_exit"`
			ExpectedExit int    `yaml:"expected_exit"`
			Signal       bool   `yaml:"signal"`
			Report       bool   `yaml:"report"`
		} `yaml:"cases"`
	}
	if err := yaml.Unmarshal(aggregateFixtures, &fixture); err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile(filepath.Join(moduleRoot(t), "scripts/contract-gates.sh"))
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("empty or duplicate aggregate case name %q", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			write := func(path, content string, mode os.FileMode) {
				t.Helper()
				path = filepath.Join(dir, path)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), mode); err != nil {
					t.Fatal(err)
				}
			}
			write("scripts/contract-gates.sh", string(script), 0o755)
			write("generated/types-0.1.0.yaml", "fixture\n", 0o644)
			write("bin/vacuum", "#!/bin/sh\nexit \"$FIXTURE_VACUUM_EXIT\"\n", 0o755)
			write("bin/oasdiff", "#!/bin/sh\nexit 0\n", 0o755)
			write("bin/go-apidiff", "#!/bin/sh\nprintf '%s\\n' \"$FIXTURE_API_OUTPUT\"\nexit \"$FIXTURE_API_EXIT\"\n", 0o755)
			write("bin/go", `#!/bin/sh
if [ "$FIXTURE_REPORT_EXIT" -ne 0 ]; then exit "$FIXTURE_REPORT_EXIT"; fi
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--write" ]; then printf 'report\n' > "$2"; exit 0; fi
  shift
done
exit 1
`, 0o755)
			git := func(args ...string) {
				t.Helper()
				cmd := exec.Command("git", args...)
				cmd.Dir = dir
				cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid",
					"GIT_COMMITTER_NAME=fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, out)
				}
			}
			git("init", "-q")
			git("add", "scripts", "generated")
			git("-c", "commit.gpgsign=false", "commit", "-qm", "fixture")
			outDir := t.TempDir()
			signal := filepath.Join(outDir, "changes.txt")
			report := filepath.Join(outDir, "report.yaml")
			if err := os.WriteFile(signal, []byte("stale"), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("bash", filepath.Join(dir, "scripts/contract-gates.sh"), "all", "HEAD", "--out", report)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "PATH="+filepath.Join(dir, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
				"APIDIFF_CHANGES_FILE="+signal, "FIXTURE_VACUUM_EXIT="+strconv.Itoa(c.VacuumExit),
				"FIXTURE_API_EXIT="+strconv.Itoa(c.APIExit), "FIXTURE_API_OUTPUT="+c.APIOutput,
				"FIXTURE_REPORT_EXIT="+strconv.Itoa(c.ReportExit))
			output, err := cmd.CombinedOutput()
			exit := 0
			if err != nil {
				if e, ok := err.(*exec.ExitError); ok {
					exit = e.ExitCode()
				} else {
					t.Fatalf("run aggregate gate: %v", err)
				}
			}
			if exit != c.ExpectedExit {
				t.Errorf("exit = %d, want %d\n%s", exit, c.ExpectedExit, output)
			}
			_, signalErr := os.Stat(signal)
			if (signalErr == nil) != c.Signal {
				t.Errorf("signal exists = %t, want %t\n%s", signalErr == nil, c.Signal, output)
			}
			if c.Signal {
				data, err := os.ReadFile(signal)
				if err != nil || len(data) != 0 {
					t.Errorf("established clean signal = %q, err = %v; want an empty file", data, err)
				}
			}
			_, reportErr := os.Stat(report)
			if (reportErr == nil) != c.Report {
				t.Errorf("report exists = %t, want %t\n%s", reportErr == nil, c.Report, output)
			}
		})
	}
	for _, name := range fixture.Required {
		if !seen[name] {
			t.Errorf("required aggregate case %q missing", name)
		}
	}
}
