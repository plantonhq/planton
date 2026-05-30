package module

import (
	"github.com/pkg/errors"
	kubernetesreferencegrantv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesreferencegrant/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesreferencegrantv1.KubernetesReferenceGrantStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createReferenceGrant(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create reference grant")
	}

	ctx.Export(OpReferenceGrantName, pulumi.String(locals.ReferenceGrantName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createReferenceGrant creates the namespaced Gateway API ReferenceGrant using
// the typed crd2pulumi SDK (gatewayv1.NewReferenceGrant, served as
// gateway.networking.k8s.io/v1), consistent with every other OpenMCF ingress
// component. The typed approach catches field-name and structure errors at
// compile time rather than at deployment time. The ReferenceGrantSpec mapping
// (the from/to lists) is built in references.go.
func createReferenceGrant(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesReferenceGrant.Spec

	_, err := gatewayv1.NewReferenceGrant(ctx, locals.ReferenceGrantName,
		&gatewayv1.ReferenceGrantArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.ReferenceGrantName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: gatewayv1.ReferenceGrantSpecArgs{
				From: buildFrom(spec.GetFrom()),
				To:   buildTo(spec.GetTo()),
			},
		},
		pulumi.Provider(kubeProvider))

	return err
}
