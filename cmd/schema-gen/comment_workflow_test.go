package main

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

//go:embed testdata/comment_workflow.yaml
var commentWorkflowFixture []byte

type commentWorkflowStep struct {
	Name string            `yaml:"name"`
	If   string            `yaml:"if"`
	Uses string            `yaml:"uses"`
	Run  string            `yaml:"run"`
	With map[string]string `yaml:"with"`
}

type commentWorkflowJob struct {
	If    string                `yaml:"if"`
	Needs yaml.Node             `yaml:"needs"`
	Steps []commentWorkflowStep `yaml:"steps"`
}

type commentWorkflow struct {
	Name string                        `yaml:"name"`
	On   map[string]yaml.Node          `yaml:"on"`
	Jobs map[string]commentWorkflowJob `yaml:"jobs"`
}

func TestCommentArtifactWorkflow(t *testing.T) {
	var fixture struct {
		ProducerWorkflow     string `yaml:"producer_workflow"`
		ProducerName         string `yaml:"producer_name"`
		CommentWorkflow      string `yaml:"comment_workflow"`
		CommentJob           string `yaml:"comment_job"`
		Activity             string `yaml:"activity"`
		SourceJob            string `yaml:"source_job"`
		UploaderJob          string `yaml:"uploader_job"`
		IntermediateArtifact string `yaml:"intermediate_artifact"`
		FinalArtifact        string `yaml:"final_artifact"`
		Placeholder          string `yaml:"placeholder"`
		MissingFilesBehavior string `yaml:"missing_files_behavior"`
	}
	if err := yaml.Unmarshal(commentWorkflowFixture, &fixture); err != nil {
		t.Fatal(err)
	}
	root, err := findModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	load := func(name string) commentWorkflow {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, ".github/workflows", name))
		if err != nil {
			t.Fatal(err)
		}
		var workflow commentWorkflow
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			t.Fatal(err)
		}
		return workflow
	}
	producer, comment := load(fixture.ProducerWorkflow), load(fixture.CommentWorkflow)
	if producer.Name != fixture.ProducerName {
		t.Fatalf("producer name = %q, want %q", producer.Name, fixture.ProducerName)
	}
	run, ok := comment.On["workflow_run"]
	if !ok || len(comment.On) != 1 {
		t.Fatal("comment workflow must trigger only on workflow_run")
	}
	var trigger struct {
		Workflows []string `yaml:"workflows"`
		Types     []string `yaml:"types"`
	}
	if err := run.Decode(&trigger); err != nil {
		t.Fatal(err)
	}
	if len(trigger.Workflows) != 1 || trigger.Workflows[0] != producer.Name ||
		len(trigger.Types) != 1 || trigger.Types[0] != fixture.Activity {
		t.Fatalf("workflow_run trigger does not identify the producer completion: %+v", trigger)
	}
	if len(comment.Jobs) != 1 {
		t.Fatal("workflow_run must contain only the comment job")
	}
	if _, ok := comment.Jobs[fixture.CommentJob]; !ok {
		t.Fatalf("comment job %q missing", fixture.CommentJob)
	}

	uploader, ok := producer.Jobs[fixture.UploaderJob]
	var needs []string
	if err := uploader.Needs.Decode(&needs); err != nil {
		t.Fatal(err)
	}
	if !ok || len(needs) != 1 || needs[0] != fixture.SourceJob ||
		!strings.Contains(uploader.If, "always()") || !strings.Contains(uploader.If, "github.event_name == 'pull_request'") {
		t.Fatalf("uploader does not run after a skipped or failed source job: %+v", uploader)
	}
	var placeholder, download, upload *commentWorkflowStep
	for i := range uploader.Steps {
		step := &uploader.Steps[i]
		switch {
		case strings.Contains(step.Run, fixture.Placeholder):
			placeholder = step
		case strings.HasPrefix(step.Uses, "actions/download-artifact@"):
			download = step
		case strings.HasPrefix(step.Uses, "actions/upload-artifact@"):
			upload = step
		}
	}
	if placeholder == nil || download == nil || upload == nil {
		t.Fatal("uploader must create a placeholder, collect the intermediate, and upload the result")
	}
	if download.If != "needs."+fixture.SourceJob+".result != 'skipped'" ||
		download.With["name"] != fixture.IntermediateArtifact ||
		upload.If != "" || upload.With["name"] != fixture.FinalArtifact ||
		upload.With["if-no-files-found"] != fixture.MissingFilesBehavior || download.With["path"] != upload.With["path"] {
		t.Fatal("uploader does not preserve the skipped-source fallback and fail-closed handoff")
	}
	temp := t.TempDir()
	cmd := exec.Command("bash", "-c", placeholder.Run)
	cmd.Env = append(os.Environ(), "RUNNER_TEMP="+temp)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("placeholder step: %v\n%s", err, output)
	}
	placeholderPath := filepath.Join(temp, "apidiff-payload", fixture.Placeholder)
	if stat, err := os.Stat(placeholderPath); err != nil || !stat.Mode().IsRegular() {
		t.Fatalf("placeholder %s is not a regular artifact file: %v", placeholderPath, err)
	}

	finalUploads, intermediateUploads := 0, 0
	for jobName, job := range producer.Jobs {
		for _, step := range job.Steps {
			if !strings.HasPrefix(step.Uses, "actions/upload-artifact@") {
				continue
			}
			switch step.With["name"] {
			case fixture.FinalArtifact:
				finalUploads++
				if jobName != fixture.UploaderJob {
					t.Errorf("final artifact is uploaded by %q", jobName)
				}
			case fixture.IntermediateArtifact:
				intermediateUploads++
				if jobName != fixture.SourceJob || !strings.Contains(step.If, "always()") {
					t.Errorf("intermediate artifact is not produced by the always-run source step")
				}
			}
		}
	}
	if finalUploads != 1 || intermediateUploads != 1 {
		t.Fatalf("final uploads = %d, intermediate uploads = %d; want one each", finalUploads, intermediateUploads)
	}
}
