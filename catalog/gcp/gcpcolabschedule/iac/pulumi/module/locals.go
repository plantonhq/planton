package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcolabschedulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcolabschedule/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpColabSchedule  *gcpcolabschedulev1alpha1.GcpColabSchedule
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcolabschedulev1alpha1.GcpColabScheduleStackInput) *Locals {
	locals := &Locals{}
	locals.GcpColabSchedule = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
