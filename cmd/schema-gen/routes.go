package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/peasant-labs/schema/openapi"
)

// Route members are a projection of the built Village API and Local API
// documents: one typed member per operation, in Go (package route) and in
// TypeScript (the /village-api and /local-api subpaths). One collector reads the
// documents and refuses anything it cannot project faithfully, and both
// renderers read only its output, so the two languages cannot disagree.

// routeGoRelPath and routeTSRelPath are the committed generated files,
// relative to the module root.
var (
	routeGoRelPath = filepath.Join("route", "routes_gen.go")
	routeTSRelPath = filepath.Join("typescript", "src", "internal", "generated", "routes.gen.ts")
)

// routeAPI is one document that route members are generated from.
type routeAPI struct {
	// label names the document in generated comments and errors.
	label string
	// goPrefix prefixes each Go member, for example VillageGetGroup.
	goPrefix string
	// goAll is the Go ordered set of the API's members.
	goAll string
	// tsObject is the TypeScript object holding the API's members.
	tsObject string
	// tsAll is the TypeScript ordered set of the API's members.
	tsAll string
	// artifact is the generated document's file name.
	artifact func() string
}

var (
	villageRouteAPI = routeAPI{
		label: "Village API", goPrefix: "Village", goAll: "AllVillage",
		tsObject: "VillageRoute", tsAll: "ALL_VILLAGE_ROUTES",
		artifact: func() string { return "village-api-" + openapi.VillageAPIVersion + ".json" },
	}
	localRouteAPI = routeAPI{
		label: "Local API", goPrefix: "Local", goAll: "AllLocal",
		tsObject: "LocalRoute", tsAll: "ALL_LOCAL_ROUTES",
		artifact: func() string { return "peasantlocal-api-" + openapi.PeasantLocalAPIVersion + ".json" },
	}
	// routeAPIs is the ordered set of documents route members come from.
	routeAPIs = []routeAPI{villageRouteAPI, localRouteAPI}
)

// routeMethod is one member of the closed method set. Its order is the order of
// route.AllMethods, which a test holds equal.
type routeMethod struct {
	docKey   string // the OpenAPI path-item key
	wire     string // the method on the wire
	goConst  string // the route package constant
	tsMember string // the TypeScript HttpMethod member
}

var routeMethods = []routeMethod{
	{docKey: "get", wire: "GET", goConst: "MethodGet", tsMember: "Get"},
	{docKey: "post", wire: "POST", goConst: "MethodPost", tsMember: "Post"},
	{docKey: "put", wire: "PUT", goConst: "MethodPut", tsMember: "Put"},
	{docKey: "patch", wire: "PATCH", goConst: "MethodPatch", tsMember: "Patch"},
	{docKey: "delete", wire: "DELETE", goConst: "MethodDelete", tsMember: "Delete"},
}

// routeOtherMethods are path-item keys that name an HTTP method outside the set.
var routeOtherMethods = map[string]bool{"head": true, "options": true, "trace": true}

// routeParamKind is the closed set of path parameter types a member accepts.
type routeParamKind string

const (
	routeParamString  routeParamKind = "string"
	routeParamInteger routeParamKind = "integer"
)

// routeParam is one path parameter, in template order.
type routeParam struct {
	Name string         `yaml:"name"`
	Kind routeParamKind `yaml:"kind"`
}

// routeOperation is one collected operation. Name is the member name inside
// its API: the operationId with an upper-case first letter.
type routeOperation struct {
	Name        string       `yaml:"name"`
	Method      string       `yaml:"method"`
	Pattern     string       `yaml:"pattern"`
	OperationID string       `yaml:"operation_id"`
	Deprecated  bool         `yaml:"deprecated"`
	Params      []routeParam `yaml:"params"`

	method   routeMethod
	segments []routeSegment
}

// routeSegment is a literal run of the template or one parameter.
type routeSegment struct {
	literal string
	param   *routeParam
}

var (
	routeIdentifierPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
	routeParamNamePattern  = regexp.MustCompile(`^[a-z][A-Za-z0-9]*$`)
	routeTemplatePattern   = regexp.MustCompile(`\{([^{}]*)\}`)
	// routeReservedParamNames are route package identifiers a generated path
	// builder calls; a parameter with one of these names would shadow it.
	routeReservedParamNames = map[string]bool{"pathSegment": true, "intPathSegment": true, "checkPathSegment": true, "err": true}
)

type routeDocument struct {
	Paths      map[string]map[string]json.RawMessage `json:"paths"`
	Components struct {
		Schemas map[string]routeSchema `json:"schemas"`
	} `json:"components"`
}

type routeDocOperation struct {
	OperationID string          `json:"operationId"`
	Deprecated  bool            `json:"deprecated"`
	Parameters  []routeDocParam `json:"parameters"`
}

type routeDocParam struct {
	Ref    string       `json:"$ref"`
	In     string       `json:"in"`
	Name   string       `json:"name"`
	Schema *routeSchema `json:"schema"`
}

type routeSchema struct {
	Ref  string          `json:"$ref"`
	Type json.RawMessage `json:"type"`
}

// collectRoutes reads every operation of one built document, in pattern order
// and then route.AllMethods order. It refuses a document it cannot project
// faithfully, naming the operation and the fix.
func collectRoutes(api routeAPI, document []byte) ([]routeOperation, error) {
	var doc routeDocument
	if err := json.Unmarshal(document, &doc); err != nil {
		return nil, routeError(api, "the document", "is not valid OpenAPI JSON: "+err.Error(), "regenerate the documents with `go run ./cmd/schema-gen` before the route members")
	}
	if len(doc.Paths) == 0 {
		return nil, routeError(api, "the document", "declares no paths", "build the document from its operation declarations before generating route members")
	}

	patterns := make([]string, 0, len(doc.Paths))
	for pattern := range doc.Paths {
		patterns = append(patterns, pattern)
	}
	sort.Strings(patterns)

	var ops []routeOperation
	identifiers := map[string]string{}
	operationIDs := map[string]string{}
	for _, pattern := range patterns {
		item := doc.Paths[pattern]
		where := "path " + pattern
		if _, ok := item["parameters"]; ok {
			return nil, routeError(api, where, "declares path-level parameters, which route member generation does not read", "declare the path parameters on each operation")
		}
		for key := range item {
			if routeOtherMethods[key] {
				return nil, routeError(api, where, "declares method "+strings.ToUpper(key)+", which is not in the route method set GET, POST, PUT, PATCH, DELETE", "declare the operation with a method in the set, or add the method to route.Method and the generator's method table together")
			}
		}
		segments, names, err := parseRouteTemplate(api, pattern)
		if err != nil {
			return nil, err
		}
		for _, method := range routeMethods {
			raw, ok := item[method.docKey]
			if !ok {
				continue
			}
			where := method.wire + " " + pattern
			var docOp routeDocOperation
			if err := json.Unmarshal(raw, &docOp); err != nil {
				return nil, routeError(api, where, "is not a valid operation object: "+err.Error(), "regenerate the documents with `go run ./cmd/schema-gen`")
			}
			op, err := collectRouteOperation(api, doc, where, method, pattern, segments, names, docOp)
			if err != nil {
				return nil, err
			}
			if previous, ok := operationIDs[op.OperationID]; ok {
				return nil, routeError(api, where, "repeats operationId "+op.OperationID+", already used by "+previous, "give each operation a unique operationId")
			}
			operationIDs[op.OperationID] = where
			for _, identifier := range op.goIdentifiers(api) {
				if previous, ok := identifiers[identifier]; ok {
					return nil, routeError(api, where, "would generate Go identifier "+identifier+", already generated for "+previous, "rename one of the two operationIds")
				}
				identifiers[identifier] = where
			}
			ops = append(ops, op)
		}
	}
	return ops, nil
}

func collectRouteOperation(api routeAPI, doc routeDocument, where string, method routeMethod, pattern string, segments []routeSegment, names []string, docOp routeDocOperation) (routeOperation, error) {
	if docOp.OperationID == "" {
		return routeOperation{}, routeError(api, where, "has no operationId, so its route member has no name", "declare an operationId on the operation")
	}
	if !routeIdentifierPattern.MatchString(docOp.OperationID) {
		return routeOperation{}, routeError(api, where, "has operationId "+strconv.Quote(docOp.OperationID)+", which is not an identifier of ASCII letters and digits", "rename the operationId in lower camel case, for example getTranscriptSegments")
	}

	declared := map[string]routeParamKind{}
	for _, param := range docOp.Parameters {
		if param.Ref != "" {
			return routeOperation{}, routeError(api, where, "declares the referenced parameter "+param.Ref+", which route member generation does not resolve", "declare the parameter inline on the operation")
		}
		if param.In != "path" {
			continue
		}
		if !containsString(names, param.Name) {
			return routeOperation{}, routeError(api, where, "declares path parameter "+strconv.Quote(param.Name)+", which is not in the path template", "remove the parameter or add {"+param.Name+"} to the path")
		}
		kind, err := routeParamKindOf(api, doc, where, param)
		if err != nil {
			return routeOperation{}, err
		}
		declared[param.Name] = kind
	}

	op := routeOperation{
		Name:        strings.ToUpper(docOp.OperationID[:1]) + docOp.OperationID[1:],
		Method:      method.wire,
		Pattern:     pattern,
		OperationID: docOp.OperationID,
		Deprecated:  docOp.Deprecated,
		method:      method,
	}
	for _, segment := range segments {
		if segment.param == nil {
			op.segments = append(op.segments, segment)
			continue
		}
		kind, ok := declared[segment.param.Name]
		if !ok {
			return routeOperation{}, routeError(api, where, "has {"+segment.param.Name+"} in its path template but declares no path parameter of that name", "declare the path parameter on the operation")
		}
		param := routeParam{Name: segment.param.Name, Kind: kind}
		op.Params = append(op.Params, param)
		op.segments = append(op.segments, routeSegment{param: &param})
	}
	return op, nil
}

// parseRouteTemplate splits a path template into literal runs and parameters.
func parseRouteTemplate(api routeAPI, pattern string) ([]routeSegment, []string, error) {
	where := "path " + pattern
	if !strings.HasPrefix(pattern, "/") {
		return nil, nil, routeError(api, where, "does not start with /", "declare the path from the server root")
	}
	var segments []routeSegment
	var names []string
	offset := 0
	for _, match := range routeTemplatePattern.FindAllStringSubmatchIndex(pattern, -1) {
		literal := pattern[offset:match[0]]
		if strings.ContainsAny(literal, "{}") {
			return nil, nil, routeError(api, where, "has an unbalanced brace", "write each path parameter as {name}")
		}
		segments = append(segments, routeSegment{literal: literal})
		name := pattern[match[2]:match[3]]
		if !routeParamNamePattern.MatchString(name) {
			return nil, nil, routeError(api, where, "has path parameter "+strconv.Quote(name)+", which is not a lower camel identifier of ASCII letters and digits", "rename the parameter, for example {transcriptId}")
		}
		if token.IsKeyword(name) || routeReservedParamNames[name] {
			return nil, nil, routeError(api, where, "has path parameter "+strconv.Quote(name)+", which is reserved in the generated Go path builder", "rename the parameter")
		}
		if containsString(names, name) {
			return nil, nil, routeError(api, where, "repeats path parameter "+strconv.Quote(name), "give each path parameter a unique name")
		}
		names = append(names, name)
		segments = append(segments, routeSegment{param: &routeParam{Name: name}})
		offset = match[1]
	}
	tail := pattern[offset:]
	if strings.ContainsAny(tail, "{}") {
		return nil, nil, routeError(api, where, "has an unbalanced brace", "write each path parameter as {name}")
	}
	segments = append(segments, routeSegment{literal: tail})
	return segments, names, nil
}

// routeParamKindOf reads a path parameter's type, following one component
// reference.
func routeParamKindOf(api routeAPI, doc routeDocument, where string, param routeDocParam) (routeParamKind, error) {
	if param.Schema == nil {
		return "", routeError(api, where, "declares path parameter "+strconv.Quote(param.Name)+" without a schema", "give the parameter a string or integer schema")
	}
	schema := *param.Schema
	if schema.Ref != "" {
		const prefix = "#/components/schemas/"
		target, ok := doc.Components.Schemas[strings.TrimPrefix(schema.Ref, prefix)]
		if !strings.HasPrefix(schema.Ref, prefix) || !ok {
			return "", routeError(api, where, "declares path parameter "+strconv.Quote(param.Name)+" with the unresolved schema reference "+schema.Ref, "reference a component schema the document declares")
		}
		schema = target
	}
	var typ string
	if len(schema.Type) > 0 {
		_ = json.Unmarshal(schema.Type, &typ)
	}
	switch typ {
	case string(routeParamString):
		return routeParamString, nil
	case string(routeParamInteger):
		return routeParamInteger, nil
	}
	declaredType := strings.TrimSpace(string(schema.Type))
	if declaredType == "" {
		declaredType = "(none)"
	}
	return "", routeError(api, where, "declares path parameter "+strconv.Quote(param.Name)+" with unsupported path parameter type "+declaredType+"; a member accepts only string and integer", "declare the parameter as a string or an integer")
}

func routeError(api routeAPI, where, what, fix string) error {
	return fmt.Errorf("route member generation from the %s document: %s %s; no route members were written; %s, then run `go run ./cmd/schema-gen`", api.label, where, what, fix)
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func (op routeOperation) goName(api routeAPI) string { return api.goPrefix + op.Name }

func (op routeOperation) goIdentifiers(api routeAPI) []string {
	identifiers := []string{op.goName(api)}
	if len(op.Params) > 0 {
		identifiers = append(identifiers, op.goName(api)+"Path")
	}
	return identifiers
}

// buildRouteArtifacts collects both documents and renders the Go and the
// TypeScript files, keyed by their path relative to the module root. main()
// writes them and the freshness test diffs them, so the two cannot disagree.
func buildRouteArtifacts(specs openapi.SpecArtifacts) (map[string][]byte, error) {
	collected := make([][]routeOperation, len(routeAPIs))
	for i, api := range routeAPIs {
		document, ok := specs[api.artifact()]
		if !ok {
			return nil, routeError(api, "document "+api.artifact(), "was not built", "keep the document in openapi.GenerateSpecArtifacts")
		}
		ops, err := collectRoutes(api, document)
		if err != nil {
			return nil, err
		}
		collected[i] = ops
	}
	goSource, err := renderRouteGo(collected)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{
		routeGoRelPath: goSource,
		routeTSRelPath: renderRouteTS(collected),
	}, nil
}

// generateRoutes writes the route member files under moduleRoot.
func generateRoutes(moduleRoot string, specs openapi.SpecArtifacts) error {
	artifacts, err := buildRouteArtifacts(specs)
	if err != nil {
		return err
	}
	for _, rel := range sortedKeys(artifacts) {
		path := filepath.Join(moduleRoot, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("create dir for %s: %w", path, err)
		}
		if err := os.WriteFile(path, artifacts[rel], 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

func sortedKeys(values map[string][]byte) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

const routeGeneratedHeader = "// Code generated by `go run ./cmd/schema-gen`. DO NOT EDIT.\n"

func renderRouteGo(collected [][]routeOperation) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(routeGeneratedHeader)
	b.WriteString("\npackage route\n")
	for i, api := range routeAPIs {
		ops := collected[i]
		for _, op := range ops {
			name := op.goName(api)
			fmt.Fprintf(&b, "\n// %s is %s %s\n// (operationId %s) in the %s document.\n", name, op.Method, op.Pattern, op.OperationID, api.label)
			writeGoDeprecation(&b, api, op)
			fmt.Fprintf(&b, "var %s = Route{method: %s, pattern: %s, operationID: %s}\n", name, op.method.goConst, strconv.Quote(op.Pattern), strconv.Quote(op.OperationID))
			if len(op.Params) == 0 {
				continue
			}
			fmt.Fprintf(&b, "\n// %sPath returns the path of %s.\n// Each string parameter is escaped as one path segment with url.PathEscape;\n// an empty, \".\" or \"..\" value is refused with a *PathSegmentError.\n", name, name)
			writeGoDeprecation(&b, api, op)
			fmt.Fprintf(&b, "func %sPath(%s) (string, error) {\n", name, goRouteParams(op.Params))
			for _, param := range op.Params {
				if param.Kind == routeParamString {
					fmt.Fprintf(&b, "\tif err := checkPathSegment(%s, %s, %s); err != nil {\n\t\treturn \"\", err\n\t}\n", name, strconv.Quote(param.Name), param.Name)
				}
			}
			fmt.Fprintf(&b, "\treturn %s, nil\n}\n", goRouteBody(op.segments))
		}
		fmt.Fprintf(&b, "\n// %s is every operation of the %s document, ordered by pattern and then\n// by AllMethods.\nvar %s = []Route{\n", api.goAll, api.label, api.goAll)
		for _, op := range ops {
			fmt.Fprintf(&b, "\t%s,\n", op.goName(api))
		}
		b.WriteString("}\n")
	}
	formatted, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("route member generation: the rendered Go source does not format (%w); no route members were written; fix renderRouteGo in cmd/schema-gen/routes.go", err)
	}
	return formatted, nil
}

func writeGoDeprecation(b *bytes.Buffer, api routeAPI, op routeOperation) {
	if op.Deprecated {
		fmt.Fprintf(b, "//\n// Deprecated: the %s document marks %s deprecated.\n", api.label, op.OperationID)
	}
}

// goRouteParams joins consecutive parameters of one type, as gofmt users write.
func goRouteParams(params []routeParam) string {
	var parts []string
	for i, param := range params {
		if i+1 < len(params) && params[i+1].Kind == param.Kind {
			parts = append(parts, param.Name)
			continue
		}
		parts = append(parts, param.Name+" "+goRouteParamType(param.Kind))
	}
	return strings.Join(parts, ", ")
}

func goRouteParamType(kind routeParamKind) string {
	if kind == routeParamInteger {
		return "int"
	}
	return "string"
}

func goRouteBody(segments []routeSegment) string {
	var parts []string
	for _, segment := range segments {
		switch {
		case segment.param == nil && segment.literal == "":
			continue
		case segment.param == nil:
			parts = append(parts, strconv.Quote(segment.literal))
		case segment.param.Kind == routeParamInteger:
			parts = append(parts, "intPathSegment("+segment.param.Name+")")
		default:
			parts = append(parts, "pathSegment("+segment.param.Name+")")
		}
	}
	return strings.Join(parts, " + ")
}

func renderRouteTS(collected [][]routeOperation) []byte {
	var b bytes.Buffer
	b.WriteString(routeGeneratedHeader)
	b.WriteString("import { routeMember } from \"../../route-member.js\";\n\n")
	b.WriteString("/** HttpMethod is the closed set of HTTP methods the Village API and Local API documents declare. */\n")
	b.WriteString("export const HttpMethod = Object.freeze({\n")
	for _, method := range routeMethods {
		fmt.Fprintf(&b, "  %s: %s,\n", method.tsMember, jsString(method.wire))
	}
	b.WriteString("} as const);\nexport type HttpMethod = (typeof HttpMethod)[keyof typeof HttpMethod];\n")
	b.WriteString("/** ALL_HTTP_METHODS is the canonical order of HttpMethod. */\nexport const ALL_HTTP_METHODS = Object.freeze([")
	for i, method := range routeMethods {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("HttpMethod." + method.tsMember)
	}
	b.WriteString("] as const);\n")
	b.WriteString("export function isHttpMethod(value: unknown): value is HttpMethod {\n  return typeof value === \"string\" && (ALL_HTTP_METHODS as readonly string[]).includes(value);\n}\n")
	for i, api := range routeAPIs {
		ops := collected[i]
		fmt.Fprintf(&b, "\n/** %s holds one member per operation of the %s document, named by operationId. */\n", api.tsObject, api.label)
		fmt.Fprintf(&b, "export const %s = Object.freeze({\n", api.tsObject)
		for _, op := range ops {
			fmt.Fprintf(&b, "  /**\n   * `%s %s` (operationId `%s`).\n", op.Method, op.Pattern, op.OperationID)
			if op.Deprecated {
				fmt.Fprintf(&b, "   * @deprecated The %s document marks this operation deprecated.\n", api.label)
			}
			fmt.Fprintf(&b, "   */\n  %s: routeMember(HttpMethod.%s, %s, %s),\n", op.Name, op.method.tsMember, jsString(op.Pattern), jsString(op.OperationID))
		}
		b.WriteString("} as const);\n")
		fmt.Fprintf(&b, "/** %s is every member of %s, ordered by pattern and then by ALL_HTTP_METHODS. */\n", api.tsAll, api.tsObject)
		fmt.Fprintf(&b, "export const %s = Object.freeze([\n", api.tsAll)
		for _, op := range ops {
			fmt.Fprintf(&b, "  %s.%s,\n", api.tsObject, op.Name)
		}
		b.WriteString("] as const);\n")
	}
	return b.Bytes()
}

// jsString quotes a value as a JSON string, which is also a TypeScript string
// literal.
func jsString(value string) string {
	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
	return strings.TrimSuffix(b.String(), "\n")
}
