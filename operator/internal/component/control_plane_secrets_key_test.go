package component

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

func secretsKeyPlatform() *v1.PlantonPlatform {
	p := ownershipPlatform()
	off := false
	p.Spec.Vault = &v1.OpenBAOSpec{Enabled: &off}
	return p
}

func readSecretsKeySecret(t *testing.T, c client.Client, name string) *corev1.Secret {
	t.Helper()
	var s corev1.Secret
	if err := c.Get(context.Background(), types.NamespacedName{Name: name, Namespace: "planton"}, &s); err != nil {
		t.Fatalf("reading Secret %s: %v", name, err)
	}
	return &s
}

// The key is the exact form the control plane decodes -- 32 bytes, standard
// base64, no stray whitespace -- or the control plane refuses to boot.
func TestGenerateSecretsKey_IsWhatTheControlPlaneDecodes(t *testing.T) {
	key, err := resources.GenerateSecretsKey()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		t.Fatalf("standard base64, never base64url: %v", err)
	}
	if len(raw) != 32 {
		t.Errorf("AES-256 needs 32 bytes, got %d", len(raw))
	}
	if strings.TrimSpace(key) != key {
		t.Error("no stray whitespace: the value travels as an env var verbatim")
	}
	other, _ := resources.GenerateSecretsKey()
	if other == key {
		t.Error("two mints must differ")
	}
}

// With nothing named, the operator mints the key into a Secret it owns,
// deleted with the platform, and says so on the Secret.
func TestEnsureSecretsKey_MintsIntoTheOperatorsOwnSecret(t *testing.T) {
	p := secretsKeyPlatform()
	c := vaultFakeClient(t)
	owner := (&Base{}).OwnerReferenceFor(p)

	if err := ensureSecretsKey(context.Background(), c, p, owner); err != nil {
		t.Fatal(err)
	}

	s := readSecretsKeySecret(t, c, "planton-secrets-key")
	if len(s.Data[resources.SecretsKeySecretKey]) == 0 {
		t.Fatal("the key must be minted")
	}
	if len(s.OwnerReferences) != 1 || s.OwnerReferences[0].UID != p.UID {
		t.Errorf("the operator's Secret is owned by the platform, got %+v", s.OwnerReferences)
	}
	if note := s.Annotations[resources.SecretsKeyAnnotation]; !strings.Contains(note, "deleted with the platform") {
		t.Errorf("the note must say who keeps it, got %q", note)
	}
}

// A named Secret is the adopter's: no owner reference, so it outlives the
// platform.
func TestEnsureSecretsKey_AnAdopterNamedSecretOutlivesThePlatform(t *testing.T) {
	p := secretsKeyPlatform()
	p.Spec.ControlPlane = &v1.ControlPlaneSpec{SecretsKeySecretName: "acme-platform-key"}
	c := vaultFakeClient(t)

	if err := ensureSecretsKey(context.Background(), c, p, (&Base{}).OwnerReferenceFor(p)); err != nil {
		t.Fatal(err)
	}

	s := readSecretsKeySecret(t, c, "acme-platform-key")
	if len(s.Data[resources.SecretsKeySecretKey]) == 0 {
		t.Fatal("the key must be minted into the adopter's Secret")
	}
	if len(s.OwnerReferences) != 0 {
		t.Errorf("the adopter's Secret carries no owner reference, got %+v", s.OwnerReferences)
	}
	if note := s.Annotations[resources.SecretsKeyAnnotation]; !strings.Contains(note, "You own this Secret") {
		t.Errorf("the note must say the adopter owns it, got %q", note)
	}
}

// A key already there is never replaced: every stored secret is sealed under it.
func TestEnsureSecretsKey_NeverReplacesAKeyThatIsThere(t *testing.T) {
	p := secretsKeyPlatform()
	held := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "planton-secrets-key", Namespace: "planton"},
		Data:       map[string][]byte{resources.SecretsKeySecretKey: []byte("the-key-everything-is-sealed-under")},
	}
	c := vaultFakeClient(t, held)

	if err := ensureSecretsKey(context.Background(), c, p, (&Base{}).OwnerReferenceFor(p)); err != nil {
		t.Fatal(err)
	}

	if got := string(readSecretsKeySecret(t, c, "planton-secrets-key").Data[resources.SecretsKeySecretKey]); got != "the-key-everything-is-sealed-under" {
		t.Errorf("the held key must survive untouched, got %q", got)
	}
}

// A Secret the adopter created empty gets the key minted into it.
func TestEnsureSecretsKey_FillsAnEmptySecret(t *testing.T) {
	p := secretsKeyPlatform()
	p.Spec.ControlPlane = &v1.ControlPlaneSpec{SecretsKeySecretName: "acme-platform-key"}
	empty := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "acme-platform-key", Namespace: "planton"}}
	c := vaultFakeClient(t, empty)

	if err := ensureSecretsKey(context.Background(), c, p, (&Base{}).OwnerReferenceFor(p)); err != nil {
		t.Fatal(err)
	}

	s := readSecretsKeySecret(t, c, "acme-platform-key")
	if _, err := base64.StdEncoding.DecodeString(string(s.Data[resources.SecretsKeySecretKey])); err != nil || len(s.Data[resources.SecretsKeySecretKey]) == 0 {
		t.Errorf("an empty Secret must get a key, got %q", s.Data[resources.SecretsKeySecretKey])
	}
	if s.Annotations[resources.SecretsKeyAnnotation] == "" {
		t.Error("the filled Secret says what it is")
	}
}

// The binding follows the vault: the key reaches the control plane exactly
// when the vault is off, named after whichever Secret holds it.
func TestBuildConfig_SecretsKeyFollowsTheVault(t *testing.T) {
	cp := &ControlPlane{}
	p := ingressPlatform(false)
	if cfg := cp.buildConfig(p, nil); cfg.SecretsKey != nil || cfg.Vault == nil {
		t.Errorf("with the vault on there is no secrets key, got %+v", cfg.SecretsKey)
	}

	off := false
	p.Spec.Vault = &v1.OpenBAOSpec{Enabled: &off}
	cfg := cp.buildConfig(p, nil)
	if cfg.SecretsKey == nil || cfg.SecretsKey.SecretName != resources.SecretsKeySecretName(p.Name) {
		t.Errorf("with the vault off the operator's own Secret holds the key, got %+v", cfg.SecretsKey)
	}

	p.Spec.ControlPlane = &v1.ControlPlaneSpec{SecretsKeySecretName: "acme-platform-key"}
	if cfg := cp.buildConfig(p, nil); cfg.SecretsKey == nil || cfg.SecretsKey.SecretName != "acme-platform-key" {
		t.Errorf("a named Secret holds the key, got %+v", cfg.SecretsKey)
	}
}
