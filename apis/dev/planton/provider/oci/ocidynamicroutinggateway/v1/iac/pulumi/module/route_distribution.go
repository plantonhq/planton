package module

import (
	"fmt"
	"strings"

	ocidynamicroutinggatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocidynamicroutinggateway/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var distributionTypeMap = map[ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_DrgRouteDistribution_DistributionType]string{
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_DrgRouteDistribution_import_routes: "IMPORT",
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_DrgRouteDistribution_export_routes: "EXPORT",
}

var matchTypeMap = map[ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_MatchCriteria_MatchType]string{
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_MatchCriteria_match_all:           "MATCH_ALL",
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_MatchCriteria_drg_attachment_type: "DRG_ATTACHMENT_TYPE",
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_MatchCriteria_drg_attachment_id:   "DRG_ATTACHMENT_ID",
}

// createRouteDistributions creates all DRG route distributions (without
// their statements). Statements are created separately in
// createDistributionStatements after attachments exist.
func createRouteDistributions(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	drg *core.Drg,
) (map[string]*core.DrgRouteDistribution, error) {
	spec := locals.OciDynamicRoutingGateway.Spec
	distMap := make(map[string]*core.DrgRouteDistribution)

	for _, distSpec := range spec.RouteDistributions {
		createdDist, err := core.NewDrgRouteDistribution(ctx, distSpec.DisplayName, &core.DrgRouteDistributionArgs{
			DrgId:            drg.ID(),
			DistributionType: pulumi.String(distributionTypeMap[distSpec.DistributionType]),
			DisplayName:      pulumi.StringPtr(distSpec.DisplayName),
			FreeformTags:     pulumi.ToStringMap(locals.FreeformTags),
		}, pulumiOciOpt(provider), pulumi.Parent(drg))
		if err != nil {
			return nil, fmt.Errorf("failed to create route distribution %s: %w", distSpec.DisplayName, err)
		}
		distMap[distSpec.DisplayName] = createdDist
	}

	return distMap, nil
}

// createDistributionStatements creates statements for all route
// distributions. Called after attachments exist so that drg_attachment_id
// match criteria can resolve attachment OCIDs.
func createDistributionStatements(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	distMap map[string]*core.DrgRouteDistribution,
	attachmentIdMap map[string]pulumi.IDOutput,
) error {
	spec := locals.OciDynamicRoutingGateway.Spec

	for _, distSpec := range spec.RouteDistributions {
		dist, ok := distMap[distSpec.DisplayName]
		if !ok {
			continue
		}

		for _, stmtSpec := range distSpec.Statements {
			if err := createStatement(ctx, provider, dist, distSpec.DisplayName, stmtSpec, attachmentIdMap); err != nil {
				return fmt.Errorf("failed to create statement in distribution %s: %w", distSpec.DisplayName, err)
			}
		}
	}

	return nil
}

func createStatement(
	ctx *pulumi.Context,
	provider *oci.Provider,
	dist *core.DrgRouteDistribution,
	distName string,
	stmtSpec *ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_DistributionStatement,
	attachmentIdMap map[string]pulumi.IDOutput,
) error {
	mc := stmtSpec.MatchCriteria
	resourceName := fmt.Sprintf("%s-stmt-%d", distName, stmtSpec.Priority)

	matchCriteriaArgs := core.DrgRouteDistributionStatementMatchCriteriaArgs{
		MatchType: pulumi.StringPtr(matchTypeMap[mc.MatchType]),
	}

	if mc.AttachmentType != "" {
		matchCriteriaArgs.AttachmentType = pulumi.StringPtr(strings.ToUpper(mc.AttachmentType))
	}

	if mc.DrgAttachmentName != "" {
		if attId, ok := attachmentIdMap[mc.DrgAttachmentName]; ok {
			matchCriteriaArgs.DrgAttachmentId = attId
		}
	}

	_, err := core.NewDrgRouteDistributionStatement(ctx, resourceName, &core.DrgRouteDistributionStatementArgs{
		DrgRouteDistributionId: dist.ID(),
		Action:                 pulumi.String("ACCEPT"),
		MatchCriteria:          &matchCriteriaArgs,
		Priority:               pulumi.Int(int(stmtSpec.Priority)),
	}, pulumiOciOpt(provider), pulumi.Parent(dist))

	return err
}
