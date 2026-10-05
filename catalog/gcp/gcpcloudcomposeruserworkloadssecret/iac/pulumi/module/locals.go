package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudcomposeruserworkloadssecretv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudcomposeruserworkloadssecret/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig                   *gcpprovider.GcpProviderConfig
	GcpCloudComposerUserWorkloadsSecret *gcpcloudcomposeruserworkloadssecretv1alpha1.GcpCloudComposerUserWorkloadsSecret
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudcomposeruserworkloadssecretv1alpha1.GcpCloudComposerUserWorkloadsSecretIacInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudComposerUserWorkloadsSecret = iacInput.Target

	// Kubernetes Secrets carry no GCP labels surface — no platform
	// attribution labels are stamped, identically on both engines.

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
