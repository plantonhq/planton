package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpkmskeyringv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpkmskeyring/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpKmsKeyRing     *gcpkmskeyringv1.GcpKmsKeyRing
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpkmskeyringv1.GcpKmsKeyRingStackInput) *Locals {
	locals := &Locals{}
	locals.GcpKmsKeyRing = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpKmsKeyRing.Spec.KeyRingName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpKmsKeyRing.String()),
	}

	if locals.GcpKmsKeyRing.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpKmsKeyRing.Metadata.Org
	}
	if locals.GcpKmsKeyRing.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpKmsKeyRing.Metadata.Env
	}
	if locals.GcpKmsKeyRing.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpKmsKeyRing.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
