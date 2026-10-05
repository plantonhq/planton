package module

import (
	"strconv"

	awsmemorydbclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmemorydbcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsmemorydbclusterv1alpha1.AwsMemorydbCluster
	Spec   *awsmemorydbclusterv1alpha1.AwsMemorydbClusterSpec

	// ClusterName is metadata.name -- the AWS cluster name is create-time
	// immutable, and metadata.name is the naming basis both engines share so
	// a manifest deploys identically on either. AWS caps it at 40 characters.
	// The module-managed subnet group and parameter group derive their names
	// from it too, so everything the module owns is discoverable by one name.
	ClusterName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsmemorydbclusterv1alpha1.AwsMemorydbClusterIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata
	locals.ClusterName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsMemorydbCluster.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
