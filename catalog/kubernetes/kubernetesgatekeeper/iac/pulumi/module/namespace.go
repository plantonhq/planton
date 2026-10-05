package module

import (
	"github.com/pkg/errors"
	kubernetesgatekeeperv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgatekeeper/v1alpha1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// namespace conditionally creates the installation namespace based on the
// create_namespace flag. Returns the created namespace resource (or nil
// when create_namespace is false). Terraform equivalent:
// kubernetes_namespace_v1 with count.
//
// The self-management exemption label must be DECLARED here, not left to
// the chart's post-install label_namespace hook: the hook stamps it onto
// a namespace this module owns, so an undeclared label is permanent
// config↔live drift — every later apply would STRIP the exemption and
// let Gatekeeper police its own namespace (fail-closed via the
// label-guard check webhook). Rendered whenever the hook is enabled
// (empty = the chart default, true), matching what the hook itself
// would stamp. Terraform twin: locals.namespace_labels.
func namespace(ctx *pulumi.Context,
	iacInput *kubernetesgatekeeperv1alpha1.KubernetesGatekeeperIacInput,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
) (*kubernetescorev1.Namespace, error) {
	if !iacInput.Target.Spec.CreateNamespace {
		return nil, nil
	}

	namespaceLabels := make(map[string]string, len(locals.Labels)+1)
	for k, v := range locals.Labels {
		namespaceLabels[k] = v
	}
	hooks := iacInput.Target.Spec.GetHooks()
	if hooks == nil || hooks.LabelNamespace == nil || hooks.GetLabelNamespace() {
		namespaceLabels["admission.gatekeeper.sh/ignore"] = "no-self-managing"
	}

	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(
				&kubernetesmeta.ObjectMetaArgs{
					Name:   pulumi.String(locals.Namespace),
					Labels: pulumi.ToStringMap(namespaceLabels),
				}),
		}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	return createdNamespace, nil
}
