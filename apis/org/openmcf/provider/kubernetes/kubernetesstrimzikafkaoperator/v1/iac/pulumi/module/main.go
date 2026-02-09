package module

import (
	"github.com/pkg/errors"
	kubernetesstrimzikafkaoperatorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesstrimzikafkaoperator/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry‑point called by the OpenMCF engine.
func Resources(
	ctx *pulumi.Context,
	stackInput *kubernetesstrimzikafkaoperatorv1.KubernetesStrimziKafkaOperatorStackInput,
) error {
	// ------------------------------------------------------------------
	// Provider set‑up
	// ------------------------------------------------------------------
	k8sProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx,
		stackInput.ProviderConfig,
		"kubernetes",
	)
	if err != nil {
		return errors.Wrap(err, "failed to set up Kubernetes provider")
	}

	// Compute locals
	l := newLocals(stackInput)

	// ------------------------------ namespace ----------------------------
	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput.Target, l, k8sProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	// When create_namespace is false, createdNamespace is nil and namespaceDeps is empty.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ------------------------------------------------------------------
	// Helm install
	// ------------------------------------------------------------------
	if err := kafkaOperator(ctx, stackInput.Target, l, k8sProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to install Kafka operator resources")
	}

	return nil
}
