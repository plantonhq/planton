package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/strimzioperator/kubernetes/kafka/v1beta2"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kafkaTopics(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	createdKafkaCluster *v1beta2.Kafka, namespaceDeps []pulumi.ResourceOption) error {
	for _, kafkaTopic := range locals.KubernetesKafka.Spec.KafkaTopics {

		config := vars.KafkaTopicDefaultConfig
		for k, v := range kafkaTopic.Config {
			config[k] = v
		}

		opts := append([]pulumi.ResourceOption{pulumi.Parent(createdKafkaCluster)}, namespaceDeps...)
		_, err := v1beta2.NewKafkaTopic(ctx,
			kafkaTopic.Name,
			&v1beta2.KafkaTopicArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name:      pulumi.String(kafkaTopic.Name),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Spec: v1beta2.KafkaTopicSpecArgs{
					Config:     convertstringmaps.ConvertGoStringMapToPulumiMap(config),
					Partitions: pulumi.Int(int(kafkaTopic.GetPartitions())),
					Replicas:   pulumi.Int(int(kafkaTopic.GetReplicas())),
					TopicName:  pulumi.String(kafkaTopic.Name),
				},
			}, opts...)
		if err != nil {
			return errors.Wrapf(err, "failed to create kafka-topic %s", kafkaTopic.Name)
		}
	}
	return nil
}
