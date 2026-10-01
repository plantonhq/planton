package module

import (
	"strings"

	azureaksnodepoolv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureaksnodepool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureAksNodePool *azureaksnodepoolv1alpha1.AzureAksNodePool

	// KubernetesClusterId is a StringValueOrRef field; the platform
	// middleware resolves valueFrom references before IaC modules run, so
	// GetValue() always returns the resolved literal ARM ID of the parent
	// cluster.
	KubernetesClusterId string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureaksnodepoolv1alpha1.AzureAksNodePoolStackInput) *Locals {
	locals := &Locals{}

	locals.AzureAksNodePool = stackInput.Target
	target := stackInput.Target

	locals.KubernetesClusterId = target.Spec.KubernetesClusterId.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureAksNodePool.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}

// enum-name -> ARM string maps. Empty results mean "leave unset so Azure
// applies its default" -- keeping an unspecified spec and Azure's default
// identical on both engines.

func osSkuToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolOsSku) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsSku_UBUNTU:
		return "Ubuntu"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsSku_UBUNTU_2204:
		return "Ubuntu2204"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsSku_UBUNTU_2404:
		return "Ubuntu2404"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsSku_AZURE_LINUX:
		return "AzureLinux"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsSku_AZURE_LINUX_3:
		return "AzureLinux3"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsSku_WINDOWS_2019:
		return "Windows2019"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsSku_WINDOWS_2022:
		return "Windows2022"
	}
	return ""
}

func osDiskTypeToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolOsDiskType) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsDiskType_MANAGED:
		return "Managed"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolOsDiskType_EPHEMERAL:
		return "Ephemeral"
	}
	return ""
}

func kubeletDiskTypeToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolKubeletDiskType) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolKubeletDiskType_OS:
		return "OS"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolKubeletDiskType_TEMPORARY:
		return "Temporary"
	}
	return ""
}

func gpuInstanceToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolGpuInstance) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolGpuInstance_MIG1G:
		return "MIG1g"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolGpuInstance_MIG2G:
		return "MIG2g"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolGpuInstance_MIG3G:
		return "MIG3g"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolGpuInstance_MIG4G:
		return "MIG4g"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolGpuInstance_MIG7G:
		return "MIG7g"
	}
	return ""
}

func gpuDriverToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolGpuDriver) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolGpuDriver_INSTALL:
		return "Install"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolGpuDriver_NONE:
		return "None"
	}
	return ""
}

func scaleDownModeToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolScaleDownMode) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolScaleDownMode_DELETE:
		return "Delete"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolScaleDownMode_DEALLOCATE:
		return "Deallocate"
	}
	return ""
}

func workloadRuntimeToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolWorkloadRuntime) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolWorkloadRuntime_OCI_CONTAINER:
		return "OCIContainer"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolWorkloadRuntime_KATA_MSHV_VM_ISOLATION:
		return "KataMshvVmIsolation"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolWorkloadRuntime_WASM_WASI:
		return "WasmWasi"
	}
	return ""
}

func cpuManagerPolicyToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolCpuManagerPolicy) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolCpuManagerPolicy_CPU_MANAGER_NONE:
		return "none"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolCpuManagerPolicy_CPU_MANAGER_STATIC:
		return "static"
	}
	return ""
}

func topologyManagerPolicyToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolTopologyManagerPolicy) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTopologyManagerPolicy_TOPOLOGY_NONE:
		return "none"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTopologyManagerPolicy_BEST_EFFORT:
		return "best-effort"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTopologyManagerPolicy_RESTRICTED:
		return "restricted"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTopologyManagerPolicy_SINGLE_NUMA_NODE:
		return "single-numa-node"
	}
	return ""
}

func transparentHugePageToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePage) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePage_THP_ALWAYS:
		return "always"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePage_THP_MADVISE:
		return "madvise"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePage_THP_NEVER:
		return "never"
	}
	return ""
}

func transparentHugePageDefragToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePageDefrag) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePageDefrag_DEFRAG_ALWAYS:
		return "always"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePageDefrag_DEFRAG_DEFER:
		return "defer"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePageDefrag_DEFRAG_DEFER_MADVISE:
		return "defer+madvise"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePageDefrag_DEFRAG_MADVISE:
		return "madvise"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolTransparentHugePageDefrag_DEFRAG_NEVER:
		return "never"
	}
	return ""
}

func undrainableNodeBehaviorToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolUndrainableNodeBehavior) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolUndrainableNodeBehavior_CORDON:
		return "Cordon"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolUndrainableNodeBehavior_SCHEDULE:
		return "Schedule"
	}
	return ""
}

func hostPortProtocolToArm(v azureaksnodepoolv1alpha1.AzureAksNodePoolHostPortProtocol) string {
	switch v {
	case azureaksnodepoolv1alpha1.AzureAksNodePoolHostPortProtocol_TCP:
		return "TCP"
	case azureaksnodepoolv1alpha1.AzureAksNodePoolHostPortProtocol_UDP:
		return "UDP"
	}
	return ""
}
