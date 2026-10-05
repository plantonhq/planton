package module

import (
	"github.com/pkg/errors"
	kuberneteskafkaconnectv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkaconnect/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one Strimzi-operator-managed Kafka Connect cluster:
//
//  1. the namespace (optional, create_namespace),
//  2. the JMX exporter rules ConfigMap (optional, metrics.enabled) —
//     module-owned, referenced by the KafkaConnect CR's metricsConfig,
//  3. the kafka.strimzi.io/v1 KafkaConnect CR itself, always annotated
//     strimzi.io/use-connector-resources: "true" so connectors on this
//     cluster are managed declaratively through KubernetesKafkaConnector
//     resources (the operator reverts REST-API-made changes).
//
// The Strimzi cluster operator (KubernetesStrimziKafkaOperator, the
// registry prerequisite) reconciles the CR into Connect worker pods and
// the REST API Service — and, when spec.build is set, runs the
// Kaniko/Buildah image build first.
//
// No await machinery, deliberately: worker readiness depends on the
// operator (image pulls or builds, Connect group formation) that is not
// part of applying the resources — the never-block-on-a-controller
// posture of every operator-CR kind in the catalog.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskafkaconnectv1alpha1.KubernetesKafkaConnectIacInput) error {
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

	// --------------------------- KafkaConnect CR ---------------------------
	if _, err := createKafkaConnect(ctx, locals, kubernetesProvider, dependencies); err != nil {
		return errors.Wrap(err, "failed to create kafka connect cluster")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpConnectName, pulumi.String(locals.ConnectName))
	ctx.Export(OpRestApiServiceName, pulumi.String(locals.RestApiServiceName))
	ctx.Export(OpRestApiEndpoint, pulumi.String(locals.RestApiEndpoint))

	return nil
}
