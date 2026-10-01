package module

// A secret an author puts in containers[].secret_environment is stored by the module in Secrets
// Manager, and the container definition names it under "secrets" by its pinned ARN; the value
// itself never enters the registered task definition, where every viewer of the task definition
// reads it. A literal environment entry beside it stays a literal.

import (
	"encoding/json"
	"strings"
	"testing"

	awsecstaskdefinitionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecstaskdefinition/v1alpha1"
)

func TestASecretEnvironmentEntryIsNamedByItsArn_neverWrittenIntoTheTaskDefinition(t *testing.T) {
	const secretValue = "sk_live_do_not_leak"
	const pinnedArn = "arn:aws:secretsmanager:us-east-1:123456789012:secret:billing/app/STRIPE_KEY-AbCdEf:::v-1"
	spec := &awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinitionSpec{
		Containers: []*awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinitionContainer{{
			Name:              "app",
			Image:             "123456789012.dkr.ecr.us-east-1.amazonaws.com/billing:1",
			Environment:       map[string]string{"MODE": "prod"},
			SecretEnvironment: map[string]string{"STRIPE_KEY": secretValue},
		}},
	}

	document, err := buildContainerDefinitions(spec, "/ecs/billing", map[string]map[string]string{
		"app": {"STRIPE_KEY": pinnedArn},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(document, secretValue) {
		t.Fatalf("the secret's value is written into the task definition: %s", document)
	}

	var definitions []struct {
		Environment []struct{ Name, Value string } `json:"environment"`
		Secrets     []struct {
			Name      string `json:"name"`
			ValueFrom string `json:"valueFrom"`
		} `json:"secrets"`
	}
	if err := json.Unmarshal([]byte(document), &definitions); err != nil {
		t.Fatal(err)
	}
	app := definitions[0]
	if len(app.Secrets) != 1 || app.Secrets[0].Name != "STRIPE_KEY" || app.Secrets[0].ValueFrom != pinnedArn {
		t.Errorf("the secret must be named under secrets by its pinned ARN; got %+v", app.Secrets)
	}
	if len(app.Environment) != 1 || app.Environment[0].Name != "MODE" || app.Environment[0].Value != "prod" {
		t.Errorf("only the literal is plain environment; got %+v", app.Environment)
	}
}
