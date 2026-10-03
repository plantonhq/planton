package module

import (
	"github.com/pkg/errors"
	kubernetestcproutev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestcproute/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *kubernetestcproutev1alpha1.KubernetesTcpRouteIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createTcpRoute(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create tcp route")
	}

	ctx.Export(OpRouteName, pulumi.String(locals.RouteName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createTcpRoute creates the namespaced Gateway API TCPRoute using the typed
// crd2pulumi SDK (gatewayv1.NewTCPRoute). TCPRoute is a standard-channel GA
// resource served as gateway.networking.k8s.io/v1 (it was experimental
// v1alpha2 in earlier Gateway API releases). The typed approach catches
// field-name and structure errors at compile time. A TCP route has no
// hostnames, matches, or filters; the spec mapping is split across
// parent_refs.go and rules.go.
func createTcpRoute(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesTcpRoute.Spec

	tcpRouteSpec := gatewayv1.TCPRouteSpecArgs{
		Rules: buildRules(spec.GetRules()),
	}

	if parentRefs := spec.GetParentRefs(); len(parentRefs) > 0 {
		tcpRouteSpec.ParentRefs = buildParentRefs(parentRefs)
	}

	_, err := gatewayv1.NewTCPRoute(ctx, locals.RouteName,
		&gatewayv1.TCPRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.RouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: tcpRouteSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}
