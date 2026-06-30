package module

import (
	"strconv"

	awseksclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsekscluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals captures convenient references and computed values for the EKS cluster module.
type Locals struct {
	AwsEksCluster *awseksclusterv1.AwsEksCluster
	AwsTags       map[string]string
}

// initializeLocals creates and populates the Locals struct with computed values
// such as AWS tags based on the stack input.
func initializeLocals(ctx *pulumi.Context, stackInput *awseksclusterv1.AwsEksClusterStackInput) *Locals {
	locals := &Locals{}
	locals.AwsEksCluster = stackInput.Target

	// Build standard AWS tags for the cluster
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsEksCluster.Metadata.Org,
		awstagkeys.Environment:  locals.AwsEksCluster.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsEksCluster.String(),
		awstagkeys.ResourceId:   locals.AwsEksCluster.Metadata.Id,
	}

	return locals
}
