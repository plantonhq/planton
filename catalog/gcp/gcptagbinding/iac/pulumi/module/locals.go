package module

import (
	"strings"

	gcptagbindingv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcptagbinding/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// resourceManagerPrefix is the full-resource-name authority for the
// hierarchy nodes (organizations, folders, projects).
const resourceManagerPrefix = "//cloudresourcemanager.googleapis.com/"

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derived values the binding is built from. A binding
// carries no labels, so there is no platform-label merge here.
type Locals struct {
	GcpTagBinding *gcptagbindingv1alpha1.GcpTagBinding

	// The full resource name of the tagged resource when it is knowable
	// offline: a folder, an organization, a project given by NUMBER, or a
	// literal resource_name. Empty when the project must be looked up at
	// apply time (the project arm is an ID, or the parent is empty and the
	// provider's default project applies) -- see tagBinding.
	Parent string

	// The project arm's literal when it is not a number (a project ID to
	// resolve), or "" for the ambient project. Only meaningful when Parent
	// is empty.
	ProjectIdToResolve string

	// Whether the location-scoped binding resource is used (spec.location
	// set) instead of the global one.
	IsLocationScoped bool
}

func initializeLocals(_ *pulumi.Context, stackInput *gcptagbindingv1alpha1.GcpTagBindingStackInput) *Locals {
	target := stackInput.Target
	parent := target.Spec.Parent

	locals := &Locals{
		GcpTagBinding:    target,
		IsLocationScoped: target.Spec.Location != "",
	}

	switch {
	case parent.GetFolderId().GetValue() != "":
		locals.Parent = resourceManagerPrefix + "folders/" + strings.TrimPrefix(parent.GetFolderId().GetValue(), "folders/")
	case parent.GetOrganizationId() != "":
		locals.Parent = resourceManagerPrefix + "organizations/" + parent.GetOrganizationId()
	case parent.GetResourceName() != "":
		locals.Parent = parent.GetResourceName()
	default:
		// The project arm (or no arm at all). Google wants the project
		// NUMBER in a binding's parent: a numeric literal (or a GcpProject
		// reference, which resolves to the number) is rendered here; a
		// project ID, or the ambient project, is resolved at apply time.
		project := strings.TrimPrefix(parent.GetProjectId().GetValue(), "projects/")
		if isNumeric(project) {
			locals.Parent = resourceManagerPrefix + "projects/" + project
		} else {
			locals.ProjectIdToResolve = project
		}
	}

	return locals
}

// isNumeric reports whether s is a non-empty string of ASCII digits -- the
// shape of a project number.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
