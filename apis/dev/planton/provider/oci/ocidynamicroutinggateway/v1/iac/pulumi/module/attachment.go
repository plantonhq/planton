package module

import (
	"fmt"
	"strings"

	ocidynamicroutinggatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocidynamicroutinggateway/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var networkTypeMap = map[ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_NetworkType]string{
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_vcn:                       "VCN",
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_ipsec_tunnel:              "IPSEC_TUNNEL",
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_remote_peering_connection: "REMOTE_PEERING_CONNECTION",
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_virtual_circuit:           "VIRTUAL_CIRCUIT",
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_loopback:                  "LOOPBACK",
}

var vcnRouteTypeMap = map[ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_VcnRouteType]string{
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_vcn_cidrs:    "VCN_CIDRS",
	ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_subnet_cidrs: "SUBNET_CIDRS",
}

// createAttachments creates all DRG attachments. Returns a map of
// display_name -> attachment ID for downstream reference by route
// distribution statements and static route rules.
func createAttachments(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	drg *core.Drg,
	rtMap map[string]*core.DrgRouteTable,
	distMap map[string]*core.DrgRouteDistribution,
) (map[string]pulumi.IDOutput, error) {
	spec := locals.OciDynamicRoutingGateway.Spec
	attIdMap := make(map[string]pulumi.IDOutput)

	for _, attSpec := range spec.Attachments {
		nd := attSpec.NetworkDetails

		networkDetails := &core.DrgAttachmentNetworkDetailsArgs{
			Type: pulumi.String(networkTypeMap[nd.Type]),
			Id:   pulumi.StringPtr(nd.Id.GetValue()),
		}

		if nd.RouteTableId != "" {
			networkDetails.RouteTableId = pulumi.StringPtr(nd.RouteTableId)
		}

		if nd.VcnRouteType != ocidynamicroutinggatewayv1.OciDynamicRoutingGatewaySpec_NetworkDetails_vcn_route_type_unspecified {
			if mapped, ok := vcnRouteTypeMap[nd.VcnRouteType]; ok {
				networkDetails.VcnRouteType = pulumi.StringPtr(mapped)
			}
		}

		args := &core.DrgAttachmentArgs{
			DrgId:          drg.ID(),
			DisplayName:    pulumi.StringPtr(attSpec.DisplayName),
			NetworkDetails: networkDetails,
			FreeformTags:   pulumi.ToStringMap(locals.FreeformTags),
		}

		if attSpec.DrgRouteTableName != "" && rtMap != nil {
			if rt, ok := rtMap[attSpec.DrgRouteTableName]; ok {
				args.DrgRouteTableId = rt.ID()
			}
		}

		if attSpec.ExportDrgRouteDistributionName != "" && distMap != nil {
			if dist, ok := distMap[attSpec.ExportDrgRouteDistributionName]; ok {
				args.ExportDrgRouteDistributionId = dist.ID()
			}
		}

		createdAtt, err := core.NewDrgAttachment(ctx,
			fmt.Sprintf("%s-%s", locals.DisplayName, strings.ReplaceAll(attSpec.DisplayName, " ", "-")),
			args, pulumiOciOpt(provider), pulumi.Parent(drg))
		if err != nil {
			return nil, fmt.Errorf("failed to create drg attachment %s: %w", attSpec.DisplayName, err)
		}

		attIdMap[attSpec.DisplayName] = createdAtt.ID()
	}

	return attIdMap, nil
}
