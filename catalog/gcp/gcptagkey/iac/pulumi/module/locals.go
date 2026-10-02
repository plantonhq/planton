package module

import (
	"strings"

	gcptagkeyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcptagkey/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the two values the key resource is built from. A tag key
// carries no labels, so there is no platform-label merge here.
type Locals struct {
	GcpTagKey *gcptagkeyv1alpha1.GcpTagKey

	// The key's short name: the spec's short_name, or metadata.name when
	// the spec leaves it empty -- the same naming basis every kind uses.
	ShortName string

	// The provider's `parent` rendered from whichever owner arm the spec
	// set: `organizations/{id}` or `projects/{id}`. Exactly one arm is set
	// (proto-CEL-enforced).
	Parent string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcptagkeyv1alpha1.GcpTagKeyStackInput) *Locals {
	target := stackInput.Target

	shortName := target.Spec.ShortName
	if shortName == "" {
		shortName = target.Metadata.Name
	}

	return &Locals{
		GcpTagKey: target,
		ShortName: shortName,
		Parent:    renderParent(target.Spec.Parent),
	}
}

// renderParent assembles Google's parent resource name from the spec's
// owner arm. A project_id literal or reference is a bare id (or number),
// so the `projects/` prefix is added here; a value that already carries it
// is passed through so a hand-written full name still works.
func renderParent(parent *gcptagkeyv1alpha1.GcpTagKeyParent) string {
	if parent.GetOrganizationId() != "" {
		return "organizations/" + parent.GetOrganizationId()
	}
	project := parent.GetProjectId().GetValue()
	if strings.HasPrefix(project, "projects/") {
		return project
	}
	return "projects/" + project
}
