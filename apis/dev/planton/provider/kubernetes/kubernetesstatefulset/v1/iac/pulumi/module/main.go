package module

import (
	"github.com/pkg/errors"
	kubernetesstatefulsetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesstatefulset/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesstatefulsetv1.KubernetesStatefulSetStackInput) error {
	locals, err := initializeLocals(ctx, stackInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// Create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	// When create_namespace is false, createdNamespace is nil and namespaceDeps is empty.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Create ConfigMaps from spec before StatefulSet
	_, err = configMaps(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create configmaps")
	}

	// Create the headless service for stable network identity (required for StatefulSet)
	createdHeadlessService, err := headlessService(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create headless service")
	}

	// Create the StatefulSet
	createdStatefulSet, err := statefulSet(ctx, locals, kubernetesProvider, createdHeadlessService, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create stateful set")
	}

	// Create ClusterIP service for client access (if ports are defined)
	if err := clientService(ctx, locals, kubernetesProvider, createdStatefulSet, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create client service")
	}

	// Create kubernetes secret with app secrets
	if err := secret(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}

	// Create istio-ingress resources if ingress is enabled
	if locals.KubernetesStatefulSet.Spec.Ingress != nil && locals.KubernetesStatefulSet.Spec.Ingress.Enabled {
		if err := ingress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
			return errors.Wrap(err, "failed to create istio ingress resources")
		}
	}

	// Create pod disruption budget if enabled
	if err := podDisruptionBudget(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create pod disruption budget")
	}

	return nil
}
