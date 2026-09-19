package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// hierarchicalFirewallPolicyVerifier probes an organization- or
// folder-level firewall policy through the compute API by its
// server-assigned numeric ID (the policy_id output, Google's `name`), and
// asserts the rule set landed -- the payload, not just the shell -- and
// that every association the module declared is attached. 404 is the
// destroyed shape.
type hierarchicalFirewallPolicyVerifier struct{}

func (v *hierarchicalFirewallPolicyVerifier) IDOutputKey() string { return "policy_id" }

func (v *hierarchicalFirewallPolicyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	policyId := outputs["policy_id"]
	if policyId == "" {
		return errors.New("policy_id output missing after deploy")
	}
	policy, err := svc.Compute.FirewallPolicies.Get(policyId).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "hierarchical firewall policy %s not found after deploy", policyId)
	}
	if want := outputs["short_name"]; want != "" && policy.ShortName != want {
		return errors.Errorf("hierarchical firewall policy %s has short name %q, want the short_name output %q", policyId, policy.ShortName, want)
	}
	// Google appends two implied goto_next rules to every hierarchical
	// policy; a policy whose only rules are Google's own carried no payload
	// from the manifest.
	if len(policy.Rules) <= 2 {
		return errors.Errorf("hierarchical firewall policy %s carries only Google's implied rules after deploy (%d total); the manifest's rules did not land", policyId, len(policy.Rules))
	}
	// Every declared association must be attached; the list output flattens
	// to indexed keys (association_names.0, ...) on both engines.
	for _, name := range indexed(outputs, "association_names") {
		if !hasFirewallPolicyAssociation(policy.Associations, name) {
			return errors.Errorf("hierarchical firewall policy %s has no association %q after deploy", policyId, name)
		}
	}
	return nil
}

func (v *hierarchicalFirewallPolicyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	policyId := outputs["policy_id"]
	if policyId == "" {
		return nil
	}
	_, err := svc.Compute.FirewallPolicies.Get(policyId).Context(ctx).Do()
	if err == nil {
		return errors.Errorf("hierarchical firewall policy %s still exists after destroy", policyId)
	}
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && apiErr.Code == 404 {
		return nil
	}
	return errors.Wrapf(err, "unexpected error probing hierarchical firewall policy %s after destroy", policyId)
}

// hasFirewallPolicyAssociation reports whether the policy's read-back
// association list carries the given name (both policy families share the
// association type).
func hasFirewallPolicyAssociation(associations []*compute.FirewallPolicyAssociation, name string) bool {
	for _, association := range associations {
		if association != nil && association.Name == name {
			return true
		}
	}
	return false
}
