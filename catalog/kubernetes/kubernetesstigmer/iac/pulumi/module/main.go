package module

import (
	"github.com/pkg/errors"
	kubernetesstigmerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstigmer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources installs Stigmer: the operator chart, the Secrets the
// declaration gives values for, the restore check, the cleanup marker, and
// the one Stigmer resource, in that order. A destroy runs it backwards: the
// Stigmer goes first, the cleanup marker's hook waits for the operator to
// remove what it installed, and only then do the Secrets and the chart go.
// Terraform twin: iac/tf, the same documents (parity_test.go holds them
// together).
//
// The Stigmer is rendered UNTYPED (apiextensions.NewCustomResource with a
// plain map, render.go) from only the keys the spec declares, so the
// operator's own defaults stay authoritative. Its definition arrives inside
// the chart's release, and both engines plan it on a cluster that has never
// had the definition: Pulumi previews a custom resource of an unknown kind
// without asking the cluster, and the chart release it depends on installs
// the definition before the Stigmer is created.
func Resources(ctx *pulumi.Context, iacInput *kubernetesstigmerv1alpha1.KubernetesStigmerIacInput) error {
	locals := initializeLocals(iacInput)
	if err := preflight(iacInput.ProviderConfig, locals); err != nil {
		return err
	}

	provider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// The operator's namespace, when this resource owns it. Its labels carry
	// planton.ai/kind, the ownership mark the preflight reads.
	var operatorNamespace []pulumi.Resource
	if locals.Spec.GetOperator().GetCreateNamespace() {
		ns, err := kubernetescorev1.NewNamespace(ctx, locals.OperatorNamespace, &kubernetescorev1.NamespaceArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.OperatorNamespace),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to create the operator's namespace %s", locals.OperatorNamespace)
		}
		operatorNamespace = append(operatorNamespace, ns)
	}

	// The operator. The chart owns the Stigmer definition as a template it
	// keeps on uninstall (helm.sh/resource-policy: keep), so the module adds
	// no CRD handling of its own. The release waits for the operator to be
	// Available: an operator that stops at startup (beside another copy, say)
	// fails this deploy, instead of a Stigmer that never reconciles.
	release, err := helmv3.NewRelease(ctx, vars.OperatorRelease, &helmv3.ReleaseArgs{
		Name:            pulumi.String(vars.OperatorRelease),
		Namespace:       pulumi.String(locals.OperatorNamespace),
		Chart:           pulumi.String(vars.ChartRepository + "/" + vars.ChartName),
		Version:         pulumi.String(locals.Spec.GetVersion()),
		Values:          pulumi.ToMap(helmValues(locals.Spec)),
		CreateNamespace: pulumi.Bool(false),
	}, pulumi.Provider(provider), pulumi.DependsOn(operatorNamespace))
	if err != nil {
		return errors.Wrap(err, "failed to install the stigmer-operator chart")
	}

	// The Secrets the declaration gives values for.
	secrets := []pulumi.Resource{}
	for _, s := range materializedSecrets(locals.Spec) {
		secret, err := kubernetescorev1.NewSecret(ctx, s.Name, &kubernetescorev1.SecretArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:      pulumi.String(s.Name),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Type:       pulumi.String("Opaque"),
			StringData: pulumi.ToSecret(pulumi.ToStringMap(s.Data)).(pulumi.StringMapOutput),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to write Secret %s/%s", locals.Namespace, s.Name)
		}
		secrets = append(secrets, secret)
	}

	gate, err := keysGate(ctx, locals, provider, append([]pulumi.Resource{release}, secrets...))
	if err != nil {
		return err
	}
	marker, err := cleanupMarker(ctx, locals, provider, append(append([]pulumi.Resource{release}, secrets...), operatorNamespace...))
	if err != nil {
		return err
	}

	metadata := &kubernetesmetav1.ObjectMetaArgs{
		Name:      pulumi.String(vars.StigmerName),
		Namespace: pulumi.String(locals.Namespace),
		Labels:    pulumi.ToStringMap(locals.Labels),
	}
	if annotations := stigmerAnnotations(locals.Spec); annotations != nil {
		metadata.Annotations = pulumi.ToStringMap(annotations)
	}
	_, err = apiextensions.NewCustomResource(ctx, vars.StigmerName, &apiextensions.CustomResourceArgs{
		ApiVersion:  pulumi.String(vars.StigmerAPIVersion),
		Kind:        pulumi.String(vars.StigmerKind),
		Metadata:    metadata,
		OtherFields: map[string]interface{}{"spec": stigmerSpec(locals.Spec)},
	}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{gate, marker}))
	if err != nil {
		return errors.Wrap(err, "failed to declare the Stigmer")
	}

	for name, value := range outputs(locals) {
		ctx.Export(name, pulumi.String(value))
	}
	return nil
}
