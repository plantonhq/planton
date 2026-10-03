package module

import (
	"github.com/pkg/errors"
	kubernetesjenkinsv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesjenkins/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *kubernetesjenkinsv1alpha1.KubernetesJenkinsIacInput) error {
	locals := initializeLocals(ctx, iacInput)
	//create kubernetes-provider from the credential in the iac-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	//conditionally create namespace resource based on create_namespace flag
	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	//export name of the namespace
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	//create admin-password secret
	createdAdminPasswordSecret, err := adminCredentials(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create admin password resources")
	}

	//install the jenkins helm-chart
	if err := helmChart(ctx, locals, kubernetesProvider, createdAdminPasswordSecret, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create helm-chart resources")
	}

	//create istio-ingress resources if ingress is enabled.
	if locals.KubernetesJenkins.Spec.Ingress != nil && locals.KubernetesJenkins.Spec.Ingress.Enabled {
		if err := ingress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
			return errors.Wrap(err, "failed to create ingress resources")
		}
	}

	return nil
}
