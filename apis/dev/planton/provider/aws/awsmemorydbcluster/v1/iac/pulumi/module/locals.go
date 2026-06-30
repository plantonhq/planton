package module

import (
	"strconv"

	awsmemorydbclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsmemorydbcluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target  *awsmemorydbclusterv1.AwsMemorydbCluster
	Spec    *awsmemorydbclusterv1.AwsMemorydbClusterSpec
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsmemorydbclusterv1.AwsMemorydbClusterStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsMemorydbCluster.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
