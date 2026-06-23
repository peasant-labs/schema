package release

import "testing"

func TestRunGreenForCommit(t *testing.T) {
	t.Parallel()

	const commit = "abc123"

	tests := []struct {
		name string
		runs []WorkflowRun
		want bool
	}{
		{
			name: "completed success on the commit is green",
			runs: []WorkflowRun{
				{HeadSHA: commit, Status: WorkflowRunCompleted, Conclusion: WorkflowRunSuccess},
			},
			want: true,
		},
		{
			name: "green run found among decoys",
			runs: []WorkflowRun{
				{HeadSHA: "other", Status: WorkflowRunCompleted, Conclusion: WorkflowRunSuccess},
				{HeadSHA: commit, Status: WorkflowRunInProgress, Conclusion: WorkflowRunNoConclusion},
				{HeadSHA: commit, Status: WorkflowRunCompleted, Conclusion: WorkflowRunFailure},
				{HeadSHA: commit, Status: WorkflowRunCompleted, Conclusion: WorkflowRunSuccess},
			},
			want: true,
		},
		{
			name: "success on a different commit is not green",
			runs: []WorkflowRun{
				{HeadSHA: "other", Status: WorkflowRunCompleted, Conclusion: WorkflowRunSuccess},
			},
			want: false,
		},
		{
			name: "completed failure on the commit is not green",
			runs: []WorkflowRun{
				{HeadSHA: commit, Status: WorkflowRunCompleted, Conclusion: WorkflowRunFailure},
			},
			want: false,
		},
		{
			name: "in-progress on the commit is not green",
			runs: []WorkflowRun{
				{HeadSHA: commit, Status: WorkflowRunInProgress, Conclusion: WorkflowRunNoConclusion},
			},
			want: false,
		},
		{
			name: "success but not completed is not green",
			runs: []WorkflowRun{
				{HeadSHA: commit, Status: WorkflowRunQueued, Conclusion: WorkflowRunSuccess},
			},
			want: false,
		},
		{
			name: "no runs is not green",
			runs: nil,
			want: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := RunGreenForCommit(tc.runs, commit); got != tc.want {
				t.Fatalf("RunGreenForCommit(%+v, %q) = %v, want %v", tc.runs, commit, got, tc.want)
			}
		})
	}
}
