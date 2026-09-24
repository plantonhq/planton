package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificateauthority"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// certificateAuthority creates the authority. A subordinate named by
// reference is signed and activated on create; destroy disables it and,
// unless skip_grace_period, leaves it in Google's 30-day soft delete.
func certificateAuthority(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpPrivateCaCertificateAuthority.Spec
	resourceName := locals.GcpPrivateCaCertificateAuthority.Metadata.Name
	config := spec.Config

	// Deletion guard, honest by default: an unset spec field means true, so a
	// destroy fails until the manifest explicitly opts out (identical to the
	// Terraform module).
	deletionProtection := true
	if spec.DeletionProtection != nil {
		deletionProtection = spec.GetDeletionProtection()
	}

	configArgs := &certificateauthority.AuthorityConfigArgs{
		SubjectConfig: subjectConfig(config.SubjectConfig),
		X509Config:    x509Parameters(config.X509Config),
	}
	if config.SubjectKeyId != "" {
		configArgs.SubjectKeyId = &certificateauthority.AuthorityConfigSubjectKeyIdArgs{
			KeyId: pulumi.String(config.SubjectKeyId),
		}
	}

	args := &certificateauthority.AuthorityArgs{
		Location:               pulumi.String(spec.Location),
		Pool:                   pulumi.String(locals.PoolId),
		CertificateAuthorityId: pulumi.String(locals.CertificateAuthorityId),
		Config:                 configArgs,
		KeySpec: &certificateauthority.AuthorityKeySpecArgs{
			Algorithm:          stringPtr(spec.KeySpec.Algorithm),
			CloudKmsKeyVersion: stringPtr(spec.KeySpec.CloudKmsKeyVersion.GetValue()),
		},
		Type:                               stringPtr(spec.Type),
		Lifetime:                           stringPtr(spec.Lifetime),
		PemCaCertificate:                   stringPtr(spec.PemCaCertificate),
		GcsBucket:                          stringPtr(spec.GcsBucket.GetValue()),
		DesiredState:                       stringPtr(spec.DesiredState),
		DeletionProtection:                 pulumi.Bool(deletionProtection),
		SkipGracePeriod:                    pulumi.Bool(spec.SkipGracePeriod),
		IgnoreActiveCertificatesOnDeletion: pulumi.Bool(spec.IgnoreActiveCertificatesOnDeletion),
		Labels:                             pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if subordinate := spec.SubordinateConfig; subordinate != nil {
		subordinateArgs := &certificateauthority.AuthoritySubordinateConfigArgs{
			CertificateAuthority: stringPtr(subordinate.CertificateAuthority.GetValue()),
		}
		if len(subordinate.PemIssuerChain) > 0 {
			subordinateArgs.PemIssuerChain = &certificateauthority.AuthoritySubordinateConfigPemIssuerChainArgs{
				PemCertificates: pulumi.ToStringArray(subordinate.PemIssuerChain),
			}
		}
		args.SubordinateConfig = subordinateArgs
	}
	if urls := spec.UserDefinedAccessUrls; urls != nil {
		args.UserDefinedAccessUrls = &certificateauthority.AuthorityUserDefinedAccessUrlsArgs{
			AiaIssuingCertificateUrls: stringArray(urls.AiaIssuingCertificateUrls),
			CrlAccessUrls:             stringArray(urls.CrlAccessUrls),
		}
	}

	// Engine-side destroy stance: DELETE (the provider's default), PREVENT,
	// or ABANDON. Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := certificateauthority.NewAuthority(ctx, resourceName, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create certificate authority")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpCertificateAuthorityId, created.CertificateAuthorityId)
	ctx.Export(OpState, created.State)
	ctx.Export(OpPemCaCertificate, created.PemCaCertificates.ApplyT(func(chain []string) string {
		if len(chain) == 0 {
			return ""
		}
		return chain[0]
	}).(pulumi.StringOutput))
	ctx.Export(OpPemCaCertificates, created.PemCaCertificates)
	ctx.Export(OpCaCertificateAccessUrl, created.AccessUrls.ApplyT(func(urls []certificateauthority.AuthorityAccessUrl) string {
		if len(urls) == 0 || urls[0].CaCertificateAccessUrl == nil {
			return ""
		}
		return *urls[0].CaCertificateAccessUrl
	}).(pulumi.StringOutput))
	ctx.Export(OpCrlAccessUrls, created.AccessUrls.ApplyT(func(urls []certificateauthority.AuthorityAccessUrl) []string {
		if len(urls) == 0 {
			return []string{}
		}
		return urls[0].CrlAccessUrls
	}).(pulumi.StringArrayOutput))
	return nil
}
