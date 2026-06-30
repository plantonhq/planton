package module

import (
	"encoding/json"
	"fmt"

	kubernetesnamespacev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesnamespace/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds all derived configuration and state for the module
type Locals struct {
	// Context for Pulumi operations
	Ctx *pulumi.Context

	// Stack input containing the target resource
	StackInput *kubernetesnamespacev1.KubernetesNamespaceStackInput

	// Target namespace resource
	Target *kubernetesnamespacev1.KubernetesNamespace

	// Spec from the target
	Spec *kubernetesnamespacev1.KubernetesNamespaceSpec

	// Namespace name
	NamespaceName string

	// Combined labels (spec labels + standard labels)
	Labels map[string]string

	// Combined annotations (spec annotations + derived annotations)
	Annotations map[string]string

	// Resource quota configuration
	ResourceQuota *ResourceQuotaConfig

	// Limit range configuration
	LimitRange *LimitRangeConfig

	// Network policy configuration
	NetworkPolicy *NetworkPolicyConfig

	// Service mesh configuration
	ServiceMesh *ServiceMeshConfig

	// Pod security standard
	PodSecurityStandard string
}

// ResourceQuotaConfig holds computed resource quota values
type ResourceQuotaConfig struct {
	Enabled        bool
	CpuRequests    string
	CpuLimits      string
	MemoryRequests string
	MemoryLimits   string
	Pods           int32
	Services       int32
	ConfigMaps     int32
	Secrets        int32
	PVCs           int32
	LoadBalancers  int32
}

// LimitRangeConfig holds computed limit range values
type LimitRangeConfig struct {
	Enabled              bool
	DefaultCpuRequest    string
	DefaultCpuLimit      string
	DefaultMemoryRequest string
	DefaultMemoryLimit   string
}

// NetworkPolicyConfig holds network policy settings
type NetworkPolicyConfig struct {
	IsolateIngress           bool
	RestrictEgress           bool
	AllowedIngressNamespaces []string
	AllowedEgressCIDRs       []string
	AllowedEgressDomains     []string
}

// ServiceMeshConfig holds service mesh settings
type ServiceMeshConfig struct {
	Enabled     bool
	MeshType    string
	RevisionTag string
}

// initializeLocals creates and populates the Locals struct
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesnamespacev1.KubernetesNamespaceStackInput) *Locals {
	locals := &Locals{
		Ctx:           ctx,
		StackInput:    stackInput,
		Target:        stackInput.Target,
		Spec:          stackInput.Target.Spec,
		NamespaceName: stackInput.Target.Spec.Name,
	}

	// Build labels
	locals.Labels = buildLabels(locals)

	// Build annotations
	locals.Annotations = buildAnnotations(locals)

	// Compute resource quota configuration
	locals.ResourceQuota = computeResourceQuota(locals.Spec)

	// Compute limit range configuration
	locals.LimitRange = computeLimitRange(locals.Spec)

	// Extract network policy configuration
	locals.NetworkPolicy = extractNetworkPolicyConfig(locals.Spec)

	// Extract service mesh configuration
	locals.ServiceMesh = extractServiceMeshConfig(locals.Spec)

	// Determine pod security standard
	locals.PodSecurityStandard = getPodSecurityStandard(locals.Spec)

	return locals
}

// buildLabels combines spec labels with standard labels
func buildLabels(locals *Locals) map[string]string {
	labels := make(map[string]string)

	// Add standard labels
	labels["managed-by"] = "planton"
	labels["resource"] = locals.Target.Metadata.Name
	labels["resource-kind"] = "KubernetesNamespace"

	// Add spec labels
	for k, v := range locals.Spec.Labels {
		labels[k] = v
	}

	// Add pod security standard label if specified
	if locals.Spec.PodSecurityStandard != kubernetesnamespacev1.KubernetesNamespaceSpec_pod_security_standard_unspecified {
		pssLevel := getPodSecurityStandard(locals.Spec)
		if pssLevel != "" {
			labels["pod-security.kubernetes.io/enforce"] = pssLevel
		}
	}

	return labels
}

// buildAnnotations combines spec annotations with service mesh annotations
func buildAnnotations(locals *Locals) map[string]string {
	annotations := make(map[string]string)

	// Add spec annotations
	for k, v := range locals.Spec.Annotations {
		annotations[k] = v
	}

	// Add service mesh annotations
	if locals.Spec.ServiceMeshConfig != nil && locals.Spec.ServiceMeshConfig.Enabled {
		switch locals.Spec.ServiceMeshConfig.MeshType {
		case kubernetesnamespacev1.KubernetesNamespaceServiceMeshConfig_istio:
			// Istio sidecar injection
			if locals.Spec.ServiceMeshConfig.RevisionTag != "" {
				annotations["istio.io/rev"] = locals.Spec.ServiceMeshConfig.RevisionTag
			} else {
				annotations["istio-injection"] = "enabled"
			}
		case kubernetesnamespacev1.KubernetesNamespaceServiceMeshConfig_linkerd:
			// Linkerd sidecar injection
			annotations["linkerd.io/inject"] = "enabled"
		case kubernetesnamespacev1.KubernetesNamespaceServiceMeshConfig_consul:
			// Consul Connect injection
			annotations["consul.hashicorp.com/connect-inject"] = "true"
		}
	}

	return annotations
}

// computeResourceQuota determines resource quota values based on profile or custom settings
func computeResourceQuota(spec *kubernetesnamespacev1.KubernetesNamespaceSpec) *ResourceQuotaConfig {
	config := &ResourceQuotaConfig{Enabled: false}

	if spec.ResourceProfile == nil {
		return config
	}

	config.Enabled = true

	switch profileConfig := spec.ResourceProfile.ProfileConfig.(type) {
	case *kubernetesnamespacev1.KubernetesNamespaceResourceProfile_Preset:
		// Apply preset profile
		switch profileConfig.Preset {
		case kubernetesnamespacev1.KubernetesNamespaceResourceProfile_small:
			config.CpuRequests = "2"
			config.CpuLimits = "4"
			config.MemoryRequests = "4Gi"
			config.MemoryLimits = "8Gi"
			config.Pods = 20
			config.Services = 10
			config.ConfigMaps = 50
			config.Secrets = 50
			config.PVCs = 5
			config.LoadBalancers = 2

		case kubernetesnamespacev1.KubernetesNamespaceResourceProfile_medium:
			config.CpuRequests = "4"
			config.CpuLimits = "8"
			config.MemoryRequests = "8Gi"
			config.MemoryLimits = "16Gi"
			config.Pods = 50
			config.Services = 20
			config.ConfigMaps = 100
			config.Secrets = 100
			config.PVCs = 10
			config.LoadBalancers = 3

		case kubernetesnamespacev1.KubernetesNamespaceResourceProfile_large:
			config.CpuRequests = "8"
			config.CpuLimits = "16"
			config.MemoryRequests = "16Gi"
			config.MemoryLimits = "32Gi"
			config.Pods = 100
			config.Services = 40
			config.ConfigMaps = 200
			config.Secrets = 200
			config.PVCs = 20
			config.LoadBalancers = 5

		case kubernetesnamespacev1.KubernetesNamespaceResourceProfile_xlarge:
			config.CpuRequests = "16"
			config.CpuLimits = "32"
			config.MemoryRequests = "32Gi"
			config.MemoryLimits = "64Gi"
			config.Pods = 200
			config.Services = 80
			config.ConfigMaps = 400
			config.Secrets = 400
			config.PVCs = 40
			config.LoadBalancers = 10

		default:
			config.Enabled = false
		}

	case *kubernetesnamespacev1.KubernetesNamespaceResourceProfile_Custom:
		// Apply custom quotas
		custom := profileConfig.Custom
		if custom.Cpu != nil {
			config.CpuRequests = custom.Cpu.Requests
			config.CpuLimits = custom.Cpu.Limits
		}
		if custom.Memory != nil {
			config.MemoryRequests = custom.Memory.Requests
			config.MemoryLimits = custom.Memory.Limits
		}
		if custom.ObjectCounts != nil {
			config.Pods = custom.ObjectCounts.Pods
			config.Services = custom.ObjectCounts.Services
			config.ConfigMaps = custom.ObjectCounts.Configmaps
			config.Secrets = custom.ObjectCounts.Secrets
			config.PVCs = custom.ObjectCounts.PersistentVolumeClaims
			config.LoadBalancers = custom.ObjectCounts.LoadBalancers
		}
	}

	return config
}

// computeLimitRange determines default limit values
func computeLimitRange(spec *kubernetesnamespacev1.KubernetesNamespaceSpec) *LimitRangeConfig {
	config := &LimitRangeConfig{Enabled: false}

	if spec.ResourceProfile == nil {
		return config
	}

	if customProfile, ok := spec.ResourceProfile.ProfileConfig.(*kubernetesnamespacev1.KubernetesNamespaceResourceProfile_Custom); ok {
		if customProfile.Custom != nil && customProfile.Custom.DefaultLimits != nil {
			config.Enabled = true
			config.DefaultCpuRequest = customProfile.Custom.DefaultLimits.DefaultCpuRequest
			config.DefaultCpuLimit = customProfile.Custom.DefaultLimits.DefaultCpuLimit
			config.DefaultMemoryRequest = customProfile.Custom.DefaultLimits.DefaultMemoryRequest
			config.DefaultMemoryLimit = customProfile.Custom.DefaultLimits.DefaultMemoryLimit
		}
	}

	return config
}

// extractNetworkPolicyConfig extracts network policy settings
func extractNetworkPolicyConfig(spec *kubernetesnamespacev1.KubernetesNamespaceSpec) *NetworkPolicyConfig {
	config := &NetworkPolicyConfig{}

	if spec.NetworkConfig != nil {
		config.IsolateIngress = spec.NetworkConfig.IsolateIngress
		config.RestrictEgress = spec.NetworkConfig.RestrictEgress
		config.AllowedIngressNamespaces = spec.NetworkConfig.AllowedIngressNamespaces
		config.AllowedEgressCIDRs = spec.NetworkConfig.AllowedEgressCidrs
		config.AllowedEgressDomains = spec.NetworkConfig.AllowedEgressDomains
	}

	return config
}

// extractServiceMeshConfig extracts service mesh settings
func extractServiceMeshConfig(spec *kubernetesnamespacev1.KubernetesNamespaceSpec) *ServiceMeshConfig {
	config := &ServiceMeshConfig{}

	if spec.ServiceMeshConfig != nil {
		config.Enabled = spec.ServiceMeshConfig.Enabled
		config.RevisionTag = spec.ServiceMeshConfig.RevisionTag

		switch spec.ServiceMeshConfig.MeshType {
		case kubernetesnamespacev1.KubernetesNamespaceServiceMeshConfig_istio:
			config.MeshType = "istio"
		case kubernetesnamespacev1.KubernetesNamespaceServiceMeshConfig_linkerd:
			config.MeshType = "linkerd"
		case kubernetesnamespacev1.KubernetesNamespaceServiceMeshConfig_consul:
			config.MeshType = "consul"
		default:
			config.MeshType = ""
		}
	}

	return config
}

// getPodSecurityStandard returns the pod security standard level
func getPodSecurityStandard(spec *kubernetesnamespacev1.KubernetesNamespaceSpec) string {
	switch spec.PodSecurityStandard {
	case kubernetesnamespacev1.KubernetesNamespaceSpec_privileged:
		return "privileged"
	case kubernetesnamespacev1.KubernetesNamespaceSpec_baseline:
		return "baseline"
	case kubernetesnamespacev1.KubernetesNamespaceSpec_restricted:
		return "restricted"
	default:
		return ""
	}
}

// labelsToJSON converts labels map to JSON string
func labelsToJSON(labels map[string]string) string {
	if len(labels) == 0 {
		return "{}"
	}
	b, err := json.Marshal(labels)
	if err != nil {
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}

// annotationsToJSON converts annotations map to JSON string
func annotationsToJSON(annotations map[string]string) string {
	if len(annotations) == 0 {
		return "{}"
	}
	b, err := json.Marshal(annotations)
	if err != nil {
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}
