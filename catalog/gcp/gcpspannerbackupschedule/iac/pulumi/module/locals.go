package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpspannerbackupschedulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpspannerbackupschedule/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds resolved values used across the Pulumi module.
// Note: Spanner backup schedules do not support GCP labels. Labels are
// managed at the instance level only (see GcpSpannerInstance).
type Locals struct {
	GcpProviderConfig        *gcpprovider.GcpProviderConfig
	GcpSpannerBackupSchedule *gcpspannerbackupschedulev1alpha1.GcpSpannerBackupSchedule
	ScheduleName             string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpspannerbackupschedulev1alpha1.GcpSpannerBackupScheduleIacInput) *Locals {
	locals := &Locals{}
	locals.GcpSpannerBackupSchedule = iacInput.Target

	locals.ScheduleName = locals.GcpSpannerBackupSchedule.Spec.ScheduleName
	if locals.ScheduleName == "" {
		locals.ScheduleName = locals.GcpSpannerBackupSchedule.Metadata.Name
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
