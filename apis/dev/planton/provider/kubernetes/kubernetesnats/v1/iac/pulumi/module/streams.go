package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kubernetesnatsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesnats/v1"
	nackv1beta2 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/nack/kubernetes/jetstream/v1beta2"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// streams creates JetStream Stream and Consumer custom resources.
// These resources are reconciled by the NACK controller to create actual
// JetStream streams and consumers in the NATS cluster.
//
// This must be deployed after:
// 1. NATS Helm chart (provides the NATS server)
// 2. NACK CRDs (so Kubernetes accepts the CR schemas)
// 3. NACK controller (so the CRs get reconciled)
func streams(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	nackController pulumi.Resource, namespaceDeps []pulumi.ResourceOption) error {

	// Skip if NACK controller is not enabled or no streams defined
	if locals.KubernetesNats.Spec.NackController == nil ||
		!locals.KubernetesNats.Spec.NackController.Enabled {
		return nil
	}

	if len(locals.KubernetesNats.Spec.Streams) == 0 {
		return nil
	}

	// Streams depend on NACK controller
	deps := []pulumi.Resource{nackController}

	// Track created stream names for output
	var createdStreams []string

	// Create each stream and its consumers
	for _, stream := range locals.KubernetesNats.Spec.Streams {
		streamResource, err := createStream(ctx, locals, kubernetesProvider, stream, deps, namespaceDeps)
		if err != nil {
			return errors.Wrapf(err, "failed to create stream %s", stream.Name)
		}

		createdStreams = append(createdStreams, stream.Name)

		// Create consumers for this stream
		for _, consumer := range stream.Consumers {
			if err := createConsumer(ctx, locals, kubernetesProvider, stream, consumer, streamResource, namespaceDeps); err != nil {
				return errors.Wrapf(err, "failed to create consumer %s for stream %s",
					consumer.DurableName, stream.Name)
			}
		}
	}

	// Export list of created streams
	ctx.Export(OpStreamsCreated, pulumi.ToStringArray(createdStreams))

	return nil
}

// createStream creates a JetStream Stream custom resource using strongly-typed nack types
func createStream(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	stream *kubernetesnatsv1.KubernetesNatsStream, deps []pulumi.Resource, namespaceDeps []pulumi.ResourceOption) (pulumi.Resource, error) {

	// Build stream spec
	spec := &nackv1beta2.StreamSpecArgs{
		Name:     pulumi.StringPtr(stream.Name),
		Subjects: pulumi.ToStringArray(stream.Subjects),
	}

	// Storage type
	if stream.Storage != kubernetesnatsv1.StreamStorageEnum_unspecified {
		spec.Storage = pulumi.StringPtr(storageEnumToString(stream.Storage))
	}

	// Replicas
	if stream.Replicas > 0 {
		spec.Replicas = pulumi.IntPtr(int(stream.Replicas))
	}

	// Retention policy
	if stream.Retention != kubernetesnatsv1.StreamRetentionEnum_unspecified {
		spec.Retention = pulumi.StringPtr(retentionEnumToString(stream.Retention))
	}

	// Max age
	if stream.MaxAge != "" {
		spec.MaxAge = pulumi.StringPtr(stream.MaxAge)
	}

	// Max bytes (-1 for unlimited is the default)
	if stream.MaxBytes != 0 {
		spec.MaxBytes = pulumi.IntPtr(int(stream.MaxBytes))
	}

	// Max messages
	if stream.MaxMsgs != 0 {
		spec.MaxMsgs = pulumi.IntPtr(int(stream.MaxMsgs))
	}

	// Max message size
	if stream.MaxMsgSize != 0 {
		spec.MaxMsgSize = pulumi.IntPtr(int(stream.MaxMsgSize))
	}

	// Max consumers
	if stream.MaxConsumers != 0 {
		spec.MaxConsumers = pulumi.IntPtr(int(stream.MaxConsumers))
	}

	// Discard policy
	if stream.Discard != kubernetesnatsv1.StreamDiscardEnum_unspecified {
		spec.Discard = pulumi.StringPtr(discardEnumToString(stream.Discard))
	}

	// Description
	if stream.Description != "" {
		spec.Description = pulumi.StringPtr(stream.Description)
	}

	// Resource name: lowercase stream name with instance prefix for uniqueness
	resourceName := fmt.Sprintf("%s-stream-%s", locals.KubernetesNats.Metadata.Name,
		strings.ToLower(stream.Name))

	streamOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn(deps),
	}, namespaceDeps...)
	streamCR, err := nackv1beta2.NewStream(ctx, resourceName,
		&nackv1beta2.StreamArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(strings.ToLower(stream.Name)),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: spec,
		},
		streamOpts...,
	)
	if err != nil {
		return nil, err
	}

	return streamCR, nil
}

// createConsumer creates a JetStream Consumer custom resource using strongly-typed nack types
func createConsumer(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource,
	stream *kubernetesnatsv1.KubernetesNatsStream, consumer *kubernetesnatsv1.KubernetesNatsConsumer,
	streamResource pulumi.Resource, namespaceDeps []pulumi.ResourceOption) error {

	// Build consumer spec
	spec := &nackv1beta2.ConsumerSpecArgs{
		StreamName:  pulumi.StringPtr(stream.Name),
		DurableName: pulumi.StringPtr(consumer.DurableName),
	}

	// Deliver policy
	if consumer.DeliverPolicy != kubernetesnatsv1.ConsumerDeliverPolicyEnum_unspecified {
		spec.DeliverPolicy = pulumi.StringPtr(deliverPolicyEnumToString(consumer.DeliverPolicy))
	}

	// Ack policy
	if consumer.AckPolicy != kubernetesnatsv1.ConsumerAckPolicyEnum_unspecified {
		spec.AckPolicy = pulumi.StringPtr(ackPolicyEnumToString(consumer.AckPolicy))
	}

	// Filter subject
	if consumer.FilterSubject != "" {
		spec.FilterSubject = pulumi.StringPtr(consumer.FilterSubject)
	}

	// Deliver subject (for push consumers)
	if consumer.DeliverSubject != "" {
		spec.DeliverSubject = pulumi.StringPtr(consumer.DeliverSubject)
	}

	// Deliver group (queue group)
	if consumer.DeliverGroup != "" {
		spec.DeliverGroup = pulumi.StringPtr(consumer.DeliverGroup)
	}

	// Max ack pending
	if consumer.MaxAckPending > 0 {
		spec.MaxAckPending = pulumi.IntPtr(int(consumer.MaxAckPending))
	}

	// Max deliver
	if consumer.MaxDeliver != 0 {
		spec.MaxDeliver = pulumi.IntPtr(int(consumer.MaxDeliver))
	}

	// Ack wait
	if consumer.AckWait != "" {
		spec.AckWait = pulumi.StringPtr(consumer.AckWait)
	}

	// Replay policy
	if consumer.ReplayPolicy != kubernetesnatsv1.ConsumerReplayPolicyEnum_unspecified {
		spec.ReplayPolicy = pulumi.StringPtr(replayPolicyEnumToString(consumer.ReplayPolicy))
	}

	// Description
	if consumer.Description != "" {
		spec.Description = pulumi.StringPtr(consumer.Description)
	}

	// Resource name: stream-consumer combination with instance prefix
	resourceName := fmt.Sprintf("%s-consumer-%s-%s",
		locals.KubernetesNats.Metadata.Name,
		strings.ToLower(stream.Name),
		strings.ToLower(consumer.DurableName))

	// Consumer depends on the stream being created first
	deps := []pulumi.Resource{streamResource}

	consumerOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.DependsOn(deps),
	}, namespaceDeps...)
	_, err := nackv1beta2.NewConsumer(ctx, resourceName,
		&nackv1beta2.ConsumerArgs{
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(strings.ToLower(consumer.DurableName)),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: spec,
		},
		consumerOpts...,
	)

	return err
}

// Enum to string conversion helpers
// These convert proto enum values to NACK CRD string values

func storageEnumToString(s kubernetesnatsv1.StreamStorageEnum_Value) string {
	switch s {
	case kubernetesnatsv1.StreamStorageEnum_file:
		return "file"
	case kubernetesnatsv1.StreamStorageEnum_memory:
		return "memory"
	default:
		return "memory" // default per NACK CRD
	}
}

func retentionEnumToString(r kubernetesnatsv1.StreamRetentionEnum_Value) string {
	switch r {
	case kubernetesnatsv1.StreamRetentionEnum_limits:
		return "limits"
	case kubernetesnatsv1.StreamRetentionEnum_interest:
		return "interest"
	case kubernetesnatsv1.StreamRetentionEnum_workqueue:
		return "workqueue"
	default:
		return "limits" // default per NACK CRD
	}
}

func discardEnumToString(d kubernetesnatsv1.StreamDiscardEnum_Value) string {
	switch d {
	case kubernetesnatsv1.StreamDiscardEnum_old:
		return "old"
	// Note: Proto enum is "new_msgs" because "new" is a reserved keyword in Java.
	// We convert it back to "new" for NACK CRDs.
	case kubernetesnatsv1.StreamDiscardEnum_new_msgs:
		return "new"
	default:
		return "old" // default per NACK CRD
	}
}

func deliverPolicyEnumToString(d kubernetesnatsv1.ConsumerDeliverPolicyEnum_Value) string {
	switch d {
	case kubernetesnatsv1.ConsumerDeliverPolicyEnum_all:
		return "all"
	case kubernetesnatsv1.ConsumerDeliverPolicyEnum_last:
		return "last"
	// Note: Proto enum is "new_msgs" because "new" is a reserved keyword in Java.
	// We convert it back to "new" for NACK CRDs.
	case kubernetesnatsv1.ConsumerDeliverPolicyEnum_new_msgs:
		return "new"
	default:
		return "all" // default per NACK CRD
	}
}

func ackPolicyEnumToString(a kubernetesnatsv1.ConsumerAckPolicyEnum_Value) string {
	switch a {
	case kubernetesnatsv1.ConsumerAckPolicyEnum_none:
		return "none"
	case kubernetesnatsv1.ConsumerAckPolicyEnum_all:
		return "all"
	case kubernetesnatsv1.ConsumerAckPolicyEnum_explicit:
		return "explicit"
	default:
		return "none" // default per NACK CRD
	}
}

func replayPolicyEnumToString(r kubernetesnatsv1.ConsumerReplayPolicyEnum_Value) string {
	switch r {
	case kubernetesnatsv1.ConsumerReplayPolicyEnum_original:
		return "original"
	case kubernetesnatsv1.ConsumerReplayPolicyEnum_instant:
		return "instant"
	default:
		return "instant" // default per NACK CRD
	}
}
