package module

import (
	"strconv"
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudarmorpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudarmorpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpCloudArmorPolicy *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorPolicy
	GcpLabels           map[string]string

	// ProjectId is empty when the manifest omits it — the provider's default
	// project then applies (the same ambient contract the Terraform module
	// honors by passing null).
	ProjectId string

	// PolicyName falls back to metadata.name — explicit conditional, so both
	// engines derive the identical cloud-side name.
	PolicyName string

	// IsRegional is the scope selector: an empty spec.region builds the
	// global security policy, a region name builds the regional one. The
	// regional collection carries no labels, so GcpLabels applies to the
	// global arm only.
	IsRegional bool
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcloudarmorpolicyv1alpha1.GcpCloudArmorPolicyStackInput) *Locals {
	locals := &Locals{}
	locals.GcpCloudArmorPolicy = stackInput.Target
	locals.GcpProviderConfig = stackInput.ProviderConfig

	locals.ProjectId = locals.GcpCloudArmorPolicy.Spec.ProjectId.GetValue()

	locals.PolicyName = locals.GcpCloudArmorPolicy.Spec.PolicyName
	if locals.PolicyName == "" {
		locals.PolicyName = locals.GcpCloudArmorPolicy.Metadata.Name
	}

	locals.IsRegional = locals.GcpCloudArmorPolicy.Spec.Region != ""

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpCloudArmorPolicy.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.PolicyName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpCloudArmorPolicy.String())

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
