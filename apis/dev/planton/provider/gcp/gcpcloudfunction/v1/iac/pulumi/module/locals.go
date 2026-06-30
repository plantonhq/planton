package module

import (
	"strconv"
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpcloudfunctionv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcloudfunction/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
)

// Locals holds handy references and derived values used across this module.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpCloudFunction  *gcpcloudfunctionv1.GcpCloudFunction
	GcpLabels         map[string]string
	FunctionName      string
	IsHttpTrigger     bool
}

// initializeLocals fills the Locals struct from the incoming stack input.
func initializeLocals(stackInput *gcpcloudfunctionv1.GcpCloudFunctionStackInput) *Locals {
	locals := &Locals{}

	locals.GcpCloudFunction = stackInput.Target
	target := stackInput.Target
	spec := target.Spec

	locals.GcpProviderConfig = stackInput.ProviderConfig

	// Determine function name: use spec.FunctionName if provided, otherwise metadata.Name
	if spec.FunctionName != "" {
		locals.FunctionName = spec.FunctionName
	} else {
		locals.FunctionName = target.Metadata.Name
	}

	// Determine if trigger is HTTP or Event-driven
	// Default to HTTP if trigger is nil or trigger_type is 0 (HTTP)
	locals.IsHttpTrigger = spec.Trigger == nil ||
		spec.Trigger.TriggerType == gcpcloudfunctionv1.GcpCloudFunctionTriggerType_HTTP

	// Build GCP labels
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpCloudFunction.String()),
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
