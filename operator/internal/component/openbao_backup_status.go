package component

import (
	"fmt"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// The archive's answer about the vault. Every fact here follows from the
// declaration and the database's backup state -- the vault stores in the
// platform's database, so "is it archived" is "is the database archived",
// and "what opens the restored vault" is what the spec declares -- so the
// slot is written beside status.backup on every pass, by the component that
// writes status.backup, without probing the vault.

// vaultSealWord names what opens the vault: the cloud seal the spec declares,
// or the built-in key shares when it declares none -- the one vocabulary the
// rendering, the init Secret's annotation, and this status share.
func vaultSealWord(planton *v1.PlantonPlatform) string {
	return sealOptionsFrom(planton).Word()
}

// vaultInitSecretName is the Secret that holds the vault's keys and root
// token: the adopter's own when the spec names one, the operator's otherwise.
func vaultInitSecretName(planton *v1.PlantonPlatform) string {
	if planton.Spec.Vault != nil && planton.Spec.Vault.InitSecretName != "" {
		return planton.Spec.Vault.InitSecretName
	}
	return resources.OpenBAOInitSecretName(planton.Name)
}

// vaultInitSecretIsAdoptersOwn reports whether the keys Secret outlives the
// platform: named by the adopter, the operator never deletes it; unnamed, it
// is owner-referenced and dies with the platform.
func vaultInitSecretIsAdoptersOwn(planton *v1.PlantonPlatform) bool {
	return planton.Spec.Vault != nil && planton.Spec.Vault.InitSecretName != ""
}

// vaultBackupStatus derives status.backup.vault from the declaration and the
// database's backup state.
func vaultBackupStatus(planton *v1.PlantonPlatform, state v1.BackupState) *v1.VaultBackupStatus {
	if !isVaultEnabled(planton) {
		return &v1.VaultBackupStatus{
			Covered: false,
			Message: "The bundled vault is not deployed; there is nothing to archive.",
		}
	}
	seal := sealOptionsFrom(planton)
	initSecret := vaultInitSecretName(planton)
	if state == v1.BackupStateNotConfigured || state == "" {
		return &v1.VaultBackupStatus{
			Covered:        false,
			Seal:           seal.Word(),
			InitSecretName: initSecret,
			Message:        "No backup is declared; the vault's data lives in this database and nothing copies it anywhere.",
		}
	}
	return &v1.VaultBackupStatus{
		Covered:        true,
		Seal:           seal.Word(),
		InitSecretName: initSecret,
		Message:        vaultCoverageMessage(seal, initSecret, vaultInitSecretIsAdoptersOwn(planton)),
	}
}

// vaultCoverageMessage is the sentence a person planning for the bad day
// reads: the archive carries the vault; here is what opens it and what to
// keep. Under the built-in seal the keys Secret IS the way back in; under a
// cloud seal the restored vault opens itself and the Secret is break-glass.
func vaultCoverageMessage(seal *resources.OpenBAOSealOptions, initSecret string, adoptersOwn bool) string {
	const carries = "The archive carries the vault -- it stores in this database. "
	if seal.Word() == resources.OpenBAOSealShamir {
		if adoptersOwn {
			return carries + fmt.Sprintf("It is sealed with the built-in key shares held in Secret %s, which you own: "+
				"a restore needs that Secret present in the new cluster, and it is the one object to keep a copy of outside the cluster.", initSecret)
		}
		return carries + fmt.Sprintf("It is sealed with the built-in key shares held in Secret %s, which is deleted with the platform: "+
			"a restore cannot open the vault unless you keep a copy of that Secret outside the cluster and recreate it first -- "+
			"name a Secret you own in spec.vault.initSecretName, or seal the vault with a cloud key (spec.vault.autoUnseal).", initSecret)
	}
	return carries + fmt.Sprintf("It is sealed by %s: a restored vault opens itself from your key. "+
		"Secret %s holds the root token and recovery keys (the vault's break-glass) -- keep a copy of it outside the cluster.", seal.Human(), initSecret)
}

// sealedWithoutKeysMessage is the vault component's sentence for a vault
// that is initialized and sealed while the Secret that should hold its keys
// does not exist: the restored case above all (the data came back from an
// archive; the keys did not), and the case of a Secret deleted from under a
// running platform. One root cause, the object named, the two ways out.
func sealedWithoutKeysMessage(planton *v1.PlantonPlatform, initSecret string) string {
	origin := "The vault is initialized and sealed"
	if planton.Status.Backup != nil && planton.Status.Backup.RestoredFrom != "" {
		origin = fmt.Sprintf("The vault came back from archive %s initialized and sealed", planton.Status.Backup.RestoredFrom)
	}
	return fmt.Sprintf("%s, but Secret %s (keys %s, %s) does not exist in namespace %s, so the operator holds no keys to open it. "+
		"Recreate that Secret from the copy you kept; or, on a platform declared with spec.vault.autoUnseal, the restored vault opens itself from your cloud key.",
		origin, initSecret, resources.OpenBAOInitSecretUnsealKeysKey, resources.OpenBAOInitSecretRootTokenKey, planton.Namespace)
}
