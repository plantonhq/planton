package runner

import (
	"strings"
	"testing"
)

func TestRequireLaneEngine_RefusesAnEngineTheKindDoesNotRunOn(t *testing.T) {
	t.Setenv("PLANTON_E2E_TF_BINARY", "")
	if err := requireLaneEngine("stripewebhookendpoint", "terraform"); err != nil {
		t.Errorf("an OpenTofu-only kind runs on the default tofu binary, got %v", err)
	}
	if err := requireLaneEngine("stripewebhookendpoint", "pulumi"); err == nil || !strings.Contains(err.Error(), "runs on OpenTofu only") {
		t.Errorf("Pulumi must be refused for an OpenTofu-only kind, got %v", err)
	}

	t.Setenv("PLANTON_E2E_TF_BINARY", "terraform")
	if err := requireLaneEngine("stripewebhookendpoint", "terraform"); err == nil || !strings.Contains(err.Error(), "Terraform cannot deploy it") {
		t.Errorf("the Terraform binary must be refused for an OpenTofu-only kind, got %v", err)
	}
	if err := requireLaneEngine("openfgastore", "terraform"); err != nil {
		t.Errorf("OpenFGA declares Terraform, got %v", err)
	}
}

func TestRequireLaneEngine_UndeclaredKindRunsAnywhere(t *testing.T) {
	t.Setenv("PLANTON_E2E_TF_BINARY", "terraform")
	for _, engine := range []string{"pulumi", "terraform"} {
		if err := requireLaneEngine("awss3bucket", engine); err != nil {
			t.Errorf("%s: %v", engine, err)
		}
	}
}
