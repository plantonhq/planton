package module

import (
	gcpservicenetworkingconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpservicenetworkingconnection/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpServiceNetworkingConnection *gcpservicenetworkingconnectionv1alpha1.GcpServiceNetworkingConnection

	// Empty service falls through to the Google managed-services producer —
	// the one behind Cloud SQL, AlloyDB, Memorystore, and Filestore private
	// IP.
	Service string
}

func initializeLocals(ctx *pulumi.Context, iacInput *gcpservicenetworkingconnectionv1alpha1.GcpServiceNetworkingConnectionIacInput) *Locals {
	target := iacInput.Target

	service := target.Spec.Service
	if service == "" {
		service = "servicenetworking.googleapis.com"
	}

	return &Locals{
		GcpServiceNetworkingConnection: target,
		Service:                        service,
	}
}
