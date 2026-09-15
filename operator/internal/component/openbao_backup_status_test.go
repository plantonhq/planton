package component

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/plantonhq/planton/operator/api/v1"
)

// status.backup.vault follows from the declaration and the backup state: the
// four postures and every seal word, each with the sentence a person plans
// the bad day from.

func vaultPlatform(vault *v1.OpenBAOSpec) *v1.PlantonPlatform {
	return &v1.PlantonPlatform{
		ObjectMeta: metav1.ObjectMeta{Name: "planton", Namespace: "planton"},
		Spec:       v1.PlantonPlatformSpec{Version: "v0.0.62", Vault: vault},
	}
}

func TestVaultBackupStatus_Postures(t *testing.T) {
	off := false

	t.Run("an opted-out vault has nothing to archive", func(t *testing.T) {
		got := vaultBackupStatus(vaultPlatform(&v1.OpenBAOSpec{Enabled: &off}), v1.BackupStateHealthy)
		if got.Covered || got.Seal != "" || got.InitSecretName != "" || !strings.Contains(got.Message, "not deployed") {
			t.Errorf("got %+v", got)
		}
	})
	t.Run("no backup declared: not covered, in words, with the seal and Secret still named", func(t *testing.T) {
		got := vaultBackupStatus(vaultPlatform(nil), v1.BackupStateNotConfigured)
		if got.Covered || got.Seal != "shamir" || got.InitSecretName != "planton-openbao-init" || !strings.Contains(got.Message, "No backup is declared") {
			t.Errorf("got %+v", got)
		}
	})
	t.Run("the built-in seal with the adopter's own Secret: covered, the Secret to keep", func(t *testing.T) {
		got := vaultBackupStatus(vaultPlatform(&v1.OpenBAOSpec{InitSecretName: "planton-vault-keys"}), v1.BackupStateHealthy)
		if !got.Covered || got.Seal != "shamir" || got.InitSecretName != "planton-vault-keys" {
			t.Fatalf("got %+v", got)
		}
		for _, want := range []string{"carries the vault", "planton-vault-keys, which you own", "restore needs that Secret present", "keep a copy"} {
			if !strings.Contains(got.Message, want) {
				t.Errorf("message lacks %q: %s", want, got.Message)
			}
		}
	})
	t.Run("the built-in seal with the operator's Secret: covered, and the honest warning", func(t *testing.T) {
		got := vaultBackupStatus(vaultPlatform(nil), v1.BackupStateDeploying)
		if !got.Covered || got.InitSecretName != "planton-openbao-init" {
			t.Fatalf("got %+v", got)
		}
		for _, want := range []string{"deleted with the platform", "spec.vault.initSecretName", "spec.vault.autoUnseal"} {
			if !strings.Contains(got.Message, want) {
				t.Errorf("message lacks %q: %s", want, got.Message)
			}
		}
	})
	t.Run("every cloud seal: covered, opens itself, the Secret is break-glass", func(t *testing.T) {
		seals := map[string]*v1.OpenBAOAutoUnsealSpec{
			"awsKms":        {AwsKms: &v1.OpenBAOAwsKmsSealSpec{Region: "us-west-2", KmsKeyID: "alias/x"}},
			"gcpKms":        {GcpKms: &v1.OpenBAOGcpKmsSealSpec{Project: "p", Region: "global", KeyRing: "r", CryptoKey: "k"}},
			"azureKeyVault": {AzureKeyVault: &v1.OpenBAOAzureKeyVaultSealSpec{VaultName: "v", KeyName: "k", TenantID: "t"}},
			"transit":       {Transit: &v1.OpenBAOTransitSealSpec{Address: "http://key-holder:8200", KeyName: "k"}},
		}
		for word, seal := range seals {
			got := vaultBackupStatus(vaultPlatform(&v1.OpenBAOSpec{AutoUnseal: seal, InitSecretName: "planton-vault-keys"}), v1.BackupStateHealthy)
			if !got.Covered || got.Seal != word {
				t.Errorf("%s: got %+v", word, got)
			}
			for _, want := range []string{"opens itself", "planton-vault-keys", "break-glass"} {
				if !strings.Contains(got.Message, want) {
					t.Errorf("%s: message lacks %q: %s", word, want, got.Message)
				}
			}
		}
	})
	t.Run("an autoUnseal block with no arm reads as the built-in seal", func(t *testing.T) {
		if got := vaultSealWord(vaultPlatform(&v1.OpenBAOSpec{AutoUnseal: &v1.OpenBAOAutoUnsealSpec{}})); got != "shamir" {
			t.Errorf("seal = %q", got)
		}
	})
}

// A vault that is initialized while the operator holds no keys says where it
// came from, which Secret is missing with its two keys, and the two ways out.
func TestSealedWithoutKeysMessage(t *testing.T) {
	restored := vaultPlatform(nil)
	restored.Status.Backup = &v1.BackupStatus{State: v1.BackupStateDeploying, RestoredFrom: "planton-postgres-deadbeef"}
	msg := sealedWithoutKeysMessage(restored, "planton-openbao-init")
	for _, want := range []string{
		"came back from archive planton-postgres-deadbeef",
		"Secret planton-openbao-init (keys unseal-keys, root-token) does not exist in namespace planton",
		"Recreate that Secret from the copy you kept",
		"spec.vault.autoUnseal",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message lacks %q: %s", want, msg)
		}
	}

	fresh := vaultPlatform(nil)
	msg = sealedWithoutKeysMessage(fresh, "planton-openbao-init")
	if strings.Contains(msg, "came back from archive") || !strings.HasPrefix(msg, "The vault is initialized and sealed") {
		t.Errorf("a platform never restored must not name an archive: %s", msg)
	}
}
