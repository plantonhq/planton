package runner

import (
	"errors"
	"testing"
	"time"

	"github.com/plantonhq/planton/e2e/framework/provider"
)

// propagationJudge is a harness double that reads every refusal of "held" as
// a grant still propagating.
type propagationJudge struct{ provider.Harness }

func (propagationJudge) HeldPermissionRefused(_ *provider.KindTestContext, err error) (string, bool) {
	return "p", err.Error() == "refused held"
}

func TestDeployWhileGrantsPropagateRetriesOnlyAHeldRefusal(t *testing.T) {
	interval, budget := propagationRetryInterval, propagationRetryBudget
	propagationRetryInterval, propagationRetryBudget = time.Millisecond, time.Second
	defer func() { propagationRetryInterval, propagationRetryBudget = interval, budget }()

	declared := &provider.KindTestContext{T: t, IdentityProviderConfig: "lane.yaml"}

	calls := 0
	err := deployWhileGrantsPropagate(declared, propagationJudge{}, func() error {
		calls++
		if calls < 3 {
			return errors.New("refused held")
		}
		return nil
	})
	if err != nil || calls != 3 {
		t.Fatalf("a held refusal is retried until the deploy succeeds: err %v after %d call(s)", err, calls)
	}

	calls = 0
	err = deployWhileGrantsPropagate(declared, propagationJudge{}, func() error {
		calls++
		return errors.New("refused missing")
	})
	if err == nil || calls != 1 {
		t.Fatalf("any other failure fails at once: err %v after %d call(s)", err, calls)
	}

	calls = 0
	ambient := &provider.KindTestContext{T: t}
	err = deployWhileGrantsPropagate(ambient, propagationJudge{}, func() error {
		calls++
		return errors.New("refused held")
	})
	if err == nil || calls != 1 {
		t.Fatalf("a lane without a declared identity deploys once: err %v after %d call(s)", err, calls)
	}
}

// The untaint reads every tainted resource from the state, child modules
// included, and nothing else.
func TestTaintedAddresses(t *testing.T) {
	state := []byte(`{"values":{"root_module":{"resources":[
		{"address":"google_privateca_certificate_template.this","tainted":true},
		{"address":"google_project_service.api"}],
		"child_modules":[{"resources":[{"address":"module.m.google_x.y","tainted":true}]}]}}}`)
	got, err := taintedAddresses(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "google_privateca_certificate_template.this" || got[1] != "module.m.google_x.y" {
		t.Fatalf("tainted = %v", got)
	}
	if got, err := taintedAddresses([]byte(`{"format_version":"1.0"}`)); err != nil || len(got) != 0 {
		t.Fatalf("an empty state lists nothing: %v, %v", got, err)
	}
}
