// Command openapi-codegen-split runs oapi-codegen once per paths file of a
// spec split by tools/openapi-split, producing one Go package per file that
// mirrors the paths hierarchy:
//
//	paths/Auth.yaml                              -> api/auth            (package auth)
//	paths/Auth/Headless Identity API Demo.yaml   -> api/auth/headlessidentityapidemo
//	paths/Bulk v2/Query.yaml                     -> api/bulkv2/query
//
// Each package gets the operations of its file plus the shared parts of the
// main document (info, servers, components, security, tags), so models are
// generated next to the operations that use them.
//
// The root package (the `package`/`output` of the base config, e.g. api/api.gen.go)
// is regenerated as an aggregator: its ServerInterface embeds every package's
// ServerInterface, Unimplemented embeds every package's Unimplemented, and
// HandlerWithOptions mounts every package's routes on one chi router.
//
// Usage: openapi-codegen-split -oapi-codegen <binary> -config <oapi-codegen.yaml> <openapi.yaml>
//
// Previously generated files (named like the base config's output) under the
// output directory are removed first, so renamed/deleted folders don't linger.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"go/token"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"text/template"
	"unicode"

	"gopkg.in/yaml.v3"
)

// group is one paths file and the Go package generated from it.
type group struct {
	file    string   // paths file relative to the spec directory, e.g. "paths/Auth.yaml"
	folders []string // folder chain, e.g. ["Auth", "Headless Identity API Demo"]
	segs    []string // package path segments below the output dir, e.g. ["auth", "headlessidentityapidemo"]
	paths   *yaml.Node
	Alias   string // import alias in the aggregator, e.g. "authHeadlessidentityapidemo"
	Import  string // import path
	outFile string // generated file
}

func main() {
	codegen := flag.String("oapi-codegen", "oapi-codegen", "oapi-codegen binary")
	configPath := flag.String("config", "", "base oapi-codegen config (its package/output name the root package)")
	pathsDir := flag.String("paths", "paths", "paths directory, relative to the spec, that the hierarchy is taken from")
	flag.Parse()
	if flag.NArg() != 1 || *configPath == "" {
		fmt.Fprintln(os.Stderr, "usage: openapi-codegen-split -oapi-codegen <binary> -config <oapi-codegen.yaml> <openapi.yaml>")
		os.Exit(2)
	}
	if err := run(*codegen, *configPath, *pathsDir, flag.Arg(0)); err != nil {
		fmt.Fprintln(os.Stderr, "openapi-codegen-split:", err)
		os.Exit(1)
	}
}

func run(codegen, configPath, pathsDir, specPath string) error {
	config, err := readYAML(configPath)
	if err != nil {
		return err
	}
	rootPkg, rootOut := scalar(config, "package"), scalar(config, "output")
	if rootPkg == "" || rootOut == "" {
		return fmt.Errorf("%s must set package and output", configPath)
	}
	outDir, outName := filepath.Dir(rootOut), filepath.Base(rootOut)
	rootImport, err := importPath(outDir)
	if err != nil {
		return err
	}

	doc, err := readYAML(specPath)
	if err != nil {
		return err
	}
	groups, err := groupPaths(doc, pathsDir)
	if err != nil {
		return fmt.Errorf("%s: %w", specPath, err)
	}

	seenDir, seenAlias := map[string]string{}, map[string]string{}
	for _, g := range groups {
		dir := path.Join(g.segs...)
		if other, ok := seenDir[dir]; ok {
			return fmt.Errorf("%q and %q both map to package %s", other, g.file, dir)
		}
		seenDir[dir] = g.file
		g.Alias = alias(g.segs)
		if other, ok := seenAlias[g.Alias]; ok {
			return fmt.Errorf("%q and %q both map to import alias %s", other, g.file, g.Alias)
		}
		seenAlias[g.Alias] = g.file
		g.Import = rootImport + "/" + dir
		g.outFile = filepath.Join(append([]string{outDir}, append(g.segs, outName)...)...)
	}

	if err := removeGenerated(outDir, outName); err != nil {
		return err
	}
	if err := generate(codegen, config, doc, specPath, groups); err != nil {
		return err
	}
	if err := writeAggregator(rootOut, rootPkg, groups); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Generated %d packages under %s and the aggregator %s\n", len(groups), outDir, rootOut)
	return nil
}

// groupPaths groups the `$ref`ed path items of the main document by paths file,
// in order of first appearance.
func groupPaths(doc *yaml.Node, pathsDir string) ([]*group, error) {
	paths := mappingValue(doc, "paths")
	if paths == nil || paths.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("no paths")
	}
	prefix := path.Clean(filepath.ToSlash(pathsDir)) + "/"
	var groups []*group
	byFile := map[string]*group{}
	for i := 0; i+1 < len(paths.Content); i += 2 {
		key, item := paths.Content[i], paths.Content[i+1]
		ref := scalar(item, "$ref")
		if ref == "" {
			return nil, fmt.Errorf("path %s is not a $ref (run tools/openapi-split first)", key.Value)
		}
		escaped, _, _ := strings.Cut(ref, "#")
		file, err := url.PathUnescape(escaped)
		if err != nil {
			return nil, fmt.Errorf("path %s: %w", key.Value, err)
		}
		file = path.Clean(file)
		g, ok := byFile[file]
		if !ok {
			rel, found := strings.CutPrefix(file, prefix)
			if !found {
				return nil, fmt.Errorf("path %s: %s is not under %s", key.Value, file, prefix)
			}
			g = &group{file: file, paths: &yaml.Node{Kind: yaml.MappingNode}}
			g.folders = strings.Split(strings.TrimSuffix(rel, path.Ext(rel)), "/")
			byFile[file] = g
			groups = append(groups, g)
		}
		g.paths.Content = append(g.paths.Content, key, item)
	}
	assignPackages(groups)
	return groups, nil
}

// assignPackages sets the package path of every group. Each folder maps to
// packageName(folder, false); sibling folders that would collide that way
// (e.g. "expression-set" and "expressionset") keep their separators as
// underscores instead ("expression_set"). Directory and file segments of the
// same folder (Auth.yaml, Auth/...) always map to the same package segment.
func assignPackages(groups []*group) {
	siblings := map[string]map[string]bool{} // parent chain -> folder names
	for _, g := range groups {
		for i, folder := range g.folders {
			parent := strings.Join(g.folders[:i], "/")
			if siblings[parent] == nil {
				siblings[parent] = map[string]bool{}
			}
			siblings[parent][folder] = true
		}
	}
	names := map[string]string{} // folder chain -> package segment
	for parent, folders := range siblings {
		count := map[string]int{}
		for folder := range folders {
			count[packageName(folder, false)]++
		}
		for folder := range folders {
			name := packageName(folder, false)
			if count[name] > 1 {
				name = packageName(folder, true)
			}
			names[path.Join(parent, folder)] = name
		}
	}
	for _, g := range groups {
		for i := range g.folders {
			g.segs = append(g.segs, names[strings.Join(g.folders[:i+1], "/")])
		}
	}
}

// generate writes a spec and config per group next to the main spec (so the
// relative $refs resolve) and runs oapi-codegen on them in parallel.
func generate(codegen string, config, doc *yaml.Node, specPath string, groups []*group) error {
	tmpDir, err := os.MkdirTemp("", "openapi-codegen-split-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	// The specs must sit next to the main spec for its relative $refs to resolve.
	specDir := filepath.Dir(specPath)

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
		sem  = make(chan struct{}, runtime.NumCPU())
	)
	for i, g := range groups {
		sub := withPaths(doc, g.paths)
		subSpec := filepath.Join(specDir, fmt.Sprintf(".codegen-%s-%d.yaml", filepath.Base(tmpDir), i))
		if err := writeYAML(subSpec, sub); err != nil {
			return err
		}
		defer os.Remove(subSpec)
		subConfig := filepath.Join(tmpDir, fmt.Sprintf("config-%d.yaml", i))
		if err := writeYAML(subConfig, withScalars(config, "package", g.segs[len(g.segs)-1], "output", g.outFile)); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(g.outFile), 0o755); err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out, err := exec.Command(codegen, "-config", subConfig, subSpec).CombinedOutput()
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("oapi-codegen for %s: %w\n%s", g.file, err, out))
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

var aggregator = template.Must(template.New("").Parse(`// Code generated by tools/openapi-codegen-split. DO NOT EDIT.

// Package {{.Package}} aggregates the packages generated per paths file of the
// OpenAPI spec into a single server interface and chi handler.
package {{.Package}}

import (
	"net/http"

	"github.com/go-chi/chi/v5"
{{range .Groups}}
	{{.Alias}} "{{.Import}}"{{end}}
)

// ServerInterface represents all server handlers of the spec.
type ServerInterface interface {
{{- range .Groups}}
	{{.Alias}}.ServerInterface{{end}}
}

type (
{{- range .Groups}}
	{{.Alias}}Unimplemented = {{.Alias}}.Unimplemented{{end}}
)

// Unimplemented answers every operation with 501 Not Implemented.
// Embed it and override the operations you implement.
type Unimplemented struct {
{{- range .Groups}}
	{{.Alias}}Unimplemented{{end}}
}

var _ ServerInterface = Unimplemented{}

type MiddlewareFunc func(http.Handler) http.Handler

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{BaseRouter: r})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{BaseURL: baseURL, BaseRouter: r})
}

// HandlerWithOptions mounts the routes of every generated package on one router.
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter
	if r == nil {
		r = chi.NewRouter()
	}
{{- range .Groups}}
	{{.Alias}}.HandlerWithOptions(si, {{.Alias}}.ChiServerOptions{
		BaseURL:          options.BaseURL,
		BaseRouter:       r,
		Middlewares:      middlewares[{{.Alias}}.MiddlewareFunc](options.Middlewares),
		ErrorHandlerFunc: options.ErrorHandlerFunc,
	}){{end}}
	return r
}

func middlewares[T ~func(http.Handler) http.Handler](in []MiddlewareFunc) []T {
	out := make([]T, len(in))
	for i, m := range in {
		out[i] = T(m)
	}
	return out
}
`))

func writeAggregator(file, pkg string, groups []*group) error {
	sorted := append([]*group(nil), groups...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Import < sorted[j].Import })
	var buf bytes.Buffer
	if err := aggregator.Execute(&buf, map[string]any{"Package": pkg, "Groups": sorted}); err != nil {
		return err
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting %s: %w", file, err)
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, src, 0o644)
}

// removeGenerated deletes the generated files named name below dir (the
// aggregator included) and the directories left empty.
func removeGenerated(dir, name string) error {
	var dirs []string
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, p)
			return nil
		}
		if d.Name() == name {
			return os.Remove(p)
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for i := len(dirs) - 1; i > 0; i-- { // deepest first, keep dir itself
		if entries, err := os.ReadDir(dirs[i]); err == nil && len(entries) == 0 {
			_ = os.Remove(dirs[i])
		}
	}
	return nil
}

// packageName turns a folder name such as "Headless Identity API Demo" into a
// Go package name ("headlessidentityapidemo", or "headless_identity_api_demo"
// keeping separators).
func packageName(folder string, keepSeparators bool) string {
	var b strings.Builder
	separator := false
	for _, r := range strings.ToLower(folder) {
		if r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			if separator && keepSeparators && b.Len() > 0 {
				b.WriteByte('_')
			}
			separator = false
			b.WriteRune(r)
		} else {
			separator = true
		}
	}
	name := b.String()
	switch {
	case name == "":
		return "pkg"
	case unicode.IsDigit(rune(name[0])):
		return "p" + name
	case token.IsKeyword(name):
		return name + "api"
	}
	return name
}

// alias joins package path segments into a camel-case identifier that cannot
// clash with the aggregator's own imports.
func alias(segs []string) string {
	var b strings.Builder
	for i, seg := range segs {
		for j, word := range strings.Split(seg, "_") {
			if i > 0 || j > 0 {
				word = strings.ToUpper(word[:1]) + word[1:]
			}
			b.WriteString(word)
		}
	}
	switch name := b.String(); name {
	case "http", "chi":
		return name + "API"
	default:
		return name
	}
}

// importPath returns the Go import path of dir from the enclosing go.mod.
func importPath(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for root := abs; ; root = filepath.Dir(root) {
		raw, err := os.ReadFile(filepath.Join(root, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(raw), "\n") {
				if mod, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
					rel, err := filepath.Rel(root, abs)
					if err != nil {
						return "", err
					}
					return path.Join(strings.Trim(strings.TrimSpace(mod), `"`), filepath.ToSlash(rel)), nil
				}
			}
			return "", fmt.Errorf("%s/go.mod has no module directive", root)
		}
		if filepath.Dir(root) == root {
			return "", fmt.Errorf("no go.mod above %s", abs)
		}
	}
}

// withPaths returns a shallow copy of the document mapping with paths replaced.
func withPaths(doc, paths *yaml.Node) *yaml.Node {
	out := &yaml.Node{Kind: yaml.MappingNode}
	for i := 0; i+1 < len(doc.Content); i += 2 {
		value := doc.Content[i+1]
		if doc.Content[i].Value == "paths" {
			value = paths
		}
		out.Content = append(out.Content, doc.Content[i], value)
	}
	return out
}

// withScalars returns a copy of a mapping with the given keys set to string values.
func withScalars(mapping *yaml.Node, kv ...string) *yaml.Node {
	out := &yaml.Node{Kind: yaml.MappingNode, Content: append([]*yaml.Node(nil), mapping.Content...)}
	for k := 0; k+1 < len(kv); k += 2 {
		value := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: kv[k+1]}
		replaced := false
		for i := 0; i+1 < len(out.Content); i += 2 {
			if out.Content[i].Value == kv[k] {
				out.Content[i+1], replaced = value, true
			}
		}
		if !replaced {
			out.Content = append(out.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: kv[k]}, value)
		}
	}
	return out
}

func readYAML(file string) (*yaml.Node, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", file, err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s is not a YAML mapping", file)
	}
	return doc.Content[0], nil
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func scalar(mapping *yaml.Node, key string) string {
	if v := mappingValue(mapping, key); v != nil && v.Kind == yaml.ScalarNode {
		return v.Value
	}
	return ""
}

func writeYAML(file string, node *yaml.Node) error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return fmt.Errorf("encoding %s: %w", file, err)
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return os.WriteFile(file, buf.Bytes(), 0o644)
}
