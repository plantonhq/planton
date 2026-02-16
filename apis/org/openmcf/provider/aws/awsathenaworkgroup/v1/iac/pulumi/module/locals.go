package module

import (
	"strconv"

	awsathenaworkgroup "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsathenaworkgroup/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target        *awsathenaworkgroup.AwsAthenaWorkgroup
	Spec          *awsathenaworkgroup.AwsAthenaWorkgroupSpec
	AwsTags       map[string]string
	WorkgroupName string
}

func initializeLocals(ctx *pulumi.Context, in *awsathenaworkgroup.AwsAthenaWorkgroupStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.WorkgroupName = in.Target.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsAthenaWorkgroup.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
