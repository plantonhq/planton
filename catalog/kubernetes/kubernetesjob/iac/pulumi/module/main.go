package module

import (
	"github.com/pkg/errors"
	kubernetesjobv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesjob/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys a KubernetesJob: an optional namespace, the env and
// image-pull satellite Secrets, and the batch/v1 Job itself.
//
// Identity and configuration are composed, not created: pods run as the
// ServiceAccount referenced in spec.pod.service_account, and config files are
// mounted from ConfigMaps owned by the first-class KubernetesConfigMap kind.
// This module never creates ServiceAccounts, RBAC objects, ConfigMaps,
// certificates, gateways, or routes — Jobs front no traffic.
func Resources(ctx *pulumi.Context, iacInput *kubernetesjobv1alpha1.KubernetesJobIacInput) error {
	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Conditional namespace dependency (Pulumi equivalent of Terraform depends_on):
	// empty when the namespace pre-exists or is owned by a KubernetesNamespace resource.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Satellite secrets are created BEFORE the Job: pods reference them by name
	// at startup, and a pod that starts before its env secret exists crashes —
	// for a Job that failure burns retries from the backoff budget.
	if err := secret(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create env secret")
	}

	createdImagePullSecret, err := imagePullSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create image pull secret")
	}

	if _, err := job(ctx, locals, kubernetesProvider, createdImagePullSecret, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create job")
	}

	return nil
}
