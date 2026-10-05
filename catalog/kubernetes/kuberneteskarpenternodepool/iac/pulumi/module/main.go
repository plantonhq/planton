package module

import (
	"github.com/pkg/errors"
	kuberneteskarpenternodepoolv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskarpenternodepool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	karpenterv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/karpenter/kubernetes/karpenter/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *kuberneteskarpenternodepoolv1alpha1.KubernetesKarpenterNodePoolIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createNodePool(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create node pool")
	}

	ctx.Export(OpNodePoolName, pulumi.String(locals.NodePoolName))

	return nil
}

// createNodePool creates the cluster-scoped karpenter.sh/v1 NodePool using
// the typed crd2pulumi SDK (karpenterv1.NewNodePool), which catches
// field-name and structure errors at compile time. The spec mapping is
// split across template.go (the NodeClaim template) and disruption.go
// (consolidation policy and budgets); the scalar top-level fields — limits,
// weight, and the alpha static-capacity replicas — are mapped inline.
// Unset optionals are omitted entirely so the apiserver applies the CRD's
// own defaults (weight in particular is presence-sensitive: an absent
// weight means 0, but 0 is not an accepted literal value).
func createNodePool(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	spec := locals.KubernetesKarpenterNodePool.Spec

	nodePoolSpec := karpenterv1.NodePoolSpecArgs{
		Template: buildTemplate(spec.GetTemplate()),
	}

	if disruption := spec.GetDisruption(); disruption != nil {
		nodePoolSpec.Disruption = buildDisruption(disruption)
	}

	// limits values are Kubernetes quantities kept as strings (the CRD field
	// is int-or-string; the string form round-trips every quantity).
	if limits := spec.GetLimits(); len(limits) > 0 {
		limitsMap := pulumi.Map{}
		for resourceName, quantity := range limits {
			limitsMap[resourceName] = pulumi.String(quantity)
		}
		nodePoolSpec.Limits = limitsMap
	}

	if spec.Weight != nil {
		nodePoolSpec.Weight = pulumi.Int(int(spec.GetWeight()))
	}

	// replicas switches the pool into ALPHA static-capacity mode, so its
	// presence matters: only render it when explicitly set (0 is a valid
	// static size and must survive).
	if spec.Replicas != nil {
		nodePoolSpec.Replicas = pulumi.Int(int(spec.GetReplicas()))
	}

	_, err := karpenterv1.NewNodePool(ctx, locals.NodePoolName,
		&karpenterv1.NodePoolArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.NodePoolName),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
			Spec: nodePoolSpec,
		},
		pulumi.Provider(kubeProvider))

	return err
}
