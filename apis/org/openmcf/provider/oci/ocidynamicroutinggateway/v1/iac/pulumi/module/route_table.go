package module

import (
	"fmt"

	ocidynamicroutinggatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocidynamicroutinggateway/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createRouteTables creates all DRG route tables (without their static route
// rules). Rules are created separately in createStaticRouteRules after
// attachments exist.
func createRouteTables(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	drg *core.Drg,
	distMap map[string]*core.DrgRouteDistribution,
) (map[string]*core.DrgRouteTable, error) {
	spec := locals.OciDynamicRoutingGateway.Spec
	rtMap := make(map[string]*core.DrgRouteTable)

	for _, rtSpec := range spec.RouteTables {
		args := &core.DrgRouteTableArgs{
			DrgId:        drg.ID(),
			DisplayName:  pulumi.StringPtr(rtSpec.DisplayName),
			FreeformTags: pulumi.ToStringMap(locals.FreeformTags),
		}

		if rtSpec.IsEcmpEnabled {
			args.IsEcmpEnabled = pulumi.BoolPtr(true)
		}

		if rtSpec.ImportDrgRouteDistributionName != "" {
			if dist, ok := distMap[rtSpec.ImportDrgRouteDistributionName]; ok {
				args.ImportDrgRouteDistributionId = dist.ID()
			}
		}

		createdRt, err := core.NewDrgRouteTable(ctx, rtSpec.DisplayName, args,
			pulumiOciOpt(provider), pulumi.Parent(drg))
		if err != nil {
			return nil, fmt.Errorf("failed to create route table %s: %w", rtSpec.DisplayName, err)
		}
		rtMap[rtSpec.DisplayName] = createdRt
	}

	return rtMap, nil
}

// createStaticRouteRules creates static route rules for all route tables.
// Called after attachments exist so that next_hop_attachment_name can resolve
// to attachment OCIDs.
func createStaticRouteRules(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	rtMap map[string]*core.DrgRouteTable,
	attachmentIdMap map[string]pulumi.IDOutput,
) error {
	spec := locals.OciDynamicRoutingGateway.Spec

	for _, rtSpec := range spec.RouteTables {
		rt, ok := rtMap[rtSpec.DisplayName]
		if !ok {
			continue
		}

		for _, ruleSpec := range rtSpec.StaticRouteRules {
			if err := createRouteRule(ctx, provider, rt, rtSpec.DisplayName, ruleSpec, attachmentIdMap); err != nil {
				return fmt.Errorf("failed to create route rule in table %s: %w", rtSpec.DisplayName, err)
			}
		}
	}

	return nil
}

func createRouteRule(
	ctx *pulumi.Context,
	provider *oci.Provider,
	rt *core.DrgRouteTable,
	rtName string,
	ruleSpec *ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_StaticRouteRule,
	attachmentIdMap map[string]pulumi.IDOutput,
) error {
	resourceName := fmt.Sprintf("%s-%s", rtName, ruleSpec.Destination)

	args := &core.DrgRouteTableRouteRuleArgs{
		DrgRouteTableId: rt.ID(),
		Destination:     pulumi.String(ruleSpec.Destination),
		DestinationType: pulumi.String("CIDR_BLOCK"),
	}

	if attId, ok := attachmentIdMap[ruleSpec.NextHopAttachmentName]; ok {
		args.NextHopDrgAttachmentId = attId
	}

	_, err := core.NewDrgRouteTableRouteRule(ctx, resourceName, args,
		pulumiOciOpt(provider), pulumi.Parent(rt))

	return err
}
