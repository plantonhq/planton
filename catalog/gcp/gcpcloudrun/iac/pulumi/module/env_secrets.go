package module

import (
	gcpcloudrunv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrun/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
)

// secretVariables lists the env entries whose value the module keeps in
// Secret Manager (the secret_value arm), addressed by container position.
func secretVariables(spec *gcpcloudrunv1alpha1.GcpCloudRunSpec) []envsecrets.Variable {
	var variables []envsecrets.Variable
	for containerIndex, container := range spec.Containers {
		for _, envVar := range container.Env {
			if envVar.SecretValue == "" {
				continue
			}
			variables = append(variables, envsecrets.Variable{
				ContainerIndex: containerIndex,
				Container:      container.Name,
				Name:           envVar.Name,
				Value:          envVar.SecretValue,
			})
		}
	}
	return variables
}

// secretPlacement puts the secrets where the service is: its project, the
// regions it serves from (every multi-region region, else its one region),
// and its runtime identity as the only reader.
func secretPlacement(locals *Locals) envsecrets.Placement {
	spec := locals.GcpCloudRun.Spec
	regions := []string{spec.Region}
	if spec.Region == "global" && spec.MultiRegionSettings != nil {
		regions = spec.MultiRegionSettings.Regions
	}
	return envsecrets.Placement{
		Kind:                  envsecrets.KindService,
		Resource:              locals.ServiceName,
		Location:              spec.Region,
		ReplicaRegions:        regions,
		Project:               spec.ProjectId.GetValue(),
		RuntimeServiceAccount: spec.ServiceAccount.GetValue(),
		Labels:                locals.GcpLabels,
	}
}
