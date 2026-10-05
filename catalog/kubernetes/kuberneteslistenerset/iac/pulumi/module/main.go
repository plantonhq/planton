package module

import (
	"github.com/pkg/errors"
	kuberneteslistenersetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteslistenerset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *kuberneteslistenersetv1alpha1.KubernetesListenerSetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createListenerSet(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create listener set")
	}

	ctx.Export(OpListenerSetName, pulumi.String(locals.ListenerSetName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpGatewayName, pulumi.String(locals.GatewayName))

	return nil
}

// createListenerSet creates the namespaced Gateway API ListenerSet using the
// typed crd2pulumi SDK (gatewayv1.NewListenerSet). ListenerSet is a
// standard-channel resource served as gateway.networking.k8s.io/v1. The typed
// approach catches field-name and structure errors at compile time. The
// listener entries are mapped in listeners.go; the parent Gateway reference is
// inline here (group/kind/name/namespace only — a ListenerSet always attaches
// to the Gateway as a whole).
//
// No await customization, deliberately: the per-listener Accepted/Programmed
// conditions appear when a Gateway controller reconciles the resource, which
// is not part of applying it — the same never-block-on-a-controller posture
// as the route kinds.
func createListenerSet(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesListenerSet.Spec

	parentRef := spec.GetParentRef()
	parentRefArgs := gatewayv1.ListenerSetSpecParentRefArgs{
		Name: pulumi.String(locals.GatewayName),
	}
	if group := parentRef.GetGroup(); group != "" {
		parentRefArgs.Group = pulumi.String(group)
	}
	if kind := parentRef.GetKind(); kind != "" {
		parentRefArgs.Kind = pulumi.String(kind)
	}
	if namespace := parentRef.GetNamespace(); namespace != "" {
		parentRefArgs.Namespace = pulumi.String(namespace)
	}

	_, err := gatewayv1.NewListenerSet(ctx, locals.ListenerSetName,
		&gatewayv1.ListenerSetArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.ListenerSetName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewayv1.ListenerSetSpecArgs{
				ParentRef: parentRefArgs,
				Listeners: buildListeners(spec.GetListeners()),
			},
		},
		pulumi.Provider(kubeProvider))

	return err
}
