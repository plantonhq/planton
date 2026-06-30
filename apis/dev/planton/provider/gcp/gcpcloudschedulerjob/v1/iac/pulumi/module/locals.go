package module

import (
	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpcloudschedulerjobv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudschedulerjob/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpCloudSchedulerJob *gcpcloudschedulerjobv1.GcpCloudSchedulerJob
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcloudschedulerjobv1.GcpCloudSchedulerJobStackInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudSchedulerJob = stackInput.Target
	locals.GcpProviderConfig = stackInput.ProviderConfig
	// Note: Cloud Scheduler jobs do NOT support GCP labels.
	// No label computation needed (unlike most GCP components).
	return locals
}
