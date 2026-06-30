package module

import (
	kubernetesgatewayapicrdsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgatewayapicrds/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values used throughout the module
type Locals struct {
	// Version of Gateway API to install
	Version string

	// Whether to use experimental channel
	IsExperimental bool

	// Channel name (for outputs)
	ChannelName string

	// URL to download the CRD manifests
	ManifestURL string

	// Resource name for the CRDs
	ResourceName string

	// Labels for resources
	Labels map[string]string
}

// initializeLocals computes values from the stack input
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesgatewayapicrdsv1.KubernetesGatewayApiCrdsStackInput) *Locals {
	spec := stackInput.Target.Spec
	metadata := stackInput.Target.Metadata

	// Determine version
	version := DefaultVersion
	if spec.Version != nil && *spec.Version != "" {
		version = *spec.Version
	}

	// Determine channel
	isExperimental := false
	channelName := "standard"
	if spec.InstallChannel != nil {
		if spec.InstallChannel.Channel == kubernetesgatewayapicrdsv1.KubernetesGatewayApiCrdsSpec_InstallChannel_experimental {
			isExperimental = true
			channelName = "experimental"
		}
	}

	// Resource name based on metadata name
	resourceName := metadata.Name + "-gateway-api-crds"

	// Standard labels
	labels := map[string]string{
		"app.kubernetes.io/name":       "gateway-api-crds",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "planton",
		"app.kubernetes.io/component":  "crds",
		"gateway-api/version":          version,
		"gateway-api/channel":          channelName,
	}

	return &Locals{
		Version:        version,
		IsExperimental: isExperimental,
		ChannelName:    channelName,
		ManifestURL:    GetManifestURL(version, isExperimental),
		ResourceName:   resourceName,
		Labels:         labels,
	}
}
