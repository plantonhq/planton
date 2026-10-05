package module

import (
	"strconv"

	awskinesisstream "github.com/plantonhq/planton/catalog/aws/awskinesisstream/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target     *awskinesisstream.AwsKinesisStream
	Spec       *awskinesisstream.AwsKinesisStreamSpec
	AwsTags    map[string]string
	StreamName string
}

func initializeLocals(ctx *pulumi.Context, in *awskinesisstream.AwsKinesisStreamIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.StreamName = in.Target.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.Target.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsKinesisStream.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
