package verify

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/smithy-go"
	pkgerrors "github.com/pkg/errors"
)

// The ELBv2 family shares one API surface: load balancers (ALB and NLB alike),
// target groups, listeners, and listener rules are all addressed by ARN
// through Describe* calls, and each has a typed *NotFound error code that is
// the "absent" signal. One verifier type per resource class, parameterized by
// the component name for error messages -- the ALB and NLB verifiers are the
// same loadBalancerVerifier because AWS does not distinguish them at the
// verification API.

// loadBalancerVerifier verifies an ALB or NLB via DescribeLoadBalancers,
// keyed on the load balancer ARN.
type loadBalancerVerifier struct {
	kind string
}

func (*loadBalancerVerifier) IDOutputKey() string { return "load_balancer_arn" }

func (v *loadBalancerVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := loadBalancerExists(ctx, cfg, region, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-exists failed for %q", v.kind, id)
	}
	if !exists {
		return pkgerrors.Errorf("%s %q not found after deploy", v.kind, id)
	}
	return nil
}

func (v *loadBalancerVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := loadBalancerExists(ctx, cfg, region, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "%s verify-absent failed for %q", v.kind, id)
	}
	if exists {
		return pkgerrors.Errorf("%s %q still exists after destroy", v.kind, id)
	}
	return nil
}

func loadBalancerExists(ctx context.Context, cfg aws.Config, region, arn string) (bool, error) {
	client := elbv2Client(cfg, region)
	_, err := client.DescribeLoadBalancers(ctx, &elbv2.DescribeLoadBalancersInput{
		LoadBalancerArns: []string{arn},
	})
	return interpretElbv2Lookup(err, "LoadBalancerNotFound")
}

// targetGroupVerifier verifies an AwsLbTargetGroup via DescribeTargetGroups,
// keyed on the target group ARN.
type targetGroupVerifier struct{}

func (*targetGroupVerifier) IDOutputKey() string { return "target_group_arn" }

func (*targetGroupVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := targetGroupExists(ctx, cfg, region, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awslbtargetgroup verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awslbtargetgroup %q not found after deploy", id)
	}
	return nil
}

func (*targetGroupVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := targetGroupExists(ctx, cfg, region, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awslbtargetgroup verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awslbtargetgroup %q still exists after destroy", id)
	}
	return nil
}

func targetGroupExists(ctx context.Context, cfg aws.Config, region, arn string) (bool, error) {
	client := elbv2Client(cfg, region)
	_, err := client.DescribeTargetGroups(ctx, &elbv2.DescribeTargetGroupsInput{
		TargetGroupArns: []string{arn},
	})
	return interpretElbv2Lookup(err, "TargetGroupNotFound")
}

// listenerVerifier verifies an AwsLbListener via DescribeListeners, keyed on
// the listener ARN.
type listenerVerifier struct{}

func (*listenerVerifier) IDOutputKey() string { return "listener_arn" }

func (*listenerVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := listenerExists(ctx, cfg, region, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awslblistener verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awslblistener %q not found after deploy", id)
	}
	return nil
}

func (*listenerVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := listenerExists(ctx, cfg, region, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awslblistener verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awslblistener %q still exists after destroy", id)
	}
	return nil
}

func listenerExists(ctx context.Context, cfg aws.Config, region, arn string) (bool, error) {
	client := elbv2Client(cfg, region)
	_, err := client.DescribeListeners(ctx, &elbv2.DescribeListenersInput{
		ListenerArns: []string{arn},
	})
	return interpretElbv2Lookup(err, "ListenerNotFound")
}

// listenerRuleVerifier verifies an AwsLbListenerRule via DescribeRules, keyed
// on the rule ARN. Destroying a listener cascades its rules at AWS, so the
// absent check also treats a missing parent listener as absent.
type listenerRuleVerifier struct{}

func (*listenerRuleVerifier) IDOutputKey() string { return "rule_arn" }

func (*listenerRuleVerifier) VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := listenerRuleExists(ctx, cfg, region, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awslblistenerrule verify-exists failed for %q", id)
	}
	if !exists {
		return pkgerrors.Errorf("awslblistenerrule %q not found after deploy", id)
	}
	return nil
}

func (*listenerRuleVerifier) VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error {
	exists, err := listenerRuleExists(ctx, cfg, region, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "awslblistenerrule verify-absent failed for %q", id)
	}
	if exists {
		return pkgerrors.Errorf("awslblistenerrule %q still exists after destroy", id)
	}
	return nil
}

func listenerRuleExists(ctx context.Context, cfg aws.Config, region, arn string) (bool, error) {
	client := elbv2Client(cfg, region)
	_, err := client.DescribeRules(ctx, &elbv2.DescribeRulesInput{
		RuleArns: []string{arn},
	})
	// A destroyed parent listener cascades its rules, and AWS then reports
	// ListenerNotFound rather than RuleNotFound for the orphaned rule ARN.
	return interpretElbv2Lookup(err, "RuleNotFound", "ListenerNotFound")
}

// elbv2Client builds a regional ELBv2 client; verification always runs in the
// same region the scenario deployed to.
func elbv2Client(cfg aws.Config, region string) *elbv2.Client {
	if region != "" {
		cfg.Region = region
	}
	return elbv2.NewFromConfig(cfg)
}

// interpretElbv2Lookup collapses a Describe* result into (exists, error):
// nil error means the resource exists, any listed not-found code means it is
// absent, and everything else is a genuine failure that must surface.
func interpretElbv2Lookup(err error, notFoundCodes ...string) (bool, error) {
	if err == nil {
		return true, nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		for _, code := range notFoundCodes {
			if apiErr.ErrorCode() == code {
				return false, nil
			}
		}
	}
	return false, err
}
