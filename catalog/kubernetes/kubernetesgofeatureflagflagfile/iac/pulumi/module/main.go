package module

import (
	"github.com/pkg/errors"
	kubernetesgofeatureflagflagfilev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgofeatureflagflagfile/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources renders the flag file into one ConfigMap named after the
// resource, under the spec's data key. A KubernetesGoFeatureFlag relay's
// `config_map` retriever reads it through the Kubernetes API on every poll,
// so an edit reaches evaluations within one polling interval and nothing
// restarts. Terraform twin: kubernetes_config_map_v1.flag_file.
func Resources(ctx *pulumi.Context, iacInput *kubernetesgofeatureflagflagfilev1alpha1.KubernetesGoFeatureFlagFlagFileIacInput) error {
	locals := initializeLocals(iacInput)

	document, err := renderFlagFile(locals.Spec)
	if err != nil {
		return err
	}

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	_, err = kubernetescorev1.NewConfigMap(ctx, locals.Name, &kubernetescorev1.ConfigMapArgs{
		Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
			Name:      pulumi.String(locals.Name),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		}),
		Data: pulumi.StringMap{locals.Key: pulumi.String(document)},
	}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create %s config map", locals.Name)
	}

	ctx.Export(OpConfigMapName, pulumi.String(locals.Name))
	ctx.Export(OpKey, pulumi.String(locals.Key))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	return nil
}
