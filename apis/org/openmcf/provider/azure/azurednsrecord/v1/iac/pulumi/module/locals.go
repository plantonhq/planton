package module

import (
	"strings"

	azurednsrecordv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurednsrecord/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureDnsRecord    *azurednsrecordv1.AzureDnsRecord
	AzureTags         map[string]string
	ZoneName          string
	ResourceGroupName string
	RecordName        string
	TTL            int
	MxPriority     int
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurednsrecordv1.AzureDnsRecordStackInput) *Locals {
	locals := &Locals{}

	locals.AzureDnsRecord = stackInput.Target
	target := stackInput.Target
	spec := target.Spec

	// Extract zone_name from literal or value_from
	if spec.ZoneName != nil {
		if spec.ZoneName.GetValue() != "" {
			locals.ZoneName = spec.ZoneName.GetValue()
		} else if spec.ZoneName.GetValueFrom() != nil {
			// The value_from reference would be resolved by the CLI before reaching here
			// For now, we'll handle the literal case
			locals.ZoneName = ""
		}
	}

	// The resource_group field is a StringValueOrRef. The platform middleware resolves
	// valueFrom references before IaC modules run, so .GetValue() always returns the
	// resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()
	locals.RecordName = spec.Name

	// Set TTL with default value
	if spec.TtlSeconds != nil {
		locals.TTL = int(*spec.TtlSeconds)
	} else {
		locals.TTL = 300 // Default TTL
	}

	// Set MX Priority with default value
	if spec.MxPriority != nil {
		locals.MxPriority = int(*spec.MxPriority)
	} else {
		locals.MxPriority = 10 // Default MX priority
	}

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureDnsRecord.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags["environment"] = target.Metadata.Env
	}

	return locals
}
