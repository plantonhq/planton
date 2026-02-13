package module

import (
	azurevirtualmachinev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurevirtualmachine/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureVirtualMachine *azurevirtualmachinev1.AzureVirtualMachine
	ResourceGroupName   string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurevirtualmachinev1.AzureVirtualMachineStackInput) *Locals {
	locals := &Locals{}

	locals.AzureVirtualMachine = stackInput.Target
	target := stackInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware resolves
	// valueFrom references before IaC modules run, so .GetValue() always returns the
	// resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	return locals
}
