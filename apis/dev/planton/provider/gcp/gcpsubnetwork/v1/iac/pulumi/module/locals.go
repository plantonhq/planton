package module

import (
	"strconv"
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpsubnetworkv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpsubnetwork/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals aggregates frequently used input data so the rest of the module
// can reference short paths like locals.GcpSubnetwork.Metadata.Name rather
// than drilling into the original stackInput every time.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpSubnetwork     *gcpsubnetworkv1.GcpSubnetwork
	GcpLabels         map[string]string
}

// initializeLocals populates the Locals struct and derives a canonical set
// of GCP labels from the API‑resource metadata.  No additional logic is
// performed; we simply expose the values so other files stay terse and
// Terraform‑like.
func initializeLocals(_ *pulumi.Context, input *gcpsubnetworkv1.GcpSubnetworkStackInput) *Locals {
	locals := &Locals{}
	locals.GcpSubnetwork = input.Target

	// Base labels – always present.
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: input.Target.Spec.SubnetworkName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpSubnetwork.String()),
	}

	// Optional labels copied straight from metadata if the fields are set.
	if input.Target.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = input.Target.Metadata.Org
	}

	if input.Target.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = input.Target.Metadata.Env
	}

	if input.Target.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = input.Target.Metadata.Id
	}

	locals.GcpProviderConfig = input.ProviderConfig
	return locals
}
