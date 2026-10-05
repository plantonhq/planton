// pkg/catalogkindreflect/cmd/genkindmap/main.go
// Generates pkg/catalogkindreflect/kind_map_gen.go.
//
// IMPORTANT: This must be run with the "codegen" build tag:
//
//	go run -tags codegen ./pkg/catalogkindreflect/codegen
//
// Why the build tag is needed:
//   - This generator creates kind_map_gen.go which defines ToMessageMap
//   - Other files in pkg/catalogkindreflect (e.g., new_instance.go) depend on ToMessageMap
//   - Without the build tag, "go run" would try to compile ALL pkg/catalogkindreflect files
//   - This causes a chicken-and-egg problem: can't compile without ToMessageMap,
//     but ToMessageMap doesn't exist until the generator runs
//   - The "codegen" build tag excludes files marked with "//go:build !codegen"
//
// Run via Makefile:  make generate-catalog-kind-map
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

// pascalFromSnake converts "digital_ocean" → "DigitalOcean".
func pascalFromSnake(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

// fixDigitCase: Neo4j → Neo4J; CloudflareKVNamespace → CloudflareKvNamespace
// (matches message names in generated code)
func fixDigitCase(s string) string {
	r := []rune(s)

	// 1. upper‑case the letter that follows any digit (existing rule)
	for i := 0; i < len(r)-1; i++ {
		if r[i] >= '0' && r[i] <= '9' && r[i+1] >= 'a' && r[i+1] <= 'z' {
			r[i+1] -= 'a' - 'A'
		}
	}

	// 2. inside runs of ≥2 consecutive upper‑case letters,
	//    lower‑case every rune except the first so that "KVNamespace" → "KvNamespace".
	for i := 1; i < len(r)-1; i++ {
		if r[i-1] >= 'A' && r[i-1] <= 'Z' &&
			r[i] >= 'A' && r[i] <= 'Z' &&
			r[i+1] >= 'A' && r[i+1] <= 'Z' {
			r[i] += 'a' - 'A'
		}
	}

	return string(r)
}

// lowerNoSep: AwsAlb → awsalb
func lowerNoSep(s string) string { return strings.ToLower(strings.ReplaceAll(s, "_", "")) }

// -----------------------------------------------------------------------------
// types for template
// -----------------------------------------------------------------------------

type importInfo struct{ Alias, Path string }
type entry struct {
	KindConst, Alias, MessageType string
}

// -----------------------------------------------------------------------------
// main
// -----------------------------------------------------------------------------

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	provEntries := map[string][]entry{} // provider raw name ("digital_ocean") → []entry

	imports := []importInfo{}
	aliasByPath := map[string]string{}

	for _, catalogKind := range catalogkindreflect.KindsList() {
		provider := catalogkindreflect.GetProvider(catalogKind)
		if provider == catalogkind.CatalogProvider_catalog_provider_unspecified {
			// skip unspecified
			continue
		}

		kindName := catalogKind.String() // e.g. AwsAlb

		provRaw := provider.String() // "digital_ocean" or "_test"
		// Keep leading underscore for test provider, remove underscores for others
		provSlug := provRaw
		if !strings.HasPrefix(provRaw, "_") {
			provSlug = strings.ReplaceAll(provRaw, "_", "") // "digitalocean"
		}

		lowerKind := lowerNoSep(kindName) // awsalb

		// The version segment of a kind's package path follows its declared
		// version in the registry — never a literal. A kind without a valid
		// version fails generation loudly instead of mapping to a package
		// that does not exist.
		versionDir, err := catalogkindreflect.KindVersion(catalogKind)
		if err != nil {
			return errors.Wrapf(err, "cannot derive the package path for %s", kindName)
		}

		importAlias := lowerKind + versionDir // awsalbv1

		// all providers use flat structure
		importPath := fmt.Sprintf(
			"github.com/plantonhq/planton/catalog/%s/%s/%s",
			provSlug, lowerKind, versionDir)

		// Skip kinds whose API packages have not been implemented yet.
		// The package directory is created when proto files are added and
		// buf generate produces the corresponding .pb.go files. Until then,
		// including the import would break both Gazelle resolution and Go
		// compilation.
		pkgDir := filepath.Join("catalog", provSlug, lowerKind, versionDir)
		if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "skipping %s: package dir %s not found\n", kindName, pkgDir)
			continue
		}

		alias := uniqueAlias(importPath, importAlias, &imports, aliasByPath)
		provEntries[provRaw] = append(provEntries[provRaw], entry{
			KindConst:   kindName,
			Alias:       alias,
			MessageType: fixDigitCase(kindName),
		})
	}

	// deterministic output
	for _, list := range provEntries {
		sort.Slice(list, func(i, j int) bool { return list[i].KindConst < list[j].KindConst })
	}
	sort.Slice(imports, func(i, j int) bool { return imports[i].Alias < imports[j].Alias })

	// render
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, struct {
		Imports     []importInfo
		ProvEntries map[string][]entry
		Providers   []string
	}{
		Imports:     imports,
		ProvEntries: provEntries,
		Providers:   sortedKeys(provEntries),
	}); err != nil {
		return errors.Wrap(err, "execute template")
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		src = buf.Bytes() // keep raw on formatting failure
	}

	outPath := filepath.Join("pkg", "catalogkindreflect", "kind_map_gen.go")
	if err := os.WriteFile(outPath, src, 0o644); err != nil {
		return errors.Wrapf(err, "write %s", outPath)
	}
	fmt.Printf("created %s\n", outPath)
	return nil
}

func uniqueAlias(path, base string, imports *[]importInfo, seen map[string]string) string {
	if a, ok := seen[path]; ok {
		return a
	}
	alias := base
	for i := 1; aliasExists(*imports, alias); i++ {
		alias = fmt.Sprintf("%s_%d", base, i)
	}
	*imports = append(*imports, importInfo{alias, path})
	seen[path] = alias
	return alias
}

func aliasExists(imps []importInfo, a string) bool {
	for _, v := range imps {
		if v.Alias == a {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string][]entry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// -----------------------------------------------------------------------------
// template
// -----------------------------------------------------------------------------

var tpl = template.Must(template.New("").Funcs(template.FuncMap{
	"pascal": pascalFromSnake,
}).Parse(`// Code generated by pkg/catalogkindreflect/codegen; DO NOT EDIT.
//
// This file defines ToMessageMap which maps CatalogKind enums to protobuf message instances.
// To regenerate: make generate-catalog-kind-map
//
// IMPORTANT: Files that import ToMessageMap should use "//go:build !codegen" tag
// to prevent chicken-and-egg compilation issues during generation.
// See new_instance.go for an example.
package catalogkindreflect

import (
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
{{- range .Imports }}
	{{ .Alias }} "{{ .Path }}"
{{- end }}
)

func merge(maps ...map[catalogkind.CatalogKind]proto.Message) map[catalogkind.CatalogKind]proto.Message {
	out := make(map[catalogkind.CatalogKind]proto.Message)
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

{{/* provider maps */}}
{{- range $prov, $list := .ProvEntries }}
var Provider{{ pascal $prov }}Map = map[catalogkind.CatalogKind]proto.Message{
{{- range $list }}
	catalogkind.CatalogKind_{{ .KindConst }}: &{{ .Alias }}.{{ .MessageType }}{},
{{- end }}
}
{{ end }}

var ToMessageMap = merge(
{{- range .Providers }}
	Provider{{ pascal . }}Map,
{{- end }}
)
`))
