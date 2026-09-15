package component

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// The seal is what opens the vault, and it is decided when the vault is
// created. This file is the component's half of that: mapping the
// declaration onto the rendering options, checking before the chart renders
// that the declared credentials exist and that the declaration still
// describes the seal the vault was initialized under, and the sentences a
// person reads when a cloud seal has not opened the vault -- because the
// server tells them apart only in its log, and the platform's status must
// say it in words.

// sealOptionsFrom maps the definition's seal onto the rendering options.
// Nil (the built-in key shares) when the declaration names none.
func sealOptionsFrom(planton *v1.PlantonPlatform) *resources.OpenBAOSealOptions {
	if planton.Spec.Vault == nil || planton.Spec.Vault.AutoUnseal == nil {
		return nil
	}
	seal := planton.Spec.Vault.AutoUnseal
	switch {
	case seal.AwsKms != nil:
		return &resources.OpenBAOSealOptions{AwsKms: &resources.OpenBAOAwsKmsSealOptions{
			Region:                seal.AwsKms.Region,
			KmsKeyID:              seal.AwsKms.KmsKeyID,
			AccessKeyID:           seal.AwsKms.AccessKeyID,
			CredentialsSecretName: seal.AwsKms.CredentialsSecretName,
		}}
	case seal.GcpKms != nil:
		return &resources.OpenBAOSealOptions{GcpKms: &resources.OpenBAOGcpKmsSealOptions{
			Project:   seal.GcpKms.Project,
			Region:    seal.GcpKms.Region,
			KeyRing:   seal.GcpKms.KeyRing,
			CryptoKey: seal.GcpKms.CryptoKey,
		}}
	case seal.AzureKeyVault != nil:
		return &resources.OpenBAOSealOptions{AzureKeyVault: &resources.OpenBAOAzureKeyVaultSealOptions{
			VaultName:             seal.AzureKeyVault.VaultName,
			KeyName:               seal.AzureKeyVault.KeyName,
			TenantID:              seal.AzureKeyVault.TenantID,
			ClientID:              seal.AzureKeyVault.ClientID,
			CredentialsSecretName: seal.AzureKeyVault.CredentialsSecretName,
		}}
	case seal.Transit != nil:
		return &resources.OpenBAOSealOptions{Transit: &resources.OpenBAOTransitSealOptions{
			Address:               seal.Transit.Address,
			KeyName:               seal.Transit.KeyName,
			MountPath:             seal.Transit.MountPath,
			CredentialsSecretName: seal.Transit.CredentialsSecretName,
		}}
	}
	return nil
}

// vaultServiceAccountAnnotations is the merged identity map the declaring
// module put on the CR -- rendered as given, nil when empty.
func vaultServiceAccountAnnotations(planton *v1.PlantonPlatform) map[string]string {
	if planton.Spec.Vault == nil || len(planton.Spec.Vault.ServiceAccountAnnotations) == 0 {
		return nil
	}
	return planton.Spec.Vault.ServiceAccountAnnotations
}

// preflightSealCredentialsSecret checks the adopter-owned credentials Secret
// a seal arm names exists and carries the key its wrapper reads. A missing
// Secret or key is a sentence, not an error: the chart is not rendered and
// the status says exactly what to add -- otherwise the pod would sit in
// CreateContainerConfigError naming a mechanism. Keyless arms need nothing.
func (o *OpenBAO) preflightSealCredentialsSecret(ctx context.Context, c client.Client, namespace string, seal *resources.OpenBAOSealOptions) (string, error) {
	name := seal.CredentialsSecretName()
	if name == "" {
		return "", nil
	}
	keys := seal.RequiredCredentialKeys()
	field := "spec.vault.autoUnseal." + seal.Word() + ".credentialsSecretName"
	var secret corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, &secret)
	if apierrors.IsNotFound(err) {
		return fmt.Sprintf(
			"credentials Secret %q named by %s not found in namespace %q; create it with keys %v, or leave credentialsSecretName empty to use the vault's own cloud identity -- until then the vault is not started",
			name, field, namespace, keys), nil
	}
	if err != nil {
		return "", fmt.Errorf("checking the seal credentials Secret %s: %w", name, err)
	}
	for _, key := range keys {
		if len(secret.Data[key]) == 0 {
			return fmt.Sprintf(
				"credentials Secret %q named by %s has no %q key; add it (the seal needs %v) -- until then the vault is not started",
				name, field, key, keys), nil
		}
	}
	return "", nil
}

// sealChangedMessage is the refusal for a declaration whose seal is not the
// one the vault was initialized under. The server would refuse to start
// against the old storage ("cannot seal migrate from X to Y"), and there is
// no migration path in this product: the seal is decided when the vault is
// created.
func sealChangedMessage(recorded, declared string, initSecret string) string {
	return fmt.Sprintf(
		"The vault was initialized with the seal %q (recorded on Secret %s) and the declaration now says %q. "+
			"The seal is decided when the vault is created and cannot be changed on a running platform: restore the declaration to the recorded seal, "+
			"or declare a new platform (a restore from this platform's archive keeps the recorded seal).",
		recorded, initSecret, declared)
}

// sealOpeningWindow is how long after the vault's pod started a sealed vault
// under a cloud seal is still "opening": the server retries the stored key
// every five seconds, so a pod this young may simply not have come round
// yet. Past it, a vault that is still sealed has a permission problem.
const sealOpeningWindow = time.Minute

// sealNotOpenedMessage is the sentence for a vault that is initialized and
// sealed under a cloud seal, where the operator never unseals anything.
// Under a minute since the pod started it is opening; after that the seal
// identity could describe the key (the server came up) but cannot decrypt
// with it (the vault never opens) -- and the grant is named per arm.
func sealNotOpenedMessage(seal *resources.OpenBAOSealOptions, podAge time.Duration) string {
	if podAge < sealOpeningWindow {
		return fmt.Sprintf("The vault is initialized and %s is opening it; the server retries the key every five seconds.", seal.Human())
	}
	return fmt.Sprintf(
		"The vault is initialized but %s has not opened it after %s: the seal identity can read the key (the server started) but cannot decrypt with it. "+
			"Grant it %s; the server retries every five seconds and the vault opens on its own once the grant lands.",
		seal.Human(), podAge.Round(time.Second), seal.DecryptGrant())
}

// sealStartHint is appended to a crash-loop explanation when a cloud seal is
// declared: the server configures the seal before anything else and exits
// when the wrapper's first call fails, so a vault that keeps exiting under a
// seal is, first of all, a seal check that did not pass.
func sealStartHint(seal *resources.OpenBAOSealOptions) string {
	return fmt.Sprintf(
		"Under %s the usual cause is the seal check the server makes at start: %s. The log line to look for is \"Error configuring seal\".",
		seal.Human(), seal.StartCheck())
}

// vaultPodAge reports how long the vault's server pod has been running --
// what tells a seal that is still opening from a seal that cannot decrypt.
// Zero when the pod cannot be read.
func (o *OpenBAO) vaultPodAge(ctx context.Context, c client.Client, releaseName, namespace string, now time.Time) time.Duration {
	var pods corev1.PodList
	if err := c.List(ctx, &pods, client.InNamespace(namespace), client.MatchingLabels{"app.kubernetes.io/name": releaseName}); err != nil {
		return 0
	}
	for i := range pods.Items {
		if pods.Items[i].Status.Phase == corev1.PodRunning && pods.Items[i].Status.StartTime != nil {
			return now.Sub(pods.Items[i].Status.StartTime.Time)
		}
	}
	return 0
}
