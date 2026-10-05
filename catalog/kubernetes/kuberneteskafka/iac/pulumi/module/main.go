package module

import (
	"github.com/pkg/errors"
	kuberneteskafkav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafka/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one Strimzi-operator-managed KRaft Kafka cluster:
//
//  1. the namespace (optional, create_namespace),
//  2. the JMX exporter rules ConfigMap (optional, metrics.enabled) —
//     module-owned, referenced by the Kafka CR's metricsConfig,
//  3. one kafka.strimzi.io/v1 KafkaNodePool per spec.node_pools entry,
//  4. the kafka.strimzi.io/v1 Kafka CR itself.
//
// The Strimzi cluster operator (KubernetesStrimziKafkaOperator, the
// registry prerequisite) reconciles the CRs into brokers, controllers,
// listeners, certificates, and the per-cluster entity operators that
// serve KubernetesKafkaTopic / KubernetesKafkaUser declarations.
//
// No await machinery, deliberately: cluster readiness depends on the
// operator (image pulls, KRaft quorum formation) that is not part of
// applying the resources — the never-block-on-a-controller posture of
// every operator-CR kind in the catalog.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskafkav1alpha1.KubernetesKafkaIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, iacInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var dependencies []pulumi.ResourceOption
	if createdNamespace != nil {
		dependencies = append(dependencies, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// --------------------- metrics rules ConfigMap ------------------------
	createdConfigMap, err := metricsConfigMap(ctx, locals, kubernetesProvider, dependencies)
	if err != nil {
		return err
	}
	if createdConfigMap != nil {
		dependencies = append(dependencies, pulumi.DependsOn([]pulumi.Resource{createdConfigMap}))
	}

	// ------------------------------ node pools ----------------------------
	createdPools, err := createNodePools(ctx, locals, kubernetesProvider, dependencies)
	if err != nil {
		return errors.Wrap(err, "failed to create kafka node pools")
	}
	if len(createdPools) > 0 {
		dependencies = append(dependencies, pulumi.DependsOn(createdPools))
	}

	// ------------------------------ Kafka CR -------------------------------
	if _, err := createKafkaCluster(ctx, locals, kubernetesProvider, dependencies); err != nil {
		return errors.Wrap(err, "failed to create kafka cluster")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpClusterName, pulumi.String(locals.ClusterName))
	ctx.Export(OpBootstrapServiceName, pulumi.String(locals.BootstrapServiceName))
	ctx.Export(OpInternalBootstrapEndpoint, pulumi.String(locals.InternalBootstrapEndpoint))
	ctx.Export(OpClusterCaCertSecretName, pulumi.String(locals.ClusterCaCertSecretName))

	return nil
}
