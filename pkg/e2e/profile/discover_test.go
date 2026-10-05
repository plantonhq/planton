package profile

import (
	"strings"
	"testing"

	kindv1 "github.com/plantonhq/planton/qa/catalogkinde2eprofile/v1"
	sharedpb "github.com/plantonhq/planton/shared"
)

// A profile may only claim engines its kind runs on; otherwise the matrix schedules a lane the CLI
// refuses before it starts.
func TestCheckValidatedProvisioners_RefusesAnEngineTheKindDoesNotRunOn(t *testing.T) {
	profile := func(provisioners ...sharedpb.IacProvisioner) *kindv1.CatalogKindE2EProfile {
		return &kindv1.CatalogKindE2EProfile{Spec: &kindv1.CatalogKindE2EProfileSpec{ValidatedProvisioners: provisioners}}
	}

	err := checkValidatedProvisioners("openfgastore", profile(sharedpb.IacProvisioner_terraform, sharedpb.IacProvisioner_pulumi))
	if err == nil || !strings.Contains(err.Error(), "validated_provisioners lists pulumi, but OpenFgaStore does not run on it") {
		t.Fatalf("want the refusal, got %v", err)
	}
	if err := checkValidatedProvisioners("openfgastore", profile(sharedpb.IacProvisioner_tofu, sharedpb.IacProvisioner_terraform)); err != nil {
		t.Errorf("declared engines pass: %v", err)
	}
	if err := checkValidatedProvisioners("awss3bucket", profile(sharedpb.IacProvisioner_pulumi)); err != nil {
		t.Errorf("an undeclared kind accepts every engine: %v", err)
	}
}

func TestToPascalCase_RegistryLookup(t *testing.T) {
	tests := []struct {
		kindDir string
		want    string
	}{
		{kindDir: "awslambda", want: "AwsLambda"},
		{kindDir: "awskmskey", want: "AwsKmsKey"},
		{kindDir: "awslambdaeventsourcemapping", want: "AwsLambdaEventSourceMapping"},
		{kindDir: "awssecuritygroup", want: "AwsSecurityGroup"},
		// Kubernetes kinds resolve through the registry too; Signoz's enum
		// name matches its TestKubernetesSignoz_* entrypoints (the retired
		// hand-maintained table had it wrong as "KubernetesSigNoz").
		{kindDir: "kubernetessignoz", want: "KubernetesSignoz"},
		{kindDir: "kubernetesgatewayapicrds", want: "KubernetesGatewayApiCrds"},
		// ArgoCD's test entrypoints deviate from the enum name
		// (KubernetesArgocd); the explicit override keeps its green CI lane
		// matching TestKubernetesArgoCD_*.
		{kindDir: "kubernetesargocd", want: "KubernetesArgoCD"},
	}
	for _, tc := range tests {
		t.Run(tc.kindDir, func(t *testing.T) {
			if got := toPascalCase(tc.kindDir); got != tc.want {
				t.Errorf("toPascalCase(%q) = %q, want %q", tc.kindDir, got, tc.want)
			}
		})
	}
}
