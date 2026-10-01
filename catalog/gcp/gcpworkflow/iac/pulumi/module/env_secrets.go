package module

import (
	"regexp"
	"sort"

	gcpworkflowv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpworkflow/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serviceAccountPath is the projects/{project}/serviceAccounts/ prefix the
// workflow's service_account may carry; the grant names the email after it.
var serviceAccountPath = regexp.MustCompile(`^projects/[^/]+/serviceAccounts/`)

// secretVariables lists the secret_env_vars entries the module keeps in
// Secret Manager, in key order so the plan is stable.
func secretVariables(spec *gcpworkflowv1alpha1.GcpWorkflowSpec) []envsecrets.Variable {
	names := make([]string, 0, len(spec.SecretEnvVars))
	for name := range spec.SecretEnvVars {
		names = append(names, name)
	}
	sort.Strings(names)
	variables := make([]envsecrets.Variable, 0, len(names))
	for _, name := range names {
		variables = append(variables, envsecrets.Variable{Name: name, Value: spec.SecretEnvVars[name]})
	}
	return variables
}

// secretPlacement puts the secrets where the workflow is: its project, its
// region, and its runtime identity (or the Compute Engine default service
// account it falls back to) as the only reader.
func secretPlacement(locals *Locals) envsecrets.Placement {
	spec := locals.GcpWorkflow.Spec
	return envsecrets.Placement{
		Kind:                  envsecrets.KindWorkflow,
		Resource:              locals.WorkflowName,
		Location:              spec.Region,
		NoContainers:          true,
		ReplicaRegions:        []string{spec.Region},
		Project:               spec.ProjectId.GetValue(),
		RuntimeServiceAccount: serviceAccountPath.ReplaceAllString(spec.ServiceAccount.GetValue(), ""),
		Labels:                locals.GcpLabels,
	}
}

// userEnvVars is what the workflow receives as its environment: the literal
// user_env_vars, plus each stored secret's version resource name under its
// variable -- a pointer the workflow resolves through the Secret Manager
// connector, never the value.
func userEnvVars(spec *gcpworkflowv1alpha1.GcpWorkflowSpec, secretRefs map[envsecrets.Key]envsecrets.Ref) pulumi.StringMap {
	vars := pulumi.StringMap{}
	for name, value := range spec.UserEnvVars {
		vars[name] = pulumi.String(value)
	}
	for name := range spec.SecretEnvVars {
		vars[name] = secretRefs[envsecrets.Key{Name: name}].Name
	}
	return vars
}
