package module

import (
	gcpcloudfunctionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudfunction/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
)

// secretVariables lists the secret_environment_variables entries whose value
// the module keeps in Secret Manager (the value arm), addressed by key.
func secretVariables(spec *gcpcloudfunctionv1alpha1.GcpCloudFunctionSpec) []envsecrets.Variable {
	var variables []envsecrets.Variable
	for _, envVar := range spec.GetServiceConfig().GetSecretEnvironmentVariables() {
		if envVar.Value == "" {
			continue
		}
		variables = append(variables, envsecrets.Variable{Name: envVar.Key, Value: envVar.Value})
	}
	return variables
}

// secretPlacement puts the secrets where the function is: its project, its
// region, and its runtime identity (or the Compute Engine default service
// account it falls back to) as the only reader.
func secretPlacement(locals *Locals) envsecrets.Placement {
	spec := locals.GcpCloudFunction.Spec
	return envsecrets.Placement{
		Kind:                  envsecrets.KindFunction,
		Resource:              locals.FunctionName,
		Location:              spec.Region,
		NoContainers:          true,
		ReplicaRegions:        []string{spec.Region},
		Project:               spec.ProjectId.GetValue(),
		RuntimeServiceAccount: spec.GetServiceConfig().GetServiceAccountEmail().GetValue(),
		Labels:                locals.GcpLabels,
	}
}
