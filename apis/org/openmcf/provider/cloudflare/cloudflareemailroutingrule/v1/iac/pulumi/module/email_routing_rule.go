package module

import (
	"github.com/pkg/errors"
	cloudflareemailroutingrulev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareemailroutingrule/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// emailRoutingRule creates a single Email Routing rule. The typed action is
// mapped onto the provider's generic {type, value[]}.
func emailRoutingRule(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.EmailRoutingRule, error) {
	spec := locals.CloudflareEmailRoutingRule.Spec

	zoneId := ""
	if spec.ZoneId != nil {
		zoneId = spec.ZoneId.GetValue()
	}

	matchers := cloudflare.EmailRoutingRuleMatcherArray{}
	for _, m := range spec.Matchers {
		margs := cloudflare.EmailRoutingRuleMatcherArgs{Type: pulumi.String(m.Type.String())}
		if m.Field != "" {
			margs.Field = pulumi.StringPtr(m.Field)
		}
		if m.Value != "" {
			margs.Value = pulumi.StringPtr(m.Value)
		}
		matchers = append(matchers, margs)
	}

	// Map the typed action onto the provider's generic {type, value[]}.
	values := pulumi.StringArray{}
	if action := spec.Action; action != nil {
		switch action.Type {
		case cloudflareemailroutingrulev1.CloudflareEmailRoutingRuleActionType_forward:
			for _, f := range action.ForwardTo {
				values = append(values, pulumi.String(f.GetValue()))
			}
		case cloudflareemailroutingrulev1.CloudflareEmailRoutingRuleActionType_worker:
			if action.Worker != nil {
				values = append(values, pulumi.String(action.Worker.GetValue()))
			}
		}
	}

	actions := cloudflare.EmailRoutingRuleActionArray{
		cloudflare.EmailRoutingRuleActionArgs{
			Type:   pulumi.String(spec.Action.Type.String()),
			Values: values,
		},
	}

	args := &cloudflare.EmailRoutingRuleArgs{
		ZoneId:   pulumi.String(zoneId),
		Enabled:  pulumi.Bool(spec.GetEnabled()),
		Priority: pulumi.Float64(float64(spec.Priority)),
		Matchers: matchers,
		Actions:  actions,
	}
	if spec.Name != "" {
		args.Name = pulumi.StringPtr(spec.Name)
	}

	created, err := cloudflare.NewEmailRoutingRule(
		ctx,
		"email-routing-rule",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare email routing rule")
	}

	ctx.Export(OpRuleId, created.ID())
	ctx.Export(OpZoneId, pulumi.String(zoneId))

	return created, nil
}
