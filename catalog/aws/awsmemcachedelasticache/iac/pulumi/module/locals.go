package module

import (
	"strconv"

	awsmemcachedelasticachev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmemcachedelasticache/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target  *awsmemcachedelasticachev1alpha1.AwsMemcachedElasticache
	Spec    *awsmemcachedelasticachev1alpha1.AwsMemcachedElasticacheSpec
	AwsTags map[string]string

	// ClusterIdentifier is metadata.name — create-only in AWS, and the
	// basis both engines share so a manifest deploys identically on either.
	ClusterIdentifier string
}

func initializeLocals(_ *pulumi.Context, in *awsmemcachedelasticachev1alpha1.AwsMemcachedElasticacheIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.ClusterIdentifier = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsMemcachedElasticache.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
