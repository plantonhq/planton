package component

import (
	"errors"
	"strings"
	"testing"
	"time"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// The definition's four arms map onto the rendering options field for
// field, and the built-in seal maps onto nothing.
func TestSealOptionsFrom_EveryArm(t *testing.T) {
	if sealOptionsFrom(vaultTestPlatform(nil)) != nil || sealOptionsFrom(vaultTestPlatform(&v1.OpenBAOSpec{})) != nil {
		t.Error("no declared seal is the built-in seal")
	}

	aws := sealOptionsFrom(vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: &v1.OpenBAOAutoUnsealSpec{AwsKms: &v1.OpenBAOAwsKmsSealSpec{
		Region: "us-east-1", KmsKeyID: "alias/k", AccessKeyID: "AKIA", CredentialsSecretName: "aws-creds"}}}))
	if aws.AwsKms == nil || aws.AwsKms.Region != "us-east-1" || aws.AwsKms.KmsKeyID != "alias/k" || aws.AwsKms.AccessKeyID != "AKIA" || aws.AwsKms.CredentialsSecretName != "aws-creds" {
		t.Errorf("aws arm = %+v", aws.AwsKms)
	}
	gcp := sealOptionsFrom(vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: &v1.OpenBAOAutoUnsealSpec{GcpKms: &v1.OpenBAOGcpKmsSealSpec{
		Project: "p", Region: "global", KeyRing: "r", CryptoKey: "k"}}}))
	if gcp.GcpKms == nil || gcp.GcpKms.Project != "p" || gcp.GcpKms.Region != "global" || gcp.GcpKms.KeyRing != "r" || gcp.GcpKms.CryptoKey != "k" {
		t.Errorf("gcp arm = %+v", gcp.GcpKms)
	}
	az := sealOptionsFrom(vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: &v1.OpenBAOAutoUnsealSpec{AzureKeyVault: &v1.OpenBAOAzureKeyVaultSealSpec{
		VaultName: "v", KeyName: "k", TenantID: "t", ClientID: "c", CredentialsSecretName: "az-creds"}}}))
	if az.AzureKeyVault == nil || az.AzureKeyVault.VaultName != "v" || az.AzureKeyVault.KeyName != "k" || az.AzureKeyVault.TenantID != "t" || az.AzureKeyVault.ClientID != "c" || az.AzureKeyVault.CredentialsSecretName != "az-creds" {
		t.Errorf("azure arm = %+v", az.AzureKeyVault)
	}
	tr := sealOptionsFrom(vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: &v1.OpenBAOAutoUnsealSpec{Transit: &v1.OpenBAOTransitSealSpec{
		Address: "https://bao:8200", KeyName: "k", MountPath: "seals/", CredentialsSecretName: "tok"}}}))
	if tr.Transit == nil || tr.Transit.Address != "https://bao:8200" || tr.Transit.KeyName != "k" || tr.Transit.MountPath != "seals/" || tr.Transit.CredentialsSecretName != "tok" {
		t.Errorf("transit arm = %+v", tr.Transit)
	}

	if vaultServiceAccountAnnotations(vaultTestPlatform(nil)) != nil {
		t.Error("no annotations when none are declared")
	}
	ann := vaultServiceAccountAnnotations(vaultTestPlatform(&v1.OpenBAOSpec{ServiceAccountAnnotations: map[string]string{"iam.gke.io/gcp-service-account": "sa@p.iam.gserviceaccount.com"}}))
	if ann["iam.gke.io/gcp-service-account"] != "sa@p.iam.gserviceaccount.com" {
		t.Errorf("annotations = %v", ann)
	}
}

// Every sentence a person meets when a seal misbehaves: one cause each, the
// grant or the check named per arm, the next step stated.
func TestSealSentences(t *testing.T) {
	arms := map[string]*resources.OpenBAOSealOptions{
		"aws":     {AwsKms: &resources.OpenBAOAwsKmsSealOptions{Region: "r", KmsKeyID: "k"}},
		"gcp":     {GcpKms: &resources.OpenBAOGcpKmsSealOptions{Project: "p", Region: "r", KeyRing: "kr", CryptoKey: "k"}},
		"azure":   {AzureKeyVault: &resources.OpenBAOAzureKeyVaultSealOptions{VaultName: "v", KeyName: "k", TenantID: "t"}},
		"transit": {Transit: &resources.OpenBAOTransitSealOptions{Address: "a", KeyName: "k"}},
	}
	grants := map[string]string{"aws": "kms:Decrypt", "gcp": "cryptoKeyEncrypterDecrypter", "azure": "unwrapKey", "transit": "decrypt on the transit key"}
	for name, seal := range arms {
		t.Run(name, func(t *testing.T) {
			young := sealNotOpenedMessage(seal, 10*time.Second)
			mustContain(t, young, "is opening it", "every five seconds")
			if contains(young, "cannot decrypt") {
				t.Error("a young pod is opening, not blocked")
			}
			old := sealNotOpenedMessage(seal, 3*time.Minute)
			mustContain(t, old, "has not opened it after 3m0s", "cannot decrypt with it", grants[name], "opens on its own once the grant lands")

			hint := sealStartHint(seal)
			mustContain(t, hint, "seal check the server makes at start", `"Error configuring seal"`)
		})
	}
	mustContain(t, sealStartHint(arms["gcp"]), "roles/cloudkms.viewer", "roles/cloudkms.cryptoKeyEncrypterDecrypter")

	changed := sealChangedMessage("shamir", "gcpKms p/r/kr/k", "my-vault-keys")
	mustContain(t, changed, `initialized with the seal "shamir"`, "my-vault-keys", `now says "gcpKms p/r/kr/k"`, "decided when the vault is created", "restore the declaration", "keeps the recorded seal")
}

// The rare, serious sentence: init succeeded and the keys could not be
// saved. Under the built-in seal the vault is lost; under a cloud seal it
// still opens but its break-glass is gone. Both say how to reset.
func TestInitSecretWriteFailedMessage(t *testing.T) {
	err := errors.New("the API server is unavailable")
	shamir := initSecretWriteFailedMessage(nil, "my-vault-keys", err)
	mustContain(t, shamir, "could not be written to Secret my-vault-keys", "the API server is unavailable", "cannot be opened and cannot be initialized again", `drop and recreate the database "openbao"`)

	cloud := initSecretWriteFailedMessage(&resources.OpenBAOSealOptions{GcpKms: &resources.OpenBAOGcpKmsSealOptions{}}, "my-vault-keys", err)
	mustContain(t, cloud, "still opens itself from your key", "break-glass", "cannot be initialized again")
}

// An open vault with no Secret: a Ready sentence that says the platform
// works, where the vault came from, which seal opened it, what the missing
// Secret costs, and both ways back. Under the built-in seal the cost is the
// next restart; under a cloud seal, the break-glass.
func TestOpenVaultWithoutSecretMessage(t *testing.T) {
	planton := vaultTestPlatform(&v1.OpenBAOSpec{InitSecretName: "my-vault-keys"})
	shamir := openVaultWithoutSecretMessage(planton, nil, "my-vault-keys")
	mustContain(t, shamir, "OpenBAO healthy, but Secret my-vault-keys does not exist", "works until the vault next restarts", "unseal keys are the only way", "Recreate it from the copy you kept", "renamed spec.vault.initSecretName")

	planton.Status.Backup = &v1.BackupStatus{RestoredFrom: "planton-postgres-deadbeef"}
	cloud := openVaultWithoutSecretMessage(planton, &resources.OpenBAOSealOptions{AwsKms: &resources.OpenBAOAwsKmsSealOptions{}}, "my-vault-keys")
	mustContain(t, cloud, "came back from archive planton-postgres-deadbeef", "an AWS KMS key opened it", "the platform works", "break-glass", "gone until the Secret is back")
	if strings.Contains(shamir, "sign in with") || strings.Contains(cloud, "sign in with") {
		t.Error("the sentence must not describe the root token as anything's sign-in")
	}
}

// The init Secret's contract in one place: which data key a seal's keys live
// under, and what "holds keys" means.
func TestInitSecretContract(t *testing.T) {
	if initSecretKeysKey(nil) != resources.OpenBAOInitSecretUnsealKeysKey {
		t.Error("the built-in seal's keys are unseal keys")
	}
	if initSecretKeysKey(&resources.OpenBAOSealOptions{Transit: &resources.OpenBAOTransitSealOptions{}}) != resources.OpenBAOInitSecretRecoveryKeysKey {
		t.Error("a cloud seal's keys are recovery keys")
	}
	if initSecretHoldsKeys(nil) {
		t.Error("no Secret holds no keys")
	}
	for _, key := range []string{resources.OpenBAOInitSecretUnsealKeysKey, resources.OpenBAOInitSecretRecoveryKeysKey, resources.OpenBAOInitSecretRootTokenKey} {
		s := transitCredentials()
		s.Data = map[string][]byte{key: []byte("x")}
		if !initSecretHoldsKeys(s) {
			t.Errorf("a Secret with %s holds keys", key)
		}
	}
	empty := transitCredentials()
	empty.Data = map[string][]byte{"unrelated": []byte("x")}
	if initSecretHoldsKeys(empty) {
		t.Error("a Secret with unrelated data holds no vault keys")
	}
}
