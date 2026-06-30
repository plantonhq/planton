package module

import (
	"github.com/pkg/errors"
	kubernetestcproutev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetestcproute/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1alpha2 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1alpha2"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetestcproutev1.KubernetesTcpRouteStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
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
// crd2pulumi SDK (gatewayv1alpha2.NewTCPRoute). TCPRoute is an experimental-channel
// resource served as gateway.networking.k8s.io/v1alpha2 (the experimental CRDs
// must be installed -- see KubernetesGatewayApiCrds install_channel: experimental).
// The typed approach catches field-name and structure errors at compile time. A
// TCP route has no hostnames, matches, or filters; the spec mapping is split
// across parent_refs.go and rules.go.
func createTcpRoute(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesTcpRoute.Spec

	tcpRouteSpec := gatewayv1alpha2.TCPRouteSpecArgs{
		Rules: buildRules(spec.GetRules()),
	}

	if parentRefs := spec.GetParentRefs(); len(parentRefs) > 0 {
		tcpRouteSpec.ParentRefs = buildParentRefs(parentRefs)
	}
	if useDefaultGateways := spec.GetUseDefaultGateways(); useDefaultGateways != "" {
		tcpRouteSpec.UseDefaultGateways = pulumi.String(useDefaultGateways)
	}

	_, err := gatewayv1alpha2.NewTCPRoute(ctx, locals.RouteName,
		&gatewayv1alpha2.TCPRouteArgs{
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
