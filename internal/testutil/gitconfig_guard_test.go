package testutil

import (
	"bufio"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/peasant-labs/schema/testcase"
	"github.com/peasant-labs/schema/testcase/assert"
	"gopkg.in/yaml.v3"
)

// The guard below fails when a Go test package in this module can start a
// process without the package-level git isolation from RunWithIsolatedGitConfig.
//
// A package "can start a process" when its tests import os/exec, or when the
// package itself or a module package its tests import reaches os/exec through
// non-test code. Every such package must have a TestMain that calls
// RunWithIsolatedGitConfig. The rule deliberately covers every child process,
// not only `git`: shell scripts, `go` (which runs git to stamp version control
// information), and helper binaries all inherit the environment.
//
// What it cannot see: a process started without os/exec (os.StartProcess,
// syscall.Exec), git run by a dependency outside this module, a TestMain that
// calls the helper only on some path, and a test that rebuilds a child's
// environment from scratch without the isolation variables. Those need review.

//go:embed testdata/git_isolation_packages.yaml
var gitIsolationPackagesManifest []byte

//go:embed testdata/git_isolation_scanner.yaml
var gitIsolationScannerCorpus []byte

const (
	isolationCall = "RunWithIsolatedGitConfig"
	// The embed directives above must name the same files.
	gitIsolationPackagesPath = "testdata/git_isolation_packages.yaml"
	gitIsolationScannerPath  = "testdata/git_isolation_scanner.yaml"
	// testdataDirName is the directory the go tool ignores; nothing in it is a package.
	testdataDirName = "testdata"
)

type gitIsolationInput struct {
	Files map[string]string `yaml:"files"`
}

type gitIsolationExpected struct {
	Needs      []string `yaml:"needs"`
	Violations []string `yaml:"violations"`
}

type gitIsolationReport struct {
	Needs      []string
	Violations []string
}

type goPackageDir struct {
	hasTests    bool
	testExec    bool
	prodExec    bool
	isolated    bool
	testImports []string
	prodImports []string
}

// scanGitIsolation walks the module rooted at root (skipping testdata,
// hidden directories, node_modules, and nested modules) and reports, by
// slash-separated directory relative to root, which test packages need the
// isolation and which of those lack it.
func scanGitIsolation(root string) (gitIsolationReport, error) {
	modulePath, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		return gitIsolationReport{}, err
	}
	dirs := map[string]*goPackageDir{}
	fset := token.NewFileSet()
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if p != root && (name == testdataDirName || name == "node_modules" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
				return filepath.SkipDir
			}
			if p != root {
				if _, statErr := os.Stat(filepath.Join(p, "go.mod")); statErr == nil {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") {
			return nil
		}
		rel, err := filepath.Rel(root, filepath.Dir(p))
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		dir := dirs[rel]
		if dir == nil {
			dir = &goPackageDir{}
			dirs[rel] = dir
		}
		isTest := strings.HasSuffix(p, "_test.go")
		mode := parser.ImportsOnly
		if isTest {
			mode = parser.SkipObjectResolution
		}
		file, err := parser.ParseFile(fset, p, nil, mode)
		if err != nil {
			return fmt.Errorf("scan git isolation: parse %s: %w", p, err)
		}
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return fmt.Errorf("scan git isolation: import path %s in %s: %w", spec.Path.Value, p, err)
			}
			local := ""
			if importPath == modulePath {
				local = "."
			} else if rest, ok := strings.CutPrefix(importPath, modulePath+"/"); ok {
				local = rest
			}
			switch {
			case importPath == "os/exec" && isTest:
				dir.testExec = true
			case importPath == "os/exec":
				dir.prodExec = true
			case local != "" && isTest:
				dir.testImports = append(dir.testImports, local)
			case local != "":
				dir.prodImports = append(dir.prodImports, local)
			}
		}
		if isTest {
			dir.hasTests = true
			if callsIsolationFromTestMain(file) {
				dir.isolated = true
			}
		}
		return nil
	})
	if err != nil {
		return gitIsolationReport{}, err
	}

	reaches := map[string]bool{}
	visiting := map[string]bool{}
	var reachesExec func(string) bool
	reachesExec = func(pkg string) bool {
		if done, ok := reaches[pkg]; ok {
			return done
		}
		dir := dirs[pkg]
		if dir == nil || visiting[pkg] {
			return false
		}
		visiting[pkg] = true
		result := dir.prodExec
		for _, imported := range dir.prodImports {
			result = result || reachesExec(imported)
		}
		visiting[pkg] = false
		reaches[pkg] = result
		return result
	}

	var report gitIsolationReport
	for pkg, dir := range dirs {
		if !dir.hasTests {
			continue
		}
		needs := dir.testExec || reachesExec(pkg)
		for _, imported := range dir.testImports {
			needs = needs || reachesExec(imported)
		}
		if !needs {
			continue
		}
		report.Needs = append(report.Needs, pkg)
		if !dir.isolated {
			report.Violations = append(report.Violations, pkg)
		}
	}
	sort.Strings(report.Needs)
	sort.Strings(report.Violations)
	return report, nil
}

// callsIsolationFromTestMain reports whether file declares TestMain and its
// body calls a function or method named RunWithIsolatedGitConfig.
func callsIsolationFromTestMain(file *ast.File) bool {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name.Name != "TestMain" || fn.Body == nil {
			continue
		}
		found := false
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return !found
			}
			switch fun := call.Fun.(type) {
			case *ast.SelectorExpr:
				found = found || fun.Sel.Name == isolationCall
			case *ast.Ident:
				found = found || fun.Name == isolationCall
			}
			return !found
		})
		if found {
			return true
		}
	}
	return false
}

func readModulePath(goMod string) (string, error) {
	f, err := os.Open(goMod)
	if err != nil {
		return "", fmt.Errorf("scan git isolation: open %s: %w", goMod, err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(scanner.Text()), "module "); ok {
			return strings.Trim(strings.TrimSpace(rest), `"`), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan git isolation: read %s: %w", goMod, err)
	}
	return "", fmt.Errorf("scan git isolation: %s has no module directive", goMod)
}

func moduleRootDir(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod above the test's working directory")
		}
		dir = parent
	}
}

// TestEveryPackageThatStartsProcessesIsolatesGitConfig is the regression guard
// for the module itself.
func TestEveryPackageThatStartsProcessesIsolatesGitConfig(t *testing.T) {
	report, err := scanGitIsolation(moduleRootDir(t))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		RequiredNames []string `yaml:"requiredNames"`
	}
	if err := yaml.Unmarshal(gitIsolationPackagesManifest, &manifest); err != nil || len(manifest.RequiredNames) == 0 {
		t.Fatalf("read %s: names %v, error %v", gitIsolationPackagesPath, manifest.RequiredNames, err)
	}
	for _, name := range manifest.RequiredNames {
		if !slices.Contains(report.Needs, name) {
			t.Errorf("package %s runs git in its tests but the scanner no longer detects it (detected: %v); "+
				"the scanner lost coverage, so fix the scanner rather than removing the name from %s", name, report.Needs, gitIsolationPackagesPath)
		}
	}
	for _, pkg := range report.Violations {
		t.Errorf("package %s starts processes in its tests but has no TestMain calling testutil.%s; "+
			"git there reads the developer's configuration (for example commit.gpgsign), so add "+
			"`func TestMain(m *testing.M) { os.Exit(testutil.%s(m)) }` to a _test.go file in %s",
			pkg, isolationCall, isolationCall, pkg)
	}
}

// TestGitIsolationScannerCorpus proves the scanner detects each way a package
// can start processes and rejects each way the isolation can be missing.
func TestGitIsolationScannerCorpus(t *testing.T) {
	corpus, err := testcase.LoadCorpus[gitIsolationInput, gitIsolationExpected](gitIsolationScannerCorpus)
	if err != nil {
		t.Fatalf("load %s: %v", gitIsolationScannerPath, err)
	}
	assert.RequireMin(t, corpus, 6)
	assert.RequireValid(t, corpus)
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root := t.TempDir()
			for name, content := range c.Input.Files {
				p := filepath.Join(root, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			report, err := scanGitIsolation(root)
			if err != nil {
				t.Fatal(err)
			}
			if c.Classification == testcase.MustFail && len(report.Violations) == 0 {
				t.Fatalf("must-fail case produced no violation (needs %v)", report.Needs)
			}
			if c.Classification == testcase.MustPass && len(report.Violations) != 0 {
				t.Fatalf("must-pass case produced violations %v", report.Violations)
			}
			if !slices.Equal(report.Needs, c.Expected.Needs) {
				t.Errorf("needs = %v, want %v", report.Needs, c.Expected.Needs)
			}
			if !slices.Equal(report.Violations, c.Expected.Violations) {
				t.Errorf("violations = %v, want %v", report.Violations, c.Expected.Violations)
			}
		})
	}
}

// TestIsolateGitConfigOverridesEveryAmbientSource pins the variables the
// isolation sets and removes, starting from a hostile environment.
func TestIsolateGitConfigOverridesEveryAmbientSource(t *testing.T) {
	hostile := map[string]string{
		"GIT_CONFIG_GLOBAL":     "/tmp/hostile.gitconfig",
		"GIT_CONFIG_NOSYSTEM":   "0",
		"GIT_CONFIG_COUNT":      "1",
		"GIT_CONFIG_PARAMETERS": "'commit.gpgsign'='true'",
	}
	for name, value := range hostile {
		t.Setenv(name, value)
	}
	if err := isolateGitConfig(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("GIT_CONFIG_GLOBAL"); got != os.DevNull {
		t.Errorf("GIT_CONFIG_GLOBAL = %q, want %q", got, os.DevNull)
	}
	if got := os.Getenv("GIT_CONFIG_NOSYSTEM"); got != "1" {
		t.Errorf("GIT_CONFIG_NOSYSTEM = %q, want 1", got)
	}
	for _, name := range []string{"GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS"} {
		if value, set := os.LookupEnv(name); set {
			t.Errorf("%s is still set to %q; injected configuration would bypass the files", name, value)
		}
	}
}
