package module

import (
	"strconv"
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpvpcnetworkv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvpcnetwork/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals keeps frequently‑used values (metadata, labels, credentials) handy for the module.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpVpcNetwork     *gcpvpcnetworkv1alpha1.GcpVpcNetwork
	GcpLabels         map[string]string
}

// initializeLocals populates the Locals struct from the IaC input.
// It mirrors the pattern used in the gcp_gke_cluster module and applies the same Planton label strategy.
func initializeLocals(_ *pulumi.Context, iacInput *gcpvpcnetworkv1alpha1.GcpVpcNetworkIacInput) *Locals {
	locals := &Locals{}

	locals.GcpVpcNetwork = iacInput.Target

	// Standard Planton‑wide labels for GCP resources
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: locals.GcpVpcNetwork.Spec.NetworkName,
		gcplabelkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_GcpVpcNetwork.String()),
	}

	if locals.GcpVpcNetwork.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpVpcNetwork.Metadata.Org
	}

	if locals.GcpVpcNetwork.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpVpcNetwork.Metadata.Env
	}

	if locals.GcpVpcNetwork.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpVpcNetwork.Metadata.Id
	}

	locals.GcpProviderConfig = iacInput.ProviderConfig

	return locals
}
