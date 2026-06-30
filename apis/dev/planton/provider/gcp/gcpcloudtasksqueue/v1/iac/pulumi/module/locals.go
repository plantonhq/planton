package module

import (
	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpcloudtasksqueuev1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudtasksqueue/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpCloudTasksQueue *gcpcloudtasksqueuev1.GcpCloudTasksQueue
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcloudtasksqueuev1.GcpCloudTasksQueueStackInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudTasksQueue = stackInput.Target
	locals.GcpProviderConfig = stackInput.ProviderConfig
	// Note: Cloud Tasks queues do NOT support GCP labels.
	// No label computation needed (unlike most GCP components).
	return locals
}
