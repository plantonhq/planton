package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/explain"
)

// These are the guardrails for the catalog's AUTHORED layer -- the GUIDE.md
// files and the patterns library. Authored wisdom is allowed to age; it is
// never allowed to be silently schema-wrong or structurally lost. So the
// checks are structural, not editorial: placement, embedded-manifest
// validity, declared kind names, and link integrity. Editorial quality is
// the audit rules' job, not CI's.

// wisdomFile is one authored file with its repo-relative path and content.
type wisdomFile struct {
	relPath string
	content string
}

// wisdomFiles collects every authored wisdom file in the catalog: kind-level
// and catalog-root GUIDE.md files plus everything in the patterns library.
func wisdomFiles(t *testing.T, root string) []wisdomFile {
	t.Helper()
	catalogRoot := filepath.Join(root, catalogPathPrefix)
	var files []wisdomFile
	err := filepath.WalkDir(catalogRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		inPatterns := strings.Contains(path, string(filepath.Separator)+"_patterns"+string(filepath.Separator))
		if d.Name() != "GUIDE.md" && !inPatterns {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, wisdomFile{relPath: rel, content: string(raw)})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// registryKindDirs derives every registered kind's root the
// same way the generator does (the parent of the proto descriptor's source
// directory) -- the single source of truth for where a kind-level GUIDE.md
// may live.
func registryKindDirs(t *testing.T) map[string]bool {
	t.Helper()
	dirs := map[string]bool{}
	for _, name := range explain.KindNames() {
		res, err := explain.ResolveKindName(name)
		if err != nil {
			t.Fatalf("resolve kind %s: %v", name, err)
		}
		protoDir := filepath.Dir(res.Message.ParentFile().Path())
		if strings.HasPrefix(protoDir, catalogPathPrefix) {
			dirs[filepath.Dir(protoDir)] = true
		}
	}
	return dirs
}

// TestWisdomGuidePlacement pins where authored guides may live: at a
// registered kind's root (where the generated page links them) or
// at the catalog root. A guide anywhere else is invisible to every reader --
// a misplaced file, usually a typo'd directory or a kind that was renamed.
func TestWisdomGuidePlacement(t *testing.T) {
	root := repoRoot(t)
	kindEntries := registryKindDirs(t)
	catalogRootGuide := filepath.Join(catalogPathPrefix, "_docs", "GUIDE.md")

	for _, file := range wisdomFiles(t, root) {
		if filepath.Base(file.relPath) != "GUIDE.md" {
			continue
		}
		if file.relPath == catalogRootGuide {
			continue
		}
		if !kindEntries[filepath.Dir(file.relPath)] {
			t.Errorf("%s is not at a registered kind's root (and not the catalog-root guide) -- no generated page will ever link it", file.relPath)
		}
	}
}

// completeManifestDocs extracts every complete manifest (a YAML document
// declaring both apiVersion and kind) from a file's fenced yaml blocks.
// Partial snippets -- fragments without the envelope -- are exempt by
// construction: wisdom is allowed to quote fragments, but anything that
// presents itself as a deployable manifest must actually validate.
func completeManifestDocs(content string) []string {
	var docs []string
	for _, block := range regexp.MustCompile("(?s)```yaml\n(.*?)```").FindAllStringSubmatch(content, -1) {
		for _, doc := range strings.Split(block[1], "\n---\n") {
			hasAPIVersion := regexp.MustCompile(`(?m)^apiVersion:`).MatchString(doc)
			hasKind := regexp.MustCompile(`(?m)^kind:`).MatchString(doc)
			if hasAPIVersion && hasKind {
				docs = append(docs, doc)
			}
		}
	}
	return docs
}

// TestWisdomManifestsValidate protovalidates every complete manifest
// embedded in authored wisdom. An invalid example in a guide teaches every
// reader a broken shape -- the same law the generator enforces for hack
// manifests.
func TestWisdomManifestsValidate(t *testing.T) {
	root := repoRoot(t)
	for _, file := range wisdomFiles(t, root) {
		for i, doc := range completeManifestDocs(file.content) {
			loaded, err := manifest.LoadManifestBytes([]byte(doc), file.relPath)
			if err != nil {
				t.Errorf("%s embedded manifest #%d does not load: %v", file.relPath, i+1, err)
				continue
			}
			if err := manifest.ValidateLoaded(loaded); err != nil {
				t.Errorf("%s embedded manifest #%d fails validation: %v", file.relPath, i+1, err)
			}
		}
	}
}

// patternFrontmatterKinds parses the `kinds:` list from a pattern's
// frontmatter. Hand-rolled on purpose: the convention is a plain block list
// between the opening `---` fences, and a parser this small cannot drift
// from it.
func patternFrontmatterKinds(content string) ([]string, bool) {
	rest, ok := strings.CutPrefix(content, "---\n")
	if !ok {
		return nil, false
	}
	front, _, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return nil, false
	}
	var kinds []string
	inKinds := false
	for _, line := range strings.Split(front, "\n") {
		switch {
		case strings.HasPrefix(line, "kinds:"):
			inKinds = true
		case inKinds && strings.HasPrefix(strings.TrimSpace(line), "- "):
			kinds = append(kinds, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- ")))
		case inKinds && !strings.HasPrefix(line, " "):
			inKinds = false
		}
	}
	return kinds, true
}

// TestWisdomPatternKindsResolve pins the patterns convention: every pattern
// declares the kinds it composes in frontmatter, and every declared kind
// resolves in the registry. A stale kind name means the pattern quietly
// stopped describing anything deployable.
func TestWisdomPatternKindsResolve(t *testing.T) {
	root := repoRoot(t)
	for _, file := range wisdomFiles(t, root) {
		if filepath.Base(file.relPath) == "GUIDE.md" || filepath.Base(file.relPath) == "README.md" {
			continue
		}
		kinds, ok := patternFrontmatterKinds(file.content)
		if !ok || len(kinds) == 0 {
			t.Errorf("%s declares no kinds in frontmatter -- every pattern opens with `---`, a `kinds:` block list, `---`", file.relPath)
			continue
		}
		for _, kind := range kinds {
			if _, err := explain.ResolveKindName(kind); err != nil {
				t.Errorf("%s declares kind %q which does not resolve in the registry: %v", file.relPath, kind, err)
			}
		}
	}
}

// TestWisdomLinksResolve checks every relative link in authored wisdom
// points at a file that exists. Wisdom links reference pages, guides, and
// patterns liberally; a broken link strands the reader exactly where the
// next answer should have been.
func TestWisdomLinksResolve(t *testing.T) {
	root := repoRoot(t)
	linkPattern := regexp.MustCompile(`\]\(([^)]+)\)`)
	for _, file := range wisdomFiles(t, root) {
		for _, match := range linkPattern.FindAllStringSubmatch(file.content, -1) {
			target := match[1]
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") ||
				strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "#") {
				continue
			}
			target, _, _ = strings.Cut(target, "#")
			resolved := filepath.Join(root, filepath.Dir(file.relPath), target)
			if _, err := os.Stat(resolved); err != nil {
				t.Errorf("%s links %q which does not resolve: %v", file.relPath, match[1], err)
			}
		}
	}
}
