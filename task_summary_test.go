package schema_test

import (
	_ "embed"
	"testing"

	"github.com/peasant-labs/schema"
	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
)

//go:embed testdata/local-api/task_summary_read_files.yaml
var taskSummaryReadFilesCasesYAML []byte

// readFilesFixtureInput isolates the one field under test
// so a missing YAML key decodes to a true Go nil, distinguishing "omitted"
// from "explicit empty array" the same way TaskSummary.Validate does.
type readFilesFixtureInput struct {
	ReadFiles []string `yaml:"readFiles"`
}

// TestTaskSummary_ReadFilesFixtureContract drives every
// ReadFiles case in testdata/local-api/task_summary_read_files.yaml against
// the real TaskSummary.Validate.
func TestTaskSummary_ReadFilesFixtureContract(t *testing.T) {
	corpus, err := testcase.LoadCorpus[readFilesFixtureInput, bool](taskSummaryReadFilesCasesYAML)
	if err != nil {
		t.Fatalf("load task summary read files corpus: %v", err)
	}
	assert.RequireMin(t, corpus, 6)
	assert.RequireValid(t, corpus)

	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			task := schema.NewTaskSummary("session-a", 0)
			task.ReadFiles = c.Input.ReadFiles
			err := task.Validate()
			valid := err == nil
			if valid != c.Expected {
				t.Fatalf("Validate() error=%v (valid=%v), want valid=%v", err, valid, c.Expected)
			}
			if c.Classification == testcase.MustFail {
				requireActionableValidationError(t, err)
			}
		})
	}
}
