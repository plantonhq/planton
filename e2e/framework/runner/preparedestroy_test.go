package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/plantonhq/planton/e2e/framework/provider"
)

// destroyPreparer is a harness double that counts its calls and fails for
// the class "refused".
type destroyPreparer struct {
	provider.Harness
	calls *int
}

func (p destroyPreparer) PrepareDestroy(_ context.Context, _ *provider.KindTestContext, class string) error {
	*p.calls++
	if class == "refused" {
		return errors.New("the owner's request was refused")
	}
	return nil
}

// A scenario that names no class has no step; one that names a class hands
// it to the harness; a refused step and a harness without the capability
// both fail, so the destroy they guard fails loudly.
func TestDestroyPreparationRunsOnlyWhenDeclared(t *testing.T) {
	tc := &provider.KindTestContext{T: t, Provider: "gcp"}
	calls := 0
	harness := destroyPreparer{calls: &calls}
	ctx := context.Background()

	if step := destroyPreparation(ctx, tc, harness, scenarioWithAnnotations(t, "")); step != nil {
		t.Fatal("no annotation: want no step")
	}

	step := destroyPreparation(ctx, tc, harness, scenarioWithAnnotations(t, "    "+PrepareDestroyAnnotation+": consensus-peering\n"))
	if step == nil || step() != nil || calls != 1 {
		t.Fatalf("declared class: step declared %t, calls %d", step != nil, calls)
	}

	step = destroyPreparation(ctx, tc, harness, scenarioWithAnnotations(t, "    "+PrepareDestroyAnnotation+": refused\n"))
	if step == nil || step() == nil {
		t.Fatal("a refused step must fail")
	}

	step = destroyPreparation(ctx, tc, propagationJudge{}, scenarioWithAnnotations(t, "    "+PrepareDestroyAnnotation+": consensus-peering\n"))
	if step == nil || step() == nil {
		t.Fatal("a harness without the capability must fail the step")
	}
}

// Every destroy attempt is preceded by the step, and a failed step fails the
// attempt without running the destroy.
func TestDestroyAttemptTakesTheStepFirst(t *testing.T) {
	calls := 0
	destroys := 0
	destroy := func() error { destroys++; return nil }
	if err := destroyAttempt(func() error { calls++; return nil }, destroy); err != nil || calls != 1 || destroys != 1 {
		t.Fatalf("step then destroy: err %v, calls %d, destroys %d", err, calls, destroys)
	}
	if err := destroyAttempt(func() error { return errors.New("refused") }, destroy); err == nil || destroys != 1 {
		t.Fatalf("a failed step must fail the attempt before the destroy: err %v, destroys %d", err, destroys)
	}
	if err := destroyAttempt(nil, destroy); err != nil || destroys != 2 {
		t.Fatalf("no step: err %v, destroys %d", err, destroys)
	}
}
