package component

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// secretsKeySecretName is the Secret that holds the platform's secrets key:
// the adopter's own when the spec names one, the operator's otherwise.
func secretsKeySecretName(planton *v1.PlantonPlatform) string {
	if secretsKeySecretIsAdoptersOwn(planton) {
		return planton.Spec.ControlPlane.SecretsKeySecretName
	}
	return resources.SecretsKeySecretName(planton.Name)
}

// secretsKeySecretIsAdoptersOwn reports whether the secrets key outlives the
// platform: named by the adopter, the operator never deletes it; unnamed, it
// is owner-referenced and dies with the platform.
func secretsKeySecretIsAdoptersOwn(planton *v1.PlantonPlatform) bool {
	return planton.Spec.ControlPlane != nil && planton.Spec.ControlPlane.SecretsKeySecretName != ""
}

// ensureSecretsKey makes sure the platform's secrets key exists before the
// Deployment references it. Create-once: a Secret that already holds a key is
// never touched -- that key is what every stored secret is sealed under. A
// Secret the adopter created empty (or without the key) gets one minted into
// it, with the note that says what it is.
func ensureSecretsKey(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, ownerRef *metav1.OwnerReference) error {
	log := logf.FromContext(ctx)
	name := secretsKeySecretName(planton)
	adoptersOwn := secretsKeySecretIsAdoptersOwn(planton)
	if adoptersOwn {
		// The adopter's Secret outlives the platform: no owner reference.
		ownerRef = nil
	}

	var existing corev1.Secret
	err := c.Get(ctx, types.NamespacedName{Name: name, Namespace: planton.Namespace}, &existing)
	switch {
	case apierrors.IsNotFound(err):
		key, genErr := resources.GenerateSecretsKey()
		if genErr != nil {
			return genErr
		}
		if err := c.Create(ctx, resources.SecretsKeySecret(name, planton.Namespace, key, adoptersOwn, ownerRef)); err != nil {
			return fmt.Errorf("creating Secret %s: %w", name, err)
		}
		log.Info("Minted the platform's secrets key", "secret", name)
		return nil
	case err != nil:
		return fmt.Errorf("getting Secret %s: %w", name, err)
	}

	if len(existing.Data[resources.SecretsKeySecretKey]) == 0 {
		key, genErr := resources.GenerateSecretsKey()
		if genErr != nil {
			return genErr
		}
		patch := client.MergeFrom(existing.DeepCopy())
		if existing.Data == nil {
			existing.Data = map[string][]byte{}
		}
		existing.Data[resources.SecretsKeySecretKey] = []byte(key)
		if existing.Annotations == nil {
			existing.Annotations = map[string]string{}
		}
		existing.Annotations[resources.SecretsKeyAnnotation] = resources.SecretsKeyNote(adoptersOwn)
		if err := c.Patch(ctx, &existing, patch); err != nil {
			return fmt.Errorf("adding the secrets key to Secret %s: %w", name, err)
		}
		log.Info("Minted the platform's secrets key into an existing Secret", "secret", name)
	}
	return nil
}
