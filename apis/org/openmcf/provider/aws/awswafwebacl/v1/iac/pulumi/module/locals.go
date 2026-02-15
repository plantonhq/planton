package module

import (
	"strconv"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"

	awswafwebaclv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awswafwebacl/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the Web ACL resource definition from the stack input and a map
// of AWS tags to apply to all created resources.
type Locals struct {
	WebAcl  *awswafwebaclv1.AwsWafWebAcl
	AwsTags map[string]string
}

// initializeLocals reads the stack input and builds the Locals instance,
// analogous to a Terraform locals block.
func initializeLocals(ctx *pulumi.Context, stackInput *awswafwebaclv1.AwsWafWebAclStackInput) *Locals {
	locals := &Locals{}

	locals.WebAcl = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.WebAcl.Metadata.Org,
		awstagkeys.Environment:  locals.WebAcl.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsWafWebAcl.String(),
		awstagkeys.ResourceId:   locals.WebAcl.Metadata.Id,
	}

	return locals
}
