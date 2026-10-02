package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
)

// getForwardingRule reads the rule from the global or regional API
// collection by the region output, the same switch the modules make.
func getForwardingRule(ctx context.Context, svc *Services, outputs map[string]string) (*compute.ForwardingRule, error) {
	name := outputs["forwarding_rule_name"]
	if region := outputs["region"]; region != "" {
		return svc.Compute.ForwardingRules.Get(svc.Project, region, name).Context(ctx).Do()
	}
	return svc.Compute.GlobalForwardingRules.Get(svc.Project, name).Context(ctx).Do()
}

// globalForwardingRuleVerifier probes a forwarding rule by name -- the global
// or regional collection, chosen by the region output -- and confirms the
// frontend actually formed: a VIP was bound and a sink (target proxy or
// passthrough backend service) is wired — the rule is the node DNS points
// at, so both must hold.
type globalForwardingRuleVerifier struct{}

func (v *globalForwardingRuleVerifier) IDOutputKey() string { return "self_link" }

func (v *globalForwardingRuleVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["forwarding_rule_name"]
	rule, err := getForwardingRule(ctx, svc, outputs)
	if err != nil {
		return errors.Wrapf(err, "global forwarding rule %s not found after deploy", name)
	}
	if rule.IPAddress == "" {
		return errors.Errorf("global forwarding rule %s has no VIP bound", name)
	}
	// Exactly one sink is wired: a proxy-based load balancer's target, or a
	// passthrough Network Load Balancer's backend service (regional only).
	if rule.Target == "" && rule.BackendService == "" {
		return errors.Errorf("forwarding rule %s has neither a target nor a backend service wired", name)
	}
	// An internal passthrough NLB that asked for a service label must have
	// been given its internal DNS name.
	if outputs["service_name"] == "" && rule.ServiceLabel != "" {
		return errors.Errorf("forwarding rule %s set service_label %q but reports no service_name", name, rule.ServiceLabel)
	}
	return nil
}

func (v *globalForwardingRuleVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["forwarding_rule_name"]
	_, err := getForwardingRule(ctx, svc, outputs)
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing global forwarding rule %s after destroy", name)
	}
	return errors.Errorf("global forwarding rule %s still exists after destroy", name)
}
