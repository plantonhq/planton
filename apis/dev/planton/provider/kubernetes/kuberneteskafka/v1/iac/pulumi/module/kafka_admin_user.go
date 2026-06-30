package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/strimzioperator/kubernetes/kafka/v1beta2"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func kafkaAdminUser(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	createdKafkaCluster *v1beta2.Kafka, namespaceDeps []pulumi.ResourceOption) error {

	labels := locals.Labels
	//add the label required to create the admin secret for the target kafka-cluster
	labels[vars.ClusterLabelKey] = locals.KubernetesKafka.Metadata.Name

	opts := append([]pulumi.ResourceOption{pulumi.Parent(createdKafkaCluster)}, namespaceDeps...)
	_, err := v1beta2.NewKafkaUser(ctx,
		"admin-user",
		&v1beta2.KafkaUserArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.AdminUsername),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(labels),
			},
			Spec: v1beta2.KafkaUserSpecArgs{
				Authentication: v1beta2.KafkaUserSpecAuthenticationArgs{
					Type: pulumi.String("scram-sha-512"),
				},
			},
		}, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create kafka admin user")
	}
	return nil
}
