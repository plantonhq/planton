package module

import (
	"strconv"

	awsecsclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsEcsCluster *awsecsclusterv1alpha1.AwsEcsCluster

	// ClusterName is metadata.name -- create-only in AWS (changing it
	// replaces the cluster), and the basis both engines share so a
	// manifest deploys identically on either.
	ClusterName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsecsclusterv1alpha1.AwsEcsClusterIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEcsCluster = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.ClusterName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEcsCluster.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
