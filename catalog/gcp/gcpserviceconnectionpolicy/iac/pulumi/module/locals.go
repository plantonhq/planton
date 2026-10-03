package module

import (
	"strings"

	gcpserviceconnectionpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpserviceconnectionpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// computeSelfLinkPrefix is stripped from network/subnet values: the
// Service Connectivity API rejects full https:// self-link URLs and
// requires relative resource paths. Stripping is a no-op for values
// already in relative form — identical normalization to the Terraform
// module.
const computeSelfLinkPrefix = "https://www.googleapis.com/compute/v1/"

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpServiceConnectionPolicy *gcpserviceconnectionpolicyv1alpha1.GcpServiceConnectionPolicy
	GcpLabels                  map[string]string

	// Policy name defaults to metadata.name when policy_name is omitted.
	PolicyName string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpserviceconnectionpolicyv1alpha1.GcpServiceConnectionPolicyIacInput) *Locals {
	locals := &Locals{}
	locals.GcpServiceConnectionPolicy = iacInput.Target

	locals.PolicyName = locals.GcpServiceConnectionPolicy.Spec.PolicyName
	if locals.PolicyName == "" {
		locals.PolicyName = locals.GcpServiceConnectionPolicy.Metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts — identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range locals.GcpServiceConnectionPolicy.Spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = locals.PolicyName
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(catalogkind.CatalogKind_GcpServiceConnectionPolicy.String())

	if locals.GcpServiceConnectionPolicy.Metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = locals.GcpServiceConnectionPolicy.Metadata.Org
	}
	if locals.GcpServiceConnectionPolicy.Metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = locals.GcpServiceConnectionPolicy.Metadata.Env
	}
	if locals.GcpServiceConnectionPolicy.Metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = locals.GcpServiceConnectionPolicy.Metadata.Id
	}

	return locals
}

// toResourcePath normalizes a compute self-link URL to the relative
// resource path the Service Connectivity API expects.
func toResourcePath(value string) string {
	return strings.TrimPrefix(value, computeSelfLinkPrefix)
}
