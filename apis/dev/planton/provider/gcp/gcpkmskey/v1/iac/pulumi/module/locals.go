package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	gcpkmskeyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpkmskey/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpKmsKey         *gcpkmskeyv1.GcpKmsKey
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpkmskeyv1.GcpKmsKeyStackInput) *Locals {
	locals := &Locals{}
	locals.GcpKmsKey = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: locals.GcpKmsKey.Spec.KeyName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpKmsKey.String()),
	}

	if locals.GcpKmsKey.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpKmsKey.Metadata.Org
	}
	if locals.GcpKmsKey.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpKmsKey.Metadata.Env
	}
	if locals.GcpKmsKey.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpKmsKey.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
