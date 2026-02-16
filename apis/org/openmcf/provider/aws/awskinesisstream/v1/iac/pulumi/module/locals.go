package module

import (
	"strconv"

	awskinesisstream "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awskinesisstream/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target     *awskinesisstream.AwsKinesisStream
	Spec       *awskinesisstream.AwsKinesisStreamSpec
	AwsTags    map[string]string
	StreamName string
}

func initializeLocals(ctx *pulumi.Context, in *awskinesisstream.AwsKinesisStreamStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.StreamName = in.Target.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsKinesisStream.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
