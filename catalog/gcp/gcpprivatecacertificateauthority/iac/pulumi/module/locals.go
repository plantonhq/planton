package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpprivatecacertificateauthorityv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecacertificateauthority/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig                *gcpprovider.GcpProviderConfig
	GcpPrivateCaCertificateAuthority *gcpprivatecacertificateauthorityv1alpha1.GcpPrivateCaCertificateAuthority
	GcpLabels                        map[string]string

	// PoolId is the bare pool ID Google's resource is keyed by. The spec's
	// pool arrives either as the pool's full resource path (a
	// GcpPrivateCaPool reference resolves to its name output) or as the bare
	// ID -- identical to the Terraform module's locals.pool.
	PoolId string

	// CertificateAuthorityId is spec.certificate_authority_id when set,
	// otherwise metadata.name -- identical to the Terraform module's
	// locals.certificate_authority_id.
	CertificateAuthorityId string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpprivatecacertificateauthorityv1alpha1.GcpPrivateCaCertificateAuthorityStackInput) *Locals {
	locals := &Locals{}
	locals.GcpPrivateCaCertificateAuthority = stackInput.Target
	metadata := locals.GcpPrivateCaCertificateAuthority.Metadata
	spec := locals.GcpPrivateCaCertificateAuthority.Spec

	locals.PoolId = lastSegment(spec.Pool.GetValue())
	locals.CertificateAuthorityId = spec.CertificateAuthorityId
	if locals.CertificateAuthorityId == "" {
		locals.CertificateAuthorityId = metadata.Name
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpPrivateCaCertificateAuthority.String())

	if metadata.Org != "" {
		locals.GcpLabels[gcplabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		locals.GcpLabels[gcplabelkeys.Environment] = metadata.Env
	}
	if metadata.Id != "" {
		locals.GcpLabels[gcplabelkeys.ResourceId] = metadata.Id
	}

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}

// lastSegment returns what follows the last "/" -- the bare ID of a full
// resource path, or the value itself when it is already bare.
func lastSegment(value string) string {
	return value[strings.LastIndex(value, "/")+1:]
}
