package module

import (
	"strconv"

	awstgwrtv1 "github.com/plantonhq/planton/catalog/aws/awstransitgatewayroutetable/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set applied to the route table.
type Locals struct {
	RouteTable *awstgwrtv1.AwsTransitGatewayRouteTable
	AwsTags    map[string]string
}

// initializeLocals reads the IaC input and builds the Locals instance.
func initializeLocals(ctx *pulumi.Context, iacInput *awstgwrtv1.AwsTransitGatewayRouteTableIacInput) *Locals {
	locals := &Locals{}

	locals.RouteTable = iacInput.Target

	// Identity tags match the Terraform module key-for-key. The Name tag IS
	// the route table's console identity -- route tables have no name
	// attribute of their own.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.RouteTable.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.RouteTable.Metadata.Org,
		awstagkeys.Environment:  locals.RouteTable.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsTransitGatewayRouteTable.String(),
		awstagkeys.ResourceId:   locals.RouteTable.Metadata.Id,
	}

	return locals
}
