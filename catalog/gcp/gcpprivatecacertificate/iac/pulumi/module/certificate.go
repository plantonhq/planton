package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificateauthority"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// certificate issues the certificate from a CSR or from structured config.
// Everything but labels is immutable; destroy revokes it.
func certificate(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpPrivateCaCertificate.Spec
	resourceName := locals.GcpPrivateCaCertificate.Metadata.Name

	args := &certificateauthority.CertificateArgs{
		Location:             pulumi.String(spec.Location),
		Pool:                 pulumi.String(locals.PoolId),
		Name:                 pulumi.String(locals.CertificateId),
		CertificateAuthority: stringPtr(locals.CertificateAuthorityId),
		CertificateTemplate:  stringPtr(spec.CertificateTemplate.GetValue()),
		Lifetime:             stringPtr(spec.Lifetime),
		PemCsr:               stringPtr(spec.PemCsr),
		Labels:               pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if config := spec.Config; config != nil {
		// PEM is the only key format Google accepts; an empty format sends it.
		format := config.PublicKey.GetFormat()
		if format == "" {
			format = "PEM"
		}
		configArgs := &certificateauthority.CertificateConfigArgs{
			SubjectConfig: subjectConfig(config.SubjectConfig),
			X509Config:    x509Parameters(config.X509Config),
			PublicKey: &certificateauthority.CertificateConfigPublicKeyArgs{
				Format: pulumi.String(format),
				Key:    pulumi.String(config.PublicKey.GetKey()),
			},
		}
		if config.SubjectKeyId != "" {
			configArgs.SubjectKeyId = &certificateauthority.CertificateConfigSubjectKeyIdArgs{
				KeyId: pulumi.String(config.SubjectKeyId),
			}
		}
		args.Config = configArgs
	}

	// Engine-side destroy stance: DELETE (the provider's default, which
	// revokes), PREVENT, or ABANDON. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := certificateauthority.NewCertificate(ctx, resourceName, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create certificate")
	}

	ctx.Export(OpName, created.ID())
	ctx.Export(OpCertificateId, created.Name)
	ctx.Export(OpPemCertificate, created.PemCertificate)
	ctx.Export(OpPemCertificateChain, created.PemCertificateChains)
	ctx.Export(OpIssuerCertificateAuthority, created.IssuerCertificateAuthority)
	return nil
}
