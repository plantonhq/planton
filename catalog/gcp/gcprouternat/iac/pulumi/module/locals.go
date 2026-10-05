package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcprouternatv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcprouternat/v1alpha1"
)

// Locals collects frequently used input values (mirrors the Terraform
// "locals" pattern). Routers and NATs accept no labels in the GCP API, so
// there is no label set to derive for this kind.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpRouterNat      *gcprouternatv1alpha1.GcpRouterNat
}

// initializeLocals converts the iac-input into a struct that is easy to
// reference across the module.
func initializeLocals(iacInput *gcprouternatv1alpha1.GcpRouterNatIacInput) *Locals {
	return &Locals{
		GcpProviderConfig: iacInput.ProviderConfig,
		GcpRouterNat:      iacInput.Target,
	}
}
