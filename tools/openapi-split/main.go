// Command openapi-split splits an OpenAPI document into one file per tag
// (Postman/Bruno folder) and rewrites the main document to $ref them.
//
// postman-to-openapi tags every operation with its folder chain, e.g.
// "Auth > Headless Identity API Demo". Each path item is moved to the file of
// its first operation's tag, mirroring the Bruno folder structure:
//
//	Auth                              -> paths/Auth.yaml
//	Auth > Headless Identity API Demo -> paths/Auth/Headless Identity API Demo.yaml
//
// A folder file maps paths to path items, and the main document references
// them, e.g. `$ref: paths/Auth.yaml#/~1services~1oauth2~1token`. OpenAPI 3.0
// only allows a $ref for a whole path item, so a path whose methods span
// several folders goes to the folder of its first method.
//
// Usage: openapi-split <openapi.yaml> [paths-dir]
//
// The document is rewritten in place. paths-dir defaults to "paths" next to
// it; it is regenerated from scratch on every run.
package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var methods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true, "trace": true,
}

// folderFile is the YAML document of one folder, keeping the paths in order.
type folderFile struct {
	file  string // path relative to the paths directory, e.g. "Auth/Query.yaml"
	paths *yaml.Node
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "usage: openapi-split <openapi.yaml> [paths-dir]")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2:]...); err != nil {
		fmt.Fprintln(os.Stderr, "openapi-split:", err)
		os.Exit(1)
	}
}

func run(specPath string, args ...string) error {
	pathsDir := filepath.Join(filepath.Dir(specPath), "paths")
	if len(args) > 0 {
		pathsDir = args[0]
	}
	refBase, err := filepath.Rel(filepath.Dir(specPath), pathsDir)
	if err != nil {
		return err
	}

	raw, err := os.ReadFile(specPath)
	if err != nil {
		return err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parsing %s: %w", specPath, err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return fmt.Errorf("%s is not an OpenAPI document", specPath)
	}
	paths := mappingValue(doc.Content[0], "paths")
	if paths == nil || paths.Kind != yaml.MappingNode {
		return fmt.Errorf("%s has no paths", specPath)
	}

	var order []string
	folders := map[string]*folderFile{}
	for i := 0; i < len(paths.Content); i += 2 {
		key, item := paths.Content[i], paths.Content[i+1]
		if mappingValue(item, "$ref") != nil {
			return fmt.Errorf("%s is already split (path %s is a $ref)", specPath, key.Value)
		}
		file := folderFileName(firstTag(item))
		folder, ok := folders[file]
		if !ok {
			folder = &folderFile{file: file, paths: &yaml.Node{Kind: yaml.MappingNode}}
			folders[file] = folder
			order = append(order, file)
		}
		folder.paths.Content = append(folder.paths.Content, key, item)
		paths.Content[i+1] = refNode(refBase, file, key.Value)
	}

	if err := os.RemoveAll(pathsDir); err != nil {
		return err
	}
	for _, file := range order {
		if err := writeYAML(filepath.Join(pathsDir, filepath.FromSlash(file)), folders[file].paths); err != nil {
			return err
		}
	}
	if err := writeYAML(specPath, &doc); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Split %d paths into %d files under %s\n", len(paths.Content)/2, len(order), pathsDir)
	return nil
}

// mappingValue returns the value of key in a YAML mapping, or nil.
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

// firstTag returns the first tag of the first operation of a path item.
func firstTag(item *yaml.Node) string {
	for i := 0; i+1 < len(item.Content); i += 2 {
		if !methods[item.Content[i].Value] {
			continue
		}
		if tags := mappingValue(item.Content[i+1], "tags"); tags != nil && len(tags.Content) > 0 {
			return tags.Content[0].Value
		}
	}
	return ""
}

// folderFileName turns a folder chain such as "Auth > Headless Identity API Demo"
// into "Auth/Headless Identity API Demo.yaml".
func folderFileName(tag string) string {
	var segments []string
	for _, segment := range strings.Split(tag, " > ") {
		segment = strings.TrimSpace(strings.NewReplacer("/", "-", "\\", "-").Replace(segment))
		if segment != "" && segment != "." && segment != ".." {
			segments = append(segments, segment)
		}
	}
	if len(segments) == 0 {
		segments = []string{"default"}
	}
	return strings.Join(segments, "/") + ".yaml"
}

// refNode builds `$ref: <base>/<file>#/<json pointer to the path>`, percent-encoded
// as a URI reference (folder names contain spaces, paths contain braces).
func refNode(base, file, path string) *yaml.Node {
	segments := strings.Split(filepath.ToSlash(filepath.Join(base, filepath.FromSlash(file))), "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	pointer := strings.NewReplacer("~", "~0", "/", "~1").Replace(path)
	ref := strings.Join(segments, "/") + "#/" + url.PathEscape(pointer)
	return &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{
		{Kind: yaml.ScalarNode, Value: "$ref"},
		{Kind: yaml.ScalarNode, Value: ref},
	}}
}

func writeYAML(path string, node *yaml.Node) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}
	if err := enc.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
