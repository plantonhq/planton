package module

import (
	"github.com/pkg/errors"
	kubernetesbackendtlspolicyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesbackendtlspolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *kubernetesbackendtlspolicyv1alpha1.KubernetesBackendTlsPolicyIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createBackendTlsPolicy(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create backend tls policy")
	}

	ctx.Export(OpPolicyName, pulumi.String(locals.PolicyName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createBackendTlsPolicy creates the namespaced Gateway API BackendTLSPolicy
// using the typed crd2pulumi SDK (gatewayv1.NewBackendTLSPolicy).
// BackendTLSPolicy is served as gateway.networking.k8s.io/v1 (the v1alpha3
// version is deprecated upstream and no longer served). The typed approach
// catches field-name and structure errors at compile time. The spec mapping
// is split across target_refs.go and validation.go; options is the only
// piece mapped inline (a plain string map, set only when non-empty so the
// key is omitted from the rendered CR otherwise).
func createBackendTlsPolicy(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesBackendTlsPolicy.Spec

	policySpec := gatewayv1.BackendTLSPolicySpecArgs{
		TargetRefs: buildTargetRefs(spec.GetTargetRefs()),
		Validation: buildValidation(spec.GetValidation()),
	}

	if options := spec.GetOptions(); len(options) > 0 {
		policySpec.Options = pulumi.ToStringMap(options)
	}

	_, err := gatewayv1.NewBackendTLSPolicy(ctx, locals.PolicyName,
		&gatewayv1.BackendTLSPolicyArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.PolicyName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: policySpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}
