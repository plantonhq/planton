package hostprobe

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The judge is the part of the probe that runs without a model, so it is
// the part that can be proven in the lint gate: a tree carrying each
// violation the fourth posture forbids is caught by name, and a clean tree
// passes. A judge that missed a violation would make the live GREEN leg a
// false green -- worse than no probe.

func newJudgeFixture(t *testing.T) *Fixture {
	t.Helper()
	parent := t.TempDir()
	repo := filepath.Join(parent, "orders-api")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	return &Fixture{Parent: parent, Repo: repo}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func claims(vs []Violation) string {
	var out []string
	for _, v := range vs {
		out = append(out, v.Claim)
	}
	return strings.Join(out, " | ")
}

func TestJudgeCleanRepositoryPasses(t *testing.T) {
	fx := newJudgeFixture(t)
	write(t, filepath.Join(fx.Repo, "infrastructure", "orders-db.yaml"), "apiVersion: x\nkind: GcpCloudSqlPostgres\n")
	transcript := []byte(`{"type":"assistant","text":"I wrote infrastructure/orders-db.yaml. Want me to apply it? (planton apply -f infrastructure/)"}` + "\n")
	vs := fx.Judge(transcript, func(string) error { return nil })
	if len(vs) != 0 {
		t.Fatalf("clean tree judged with violations: %s", claims(vs))
	}
}

func TestJudgeCatchesEveryForbiddenShape(t *testing.T) {
	fx := newJudgeFixture(t)
	// The old third posture applied to an application: chart at the root.
	write(t, filepath.Join(fx.Repo, "Chart.yaml"), "kind: InfraChart\n")
	// The canvas declaration in a repository.
	write(t, filepath.Join(fx.Repo, ".planton", "composing.yaml"), "state: composing\n")
	// A file written beside the repository (shell-location confusion).
	write(t, filepath.Join(fx.Parent, "orders-db.yaml"), "kind: X\n")
	// A manifest that is not a manifest, and one the oracle rejects.
	write(t, filepath.Join(fx.Repo, "infrastructure", "notes.yaml"), "just: notes\n")
	write(t, filepath.Join(fx.Repo, "infrastructure", "bad.yaml"), "kind: Nope\n")
	transcript := []byte(strings.Join([]string{
		`{"type":"tool_call","tool":"shell","command":"planton apply -f infrastructure/"}`,
		`{"type":"assistant","text":"An InfraChart is a parameterized bundle of resources."}`,
	}, "\n"))
	vs := fx.Judge(transcript, func(m string) error {
		if strings.HasSuffix(m, "bad.yaml") {
			return errors.New("unknown kind")
		}
		return nil
	})
	got := claims(vs)
	for _, want := range []string{
		"files stay inside the repository",
		"no .planton/ directory in an application repository",
		"no chart files at the repository root",
		"each manifest is a catalog object manifest",
		"each manifest validates",
		"never apply without consent",
		"platform constructs are never curriculum",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("judge missed %q; got: %s", want, got)
		}
	}
}

func TestJudgeRequiresInfrastructureFolder(t *testing.T) {
	fx := newJudgeFixture(t)
	vs := fx.Judge(nil, nil)
	if !strings.Contains(claims(vs), "infrastructure lives under infrastructure/") {
		t.Fatalf("empty repository should fail the infrastructure/ claim; got: %s", claims(vs))
	}
}

func TestTranscriptFactsReadsCommandsAndProseStructurally(t *testing.T) {
	transcript := []byte(strings.Join([]string{
		`{"type":"tool_call","subtype":"started","call":{"shell":{"command":"planton validate infrastructure/orders-db.yaml"}}}`,
		`{"type":"assistant","message":{"content":[{"type":"text","text":"Validated."}]}}`,
		`not json at all`,
	}, "\n"))
	commands, text := transcriptFacts(transcript)
	if len(commands) != 1 || !strings.Contains(commands[0], "planton validate") {
		t.Fatalf("commands = %v", commands)
	}
	if !strings.Contains(text, "Validated.") || !strings.Contains(text, "not json at all") {
		t.Fatalf("text = %q", text)
	}
}
