package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudcomposeruserworkloadsconfigmapv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudcomposeruserworkloadsconfigmap/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig                      *gcpprovider.GcpProviderConfig
	GcpCloudComposerUserWorkloadsConfigMap *gcpcloudcomposeruserworkloadsconfigmapv1alpha1.GcpCloudComposerUserWorkloadsConfigMap
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudcomposeruserworkloadsconfigmapv1alpha1.GcpCloudComposerUserWorkloadsConfigMapIacInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudComposerUserWorkloadsConfigMap = iacInput.Target

	// Kubernetes ConfigMaps carry no GCP labels surface — no platform
	// attribution labels are stamped, identically on both engines.

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
