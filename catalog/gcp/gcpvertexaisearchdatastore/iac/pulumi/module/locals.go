package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvertexaisearchdatastorev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaisearchdatastore/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig          *gcpprovider.GcpProviderConfig
	GcpVertexAiSearchDataStore *gcpvertexaisearchdatastorev1alpha1.GcpVertexAiSearchDataStore

	// DataStoreId is the store's GCP id: spec.data_store_id when set,
	// otherwise metadata.name -- the same fallback the Terraform module
	// applies in locals.tf.
	DataStoreId string

	// DisplayName is spec.display_name when set, otherwise metadata.name;
	// Google requires a display name, so the module always sends one.
	DisplayName string
}

// initializeLocals derives the two defaulted names. Discovery Engine
// resources carry no labels, so there is no attribution label set.
func initializeLocals(_ *pulumi.Context, stackInput *gcpvertexaisearchdatastorev1alpha1.GcpVertexAiSearchDataStoreStackInput) *Locals {
	locals := &Locals{}
	locals.GcpVertexAiSearchDataStore = stackInput.Target

	locals.DataStoreId = locals.GcpVertexAiSearchDataStore.Spec.DataStoreId
	if locals.DataStoreId == "" {
		locals.DataStoreId = locals.GcpVertexAiSearchDataStore.Metadata.Name
	}
	locals.DisplayName = locals.GcpVertexAiSearchDataStore.Spec.DisplayName
	if locals.DisplayName == "" {
		locals.DisplayName = locals.GcpVertexAiSearchDataStore.Metadata.Name
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
