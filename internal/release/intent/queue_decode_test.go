package intent_test

// Queue-response decoder suite: drives intent.DecodeQueueResponse over the
// decode corpus. Every failure must be a plain six-part error with no gate
// value, never a typed rejection and never an empty queue.

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/peasant-labs/schema/internal/release/intent"
)

//go:embed testdata/queue_decode_corpus.yaml
var queueDecodeCorpusYAML []byte

type queueDecodeInput struct {
	Document string `yaml:"document"`
}

type queueDecodeExpected struct {
	Entries []intent.QueueEntry `yaml:"entries"`
}

func TestDecodeQueueResponse(t *testing.T) {
	corpus := loadNamedCorpus[queueDecodeInput, queueDecodeExpected](t, queueDecodeCorpusYAML)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := intent.DecodeQueueResponse([]byte(c.Input.Document))
			if c.Classification == "must-fail" {
				if err == nil {
					t.Fatalf("DecodeQueueResponse: expected an error, got %+v", got)
				}
				requireSixLabels(t, err)
				return
			}
			if err != nil {
				t.Fatalf("DecodeQueueResponse: unexpected error: %v", err)
			}
			if !reflect.DeepEqual(append([]intent.QueueEntry(nil), got...), append([]intent.QueueEntry(nil), c.Expected.Entries...)) {
				t.Errorf("Entries = %+v, want %+v", got, c.Expected.Entries)
			}
		})
	}
}
