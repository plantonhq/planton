package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	awstgwv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awstransitgateway/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the Transit Gateway resource definition from the stack input
// and a map of AWS tags to apply to all created resources.
type Locals struct {
	TransitGateway *awstgwv1.AwsTransitGateway
	AwsTags        map[string]string
}

// initializeLocals reads the stack input and builds the Locals instance.
func initializeLocals(ctx *pulumi.Context, stackInput *awstgwv1.AwsTransitGatewayStackInput) *Locals {
	locals := &Locals{}

	locals.TransitGateway = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.TransitGateway.Metadata.Org,
		awstagkeys.Environment:  locals.TransitGateway.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsTransitGateway.String(),
		awstagkeys.ResourceId:   locals.TransitGateway.Metadata.Id,
	}

	return locals
}

// enableDisable converts a boolean to the AWS-style "enable"/"disable" string.
func enableDisable(b bool) string {
	if b {
		return "enable"
	}
	return "disable"
}
