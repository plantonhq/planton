package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ocicontainerenginenodepoolv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocicontainerenginenodepool/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/containerengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func nodePool(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciContainerEngineNodePool.Spec

	args := &containerengine.NodePoolArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		ClusterId:     pulumi.String(spec.ClusterId.GetValue()),
		NodeShape:     pulumi.String(spec.NodeShape),
		Name:          pulumi.StringPtr(locals.DisplayName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.KubernetesVersion != "" {
		args.KubernetesVersion = pulumi.StringPtr(spec.KubernetesVersion)
	}

	if spec.NodeShapeConfig != nil {
		args.NodeShapeConfig = buildNodeShapeConfig(spec.NodeShapeConfig)
	}

	if spec.NodeSourceDetails != nil {
		args.NodeSourceDetails = buildNodeSourceDetails(spec.NodeSourceDetails)
	}

	args.NodeConfigDetails = buildNodeConfigDetails(spec.NodeConfigDetails, locals.FreeformTags)

	if spec.SshPublicKey != "" {
		args.SshPublicKey = pulumi.StringPtr(spec.SshPublicKey)
	}

	if len(spec.InitialNodeLabels) > 0 {
		args.InitialNodeLabels = buildInitialNodeLabels(spec.InitialNodeLabels)
	}

	if len(spec.NodeMetadata) > 0 {
		args.NodeMetadata = pulumi.ToStringMap(spec.NodeMetadata)
	}

	if spec.NodeEvictionSettings != nil {
		args.NodeEvictionNodePoolSettings = buildNodeEvictionSettings(spec.NodeEvictionSettings)
	}

	if spec.NodePoolCyclingDetails != nil {
		args.NodePoolCyclingDetails = buildNodePoolCyclingDetails(spec.NodePoolCyclingDetails)
	}

	createdNodePool, err := containerengine.NewNodePool(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci container engine node pool")
	}

	ctx.Export(OpNodePoolId, createdNodePool.ID())
	ctx.Export(OpKubernetesVersion, createdNodePool.KubernetesVersion)

	return nil
}

func buildNodeShapeConfig(nsc *ocicontainerenginenodepoolv1.OciContainerEngineNodePoolSpec_NodeShapeConfig) *containerengine.NodePoolNodeShapeConfigArgs {
	args := &containerengine.NodePoolNodeShapeConfigArgs{}

	if nsc.Ocpus != 0 {
		args.Ocpus = pulumi.Float64Ptr(float64(nsc.Ocpus))
	}

	if nsc.MemoryInGbs != 0 {
		args.MemoryInGbs = pulumi.Float64Ptr(float64(nsc.MemoryInGbs))
	}

	return args
}

func buildNodeSourceDetails(nsd *ocicontainerenginenodepoolv1.OciContainerEngineNodePoolSpec_NodeSourceDetails) *containerengine.NodePoolNodeSourceDetailsArgs {
	args := &containerengine.NodePoolNodeSourceDetailsArgs{
		ImageId:    pulumi.String(nsd.ImageId),
		SourceType: pulumi.String("IMAGE"),
	}

	if nsd.BootVolumeSizeInGbs > 0 {
		args.BootVolumeSizeInGbs = pulumi.StringPtr(fmt.Sprintf("%d", nsd.BootVolumeSizeInGbs))
	}

	return args
}

func buildNodeConfigDetails(ncd *ocicontainerenginenodepoolv1.OciContainerEngineNodePoolSpec_NodeConfigDetails, freeformTags map[string]string) *containerengine.NodePoolNodeConfigDetailsArgs {
	args := &containerengine.NodePoolNodeConfigDetailsArgs{
		Size:         pulumi.Int(int(ncd.Size)),
		FreeformTags: pulumi.ToStringMap(freeformTags),
	}

	placementConfigs := make(containerengine.NodePoolNodeConfigDetailsPlacementConfigArray, len(ncd.PlacementConfigs))
	for i, pc := range ncd.PlacementConfigs {
		placementConfigs[i] = buildPlacementConfig(pc)
	}
	args.PlacementConfigs = placementConfigs

	if len(ncd.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(ncd.NsgIds))
		for i, nsg := range ncd.NsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.NsgIds = nsgIds
	}

	if ncd.KmsKeyId != nil && ncd.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.StringPtr(ncd.KmsKeyId.GetValue())
	}

	if ncd.IsPvEncryptionInTransitEnabled {
		args.IsPvEncryptionInTransitEnabled = pulumi.BoolPtr(true)
	}

	if ncd.PodNetworkOptionDetails != nil {
		args.NodePoolPodNetworkOptionDetails = buildPodNetworkOptionDetails(ncd.PodNetworkOptionDetails)
	}

	return args
}

func buildPlacementConfig(pc *ocicontainerenginenodepoolv1.OciContainerEngineNodePoolSpec_PlacementConfig) *containerengine.NodePoolNodeConfigDetailsPlacementConfigArgs {
	args := &containerengine.NodePoolNodeConfigDetailsPlacementConfigArgs{
		AvailabilityDomain: pulumi.String(pc.AvailabilityDomain),
		SubnetId:           pulumi.String(pc.SubnetId.GetValue()),
	}

	if len(pc.FaultDomains) > 0 {
		args.FaultDomains = pulumi.ToStringArray(pc.FaultDomains)
	}

	if pc.CapacityReservationId != nil && pc.CapacityReservationId.GetValue() != "" {
		args.CapacityReservationId = pulumi.StringPtr(pc.CapacityReservationId.GetValue())
	}

	if pc.PreemptibleNodeConfig != nil {
		args.PreemptibleNodeConfig = buildPreemptibleNodeConfig(pc.PreemptibleNodeConfig)
	}

	return args
}

func buildPreemptibleNodeConfig(pnc *ocicontainerenginenodepoolv1.OciContainerEngineNodePoolSpec_PreemptibleNodeConfig) *containerengine.NodePoolNodeConfigDetailsPlacementConfigPreemptibleNodeConfigArgs {
	preemptionAction := &containerengine.NodePoolNodeConfigDetailsPlacementConfigPreemptibleNodeConfigPreemptionActionArgs{
		Type: pulumi.String("TERMINATE"),
	}

	if pnc.IsPreserveBootVolume != nil {
		preemptionAction.IsPreserveBootVolume = pulumi.BoolPtr(*pnc.IsPreserveBootVolume)
	}

	return &containerengine.NodePoolNodeConfigDetailsPlacementConfigPreemptibleNodeConfigArgs{
		PreemptionAction: preemptionAction,
	}
}

func buildPodNetworkOptionDetails(pnod *ocicontainerenginenodepoolv1.OciContainerEngineNodePoolSpec_PodNetworkOptionDetails) *containerengine.NodePoolNodeConfigDetailsNodePoolPodNetworkOptionDetailsArgs {
	args := &containerengine.NodePoolNodeConfigDetailsNodePoolPodNetworkOptionDetailsArgs{
		CniType: pulumi.String(strings.ToUpper(pnod.CniType.String())),
	}

	if pnod.MaxPodsPerNode > 0 {
		args.MaxPodsPerNode = pulumi.IntPtr(int(pnod.MaxPodsPerNode))
	}

	if len(pnod.PodNsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(pnod.PodNsgIds))
		for i, nsg := range pnod.PodNsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.PodNsgIds = nsgIds
	}

	if len(pnod.PodSubnetIds) > 0 {
		subnetIds := make(pulumi.StringArray, len(pnod.PodSubnetIds))
		for i, s := range pnod.PodSubnetIds {
			subnetIds[i] = pulumi.String(s.GetValue())
		}
		args.PodSubnetIds = subnetIds
	}

	return args
}

func buildInitialNodeLabels(labels []*ocicontainerenginenodepoolv1.OciContainerEngineNodePoolSpec_NodeLabel) containerengine.NodePoolInitialNodeLabelArray {
	result := make(containerengine.NodePoolInitialNodeLabelArray, len(labels))
	for i, label := range labels {
		result[i] = &containerengine.NodePoolInitialNodeLabelArgs{
			Key:   pulumi.StringPtr(label.Key),
			Value: pulumi.StringPtr(label.Value),
		}
	}
	return result
}

func buildNodeEvictionSettings(nes *ocicontainerenginenodepoolv1.OciContainerEngineNodePoolSpec_NodeEvictionSettings) *containerengine.NodePoolNodeEvictionNodePoolSettingsArgs {
	args := &containerengine.NodePoolNodeEvictionNodePoolSettingsArgs{}

	if nes.EvictionGraceDuration != "" {
		args.EvictionGraceDuration = pulumi.StringPtr(nes.EvictionGraceDuration)
	}

	if nes.IsForceActionAfterGraceDuration != nil {
		args.IsForceActionAfterGraceDuration = pulumi.BoolPtr(*nes.IsForceActionAfterGraceDuration)
	}

	if nes.IsForceDeleteAfterGraceDuration != nil {
		args.IsForceDeleteAfterGraceDuration = pulumi.BoolPtr(*nes.IsForceDeleteAfterGraceDuration)
	}

	return args
}

func buildNodePoolCyclingDetails(npcd *ocicontainerenginenodepoolv1.OciContainerEngineNodePoolSpec_NodePoolCyclingDetails) *containerengine.NodePoolNodePoolCyclingDetailsArgs {
	args := &containerengine.NodePoolNodePoolCyclingDetailsArgs{
		IsNodeCyclingEnabled: pulumi.BoolPtr(npcd.IsNodeCyclingEnabled),
	}

	if npcd.MaximumSurge != "" {
		args.MaximumSurge = pulumi.StringPtr(npcd.MaximumSurge)
	}

	if npcd.MaximumUnavailable != "" {
		args.MaximumUnavailable = pulumi.StringPtr(npcd.MaximumUnavailable)
	}

	return args
}
