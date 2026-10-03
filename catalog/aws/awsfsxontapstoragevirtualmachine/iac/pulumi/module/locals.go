package module

import (
	"strconv"

	awsfsxontapstoragevirtualmachinev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsfsxontapstoragevirtualmachine/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsFsxOntapStorageVirtualMachine *awsfsxontapstoragevirtualmachinev1alpha1.AwsFsxOntapStorageVirtualMachine
	AwsTags                          map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsfsxontapstoragevirtualmachinev1alpha1.AwsFsxOntapStorageVirtualMachineIacInput) *Locals {
	locals := &Locals{}
	locals.AwsFsxOntapStorageVirtualMachine = iacInput.Target

	// Resource-identity tags follow the catalog convention. The Name tag is
	// the resource's metadata.name — distinct from spec.name, the
	// ONTAP-internal SVM identity; the Terraform module pins the same basis,
	// keeping the two engines' physical identity converged.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsFsxOntapStorageVirtualMachine.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsFsxOntapStorageVirtualMachine.Metadata.Org,
		awstagkeys.Environment:  locals.AwsFsxOntapStorageVirtualMachine.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsFsxOntapStorageVirtualMachine.String(),
		awstagkeys.ResourceId:   locals.AwsFsxOntapStorageVirtualMachine.Metadata.Id,
	}

	return locals
}
