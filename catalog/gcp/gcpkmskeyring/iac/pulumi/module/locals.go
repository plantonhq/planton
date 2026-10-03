package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpkmskeyringv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpkmskeyring/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpKmsKeyRing     *gcpkmskeyringv1alpha1.GcpKmsKeyRing
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpkmskeyringv1alpha1.GcpKmsKeyRingIacInput) *Locals {
	locals := &Locals{}
	locals.GcpKmsKeyRing = iacInput.Target

	// The key ring resource has no labels surface in the Cloud KMS API —
	// no platform attribution labels are computed or stamped, identically
	// on both engines. (Labels live on the crypto keys inside the ring.)

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
