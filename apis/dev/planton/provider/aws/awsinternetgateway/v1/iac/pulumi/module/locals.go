package module

import (
	"strconv"

	awsinternetgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsinternetgateway/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsInternetGateway *awsinternetgatewayv1.AwsInternetGateway
	AwsTags            map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *awsinternetgatewayv1.AwsInternetGatewayStackInput) *Locals {
	locals := &Locals{}
	locals.AwsInternetGateway = stackInput.Target

	metadata := stackInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsInternetGateway.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
