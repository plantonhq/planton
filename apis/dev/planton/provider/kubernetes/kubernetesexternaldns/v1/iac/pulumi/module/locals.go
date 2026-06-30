package module

import (
	"fmt"

	kubernetesexternaldnsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesexternaldns/v1"
)

// Locals holds all computed values for the ExternalDNS module
type Locals struct {
	// KubernetesExternalDns is the target resource
	KubernetesExternalDns *kubernetesexternaldnsv1.KubernetesExternalDns

	// Namespace where resources will be deployed
	Namespace string

	// HelmChartVersion for the ExternalDNS chart
	HelmChartVersion string

	// ExternalDnsVersion for the ExternalDNS image tag
	ExternalDnsVersion string

	// ReleaseName is the Helm release name (same as metadata.name)
	ReleaseName string

	// KsaName is the Kubernetes ServiceAccount name
	KsaName string

	// ProviderType indicates which DNS provider is configured (google, aws, azure, cloudflare)
	ProviderType string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	CloudflareApiTokenSecretName string

	// GKE-specific configuration
	GkeProjectId string
	GkeDnsZoneId string
	GkeGsaEmail  string
	IsGke        bool

	// EKS-specific configuration
	EksRoute53ZoneId string
	EksIrsaRoleArn   string
	IsEks            bool

	// AKS-specific configuration
	AksDnsZoneId               string
	AksManagedIdentityClientId string
	IsAks                      bool

	// Cloudflare-specific configuration
	CfApiToken   string
	CfDnsZoneId  string
	CfIsProxied  bool
	IsCloudflare bool
}

// initializeLocals creates and populates the Locals struct
func initializeLocals(stackInput *kubernetesexternaldnsv1.KubernetesExternalDnsStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	locals := &Locals{
		KubernetesExternalDns: target,
	}

	// Namespace with default
	locals.Namespace = getNamespace(spec)

	// Versions with defaults
	locals.HelmChartVersion = getHelmChartVersion(spec)
	locals.ExternalDnsVersion = getExternalDnsVersion(spec)

	// Release name and ServiceAccount name (both use metadata.name)
	locals.ReleaseName = target.Metadata.Name
	locals.KsaName = target.Metadata.Name

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	locals.CloudflareApiTokenSecretName = fmt.Sprintf("%s-cloudflare-api-token", target.Metadata.Name)

	// Determine provider type and set provider-specific locals
	locals.IsGke = spec.GetGke() != nil
	locals.IsEks = spec.GetEks() != nil
	locals.IsAks = spec.GetAks() != nil
	locals.IsCloudflare = spec.GetCloudflare() != nil

	switch {
	case locals.IsGke:
		locals.ProviderType = "google"
		gke := spec.GetGke()
		locals.GkeProjectId = gke.ProjectId.GetValue()
		locals.GkeDnsZoneId = gke.DnsZoneId.GetValue()
		locals.GkeGsaEmail = fmt.Sprintf("%s@%s.iam.gserviceaccount.com", locals.KsaName, locals.GkeProjectId)

	case locals.IsEks:
		locals.ProviderType = "aws"
		eks := spec.GetEks()
		locals.EksRoute53ZoneId = eks.Route53ZoneId.GetValue()
		locals.EksIrsaRoleArn = eks.IrsaRoleArnOverride

	case locals.IsAks:
		locals.ProviderType = "azure"
		aks := spec.GetAks()
		locals.AksDnsZoneId = aks.DnsZoneId.GetValue()
		locals.AksManagedIdentityClientId = aks.ManagedIdentityClientId

	case locals.IsCloudflare:
		locals.ProviderType = "cloudflare"
		cf := spec.GetCloudflare()
		locals.CfApiToken = cf.ApiToken
		locals.CfDnsZoneId = cf.DnsZoneId.GetValue()
		locals.CfIsProxied = cf.IsProxied

	default:
		locals.ProviderType = "unknown"
	}

	return locals
}

// getNamespace returns the namespace from spec, with default if not specified
func getNamespace(spec *kubernetesexternaldnsv1.KubernetesExternalDnsSpec) string {
	namespace := spec.Namespace.GetValue()
	if namespace == "" {
		return "external-dns" // default
	}
	return namespace
}

// getHelmChartVersion returns the Helm chart version from spec, with default if not specified
func getHelmChartVersion(spec *kubernetesexternaldnsv1.KubernetesExternalDnsSpec) string {
	if spec.HelmChartVersion != nil && *spec.HelmChartVersion != "" {
		return *spec.HelmChartVersion
	}
	return "1.19.0" // default
}

// getExternalDnsVersion returns the ExternalDNS version from spec, with default if not specified
func getExternalDnsVersion(spec *kubernetesexternaldnsv1.KubernetesExternalDnsSpec) string {
	if spec.ExternalDnsVersion != nil && *spec.ExternalDnsVersion != "" {
		return *spec.ExternalDnsVersion
	}
	return "v0.19.0" // default
}
