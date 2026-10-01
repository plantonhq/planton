package module

// A secret an author puts in software_config.secret_env_variables is stored by the module; Airflow
// receives only the stored version's resource name under that variable, never the value, which
// every viewer of the environment would read. Literal env_variables stay literals beside it.

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	gcpcloudcomposerenvironmentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudcomposerenvironment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func composerSpec() *gcpcloudcomposerenvironmentv1alpha1.GcpCloudComposerEnvironmentSpec {
	return &gcpcloudcomposerenvironmentv1alpha1.GcpCloudComposerEnvironmentSpec{
		Region: "us-central1",
		SoftwareConfig: &gcpcloudcomposerenvironmentv1alpha1.GcpCloudComposerSoftwareConfig{
			EnvVariables:       map[string]string{"MODE": "prod"},
			SecretEnvVariables: map[string]string{"WAREHOUSE_PASSWORD": "pw_do_not_leak"},
		},
		NodeConfig: &gcpcloudcomposerenvironmentv1alpha1.GcpCloudComposerNodeConfig{
			ServiceAccount: &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/acme/serviceAccounts/airflow@acme.iam.gserviceaccount.com"}},
		},
	}
}

func TestAStoredValueReachesAirflowAsAVersionName_neverTheValue(t *testing.T) {
	spec := composerSpec()
	variables := secretVariables(spec)
	if len(variables) != 1 || variables[0].Name != "WAREHOUSE_PASSWORD" {
		t.Fatalf("every secret_env_variables entry is stored; got %+v", variables)
	}
	name := pulumi.String("projects/123/secrets/composer_us-central1_etl_WAREHOUSE_PASSWORD/versions/1").ToStringOutput()
	vars := envVariables(spec.SoftwareConfig, map[envsecrets.Key]envsecrets.Ref{{Name: "WAREHOUSE_PASSWORD"}: {Name: name}})
	if len(vars) != 2 || vars["WAREHOUSE_PASSWORD"] != name || vars["MODE"] != pulumi.String("prod") {
		t.Errorf("Airflow receives the literal and the version's resource name; got %v", vars)
	}
}

func TestThePlacementGrantsTheNodeServiceAccountsEmail(t *testing.T) {
	spec := composerSpec()
	placement := secretPlacement(&Locals{GcpCloudComposerEnvironment: &gcpcloudcomposerenvironmentv1alpha1.GcpCloudComposerEnvironment{Spec: spec}}, "etl")
	if got := envsecrets.SecretID(placement, secretVariables(spec)[0]); got != "composer_us-central1_etl_WAREHOUSE_PASSWORD" {
		t.Errorf("secret id = %q", got)
	}
	if placement.RuntimeServiceAccount != "airflow@acme.iam.gserviceaccount.com" {
		t.Errorf("the grant names the node account's email; got %q", placement.RuntimeServiceAccount)
	}
}
