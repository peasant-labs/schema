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
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/apidiff_comment.yaml
var commentFixtures []byte

type commentPRMode string

const (
	commentPRDefault            commentPRMode = ""
	commentPRNone               commentPRMode = "none"
	commentPRClosed             commentPRMode = "closed"
	commentPROtherRepository    commentPRMode = "other-repository"
	commentPROtherBranch        commentPRMode = "other-branch"
	commentPRStale              commentPRMode = "stale"
	commentPRAdvances           commentPRMode = "advances"
	commentPRAmbiguous          commentPRMode = "ambiguous"
	commentPRPaginatedMatch     commentPRMode = "paginated-match"
	commentPRPaginatedAmbiguity commentPRMode = "paginated-ambiguity"
)

type commentRepository struct {
	FullName string `json:"full_name"`
}

type commentHead struct {
	SHA  string            `json:"sha"`
	Ref  string            `json:"ref"`
	Repo commentRepository `json:"repo"`
}

type commentPull struct {
	Number int         `json:"number"`
	State  string      `json:"state"`
	Head   commentHead `json:"head"`
}

type commentWorkflowRun struct {
	HeadSHA        string            `json:"head_sha"`
	HeadBranch     string            `json:"head_branch"`
	HeadRepository commentRepository `json:"head_repository"`
}

type commentEvent struct {
	WorkflowRun commentWorkflowRun `json:"workflow_run"`
}

func TestReviewedAPIDiffComment(t *testing.T) {
	skipIfMissing(t, "node")
	var fixtures struct {
		Required []string `yaml:"required"`
		Cases    []struct {
			Name             string        `yaml:"name"`
			ChangesPresent   bool          `yaml:"changes_present"`
			Changes          string        `yaml:"changes"`
			Compatible       string        `yaml:"compatible"`
			Previous         bool          `yaml:"previous"`
			MissingDirectory bool          `yaml:"missing_directory"`
			ArtifactNumber   string        `yaml:"artifact_number"`
			PRMode           commentPRMode `yaml:"pr_mode"`
			Repeat           int           `yaml:"repeat"`
			ExpectedExit     int           `yaml:"expected_exit"`
			ExpectedMethod   string        `yaml:"expected_method"`
			ExpectedPages    int           `yaml:"expected_pages"`
			Contains         []string      `yaml:"contains"`
			Excludes         []string      `yaml:"excludes"`
			MaxBodyBytes     int           `yaml:"max_body_bytes"`
		} `yaml:"cases"`
	}
	if err := yaml.Unmarshal(commentFixtures, &fixtures); err != nil {
		t.Fatal(err)
	}
	const sha = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const newerSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	const headRepo = "contributor/schema"
	const headBranch = "change"
	pull := func(number int, state, repo, branch, hash string) commentPull {
		return commentPull{Number: number, State: state, Head: commentHead{
			SHA: hash, Ref: branch, Repo: commentRepository{FullName: repo},
		}}
	}
	seen := make(map[string]bool)
	for _, c := range fixtures.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatalf("empty or duplicate comment case %q", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			dir := t.TempDir()
			evidence := filepath.Join(dir, "evidence")
			if !c.MissingDirectory {
				if err := os.Mkdir(evidence, 0o700); err != nil {
					t.Fatal(err)
				}
				if c.ChangesPresent {
					changes := c.Changes
					if c.Repeat > 0 {
						changes = strings.Repeat(changes, c.Repeat)
					}
					if err := os.WriteFile(filepath.Join(evidence, "apidiff-changes.txt"), []byte(changes), 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.WriteFile(filepath.Join(evidence, "apidiff-compatible.txt"), []byte(c.Compatible), 0o600); err != nil {
					t.Fatal(err)
				}
				if c.ArtifactNumber != "" {
					if err := os.WriteFile(filepath.Join(evidence, "pr-number.txt"), []byte(c.ArtifactNumber), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			event := commentEvent{WorkflowRun: commentWorkflowRun{
				HeadSHA: sha, HeadBranch: headBranch, HeadRepository: commentRepository{FullName: headRepo},
			}}
			eventData, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			eventPath := filepath.Join(dir, "event.json")
			if err := os.WriteFile(eventPath, eventData, 0o600); err != nil {
				t.Fatal(err)
			}
			var method, body string
			var pages int
			var mu sync.Mutex
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-token" {
					http.Error(w, "missing token", http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				mu.Lock()
				defer mu.Unlock()
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/repos/peasant-labs/schema/commits/"+sha+"/pulls":
					pages++
					page, _ := strconv.Atoi(r.URL.Query().Get("page"))
					main := pull(42, "open", headRepo, headBranch, sha)
					pulls := []commentPull{main}
					switch c.PRMode {
					case commentPRDefault, commentPRAdvances:
					case commentPRNone:
						pulls = nil
					case commentPRClosed:
						pulls = []commentPull{pull(42, "closed", headRepo, headBranch, sha)}
					case commentPROtherRepository:
						pulls = []commentPull{pull(42, "open", "other/schema", headBranch, sha)}
					case commentPROtherBranch:
						pulls = []commentPull{pull(42, "open", headRepo, "other", sha)}
					case commentPRStale:
						pulls = []commentPull{pull(42, "open", headRepo, headBranch, newerSHA)}
					case commentPRAmbiguous:
						pulls = append(pulls, pull(43, "open", headRepo, headBranch, sha))
					case commentPRPaginatedMatch, commentPRPaginatedAmbiguity:
						pulls = []commentPull{}
						if page == 1 {
							if c.PRMode == commentPRPaginatedAmbiguity {
								pulls = append(pulls, main)
							}
							for len(pulls) < 100 {
								pulls = append(pulls, pull(500+len(pulls), "closed", headRepo, headBranch, sha))
							}
						} else if c.PRMode == commentPRPaginatedAmbiguity {
							pulls = append(pulls, pull(43, "open", headRepo, headBranch, sha))
						} else {
							pulls = append(pulls, main)
						}
					default:
						t.Errorf("unknown PR fixture mode %q", c.PRMode)
					}
					if pulls == nil {
						pulls = []commentPull{}
					}
					if err := json.NewEncoder(w).Encode(pulls); err != nil {
						t.Error(err)
					}
				case r.Method == http.MethodGet && r.URL.Path == "/repos/peasant-labs/schema/pulls/42":
					hash := sha
					if c.PRMode == commentPRAdvances {
						hash = newerSHA
					}
					if err := json.NewEncoder(w).Encode(pull(42, "open", headRepo, headBranch, hash)); err != nil {
						t.Error(err)
					}
				case r.Method == http.MethodGet && r.URL.Path == "/repos/peasant-labs/schema/issues/42/comments":
					if c.Previous {
						fmt.Fprint(w, `[{"id":7,"body":"<!-- go-apidiff-gate -->\nprevious"}]`)
					} else {
						fmt.Fprint(w, `[]`)
					}
				case r.Method == http.MethodPost || r.Method == http.MethodPatch:
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
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					http.Error(w, "unexpected request", http.StatusNotFound)
				}
			}))
			defer server.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "node", filepath.Join(moduleRoot(t), "scripts/post-go-apidiff-comment.mjs"))
			cmd.Env = append(os.Environ(), "GITHUB_REPOSITORY=peasant-labs/schema", "GH_TOKEN=test-token",
				"EVIDENCE_DIR="+evidence, "GITHUB_EVENT_PATH="+eventPath, "GITHUB_API_URL="+server.URL)
			output, err := cmd.CombinedOutput()
			exit := 0
			if err != nil {
				if e, ok := err.(*exec.ExitError); ok {
					exit = e.ExitCode()
				} else {
					t.Fatalf("comment script: %v\n%s", err, output)
				}
			}
			if exit != c.ExpectedExit {
				t.Fatalf("exit = %d, want %d\n%s", exit, c.ExpectedExit, output)
			}
			mu.Lock()
			defer mu.Unlock()
			if method == "" {
				method = "no-write"
			}
			if method != c.ExpectedMethod {
				t.Fatalf("method = %s, want %s\n%s", method, c.ExpectedMethod, output)
			}
			if c.ExpectedPages > 0 && pages != c.ExpectedPages {
				t.Errorf("pull request pages = %d, want %d", pages, c.ExpectedPages)
			}
			if method != "no-write" && !strings.HasPrefix(body, "<!-- go-apidiff-gate -->") {
				t.Errorf("comment lost its sticky marker: %s", body)
			}
			if c.MaxBodyBytes > 0 && len(body) > c.MaxBodyBytes {
				t.Errorf("comment has %d bytes, limit %d", len(body), c.MaxBodyBytes)
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
