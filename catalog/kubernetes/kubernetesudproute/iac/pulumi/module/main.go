package module

import (
	"github.com/pkg/errors"
	kubernetesudproutev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesudproute/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *kubernetesudproutev1alpha1.KubernetesUdpRouteIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createUdpRoute(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create udp route")
	}

	ctx.Export(OpRouteName, pulumi.String(locals.RouteName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createUdpRoute creates the namespaced Gateway API UDPRoute using the typed
// crd2pulumi SDK (gatewayv1.NewUDPRoute). UDPRoute is a standard-channel GA
// resource served as gateway.networking.k8s.io/v1 (it was experimental
// v1alpha2 in earlier Gateway API releases). The typed approach catches
// field-name and structure errors at compile time. A UDP route has no
// hostnames, matches, or filters; the spec mapping is split across
// parent_refs.go and rules.go.
func createUdpRoute(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesUdpRoute.Spec

	udpRouteSpec := gatewayv1.UDPRouteSpecArgs{
		Rules: buildRules(spec.GetRules()),
	}

	if parentRefs := spec.GetParentRefs(); len(parentRefs) > 0 {
		udpRouteSpec.ParentRefs = buildParentRefs(parentRefs)
	}

	_, err := gatewayv1.NewUDPRoute(ctx, locals.RouteName,
		&gatewayv1.UDPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.RouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: udpRouteSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}
