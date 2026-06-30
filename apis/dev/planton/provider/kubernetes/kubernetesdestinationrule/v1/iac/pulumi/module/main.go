package module

import (
	"github.com/pkg/errors"
	kubernetesdestinationrulev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesdestinationrule/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	istionetworkingv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/istio/kubernetes/networking/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesdestinationrulev1.KubernetesDestinationRuleStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createDestinationRule(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create destination rule")
	}

	ctx.Export(OpDestinationRuleName, pulumi.String(locals.DestinationRuleName))
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return nil
}

// createDestinationRule creates the namespaced Istio DestinationRule using the typed
// crd2pulumi SDK (istionetworkingv1.NewDestinationRule), consistent with every other Planton
// Istio component. The typed approach catches field-name and structure errors at compile time.
// Only `host` is always set (it is required upstream); every other block is attached
// only when present (the per-path builders in traffic_policy.go return nil for absent protos),
// so unset fields fall through to istiod's defaults.
func createDestinationRule(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesDestinationRule.Spec

	// The typed resource's Spec field is a PtrInput satisfied by the Args value itself
	// (not the SpecPtr() wrapper, which marshals to the wrong element type); assigned
	// directly below, mirroring the sibling Istio components.
	drSpec := istionetworkingv1.DestinationRuleSpecArgs{
		Host: pulumi.String(spec.GetHost()),
	}

	if exportTo := spec.GetExportTo(); len(exportTo) > 0 {
		drSpec.ExportTo = pulumi.ToStringArray(exportTo)
	}
	if tp := buildTrafficPolicy(spec.GetTrafficPolicy()); tp != nil {
		drSpec.TrafficPolicy = tp
	}
	if subsets := buildSubsets(spec.GetSubsets()); subsets != nil {
		drSpec.Subsets = subsets
	}
	if selector := spec.GetWorkloadSelector(); selector != nil && len(selector.GetMatchLabels()) > 0 {
		drSpec.WorkloadSelector = istionetworkingv1.DestinationRuleSpecWorkloadSelectorArgs{
			MatchLabels: pulumi.ToStringMap(selector.GetMatchLabels()),
		}
	}

	_, err := istionetworkingv1.NewDestinationRule(ctx, locals.DestinationRuleName,
		&istionetworkingv1.DestinationRuleArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.DestinationRuleName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: drSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}

// buildSubsets maps the Planton subsets to the typed SDK subset args. Each subset's traffic
// policy is built with the subset-path builders (see traffic_policy.go).
func buildSubsets(subsets []*kubernetesdestinationrulev1.KubernetesDestinationRuleSubset) istionetworkingv1.DestinationRuleSpecSubsetsArrayInput {
	if len(subsets) == 0 {
		return nil
	}
	out := istionetworkingv1.DestinationRuleSpecSubsetsArray{}
	for _, s := range subsets {
		args := istionetworkingv1.DestinationRuleSpecSubsetsArgs{
			Name: pulumi.String(s.GetName()),
		}
		if labels := s.GetLabels(); len(labels) > 0 {
			args.Labels = pulumi.ToStringMap(labels)
		}
		if tp := buildSubsetTrafficPolicy(s.GetTrafficPolicy()); tp != nil {
			args.TrafficPolicy = tp
		}
		out = append(out, args)
	}
	return out
}
