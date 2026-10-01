package pulumimodule

import (
	"strings"
	"testing"
)

// Every Pulumi run resolves its module through GetPath, so a kind that does not run on Pulumi is
// refused there -- before a local directory is inspected, a binary downloaded or a module staged.
func TestGetPath_RefusesAKindThatDoesNotRunOnPulumi(t *testing.T) {
	_, err := GetPath(t.TempDir(), "org/project/stack", "OpenFgaStore", "", true)
	if err == nil {
		t.Fatal("OpenFGA kinds run on OpenTofu or Terraform only; GetPath must refuse Pulumi")
	}
	if !strings.Contains(err.Error(), "OpenFgaStore runs on OpenTofu or Terraform only, so Pulumi cannot deploy it") {
		t.Errorf("the refusal must name the kind and its engines, got: %v", err)
	}
}
