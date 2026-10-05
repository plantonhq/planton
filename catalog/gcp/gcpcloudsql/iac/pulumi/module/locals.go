package module

import (
	"strconv"
	"strings"

	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudsqlv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudsql/v1alpha1"
)

// Locals holds handy references and derived values used across this module.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpCloudSql       *gcpcloudsqlv1alpha1.GcpCloudSql
	GcpLabels         map[string]string
}

// initializeLocals fills the Locals struct from the incoming IaC input.
func initializeLocals(iacInput *gcpcloudsqlv1alpha1.GcpCloudSqlIacInput) *Locals {
	locals := &Locals{}

	locals.GcpCloudSql = iacInput.Target

	target := iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig

	// The instance name (not metadata.name) keys the resource-name label so
	// the label matches what is visible in the GCP console — the Terraform
	// module applies the identical set.
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Spec.InstanceName,
		gcplabelkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_GcpCloudSql.String()),
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

	return locals
}
