package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpserverlessvpcconnectorv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpserverlessvpcconnector/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig         *gcpprovider.GcpProviderConfig
	GcpServerlessVpcConnector *gcpserverlessvpcconnectorv1alpha1.GcpServerlessVpcConnector
	ConnectorName             string
}

// NOTE ON LABELS: google_vpc_access_connector has no labels surface at all
// (verified against the released provider schema), so this module — unlike
// nearly every other GCP kind — attaches no platform attribution labels.
// Both engines skip labels identically; attribution rides on the connector
// name and the Planton control plane's own records.
func initializeLocals(_ *pulumi.Context, iacInput *gcpserverlessvpcconnectorv1alpha1.GcpServerlessVpcConnectorIacInput) *Locals {
	locals := &Locals{}
	locals.GcpServerlessVpcConnector = iacInput.Target

	// Connector name defaults to metadata.name. GCP caps connector names at
	// 25 characters — shorter than most resource names — so a metadata.name
	// that exceeds the cap fails at the API, not silently truncated here.
	locals.ConnectorName = locals.GcpServerlessVpcConnector.Spec.ConnectorName
	if locals.ConnectorName == "" {
		locals.ConnectorName = locals.GcpServerlessVpcConnector.Metadata.Name
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
