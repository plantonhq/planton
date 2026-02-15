package module

import (
	"strconv"

	awsmemcachedelasticachev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsmemcachedelasticache/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target  *awsmemcachedelasticachev1.AwsMemcachedElasticache
	Spec    *awsmemcachedelasticachev1.AwsMemcachedElasticacheSpec
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsmemcachedelasticachev1.AwsMemcachedElasticacheStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsMemcachedElasticache.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
