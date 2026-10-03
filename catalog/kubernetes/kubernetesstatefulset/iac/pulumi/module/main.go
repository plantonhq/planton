package module

import (
	"github.com/pkg/errors"
	kubernetesstatefulsetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstatefulset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys a KubernetesStatefulSet: an optional namespace, the env and
// image-pull satellite Secrets, the headless governing Service, the StatefulSet
// itself, and an optional PDB.
//
// Identity and exposure are composed, not created: pods run as the
// ServiceAccount referenced in spec.pod.service_account, and external exposure
// attaches through first-class ingress kinds referencing this workload's
// exported Service handle. This module never creates ServiceAccounts, RBAC
// objects, certificates, gateways, or routes.
func Resources(ctx *pulumi.Context, iacInput *kubernetesstatefulsetv1alpha1.KubernetesStatefulSetIacInput) error {
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

	// Satellite secrets are created BEFORE the StatefulSet: pods reference them by
	// name at startup, and a pod that starts before its env secret exists crashes.
	if err := secret(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create env secret")
	}

	createdImagePullSecret, err := imagePullSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create image pull secret")
	}

	// The governing Service is created BEFORE the StatefulSet: the Kubernetes API
	// requires spec.serviceName to reference an existing Service, and pods resolve
	// peer DNS through it during bootstrap — creating it after the StatefulSet
	// would leave early pods unable to find their peers.
	createdService, err := service(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create headless service")
	}

	if err := statefulSet(ctx, locals, kubernetesProvider, createdImagePullSecret, createdService, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create stateful set")
	}

	if err := podDisruptionBudget(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create pod disruption budget")
	}

	return nil
}
