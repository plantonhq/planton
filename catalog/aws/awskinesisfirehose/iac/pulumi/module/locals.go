package module

import (
	"strconv"

	awskinesisfirehose "github.com/plantonhq/planton/catalog/aws/awskinesisfirehose/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target             *awskinesisfirehose.AwsKinesisFirehose
	Spec               *awskinesisfirehose.AwsKinesisFirehoseSpec
	AwsTags            map[string]string
	DeliveryStreamName string
}

func initializeLocals(ctx *pulumi.Context, in *awskinesisfirehose.AwsKinesisFirehoseIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.DeliveryStreamName = in.Target.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.Target.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsKinesisFirehose.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
