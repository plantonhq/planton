package module

import (
	"strconv"

	awssubnetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awssubnet/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsSubnet *awssubnetv1.AwsSubnet
	AwsTags   map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *awssubnetv1.AwsSubnetStackInput) *Locals {
	locals := &Locals{}
	locals.AwsSubnet = stackInput.Target

	metadata := stackInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsSubnet.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
