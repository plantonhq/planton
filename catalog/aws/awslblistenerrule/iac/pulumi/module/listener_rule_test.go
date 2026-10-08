package module

// A forward to one unweighted target group with no stickiness takes the simple
// target_group_arn form; any other forward takes the forward block. Stickiness
// is written into the forward block only when enabled, because AWS rejects a
// duration of 0 and a forward block without stickiness is stickiness off.

import (
	"testing"

	awslblistenerrulev1alpha1 "github.com/plantonhq/planton/catalog/aws/awslblistenerrule/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
)

const targetGroupArn = "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/api/943f017f100becff"

func forwardTo(t *testing.T, weights []int32, stickiness *awslblistenerrulev1alpha1.AwsLbListenerRuleActionForwardStickiness) *lb.ListenerRuleActionArgs {
	t.Helper()
	targetGroups := make([]*awslblistenerrulev1alpha1.AwsLbListenerRuleActionForwardTargetGroup, 0, len(weights))
	for _, weight := range weights {
		targetGroups = append(targetGroups, &awslblistenerrulev1alpha1.AwsLbListenerRuleActionForwardTargetGroup{
			Arn: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: targetGroupArn},
			},
			Weight: weight,
		})
	}
	actions, err := actionArgs([]*awslblistenerrulev1alpha1.AwsLbListenerRuleAction{{
		Type: "forward",
		Forward: &awslblistenerrulev1alpha1.AwsLbListenerRuleActionForward{
			TargetGroups: targetGroups,
			Stickiness:   stickiness,
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return actions[0].(*lb.ListenerRuleActionArgs)
}

func TestASingleUnweightedGroupWithoutStickinessTakesTheTargetGroupArnForm(t *testing.T) {
	args := forwardTo(t, []int32{0}, nil)
	if args.TargetGroupArn == nil {
		t.Fatal("target_group_arn is not set")
	}
	if args.Forward != nil {
		t.Fatal("a forward block is written")
	}
}

func TestStickinessOffWritesAForwardBlockWithoutStickiness(t *testing.T) {
	args := forwardTo(t, []int32{0}, &awslblistenerrulev1alpha1.AwsLbListenerRuleActionForwardStickiness{Enabled: false})
	if args.TargetGroupArn != nil {
		t.Fatal("target_group_arn is set")
	}
	forward, ok := args.Forward.(*lb.ListenerRuleActionForwardArgs)
	if !ok {
		t.Fatal("no forward block is written")
	}
	if forward.Stickiness != nil {
		t.Fatal("a stickiness block is written for stickiness off")
	}
}

func TestStickinessOffOnAWeightedForwardWritesNoStickiness(t *testing.T) {
	args := forwardTo(t, []int32{1}, &awslblistenerrulev1alpha1.AwsLbListenerRuleActionForwardStickiness{Enabled: false})
	forward, ok := args.Forward.(*lb.ListenerRuleActionForwardArgs)
	if !ok {
		t.Fatal("no forward block is written")
	}
	if forward.Stickiness != nil {
		t.Fatal("a stickiness block is written for stickiness off")
	}
}

func TestStickinessOnIsWrittenWithItsDuration(t *testing.T) {
	args := forwardTo(t, []int32{95, 5}, &awslblistenerrulev1alpha1.AwsLbListenerRuleActionForwardStickiness{
		Enabled:         true,
		DurationSeconds: 600,
	})
	forward, ok := args.Forward.(*lb.ListenerRuleActionForwardArgs)
	if !ok {
		t.Fatal("no forward block is written")
	}
	stickiness, ok := forward.Stickiness.(*lb.ListenerRuleActionForwardStickinessArgs)
	if !ok {
		t.Fatal("no stickiness block is written")
	}
	if stickiness.Duration == nil {
		t.Fatal("the stickiness block has no duration")
	}
}
