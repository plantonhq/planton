package module

import (
	"strconv"

	awsvpcv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsvpc/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsVpc  *awsvpcv1.AwsVpc
	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *awsvpcv1.AwsVpcStackInput) *Locals {
	locals := &Locals{}
	locals.AwsVpc = stackInput.Target

	metadata := stackInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsVpc.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
