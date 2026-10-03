package module

import (
	"strings"

	azuredataprotectionbackuppolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuredataprotectionbackuppolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDataProtectionBackupPolicy *azuredataprotectionbackuppolicyv1alpha1.AzureDataProtectionBackupPolicy

	// VaultId is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so
	// GetValue() always returns the resolved literal ARM ID.
	VaultId string

	// The kubernetes_cluster variant is the one provider resource that
	// addresses the vault by NAME + resource group instead of by ID.
	// ARM vault IDs are structured
	// (/subscriptions/{sub}/resourceGroups/{rg}/providers/
	//  Microsoft.DataProtection/backupVaults/{name}), so both derive
	// deterministically from the spec's single vault_id reference.
	VaultResourceGroupName string
	VaultName              string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuredataprotectionbackuppolicyv1alpha1.AzureDataProtectionBackupPolicyIacInput) *Locals {
	locals := &Locals{}

	locals.AzureDataProtectionBackupPolicy = iacInput.Target

	locals.VaultId = iacInput.Target.Spec.VaultId.GetValue()

	// Note: policies carry NO tags argument (they are pure
	// configuration objects on the vault) -- there is no tag map here,
	// unlike sibling modules.
	idParts := strings.Split(locals.VaultId, "/")
	if len(idParts) > 8 {
		locals.VaultResourceGroupName = idParts[4]
		locals.VaultName = idParts[8]
	}

	return locals
}
