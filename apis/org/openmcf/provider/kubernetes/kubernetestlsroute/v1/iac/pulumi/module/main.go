package module

import (
	"github.com/pkg/errors"
	kubernetestlsroutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetestlsroute/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetestlsroutev1.KubernetesTlsRouteStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createTlsRoute(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create tls route")
	}

	ctx.Export(OpRouteName, pulumi.String(locals.RouteName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createTlsRoute creates the namespaced Gateway API TLSRoute using the typed
// crd2pulumi SDK (gatewayv1.NewTLSRoute, served as gateway.networking.k8s.io/v1),
// consistent with every other OpenMCF ingress component. The typed approach
// catches field-name and structure errors at compile time rather than at
// deployment time. The TLSRouteSpec mapping is split across parent_refs.go and
// rules.go (a TLS route has no matches or filters).
func createTlsRoute(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesTlsRoute.Spec

	tlsRouteSpec := gatewayv1.TLSRouteSpecArgs{
		Hostnames: pulumi.ToStringArray(spec.GetHostnames()),
		Rules:     buildRules(spec.GetRules()),
	}

	if parentRefs := spec.GetParentRefs(); len(parentRefs) > 0 {
		tlsRouteSpec.ParentRefs = buildParentRefs(parentRefs)
	}

	_, err := gatewayv1.NewTLSRoute(ctx, locals.RouteName,
		&gatewayv1.TLSRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.RouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: tlsRouteSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}
