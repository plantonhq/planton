package module

import (
	azurebackuppolicyfilesharev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurebackuppolicyfileshare/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureBackupPolicyFileShare *azurebackuppolicyfilesharev1alpha1.AzureBackupPolicyFileShare

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

func initializeLocals(ctx *pulumi.Context, iacInput *azurebackuppolicyfilesharev1alpha1.AzureBackupPolicyFileShareIacInput) *Locals {
	locals := &Locals{}

	locals.AzureBackupPolicyFileShare = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.RecoveryVaultName = target.Spec.RecoveryVaultName.GetValue()

	return locals
}
