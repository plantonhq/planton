package module

import (
	"strconv"
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpgkeworkloadidentitybindingv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpgkeworkloadidentitybinding/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals keeps frequently‑used values (metadata, labels, credentials) handy for the module.
type Locals struct {
	GcpProviderConfig             *gcpprovider.GcpProviderConfig
	GcpGkeWorkloadIdentityBinding *gcpgkeworkloadidentitybindingv1.GcpGkeWorkloadIdentityBinding
	GcpLabels                     map[string]string
}

// initializeLocals populates the Locals struct from the stack input.
// It mirrors the pattern used in the gcp_gke_cluster module and applies the same Planton label strategy.
func initializeLocals(_ *pulumi.Context, stackInput *gcpgkeworkloadidentitybindingv1.GcpGkeWorkloadIdentityBindingStackInput) *Locals {
	locals := &Locals{}

	locals.GcpGkeWorkloadIdentityBinding = stackInput.Target

	// Standard Planton‑wide labels for GCP resources
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: locals.GcpGkeWorkloadIdentityBinding.Metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpGkeWorkloadIdentityBinding.String()),
	}

	if locals.GcpGkeWorkloadIdentityBinding.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpGkeWorkloadIdentityBinding.Metadata.Org
	}

	if locals.GcpGkeWorkloadIdentityBinding.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpGkeWorkloadIdentityBinding.Metadata.Env
	}

	if locals.GcpGkeWorkloadIdentityBinding.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpGkeWorkloadIdentityBinding.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig

	return locals
}
