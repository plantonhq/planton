package module

import (
	gcpcloudrunworkerpoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrunworkerpool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv"
)

// secretVariables lists the env entries whose value the module keeps in
// Secret Manager (the secret_value arm), addressed by container position.
func secretVariables(spec *gcpcloudrunworkerpoolv1alpha1.GcpCloudRunWorkerPoolSpec) []cloudrunenv.Variable {
	var variables []cloudrunenv.Variable
	for containerIndex, container := range spec.GetContainers() {
		for _, envVar := range container.Env {
			if envVar.SecretValue == "" {
				continue
			}
			variables = append(variables, cloudrunenv.Variable{
				ContainerIndex: containerIndex,
				Container:      container.Name,
				Name:           envVar.Name,
				Value:          envVar.SecretValue,
			})
		}
	}
	return variables
}

// secretPlacement puts the secrets where the worker pool is: its project, its
// one region, and its runtime identity as the only reader.
func secretPlacement(locals *Locals) cloudrunenv.Placement {
	spec := locals.GcpCloudRunWorkerPool.Spec
	return cloudrunenv.Placement{
		Kind:                  cloudrunenv.KindWorkerPool,
		Resource:              locals.WorkerPoolName,
		Region:                spec.Region,
		ReplicaRegions:        []string{spec.Region},
		Project:               spec.ProjectId.GetValue(),
		RuntimeServiceAccount: spec.ServiceAccount.GetValue(),
		Labels:                locals.GcpLabels,
	}
}
