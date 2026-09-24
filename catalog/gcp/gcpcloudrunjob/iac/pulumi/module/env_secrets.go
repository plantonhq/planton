package module

import (
	gcpcloudrunjobv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudrunjob/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/cloudrunenv"
)

// secretVariables lists the env entries whose value the module keeps in
// Secret Manager (the secret_value arm), addressed by container position.
func secretVariables(tmpl *gcpcloudrunjobv1alpha1.GcpCloudRunJobTemplate) []cloudrunenv.Variable {
	var variables []cloudrunenv.Variable
	for containerIndex, container := range tmpl.GetContainers() {
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

// secretPlacement puts the secrets where the job is: its project, its one
// region, and its runtime identity as the only reader.
func secretPlacement(locals *Locals) cloudrunenv.Placement {
	spec := locals.GcpCloudRunJob.Spec
	return cloudrunenv.Placement{
		Kind:                  cloudrunenv.KindJob,
		Resource:              locals.JobName,
		Region:                spec.Region,
		ReplicaRegions:        []string{spec.Region},
		Project:               spec.ProjectId.GetValue(),
		RuntimeServiceAccount: spec.GetTemplate().GetServiceAccount().GetValue(),
		Labels:                locals.GcpLabels,
	}
}
