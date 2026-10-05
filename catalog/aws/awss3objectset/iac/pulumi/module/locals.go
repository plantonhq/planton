package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awss3objectsetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awss3objectset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsS3ObjectSet *awss3objectsetv1alpha1.AwsS3ObjectSet
	AwsTags        map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awss3objectsetv1alpha1.AwsS3ObjectSetIacInput) *Locals {
	locals := &Locals{}

	locals.AwsS3ObjectSet = iacInput.Target

	target := locals.AwsS3ObjectSet

	// Resource-identity tags match the Terraform module key-for-key (the
	// canonical six-key identity map — user labels never merge into cloud
	// tags). Every object in the set carries them (S3 object tags),
	// attributing each object back to the owning set for auditing and orphan
	// cleanup.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         target.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: target.Metadata.Org,
		awstagkeys.Environment:  target.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsS3ObjectSet.String(),
		awstagkeys.ResourceId:   target.Metadata.Id,
	}

	return locals
}
