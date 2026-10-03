package module

import (
	"github.com/pkg/errors"
	kuberneteskarpenterec2nodeclassv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskarpenterec2nodeclass/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	karpenterv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/karpenter/kubernetes/karpenter/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *kuberneteskarpenterec2nodeclassv1alpha1.KubernetesKarpenterEc2NodeClassIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	if err := createEc2NodeClass(ctx, kubeProvider, locals); err != nil {
		return errors.Wrap(err, "failed to create ec2 node class")
	}

	ctx.Export(OpNodeClassName, pulumi.String(locals.NodeClassName))

	return nil
}

// createEc2NodeClass creates the CLUSTER-SCOPED karpenter.k8s.aws/v1
// EC2NodeClass using the typed crd2pulumi SDK (karpenterv1.NewEC2NodeClass) —
// no namespace anywhere. The typed approach catches field-name and structure
// errors at compile time. The spec mapping lives in node_class_spec.go;
// unset optionals are omitted entirely so CRD-side defaults (metadataOptions
// et al.) are never overridden by rendered zero values.
func createEc2NodeClass(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	locals *Locals,
) error {
	_, err := karpenterv1.NewEC2NodeClass(ctx, locals.NodeClassName,
		&karpenterv1.EC2NodeClassArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.NodeClassName),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
			Spec: buildNodeClassSpec(locals.KubernetesKarpenterEc2NodeClass.Spec),
		},
		pulumi.Provider(kubeProvider))

	return err
}
