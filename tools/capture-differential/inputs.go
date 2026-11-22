package main

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// input is one named document of one input set.
type input struct {
	set, name string
	doc       []byte
}

// captureInput mirrors the input form of testdata/transcript_capture.yaml.
type captureInput struct {
	Detail     *string    `yaml:"detail"`
	Bytes      *string    `yaml:"bytes"`
	Nest       *nestInput `yaml:"nest"`
	Kills      string     `yaml:"kills"`
	Arch64Only bool       `yaml:"arch64Only"`
}

type nestInput struct {
	Depth int    `yaml:"depth"`
	Where string `yaml:"where"`
}

type rootCase struct {
	Name  string       `yaml:"name"`
	Input captureInput `yaml:"input"`
}

type rootFixture struct {
	RecordedIndex struct {
		Cases []rootCase `yaml:"cases"`
	} `yaml:"recordedIndex"`
	Verdicts struct {
		Cases []rootCase `yaml:"cases"`
	} `yaml:"verdicts"`
	Mutants struct {
		Cases []rootCase `yaml:"cases"`
	} `yaml:"mutants"`
}

// document builds the input bytes. The root fixture test in the schema module
// builds them the same way.
func (in captureInput) document() ([]byte, error) {
	switch {
	case in.Detail != nil:
		return []byte(*in.Detail), nil
	case in.Bytes != nil:
		s, err := strconv.Unquote(*in.Bytes)
		if err != nil {
			return nil, fmt.Errorf("bytes %q is not a Go double-quoted literal: %w", *in.Bytes, err)
		}
		return []byte(s), nil
	case in.Nest != nil:
		return nestedDocument(in.Nest.Depth, in.Nest.Where)
	}
	return nil, fmt.Errorf("input has no detail, bytes or nest")
}

// nestedDocument builds a document whose total nesting depth, root included,
// is depth: inside one turn's member, or in a root member outside turns.
func nestedDocument(depth int, where string) ([]byte, error) {
	var b bytes.Buffer
	switch where {
	case "turn":
		n := depth - 3
		b.WriteString(`{"turns":[{"x":`)
		b.WriteString(strings.Repeat("[", n) + strings.Repeat("]", n))
		b.WriteString(`}]}`)
	case "outside":
		n := depth - 1
		b.WriteString(`{"x":`)
		b.WriteString(strings.Repeat("[", n) + strings.Repeat("]", n))
		b.WriteString(`,"turns":[]}`)
	default:
		return nil, fmt.Errorf("nest where %q is not turn or outside", where)
	}
	return b.Bytes(), nil
}

func loadRoot(path string) ([]input, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f rootFixture
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	var out []input
	for arm, cases := range map[string][]rootCase{"recordedIndex": f.RecordedIndex.Cases, "verdicts": f.Verdicts.Cases, "mutants": f.Mutants.Cases} {
		for _, c := range cases {
			doc, err := c.Input.document()
			if err != nil {
				return nil, fmt.Errorf("%s case %s: %w", arm, c.Name, err)
			}
			out = append(out, input{set: "root", name: arm + "/" + c.Name, doc: doc})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out, nil
}

// mutationSeeds mirrors testdata/mutation_seeds.yaml.
type mutationSeeds struct {
	Mutators []string `yaml:"mutators"`
	Inputs   int      `yaml:"inputs"`
	Seeds    []struct {
		Name   string `yaml:"name"`
		Detail string `yaml:"detail"`
	} `yaml:"seeds"`
}

// mutators is the closed mutator list.
var mutators = map[string]func(r *rand.Rand, doc []byte) []byte{
	"byte-flip":        mutateByteFlip,
	"byte-delete":      mutateByteDelete,
	"truncate":         mutateTruncate,
	"duplicate-member": mutateDuplicateMember,
	"unicode-escape":   mutateUnicodeEscape,
	"insert-multibyte": mutateInsertMultibyte,
	"nest-in-array":    mutateNestInArray,
}

const (
	maxSeeds          = 64
	maxMutatedInputs  = 20000
	mutationSeedValue = 0x5eed
)

func loadMutations(path string) ([]input, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f mutationSeeds
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(f.Seeds) == 0 || len(f.Seeds) > maxSeeds || f.Inputs <= 0 || f.Inputs > maxMutatedInputs {
		return nil, fmt.Errorf("%s must hold 1..%d seeds and 1..%d inputs, has %d seeds and %d inputs", path, maxSeeds, maxMutatedInputs, len(f.Seeds), f.Inputs)
	}
	for _, m := range f.Mutators {
		if mutators[m] == nil {
			return nil, fmt.Errorf("%s names mutator %q outside the closed list", path, m)
		}
	}
	if len(f.Mutators) != len(mutators) {
		return nil, fmt.Errorf("%s must name every mutator of the closed list exactly once", path)
	}
	if err := checkManifest(strings.TrimSuffix(path, ".yaml")+"_manifest.yaml", f.Seeds); err != nil {
		return nil, err
	}
	per := f.Inputs / (len(f.Seeds) * len(f.Mutators))
	var out []input
	for si, seed := range f.Seeds {
		out = append(out, input{set: "mutation", name: seed.Name + "/seed", doc: []byte(seed.Detail)})
		for mi, name := range f.Mutators {
			r := rand.New(rand.NewPCG(mutationSeedValue, uint64(si)<<16|uint64(mi)))
			for k := 0; k < per; k++ {
				doc := mutators[name](r, []byte(seed.Detail))
				out = append(out, input{set: "mutation", name: fmt.Sprintf("%s/%s/%d", seed.Name, name, k), doc: doc})
			}
		}
	}
	return out, nil
}

var interestingBytes = []byte{'"', '\\', '{', '}', '[', ']', ',', ':', 'u', '0', ' ', 'n', 0xff, 0xc3, 0x01}

func mutateByteFlip(r *rand.Rand, doc []byte) []byte {
	if len(doc) == 0 {
		return doc
	}
	doc[r.IntN(len(doc))] = interestingBytes[r.IntN(len(interestingBytes))]
	return doc
}

func mutateByteDelete(r *rand.Rand, doc []byte) []byte {
	if len(doc) == 0 {
		return doc
	}
	i := r.IntN(len(doc))
	return append(doc[:i:i], doc[i+1:]...)
}

func mutateTruncate(r *rand.Rand, doc []byte) []byte {
	if len(doc) == 0 {
		return doc
	}
	return doc[:r.IntN(len(doc))]
}

// skipValue returns the end of the JSON-like value at i, tolerating junk.
func skipValue(doc []byte, i int) int {
	depth, inString := 0, false
	for ; i < len(doc); i++ {
		c := doc[i]
		if inString {
			if c == '\\' {
				i++
			} else if c == '"' {
				inString = false
				if depth == 0 {
					return i + 1
				}
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth <= 0 {
				return i + 1
			}
		case ',':
			if depth == 0 {
				return i
			}
		}
	}
	return len(doc)
}

// memberStarts lists offsets of quotes that follow '{' or ','.
func memberStarts(doc []byte) []int {
	var out []int
	for i := 1; i < len(doc); i++ {
		if doc[i] == '"' && (doc[i-1] == '{' || doc[i-1] == ',') {
			out = append(out, i)
		}
	}
	return out
}

func mutateDuplicateMember(r *rand.Rand, doc []byte) []byte {
	starts := memberStarts(doc)
	if len(starts) == 0 {
		return doc
	}
	s := starts[r.IntN(len(starts))]
	nameEnd := skipValue(doc, s)
	if nameEnd >= len(doc) || doc[nameEnd] != ':' {
		return doc
	}
	end := skipValue(doc, nameEnd+1)
	member := append([]byte{','}, doc[s:end]...)
	out := append(append(append([]byte{}, doc[:end]...), member...), doc[end:]...)
	return out
}

func mutateUnicodeEscape(r *rand.Rand, doc []byte) []byte {
	var letters []int
	for i, c := range doc {
		if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') {
			letters = append(letters, i)
		}
	}
	if len(letters) == 0 {
		return doc
	}
	i := letters[r.IntN(len(letters))]
	esc := fmt.Sprintf(`\u%04x`, doc[i])
	if r.IntN(2) == 0 {
		esc = fmt.Sprintf(`\u%04X`, doc[i])
	}
	return append(append(append([]byte{}, doc[:i]...), esc...), doc[i+1:]...)
}

var multibyteRunes = []string{"é", "中", "😀", "\u2028", "\ufffd"}

func mutateInsertMultibyte(r *rand.Rand, doc []byte) []byte {
	i := r.IntN(len(doc) + 1)
	m := multibyteRunes[r.IntN(len(multibyteRunes))]
	return append(append(append([]byte{}, doc[:i]...), m...), doc[i:]...)
}

func mutateNestInArray(r *rand.Rand, doc []byte) []byte {
	var opens []int
	for i, c := range doc {
		if c == '{' || c == '[' {
			opens = append(opens, i)
		}
	}
	if len(opens) == 0 {
		return doc
	}
	i := opens[r.IntN(len(opens))]
	end := skipValue(doc, i)
	out := append(append([]byte{}, doc[:i]...), '[')
	out = append(out, doc[i:end]...)
	out = append(out, ']')
	return append(out, doc[end:]...)
}

// subjects mirrors testdata/subjects.yaml, the benchmark subjects.
type subjects struct {
	TurnCounts []int `yaml:"turnCounts"`
	Subjects   []struct {
		Name string `yaml:"name"`
		Text string `yaml:"text"`
	} `yaml:"subjects"`
}

func loadSubjectsFile(path string) (subjects, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return subjects{}, err
	}
	var f subjects
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return subjects{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return f, nil
}

// subjectDocument builds Village's benchmark turn shape inside {"turns":[...]}.
// text is inserted verbatim as JSON string content.
func subjectDocument(turns int, text string) []byte {
	var b bytes.Buffer
	b.WriteString(`{"turns":[`)
	for i := 0; i < turns; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"index":%d,"role":"user","content":"prompt %d%s","timestamp":"2026-01-01T00:00:00Z","depth":0}`, i, i, text)
	}
	b.WriteString(`]}`)
	return b.Bytes()
}

func loadSubjects(path string) ([]input, error) {
	f, err := loadSubjectsFile(path)
	if err != nil {
		return nil, err
	}
	var out []input
	for _, s := range f.Subjects {
		for _, n := range f.TurnCounts {
			out = append(out, input{set: "subject", name: fmt.Sprintf("%s/%d", s.Name, n), doc: subjectDocument(n, s.Text)})
		}
	}
	return out, nil
}

// loadFuzz reads Go's fuzz corpus files for a one-[]byte-argument target.
func loadFuzz(dirs ...string) ([]input, error) {
	var out []input
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, err
			}
			lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
			if len(lines) != 2 || lines[0] != "go test fuzz v1" || !strings.HasPrefix(lines[1], "[]byte(") || !strings.HasSuffix(lines[1], ")") {
				return nil, fmt.Errorf("fuzz file %s is not a one-[]byte corpus entry", e.Name())
			}
			doc, err := strconv.Unquote(strings.TrimSuffix(strings.TrimPrefix(lines[1], "[]byte("), ")"))
			if err != nil {
				return nil, fmt.Errorf("fuzz file %s: %w", e.Name(), err)
			}
			out = append(out, input{set: "fuzz", name: e.Name(), doc: []byte(doc)})
		}
	}
	return out, nil
}

// checkManifest requires the seed names to equal the manifest's required names.
func checkManifest(path string, seeds []struct {
	Name   string `yaml:"name"`
	Detail string `yaml:"detail"`
}) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m struct {
		RequiredNames []string `yaml:"requiredNames"`
	}
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	want := map[string]bool{}
	for _, n := range m.RequiredNames {
		want[n] = true
	}
	got := map[string]bool{}
	for _, s := range seeds {
		got[s.Name] = true
		if !want[s.Name] {
			return fmt.Errorf("seed %q is missing from %s", s.Name, path)
		}
	}
	for n := range want {
		if !got[n] {
			return fmt.Errorf("required seed %q from %s is absent", n, path)
		}
	}
	return nil
}
