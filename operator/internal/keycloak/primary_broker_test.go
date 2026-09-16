package keycloak

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/operator/internal/resources"
)

// The primary broker's decision table (DD-023), pinned without an identity
// server: the operator writes exactly one config on the redirector, converges
// only its own, and never touches an admin's.
func TestRedirectorPlan(t *testing.T) {
	ours := redirectorState{ExecutionID: "exec-1", ConfigID: "cfg-1",
		ConfigAlias: resources.IdentityPrimaryBrokerConfigAlias, DefaultProvider: resources.IdentityBrokerAlias}
	oursDrifted := redirectorState{ExecutionID: "exec-1", ConfigID: "cfg-1",
		ConfigAlias: resources.IdentityPrimaryBrokerConfigAlias, DefaultProvider: "someone-else"}
	bare := redirectorState{ExecutionID: "exec-1"}
	admins := redirectorState{ExecutionID: "exec-1", ConfigID: "cfg-9", ConfigAlias: "acme-redirect", DefaultProvider: "acme"}
	noRedirector := redirectorState{}

	cases := []struct {
		name        string
		live        redirectorState
		wantPrimary bool
		action      redirectorAction
		finding     bool
	}{
		{"primary on a bare redirector creates the config", bare, true, redirectorCreate, false},
		{"primary with our config converged is a no-op", ours, true, redirectorLeave, false},
		{"primary with our config drifted corrects it", oursDrifted, true, redirectorUpdate, false},
		{"primary over an admin's config is a finding, never a write", admins, true, redirectorLeave, true},
		{"primary with no redirector in the flow is a finding", noRedirector, true, redirectorLeave, true},
		{"off with our config removes it", ours, false, redirectorDelete, false},
		{"off with an admin's config leaves it", admins, false, redirectorLeave, false},
		{"off on a bare redirector is a no-op", bare, false, redirectorLeave, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			action, finding := redirectorPlan(tc.live, tc.wantPrimary)
			if action != tc.action {
				t.Fatalf("action = %v, want %v", action, tc.action)
			}
			if (finding != "") != tc.finding {
				t.Fatalf("finding = %q, want finding=%v", finding, tc.finding)
			}
		})
	}
}

func TestPrimarySignInCheck(t *testing.T) {
	broker := &OwnedOIDCBroker{Primary: true, DisplayName: "Sign in with Contoso"}
	converged := redirectorState{ExecutionID: "e", ConfigID: "c",
		ConfigAlias: resources.IdentityPrimaryBrokerConfigAlias, DefaultProvider: resources.IdentityBrokerAlias}

	if _, ok := primarySignInCheck(converged, &OwnedOIDCBroker{Primary: false}); ok {
		t.Fatal("no check when primary is not declared: the default shape is not a condition")
	}
	if _, ok := primarySignInCheck(converged, nil); ok {
		t.Fatal("no check without a broker")
	}

	check, ok := primarySignInCheck(converged, broker)
	if !ok || check.Verdict != VerdictPassed || check.Name != primarySignInCheckName {
		t.Fatalf("converged: %+v", check)
	}
	for _, want := range []string{"Sign in with Contoso", "/login?local=1", "planton login --local"} {
		if !strings.Contains(check.Message, want) {
			t.Fatalf("the passed sentence must name %q: %s", want, check.Message)
		}
	}

	check, _ = primarySignInCheck(redirectorState{ExecutionID: "e", ConfigID: "x", ConfigAlias: "acme"}, broker)
	if check.Verdict != VerdictUnknown || !strings.Contains(check.Message, `"acme"`) {
		t.Fatalf("an admin's config is an advisory finding naming it: %+v", check)
	}

	check, _ = primarySignInCheck(redirectorState{ExecutionID: "e"}, broker)
	if check.Verdict != VerdictUnknown || !strings.Contains(check.Message, "not yet applied") {
		t.Fatalf("a pending write is advisory, never a pass: %+v", check)
	}
}
