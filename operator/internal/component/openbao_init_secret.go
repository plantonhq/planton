package component

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/bootstrap"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// The init Secret is the one object the operator writes that it cannot
// regenerate: /sys/init hands out the keys and the root token exactly once.
// Every other credential Secret in this operator is create-once because its
// data CAN be regenerated; this one's contract has to be different --
// checked before init so the write cannot collide, written right after
// init with a retry, and never overwritten while it holds keys, because
// keys that already sit in a Secret open some other vault's data and
// replacing them is the one way a later restore could find the wrong keys.
//
// Who owns it is the declaration's choice: spec.vault.initSecretName names
// a Secret the adopter owns (written without an owner reference; the
// operator never deletes it), unset keeps an operator-owned one that dies
// with the platform. Under the built-in seal the Secret holds unseal keys;
// under a cloud seal, recovery keys. Both hold the root token.

// initSecretKeysKey is the data key the vault's keys live under for the
// seal: unseal keys reconstruct the root key, recovery keys authorize the
// break-glass quorum. Deliberately different words.
func initSecretKeysKey(seal *resources.OpenBAOSealOptions) string {
	if seal.Word() == resources.OpenBAOSealShamir {
		return resources.OpenBAOInitSecretUnsealKeysKey
	}
	return resources.OpenBAOInitSecretRecoveryKeysKey
}

// initSecretHoldsKeys reports whether a Secret already carries a vault's
// keys under either seal's key, or a root token -- the test that protects
// them from being overwritten.
func initSecretHoldsKeys(secret *corev1.Secret) bool {
	if secret == nil {
		return false
	}
	for _, key := range []string{
		resources.OpenBAOInitSecretUnsealKeysKey,
		resources.OpenBAOInitSecretRecoveryKeysKey,
		resources.OpenBAOInitSecretRootTokenKey,
	} {
		if len(secret.Data[key]) > 0 {
			return true
		}
	}
	return false
}

// readInitSecret fetches the init Secret; nil, nil when it does not exist.
func readInitSecret(ctx context.Context, c client.Client, name, namespace string) (*corev1.Secret, error) {
	var secret corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading the vault's init Secret %s: %w", name, err)
	}
	return &secret, nil
}

// initSecretWriteAttempts bounds the retry on the write after init. The
// keys exist only in this pass's memory until the write lands; a transient
// API-server failure must not be the reason a vault's keys are lost.
const initSecretWriteAttempts = 5

// writeInitSecret writes the keys and root token /sys/init returned into the
// init Secret, with the note and the seal's fingerprint as annotations. It
// creates the Secret when it is absent and fills it in place when the
// adopter pre-created it empty (a Secret managed through GitOps for its
// existence and RBAC). The adopter's own is written WITHOUT an owner
// reference; the operator's own carries one. Retries briefly on failure; the
// caller turns a final failure into the sentence that says what it costs.
//
// Deliberately a create or a merge patch, never a server-side apply: the
// steady-state pass re-asserts the annotations by apply, and an apply by
// the same field manager that owned the data would drop the data the moment
// it omitted it. The data's manager is this write; the annotations' is the
// apply; neither can remove what the other set.
func (o *OpenBAO) writeInitSecret(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, seal *resources.OpenBAOSealOptions, result *bootstrap.OpenBAOInitResult) error {
	keys := result.UnsealKeys
	if seal.Word() != resources.OpenBAOSealShamir {
		keys = result.RecoveryKeys
	}
	keysJSON, err := json.Marshal(keys)
	if err != nil {
		return fmt.Errorf("marshaling the vault's keys: %w", err)
	}

	name := vaultInitSecretName(planton)
	adoptersOwn := vaultInitSecretIsAdoptersOwn(planton)
	desired := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: planton.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": resources.ManagedByLabel,
			},
			Annotations: initSecretAnnotations(planton, seal),
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			initSecretKeysKey(seal):                 []byte(keysJSON),
			resources.OpenBAOInitSecretRootTokenKey: []byte(result.RootToken),
		},
	}
	if !adoptersOwn {
		if ownerRef := o.OwnerReferenceFor(planton); ownerRef != nil {
			desired.OwnerReferences = []metav1.OwnerReference{*ownerRef}
		}
	}

	var lastErr error
	for attempt := 0; attempt < initSecretWriteAttempts; attempt++ {
		if lastErr = o.writeInitSecretOnce(ctx, c, desired); lastErr == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return lastErr
}

// writeInitSecretOnce is one attempt: create, or fill an existing Secret
// that holds no keys yet. A Secret found holding keys is never written --
// the pre-init check makes that a race, and the race loses to the keys.
func (o *OpenBAO) writeInitSecretOnce(ctx context.Context, c client.Client, desired *corev1.Secret) error {
	existing, err := readInitSecret(ctx, c, desired.Name, desired.Namespace)
	if err != nil {
		return err
	}
	if existing == nil {
		if err := c.Create(ctx, desired.DeepCopy()); err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("creating the vault's init Secret %s: %w", desired.Name, err)
		} else if err == nil {
			return nil
		}
		// Created between our read and our create: fall through and fill it.
		if existing, err = readInitSecret(ctx, c, desired.Name, desired.Namespace); err != nil || existing == nil {
			return fmt.Errorf("re-reading the vault's init Secret %s after a concurrent create: %w", desired.Name, err)
		}
	}
	if initSecretHoldsKeys(existing) {
		return fmt.Errorf("the vault's init Secret %s already holds a vault's keys; refusing to write over them", desired.Name)
	}
	filled := existing.DeepCopy()
	if filled.Labels == nil {
		filled.Labels = map[string]string{}
	}
	for k, v := range desired.Labels {
		filled.Labels[k] = v
	}
	if filled.Annotations == nil {
		filled.Annotations = map[string]string{}
	}
	for k, v := range desired.Annotations {
		filled.Annotations[k] = v
	}
	filled.Data = desired.Data
	if err := c.Patch(ctx, filled, client.MergeFrom(existing)); err != nil {
		return fmt.Errorf("filling the vault's init Secret %s: %w", desired.Name, err)
	}
	return nil
}

// initSecretAnnotations is what the init Secret carries beside its data: the
// plain-language note for the seal and the owner, and the seal's fingerprint
// the declaration is compared with on every later pass.
func initSecretAnnotations(planton *v1.PlantonPlatform, seal *resources.OpenBAOSealOptions) map[string]string {
	return map[string]string{
		resources.OpenBAOInitSecretAnnotation: resources.OpenBAOInitSecretNote(planton.Name, seal, vaultInitSecretIsAdoptersOwn(planton)),
		resources.OpenBAOSealAnnotation:       seal.Fingerprint(),
	}
}

// ensureInitSecretAnnotations re-asserts the note and the fingerprint on the
// steady-state pass: installs initialized before the note existed gain it,
// and a Secret an adopter recreated from its data alone regains its pin --
// an open vault proves the declared seal is the one its storage holds. A
// fingerprint already present is re-asserted with ITS OWN value, never the
// declaration's: the comparison before render is what judges a mismatch,
// and this pass must never paper one over. Annotations only (a server-side
// apply that owns exactly these two keys); the key data is never touched.
func (o *OpenBAO) ensureInitSecretAnnotations(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, seal *resources.OpenBAOSealOptions, existing *corev1.Secret) error {
	annotations := initSecretAnnotations(planton, seal)
	if existing != nil {
		if recorded := existing.Annotations[resources.OpenBAOSealAnnotation]; recorded != "" {
			annotations[resources.OpenBAOSealAnnotation] = recorded
		}
	}
	if err := o.EnsureSecretAnnotations(ctx, c, vaultInitSecretName(planton), planton.Namespace, annotations); err != nil {
		return fmt.Errorf("annotating the vault's init Secret: %w", err)
	}
	return nil
}

// initSecretHoldsForeignKeysMessage is the refusal for an init Secret that
// already carries keys while the vault is not initialized. Those keys open
// some other vault's data (a platform that was reset, a Secret copied from
// another install); writing this vault's keys over them would destroy the
// only way back into that data. Two ways out, one per intent.
func initSecretHoldsForeignKeysMessage(planton *v1.PlantonPlatform, initSecret string) string {
	who := "The operator owns it and would normally have created it"
	if vaultInitSecretIsAdoptersOwn(planton) {
		who = "You named it in spec.vault.initSecretName"
	}
	return fmt.Sprintf(
		"Secret %s in namespace %s already holds a vault's keys (%s, %s, or %s), but this vault is not initialized -- those keys open a different vault's data and the operator will not write over them. %s. "+
			"For a fresh platform, delete or rename that Secret (keep a copy if the other vault's data may still matter); to bring that vault back, declare spec.database.postgresql.recoverFrom instead.",
		initSecret, planton.Namespace,
		resources.OpenBAOInitSecretUnsealKeysKey, resources.OpenBAOInitSecretRecoveryKeysKey, resources.OpenBAOInitSecretRootTokenKey,
		who)
}

// initSecretWriteFailedMessage is the sentence for the rare, serious case:
// /sys/init succeeded and the keys could not be saved. The vault's storage
// now holds an initialized barrier whose keys nobody has; it cannot be
// opened and cannot be initialized again until that storage is reset.
func initSecretWriteFailedMessage(seal *resources.OpenBAOSealOptions, initSecret string, err error) string {
	consequence := "the vault cannot be opened and cannot be initialized again"
	if seal.Word() != resources.OpenBAOSealShamir {
		consequence = "the vault still opens itself from your key, but its root token and recovery keys (the break-glass) are gone and it cannot be initialized again"
	}
	return fmt.Sprintf(
		"The vault was initialized but its keys could not be written to Secret %s (%v) and are lost: %s. "+
			"Reset the vault's storage before the operator can initialize it again: with the platform's vault pod scaled to zero, drop and recreate the database %q on the platform's PostgreSQL cluster; the operator reconciles the rest.",
		initSecret, err, consequence, resources.DBOpenBAO)
}

// openVaultWithoutSecretMessage is the Ready sentence for a vault that is
// open and working while its init Secret does not exist: a restored vault
// under a cloud seal with no kept Secret, a renamed spec.vault.initSecretName,
// an operator Secret deleted from under a running platform. Nothing running
// needs that Secret -- the operator signs in as itself and the control plane
// with the token the operator mints -- so the platform is not refused; but
// what the Secret held is gone until it is back, and under the built-in seal
// that is the only way to open the vault after its next restart.
func openVaultWithoutSecretMessage(planton *v1.PlantonPlatform, seal *resources.OpenBAOSealOptions, initSecret string) string {
	origin := "OpenBAO healthy, but Secret " + initSecret + " does not exist in namespace " + planton.Namespace
	if planton.Status.Backup != nil && planton.Status.Backup.RestoredFrom != "" {
		origin = fmt.Sprintf("OpenBAO healthy -- the vault came back from archive %s and %s opened it -- but Secret %s does not exist in namespace %s",
			planton.Status.Backup.RestoredFrom, seal.Human(), initSecret, planton.Namespace)
	}
	consequence := "the platform works, and the vault's break-glass (its root token and recovery keys) is gone until the Secret is back"
	if seal.Word() == resources.OpenBAOSealShamir {
		consequence = "the platform works until the vault next restarts, and then nothing can unseal it: that Secret's unseal keys are the only way"
	}
	return fmt.Sprintf(
		"%s: %s. "+
			"Recreate it from the copy you kept; if you renamed spec.vault.initSecretName, copy the Secret to the new name.",
		origin, consequence)
}
