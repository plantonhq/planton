package provisioner

import (
	"strings"
	"testing"

	awss3bucketv1 "github.com/plantonhq/planton/catalog/aws/awss3bucket/v1alpha1"
	openfgastorev1 "github.com/plantonhq/planton/catalog/openfga/openfgastore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/provisionerannotationkeys"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
)

func openFgaStore(provisionerLabel string) *openfgastorev1.OpenFgaStore {
	m := &openfgastorev1.OpenFgaStore{Kind: "OpenFgaStore", Metadata: &shared.CatalogObjectMetadata{Name: "store"}}
	if provisionerLabel != "" {
		m.Metadata.Annotations = map[string]string{provisionerannotationkeys.ProvisionerAnnotationKey: provisionerLabel}
	}
	return m
}

func TestRequire_RefusesAnEngineTheKindDoesNotRunOn(t *testing.T) {
	err := Require(catalogkind.CatalogKind_OpenFgaStore, ProvisionerTypePulumi)
	if err == nil {
		t.Fatal("Pulumi on an OpenTofu-and-Terraform kind must be refused")
	}
	want := "OpenFgaStore runs on OpenTofu or Terraform only, so Pulumi cannot deploy it: set planton.dev/provisioner to tofu or terraform, or leave it unset"
	if err.Error() != want {
		t.Errorf("refusal reads\n  %s\nwant\n  %s", err, want)
	}
	for _, p := range []ProvisionerType{ProvisionerTypeTofu, ProvisionerTypeTerraform} {
		if err := Require(catalogkind.CatalogKind_OpenFgaStore, p); err != nil {
			t.Errorf("%s is declared and must pass: %v", p, err)
		}
	}
}

func TestRequire_AnUndeclaredKindRunsOnEveryEngine(t *testing.T) {
	for _, p := range everyEngine {
		if err := Require(catalogkind.CatalogKind_AwsS3Bucket, p); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
	if err := RequireForKindName("NotARegisteredKind", ProvisionerTypePulumi); err != nil {
		t.Errorf("an unresolvable kind name is not narrowed: %v", err)
	}
}

func TestForManifest_RefusesALabelTheKindDoesNotRunOn(t *testing.T) {
	_, err := ForManifest(openFgaStore("pulumi"))
	if err == nil || !strings.Contains(err.Error(), "Pulumi cannot deploy it") {
		t.Fatalf("want the refusal, got %v", err)
	}
	got, err := ForManifest(openFgaStore("terraform"))
	if err != nil || got != ProvisionerTypeTerraform {
		t.Errorf("a declared label passes through: got %v, %v", got, err)
	}
}

func TestForManifest_UnlabeledManifestAsksAmongTheKindsEngines(t *testing.T) {
	got, err := ForManifest(openFgaStore(""))
	if err != nil || got != ProvisionerTypeUnspecified {
		t.Fatalf("two declared engines leave the choice to the caller: got %v, %v", got, err)
	}
	allowed, err := AllowedForManifest(openFgaStore(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(allowed) != 2 || allowed[0] != ProvisionerTypeTofu || allowed[1] != ProvisionerTypeTerraform {
		t.Errorf("the prompt offers the declared engines in order, got %v", allowed)
	}

	s3 := &awss3bucketv1.AwsS3Bucket{Kind: "AwsS3Bucket", Metadata: &shared.CatalogObjectMetadata{Name: "b"}}
	if got, _ := ForManifest(s3); got != ProvisionerTypeUnspecified {
		t.Errorf("an undeclared kind still asks, got %v", got)
	}
}

func TestModuleFamily_HclEnginesShareOneModule(t *testing.T) {
	for p, want := range map[ProvisionerType]shared.IacProvisioner{
		ProvisionerTypePulumi:    shared.IacProvisioner_pulumi,
		ProvisionerTypeTofu:      shared.IacProvisioner_terraform,
		ProvisionerTypeTerraform: shared.IacProvisioner_terraform,
	} {
		if got := p.ModuleFamily(); got != want {
			t.Errorf("%s.ModuleFamily() = %s, want %s", p, got, want)
		}
	}
}
