package tofumodule

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	openfgastorev1 "github.com/plantonhq/planton/catalog/openfga/openfgastore/v1alpha1"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/iac/terraform"
)

// Every OpenTofu and Terraform run -- the CLI's commands and the platform runner alike -- starts in
// Init, which holds the chosen binary to the kind's declared engines before it writes a backend file
// or var file. OpenFGA runs on both HCL engines, so both pass; the refusal is proven with an engine
// the kind does not declare.
func TestInit_RefusesAnEngineTheKindDoesNotRunOn_BeforeWritingAnything(t *testing.T) {
	modulePath := t.TempDir()
	manifest := &openfgastorev1.OpenFgaStore{Kind: "OpenFgaStore", Metadata: &shared.CloudResourceMetadata{Name: "store"}}

	err := Init(context.Background(), "pulumi", modulePath, manifest, terraform.TerraformBackendType_local, nil, nil, false, false, nil)
	if err == nil || !strings.Contains(err.Error(), "OpenFgaStore runs on OpenTofu or Terraform only, so Pulumi cannot deploy it") {
		t.Fatalf("want the kind's refusal, got %v", err)
	}
	entries, _ := os.ReadDir(modulePath)
	if len(entries) != 0 {
		t.Errorf("a refused run must write nothing, found %d entries in the module path", len(entries))
	}
	if _, err := os.Stat(filepath.Join(modulePath, ".terraform")); err == nil {
		t.Error("a refused run must not create .terraform")
	}
}
