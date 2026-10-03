package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/plantonhq/planton/pkg/explain"
	"gopkg.in/yaml.v3"
)

// This is the freshness guardrail for the catalog-research eval bank
// (eval_questions.yaml). The bank's recorded answers are only useful while
// they are TRUE, so every machine-checkable assertion is re-verified here
// against the committed reference pack: cited kinds must still resolve
// through the kind registry, and every ground-truth pattern must still
// match its resolved file. A schema change that invalidates a recorded
// answer fails this test with the stale question named -- the same
// rot-proofing the drift gate gives the pack itself.
//
// Kinds are resolved to directories through the registry (never through
// stored paths), so version-segment renames in the apis tree never break
// the bank.

// evalCheck is one machine-checkable assertion. Exactly one of Kind,
// Provider, or Catalog selects the file class; Pattern is an RE2 regex
// matched against the whole file.
type evalCheck struct {
	// Kind + File select a file sitting beside the kind's protos
	// (reference.md or GUIDE.md), resolved through the kind registry.
	Kind string `yaml:"kind"`
	File string `yaml:"file"`
	// Provider selects that provider's reference-index.md.
	Provider string `yaml:"provider"`
	// Catalog selects a catalog-level file relative to the catalog root
	// (_docs/reference-graph.yaml, _docs/reference-commons.md,
	// _docs/GUIDE.md, _patterns/*).
	Catalog string `yaml:"catalog"`
	Pattern string `yaml:"pattern"`
}

// evalQuestion is one benchmark question with its recorded ground truth.
type evalQuestion struct {
	ID       string      `yaml:"id"`
	Class    string      `yaml:"class"`
	Question string      `yaml:"question"`
	Answer   string      `yaml:"answer"`
	Checks   []evalCheck `yaml:"checks"`
}

type evalBank struct {
	Questions []evalQuestion `yaml:"questions"`
}

// resolveCheckPath turns a check's file selector into an absolute path,
// deriving kind directories from the registry exactly as the generator
// does.
func resolveCheckPath(root string, c evalCheck) (string, error) {
	selectors := 0
	for _, s := range []string{c.Kind, c.Provider, c.Catalog} {
		if s != "" {
			selectors++
		}
	}
	if selectors != 1 {
		return "", fmt.Errorf("exactly one of kind, provider, or catalog must be set (got %d)", selectors)
	}
	switch {
	case c.Kind != "":
		if c.File != "reference.md" && c.File != "GUIDE.md" {
			return "", fmt.Errorf("kind checks must name file reference.md or GUIDE.md (got %q)", c.File)
		}
		res, err := explain.ResolveKindName(c.Kind)
		if err != nil {
			return "", fmt.Errorf("kind %q does not resolve: %w", c.Kind, err)
		}
		// reference.md lives in the version dir beside the protos; GUIDE.md
		// lives at the kind root, one level up.
		protoDir := filepath.Dir(res.Message.ParentFile().Path())
		if c.File == "GUIDE.md" {
			return filepath.Join(root, filepath.Dir(protoDir), c.File), nil
		}
		return filepath.Join(root, protoDir, c.File), nil
	case c.Provider != "":
		return filepath.Join(root, catalogPathPrefix, c.Provider, "reference-index.md"), nil
	default:
		return filepath.Join(root, catalogPathPrefix, c.Catalog), nil
	}
}

// TestEvalQuestionBankIsFresh re-verifies every recorded ground-truth
// assertion in eval_questions.yaml against the committed catalog.
func TestEvalQuestionBankIsFresh(t *testing.T) {
	root := repoRoot(t)

	raw, err := os.ReadFile(filepath.Join(root, "pkg", "explain", "refgen", "eval_questions.yaml"))
	if err != nil {
		t.Fatalf("read eval question bank: %v", err)
	}
	var bank evalBank
	if err := yaml.Unmarshal(raw, &bank); err != nil {
		t.Fatalf("parse eval question bank: %v", err)
	}
	if len(bank.Questions) == 0 {
		t.Fatal("eval question bank is empty")
	}

	seen := map[string]bool{}
	for _, q := range bank.Questions {
		if q.ID == "" || q.Class == "" || q.Question == "" || q.Answer == "" {
			t.Errorf("question %q: id, class, question, and answer are all required", q.ID)
			continue
		}
		if seen[q.ID] {
			t.Errorf("question %q: duplicate id", q.ID)
			continue
		}
		seen[q.ID] = true
		if len(q.Checks) == 0 {
			t.Errorf("question %q: at least one machine-checkable assertion is required", q.ID)
			continue
		}
		for i, c := range q.Checks {
			path, err := resolveCheckPath(root, c)
			if err != nil {
				t.Errorf("question %q check %d: %v", q.ID, i, err)
				continue
			}
			if c.Pattern == "" {
				t.Errorf("question %q check %d: pattern is required", q.ID, i)
				continue
			}
			re, err := regexp.Compile(c.Pattern)
			if err != nil {
				t.Errorf("question %q check %d: bad pattern: %v", q.ID, i, err)
				continue
			}
			content, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("question %q check %d: cited file is missing: %v", q.ID, i, err)
				continue
			}
			if !re.Match(content) {
				rel, _ := filepath.Rel(root, path)
				t.Errorf("question %q check %d: recorded ground truth no longer matches %s -- "+
					"the catalog changed; update the question's answer and checks in eval_questions.yaml",
					q.ID, i, rel)
			}
		}
	}
}
