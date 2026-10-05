package module

import (
	"strconv"

	awsmskserverlessclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsmskserverlesscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsMskServerlessCluster *awsmskserverlessclusterv1alpha1.AwsMskServerlessCluster

	// ClusterName is metadata.name -- create-only in AWS (max 64 chars), and
	// the basis both engines share so a manifest deploys identically on
	// either.
	ClusterName string

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsmskserverlessclusterv1alpha1.AwsMskServerlessClusterIacInput) *Locals {
	locals := &Locals{}
	locals.AwsMskServerlessCluster = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.ClusterName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key. Tags are
	// the ONLY mutable surface on a serverless MSK cluster.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsMskServerlessCluster.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
