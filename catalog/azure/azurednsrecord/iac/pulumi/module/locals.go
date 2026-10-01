package module

import (
	"strings"

	azurednsrecordv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurednsrecord/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDnsRecord    *azurednsrecordv1alpha1.AzureDnsRecord
	ResourceGroupName string
	ZoneName          string
	AzureTags         map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurednsrecordv1alpha1.AzureDnsRecordStackInput) *Locals {
	locals := &Locals{}

	locals.AzureDnsRecord = stackInput.Target

	target := stackInput.Target

	// resource_group and zone_name are StringValueOrRef fields. The
	// platform middleware resolves valueFrom references before IaC modules
	// run, so .GetValue() always returns the resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.ZoneName = target.Spec.ZoneName.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide. Record-set tags land in ARM's record-set metadata.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureDnsRecord.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
