package module

import (
	"strconv"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"

	awsnlbv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsnetworkloadbalancer/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the NLB resource definition from the stack input and a map of
// AWS tags to apply to all created resources.
type Locals struct {
	Nlb     *awsnlbv1.AwsNetworkLoadBalancer
	AwsTags map[string]string
}

// initializeLocals reads the stack input and builds the Locals instance,
// analogous to a Terraform locals block.
func initializeLocals(ctx *pulumi.Context, stackInput *awsnlbv1.AwsNetworkLoadBalancerStackInput) *Locals {
	locals := &Locals{}

	locals.Nlb = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Nlb.Metadata.Org,
		awstagkeys.Environment:  locals.Nlb.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsNetworkLoadBalancer.String(),
		awstagkeys.ResourceId:   locals.Nlb.Metadata.Id,
	}

	return locals
}
