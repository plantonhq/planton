package module

import (
	"strconv"

	awsmskclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmskcluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsMskCluster *awsmskclusterv1alpha1.AwsMskCluster
	// ClusterName is metadata.name -- create-only in AWS (max 64 chars), and the
	// basis both engines share so a manifest deploys identically on either.
	ClusterName string
	Labels      map[string]string
}

func initializeLocals(ctx *pulumi.Context, in *awsmskclusterv1alpha1.AwsMskClusterIacInput) *Locals {
	locals := &Locals{}

	locals.AwsMskCluster = in.Target
	locals.ClusterName = in.Target.Metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.Labels = map[string]string{
		awstagkeys.Name:         locals.AwsMskCluster.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsMskCluster.Metadata.Org,
		awstagkeys.Environment:  locals.AwsMskCluster.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsMskCluster.String(),
		awstagkeys.ResourceId:   locals.AwsMskCluster.Metadata.Id,
	}

	return locals
}
