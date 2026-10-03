package module

import (
	"strconv"

	awsrediselasticachev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsrediselasticache/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	AwsRedisElasticache *awsrediselasticachev1alpha1.AwsRedisElasticache

	// ReplicationGroupId is metadata.name -- create-only in AWS, and the
	// basis both engines share so a manifest deploys identically on either.
	ReplicationGroupId string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsrediselasticachev1alpha1.AwsRedisElasticacheIacInput) *Locals {
	locals := &Locals{}
	locals.AwsRedisElasticache = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.ReplicationGroupId = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsRedisElasticache.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
