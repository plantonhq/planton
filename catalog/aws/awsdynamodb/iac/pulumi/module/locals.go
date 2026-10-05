package module

import (
	"strconv"

	awsdynamodbv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsdynamodb/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsDynamodb *awsdynamodbv1alpha1.AwsDynamodb

	// TableName is metadata.name -- create-only in AWS, and the basis
	// both engines share so a manifest deploys identically on either.
	TableName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsdynamodbv1alpha1.AwsDynamodbIacInput) *Locals {
	locals := &Locals{}
	locals.AwsDynamodb = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.TableName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsDynamodb.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
