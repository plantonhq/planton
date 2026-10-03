package module

import (
	"strconv"

	awskinesisstreamconsumer "github.com/plantonhq/planton/catalog/aws/awskinesisstreamconsumer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target       *awskinesisstreamconsumer.AwsKinesisStreamConsumer
	Spec         *awskinesisstreamconsumer.AwsKinesisStreamConsumerSpec
	AwsTags      map[string]string
	ConsumerName string
}

func initializeLocals(ctx *pulumi.Context, in *awskinesisstreamconsumer.AwsKinesisStreamConsumerIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.ConsumerName = in.Target.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.Target.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsKinesisStreamConsumer.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
