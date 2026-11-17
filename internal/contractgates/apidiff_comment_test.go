package contractgates

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/apidiff_comment.yaml
var commentFixtures []byte

func TestReviewedAPIDiffComment(t *testing.T) {
	skipIfMissing(t, "node")
	var fixtures struct {
		Required []string `yaml:"required"`
		Cases    []struct {
			Name           string   `yaml:"name"`
			ChangesPresent bool     `yaml:"changes_present"`
			Changes        string   `yaml:"changes"`
			Compatible     string   `yaml:"compatible"`
			Previous       bool     `yaml:"previous"`
			ExpectedMethod string   `yaml:"expected_method"`
			Contains       []string `yaml:"contains"`
			Excludes       []string `yaml:"excludes"`
		} `yaml:"cases"`
	}
	if err := yaml.Unmarshal(commentFixtures, &fixtures); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, c := range fixtures.Cases {
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			if c.ChangesPresent {
				if err := os.WriteFile(filepath.Join(dir, "apidiff-changes.txt"), []byte(c.Changes), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(dir, "apidiff-compatible.txt"), []byte(c.Compatible), 0o600); err != nil {
				t.Fatal(err)
			}
			var method, body string
			var mu sync.Mutex
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-token" {
					http.Error(w, "missing token", http.StatusUnauthorized)
					return
				}
				switch r.Method {
				case http.MethodGet:
					if r.URL.Path != "/repos/peasant-labs/schema/issues/42/comments" {
						t.Errorf("unexpected list path %s", r.URL.Path)
					}
					if c.Previous {
						fmt.Fprint(w, `[{"id":7,"body":"<!-- go-apidiff-gate -->\nprevious"}]`)
					} else {
						fmt.Fprint(w, `[]`)
					}
				case http.MethodPost, http.MethodPatch:
					mu.Lock()
					defer mu.Unlock()
					method = r.Method
					wantPath := "/repos/peasant-labs/schema/issues/42/comments"
					if r.Method == http.MethodPatch {
						wantPath = "/repos/peasant-labs/schema/issues/comments/7"
					}
					if r.URL.Path != wantPath {
						t.Errorf("write path = %s, want %s", r.URL.Path, wantPath)
					}
					var input struct {
						Body string `json:"body"`
					}
					if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
						t.Error(err)
					}
					body = input.Body
					fmt.Fprint(w, `{}`)
				default:
					http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				}
			}))
			defer server.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "node", filepath.Join(moduleRoot(t), "scripts/post-go-apidiff-comment.mjs"))
			cmd.Env = append(os.Environ(), "GITHUB_REPOSITORY=peasant-labs/schema", "PR_NUMBER=42",
				"GH_TOKEN=test-token", "EVIDENCE_DIR="+dir, "GITHUB_API_URL="+server.URL)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("comment script: %v\n%s", err, output)
			}
			mu.Lock()
			defer mu.Unlock()
			if method == "" {
				method = "no-write"
			}
			if method != c.ExpectedMethod {
				t.Fatalf("method = %s, want %s", method, c.ExpectedMethod)
			}
			if method != "no-write" && !strings.HasPrefix(body, "<!-- go-apidiff-gate -->") {
				t.Errorf("comment lost its sticky marker: %s", body)
			}
			for _, fragment := range c.Contains {
				if !strings.Contains(body, fragment) {
					t.Errorf("comment missing %q:\n%s", fragment, body)
				}
			}
			for _, fragment := range c.Excludes {
				if strings.Contains(body, fragment) {
					t.Errorf("comment unexpectedly contains %q:\n%s", fragment, body)
				}
			}
		})
	}
	for _, name := range fixtures.Required {
		if !seen[name] {
			t.Errorf("required comment case %q missing", name)
		}
	}
}
