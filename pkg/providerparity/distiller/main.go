//go:build !codegen
// +build !codegen

// Distills Terraform provider schemas into pkg/providerparity's committed
// artifacts.
//
// The provider is the only authority on its own configurable surface, and it
// is self-describing: `providers schema -json` emits every resource, argument,
// type, and flag at an exact released version. Committing that output whole
// would carry descriptions and data sources the parity accounting never
// reads; this tool resolves the pin in a throwaway OpenTofu workspace,
// reduces the schema to the configurable surface, and writes one
// deterministic gzipped artifact per provider (see
// providerparity.WriteSchemaFile for the determinism and replace-on-repin
// contract).
//
// Run via Makefile: make generate-provider-schemas
//
//	go run ./pkg/providerparity/distiller \
//	    --out-dir pkg/providerparity/schemas \
//	    --provider 'google=hashicorp/google@8.3.0' \
//	    --provider 'google-beta=hashicorp/google-beta@8.3.0'
//
// Bumping a pin is a one-line change to the Makefile invocation followed by a
// re-run: the artifact for the same provider is replaced, never accumulated,
// and the resulting parity-check failures ARE the migration work list.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/plantonhq/planton/pkg/providerparity"
)

// pin is one provider to distill: the local/artifact name, the registry
// source address, and the version constraint the workspace resolves.
type pin struct {
	name       string
	source     string
	constraint string
}

type pinList []pin

func (p *pinList) String() string {
	var parts []string
	for _, x := range *p {
		parts = append(parts, fmt.Sprintf("%s=%s@%s", x.name, x.source, x.constraint))
	}
	return strings.Join(parts, ",")
}

func (p *pinList) Set(v string) error {
	name, rest, ok := strings.Cut(v, "=")
	if !ok {
		return fmt.Errorf("provider %q: want name=source@constraint", v)
	}
	source, constraint, ok := strings.Cut(rest, "@")
	if !ok {
		return fmt.Errorf("provider %q: want name=source@constraint", v)
	}
	*p = append(*p, pin{name: name, source: source, constraint: constraint})
	return nil
}

func main() {
	var outDir string
	var pins pinList
	flag.StringVar(&outDir, "out-dir", providerparity.DefaultSchemaDir, "directory for the committed artifacts")
	flag.Var(&pins, "provider", "provider to distill as name=source@constraint (repeatable)")
	flag.Parse()
	if len(pins) == 0 {
		fatal("at least one --provider is required")
	}

	workDir, err := os.MkdirTemp("", "providerparity-distill-*")
	if err != nil {
		fatal("creating workspace: %v", err)
	}
	defer os.RemoveAll(workDir)

	if err := os.WriteFile(filepath.Join(workDir, "providers.tf"), []byte(workspaceConfig(pins)), 0o644); err != nil {
		fatal("writing workspace config: %v", err)
	}
	if out, err := tofu(workDir, "init", "-backend=false", "-input=false", "-no-color"); err != nil {
		fatal("tofu init: %v\n%s", err, out)
	}
	versions, err := resolvedVersions(filepath.Join(workDir, ".terraform.lock.hcl"))
	if err != nil {
		fatal("%v", err)
	}
	schemaJSON, err := tofu(workDir, "providers", "schema", "-json")
	if err != nil {
		fatal("tofu providers schema: %v\n%s", err, schemaJSON)
	}

	var raw rawSchemas
	if err := json.Unmarshal(schemaJSON, &raw); err != nil {
		fatal("parsing schema JSON: %v", err)
	}

	for _, p := range pins {
		body := raw.forSource(p.source)
		if body == nil {
			fatal("schema output carries no provider matching source %q", p.source)
		}
		version, ok := versions[p.source]
		if !ok {
			fatal("lock file carries no resolved version for source %q", p.source)
		}
		schema := &providerparity.Schema{
			Provider:       p.name,
			Source:         p.source,
			Version:        version,
			ProviderConfig: distillBlock(body.Provider.Block),
			Resources:      map[string]*providerparity.Block{},
		}
		for name, rs := range body.ResourceSchemas {
			schema.Resources[name] = distillBlock(rs.Block)
		}
		path, err := providerparity.WriteSchemaFile(outDir, schema)
		if err != nil {
			fatal("writing artifact for %s: %v", p.name, err)
		}
		fmt.Printf("distilled %s %s: %d resources -> %s\n", p.name, version, len(schema.Resources), path)
	}
}

// workspaceConfig renders the throwaway workspace's required_providers block.
func workspaceConfig(pins pinList) string {
	var b strings.Builder
	b.WriteString("terraform {\n  required_providers {\n")
	for _, p := range pins {
		fmt.Fprintf(&b, "    %s = {\n      source  = %q\n      version = %q\n    }\n", p.name, p.source, p.constraint)
	}
	b.WriteString("  }\n}\n")
	return b.String()
}

func tofu(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("tofu", args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// lockProviderRe pulls each provider block's source and resolved version out
// of .terraform.lock.hcl -- the one place the exact release the constraint
// resolved to is recorded.
var lockProviderRe = regexp.MustCompile(`(?s)provider\s+"([^"]+)"\s*\{.*?version\s*=\s*"([^"]+)"`)

func resolvedVersions(lockPath string) (map[string]string, error) {
	content, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", lockPath, err)
	}
	versions := map[string]string{}
	for _, m := range lockProviderRe.FindAllStringSubmatch(string(content), -1) {
		// Lock addresses are registry-qualified (registry.opentofu.org/hashicorp/google);
		// key by the trailing source address the pins use.
		addr := m[1]
		if i := strings.Index(addr, "/"); i >= 0 {
			addr = addr[i+1:]
		}
		versions[addr] = m[2]
	}
	return versions, nil
}

// rawSchemas mirrors the slice of `providers schema -json` output the
// distillation reads. Everything else (descriptions, data sources, functions)
// is deliberately not modeled.
type rawSchemas struct {
	ProviderSchemas map[string]*rawProviderBody `json:"provider_schemas"`
}

func (r rawSchemas) forSource(source string) *rawProviderBody {
	for key, body := range r.ProviderSchemas {
		if strings.HasSuffix(key, "/"+source) {
			return body
		}
	}
	return nil
}

type rawProviderBody struct {
	// Provider is the provider's own configuration block -- the arguments a
	// `provider "<name>" {}` block accepts. Part of the parity surface: the
	// provider-config accounting reads it exactly like resource blocks.
	Provider        rawResourceSchema            `json:"provider"`
	ResourceSchemas map[string]rawResourceSchema `json:"resource_schemas"`
}

type rawResourceSchema struct {
	Block rawBlock `json:"block"`
}

type rawBlock struct {
	Attributes map[string]rawAttribute `json:"attributes"`
	BlockTypes map[string]rawBlockType `json:"block_types"`
	Deprecated bool                    `json:"deprecated"`
}

type rawAttribute struct {
	Type       json.RawMessage `json:"type"`
	Required   bool            `json:"required"`
	Optional   bool            `json:"optional"`
	Computed   bool            `json:"computed"`
	Sensitive  bool            `json:"sensitive"`
	Deprecated bool            `json:"deprecated"`
}

type rawBlockType struct {
	NestingMode string   `json:"nesting_mode"`
	Block       rawBlock `json:"block"`
}

func distillBlock(rb rawBlock) *providerparity.Block {
	b := &providerparity.Block{Deprecated: rb.Deprecated}
	if len(rb.Attributes) > 0 {
		b.Attributes = map[string]*providerparity.Attribute{}
		for name, a := range rb.Attributes {
			b.Attributes[name] = &providerparity.Attribute{
				Type:       a.Type,
				Required:   a.Required,
				Optional:   a.Optional,
				Computed:   a.Computed,
				Sensitive:  a.Sensitive,
				Deprecated: a.Deprecated,
			}
		}
	}
	if len(rb.BlockTypes) > 0 {
		b.Blocks = map[string]*providerparity.NestedBlock{}
		for name, bt := range rb.BlockTypes {
			b.Blocks[name] = &providerparity.NestedBlock{
				NestingMode: bt.NestingMode,
				Block:       distillBlock(bt.Block),
			}
		}
	}
	return b
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "providerparity distiller: "+format+"\n", args...)
	os.Exit(1)
}
