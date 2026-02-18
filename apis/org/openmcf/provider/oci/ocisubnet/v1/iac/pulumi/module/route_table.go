package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ocisubnetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocisubnet/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func routeTable(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*core.RouteTable, error) {
	spec := locals.OciSubnet.Spec

	rules := make(core.RouteTableRouteRuleArray, len(spec.RouteRules))
	for i, rule := range spec.RouteRules {
		ruleArgs := core.RouteTableRouteRuleArgs{
			Destination:     pulumi.StringPtr(rule.Destination),
			DestinationType: pulumi.StringPtr(destinationTypeString(rule.DestinationType)),
			NetworkEntityId: pulumi.String(rule.NetworkEntityId.GetValue()),
		}
		if rule.Description != "" {
			ruleArgs.Description = pulumi.StringPtr(rule.Description)
		}
		rules[i] = ruleArgs
	}

	rtName := fmt.Sprintf("%s-rt", locals.DisplayName)

	createdRouteTable, err := core.NewRouteTable(ctx, rtName, &core.RouteTableArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		VcnId:         pulumi.String(spec.VcnId.GetValue()),
		DisplayName:   pulumi.StringPtr(rtName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
		RouteRules:    rules,
	}, pulumiOciOpt(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create oci route table")
	}

	return createdRouteTable, nil
}

func destinationTypeString(dt ocisubnetv1.OciSubnetSpec_RouteRule_DestinationType) string {
	switch dt {
	case ocisubnetv1.OciSubnetSpec_RouteRule_service_cidr_block:
		return "SERVICE_CIDR_BLOCK"
	case ocisubnetv1.OciSubnetSpec_RouteRule_cidr_block:
		return "CIDR_BLOCK"
	default:
		return strings.ToUpper(dt.String())
	}
}
