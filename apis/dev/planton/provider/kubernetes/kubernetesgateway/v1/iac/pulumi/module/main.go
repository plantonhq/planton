package module

import (
	"github.com/pkg/errors"
	kubernetesgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgateway/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesgatewayv1.KubernetesGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createGateway(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create gateway")
	}

	ctx.Export(OpGatewayName, pulumi.String(locals.GatewayName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpGatewayClassName, pulumi.String(locals.GatewayClassName))

	return nil
}

// createGateway creates the namespaced Gateway API Gateway using the typed
// crd2pulumi SDK (gatewayv1.NewGateway), consistent with every other Planton
// ingress component. The typed approach catches field-name and structure errors
// at compile time rather than at deployment time. The upstream GatewaySpec is
// large, so its mapping is split across listeners.go, tls.go, addresses.go,
// infrastructure.go, and selectors.go.
func createGateway(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesGateway.Spec

	gatewaySpec := gatewayv1.GatewaySpecArgs{
		GatewayClassName: pulumi.String(locals.GatewayClassName),
		Listeners:        buildListeners(spec.GetListeners()),
	}

	if addresses := spec.GetAddresses(); len(addresses) > 0 {
		gatewaySpec.Addresses = buildAddresses(addresses)
	}
	if infra := spec.GetInfrastructure(); infra != nil {
		gatewaySpec.Infrastructure = buildInfrastructure(infra)
	}
	if allowedListeners := spec.GetAllowedListeners(); allowedListeners != nil {
		gatewaySpec.AllowedListeners = buildAllowedListeners(allowedListeners)
	}
	if tls := spec.GetTls(); tls != nil {
		gatewaySpec.Tls = buildGatewayTls(tls)
	}

	_, err := gatewayv1.NewGateway(ctx, locals.GatewayName,
		&gatewayv1.GatewayArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.GatewayName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewaySpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}
