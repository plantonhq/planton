package module

import (
	gcpiamdenypolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpiamdenypolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpIamDenyPolicy *gcpiamdenypolicyv1alpha1.GcpIamDenyPolicy

	// The cloud-side policy name defaults to metadata.name when the spec
	// leaves policy_name empty — the same naming basis every kind uses.
	PolicyName string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpiamdenypolicyv1alpha1.GcpIamDenyPolicyIacInput) *Locals {
	target := iacInput.Target

	policyName := target.Spec.PolicyName
	if policyName == "" {
		policyName = target.Metadata.Name
	}

	return &Locals{
		GcpIamDenyPolicy: target,
		PolicyName:       policyName,
	}
}
