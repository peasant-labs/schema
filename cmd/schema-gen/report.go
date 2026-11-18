package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// reportSurface connects the generated documents to the version declaration
// that names their current files. The config is checked against the actual
// generator output by the repository tests.
type reportSurface struct {
	name, label, version string
	stems                []string
}

var reportSurfaces = []reportSurface{
	{name: "types", label: "Types", version: "TypesVersion", stems: []string{"types"}},
	{name: "village-api", label: "Village API", version: "VillageAPIVersion", stems: []string{"village-api", "publish-request", "annotation-push-request"}},
	{name: "local-api", label: "Local API", version: "PeasantLocalAPIVersion", stems: []string{"peasantlocal-api"}},
	{name: "segmented-read", label: "Segmented Read", version: "SegmentedReadAPIVersion", stems: []string{"segmented-read"}},
}

type reportChange string

const (
	reportNone       reportChange = "none"
	reportDocs       reportChange = "docs"
	reportFunctional reportChange = "functional"
)

type surfaceReport struct {
	Name        string       `yaml:"name"`
	Breaking    bool         `yaml:"breaking"`
	Change      reportChange `yaml:"change"`
	AheadOnBase bool         `yaml:"ahead_on_base"`
}

type contractReport struct {
	Report        int             `yaml:"report"`
	Base          string          `yaml:"base"`
	Head          string          `yaml:"head"`
	GoAPIBreaking bool            `yaml:"go_api_breaking"`
	Surfaces      []surfaceReport `yaml:"surfaces"`
}

func runReport(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	base := fs.String("base", "", "base revision")
	head := fs.String("head", "HEAD", "head revision")
	apidiff := fs.String("apidiff", "", "go-apidiff signal file")
	out := fs.String("write", "", "output YAML file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *base == "" || *head == "" || *apidiff == "" || *out == "" || fs.NArg() != 0 {
		return errors.New("schema-gen report requires --base, --head, --apidiff, and --write; supply both revisions and the established API diff signal")
	}
	root, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("find module root for report: %w", err)
	}
	rep, err := buildContractReport(root, *base, *head, *apidiff)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(rep)
	if err != nil {
		return fmt.Errorf("encode contract report: %w", err)
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		return fmt.Errorf("write contract report %s: %w", *out, err)
	}
	return nil
}

func gitOutput(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s in %s failed: %w", strings.Join(args, " "), root, err)
	}
	return out, nil
}

func reportRevision(root, ref string) (string, error) {
	out, err := gitOutput(root, "rev-parse", "--verify", ref+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("resolve report revision %q: %w; fetch the revision and retry", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func versionsAt(root, rev string) (map[string]string, error) {
	data, err := gitOutput(root, "show", rev+":versions.go")
	if err != nil {
		return nil, fmt.Errorf("read versions.go at %s: %w", rev, err)
	}
	return parseReportVersions(data, rev)
}

func parseReportVersions(data []byte, location string) (map[string]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "versions.go", data, 0)
	if err != nil {
		return nil, fmt.Errorf("parse versions.go at %s: %w", location, err)
	}
	values := make(map[string]string)
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value := spec.(*ast.ValueSpec)
			if len(value.Names) != 1 || len(value.Values) != 1 {
				continue
			}
			literal, ok := value.Values[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				continue
			}
			if s, err := strconv.Unquote(literal.Value); err == nil {
				values[value.Names[0].Name] = s
			}
		}
	}
	for _, surface := range reportSurfaces {
		if values[surface.version] == "" {
			return nil, fmt.Errorf("versions.go at %s has no literal %s; cannot resolve the %s contract document", location, surface.version, surface.name)
		}
	}
	return values, nil
}

func reportDocumentAt(root, rev, stem, version, extension string) ([]byte, string, bool, error) {
	workingPath := "generated/" + stem + "-next" + extension
	listed, err := gitOutput(root, "ls-tree", "--name-only", rev, "--", workingPath)
	if err != nil {
		return nil, "", false, fmt.Errorf("inspect working document %s at %s: %w", workingPath, rev, err)
	}
	path := "generated/" + stem + "-" + version + extension
	working := strings.TrimSpace(string(listed)) == workingPath
	if working {
		path = workingPath
	}
	data, err := gitOutput(root, "show", rev+":"+path)
	if err != nil {
		return nil, "", false, fmt.Errorf("read contract %s at %s: %w", path, rev, err)
	}
	return data, path, working, nil
}

func buildContractReport(root, baseRef, headRef, apiFile string) (contractReport, error) {
	base, err := reportRevision(root, baseRef)
	if err != nil {
		return contractReport{}, err
	}
	head, err := reportRevision(root, headRef)
	if err != nil {
		return contractReport{}, err
	}
	api, err := os.ReadFile(apiFile)
	if err != nil {
		return contractReport{}, fmt.Errorf("read go-apidiff signal %s: %w; rerun the API gate", apiFile, err)
	}
	baseVersions, err := versionsAt(root, base)
	if err != nil {
		return contractReport{}, err
	}
	headVersions, err := versionsAt(root, head)
	if err != nil {
		return contractReport{}, err
	}
	temp, err := os.MkdirTemp("", "schema-contract-report-")
	if err != nil {
		return contractReport{}, fmt.Errorf("create contract report scratch directory: %w", err)
	}
	defer os.RemoveAll(temp)
	rep := contractReport{Report: 1, Base: base, Head: head, GoAPIBreaking: len(bytes.TrimSpace(api)) != 0}
	for _, surface := range reportSurfaces {
		result := surfaceReport{Name: surface.name, Change: reportNone}
		for _, stem := range surface.stems {
			requestSchema := stem == "publish-request" || stem == "annotation-push-request"
			extension := ".json"
			if requestSchema {
				extension = ".schema.json"
			}
			oldBytes, basePath, baseWorking, err := reportDocumentAt(root, base, stem, baseVersions[surface.version], extension)
			if err != nil {
				return contractReport{}, err
			}
			newBytes, headPath, _, err := reportDocumentAt(root, head, stem, headVersions[surface.version], extension)
			if err != nil {
				return contractReport{}, err
			}
			change, breaking, err := compareReportDocument(temp, stem, oldBytes, newBytes, requestSchema)
			if err != nil {
				return contractReport{}, fmt.Errorf("compare %s (%s to %s): %w", stem, basePath, headPath, err)
			}
			result.Change = strongerReportChange(result.Change, change)
			result.Breaking = result.Breaking || breaking
			if baseWorking {
				frozenPath := "generated/" + stem + "-" + baseVersions[surface.version] + extension
				frozen, err := gitOutput(root, "show", base+":"+frozenPath)
				if err != nil {
					return contractReport{}, fmt.Errorf("read base frozen contract %s: %w", frozenPath, err)
				}
				baseChange, _, err := compareReportDocument(temp, stem, frozen, oldBytes, requestSchema)
				if err != nil {
					return contractReport{}, fmt.Errorf("compare base working and frozen %s: %w", stem, err)
				}
				result.AheadOnBase = result.AheadOnBase || baseChange == reportFunctional
			}
		}
		if result.Breaking {
			result.Change = reportFunctional
		}
		rep.Surfaces = append(rep.Surfaces, result)
	}
	return rep, nil
}

func strongerReportChange(a, b reportChange) reportChange {
	if a == reportFunctional || b == reportFunctional {
		return reportFunctional
	}
	if a == reportDocs || b == reportDocs {
		return reportDocs
	}
	return reportNone
}

func normalizeReportDocument(data []byte, requestSchema, stripDescriptions bool) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode generated JSON document: %w", err)
	}
	if doc == nil {
		return nil, errors.New("generated JSON document is not an object")
	}
	if requestSchema {
		delete(doc, "$id")
	} else {
		info, ok := doc["info"].(map[string]any) // unstring-allow: go-json-key-assert — normalize arbitrary OpenAPI documents
		if !ok {
			return nil, errors.New("generated OpenAPI document has no info object")
		}
		delete(info, "version")
	}
	if stripDescriptions {
		removeDescriptions(doc)
	}
	return json.Marshal(doc)
}

func removeDescriptions(v any) {
	switch value := v.(type) {
	case map[string]any:
		delete(value, "description")
		for _, child := range value {
			removeDescriptions(child)
		}
	case []any:
		for _, child := range value {
			removeDescriptions(child)
		}
	}
}

func compareReportDocument(temp, stem string, oldBytes, newBytes []byte, requestSchema bool) (reportChange, bool, error) {
	oldDoc, err := normalizeReportDocument(oldBytes, requestSchema, false)
	if err != nil {
		return reportNone, false, err
	}
	newDoc, err := normalizeReportDocument(newBytes, requestSchema, false)
	if err != nil {
		return reportNone, false, err
	}
	if requestSchema {
		if bytes.Equal(oldDoc, newDoc) {
			return reportNone, false, nil
		}
		oldNoDocs, err := normalizeReportDocument(oldBytes, true, true)
		if err != nil {
			return reportNone, false, err
		}
		newNoDocs, err := normalizeReportDocument(newBytes, true, true)
		if err != nil {
			return reportNone, false, err
		}
		if bytes.Equal(oldNoDocs, newNoDocs) {
			return reportDocs, false, nil
		}
		return reportFunctional, false, nil
	}
	oldPath, newPath := filepath.Join(temp, stem+"-base.json"), filepath.Join(temp, stem+"-head.json")
	if err := os.WriteFile(oldPath, oldDoc, 0o600); err != nil {
		return reportNone, false, err
	}
	if err := os.WriteFile(newPath, newDoc, 0o600); err != nil {
		return reportNone, false, err
	}
	diff, err := oasdiffSignal("diff", oldPath, newPath, "--fail-on-diff")
	if err != nil || !diff {
		return reportNone, false, err
	}
	functional, err := oasdiffSignal("diff", oldPath, newPath, "--fail-on-diff", "--exclude-elements", "description,examples,summary,title")
	if err != nil {
		return reportNone, false, err
	}
	breaking, err := oasdiffSignal("breaking", oldPath, newPath, "--fail-on", "ERR")
	if err != nil {
		return reportNone, false, err
	}
	if functional || breaking {
		return reportFunctional, breaking, nil
	}
	return reportDocs, false, nil
}

// A non-zero status alone is not a break: oasdiff also uses exit 1 for
// invocation and document errors. Require its structured result to prove it.
func oasdiffSignal(command, oldPath, newPath string, flags ...string) (bool, error) {
	args := append([]string{command, oldPath, newPath, "--format", "json", "--allow-external-refs=false"}, flags...)
	cmd := exec.Command("oasdiff", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if err == nil {
		return false, nil
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		return false, fmt.Errorf("run oasdiff %s: %w: %s", command, err, stderr.String())
	}
	if command == "diff" {
		var object map[string]json.RawMessage
		if json.Unmarshal(stdout.Bytes(), &object) != nil || len(object) == 0 {
			return false, fmt.Errorf("oasdiff diff failed without a non-empty diff: %s", stderr.String())
		}
		return true, nil
	}
	var changes []struct {
		Level *float64 `json:"level"`
	}
	if json.Unmarshal(stdout.Bytes(), &changes) != nil || changes == nil {
		return false, fmt.Errorf("oasdiff breaking failed without a readable change list: %s", stderr.String())
	}
	for _, item := range changes {
		if item.Level != nil && *item.Level >= 3 {
			return true, nil
		}
	}
	return false, fmt.Errorf("oasdiff breaking exited 1 without an ERR-level change: %s", stderr.String())
}
