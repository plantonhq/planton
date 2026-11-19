package runner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/plantonhq/planton/e2e/framework/provider"
)

// dependentsReleaser is a harness double that records the class it was asked
// to release and fails for the class "stuck".
type dependentsReleaser struct {
	provider.Harness
	released *string
}

func (r dependentsReleaser) ReleaseDependents(_ context.Context, _ *provider.KindTestContext, class string) error {
	*r.released = class
	if class == "stuck" {
		return errors.New("still held")
	}
	return nil
}

func scenarioWithAnnotations(t *testing.T, annotations string) string {
	t.Helper()
	body := "apiVersion: gcp.planton.dev/v1alpha1\nkind: GcpDialogflowCxAgent\nmetadata:\n  name: x\n"
	if annotations != "" {
		body += "  annotations:\n" + annotations
	}
	path := filepath.Join(t.TempDir(), "scenario.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// A scenario that names no class records no phase; one that names a class
// hands it to the harness and records the verdict; a harness without the
// capability fails the phase instead of skipping it.
func TestReleaseDependentsRunsOnlyWhenDeclared(t *testing.T) {
	tc := &provider.KindTestContext{T: t, Provider: "gcp"}
	var released string
	harness := dependentsReleaser{released: &released}

	if phase := releaseDependents(context.Background(), tc, harness, scenarioWithAnnotations(t, "")); phase != nil {
		t.Fatalf("no annotation: got phase %+v, want none", phase)
	}

	phase := releaseDependents(context.Background(), tc, harness, scenarioWithAnnotations(t, "    "+ReleaseDependentsAnnotation+": chat-engines\n"))
	if phase == nil || !phase.Passed || phase.Phase != PhaseReleaseDependents || released != "chat-engines" {
		t.Fatalf("declared class: phase %+v, released %q", phase, released)
	}

	phase = releaseDependents(context.Background(), tc, harness, scenarioWithAnnotations(t, "    "+ReleaseDependentsAnnotation+": stuck\n"))
	if phase == nil || phase.Passed {
		t.Fatalf("a failed release must fail the phase: %+v", phase)
	}

	phase = releaseDependents(context.Background(), tc, propagationJudge{}, scenarioWithAnnotations(t, "    "+ReleaseDependentsAnnotation+": chat-engines\n"))
	if phase == nil || phase.Passed {
		t.Fatalf("a harness without the capability must fail the phase: %+v", phase)
	}
}
