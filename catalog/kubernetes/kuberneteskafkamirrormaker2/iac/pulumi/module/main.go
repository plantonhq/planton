package module

import (
	"github.com/pkg/errors"
	kuberneteskafkamirrormaker2v1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkamirrormaker2/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources deploys one Strimzi-operator-managed MirrorMaker 2
// replication engine:
//
//  1. the namespace (optional, create_namespace),
//  2. the JMX exporter rules ConfigMap (optional, metrics.enabled) —
//     module-owned, referenced by the CR's metricsConfig,
//  3. the kafka.strimzi.io/v1 KafkaMirrorMaker2 CR itself.
//
// The Strimzi cluster operator (KubernetesStrimziKafkaOperator, the
// registry prerequisite) reconciles the CR into a Connect-style worker
// deployment running one MirrorSourceConnector + MirrorCheckpointConnector
// pair per declared mirror.
//
// No await machinery, deliberately: engine readiness depends on the
// operator (image pulls, worker group formation, connector startup) that
// is not part of applying the resources — the never-block-on-a-controller
// posture of every operator-CR kind in the catalog.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskafkamirrormaker2v1alpha1.KubernetesKafkaMirrorMaker2IacInput) error {
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

	// -------------------------- MirrorMaker 2 CR ---------------------------
	if _, err := createMirrorMaker2(ctx, locals, kubernetesProvider, dependencies); err != nil {
		return errors.Wrap(err, "failed to create mirror maker 2")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpMirrormakerName, pulumi.String(locals.MirrorMakerName))
	ctx.Export(OpRestApiEndpoint, pulumi.String(locals.RestApiEndpoint))

	return nil
}
