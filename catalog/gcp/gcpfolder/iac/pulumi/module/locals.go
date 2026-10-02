package module

import (
	"strings"

	gcpfolderv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpfolder/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the two values the folder resource is built from. A folder
// carries no labels, so there is no platform-label merge here.
type Locals struct {
	GcpFolder *gcpfolderv1alpha1.GcpFolder

	// The console display name: the spec's display_name, or metadata.name
	// when the spec leaves it empty -- the same naming basis every kind
	// uses.
	DisplayName string

	// The provider's `parent` string, rendered from whichever arm the spec
	// set: `organizations/{id}` for a top-level folder, `folders/{id}` for a
	// nested one. Exactly one arm is set (proto-CEL-enforced).
	Parent string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpfolderv1alpha1.GcpFolderStackInput) *Locals {
	target := stackInput.Target

	displayName := target.Spec.DisplayName
	if displayName == "" {
		displayName = target.Metadata.Name
	}

	return &Locals{
		GcpFolder:   target,
		DisplayName: displayName,
		Parent:      renderParent(target.Spec.Parent),
	}
}

// renderParent assembles Google's parent resource name from the spec's
// parent arm. A folder_id literal or reference is a bare numeric id (the
// GcpFolder folder_id output), so the `folders/` prefix is added here; a
// value that already carries it is passed through so a hand-written full
// name still works.
func renderParent(parent *gcpfolderv1alpha1.GcpFolderParent) string {
	if parent.GetOrganizationId() != "" {
		return "organizations/" + parent.GetOrganizationId()
	}
	folder := parent.GetFolderId().GetValue()
	if strings.HasPrefix(folder, "folders/") {
		return folder
	}
	return "folders/" + folder
}
