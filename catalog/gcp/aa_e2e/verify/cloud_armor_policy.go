package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// cloudArmorPolicyVerifier probes a Cloud Armor security policy via the
// compute API, reading the global or regional collection by the region
// output (empty = global), the same switch the modules make. Posture
// assertions confirm the default rule invariant every policy carries (a
// rule at priority 2147483647) and that the exported self_link matches the
// live resource — the value every backend-service and backend-bucket FK
// consumes.
type cloudArmorPolicyVerifier struct{}

const cloudArmorDefaultRulePriority = int64(2147483647)

func (v *cloudArmorPolicyVerifier) IDOutputKey() string { return "policy_name" }

// getCloudArmorPolicy reads the policy from whichever collection the region
// output selects.
func getCloudArmorPolicy(ctx context.Context, svc *Services, outputs map[string]string, name string) (*compute.SecurityPolicy, error) {
	if region := outputs["region"]; region != "" {
		return svc.Compute.RegionSecurityPolicies.Get(svc.Project, region, name).Context(ctx).Do()
	}
	return svc.Compute.SecurityPolicies.Get(svc.Project, name).Context(ctx).Do()
}

func (v *cloudArmorPolicyVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	policyName := outputs["policy_name"]
	if policyName == "" {
		return errors.New("policy_name output missing after deploy")
	}

	policy, err := getCloudArmorPolicy(ctx, svc, outputs, policyName)
	if err != nil {
		return errors.Wrapf(err, "cloud armor policy %s (region %q) not found after deploy", policyName, outputs["region"])
	}

	if wantSelfLink := outputs["policy_self_link"]; wantSelfLink != "" && policy.SelfLink != wantSelfLink {
		return errors.Errorf("cloud armor policy %s self_link mismatch: output %q, live %q", policyName, wantSelfLink, policy.SelfLink)
	}

	// Every Cloud Armor policy carries the default rule; its absence would
	// mean the rule set deployed incompletely.
	hasDefaultRule := false
	for _, rule := range policy.Rules {
		if rule.Priority == cloudArmorDefaultRulePriority {
			hasDefaultRule = true
			break
		}
	}
	if !hasDefaultRule {
		return errors.Errorf("cloud armor policy %s has no default rule at priority %d after deploy", policyName, cloudArmorDefaultRulePriority)
	}
	return nil
}

func (v *cloudArmorPolicyVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	policyName := outputs["policy_name"]
	if policyName == "" {
		return nil
	}

	_, err := getCloudArmorPolicy(ctx, svc, outputs, policyName)
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing cloud armor policy %s after destroy", policyName)
	}
	return errors.Errorf("cloud armor policy %s still exists after destroy", policyName)
}
