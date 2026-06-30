package module

import (
	"strconv"

	awsserverlesselasticachev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsserverlesselasticache/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target  *awsserverlesselasticachev1.AwsServerlessElasticache
	Spec    *awsserverlesselasticachev1.AwsServerlessElasticacheSpec
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsserverlesselasticachev1.AwsServerlessElasticacheStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsServerlessElasticache.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
