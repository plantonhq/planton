package module

import (
	"fmt"

	gcpgkeworkloadidentitybindingv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkeworkloadidentitybinding/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals keeps the derived values both engines must construct identically.
type Locals struct {
	GcpGkeWorkloadIdentityBinding *gcpgkeworkloadidentitybindingv1alpha1.GcpGkeWorkloadIdentityBinding

	// PoolProject is the project naming the implicit workload-identity pool
	// (<project>.svc.id.goog). Empty when the manifest omits it — the
	// provider's default project is then resolved at deploy time.
	PoolProject string

	// ServiceAccountId is the fully-qualified service-account resource name
	// the provider requires (not the bare email). The "-" project wildcard
	// lets the IAM API infer the SA's own project from the email — correct
	// even when the GSA lives in a different project than the
	// workload-identity pool. Identical construction in the Terraform module.
	ServiceAccountId string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpgkeworkloadidentitybindingv1alpha1.GcpGkeWorkloadIdentityBindingIacInput) *Locals {
	locals := &Locals{}

	locals.GcpGkeWorkloadIdentityBinding = iacInput.Target
	spec := iacInput.Target.Spec

	locals.PoolProject = spec.ProjectId.GetValue()

	locals.ServiceAccountId = fmt.Sprintf(
		"projects/-/serviceAccounts/%s",
		spec.ServiceAccountEmail.GetValue(),
	)

	return locals
}
