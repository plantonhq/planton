package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	awsgav1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsglobalaccelerator/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the Global Accelerator resource definition from the stack input
// and a map of AWS tags to apply to all created resources.
type Locals struct {
	GlobalAccelerator *awsgav1.AwsGlobalAccelerator
	AwsTags           map[string]string
}

// initializeLocals reads the stack input and builds the Locals instance.
func initializeLocals(ctx *pulumi.Context, stackInput *awsgav1.AwsGlobalAcceleratorStackInput) *Locals {
	locals := &Locals{}

	locals.GlobalAccelerator = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.GlobalAccelerator.Metadata.Org,
		awstagkeys.Environment:  locals.GlobalAccelerator.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsGlobalAccelerator.String(),
		awstagkeys.ResourceId:   locals.GlobalAccelerator.Metadata.Id,
	}

	return locals
}
