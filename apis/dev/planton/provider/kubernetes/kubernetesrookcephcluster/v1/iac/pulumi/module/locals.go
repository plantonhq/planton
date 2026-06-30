package module

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	kubernetesrookcephclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesrookcephcluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed configuration values from the stack input
type Locals struct {
	// KubernetesRookCephCluster is the target resource
	KubernetesRookCephCluster *kubernetesrookcephclusterv1.KubernetesRookCephCluster

	// Namespace is the Kubernetes namespace to deploy to
	Namespace string

	// OperatorNamespace is the namespace where the operator is installed
	OperatorNamespace string

	// Labels are common labels applied to all resources
	Labels map[string]string

	// HelmReleaseName is the name of the Helm release
	HelmReleaseName string

	// CephClusterName is the name of the CephCluster resource
	CephClusterName string

	// ChartVersion is the Helm chart version to install (without 'v' prefix)
	ChartVersion string

	// HelmValues contains computed values for the Helm release
	HelmValues map[string]interface{}

	// BlockPoolNames list of block pool names
	BlockPoolNames []string

	// BlockStorageClassNames list of block storage class names
	BlockStorageClassNames []string

	// FilesystemNames list of filesystem names
	FilesystemNames []string

	// FilesystemStorageClassNames list of filesystem storage class names
	FilesystemStorageClassNames []string

	// ObjectStoreNames list of object store names
	ObjectStoreNames []string

	// ObjectStorageClassNames list of object storage class names
	ObjectStorageClassNames []string
}

// initializeLocals creates computed values from stack input
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesrookcephclusterv1.KubernetesRookCephClusterStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesRookCephCluster = stackInput.Target

	target := stackInput.Target

	// Build common labels
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesRookCephCluster.String(),
	}

	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// Get namespace from spec
	locals.Namespace = target.Spec.Namespace.GetValue()
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Operator namespace
	locals.OperatorNamespace = target.Spec.GetOperatorNamespace()
	if locals.OperatorNamespace == "" {
		locals.OperatorNamespace = "rook-ceph"
	}

	// Helm release name based on metadata name
	locals.HelmReleaseName = target.Metadata.Name
	ctx.Export(OpHelmReleaseName, pulumi.String(locals.HelmReleaseName))

	// Ceph cluster name
	locals.CephClusterName = target.Metadata.Name
	ctx.Export(OpCephClusterName, pulumi.String(locals.CephClusterName))

	// Helm chart version without 'v' prefix
	chartVersion := target.Spec.GetHelmChartVersion()
	if chartVersion == "" {
		chartVersion = "v1.16.6"
	}
	locals.ChartVersion = strings.TrimPrefix(chartVersion, "v")

	// Collect storage resource names
	locals.collectStorageNames(target)

	// Export storage names
	ctx.Export(OpBlockPoolNames, pulumi.ToStringArray(locals.BlockPoolNames))
	ctx.Export(OpBlockStorageClassNames, pulumi.ToStringArray(locals.BlockStorageClassNames))
	ctx.Export(OpFilesystemNames, pulumi.ToStringArray(locals.FilesystemNames))
	ctx.Export(OpFilesystemStorageClassNames, pulumi.ToStringArray(locals.FilesystemStorageClassNames))
	ctx.Export(OpObjectStoreNames, pulumi.ToStringArray(locals.ObjectStoreNames))
	ctx.Export(OpObjectStorageClassNames, pulumi.ToStringArray(locals.ObjectStorageClassNames))

	// Export dashboard and toolbox commands
	dashboardPortForward := fmt.Sprintf("kubectl port-forward svc/rook-ceph-mgr-dashboard -n %s 7000:7000", locals.Namespace)
	ctx.Export(OpDashboardPortForwardCommand, pulumi.String(dashboardPortForward))
	ctx.Export(OpDashboardUrl, pulumi.String("https://localhost:7000"))

	dashboardPasswordCmd := fmt.Sprintf("kubectl -n %s get secret rook-ceph-dashboard-password -o jsonpath=\"{['data']['password']}\" | base64 -d", locals.Namespace)
	ctx.Export(OpDashboardPasswordCommand, pulumi.String(dashboardPasswordCmd))

	toolboxCmd := fmt.Sprintf("kubectl -n %s exec -it deploy/rook-ceph-tools -- bash", locals.Namespace)
	ctx.Export(OpToolboxExecCommand, pulumi.String(toolboxCmd))

	// Build Helm values
	locals.HelmValues = buildHelmValues(target, locals)

	return locals
}

// collectStorageNames collects all storage pool and class names
func (l *Locals) collectStorageNames(target *kubernetesrookcephclusterv1.KubernetesRookCephCluster) {
	// Block pools
	for _, bp := range target.Spec.BlockPools {
		l.BlockPoolNames = append(l.BlockPoolNames, bp.Name)
		if bp.StorageClass != nil && bp.StorageClass.GetEnabled() {
			l.BlockStorageClassNames = append(l.BlockStorageClassNames, bp.StorageClass.Name)
		}
	}

	// Filesystems
	for _, fs := range target.Spec.Filesystems {
		l.FilesystemNames = append(l.FilesystemNames, fs.Name)
		if fs.StorageClass != nil && fs.StorageClass.GetEnabled() {
			l.FilesystemStorageClassNames = append(l.FilesystemStorageClassNames, fs.StorageClass.Name)
		}
	}

	// Object stores
	for _, os := range target.Spec.ObjectStores {
		l.ObjectStoreNames = append(l.ObjectStoreNames, os.Name)
		if os.StorageClass != nil && os.StorageClass.GetEnabled() {
			l.ObjectStorageClassNames = append(l.ObjectStorageClassNames, os.StorageClass.Name)
		}
	}
}

// buildHelmValues constructs the Helm values map from spec
func buildHelmValues(target *kubernetesrookcephclusterv1.KubernetesRookCephCluster, locals *Locals) map[string]interface{} {
	values := map[string]interface{}{}

	// Operator namespace
	values["operatorNamespace"] = locals.OperatorNamespace

	// Cluster name
	values["clusterName"] = locals.CephClusterName

	// Ceph image configuration
	if target.Spec.CephImage != nil {
		cephImage := map[string]interface{}{}
		if target.Spec.CephImage.GetRepository() != "" {
			cephImage["repository"] = target.Spec.CephImage.GetRepository()
		}
		if target.Spec.CephImage.GetTag() != "" {
			cephImage["tag"] = target.Spec.CephImage.GetTag()
		}
		cephImage["allowUnsupported"] = target.Spec.CephImage.GetAllowUnsupported()
		values["cephImage"] = cephImage
	}

	// Toolbox configuration
	values["toolbox"] = map[string]interface{}{
		"enabled": target.Spec.GetEnableToolbox(),
	}

	// Monitoring configuration
	values["monitoring"] = map[string]interface{}{
		"enabled": target.Spec.GetEnableMonitoring(),
	}

	// Build cephClusterSpec
	cephClusterSpec := buildCephClusterSpec(target)
	values["cephClusterSpec"] = cephClusterSpec

	// Block pools
	if len(target.Spec.BlockPools) > 0 {
		values["cephBlockPools"] = buildBlockPools(target)
	} else {
		values["cephBlockPools"] = []interface{}{}
	}

	// Filesystems
	if len(target.Spec.Filesystems) > 0 {
		values["cephFileSystems"] = buildFilesystems(target)
	} else {
		values["cephFileSystems"] = []interface{}{}
	}

	// Object stores
	if len(target.Spec.ObjectStores) > 0 {
		values["cephObjectStores"] = buildObjectStores(target)
	} else {
		values["cephObjectStores"] = []interface{}{}
	}

	return values
}

// buildCephClusterSpec builds the cephClusterSpec section
func buildCephClusterSpec(target *kubernetesrookcephclusterv1.KubernetesRookCephCluster) map[string]interface{} {
	spec := map[string]interface{}{}

	cluster := target.Spec.Cluster
	if cluster == nil {
		cluster = &kubernetesrookcephclusterv1.CephClusterConfig{}
	}

	// Data directory host path
	dataDirHostPath := cluster.GetDataDirHostPath()
	if dataDirHostPath == "" {
		dataDirHostPath = "/var/lib/rook"
	}
	spec["dataDirHostPath"] = dataDirHostPath

	// Dashboard
	spec["dashboard"] = map[string]interface{}{
		"enabled": target.Spec.GetEnableDashboard(),
		"ssl":     true,
	}

	// Monitor configuration
	if cluster.Mon != nil {
		monCount := cluster.Mon.GetCount()
		if monCount == 0 {
			monCount = 3
		}
		spec["mon"] = map[string]interface{}{
			"count":                monCount,
			"allowMultiplePerNode": cluster.Mon.GetAllowMultiplePerNode(),
		}
	} else {
		spec["mon"] = map[string]interface{}{
			"count":                3,
			"allowMultiplePerNode": false,
		}
	}

	// Manager configuration
	if cluster.Mgr != nil {
		mgrCount := cluster.Mgr.GetCount()
		if mgrCount == 0 {
			mgrCount = 2
		}
		spec["mgr"] = map[string]interface{}{
			"count":                mgrCount,
			"allowMultiplePerNode": cluster.Mgr.GetAllowMultiplePerNode(),
		}
	} else {
		spec["mgr"] = map[string]interface{}{
			"count":                2,
			"allowMultiplePerNode": false,
		}
	}

	// Storage configuration
	storageSpec := buildStorageSpec(cluster.Storage)
	spec["storage"] = storageSpec

	// Network configuration
	if cluster.Network != nil {
		spec["network"] = map[string]interface{}{
			"connections": map[string]interface{}{
				"encryption": map[string]interface{}{
					"enabled": cluster.Network.GetEnableEncryption(),
				},
				"compression": map[string]interface{}{
					"enabled": cluster.Network.GetEnableCompression(),
				},
				"requireMsgr2": cluster.Network.GetRequireMsgr2(),
			},
		}
	}

	// Resource configuration
	if cluster.Resources != nil {
		resources := map[string]interface{}{}
		if cluster.Resources.Mon != nil {
			resources["mon"] = buildResourceSpec(cluster.Resources.Mon)
		}
		if cluster.Resources.Mgr != nil {
			resources["mgr"] = buildResourceSpec(cluster.Resources.Mgr)
		}
		if cluster.Resources.Osd != nil {
			resources["osd"] = buildResourceSpec(cluster.Resources.Osd)
		}
		if len(resources) > 0 {
			spec["resources"] = resources
		}
	}

	return spec
}

// buildStorageSpec builds the storage section
func buildStorageSpec(storage *kubernetesrookcephclusterv1.CephStorageSpec) map[string]interface{} {
	storageSpec := map[string]interface{}{}

	if storage == nil {
		storageSpec["useAllNodes"] = true
		storageSpec["useAllDevices"] = true
		return storageSpec
	}

	storageSpec["useAllNodes"] = storage.GetUseAllNodes()
	storageSpec["useAllDevices"] = storage.GetUseAllDevices()

	if storage.DeviceFilter != "" {
		storageSpec["deviceFilter"] = storage.DeviceFilter
	}

	// Specific nodes configuration
	if len(storage.Nodes) > 0 {
		nodes := []interface{}{}
		for _, node := range storage.Nodes {
			nodeSpec := map[string]interface{}{
				"name": node.Name,
			}
			if len(node.Devices) > 0 {
				devices := []interface{}{}
				for _, device := range node.Devices {
					devices = append(devices, map[string]interface{}{
						"name": device,
					})
				}
				nodeSpec["devices"] = devices
			}
			if node.DeviceFilter != "" {
				nodeSpec["deviceFilter"] = node.DeviceFilter
			}
			nodes = append(nodes, nodeSpec)
		}
		storageSpec["nodes"] = nodes
	}

	return storageSpec
}

// buildResourceSpec builds a resource specification
func buildResourceSpec(resources *kubernetes.ContainerResources) map[string]interface{} {
	spec := map[string]interface{}{}

	if resources.Limits != nil {
		limits := map[string]interface{}{}
		if resources.Limits.Cpu != "" {
			limits["cpu"] = resources.Limits.Cpu
		}
		if resources.Limits.Memory != "" {
			limits["memory"] = resources.Limits.Memory
		}
		if len(limits) > 0 {
			spec["limits"] = limits
		}
	}

	if resources.Requests != nil {
		requests := map[string]interface{}{}
		if resources.Requests.Cpu != "" {
			requests["cpu"] = resources.Requests.Cpu
		}
		if resources.Requests.Memory != "" {
			requests["memory"] = resources.Requests.Memory
		}
		if len(requests) > 0 {
			spec["requests"] = requests
		}
	}

	return spec
}

// buildBlockPools builds the cephBlockPools section
func buildBlockPools(target *kubernetesrookcephclusterv1.KubernetesRookCephCluster) []interface{} {
	pools := []interface{}{}

	for _, bp := range target.Spec.BlockPools {
		pool := map[string]interface{}{
			"name": bp.Name,
			"spec": map[string]interface{}{
				"failureDomain": bp.GetFailureDomain(),
				"replicated": map[string]interface{}{
					"size": bp.GetReplicatedSize(),
				},
			},
		}

		// StorageClass configuration
		if bp.StorageClass != nil {
			pool["storageClass"] = buildStorageClassSpec(bp.StorageClass)
		}

		pools = append(pools, pool)
	}

	return pools
}

// buildFilesystems builds the cephFileSystems section
func buildFilesystems(target *kubernetesrookcephclusterv1.KubernetesRookCephCluster) []interface{} {
	filesystems := []interface{}{}

	for _, fs := range target.Spec.Filesystems {
		filesystem := map[string]interface{}{
			"name": fs.Name,
			"spec": map[string]interface{}{
				"metadataPool": map[string]interface{}{
					"replicated": map[string]interface{}{
						"size": fs.GetMetadataPoolReplicatedSize(),
					},
				},
				"dataPools": []interface{}{
					map[string]interface{}{
						"failureDomain": fs.GetFailureDomain(),
						"replicated": map[string]interface{}{
							"size": fs.GetDataPoolReplicatedSize(),
						},
						"name": "data0",
					},
				},
				"metadataServer": map[string]interface{}{
					"activeCount":   fs.GetActiveMdsCount(),
					"activeStandby": fs.GetActiveStandby(),
				},
			},
		}

		// MDS resources
		if fs.MdsResources != nil {
			fsSpec := filesystem["spec"].(map[string]interface{})
			mdsSpec := fsSpec["metadataServer"].(map[string]interface{})
			mdsSpec["resources"] = buildResourceSpec(fs.MdsResources)
		}

		// StorageClass configuration
		if fs.StorageClass != nil {
			filesystem["storageClass"] = buildStorageClassSpec(fs.StorageClass)
		}

		filesystems = append(filesystems, filesystem)
	}

	return filesystems
}

// buildObjectStores builds the cephObjectStores section
func buildObjectStores(target *kubernetesrookcephclusterv1.KubernetesRookCephCluster) []interface{} {
	stores := []interface{}{}

	for _, os := range target.Spec.ObjectStores {
		store := map[string]interface{}{
			"name": os.Name,
			"spec": map[string]interface{}{
				"metadataPool": map[string]interface{}{
					"failureDomain": os.GetFailureDomain(),
					"replicated": map[string]interface{}{
						"size": os.GetMetadataPoolReplicatedSize(),
					},
				},
				"dataPool": map[string]interface{}{
					"failureDomain": os.GetFailureDomain(),
					"erasureCoded": map[string]interface{}{
						"dataChunks":   os.GetDataPoolErasureDataChunks(),
						"codingChunks": os.GetDataPoolErasureCodingChunks(),
					},
				},
				"preservePoolsOnDelete": os.GetPreservePoolsOnDelete(),
				"gateway": map[string]interface{}{
					"port":      os.GetGatewayPort(),
					"instances": os.GetGatewayInstances(),
				},
			},
		}

		// Gateway resources
		if os.GatewayResources != nil {
			storeSpec := store["spec"].(map[string]interface{})
			gatewaySpec := storeSpec["gateway"].(map[string]interface{})
			gatewaySpec["resources"] = buildResourceSpec(os.GatewayResources)
		}

		// StorageClass configuration
		if os.StorageClass != nil {
			store["storageClass"] = buildStorageClassSpec(os.StorageClass)
		}

		stores = append(stores, store)
	}

	return stores
}

// buildStorageClassSpec builds the storageClass section
func buildStorageClassSpec(sc *kubernetesrookcephclusterv1.CephStorageClassSpec) map[string]interface{} {
	storageClass := map[string]interface{}{
		"enabled":              sc.GetEnabled(),
		"name":                 sc.Name,
		"isDefault":            sc.GetIsDefault(),
		"reclaimPolicy":        sc.GetReclaimPolicy(),
		"allowVolumeExpansion": sc.GetAllowVolumeExpansion(),
		"volumeBindingMode":    sc.GetVolumeBindingMode(),
	}

	return storageClass
}
