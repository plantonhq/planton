package keycloak

import (
	"context"
	"fmt"
	"strings"

	"github.com/plantonhq/planton/operator/internal/resources"
)

// The primary broker (DD-023): when the bound manifest declares oidc.primary,
// every sign-in goes straight to the company directory instead of the
// identity server's own form. Keycloak already has the mechanism -- the
// browser flow's Identity Provider Redirector redirects unconditionally once
// an authenticator config names a defaultProvider -- so the operator owns
// exactly ONE such config on that execution, aliased
// IdentityPrimaryBrokerConfigAlias, and nothing else about the flow:
//
//   - primary declared and the broker exists: the config is present with
//     defaultProvider = IdentityBrokerAlias (created, or its value converged);
//   - otherwise (primary off, the LDAP arm, the manifest deleted): the
//     operator's config is removed if present;
//   - an admin's own config on the redirector (any other alias) is admin
//     territory: never replaced, never deleted. With primary declared it is
//     reported as a finding -- the redirect is not applied over someone
//     else's setting -- and the pass continues.
//
// The break-glass is Keycloak's own: a sign-in carrying a kc_idp_hint the
// redirector cannot route makes it step aside and the local form renders.
// Every Planton client sends IdentityBreakGlassHint for that; the
// convergence suite pins the behaviour against the pinned identity server.

// identityProviderRedirectorID is the authenticator's provider id in
// Keycloak's browser flow.
const identityProviderRedirectorID = "identity-provider-redirector"

// redirectorConfigDefaultProviderKey is the config key the redirector reads.
const redirectorConfigDefaultProviderKey = "defaultProvider"

// redirectorState is what the reconciler learned about the browser flow's
// redirector this pass: the execution and whatever config sits on it.
type redirectorState struct {
	// ExecutionID is the redirector execution; empty when the realm's browser
	// flow carries no redirector at all (an admin removed it).
	ExecutionID string
	// ConfigID / ConfigAlias describe the config attached to it, if any.
	ConfigID    string
	ConfigAlias string
	// DefaultProvider is the attached config's defaultProvider value.
	DefaultProvider string
}

// redirectorAction is the one write (or none) the pass makes.
type redirectorAction int

const (
	redirectorLeave redirectorAction = iota
	redirectorCreate
	redirectorUpdate
	redirectorDelete
)

// redirectorPlan is the pure decision: given the live redirector and whether
// the primary broker is desired, which single write converges it. Kept free
// of I/O so the never-clobber cases are pinned without an identity server.
func redirectorPlan(live redirectorState, wantPrimary bool) (redirectorAction, string) {
	ours := live.ConfigID != "" && live.ConfigAlias == resources.IdentityPrimaryBrokerConfigAlias
	foreign := live.ConfigID != "" && !ours

	switch {
	case wantPrimary && live.ExecutionID == "":
		return redirectorLeave, "the browser flow carries no Identity Provider Redirector, so primary sign-in cannot be applied; restore the redirector to the flow or clear oidc.primary"
	case wantPrimary && foreign:
		return redirectorLeave, fmt.Sprintf("the browser flow's redirector already carries the config %q, which the operator does not own; primary sign-in is not applied over it", live.ConfigAlias)
	case wantPrimary && ours && live.DefaultProvider == resources.IdentityBrokerAlias:
		return redirectorLeave, ""
	case wantPrimary && ours:
		return redirectorUpdate, ""
	case wantPrimary:
		return redirectorCreate, ""
	case ours:
		return redirectorDelete, ""
	default:
		return redirectorLeave, ""
	}
}

// readRedirector locates the redirector on the realm's browser flow and the
// config attached to it. The flow is the one the realm names (browserFlow),
// so a realm whose admin bound a copy is converged on that copy.
func readRedirector(ctx context.Context, admin *AdminClient, realm string) (redirectorState, error) {
	realmRep, err := admin.GetRealm(ctx, realm)
	if err != nil {
		return redirectorState{}, err
	}
	flowAlias, _ := realmRep["browserFlow"].(string)
	if flowAlias == "" {
		flowAlias = "browser"
	}
	executions, err := admin.ListFlowExecutions(ctx, realm, flowAlias)
	if err != nil {
		return redirectorState{}, err
	}
	var state redirectorState
	for _, exec := range executions {
		if providerID, _ := exec["providerId"].(string); providerID != identityProviderRedirectorID {
			continue
		}
		state.ExecutionID, _ = exec["id"].(string)
		state.ConfigID, _ = exec["authenticationConfig"].(string)
		break
	}
	if state.ConfigID == "" {
		return state, nil
	}
	config, err := admin.GetAuthenticatorConfig(ctx, realm, state.ConfigID)
	if err != nil {
		return redirectorState{}, err
	}
	state.ConfigAlias, _ = config["alias"].(string)
	if values, ok := config["config"].(map[string]any); ok {
		state.DefaultProvider, _ = values[redirectorConfigDefaultProviderKey].(string)
	}
	return state, nil
}

// convergePrimaryBroker applies redirectorPlan's one write. A redirect the
// pass could not apply without clobbering is not an error: the rest of
// federation converges, and the verification pass names it as a finding
// (primarySignInCheck) so it lands on the manifest and the admin screens.
func convergePrimaryBroker(ctx context.Context, admin *AdminClient, realm string, broker *OwnedOIDCBroker, report *Report) error {
	wantPrimary := broker != nil && broker.Primary
	live, err := readRedirector(ctx, admin, realm)
	if err != nil {
		return err
	}
	action, _ := redirectorPlan(live, wantPrimary)
	switch action {
	case redirectorCreate:
		if err := admin.CreateExecutionConfig(ctx, realm, live.ExecutionID, Representation{
			"alias":  resources.IdentityPrimaryBrokerConfigAlias,
			"config": map[string]string{redirectorConfigDefaultProviderKey: resources.IdentityBrokerAlias},
		}); err != nil {
			return err
		}
		report.repaired("primary sign-in set: the browser flow redirects to broker %s", resources.IdentityBrokerAlias)
	case redirectorUpdate:
		if err := admin.UpdateAuthenticatorConfig(ctx, realm, live.ConfigID, Representation{
			"id":     live.ConfigID,
			"alias":  resources.IdentityPrimaryBrokerConfigAlias,
			"config": map[string]string{redirectorConfigDefaultProviderKey: resources.IdentityBrokerAlias},
		}); err != nil {
			return err
		}
		report.repaired("primary sign-in corrected: the redirector's default provider is %s again", resources.IdentityBrokerAlias)
	case redirectorDelete:
		if err := admin.DeleteAuthenticatorConfig(ctx, realm, live.ConfigID); err != nil {
			return err
		}
		report.repaired("primary sign-in removed: the identity server's own form shows again")
	case redirectorLeave:
	}
	return nil
}

// primarySignInCheckName is the verification check's name; the product
// renders its sentence verbatim on the Directory tab's connection panel.
const primarySignInCheckName = "primarySignIn"

// primarySignInCheck is the brokered arm's verdict on primary sign-in,
// derived from the same plan the convergence applied: Passed when every
// sign-in goes to the broker as declared (the sentence names the break-glass
// path), Unknown when the declaration could not be honoured (the finding),
// and no check at all when primary is not declared -- the button-beside-form
// shape is not a condition worth a line.
func primarySignInCheck(live redirectorState, broker *OwnedOIDCBroker) (Check, bool) {
	if broker == nil || !broker.Primary {
		return Check{}, false
	}
	action, finding := redirectorPlan(live, true)
	if finding != "" {
		return Check{Name: primarySignInCheckName, Verdict: VerdictUnknown, Message: finding}, true
	}
	if action != redirectorLeave {
		// The pass that verifies is the pass that just converged; a pending
		// write here means the convergence did not run this pass.
		return Check{Name: primarySignInCheckName, Verdict: VerdictUnknown,
			Message: "primary sign-in is declared but not yet applied on the identity server; the next reconcile pass applies it"}, true
	}
	return Check{Name: primarySignInCheckName, Verdict: VerdictPassed, Message: fmt.Sprintf(
		"every sign-in goes straight to %s; the local form stays reachable for break-glass at /login?local=1 (console), `planton login --local` (CLI), and the same hint on device sign-in",
		providerNameOf(broker.DisplayName))}, true
}

// providerNameOf turns the broker's sign-in BUTTON label ("Sign in with
// Microsoft") into the provider's name a sentence can carry; composed as-is
// the verdict read "goes straight to Sign in with Microsoft (lab)" on the
// Directory tab (observed live 2026-09-17). A label shaped some other way is
// the name.
func providerNameOf(label string) string {
	const prefix = "sign in with "
	trimmed := strings.TrimSpace(label)
	if len(trimmed) > len(prefix) && strings.EqualFold(trimmed[:len(prefix)], prefix) {
		return strings.TrimSpace(trimmed[len(prefix):])
	}
	return trimmed
}
