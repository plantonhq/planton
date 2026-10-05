package module

import (
	"strconv"

	"github.com/pkg/errors"
	kuberneteskafkaconnectorv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkaconnector/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources renders one kafka.strimzi.io/v1 KafkaConnector as an UNTYPED
// CustomResource — the Strimzi CRDs type `config` with
// x-kubernetes-preserve-unknown-fields, which crd2pulumi cannot carry, so
// no generated package is shipped for the Kafka family (the same ruling as
// the KubernetesKafka module).
//
// The CR deploys into the CONNECT CLUSTER'S OWN namespace and binds to the
// cluster through the strimzi.io/cluster label — the cluster operator
// reconciles connectors only there (the spec comments carry the placement
// contract). No namespace resource is created here, deliberately: the
// namespace belongs to the KubernetesKafkaConnect resource's lifecycle.
//
// The spec body is the exact twin of the Terraform module's
// local.connector_manifest — same keys rendered and omitted, numbers as
// ints. No await machinery: reconciliation belongs to the cluster
// operator (which drives the connector through the Connect REST API), not
// to applying the resource.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskafkaconnectorv1alpha1.KubernetesKafkaConnectorIacInput) error {
	target := iacInput.Target
	spec := target.Spec

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKafkaConnector.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}
	// The Strimzi binding label rides ON TOP of the identity labels —
	// without it the cluster operator never picks the connector up.
	labels["strimzi.io/cluster"] = spec.ConnectCluster.GetValue()

	specBody := map[string]interface{}{
		"class": spec.ConnectorClass,
	}
	if spec.TasksMax != nil {
		specBody["tasksMax"] = int(spec.GetTasksMax())
	}
	if spec.GetVersion() != "" {
		specBody["version"] = spec.GetVersion()
	}
	if len(spec.GetConfig()) > 0 {
		config := make(map[string]interface{}, len(spec.GetConfig()))
		for k, v := range spec.GetConfig() {
			config[k] = v
		}
		specBody["config"] = config
	}
	if spec.GetState() != "" {
		specBody["state"] = spec.GetState()
	}
	if autoRestart := spec.GetAutoRestart(); autoRestart != nil {
		autoRestartBody := map[string]interface{}{
			"enabled": autoRestart.GetEnabled(),
		}
		if autoRestart.MaxRestarts != nil {
			autoRestartBody["maxRestarts"] = int(autoRestart.GetMaxRestarts())
		}
		specBody["autoRestart"] = autoRestartBody
	}
	// The offset ConfigMap targets are declarations only — the list/alter
	// actions run when the resource carries the strimzi.io/connector-offsets
	// annotation (an operational verb outside this module's scope).
	if listOffsets := spec.GetListOffsets(); listOffsets != nil {
		specBody["listOffsets"] = map[string]interface{}{
			"toConfigMap": map[string]interface{}{
				"name": listOffsets.ToConfigMap.GetValue(),
			},
		}
	}
	if alterOffsets := spec.GetAlterOffsets(); alterOffsets != nil {
		specBody["alterOffsets"] = map[string]interface{}{
			"fromConfigMap": map[string]interface{}{
				"name": alterOffsets.FromConfigMap.GetValue(),
			},
		}
	}

	if _, err := apiextensions.NewCustomResource(ctx, target.Metadata.Name,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("kafka.strimzi.io/v1"),
			Kind:       pulumi.String("KafkaConnector"),
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(target.Metadata.Name),
				Namespace: pulumi.String(spec.Namespace.GetValue()),
				Labels:    pulumi.ToStringMap(labels),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": specBody,
			},
		}, pulumi.Provider(kubernetesProvider)); err != nil {
		return errors.Wrap(err, "failed to create kafka connector")
	}

	ctx.Export(OpNamespace, pulumi.String(spec.Namespace.GetValue()))
	ctx.Export(OpConnectorName, pulumi.String(target.Metadata.Name))

	return nil
}
