package module

// A secret an author puts in deployment_spec.secret_env[].value is stored by the module and read by
// the agent through a Secret Manager reference; it never becomes a plain value on the agent, where
// every viewer reads it. An entry naming a secret the author owns stays exactly that reference.

import (
	"testing"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	gcpvertexaiagentenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaiagentengine/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/envsecrets"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func agentSpec() *gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSpec {
	return &gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSpec{
		Location: "us-central1",
		Agent: &gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineAgent{
			DeploymentSpec: &gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineDeploymentSpec{
				SecretEnv: []*gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSecretEnvVar{
					{Name: "OPENAI_API_KEY", Source: &gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSecretEnvVar_Value{Value: "sk-do-not-leak"}},
					{Name: "DB_PASSWORD", Source: &gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSecretEnvVar_SecretRef{
						SecretRef: &gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngineSecretRef{
							Secret:  &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "db-password"}},
							Version: "3",
						},
					}},
				},
			},
		},
	}
}

func TestAStoredValueBecomesASecretManagerReference_neverAPlainValue(t *testing.T) {
	spec := agentSpec()
	variables := secretVariables(spec)
	if len(variables) != 1 || variables[0].Name != "OPENAI_API_KEY" {
		t.Fatalf("exactly the value entry is stored; got %+v", variables)
	}
	stored := envsecrets.Ref{
		Secret:  pulumi.String("agentengine_us-central1_helper_OPENAI_API_KEY").ToStringOutput(),
		Version: pulumi.String("1").ToStringOutput(),
	}
	refs := map[envsecrets.Key]envsecrets.Ref{{Name: "OPENAI_API_KEY"}: stored}

	secrets := buildDeploymentSpec(spec.Agent.DeploymentSpec, refs).SecretEnvs.(vertex.AiReasoningEngineSpecDeploymentSpecSecretEnvArray)
	if len(secrets) != 2 {
		t.Fatalf("both entries reach the agent; got %d", len(secrets))
	}
	first := secrets[0].(*vertex.AiReasoningEngineSpecDeploymentSpecSecretEnvArgs).SecretRef.(*vertex.AiReasoningEngineSpecDeploymentSpecSecretEnvSecretRefArgs)
	if first.Secret != stored.Secret || first.Version != stored.Version {
		t.Errorf("the stored entry must read the module's own secret and version; got %v %v", first.Secret, first.Version)
	}
	second := secrets[1].(*vertex.AiReasoningEngineSpecDeploymentSpecSecretEnvArgs).SecretRef.(*vertex.AiReasoningEngineSpecDeploymentSpecSecretEnvSecretRefArgs)
	if second.Secret != pulumi.String("db-password") || second.Version != pulumi.String("3") {
		t.Errorf("an owned secret stays the author's reference; got %v %v", second.Secret, second.Version)
	}
}

func TestThePlacementGrantsTheAgentsIdentity(t *testing.T) {
	spec := agentSpec()
	resource := &gcpvertexaiagentenginev1alpha1.GcpVertexAiAgentEngine{Metadata: &shared.CloudResourceMetadata{Name: "helper"}, Spec: spec}
	placement := secretPlacement(&Locals{GcpVertexAiAgentEngine: resource})
	if got := envsecrets.SecretID(placement, secretVariables(spec)[0]); got != "agentengine_us-central1_helper_OPENAI_API_KEY" {
		t.Errorf("secret id = %q", got)
	}
	if got := envsecrets.RuntimeMember(placement, "42"); got != "serviceAccount:service-42@gcp-sa-aiplatform-re.iam.gserviceaccount.com" {
		t.Errorf("an agent without a service account grants %q", got)
	}
	spec.Agent.ServiceAccount = &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "agent@acme.iam.gserviceaccount.com"}}
	if got := envsecrets.RuntimeMember(secretPlacement(&Locals{GcpVertexAiAgentEngine: resource}), "42"); got != "serviceAccount:agent@acme.iam.gserviceaccount.com" {
		t.Errorf("an agent with a service account grants %q", got)
	}
}
