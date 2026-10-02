package module

import (
	"regexp"
	"sort"

	gcpcloudcomposerenvironmentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudcomposerenvironment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serviceAccountPath is the projects/{project}/serviceAccounts/ prefix the
// node service account may carry; the grant names the email after it.
var serviceAccountPath = regexp.MustCompile(`^projects/[^/]+/serviceAccounts/`)

// secretVariables lists the software_config.secret_env_variables entries the
// module keeps in Secret Manager, in key order so the plan is stable.
func secretVariables(spec *gcpcloudcomposerenvironmentv1alpha1.GcpCloudComposerEnvironmentSpec) []envsecrets.Variable {
	secrets := spec.GetSoftwareConfig().GetSecretEnvVariables()
	names := make([]string, 0, len(secrets))
	for name := range secrets {
		names = append(names, name)
	}
	sort.Strings(names)
	variables := make([]envsecrets.Variable, 0, len(names))
	for _, name := range names {
		variables = append(variables, envsecrets.Variable{Name: name, Value: secrets[name]})
	}
	return variables
}

// secretPlacement puts the secrets where the environment is: its project,
// its region, and its node service account (or the Compute Engine default
// service account a Composer 2 environment falls back to) as the only
// reader -- the identity Airflow's workers and schedulers run as.
func secretPlacement(locals *Locals, environmentName string) envsecrets.Placement {
	spec := locals.GcpCloudComposerEnvironment.Spec
	return envsecrets.Placement{
		Kind:                  envsecrets.KindComposer,
		Resource:              environmentName,
		Location:              spec.Region,
		NoContainers:          true,
		ReplicaRegions:        []string{spec.Region},
		Project:               spec.ProjectId.GetValue(),
		RuntimeServiceAccount: serviceAccountPath.ReplaceAllString(spec.GetNodeConfig().GetServiceAccount().GetValue(), ""),
		Labels:                locals.GcpLabels,
	}
}

// envVariables is what Airflow receives as its environment: the literal
// env_variables, plus each stored secret's version resource name under its
// variable -- a pointer DAG code resolves with the Secret Manager client,
// never the value.
func envVariables(sc *gcpcloudcomposerenvironmentv1alpha1.GcpCloudComposerSoftwareConfig, secretRefs map[envsecrets.Key]envsecrets.Ref) pulumi.StringMap {
	vars := pulumi.StringMap{}
	for name, value := range sc.EnvVariables {
		vars[name] = pulumi.String(value)
	}
	for name := range sc.SecretEnvVariables {
		vars[name] = secretRefs[envsecrets.Key{Name: name}].Name
	}
	return vars
}
