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

	// QuotaProjectID is the project the Budget API's calls bill their quota
	// to; empty lets the deploying identity's own project carry it.
	QuotaProjectID string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpbillingbudgetv1alpha1.GcpBillingBudgetIacInput) *Locals {
	target := iacInput.Target

	displayName := target.Spec.DisplayName
	if displayName == "" {
		displayName = target.Metadata.Name
	}

	return &Locals{
		GcpBillingBudget: target,
		BillingAccount:   strings.TrimPrefix(target.Spec.BillingAccount, "billingAccounts/"),
		DisplayName:      displayName,
		QuotaProjectID:   target.Spec.GetQuotaProjectId().GetValue(),
	}
}
