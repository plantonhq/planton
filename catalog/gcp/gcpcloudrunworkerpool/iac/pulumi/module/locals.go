package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudrunworkerpoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrunworkerpool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig     *gcpprovider.GcpProviderConfig
	GcpCloudRunWorkerPool *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPool
	GcpLabels             map[string]string

	// WorkerPoolName is the pool's GCP name: spec.worker_pool_name when
	// set, otherwise metadata.name -- the same fallback the Terraform
	// module applies in locals.tf.
	WorkerPoolName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolStackInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudRunWorkerPool = stackInput.Target

	locals.WorkerPoolName = locals.GcpCloudRunWorkerPool.Spec.WorkerPoolName
	if locals.WorkerPoolName == "" {
		locals.WorkerPoolName = locals.GcpCloudRunWorkerPool.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpCloudRunWorkerPool.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.WorkerPoolName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpCloudRunWorkerPool.String())

	if locals.GcpCloudRunWorkerPool.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpCloudRunWorkerPool.Metadata.Org
	}
	if locals.GcpCloudRunWorkerPool.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpCloudRunWorkerPool.Metadata.Env
	}
	if locals.GcpCloudRunWorkerPool.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpCloudRunWorkerPool.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
