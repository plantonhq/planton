package module

import (
	"strconv"

	awsnatgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsnatgateway/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsNatGateway *awsnatgatewayv1.AwsNatGateway
	AwsTags       map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *awsnatgatewayv1.AwsNatGatewayStackInput) *Locals {
	locals := &Locals{}
	locals.AwsNatGateway = stackInput.Target

	metadata := stackInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsNatGateway.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
