package module

import (
	"strconv"
	"strings"

	gcpdnszonev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdnszone/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpDnsZone *gcpdnszonev1alpha1.GcpDnsZone
	GcpLabels  map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpdnszonev1alpha1.GcpDnsZoneIacInput) *Locals {
	locals := &Locals{}

	locals.GcpDnsZone = iacInput.Target
	target := iacInput.Target
	spec := target.Spec

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_GcpDnsZone.String()),
	}

	if target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = target.Metadata.Env
	}

	for k, v := range spec.Labels {
		locals.GcpLabels[k] = v
	}

	return locals
}
