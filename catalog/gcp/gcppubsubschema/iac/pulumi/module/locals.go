package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcppubsubschemav1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppubsubschema/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpPubSubSchema   *gcppubsubschemav1alpha1.GcpPubSubSchema
}

func initializeLocals(_ *pulumi.Context, iacInput *gcppubsubschemav1alpha1.GcpPubSubSchemaIacInput) *Locals {
	locals := &Locals{}
	locals.GcpPubSubSchema = iacInput.Target

	// The schema resource has no labels surface in the Pub/Sub API — no
	// platform attribution labels are stamped, identically on both engines.

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
