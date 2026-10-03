package module

import (
	gcpmonitoringdashboardv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmonitoringdashboard/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs. The dashboard resource
// carries no labels argument (its whole configuration is the JSON body), so
// there is no platform-label merge here.
type Locals struct {
	GcpMonitoringDashboard *gcpmonitoringdashboardv1alpha1.GcpMonitoringDashboard
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpmonitoringdashboardv1alpha1.GcpMonitoringDashboardIacInput) *Locals {
	return &Locals{
		GcpMonitoringDashboard: iacInput.Target,
	}
}
