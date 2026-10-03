package module

import (
	"strconv"

	awseksclusterv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsekscluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals captures convenient references and computed values for the EKS cluster module.
type Locals struct {
	AwsEksCluster *awseksclusterv1alpha1.AwsEksCluster
	AwsTags       map[string]string
}

// initializeLocals creates and populates the Locals struct with computed values
// such as AWS tags based on the IaC input.
func initializeLocals(ctx *pulumi.Context, iacInput *awseksclusterv1alpha1.AwsEksClusterIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEksCluster = iacInput.Target

	// Build standard AWS tags for the cluster
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsEksCluster.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsEksCluster.Metadata.Org,
		awstagkeys.Environment:  locals.AwsEksCluster.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEksCluster.String(),
		awstagkeys.ResourceId:   locals.AwsEksCluster.Metadata.Id,
	}

	return locals
}
