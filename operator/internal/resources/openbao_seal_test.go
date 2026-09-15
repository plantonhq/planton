package resources

import (
	"reflect"
	"strings"
	"testing"
)

// The built-in seal is the absence of a seal: no stanza, no environment, no
// credential, and the word every status sentence uses for it.
func TestOpenBAOSealOptions_BuiltInSeal(t *testing.T) {
	var seal *OpenBAOSealOptions
	if seal.Word() != OpenBAOSealShamir || seal.Fingerprint() != OpenBAOSealShamir {
		t.Errorf("a nil seal is the built-in seal, got word=%q fingerprint=%q", seal.Word(), seal.Fingerprint())
	}
	if seal.Stanza() != "" || seal.PlainEnv() != nil || seal.SecretEnv() != nil || seal.CredentialsSecretName() != "" {
		t.Errorf("a nil seal renders nothing, got stanza=%q plain=%v secret=%v", seal.Stanza(), seal.PlainEnv(), seal.SecretEnv())
	}
	if seal.Human() != "the built-in key shares" {
		t.Errorf("Human() = %q", seal.Human())
	}
}

// Each arm's stanza carries the seal type and the parameters that name the
// key -- byte-faithful to the standalone kind's rendering -- and never a
// credential parameter.
func TestOpenBAOSealOptions_Stanza(t *testing.T) {
	cases := []struct {
		name string
		seal *OpenBAOSealOptions
		want string
	}{
		{
			"aws",
			&OpenBAOSealOptions{AwsKms: &OpenBAOAwsKmsSealOptions{Region: "us-east-1", KmsKeyID: "alias/k", AccessKeyID: "AKIA", CredentialsSecretName: "s"}},
			"seal \"awskms\" {\n  region = \"us-east-1\"\n  kms_key_id = \"alias/k\"\n}\n",
		},
		{
			"gcp",
			&OpenBAOSealOptions{GcpKms: &OpenBAOGcpKmsSealOptions{Project: "p", Region: "global", KeyRing: "r", CryptoKey: "k"}},
			"seal \"gcpckms\" {\n  project = \"p\"\n  region = \"global\"\n  key_ring = \"r\"\n  crypto_key = \"k\"\n}\n",
		},
		{
			"azure keyless",
			&OpenBAOSealOptions{AzureKeyVault: &OpenBAOAzureKeyVaultSealOptions{VaultName: "v", KeyName: "k", TenantID: "t"}},
			"seal \"azurekeyvault\" {\n  vault_name = \"v\"\n  key_name = \"k\"\n  tenant_id = \"t\"\n}\n",
		},
		{
			"azure service principal",
			&OpenBAOSealOptions{AzureKeyVault: &OpenBAOAzureKeyVaultSealOptions{VaultName: "v", KeyName: "k", TenantID: "t", ClientID: "c", CredentialsSecretName: "s"}},
			"seal \"azurekeyvault\" {\n  vault_name = \"v\"\n  key_name = \"k\"\n  tenant_id = \"t\"\n  client_id = \"c\"\n}\n",
		},
		{
			"transit default mount",
			&OpenBAOSealOptions{Transit: &OpenBAOTransitSealOptions{Address: "https://bao:8200", KeyName: "k", CredentialsSecretName: "s"}},
			"seal \"transit\" {\n  address = \"https://bao:8200\"\n  key_name = \"k\"\n  mount_path = \"transit/\"\n}\n",
		},
		{
			"transit custom mount",
			&OpenBAOSealOptions{Transit: &OpenBAOTransitSealOptions{Address: "https://bao:8200", KeyName: "k", MountPath: "seals/"}},
			"seal \"transit\" {\n  address = \"https://bao:8200\"\n  key_name = \"k\"\n  mount_path = \"seals/\"\n}\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.seal.Stanza(); got != tc.want {
				t.Errorf("stanza\n got: %q\nwant: %q", got, tc.want)
			}
			for _, f := range []string{"secret_key", "client_secret", "token", "credentials", "AKIA", "access_key"} {
				if strings.Contains(tc.seal.Stanza(), f) {
					t.Errorf("stanza must not carry %q", f)
				}
			}
		})
	}
}

// Identifiers ride plain variables, credentials ride the named Secret keyed
// by the variable name, and a keyless arm needs neither.
func TestOpenBAOSealOptions_Environment(t *testing.T) {
	aws := &OpenBAOSealOptions{AwsKms: &OpenBAOAwsKmsSealOptions{Region: "us-east-1", KmsKeyID: "k", AccessKeyID: "AKIA", CredentialsSecretName: "aws-creds"}}
	if got := aws.PlainEnv(); !reflect.DeepEqual(got, map[string]string{OpenBAOSealEnvAwsRegion: "us-east-1", OpenBAOSealEnvAwsAccessKeyID: "AKIA"}) {
		t.Errorf("aws plain env = %v", got)
	}
	if got := aws.SecretEnv(); !reflect.DeepEqual(got, []OpenBAOSecretEnv{{EnvName: OpenBAOSealEnvAwsSecretAccessKey, SecretName: "aws-creds", SecretKey: OpenBAOSealEnvAwsSecretAccessKey}}) {
		t.Errorf("aws secret env = %v", got)
	}
	if got := aws.RequiredCredentialKeys(); !reflect.DeepEqual(got, []string{OpenBAOSealEnvAwsSecretAccessKey}) {
		t.Errorf("aws required keys = %v", got)
	}

	awsKeyless := &OpenBAOSealOptions{AwsKms: &OpenBAOAwsKmsSealOptions{Region: "us-east-1", KmsKeyID: "k"}}
	if got := awsKeyless.PlainEnv(); !reflect.DeepEqual(got, map[string]string{OpenBAOSealEnvAwsRegion: "us-east-1"}) {
		t.Errorf("keyless aws plain env = %v", got)
	}
	if awsKeyless.SecretEnv() != nil || awsKeyless.RequiredCredentialKeys() != nil || awsKeyless.CredentialsSecretName() != "" {
		t.Error("a keyless arm names no Secret and projects no credential")
	}

	gcp := &OpenBAOSealOptions{GcpKms: &OpenBAOGcpKmsSealOptions{Project: "p", Region: "global", KeyRing: "r", CryptoKey: "k"}}
	if gcp.PlainEnv() != nil || gcp.SecretEnv() != nil || gcp.CredentialsSecretName() != "" {
		t.Error("the GCP arm is keyless by construction: its identity is a ServiceAccount annotation")
	}

	azure := &OpenBAOSealOptions{AzureKeyVault: &OpenBAOAzureKeyVaultSealOptions{VaultName: "v", KeyName: "k", TenantID: "t", ClientID: "c", CredentialsSecretName: "az-creds"}}
	if got := azure.SecretEnv(); !reflect.DeepEqual(got, []OpenBAOSecretEnv{{EnvName: OpenBAOSealEnvAzureClientSecret, SecretName: "az-creds", SecretKey: OpenBAOSealEnvAzureClientSecret}}) {
		t.Errorf("azure secret env = %v", got)
	}

	transit := &OpenBAOSealOptions{Transit: &OpenBAOTransitSealOptions{Address: "https://bao:8200", KeyName: "k", CredentialsSecretName: "bao-token"}}
	if got := transit.SecretEnv(); !reflect.DeepEqual(got, []OpenBAOSecretEnv{{EnvName: OpenBAOSealEnvTransitToken, SecretName: "bao-token", SecretKey: OpenBAOSealEnvTransitToken}}) {
		t.Errorf("transit secret env = %v", got)
	}
}

// The fingerprint is the seal word plus the key it names: the same key is
// the same seal, a different key or a different type is a different seal.
func TestOpenBAOSealOptions_Fingerprint(t *testing.T) {
	a := &OpenBAOSealOptions{GcpKms: &OpenBAOGcpKmsSealOptions{Project: "p", Region: "global", KeyRing: "r", CryptoKey: "k"}}
	same := &OpenBAOSealOptions{GcpKms: &OpenBAOGcpKmsSealOptions{Project: "p", Region: "global", KeyRing: "r", CryptoKey: "k"}}
	otherKey := &OpenBAOSealOptions{GcpKms: &OpenBAOGcpKmsSealOptions{Project: "p", Region: "global", KeyRing: "r", CryptoKey: "k2"}}
	if a.Fingerprint() != same.Fingerprint() {
		t.Error("the same key must fingerprint the same")
	}
	if a.Fingerprint() == otherKey.Fingerprint() {
		t.Error("a different key must fingerprint differently")
	}
	if got := a.Fingerprint(); got != "gcpKms p/global/r/k" {
		t.Errorf("gcp fingerprint = %q", got)
	}
	aws := &OpenBAOSealOptions{AwsKms: &OpenBAOAwsKmsSealOptions{Region: "us-east-1", KmsKeyID: "alias/k", AccessKeyID: "rotated-identifier"}}
	if got := aws.Fingerprint(); got != "awsKms us-east-1/alias/k" {
		t.Errorf("aws fingerprint must name the key, never the credential identifier, got %q", got)
	}
	tr := &OpenBAOSealOptions{Transit: &OpenBAOTransitSealOptions{Address: "https://bao:8200", KeyName: "k"}}
	if got := tr.Fingerprint(); got != "transit https://bao:8200/transit/k" {
		t.Errorf("transit fingerprint = %q", got)
	}
	az := &OpenBAOSealOptions{AzureKeyVault: &OpenBAOAzureKeyVaultSealOptions{VaultName: "v", KeyName: "k", TenantID: "t"}}
	if got := az.Fingerprint(); got != "azureKeyVault v/k" {
		t.Errorf("azure fingerprint = %q", got)
	}
}

// Every arm names the grant a sealed vault is missing and what its wrapper
// checks at start; the built-in seal names neither.
func TestOpenBAOSealOptions_Grants(t *testing.T) {
	arms := map[string]*OpenBAOSealOptions{
		OpenBAOSealAwsKms:        {AwsKms: &OpenBAOAwsKmsSealOptions{Region: "r", KmsKeyID: "k"}},
		OpenBAOSealGcpKms:        {GcpKms: &OpenBAOGcpKmsSealOptions{Project: "p", Region: "r", KeyRing: "kr", CryptoKey: "k"}},
		OpenBAOSealAzureKeyVault: {AzureKeyVault: &OpenBAOAzureKeyVaultSealOptions{VaultName: "v", KeyName: "k", TenantID: "t"}},
		OpenBAOSealTransit:       {Transit: &OpenBAOTransitSealOptions{Address: "a", KeyName: "k"}},
	}
	for word, seal := range arms {
		if seal.Word() != word {
			t.Errorf("Word() = %q, want %q", seal.Word(), word)
		}
		if seal.DecryptGrant() == "" || seal.StartCheck() == "" {
			t.Errorf("%s must name its decrypt grant and its start check", word)
		}
	}
	if got := arms[OpenBAOSealGcpKms].StartCheck(); !strings.Contains(got, "roles/cloudkms.viewer") || !strings.Contains(got, "cryptoKeyEncrypterDecrypter") {
		t.Errorf("the GCP start check must name both roles, got %q", got)
	}
	var none *OpenBAOSealOptions
	if none.DecryptGrant() != "" || none.StartCheck() != "" {
		t.Error("the built-in seal has no grant to name")
	}
}
