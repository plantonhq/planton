package module

import (
	"github.com/pkg/errors"
	kubernetesgatewayclassv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgatewayclass/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesgatewayclassv1.KubernetesGatewayClassStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createGatewayClass(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create gateway class")
	}

	ctx.Export(OpGatewayClassName, pulumi.String(locals.GatewayClassName))
	ctx.Export(OpControllerName, pulumi.String(locals.ControllerName))

	return nil
}

// createGatewayClass creates the cluster-scoped Gateway API GatewayClass using
// the typed crd2pulumi SDK (gatewayv1.NewGatewayClass), consistent with how all
// other Planton ingress components consume the Gateway API typed resources. The
// typed approach catches field name and structure errors at compile time rather
// than at deployment time.
func createGatewayClass(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesGatewayClass.Spec

	gatewayClassSpec := gatewayv1.GatewayClassSpecArgs{
		ControllerName: pulumi.String(spec.ControllerName),
	}

	// description is optional upstream; only set it when provided.
	if description := spec.GetDescription(); description != "" {
		gatewayClassSpec.Description = pulumi.String(description)
	}

	// parameters_ref is optional; map the structured reference when present.
	if paramsRef := spec.GetParametersRef(); paramsRef != nil {
		parametersRefArgs := gatewayv1.GatewayClassSpecParametersRefArgs{
			Group: pulumi.String(paramsRef.GetGroup()),
			Kind:  pulumi.String(paramsRef.GetKind()),
			Name:  pulumi.String(paramsRef.GetName()),
		}
		// namespace must be set only for namespace-scoped parameters resources.
		if namespace := paramsRef.GetNamespace(); namespace != "" {
			parametersRefArgs.Namespace = pulumi.String(namespace)
		}
		gatewayClassSpec.ParametersRef = parametersRefArgs
	}

	_, err := gatewayv1.NewGatewayClass(ctx, locals.GatewayClassName,
		&gatewayv1.GatewayClassArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.GatewayClassName),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewayClassSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}
