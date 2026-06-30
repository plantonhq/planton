package module

import (
	"strconv"

	awsfsxontapstoragevirtualmachinev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsfsxontapstoragevirtualmachine/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxOntapStorageVirtualMachine *awsfsxontapstoragevirtualmachinev1.AwsFsxOntapStorageVirtualMachine
	AwsTags                          map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awsfsxontapstoragevirtualmachinev1.AwsFsxOntapStorageVirtualMachineStackInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxOntapStorageVirtualMachine = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxOntapStorageVirtualMachine.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxOntapStorageVirtualMachine.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsFsxOntapStorageVirtualMachine.String(),
		awstagkeys.ResourceId:   locals.AwsFsxOntapStorageVirtualMachine.Metadata.Id,
	}

	return locals
}
