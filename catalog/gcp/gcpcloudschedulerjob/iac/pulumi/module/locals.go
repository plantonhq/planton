package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudschedulerjobv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudschedulerjob/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpCloudSchedulerJob *gcpcloudschedulerjobv1alpha1.GcpCloudSchedulerJob
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudschedulerjobv1alpha1.GcpCloudSchedulerJobIacInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudSchedulerJob = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig
	// Note: Cloud Scheduler jobs do NOT support GCP labels.
	// No label computation needed (unlike most GCP kinds).
	return locals
}
