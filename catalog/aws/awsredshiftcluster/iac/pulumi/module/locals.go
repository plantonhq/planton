package module

import (
	"strconv"

	awsredshiftclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsredshiftcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsRedshiftCluster *awsredshiftclusterv1alpha1.AwsRedshiftCluster

	// ClusterIdentifier is metadata.name -- create-only in AWS, and the
	// basis both engines share so a manifest deploys identically on either.
	ClusterIdentifier string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsredshiftclusterv1alpha1.AwsRedshiftClusterIacInput) *Locals {
	locals := &Locals{}
	locals.AwsRedshiftCluster = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.ClusterIdentifier = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsRedshiftCluster.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
