package module

import (
	"strings"

	azureprivatednsrecordv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednsrecord/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzurePrivateDnsRecord *azureprivatednsrecordv1alpha1.AzurePrivateDnsRecord

	// ResourceGroupName and ZoneName are parsed from the spec's
	// private_dns_zone_id (a StringValueOrRef; the platform middleware
	// resolves valueFrom references before IaC modules run). The
	// Terraform provider addresses private record sets by the zone's
	// ARM id directly; this SDK addresses them by (resource group, zone
	// name) -- the same ARM object either way, so the id is split into
	// its segments here.
	ResourceGroupName string
	ZoneName          string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order. ARM stores these as record-set
	// METADATA (record sets carry no ARM tags proper).
	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureprivatednsrecordv1alpha1.AzurePrivateDnsRecordStackInput) *Locals {
	locals := &Locals{}

	locals.AzurePrivateDnsRecord = stackInput.Target
	target := stackInput.Target

	// The zone id's shape is /subscriptions/{sub}/resourceGroups/{rg}
	// /providers/Microsoft.Network/privateDnsZones/{zone}. Segment names
	// are matched case-insensitively (ARM treats them that way), so ids
	// composed by hand survive the split.
	segments := strings.Split(target.Spec.PrivateDnsZoneId.GetValue(), "/")
	for i := 0; i+1 < len(segments); i++ {
		switch {
		case strings.EqualFold(segments[i], "resourceGroups"):
			locals.ResourceGroupName = segments[i+1]
		case strings.EqualFold(segments[i], "privateDnsZones"):
			locals.ZoneName = segments[i+1]
		}
	}

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzurePrivateDnsRecord.String()),
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

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
