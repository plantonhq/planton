package module

import (
	"github.com/pkg/errors"
	kubernetesdeploymentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesdeployment/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesdeploymentv1.KubernetesDeploymentStackInput) error {
	locals, err := initializeLocals(ctx, stackInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	//create kubernetes-provider from the credential in the stack-input
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

	//create ConfigMaps from spec before deployment
	_, err = configMaps(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create configmaps")
	}

	//create kubernetes deployment resources
	createdDeployment, err := deployment(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create microservice deployment")
	}

	//create kubernetes service resources
	if err := service(ctx, locals, kubernetesProvider, createdDeployment, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create microservice kubernetes service resource")
	}

	//create kubernetes secret with app secrets
	if err := secret(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}

	//create istio-ingress resources if ingress is enabled.
	if locals.KubernetesDeployment.Spec.Ingress != nil && locals.KubernetesDeployment.Spec.Ingress.Enabled {
		if err := ingress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
			return errors.Wrap(err, "failed to create istio ingress resources")
		}
	}

	//create pod disruption budget if enabled
	if err := podDisruptionBudget(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create pod disruption budget")
	}

	return nil
}
