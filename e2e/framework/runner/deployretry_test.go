package runner

import (
	"errors"
	"testing"
	"time"

	"github.com/plantonhq/planton/e2e/framework/provider"
)

// A scenario that declares a deploy retry re-runs a refused deploy until it
// succeeds; one that declares none deploys once.
func TestDeployMaybeRetryOnlyWhenDeclared(t *testing.T) {
	interval, budget := deployRetryInterval, deployRetryBudget
	deployRetryInterval, deployRetryBudget = time.Millisecond, time.Second
	defer func() { deployRetryInterval, deployRetryBudget = interval, budget }()

	declared := &provider.KindTestContext{T: t, ManifestPath: scenarioWithAnnotations(t, "    "+DeployRetryAnnotation+": a field the store's new schema indexes is refused for a while\n")}
	calls := 0
	err := deployMaybeRetry(declared, func() error {
		calls++
		if calls < 3 {
			return errors.New("Unsupported field")
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("declared: err %v after %d call(s)", err, calls)
	}

	plain := &provider.KindTestContext{T: t, ManifestPath: scenarioWithAnnotations(t, "")}
	calls = 0
	if err := deployMaybeRetry(plain, func() error { calls++; return errors.New("refused") }); err == nil || calls != 1 {
		t.Fatalf("undeclared: err %v after %d call(s)", err, calls)
	}
}
