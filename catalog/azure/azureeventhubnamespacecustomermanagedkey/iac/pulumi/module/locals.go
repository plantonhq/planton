package module

import (
	azureeventhubnamespacecustomermanagedkeyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventhubnamespacecustomermanagedkey/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureEventHubNamespaceCustomerManagedKey *azureeventhubnamespacecustomermanagedkeyv1alpha1.AzureEventHubNamespaceCustomerManagedKey
	EventhubNamespaceId                      string
}

func initializeLocals(ctx *pulumi.Context, iacInput *azureeventhubnamespacecustomermanagedkeyv1alpha1.AzureEventHubNamespaceCustomerManagedKeyIacInput) *Locals {
	locals := &Locals{}

	locals.AzureEventHubNamespaceCustomerManagedKey = iacInput.Target

	// The eventhub_namespace_id field is a StringValueOrRef. The platform
	// middleware resolves valueFrom references before IaC modules run, so
	// .GetValue() always returns the resolved literal ARM id.
	locals.EventhubNamespaceId = iacInput.Target.Spec.EventhubNamespaceId.GetValue()

	// The CMK configuration carries no Azure tags: it is a property of the
	// namespace, not an ARM object of its own, so the platform's identity
	// tags live on the namespace (and its cluster).

	return locals
}
