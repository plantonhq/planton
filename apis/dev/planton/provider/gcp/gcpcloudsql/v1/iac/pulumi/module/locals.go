package module

import (
	"strconv"
	"strings"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpcloudsqlv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudsql/v1"
)

// Locals holds handy references and derived values used across this module.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpCloudSql       *gcpcloudsqlv1.GcpCloudSql
	GcpLabels         map[string]string
}

// initializeLocals fills the Locals struct from the incoming stack input.
func initializeLocals(stackInput *gcpcloudsqlv1.GcpCloudSqlStackInput) *Locals {
	locals := &Locals{}

	locals.GcpCloudSql = stackInput.Target

	target := stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpCloudSql.String()),
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
