package module

// A secret an author puts in secret_env_vars is stored by the module; the workflow receives only the
// stored version's resource name under that variable, never the value, which every viewer of the
// workflow would read. Literal user_env_vars stay literals beside it.

import (
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	gcpworkflowv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpworkflow/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func workflowSpec() *gcpworkflowv1alpha1.GcpWorkflowSpec {
	return &gcpworkflowv1alpha1.GcpWorkflowSpec{
		Region:        "us-central1",
		UserEnvVars:   map[string]string{"MODE": "prod"},
		SecretEnvVars: map[string]string{"API_TOKEN": "tok_do_not_leak", "B_KEY": "b"},
	}
}

func TestAStoredValueReachesTheWorkflowAsAVersionName_neverTheValue(t *testing.T) {
	spec := workflowSpec()
	variables := secretVariables(spec)
	if len(variables) != 2 || variables[0].Name != "API_TOKEN" || variables[1].Name != "B_KEY" {
		t.Fatalf("every secret_env_vars entry is stored, in key order; got %+v", variables)
	}
	name := pulumi.String("projects/123/secrets/workflow_us-central1_orders_API_TOKEN/versions/1").ToStringOutput()
	refs := map[envsecrets.Key]envsecrets.Ref{{Name: "API_TOKEN"}: {Name: name}, {Name: "B_KEY"}: {Name: name}}

	vars := userEnvVars(spec, refs)
	if len(vars) != 3 {
		t.Fatalf("the workflow receives the literals and one pointer per secret; got %d", len(vars))
	}
	if vars["API_TOKEN"] != name {
		t.Errorf("the secret variable must carry the version's resource name; got %v", vars["API_TOKEN"])
	}
	if vars["MODE"] != pulumi.String("prod") {
		t.Errorf("a literal stays a literal; got %v", vars["MODE"])
	}
}

func TestThePlacementGrantsTheWorkflowsEmailIdentity(t *testing.T) {
	spec := workflowSpec()
	spec.ServiceAccount = &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/acme/serviceAccounts/wf@acme.iam.gserviceaccount.com"}}
	placement := secretPlacement(&Locals{GcpWorkflow: &gcpworkflowv1alpha1.GcpWorkflow{Spec: spec}, WorkflowName: "orders"})
	if got := envsecrets.SecretID(placement, secretVariables(spec)[0]); got != "workflow_us-central1_orders_API_TOKEN" {
		t.Errorf("secret id = %q", got)
	}
	if placement.RuntimeServiceAccount != "wf@acme.iam.gserviceaccount.com" {
		t.Errorf("the grant names the account's email; got %q", placement.RuntimeServiceAccount)
	}
}
