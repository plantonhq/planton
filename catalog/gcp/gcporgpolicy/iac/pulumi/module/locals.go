package module

import (
	"strings"

	"github.com/pkg/errors"
	gcporgpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcporgpolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derived values the policy resource is built from. A
// policy carries no labels, so there is no platform-label merge here.
type Locals struct {
	GcpOrgPolicy *gcporgpolicyv1alpha1.GcpOrgPolicy

	// The constraint the policy configures: the spec's literal
	// `constraint`, or the resolved `custom_constraint` reference (the
	// `custom.<name>` handle). Exactly one is set (proto-CEL-enforced).
	Constraint string

	// The provider's `parent` rendered from the scope arm the spec set --
	// `projects/{id}`, `folders/{id}`, or `organizations/{id}` -- or empty
	// when the scope is empty, in which case the resource function resolves
	// the provider's default project at apply time (the same fallback the
	// Terraform module expresses with a null project).
	Parent string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcporgpolicyv1alpha1.GcpOrgPolicyStackInput) (*Locals, error) {
	target := stackInput.Target

	constraint := target.Spec.Constraint
	if constraint == "" {
		constraint = target.Spec.GetCustomConstraint().GetValue()
	}
	if constraint == "" {
		return nil, errors.New("constraint is empty after reference resolution -- the custom_constraint reference did not resolve")
	}

	return &Locals{
		GcpOrgPolicy: target,
		Constraint:   constraint,
		Parent:       renderParent(target.Spec.Scope),
	}, nil
}

// renderParent assembles Google's parent resource name from the spec's
// scope arm. Literals and references are bare ids (a project id or number,
// a folder id, an organization id), so the prefix is added here; a value
// that already carries its prefix is passed through so a hand-written full
// name still works. An empty scope renders "" -- the caller resolves the
// provider's default project.
func renderParent(scope *gcporgpolicyv1alpha1.GcpOrgPolicyScope) string {
	switch {
	case scope.GetProjectId().GetValue() != "":
		return prefixed("projects/", scope.GetProjectId().GetValue())
	case scope.GetFolderId().GetValue() != "":
		return prefixed("folders/", scope.GetFolderId().GetValue())
	case scope.GetOrganizationId() != "":
		return "organizations/" + scope.GetOrganizationId()
	default:
		return ""
	}
}

// prefixed returns value with the prefix added unless it is already there.
func prefixed(prefix, value string) string {
	if strings.HasPrefix(value, prefix) {
		return value
	}
	return prefix + value
}
