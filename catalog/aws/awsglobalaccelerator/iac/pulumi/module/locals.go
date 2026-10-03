package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awsgav1 "github.com/plantonhq/planton/catalog/aws/awsglobalaccelerator/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the Global Accelerator resource definition from the IaC input
// and a map of AWS tags to apply to all created resources.
type Locals struct {
	GlobalAccelerator *awsgav1.AwsGlobalAccelerator
	AwsTags           map[string]string
}

// initializeLocals reads the IaC input and builds the Locals instance.
func initializeLocals(ctx *pulumi.Context, iacInput *awsgav1.AwsGlobalAcceleratorIacInput) *Locals {
	locals := &Locals{}

	locals.GlobalAccelerator = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.GlobalAccelerator.Metadata.Org,
		awstagkeys.Environment:  locals.GlobalAccelerator.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsGlobalAccelerator.String(),
		awstagkeys.ResourceId:   locals.GlobalAccelerator.Metadata.Id,
	}

	return locals
}
