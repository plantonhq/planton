package resources

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The platform's secrets key: the key-encryption-key the control plane's
// built-in local secret backend seals every value under when the bundled
// vault is off. Values live envelope-encrypted in the platform's own
// database; this key never does -- it lives in a Secret and reaches the
// control plane as PLANTON_LOCAL_SECRETS_KEK, read once at start.
const (
	// SecretsKeySecretKey is the data key the secrets key is stored under.
	SecretsKeySecretKey = "kek"

	// SecretsKeyAnnotation marks the secrets key's Secret as self-describing:
	// whoever finds it must not have to guess what it seals or what losing it
	// costs.
	SecretsKeyAnnotation = "planton.ai/secrets-key"

	// secretsKeyBytes is AES-256: the control plane refuses any other length.
	secretsKeyBytes = 32
)

// SecretsKeySecretName returns the operator-owned secrets key Secret's name,
// "{crName}-secrets-key" -- where the key goes when the declaration names no
// Secret of its own.
func SecretsKeySecretName(crName string) string {
	return fmt.Sprintf("%s-secrets-key", crName)
}

// GenerateSecretsKey returns 32 random bytes in STANDARD base64 with padding
// -- the exact form the control plane decodes (java.util.Base64's basic
// decoder refuses the URL-safe alphabet GeneratePassword uses).
func GenerateSecretsKey() (string, error) {
	b := make([]byte, secretsKeyBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating the platform's secrets key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// SecretsKeyNote renders the annotation's plain-language explanation: what
// the key seals, what losing it costs, and who keeps the Secret.
func SecretsKeyNote(adoptersOwn bool) string {
	var b strings.Builder
	b.WriteString("The platform's secrets key. With the bundled vault off, the control plane keeps every secret in the platform's database, ")
	b.WriteString("envelope-encrypted under this key, and the key never enters the database. ")
	b.WriteString("Without it, no stored secret can be read -- not by the platform, not after a restore. ")
	if adoptersOwn {
		b.WriteString("You own this Secret (spec.controlPlane.secretsKeySecretName): the operator wrote the key into it once and never deletes it, and deleting the PlantonPlatform leaves it standing. ")
		b.WriteString("Keep a copy of it outside the cluster -- it is what opens a restored database's secrets.")
	} else {
		b.WriteString("The operator owns this Secret and it is deleted with the platform. ")
		b.WriteString("Teams that back up the platform name a Secret they own in spec.controlPlane.secretsKeySecretName and keep a copy of it outside the cluster.")
	}
	return b.String()
}

// SecretsKeySecret builds the Secret holding a freshly minted secrets key.
// An adopter-named Secret carries no owner reference (nil ownerRef): it
// outlives the platform.
func SecretsKeySecret(name, namespace, key string, adoptersOwn bool, ownerRef *metav1.OwnerReference) *corev1.Secret {
	secret := NewCredentialSecret(name, namespace, SecretsKeySecretKey, key, ownerRef)
	secret.Annotations = map[string]string{SecretsKeyAnnotation: SecretsKeyNote(adoptersOwn)}
	return secret
}

// SecretsKeyBinding hands the platform's secrets key to the control plane.
// Set only when the bundled vault is off: the key is what makes the built-in
// local backend servable, and the default secret backend the control plane
// seeds follows from it.
type SecretsKeyBinding struct {
	// SecretName is the Secret holding the key under SecretsKeySecretKey.
	SecretName string
}

// secretsKeyEnvVars renders the key by Secret reference, and points the
// backend-credential store (pasted cloud keys, vault tokens of a customer's
// own vault) at the same key instead of the absent platform vault.
func secretsKeyEnvVars(binding *SecretsKeyBinding) []corev1.EnvVar {
	if binding == nil {
		return nil
	}
	return []corev1.EnvVar{
		// ── the platform's secrets key (local backend + credential store) ──
		secretEnv("PLANTON_LOCAL_SECRETS_KEK", binding.SecretName, SecretsKeySecretKey),
		{Name: "PLANTON_CREDENTIALS_PROVIDER", Value: "local"},
	}
}
