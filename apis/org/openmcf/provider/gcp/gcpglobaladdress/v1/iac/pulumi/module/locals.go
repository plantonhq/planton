package module

import (
	"strconv"
	"strings"

	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	gcpglobaladdressv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpglobaladdress/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpGlobalAddress  *gcpglobaladdressv1.GcpGlobalAddress
	GcpLabels         map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpglobaladdressv1.GcpGlobalAddressStackInput) *Locals {
	locals := &Locals{}
	locals.GcpGlobalAddress = stackInput.Target
	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     strconv.FormatBool(true),
		gcplabelkeys.ResourceName: locals.GcpGlobalAddress.Spec.AddressName,
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpGlobalAddress.String()),
	}

	if locals.GcpGlobalAddress.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpGlobalAddress.Metadata.Org
	}
	if locals.GcpGlobalAddress.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpGlobalAddress.Metadata.Env
	}
	if locals.GcpGlobalAddress.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpGlobalAddress.Metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
