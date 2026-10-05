package module

import (
	"github.com/pkg/errors"
	kubernetesdeploymentv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesdeployment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys a KubernetesDeployment: an optional namespace, the env and
// image-pull satellite Secrets, the Deployment itself, its fronting Service,
// and optional HPA and PDB.
//
// Identity and exposure are composed, not created: pods run as the
// ServiceAccount referenced in spec.pod.service_account, and external exposure
// attaches through first-class ingress kinds referencing this workload's
// exported Service handle. This module never creates ServiceAccounts, RBAC
// objects, certificates, gateways, or routes.
func Resources(ctx *pulumi.Context, iacInput *kubernetesdeploymentv1alpha1.KubernetesDeploymentIacInput) error {
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

	// Satellite secrets are created BEFORE the Deployment: pods reference them by
	// name at startup, and a pod that starts before its env secret exists crashes.
	if err := secret(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create env secret")
	}

	createdImagePullSecret, err := imagePullSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create image pull secret")
	}

	createdDeployment, err := deployment(ctx, locals, kubernetesProvider, createdImagePullSecret, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}

	if err := service(ctx, locals, kubernetesProvider, createdDeployment, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create service")
	}

	if err := horizontalPodAutoscaler(ctx, locals, kubernetesProvider, createdDeployment); err != nil {
		return errors.Wrap(err, "failed to create horizontal pod autoscaler")
	}

	if err := podDisruptionBudget(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create pod disruption budget")
	}

	return nil
}
