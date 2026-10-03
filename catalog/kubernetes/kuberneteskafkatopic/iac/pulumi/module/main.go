package module

import (
	"strconv"

	"github.com/pkg/errors"
	kuberneteskafkatopicv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkatopic/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources renders one kafka.strimzi.io/v1 KafkaTopic as an UNTYPED
// CustomResource — the Strimzi CRDs type `config` with
// x-kubernetes-preserve-unknown-fields, which crd2pulumi cannot carry, so
// no generated package is shipped for the Kafka family (the same ruling as
// the KubernetesKafka module).
//
// The CR deploys into the KAFKA CLUSTER'S OWN namespace and binds to the
// cluster through the strimzi.io/cluster label — the cluster's topic
// operator watches only there (the spec comments carry the placement
// contract). No namespace resource is created here, deliberately: the
// namespace belongs to the KubernetesKafka resource's lifecycle.
//
// The spec body is the exact twin of the Terraform module's
// local.topic_manifest — same keys rendered and omitted, numbers as ints.
// No await machinery: reconciliation belongs to the topic operator, not to
// applying the resource.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskafkatopicv1alpha1.KubernetesKafkaTopicIacInput) error {
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
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKafkaTopic.String(),
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
	// without it the topic operator never picks the resource up.
	labels["strimzi.io/cluster"] = spec.KafkaCluster.GetValue()

	// The actual Kafka topic name: the override when set (dots,
	// underscores, uppercase — names Kubernetes metadata cannot carry),
	// metadata.name otherwise. Twin of TF's coalesce.
	topicName := spec.GetTopicName()
	if topicName == "" {
		topicName = target.Metadata.Name
	}

	specBody := map[string]interface{}{}
	if spec.GetTopicName() != "" {
		specBody["topicName"] = spec.GetTopicName()
	}
	if spec.Partitions != nil {
		specBody["partitions"] = int(spec.GetPartitions())
	}
	if spec.Replicas != nil {
		specBody["replicas"] = int(spec.GetReplicas())
	}
	if len(spec.GetConfig()) > 0 {
		config := make(map[string]interface{}, len(spec.GetConfig()))
		for k, v := range spec.GetConfig() {
			config[k] = v
		}
		specBody["config"] = config
	}

	if _, err := apiextensions.NewCustomResource(ctx, target.Metadata.Name,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("kafka.strimzi.io/v1"),
			Kind:       pulumi.String("KafkaTopic"),
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(target.Metadata.Name),
				Namespace: pulumi.String(spec.Namespace.GetValue()),
				Labels:    pulumi.ToStringMap(labels),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": specBody,
			},
		}, pulumi.Provider(kubernetesProvider)); err != nil {
		return errors.Wrap(err, "failed to create kafka topic")
	}

	ctx.Export(OpNamespace, pulumi.String(spec.Namespace.GetValue()))
	ctx.Export(OpTopicName, pulumi.String(topicName))

	return nil
}
