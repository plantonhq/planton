package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	gcpcloudarmorpolicyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpcloudarmorpolicy/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpCloudArmorPolicy *gcpcloudarmorpolicyv1.GcpCloudArmorPolicy
	GcpLabels           map[string]string
	PolicyName          string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcloudarmorpolicyv1.GcpCloudArmorPolicyStackInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudArmorPolicy = stackInput.Target
	locals.GcpProviderConfig = stackInput.ProviderConfig

	// Policy name: use explicit value or fall back to metadata.name.
	locals.PolicyName = locals.GcpCloudArmorPolicy.Spec.PolicyName
	if locals.PolicyName == "" {
		locals.PolicyName = locals.GcpCloudArmorPolicy.Metadata.Name
	}

	locals.GcpLabels = map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: strings.ToLower(locals.GcpCloudArmorPolicy.Metadata.Name),
		gcplabelkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_GcpCloudArmorPolicy.String()),
	}

	if locals.GcpCloudArmorPolicy.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpCloudArmorPolicy.Metadata.Org
	}
	if locals.GcpCloudArmorPolicy.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpCloudArmorPolicy.Metadata.Env
	}
	if locals.GcpCloudArmorPolicy.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpCloudArmorPolicy.Metadata.Id
	}

	return locals
}
