package module

import (
	"strings"

	gcpbillingbudgetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbillingbudget/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derivations both engines share.
type Locals struct {
	GcpBillingBudget *gcpbillingbudgetv1alpha1.GcpBillingBudget

	// BillingAccount is the bare account ID: the spec accepts the ID or the
	// billingAccounts/ resource name, and the provider wants the ID.
	BillingAccount string

	// DisplayName is the spec's display_name, or metadata.name when the
	// spec leaves it empty — the same naming basis every kind uses.
	DisplayName string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpbillingbudgetv1alpha1.GcpBillingBudgetStackInput) *Locals {
	target := stackInput.Target

	displayName := target.Spec.DisplayName
	if displayName == "" {
		displayName = target.Metadata.Name
	}

	return &Locals{
		GcpBillingBudget: target,
		BillingAccount:   strings.TrimPrefix(target.Spec.BillingAccount, "billingAccounts/"),
		DisplayName:      displayName,
	}
}
