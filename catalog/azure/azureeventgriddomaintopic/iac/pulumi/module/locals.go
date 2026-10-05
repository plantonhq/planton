package module

import (
	"strings"

	azureeventgriddomaintopicv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureeventgriddomaintopic/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureEventgridDomainTopic *azureeventgriddomaintopicv1alpha1.AzureEventgridDomainTopic

	// ResourceGroupName and DomainName are parsed from the spec's
	// domain_id (a StringValueOrRef; the platform middleware resolves
	// valueFrom references before IaC modules run). The provider
	// addresses domain topics by (resource group, domain name) while
	// the spec takes the domain's ARM id -- the same ARM object either
	// way, so the id is split into its segments here.
	ResourceGroupName string
	DomainName        string
}

// No tags: the provider carries no tags argument on domain topics
// (they are addressing entries under the domain).

func initializeLocals(ctx *pulumi.Context, iacInput *azureeventgriddomaintopicv1alpha1.AzureEventgridDomainTopicIacInput) *Locals {
	locals := &Locals{}

	locals.AzureEventgridDomainTopic = iacInput.Target
	target := iacInput.Target

	// The domain id's shape is /subscriptions/{sub}/resourceGroups/{rg}
	// /providers/Microsoft.EventGrid/domains/{domain}. Segment names are
	// matched case-insensitively (ARM treats them that way), so ids
	// composed by hand survive the split.
	segments := strings.Split(target.Spec.DomainId.GetValue(), "/")
	for i := 0; i+1 < len(segments); i++ {
		switch {
		case strings.EqualFold(segments[i], "resourceGroups"):
			locals.ResourceGroupName = segments[i+1]
		case strings.EqualFold(segments[i], "domains"):
			locals.DomainName = segments[i+1]
		}
	}

	return locals
}
