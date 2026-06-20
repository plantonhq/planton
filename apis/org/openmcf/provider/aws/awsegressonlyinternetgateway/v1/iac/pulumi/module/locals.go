package module

import (
	"strconv"

	awsegressonlyinternetgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsegressonlyinternetgateway/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsEgressOnlyInternetGateway *awsegressonlyinternetgatewayv1.AwsEgressOnlyInternetGateway
	AwsTags                      map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *awsegressonlyinternetgatewayv1.AwsEgressOnlyInternetGatewayStackInput) *Locals {
	locals := &Locals{}
	locals.AwsEgressOnlyInternetGateway = stackInput.Target

	metadata := stackInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsEgressOnlyInternetGateway.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
