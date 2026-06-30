package module

import (
	"github.com/pkg/errors"
	kuberneteshelmreleasev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteshelmrelease/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kuberneteshelmreleasev1.KubernetesHelmReleaseStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------ helm chart ----------------------------
	helmOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	_, err = helmv3.NewChart(ctx,
		locals.KubernetesHelmRelease.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String(locals.KubernetesHelmRelease.Spec.Name),
			Version:   pulumi.String(locals.KubernetesHelmRelease.Spec.Version),
			Namespace: pulumi.String(locals.Namespace),
			Values:    convertstringmaps.ConvertGoStringMapToPulumiMap(locals.KubernetesHelmRelease.Spec.Values),
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String(locals.KubernetesHelmRelease.Spec.Repo),
			},
		}, helmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create helm-chart")
	}
	return nil
}
