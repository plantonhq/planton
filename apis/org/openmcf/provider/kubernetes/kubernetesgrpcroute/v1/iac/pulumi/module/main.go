package module

import (
	"github.com/pkg/errors"
	kubernetesgrpcroutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesgrpcroute/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesgrpcroutev1.KubernetesGrpcRouteStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createGrpcRoute(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create grpc route")
	}

	ctx.Export(OpRouteName, pulumi.String(locals.RouteName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createGrpcRoute creates the namespaced Gateway API GRPCRoute using the typed
// crd2pulumi SDK (gatewayv1.NewGRPCRoute), consistent with every other OpenMCF
// ingress component. The typed approach catches field-name and structure errors
// at compile time rather than at deployment time. The GRPCRouteSpec mapping is
// split across parent_refs.go, rules.go, matches.go, filters.go, and
// backend_refs.go.
func createGrpcRoute(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesGrpcRoute.Spec

	grpcRouteSpec := gatewayv1.GRPCRouteSpecArgs{
		Rules: buildRules(spec.GetRules()),
	}

	if parentRefs := spec.GetParentRefs(); len(parentRefs) > 0 {
		grpcRouteSpec.ParentRefs = buildParentRefs(parentRefs)
	}
	if hostnames := spec.GetHostnames(); len(hostnames) > 0 {
		grpcRouteSpec.Hostnames = pulumi.ToStringArray(hostnames)
	}

	_, err := gatewayv1.NewGRPCRoute(ctx, locals.RouteName,
		&gatewayv1.GRPCRouteArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.RouteName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: grpcRouteSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}
