package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudtasksqueuev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudtasksqueue/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpCloudTasksQueue *gcpcloudtasksqueuev1alpha1.GcpCloudTasksQueue
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudtasksqueuev1alpha1.GcpCloudTasksQueueIacInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudTasksQueue = iacInput.Target
	locals.GcpProviderConfig = iacInput.ProviderConfig
	// Note: Cloud Tasks queues do NOT support GCP labels.
	// No label computation needed (unlike most GCP components).
	return locals
}
