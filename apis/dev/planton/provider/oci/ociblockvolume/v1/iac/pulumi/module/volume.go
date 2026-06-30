package module

import (
	"fmt"

	"github.com/pkg/errors"
	ociblockvolumev1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociblockvolume/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func volume(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) (*core.Volume, error) {
	spec := locals.OciBlockVolume.Spec

	volumeArgs := &core.VolumeArgs{
		CompartmentId:      pulumi.String(spec.CompartmentId.GetValue()),
		AvailabilityDomain: pulumi.String(spec.AvailabilityDomain),
		FreeformTags:       pulumi.ToStringMap(locals.FreeformTags),
		SizeInGbs:          pulumi.String(fmt.Sprintf("%d", spec.SizeInGbs)),
	}

	if spec.DisplayName != "" {
		volumeArgs.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}

	if spec.VpusPerGb != nil {
		volumeArgs.VpusPerGb = pulumi.String(fmt.Sprintf("%d", *spec.VpusPerGb))
	}

	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		volumeArgs.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	if spec.IsReservationsEnabled {
		volumeArgs.IsReservationsEnabled = pulumi.BoolPtr(true)
	}

	if spec.XrcKmsKeyId != nil && spec.XrcKmsKeyId.GetValue() != "" {
		volumeArgs.XrcKmsKeyId = pulumi.StringPtr(spec.XrcKmsKeyId.GetValue())
	}

	if len(spec.AutotunePolicies) > 0 {
		volumeArgs.AutotunePolicies = buildAutotunePolicies(spec.AutotunePolicies)
	}

	if len(spec.BlockVolumeReplicas) > 0 {
		volumeArgs.BlockVolumeReplicas = buildBlockVolumeReplicas(spec.BlockVolumeReplicas)
	}

	createdVolume, err := core.NewVolume(ctx, locals.DisplayName, volumeArgs, pulumiOciOpt(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create volume")
	}

	ctx.Export(OpVolumeId, createdVolume.ID())

	return createdVolume, nil
}

func buildAutotunePolicies(policies []*ociblockvolumev1.OciBlockVolumeSpec_AutotunePolicy) core.VolumeAutotunePolicyArray {
	var result core.VolumeAutotunePolicyArray
	for _, p := range policies {
		entry := core.VolumeAutotunePolicyArgs{
			AutotuneType: pulumi.String(autotuneTypeMap[p.AutotuneType]),
		}
		if p.MaxVpusPerGb > 0 {
			entry.MaxVpusPerGb = pulumi.StringPtr(fmt.Sprintf("%d", p.MaxVpusPerGb))
		}
		result = append(result, entry)
	}
	return result
}

func buildBlockVolumeReplicas(replicas []*ociblockvolumev1.OciBlockVolumeSpec_BlockVolumeReplica) core.VolumeBlockVolumeReplicaArray {
	var result core.VolumeBlockVolumeReplicaArray
	for _, r := range replicas {
		entry := core.VolumeBlockVolumeReplicaArgs{
			AvailabilityDomain: pulumi.String(r.AvailabilityDomain),
		}
		if r.DisplayName != "" {
			entry.DisplayName = pulumi.StringPtr(r.DisplayName)
		}
		if r.XrrKmsKeyId != nil && r.XrrKmsKeyId.GetValue() != "" {
			entry.XrrKmsKeyId = pulumi.StringPtr(r.XrrKmsKeyId.GetValue())
		}
		result = append(result, entry)
	}
	return result
}
