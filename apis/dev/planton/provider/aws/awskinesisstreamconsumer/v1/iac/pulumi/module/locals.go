package module

import (
	"strconv"

	awskinesisstreamconsumer "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awskinesisstreamconsumer/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target       *awskinesisstreamconsumer.AwsKinesisStreamConsumer
	Spec         *awskinesisstreamconsumer.AwsKinesisStreamConsumerSpec
	AwsTags      map[string]string
	ConsumerName string
}

func initializeLocals(ctx *pulumi.Context, in *awskinesisstreamconsumer.AwsKinesisStreamConsumerStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.ConsumerName = in.Target.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsKinesisStreamConsumer.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
