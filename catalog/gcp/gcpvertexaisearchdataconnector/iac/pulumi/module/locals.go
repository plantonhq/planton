package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaisearchdataconnectorv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaisearchdataconnector/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig              *gcpprovider.GcpProviderConfig
	GcpVertexAiSearchDataConnector *gcpvertexaisearchdataconnectorv1alpha1.GcpVertexAiSearchDataConnector

	// CollectionId is the collection's GCP id: spec.collection_id when
	// set, otherwise metadata.name -- the same fallback the Terraform
	// module applies in locals.tf.
	CollectionId string

	// CollectionDisplayName is spec.collection_display_name when set,
	// otherwise metadata.name; Google requires one.
	CollectionDisplayName string
}

// initializeLocals derives the two defaulted names. Discovery Engine
// resources carry no labels, so there is no attribution label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpvertexaisearchdataconnectorv1alpha1.GcpVertexAiSearchDataConnectorStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiSearchDataConnector = stackInput.Target
	spec := locals.GcpVertexAiSearchDataConnector.Spec

	locals.CollectionId = spec.CollectionId
	if locals.CollectionId == "" {
		locals.CollectionId = locals.GcpVertexAiSearchDataConnector.Metadata.Name
	}
	locals.CollectionDisplayName = spec.CollectionDisplayName
	if locals.CollectionDisplayName == "" {
		locals.CollectionDisplayName = locals.GcpVertexAiSearchDataConnector.Metadata.Name
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
