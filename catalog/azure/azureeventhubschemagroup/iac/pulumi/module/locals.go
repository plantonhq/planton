package module

import (
	azureeventhubschemagroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubschemagroup/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureEventHubSchemaGroup *azureeventhubschemagroupv1alpha1.AzureEventHubSchemaGroup
	NamespaceId              string
}

// Enum wire maps to ARM's values. Both enums are required in the spec
// (unspecified is rejected at validation), so no unspecified fallback row
// is needed -- an unmapped value would send the empty string and fail
// loudly at the provider, which is the right outcome.
var schemaCompatibilityStrings = map[azureeventhubschemagroupv1alpha1.AzureEventHubSchemaCompatibility]string{
	azureeventhubschemagroupv1alpha1.AzureEventHubSchemaCompatibility_NONE:     "None",
	azureeventhubschemagroupv1alpha1.AzureEventHubSchemaCompatibility_BACKWARD: "Backward",
	azureeventhubschemagroupv1alpha1.AzureEventHubSchemaCompatibility_FORWARD:  "Forward",
}

var schemaTypeStrings = map[azureeventhubschemagroupv1alpha1.AzureEventHubSchemaType]string{
	azureeventhubschemagroupv1alpha1.AzureEventHubSchemaType_AVRO: "Avro",
	azureeventhubschemagroupv1alpha1.AzureEventHubSchemaType_JSON: "Json",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureeventhubschemagroupv1alpha1.AzureEventHubSchemaGroupIacInput) *Locals {
	locals := &Locals{}

	locals.AzureEventHubSchemaGroup = iacInput.Target
	locals.NamespaceId = iacInput.Target.Spec.NamespaceId.GetValue()

	// Schema groups carry no Azure tags: ARM does not support tags on
	// Event Hubs entities, so the platform's identity tags live on the
	// parent namespace.

	return locals
}
