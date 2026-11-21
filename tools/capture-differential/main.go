// Command capture-differential compares schema.NewTranscriptCookTurns across
// two builds: the pinned schema release (build tag capture_baseline) and the
// working tree (through a go.work). Each build prints one verdict line per
// input; compare fails on the first differing line. See the Makefile targets
// capture-differential and capture-bench. It is local evidence, not a CI gate.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "run":
		err = runCommand(os.Args[2:])
	case "compare":
		err = compareCommand(os.Args[2:])
	case "record":
		err = recordCommand(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "capture-differential (%s build): %v\n", adapterName, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: capture-differential run -root FILE -seeds FILE -subjects FILE -fuzz DIR[,DIR] | compare BASE CAND | record -root FILE")
	os.Exit(2)
}

func runCommand(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	root := fs.String("root", "", "root capture fixture")
	seeds := fs.String("seeds", "", "mutation seed fixture")
	subjectsPath := fs.String("subjects", "", "benchmark subject fixture")
	fuzzDirs := fs.String("fuzz", "", "comma-separated fuzz corpus directories")
	_ = fs.Parse(args)
	type loaded struct {
		set    string
		inputs []input
	}
	var sets []loaded
	for _, step := range []struct {
		set  string
		load func() ([]input, error)
	}{
		{"root", func() ([]input, error) { return loadRoot(*root) }},
		{"mutation", func() ([]input, error) { return loadMutations(*seeds) }},
		{"subject", func() ([]input, error) { return loadSubjects(*subjectsPath) }},
		{"fuzz", func() ([]input, error) { return loadFuzz(strings.Split(*fuzzDirs, ",")...) }},
	} {
		in, err := step.load()
		if err != nil {
			return fmt.Errorf("load %s inputs: %w", step.set, err)
		}
		fmt.Fprintf(os.Stderr, "%s build: %s set: %d inputs\n", adapterName, step.set, len(in))
		sets = append(sets, loaded{step.set, in})
	}
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, s := range sets {
		for _, in := range s.inputs {
			o, err := observe(in.doc)
			if err != nil {
				return fmt.Errorf("%s %s: %w", in.set, in.name, err)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", in.set, in.name, mustJSON(o))
		}
	}
	return nil
}

func compareCommand(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("compare needs the baseline and candidate line files")
	}
	read := func(path string) ([]string, error) {
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n"), nil
	}
	base, err := read(args[0])
	if err != nil {
		return err
	}
	cand, err := read(args[1])
	if err != nil {
		return err
	}
	if len(base) != len(cand) {
		return fmt.Errorf("baseline printed %d lines and candidate %d; the input sets differ", len(base), len(cand))
	}
	counts := map[string]map[string]int{}
	for i := range base {
		if base[i] != cand[i] {
			fields := strings.SplitN(base[i], "\t", 3)
			fmt.Fprintf(os.Stderr, "DIFFERENCE at line %d, case %s\n  baseline:  %s\n  candidate: %s\n", i+1, strings.Join(fields[:2], " "), base[i], cand[i])
			return fmt.Errorf("verdicts differ")
		}
		fields := strings.SplitN(base[i], "\t", 3)
		var o observation
		if err := json.Unmarshal([]byte(fields[2]), &o); err != nil {
			return fmt.Errorf("line %d: %w", i+1, err)
		}
		verdict := "accept"
		if o.Refusal != "" {
			verdict = "refuse " + o.Refusal
		}
		if counts[fields[0]] == nil {
			counts[fields[0]] = map[string]int{}
		}
		counts[fields[0]][verdict]++
	}
	sets := make([]string, 0, len(counts))
	for s := range counts {
		sets = append(sets, s)
	}
	sort.Strings(sets)
	fmt.Printf("identical: %d lines\n", len(base))
	for _, s := range sets {
		verdicts := make([]string, 0, len(counts[s]))
		total := 0
		for v, n := range counts[s] {
			verdicts = append(verdicts, fmt.Sprintf("%s=%d", v, n))
			total += n
		}
		sort.Strings(verdicts)
		fmt.Printf("%s (%d): %s\n", s, total, strings.Join(verdicts, ", "))
	}
	return nil
}
