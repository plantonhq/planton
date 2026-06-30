package module

import (
	"fmt"

	"github.com/pkg/errors"
	ocistreampoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocistreampool/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/streaming"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func streamPool(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciStreamPool.Spec

	args := &streaming.StreamPoolArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		Name:          pulumi.String(locals.PoolName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.KafkaSettings != nil {
		args.KafkaSettings = buildKafkaSettings(spec.KafkaSettings)
	}

	if spec.KmsKeyId != nil {
		args.CustomEncryptionKey = &streaming.StreamPoolCustomEncryptionKeyArgs{
			KmsKeyId: pulumi.String(spec.KmsKeyId.GetValue()),
		}
	}

	if spec.PrivateEndpointSettings != nil {
		args.PrivateEndpointSettings = buildPrivateEndpointSettings(spec.PrivateEndpointSettings)
	}

	pool, err := streaming.NewStreamPool(ctx, locals.PoolName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create stream pool")
	}

	ctx.Export(OpStreamPoolId, pool.ID())
	ctx.Export(OpEndpointFqdn, pool.EndpointFqdn)

	bootstrapServers := pool.KafkaSettings.ApplyT(func(ks streaming.StreamPoolKafkaSettings) string {
		if ks.BootstrapServers != nil {
			return *ks.BootstrapServers
		}
		return ""
	}).(pulumi.StringOutput)
	ctx.Export(OpKafkaBootstrapServers, bootstrapServers)

	for _, s := range spec.Streams {
		streamName := fmt.Sprintf("%s-%s", locals.PoolName, s.Name)
		streamArgs := &streaming.StreamArgs{
			Name:         pulumi.String(s.Name),
			Partitions:   pulumi.Int(int(s.Partitions)),
			StreamPoolId: pool.ID(),
			FreeformTags: pulumi.ToStringMap(locals.FreeformTags),
		}

		if s.RetentionInHours != nil {
			streamArgs.RetentionInHours = pulumi.Int(int(*s.RetentionInHours))
		}

		_, err := streaming.NewStream(ctx, streamName, streamArgs,
			pulumiOciOpt(provider), pulumi.DependsOn([]pulumi.Resource{pool}))
		if err != nil {
			return errors.Wrapf(err, "failed to create stream %s", s.Name)
		}
	}

	return nil
}

func buildKafkaSettings(
	ks *ocistreampoolv1.OciStreamPoolSpec_KafkaSettings,
) *streaming.StreamPoolKafkaSettingsArgs {
	kafkaArgs := &streaming.StreamPoolKafkaSettingsArgs{}

	if ks.AutoCreateTopicsEnable != nil {
		kafkaArgs.AutoCreateTopicsEnable = pulumi.Bool(*ks.AutoCreateTopicsEnable)
	}

	if ks.LogRetentionHours != nil {
		kafkaArgs.LogRetentionHours = pulumi.Int(int(*ks.LogRetentionHours))
	}

	if ks.NumPartitions != nil {
		kafkaArgs.NumPartitions = pulumi.Int(int(*ks.NumPartitions))
	}

	return kafkaArgs
}

func buildPrivateEndpointSettings(
	pe *ocistreampoolv1.OciStreamPoolSpec_PrivateEndpointSettings,
) *streaming.StreamPoolPrivateEndpointSettingsArgs {
	peArgs := &streaming.StreamPoolPrivateEndpointSettingsArgs{
		SubnetId: pulumi.String(pe.SubnetId.GetValue()),
	}

	if len(pe.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(pe.NsgIds))
		for i, n := range pe.NsgIds {
			nsgIds[i] = pulumi.String(n.GetValue())
		}
		peArgs.NsgIds = nsgIds
	}

	if pe.PrivateEndpointIp != "" {
		peArgs.PrivateEndpointIp = pulumi.String(pe.PrivateEndpointIp)
	}

	return peArgs
}
