package module

import (
	azurebackuppolicyvmv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackuppolicyvm/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureBackupPolicyVm *azurebackuppolicyvmv1alpha1.AzureBackupPolicyVm

	// ResourceGroupName and RecoveryVaultName are StringValueOrRef
	// fields; the platform middleware resolves valueFrom references
	// before IaC modules run, so GetValue() always returns the
	// resolved literal names.
	ResourceGroupName string
	RecoveryVaultName string
}

// The backup policy resource carries NO tags argument on the provider
// (ARM backup policies are untagged), so this module derives no tag
// map -- deliberately unlike its vault sibling.

func initializeLocals(ctx *pulumi.Context, iacInput *azurebackuppolicyvmv1alpha1.AzureBackupPolicyVmIacInput) *Locals {
	locals := &Locals{}

	locals.AzureBackupPolicyVm = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.RecoveryVaultName = target.Spec.RecoveryVaultName.GetValue()

	return locals
}
