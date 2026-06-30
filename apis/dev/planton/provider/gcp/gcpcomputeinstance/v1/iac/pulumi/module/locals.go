package module

import (
	"strconv"
	"strings"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpcomputeinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpcomputeinstance/v1"
)

// Locals holds handy references and derived values used across this module.
type Locals struct {
	GcpProviderConfig  *gcpprovider.GcpProviderConfig
	GcpComputeInstance *gcpcomputeinstancev1.GcpComputeInstance
	GcpLabels          map[string]string
}

// initializeLocals fills the Locals struct from the incoming stack input.
func initializeLocals(stackInput *gcpcomputeinstancev1.GcpComputeInstanceStackInput) *Locals {
	locals := &Locals{}

	locals.GcpComputeInstance = stackInput.Target

	target := stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: target.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpComputeInstance.String()),
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

	// Merge user-provided labels with system labels
	if target.Spec.Labels != nil {
		for k, v := range target.Spec.Labels {
			locals.GcpLabels[k] = v
		}
	}

	return locals
}
