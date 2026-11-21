package main

import (
	"fmt"
	"runtime"
	"sort"
	"testing"
	"time"

	schema "github.com/peasant-labs/schema"
)

// benchTurns is the turn count of both benchmark subjects.
const benchTurns = 100000

// partitionPass, in each adapter, is the schema work of Village's partition
// build in its order: capture, cook, the per-kept-turn facts, then the cut.

// retainedBytes is the live heap the captured turns keep, with the input held.
func retainedBytes(doc []byte) (uint64, error) {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&before)
	turns, err := schema.NewTranscriptCookTurns(doc)
	if err != nil {
		return 0, err
	}
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(turns)
	runtime.KeepAlive(doc)
	return after.HeapInuse - before.HeapInuse, nil
}

// BenchmarkPartitionPass is evidence only: it reports timing and memory and
// asserts nothing about them.
func BenchmarkPartitionPass(b *testing.B) {
	f, err := loadSubjectsFile("testdata/subjects.yaml")
	if err != nil {
		b.Fatal(err)
	}
	for _, s := range f.Subjects {
		doc := subjectDocument(benchTurns, s.Text)
		b.Run(s.Name, func(b *testing.B) {
			retained, err := retainedBytes(doc)
			if err != nil {
				b.Fatal(err)
			}
			for i := 0; i < 5; i++ {
				runtime.GC()
				if _, err := partitionPass(doc); err != nil {
					b.Fatal(err)
				}
			}
			samples := make([]time.Duration, 0, b.N)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				runtime.GC()
				b.StartTimer()
				started := time.Now()
				if _, err := partitionPass(doc); err != nil {
					b.Fatal(err)
				}
				samples = append(samples, time.Since(started))
			}
			b.StopTimer()
			sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
			p95 := samples[(len(samples)*95+99)/100-1]
			b.ReportMetric(float64(p95.Nanoseconds()), "p95-ns/op")
			b.ReportMetric(float64(retained), "retained-B")
			b.Logf("%s build: %d turns, %d input bytes", adapterName, benchTurns, len(doc))
		})
	}
}

func TestBenchmarkSubjectsCapture(t *testing.T) {
	f, err := loadSubjectsFile("testdata/subjects.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range f.Subjects {
		if _, err := partitionPass(subjectDocument(10, s.Text)); err != nil {
			t.Fatal(fmt.Errorf("subject %s: %w", s.Name, err))
		}
	}
}
