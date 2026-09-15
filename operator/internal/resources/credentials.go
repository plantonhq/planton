package resources

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// credentialPasswordLength is the number of random bytes generated for
	// passwords. 24 bytes -> 32-character URL-safe base64 string.
	credentialPasswordLength = 24

	// ManagedByLabel is the value used for the app.kubernetes.io/managed-by
	// label on all resources created by the operator.
	ManagedByLabel = "planton-operator"
)

// GeneratePassword returns a cryptographically random, URL-safe base64 string
// suitable for use as a database or service password. The returned string is
// 32 characters long (24 random bytes, base64-encoded without padding).
// URL-safe encoding avoids shell-escaping issues in environment variables.
func GeneratePassword() (string, error) {
	b := make([]byte, credentialPasswordLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random password: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// NewBasicAuthSecret builds a kubernetes.io/basic-auth Secret (username and
// password under the keys that type defines) with the standard managed-by
// label. The shape CloudNativePG reads a managed role's password from, and
// the shape its own generated credentials carry.
func NewBasicAuthSecret(name, namespace, username, password string, ownerRef *metav1.OwnerReference) *corev1.Secret {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": ManagedByLabel,
			},
		},
		Type: corev1.SecretTypeBasicAuth,
		Data: map[string][]byte{
			BasicAuthUsernameKey: []byte(username),
			BasicAuthPasswordKey: []byte(password),
		},
	}
	if ownerRef != nil {
		secret.OwnerReferences = []metav1.OwnerReference{*ownerRef}
	}
	return secret
}

// NewCredentialSecret builds a Kubernetes Secret containing a single key-value
// credential pair. The Secret uses Opaque type and includes the standard
// managed-by label for operator-created resources.
func NewCredentialSecret(name, namespace, dataKey, password string, ownerRef *metav1.OwnerReference) *corev1.Secret {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": ManagedByLabel,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			dataKey: []byte(password),
		},
	}

	if ownerRef != nil {
		secret.OwnerReferences = []metav1.OwnerReference{*ownerRef}
	}

	return secret
}
