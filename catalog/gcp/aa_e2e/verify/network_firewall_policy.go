package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// networkFirewallPolicyVerifier probes a project's network firewall policy
// through the compute API by the policy_name output, on whichever resource
// family the region output selects (empty = global), and asserts the rule
// set landed -- the payload, not just the shell -- and that every
// association the module declared is attached to its network. 404 is the
// destroyed shape.
type networkFirewallPolicyVerifier struct{}

func (v *networkFirewallPolicyVerifier) IDOutputKey() string { return "policy_name" }

// getNetworkFirewallPolicy reads the policy from the global or regional
// family by the region output, the same switch the modules make.
func getNetworkFirewallPolicy(ctx context.Context, svc *Services, outputs map[string]string) (*compute.FirewallPolicy, error) {
	name := outputs["policy_name"]
	if region := outputs["region"]; region != "" {
		return svc.Compute.RegionNetworkFirewallPolicies.Get(svc.Project, region, name).Context(ctx).Do()
	}
	return svc.Compute.NetworkFirewallPolicies.Get(svc.Project, name).Context(ctx).Do()
}

func (v *networkFirewallPolicyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["policy_name"]
	if name == "" {
		return errors.New("policy_name output missing after deploy")
	}
	policy, err := getNetworkFirewallPolicy(ctx, svc, outputs)
	if err != nil {
		return errors.Wrapf(err, "network firewall policy %s (region %q) not found after deploy", name, outputs["region"])
	}
	// A network policy carries no implied rules: an empty rule list means the
	// manifest's rules did not land.
	if len(policy.Rules) == 0 {
		return errors.Errorf("network firewall policy %s has no rules after deploy", name)
	}
	// Every declared association must be attached; the list output flattens
	// to indexed keys (association_names.0, ...) on both engines.
	for _, associationName := range indexed(outputs, "association_names") {
		if !hasFirewallPolicyAssociation(policy.Associations, associationName) {
			return errors.Errorf("network firewall policy %s has no association %q after deploy", name, associationName)
		}
	}
	return nil
}

func (v *networkFirewallPolicyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["policy_name"]
	if name == "" {
		return nil
	}
	_, err := getNetworkFirewallPolicy(ctx, svc, outputs)
	if err == nil {
		return errors.Errorf("network firewall policy %s still exists after destroy", name)
	}
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && apiErr.Code == 404 {
		return nil
	}
	return errors.Wrapf(err, "unexpected error probing network firewall policy %s after destroy", name)
}
