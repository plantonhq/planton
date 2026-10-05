package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/explain"
)

// catalogPathPrefix roots every catalog-kind proto file; the provider is
// the path segment right after it. Deriving both the provider and the output
// directory from the descriptor's source path keeps this generator agnostic
// to the version segment (v1alpha1 today, later channels tomorrow) -- never
// reconstruct these paths from kind metadata.
const catalogPathPrefix = "catalog/"

// commonsContent is the catalog-commons page (the manifest grammar shared by
// every kind, stated once). It ships through the generator so the freshness
// gate owns it like every other generated file; edit commons.md here and
// regenerate.
//
//go:embed commons.md
var commonsContent string

// Summary is one generation run's full account: the files to write plus
// every degradation, so a caller can render an honest report and CI can fail
// on real catalog bugs instead of silently shipping thinner pages.
type Summary struct {
	// Files maps repo-relative output paths (catalog/.../reference.md and the
	// catalog-level index/graph/commons files) to rendered content.
	Files map[string]string
	// MissingManifests lists kinds with no e2e/manifest.yaml; their
	// pages render without an Example section.
	MissingManifests []string
	// InvalidManifests lists kinds whose hack manifest failed loading or
	// protovalidation. Each entry is a REAL catalog bug: the page renders
	// without the broken example, and the run must be reported as failed.
	InvalidManifests []InvalidManifest
}

// InvalidManifest names one hack manifest that failed to load or validate.
type InvalidManifest struct {
	Kind string
	Err  error
}

// kindEntry carries one kind through the generation passes.
type kindEntry struct {
	name string
	// protoDir is the descriptor-derived source directory
	// (catalog/<provider>/<kind>/<version>).
	protoDir string
	provider string
	report   *explain.Report
	example  string
	// hasGuide records whether an authored GUIDE.md sits beside the kind's
	// protos -- surfaced on the page head and in the indexes, so guide
	// coverage is a one-glance answer.
	hasGuide bool
}

// graphEdge is one foreign-key edge for the catalog graph: from's field can
// read target on to through a valueFrom reference.
type graphEdge struct {
	From   string
	Field  string
	To     string
	Target string
}

// Generate renders the reference page for every catalog kind plus the
// catalog-level files (per-provider indexes, the root index, the
// foreign-key graph, the commons page). It touches nothing on disk except
// reads; writing is the caller's decision so tests can byte-compare in
// memory.
//
// Generation is always whole-catalog: a kind's Referenced By section and
// every catalog-level file depend on every other kind's schema, so a scoped
// run could never be trusted to leave the committed tree consistent.
func Generate(repoRoot string) (*Summary, error) {
	engine := explain.DefaultEngine()
	summary := &Summary{Files: map[string]string{}}

	// Pass 1: explain every catalog kind. KindNames is sorted, so every
	// derived artifact inherits a deterministic order.
	var entries []kindEntry
	for _, name := range explain.KindNames() {
		res, err := explain.ResolveKindName(name)
		if err != nil {
			return nil, errors.Wrapf(err, "resolve kind %s", name)
		}
		protoDir := filepath.Dir(res.Message.ParentFile().Path())
		if !strings.HasPrefix(protoDir, catalogPathPrefix) {
			continue
		}
		provider, _, _ := strings.Cut(strings.TrimPrefix(protoDir, catalogPathPrefix), "/")
		if strings.HasPrefix(provider, "_") {
			// Underscore providers (_test) hold registered test-scaffolding
			// kinds, not catalog knowledge.
			continue
		}

		report, err := engine.Explain(res, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "explain %s", name)
		}

		entry := kindEntry{name: name, protoDir: protoDir, provider: provider, report: report}
		// protoDir is the version directory (the versioned contract);
		// the living kind -- GUIDE.md, README.md, e2e/, iac/ -- sits
		// one level up at the kind root.
		kindPath := filepath.Dir(filepath.Join(repoRoot, protoDir))
		entry.hasGuide = hasGuide(kindPath)
		switch example, err := loadValidatedExample(res, kindPath); {
		case err != nil:
			summary.InvalidManifests = append(summary.InvalidManifests, InvalidManifest{Kind: name, Err: err})
		case example == "":
			summary.MissingManifests = append(summary.MissingManifests, name)
		default:
			entry.example = example
		}
		entries = append(entries, entry)
	}

	// Pass 2: the catalog-wide edge list, projected two ways -- the graph
	// file and each target kind's inbound (Referenced By) edges.
	var edges []graphEdge
	inbound := map[string][]explain.InboundRef{}
	for _, entry := range entries {
		for _, edge := range explain.ReferenceEdges(entry.report) {
			edges = append(edges, graphEdge{
				From:   entry.name,
				Field:  edge.FieldPath,
				To:     edge.Kind,
				Target: edge.TargetFieldPath,
			})
			inbound[edge.Kind] = append(inbound[edge.Kind], explain.InboundRef{
				Kind:            entry.name,
				FieldPath:       edge.FieldPath,
				TargetFieldPath: edge.TargetFieldPath,
			})
		}
	}

	// Pass 3: render every page and the catalog-level files.
	for _, entry := range entries {
		kindPath := filepath.Dir(filepath.Join(repoRoot, entry.protoDir))
		opts := explain.MarkdownOptions{
			ExampleYAML:  entry.example,
			ReferencedBy: inbound[entry.name],
			SeeAlso:      seeAlso(kindPath),
			HasGuide:     entry.hasGuide,
		}
		summary.Files[filepath.Join(entry.protoDir, "reference.md")] = explain.RenderMarkdown(entry.report, opts)
	}
	renderProviderIndexes(summary, entries)
	renderRootIndex(summary, entries, repoRoot)
	renderGraph(summary, edges)
	summary.Files[catalogPath("reference-commons.md")] = commonsContent

	return summary, nil
}

// catalogPath places a catalog-level file in the catalog's _docs/ home,
// beside the provider directories (which, by the catalog invariant, are the
// only non-underscore entries at the catalog root).
func catalogPath(name string) string {
	return filepath.Join(catalogPathPrefix, "_docs", name)
}

// generatedBanner is the do-not-edit header every generated markdown file
// opens with. sourceHint names where the content actually comes from, so a
// wrong fact gets fixed at its source instead of on the page.
func generatedBanner(b *strings.Builder, sourceHint string) {
	b.WriteString("> Generated by `make generate-reference` -- do not edit by hand. ")
	b.WriteString(sourceHint)
	b.WriteString("\n\n")
}

// renderProviderIndexes emits one kind index per provider: the entry point
// an agent reads to map kind names to reference pages.
func renderProviderIndexes(summary *Summary, entries []kindEntry) {
	byProvider := map[string][]kindEntry{}
	for _, entry := range entries {
		byProvider[entry.provider] = append(byProvider[entry.provider], entry)
	}
	for provider, kinds := range byProvider {
		var b strings.Builder
		fmt.Fprintf(&b, "# %s Kinds -- Reference Index\n\n", provider)
		generatedBanner(&b, "Kind facts come from the protobuf schemas.")
		fmt.Fprintf(&b, "%d kinds. Shared manifest grammar and the search grammar of every page:\n[reference-commons.md](../_docs/reference-commons.md). Cross-kind wiring:\n[reference-graph.yaml](../_docs/reference-graph.yaml).\n\n", len(kinds))
		// The Example and Guide columns double as coverage x-rays: a blank
		// Example is a kind missing its hack manifest, a blank Guide is a
		// kind nobody has written wisdom for yet.
		b.WriteString("| Kind | Purpose | Example | Guide |\n")
		b.WriteString("|---|---|---|---|\n")
		for _, entry := range kinds {
			relPage := strings.TrimPrefix(entry.protoDir, catalogPathPrefix+provider+"/") + "/reference.md"
			example := ""
			if entry.example != "" {
				example = "yes"
			}
			guide := ""
			if entry.hasGuide {
				guide = "yes"
			}
			fmt.Fprintf(&b, "| [%s](%s) | %s | %s | %s |\n",
				entry.name, relPage, indexCell(firstSentence(entry.report.Doc)), example, guide)
		}
		summary.Files[filepath.Join(catalogPathPrefix, provider, "reference-index.md")] = b.String()
	}
}

// renderRootIndex emits the catalog's front door: where everything lives and
// how to read it. It probes the repo for the authored catalog-level wisdom
// (the catalog GUIDE.md and the patterns library) with the same
// exists-or-no-mention idiom the kind pages use for their guides.
func renderRootIndex(summary *Summary, entries []kindEntry, repoRoot string) {
	providers := map[string]int{}
	withExample := map[string]int{}
	withGuide := map[string]int{}
	for _, entry := range entries {
		providers[entry.provider]++
		if entry.example != "" {
			withExample[entry.provider]++
		}
		if entry.hasGuide {
			withGuide[entry.provider]++
		}
	}
	names := make([]string, 0, len(providers))
	for provider := range providers {
		names = append(names, provider)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("# Infra Catalog -- Reference Index\n\n")
	generatedBanner(&b, "Kind facts come from the protobuf schemas.")
	fmt.Fprintf(&b, "%d kinds across %d providers. Every kind has a `reference.md` co-located\nwith its protos: the complete spec field reference, validation rules,\noutputs, cross-kind references (both directions), and a validated example.\n\nStart with:\n\n", len(entries), len(names))
	docsRoot := filepath.Join(repoRoot, catalogPathPrefix, "_docs")
	if _, err := os.Stat(filepath.Join(docsRoot, "GUIDE.md")); err == nil {
		b.WriteString("- [GUIDE.md](GUIDE.md) -- how to use this catalog: finding compatible\n  alternatives when the asked-for software has no kind of its own, and the\n  conventions that span providers.\n")
	}
	b.WriteString("- [reference-commons.md](reference-commons.md) -- the manifest grammar shared\n  by every kind (envelope, metadata, value/valueFrom, fieldPath spelling)\n  and the search grammar for reading these pages.\n")
	b.WriteString("- [reference-graph.yaml](reference-graph.yaml) -- every foreign-key edge in\n  the catalog, for wiring resources and planning architectures.\n")
	if _, err := os.Stat(filepath.Join(repoRoot, catalogPathPrefix, "_patterns", "README.md")); err == nil {
		b.WriteString("- [_patterns/](../_patterns/) -- authored architecture patterns: named\n  compositions of multiple kinds with validated wiring, and the trade-offs\n  behind them.\n")
	}
	b.WriteString("- The per-provider indexes below, which link every kind's page.\n\n")
	b.WriteString("| Provider | Kinds | With Example | With Guide |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, provider := range names {
		fmt.Fprintf(&b, "| [%s](../%s/reference-index.md) | %d | %d | %d |\n",
			provider, provider, providers[provider], withExample[provider], withGuide[provider])
	}
	summary.Files[catalogPath("reference-index.md")] = b.String()
}

// renderGraph emits the catalog's foreign-key edge list. Hand-built line by
// line: YAML libraries do not promise deterministic serialization, and this
// file is byte-compared by the freshness gate.
func renderGraph(summary *Summary, edges []graphEdge) {
	var b strings.Builder
	b.WriteString("# Generated by `make generate-reference` -- do not edit by hand.\n")
	b.WriteString("# Every foreign-key edge in the catalog: `from`'s `field` accepts a valueFrom\n")
	b.WriteString("# reference reading `target` on a `to` resource. Field paths are quoted\n")
	b.WriteString("# verbatim from the schemas ([] marks list elements, .* map values).\n")
	b.WriteString("edges:\n")
	for _, edge := range edges {
		fmt.Fprintf(&b, "  - from: %q\n    field: %q\n    to: %q\n    target: %q\n",
			edge.From, edge.Field, edge.To, edge.Target)
	}
	summary.Files[catalogPath("reference-graph.yaml")] = b.String()
}

// firstSentence truncates documentation to its first sentence, collapsed
// onto one line -- the right size for an index cell.
func firstSentence(doc string) string {
	text := strings.Join(strings.Fields(doc), " ")
	if i := strings.Index(text, ". "); i >= 0 {
		return text[:i+1]
	}
	return text
}

// indexCell makes arbitrary prose safe inside a markdown table row.
func indexCell(text string) string {
	return strings.ReplaceAll(text, "|", "\\|")
}

// loadValidatedExample reads the kind's base test manifest
// (e2e/manifest.yaml at the kind root -- the same manifest the E2E
// framework deploys) and admits it as the page's example only after every
// document round-trips into its typed message and passes protovalidate. A
// manifest may legitimately hold several documents (a kind deployed with its
// prerequisites), so documents are split and validated individually -- the
// single-manifest refusal belongs to one-resource-per-run commands, not to
// catalog tooling. At least one document must declare the page's own kind.
// An empty string with nil error means no manifest exists (graceful
// degradation); an error means the manifest exists and is broken.
func loadValidatedExample(res explain.Resource, kindPath string) (string, error) {
	path := filepath.Join(kindPath, "e2e", "manifest.yaml")
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", errors.Wrapf(err, "read %s", path)
	}
	docs, err := manifest.SplitDocuments(raw)
	if err != nil {
		return "", errors.Wrapf(err, "split %s", path)
	}
	kindPresent := false
	for i, doc := range docs {
		loaded, err := manifest.LoadManifestBytes(doc, path)
		if err != nil {
			return "", errors.Wrapf(err, "load %s document %d", path, i+1)
		}
		if loaded.ProtoReflect().Descriptor().FullName() == res.Message.FullName() {
			kindPresent = true
		}
		if err := manifest.ValidateLoaded(loaded); err != nil {
			return "", errors.Wrapf(err, "validate %s document %d", path, i+1)
		}
	}
	if !kindPresent {
		return "", errors.Errorf("%s declares no document of kind %s", path, res.Message.FullName())
	}
	return string(raw), nil
}

// hasGuide reports whether an authored GUIDE.md sits at the kind root.
// Guide presence flows into the page head and the indexes; the freshness gate
// makes adding or removing a guide without regenerating a visible failure
// instead of silent staleness.
func hasGuide(kindPath string) bool {
	_, err := os.Stat(filepath.Join(kindPath, "GUIDE.md"))
	return err == nil
}

// seeAlso links the page to the kind's hand-written prose when it exists.
// Reference pages link to essays, never duplicate them. The README sits at
// the kind root, one level above the version dir that holds the page.
func seeAlso(kindPath string) []explain.MarkdownLink {
	var links []explain.MarkdownLink
	if _, err := os.Stat(filepath.Join(kindPath, "README.md")); err == nil {
		links = append(links, explain.MarkdownLink{Title: "Overview", Path: "../README.md"})
	}
	return links
}

// SortedPaths returns the output paths in stable order for writing and
// reporting.
func (s *Summary) SortedPaths() []string {
	paths := make([]string, 0, len(s.Files))
	for path := range s.Files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
