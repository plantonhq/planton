package module

import (
	"strconv"

	awssecuritygroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awssecuritygroup/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the AWS Security Group resource definition from the stack input
// and a map of AWS tags to apply to resources.
type Locals struct {
	AwsSecurityGroup *awssecuritygroupv1.AwsSecurityGroup
	AwsTags          map[string]string
}

// initializeLocals is similar to Terraform "locals" usage. It reads
// values from AwsSecurityGroupStackInput to build a Locals instance.
func initializeLocals(ctx *pulumi.Context, stackInput *awssecuritygroupv1.AwsSecurityGroupStackInput) *Locals {
	locals := &Locals{}

	locals.AwsSecurityGroup = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsSecurityGroup.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsSecurityGroup.Metadata.Org,
		awstagkeys.Environment:  locals.AwsSecurityGroup.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsSecurityGroup.String(),
		awstagkeys.ResourceId:   locals.AwsSecurityGroup.Metadata.Id,
	}

	return locals
}
