package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpprivatecacertificatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecacertificate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig       *gcpprovider.GcpProviderConfig
	GcpPrivateCaCertificate *gcpprivatecacertificatev1alpha1.GcpPrivateCaCertificate
	GcpLabels               map[string]string

	// PoolId is the bare pool ID Google's resource is keyed by (the pool
	// reference resolves to the full path; a literal may be either) --
	// identical to the Terraform module's locals.pool.
	PoolId string

	// CertificateId is spec.certificate_id when set, otherwise
	// metadata.name -- identical to the Terraform module's
	// locals.certificate_id.
	CertificateId string

	// CertificateAuthorityId is the bare ID of the signing authority, or
	// empty to let the pool choose -- identical to the Terraform module's
	// locals.certificate_authority.
	CertificateAuthorityId string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpprivatecacertificatev1alpha1.GcpPrivateCaCertificateStackInput) *Locals {
	locals := &Locals{}
	locals.GcpPrivateCaCertificate = stackInput.Target
	metadata := locals.GcpPrivateCaCertificate.Metadata
	spec := locals.GcpPrivateCaCertificate.Spec

	locals.PoolId = lastSegment(spec.Pool.GetValue())
	locals.CertificateId = spec.CertificateId
	if locals.CertificateId == "" {
		locals.CertificateId = metadata.Name
	}
	locals.CertificateAuthorityId = lastSegment(spec.CertificateAuthority.GetValue())

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = map[string]string{}
	for key, value := range spec.Labels {
		locals.GcpLabels[key] = value
	}
	locals.GcpLabels[gcplabelkeys.Resource] = "true"
	locals.GcpLabels[gcplabelkeys.ResourceName] = metadata.Name
	locals.GcpLabels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpPrivateCaCertificate.String())

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
