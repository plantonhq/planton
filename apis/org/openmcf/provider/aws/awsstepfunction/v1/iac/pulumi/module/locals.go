package module

import (
	"strconv"

	awsstepfunctionv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsstepfunction/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target  *awsstepfunctionv1.AwsStepFunction
	Spec    *awsstepfunctionv1.AwsStepFunctionSpec
	AwsTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsstepfunctionv1.AwsStepFunctionStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsStepFunction.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
