package resources

import (
	"fmt"
	"sort"
	"strings"
)

// The vault's seal is the one thing a restore cannot bring back: the archive
// carries the vault's data encrypted under a key the seal protects, so what
// opens a restored vault is decided here, at creation, and never changes
// after. This file renders a declared seal into the three seams the chart
// offers -- the `seal` stanza in the server's config (non-credential
// parameters only; the config lands in a ConfigMap), plain environment
// variables for public identifiers, and Secret-projected environment
// variables for credentials -- and names what a person must grant before
// the server starts, because every seal wrapper probes its key when the
// seal is configured, before init can open.
//
// Reference: the standalone OpenBao kind's module (bao_config.go renders the
// same stanzas; seal_secret.go delivers the same variables) -- the shapes
// are byte-faithful so an agent who learned one vault has learned both.

// The environment variables the seal wrappers read their credentials and
// identifiers from -- the cloud SDKs' standard variables; transit follows
// the api client's token variable. A declared credentials Secret is keyed
// by exactly these names, so each key is projected as the variable of the
// same name and nothing is renamed on the way.
const (
	OpenBAOSealEnvAwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	OpenBAOSealEnvAwsRegion          = "AWS_REGION"
	OpenBAOSealEnvAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	OpenBAOSealEnvAzureClientSecret  = "AZURE_CLIENT_SECRET"
	OpenBAOSealEnvTransitToken       = "VAULT_TOKEN"
)

// The words status and annotations speak for the seal. Strings, never a CRD
// enum, for the reason BackupState is not one: an older installed definition
// must not refuse a newer operator's word.
const (
	OpenBAOSealShamir        = "shamir"
	OpenBAOSealAwsKms        = "awsKms"
	OpenBAOSealGcpKms        = "gcpKms"
	OpenBAOSealAzureKeyVault = "azureKeyVault"
	OpenBAOSealTransit       = "transit"
)

// OpenBAOSealAnnotation is written on the init Secret at initialization and
// carries the seal's fingerprint (Fingerprint below). The seal is decided
// when the vault is created: the server refuses to start against storage
// initialized under a different seal, so the operator compares the
// declaration with this record before rendering and refuses a change in
// words instead of letting the pod crash-loop.
const OpenBAOSealAnnotation = "planton.ai/openbao-seal"

// OpenBAOSealOptions describes the declared seal: exactly one arm, or nil
// for the built-in key shares. The component builds it from the definition
// (CEL guarantees the one-arm rule); this package renders it.
type OpenBAOSealOptions struct {
	AwsKms        *OpenBAOAwsKmsSealOptions
	GcpKms        *OpenBAOGcpKmsSealOptions
	AzureKeyVault *OpenBAOAzureKeyVaultSealOptions
	Transit       *OpenBAOTransitSealOptions
}

// OpenBAOAwsKmsSealOptions: an AWS KMS key. Keyless through IRSA or an
// instance profile on the vault's pods; a static key pair is the fallback,
// its public half a plain variable and its secret half projected from the
// credentials Secret. The wrapper describes the key at start
// (kms:DescribeKey), wraps on init and unwraps on every unseal (kms:Encrypt,
// kms:Decrypt).
type OpenBAOAwsKmsSealOptions struct {
	Region   string
	KmsKeyID string
	// AccessKeyID is an identifier, not a credential; set only with
	// CredentialsSecretName.
	AccessKeyID string
	// CredentialsSecretName names an adopter-owned Secret carrying
	// OpenBAOSealEnvAwsSecretAccessKey. Empty means keyless.
	CredentialsSecretName string
}

// OpenBAOGcpKmsSealOptions: a Google Cloud KMS key, keyless by construction
// -- the identity is the vault ServiceAccount's Workload Identity annotation
// (or the node's ambient credentials). It needs TWO roles on the key:
// roles/cloudkms.cryptoKeyEncrypterDecrypter to wrap and unwrap, AND
// roles/cloudkms.viewer for the cloudkms.cryptoKeys.get the wrapper makes
// when the seal is configured at start; with only the first the pod
// crash-loops on "Permission 'cloudkms.cryptoKeys.get' denied".
type OpenBAOGcpKmsSealOptions struct {
	Project   string
	Region    string
	KeyRing   string
	CryptoKey string
}

// OpenBAOAzureKeyVaultSealOptions: a key in an Azure Key Vault. Keyless
// through Workload Identity or a managed identity on the vault's pods; a
// service principal is the fallback, its client id in the stanza and its
// secret projected from the credentials Secret. The identity needs get,
// wrapKey, and unwrapKey on the key.
type OpenBAOAzureKeyVaultSealOptions struct {
	VaultName string
	KeyName   string
	TenantID  string
	// ClientID is an identifier, not a credential; set only with
	// CredentialsSecretName.
	ClientID string
	// CredentialsSecretName names an adopter-owned Secret carrying
	// OpenBAOSealEnvAzureClientSecret. Empty means keyless.
	CredentialsSecretName string
}

// OpenBAOTransitSealOptions: the transit engine of a central OpenBao or
// Vault. The wrapper test-encrypts on the key when the seal is configured,
// so the central instance must be reachable and unsealed at every start and
// the token must be able to encrypt and decrypt on the key's path.
type OpenBAOTransitSealOptions struct {
	Address string
	KeyName string
	// MountPath defaults to "transit/".
	MountPath string
	// CredentialsSecretName names an adopter-owned Secret carrying
	// OpenBAOSealEnvTransitToken. Empty means the central instance is
	// reached without a token.
	CredentialsSecretName string
}

// OpenBAOSecretEnv is one Secret-projected environment variable the chart
// renders through server.extraSecretEnvironmentVars.
type OpenBAOSecretEnv struct {
	EnvName    string
	SecretName string
	SecretKey  string
}

// Word names the seal the way status.backup.vault.seal and the init Secret's
// annotation say it: the declared arm, or the built-in key shares when none.
func (o *OpenBAOSealOptions) Word() string {
	switch {
	case o == nil:
		return OpenBAOSealShamir
	case o.AwsKms != nil:
		return OpenBAOSealAwsKms
	case o.GcpKms != nil:
		return OpenBAOSealGcpKms
	case o.AzureKeyVault != nil:
		return OpenBAOSealAzureKeyVault
	case o.Transit != nil:
		return OpenBAOSealTransit
	}
	return OpenBAOSealShamir
}

// Human is the seal as a person reads it in a sentence.
func (o *OpenBAOSealOptions) Human() string {
	switch o.Word() {
	case OpenBAOSealAwsKms:
		return "an AWS KMS key"
	case OpenBAOSealGcpKms:
		return "a Google Cloud KMS key"
	case OpenBAOSealAzureKeyVault:
		return "an Azure Key Vault key"
	case OpenBAOSealTransit:
		return "a central OpenBao's transit key"
	}
	return "the built-in key shares"
}

// Fingerprint is the seal word plus the key it names -- what the init
// Secret records at initialization and what a later declaration is compared
// with. Two declarations that would open the same storage have the same
// fingerprint; a different key ring, key, vault, or address is a different
// seal, and the server would refuse to start against the old storage.
func (o *OpenBAOSealOptions) Fingerprint() string {
	switch {
	case o == nil:
		return OpenBAOSealShamir
	case o.AwsKms != nil:
		return fmt.Sprintf("%s %s/%s", OpenBAOSealAwsKms, o.AwsKms.Region, o.AwsKms.KmsKeyID)
	case o.GcpKms != nil:
		g := o.GcpKms
		return fmt.Sprintf("%s %s/%s/%s/%s", OpenBAOSealGcpKms, g.Project, g.Region, g.KeyRing, g.CryptoKey)
	case o.AzureKeyVault != nil:
		return fmt.Sprintf("%s %s/%s", OpenBAOSealAzureKeyVault, o.AzureKeyVault.VaultName, o.AzureKeyVault.KeyName)
	case o.Transit != nil:
		return fmt.Sprintf("%s %s/%s%s", OpenBAOSealTransit, o.Transit.Address, o.Transit.mountPath(), o.Transit.KeyName)
	}
	return OpenBAOSealShamir
}

// Stanza renders the `seal` block for the server's config: the seal type and
// its NON-credential parameters only. Empty for the built-in seal (a Shamir
// seal is simply not configured). Every line here lands in a ConfigMap.
func (o *OpenBAOSealOptions) Stanza() string {
	if o == nil {
		return ""
	}
	var b strings.Builder
	switch {
	case o.AwsKms != nil:
		b.WriteString("seal \"awskms\" {\n")
		fmt.Fprintf(&b, "  region = \"%s\"\n", o.AwsKms.Region)
		fmt.Fprintf(&b, "  kms_key_id = \"%s\"\n", o.AwsKms.KmsKeyID)
		b.WriteString("}\n")
	case o.GcpKms != nil:
		b.WriteString("seal \"gcpckms\" {\n")
		fmt.Fprintf(&b, "  project = \"%s\"\n", o.GcpKms.Project)
		fmt.Fprintf(&b, "  region = \"%s\"\n", o.GcpKms.Region)
		fmt.Fprintf(&b, "  key_ring = \"%s\"\n", o.GcpKms.KeyRing)
		fmt.Fprintf(&b, "  crypto_key = \"%s\"\n", o.GcpKms.CryptoKey)
		b.WriteString("}\n")
	case o.AzureKeyVault != nil:
		b.WriteString("seal \"azurekeyvault\" {\n")
		fmt.Fprintf(&b, "  vault_name = \"%s\"\n", o.AzureKeyVault.VaultName)
		fmt.Fprintf(&b, "  key_name = \"%s\"\n", o.AzureKeyVault.KeyName)
		fmt.Fprintf(&b, "  tenant_id = \"%s\"\n", o.AzureKeyVault.TenantID)
		if o.AzureKeyVault.ClientID != "" {
			fmt.Fprintf(&b, "  client_id = \"%s\"\n", o.AzureKeyVault.ClientID)
		}
		b.WriteString("}\n")
	case o.Transit != nil:
		b.WriteString("seal \"transit\" {\n")
		fmt.Fprintf(&b, "  address = \"%s\"\n", o.Transit.Address)
		fmt.Fprintf(&b, "  key_name = \"%s\"\n", o.Transit.KeyName)
		fmt.Fprintf(&b, "  mount_path = \"%s\"\n", o.Transit.mountPath())
		b.WriteString("}\n")
	}
	return b.String()
}

// PlainEnv is the NON-secret environment a seal arm needs -- identifiers,
// rendered through the chart's plain extraEnvironmentVars. Nil when the arm
// needs none.
func (o *OpenBAOSealOptions) PlainEnv() map[string]string {
	if o == nil || o.AwsKms == nil {
		return nil
	}
	env := map[string]string{OpenBAOSealEnvAwsRegion: o.AwsKms.Region}
	if o.AwsKms.AccessKeyID != "" {
		env[OpenBAOSealEnvAwsAccessKeyID] = o.AwsKms.AccessKeyID
	}
	return env
}

// CredentialsSecretName returns the adopter-owned credentials Secret the arm
// names, or "" for a keyless posture (the GCP arm is keyless by
// construction and never names one).
func (o *OpenBAOSealOptions) CredentialsSecretName() string {
	switch {
	case o == nil:
		return ""
	case o.AwsKms != nil:
		return o.AwsKms.CredentialsSecretName
	case o.AzureKeyVault != nil:
		return o.AzureKeyVault.CredentialsSecretName
	case o.Transit != nil:
		return o.Transit.CredentialsSecretName
	}
	return ""
}

// RequiredCredentialKeys lists the keys the credentials Secret must carry --
// what the component preflights before rendering, so a missing key is a
// sentence instead of a pod that cannot start. Nil for a keyless posture.
func (o *OpenBAOSealOptions) RequiredCredentialKeys() []string {
	if o.CredentialsSecretName() == "" {
		return nil
	}
	switch {
	case o.AwsKms != nil:
		return []string{OpenBAOSealEnvAwsSecretAccessKey}
	case o.AzureKeyVault != nil:
		return []string{OpenBAOSealEnvAzureClientSecret}
	case o.Transit != nil:
		return []string{OpenBAOSealEnvTransitToken}
	}
	return nil
}

// SecretEnv is the credential the arm needs, projected from the named Secret
// as the variable of the same name -- rendered through the chart's
// extraSecretEnvironmentVars. Nil for a keyless posture.
func (o *OpenBAOSealOptions) SecretEnv() []OpenBAOSecretEnv {
	name := o.CredentialsSecretName()
	keys := o.RequiredCredentialKeys()
	if name == "" || len(keys) == 0 {
		return nil
	}
	sort.Strings(keys)
	out := make([]OpenBAOSecretEnv, 0, len(keys))
	for _, key := range keys {
		out = append(out, OpenBAOSecretEnv{EnvName: key, SecretName: name, SecretKey: key})
	}
	return out
}

// DecryptGrant names, per arm, the permission the seal identity needs to
// unwrap the vault's key -- the one a vault that stays sealed under a cloud
// seal is missing (the identity could describe the key at start, so the
// server came up; it cannot decrypt with it, so the vault never opens).
func (o *OpenBAOSealOptions) DecryptGrant() string {
	switch o.Word() {
	case OpenBAOSealAwsKms:
		return "kms:Decrypt on the key"
	case OpenBAOSealGcpKms:
		return "roles/cloudkms.cryptoKeyEncrypterDecrypter on the key"
	case OpenBAOSealAzureKeyVault:
		return "the unwrapKey permission on the key"
	case OpenBAOSealTransit:
		return "decrypt on the transit key's path for the token it presents"
	}
	return ""
}

// StartCheck names, per arm, what the seal wrapper does when the server
// configures the seal at start -- and therefore what a crash-looping vault
// under a cloud seal most likely lacks. The GCP arm's second role is the
// case every operator has met once.
func (o *OpenBAOSealOptions) StartCheck() string {
	switch o.Word() {
	case OpenBAOSealAwsKms:
		return "the identity must be able to describe the key (kms:DescribeKey) as well as encrypt and decrypt with it"
	case OpenBAOSealGcpKms:
		return "the identity needs roles/cloudkms.viewer on the key (for the cloudkms.cryptoKeys.get made at start) beside roles/cloudkms.cryptoKeyEncrypterDecrypter"
	case OpenBAOSealAzureKeyVault:
		return "the identity must be able to get the key as well as wrapKey and unwrapKey with it"
	case OpenBAOSealTransit:
		return "the central instance must be reachable and unsealed, the transit engine mounted, and the token able to encrypt on the key"
	}
	return ""
}

func (t *OpenBAOTransitSealOptions) mountPath() string {
	if t.MountPath == "" {
		return "transit/"
	}
	return t.MountPath
}
