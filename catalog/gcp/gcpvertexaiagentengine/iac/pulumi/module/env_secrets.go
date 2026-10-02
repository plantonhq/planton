package module

import (
	gcpvertexaiagentenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaiagentengine/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
)

// secretVariables lists the deployment_spec.secret_env entries whose value
// the module keeps in Secret Manager (the value arm), addressed by name.
func secretVariables(spec *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSpec) []envsecrets.Variable {
	var variables []envsecrets.Variable
	for _, envVar := range spec.GetAgent().GetDeploymentSpec().GetSecretEnv() {
		value, stored := envVar.Source.(*gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSecretEnvVar_Value)
		if !stored {
			continue
		}
		variables = append(variables, envsecrets.Variable{Name: envVar.Name, Value: value.Value})
	}
	return variables
}

// secretPlacement puts the secrets where the agent is: its project, its
// location, and its identity -- spec.service_account, or the project's
// Reasoning Engine service agent it falls back to -- as the only reader.
// The id carries metadata.name because Google assigns the agent's own id.
func secretPlacement(locals *Locals) envsecrets.Placement {
	spec := locals.GcpVertexAiAgentEngine.Spec
	return envsecrets.Placement{
		Kind:                  envsecrets.KindAgentEngine,
		Resource:              locals.GcpVertexAiAgentEngine.Metadata.Name,
		Location:              spec.Location,
		NoContainers:          true,
		ReplicaRegions:        []string{spec.Location},
		Project:               spec.ProjectId.GetValue(),
		RuntimeServiceAccount: spec.GetAgent().GetServiceAccount().GetValue(),
		DefaultAccount:        envsecrets.ReasoningEngineServiceAgent,
		Labels:                locals.GcpLabels,
	}
}
