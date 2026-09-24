package module

import (
	gcpcloudrunv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrun/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv"
)

// secretVariables lists the env entries whose value the module keeps in
// Secret Manager (the secret_value arm), addressed by container position.
func secretVariables(spec *gcpcloudrunv1alpha1.GcpCloudRunSpec) []cloudrunenv.Variable {
	var variables []cloudrunenv.Variable
	for containerIndex, container := range spec.Containers {
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

// secretPlacement puts the secrets where the service is: its project, the
// regions it serves from (every multi-region region, else its one region),
// and its runtime identity as the only reader.
func secretPlacement(locals *Locals) cloudrunenv.Placement {
	spec := locals.GcpCloudRun.Spec
	regions := []string{spec.Region}
	if spec.Region == "global" && spec.MultiRegionSettings != nil {
		regions = spec.MultiRegionSettings.Regions
	}
	return cloudrunenv.Placement{
		Kind:                  cloudrunenv.KindService,
		Resource:              locals.ServiceName,
		Region:                spec.Region,
		ReplicaRegions:        regions,
		Project:               spec.ProjectId.GetValue(),
		RuntimeServiceAccount: spec.ServiceAccount.GetValue(),
		Labels:                locals.GcpLabels,
	}
}
