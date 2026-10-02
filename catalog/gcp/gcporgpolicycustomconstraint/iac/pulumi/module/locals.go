package module

import (
	gcporgpolicycustomconstraintv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcporgpolicycustomconstraint/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// customConstraintPrefix is the namespace Google gives every
// organization-defined constraint. The spec takes the bare name; the
// module owns the prefix so it can never be doubled or forgotten.
const customConstraintPrefix = "custom."

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the two names the constraint resource is built from. A
// constraint carries no labels, so there is no platform-label merge here.
type Locals struct {
	GcpOrgPolicyCustomConstraint *gcporgpolicycustomconstraintv1alpha1.GcpOrgPolicyCustomConstraint

	// The constraint's name as Google knows it: `custom.<constraint_name>`,
	// with constraint_name defaulting to metadata.name when the spec leaves
	// it empty -- the same naming basis every kind uses.
	ConstraintName string

	// The provider's `parent`: `organizations/{organization_id}`.
	Parent string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcporgpolicycustomconstraintv1alpha1.GcpOrgPolicyCustomConstraintStackInput) *Locals {
	target := stackInput.Target

	name := target.Spec.ConstraintName
	if name == "" {
		name = target.Metadata.Name
	}

	return &Locals{
		GcpOrgPolicyCustomConstraint: target,
		ConstraintName:               customConstraintPrefix + name,
		Parent:                       "organizations/" + target.Spec.OrganizationId,
	}
}
