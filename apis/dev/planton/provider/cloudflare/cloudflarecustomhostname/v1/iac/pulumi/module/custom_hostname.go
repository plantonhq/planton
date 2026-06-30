package module

import (
	"github.com/pkg/errors"
	cloudflarecustomhostnamev1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarecustomhostname/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// customHostname onboards a customer hostname onto a Cloudflare for SaaS zone and
// exports the ownership-verification records the customer needs to activate it.
// ssl defaults (bundle_method "ubiquitous", type "dv") are coalesced here to match
// the control-plane middleware and the Terraform module byte-for-byte.
func customHostname(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareCustomHostname.Spec

	args := &cloudflare.CustomHostnameArgs{
		ZoneId:   pulumi.String(spec.ZoneId.GetValue()),
		Hostname: pulumi.String(spec.Hostname),
	}
	if spec.CustomOriginServer != nil && spec.CustomOriginServer.GetValue() != "" {
		args.CustomOriginServer = pulumi.StringPtr(spec.CustomOriginServer.GetValue())
	}
	if spec.CustomOriginSni != "" {
		args.CustomOriginSni = pulumi.StringPtr(spec.CustomOriginSni)
	}
	if len(spec.CustomMetadata) > 0 {
		args.CustomMetadata = pulumi.ToStringMap(spec.CustomMetadata)
	}
	if spec.Ssl != nil {
		args.Ssl = buildSsl(spec.Ssl)
	}

	created, err := cloudflare.NewCustomHostname(
		ctx,
		"custom-hostname",
		args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudflare custom hostname")
	}

	ctx.Export(OpCustomHostnameId, created.ID())
	ctx.Export(OpStatus, created.Status)
	ctx.Export(OpOwnershipVerificationName, created.OwnershipVerification.Name())
	ctx.Export(OpOwnershipVerificationType, created.OwnershipVerification.Type())
	ctx.Export(OpOwnershipVerificationValue, created.OwnershipVerification.Value())
	ctx.Export(OpOwnershipVerificationHttpUrl, created.OwnershipVerificationHttp.HttpUrl())
	ctx.Export(OpOwnershipVerificationHttpBody, created.OwnershipVerificationHttp.HttpBody())
	ctx.Export(OpVerificationErrors, created.VerificationErrors)
	ctx.Export(OpCreatedAt, created.CreatedAt)

	return nil
}

// buildSsl maps the spec ssl block onto the provider's ssl args. Note the Pulumi
// SDK names differ from the Terraform provider for two fields: the spec's
// `custom_cert_bundle` maps to `CustomCertBundles` and `settings.tls_1_3` maps to
// `Tls13`; both carry identical data — see docs/README.md ("Field-name nuance").
func buildSsl(s *cloudflarecustomhostnamev1.CloudflareCustomHostnameSsl) *cloudflare.CustomHostnameSslArgs {
	args := &cloudflare.CustomHostnameSslArgs{}

	bundleMethod := s.GetBundleMethod()
	if bundleMethod == "" {
		bundleMethod = "ubiquitous"
	}
	args.BundleMethod = pulumi.StringPtr(bundleMethod)

	sslType := s.GetType()
	if sslType == "" {
		sslType = "dv"
	}
	args.Type = pulumi.StringPtr(sslType)

	if s.CertificateAuthority != "" {
		args.CertificateAuthority = pulumi.StringPtr(s.CertificateAuthority)
	}
	if s.CloudflareBranding {
		args.CloudflareBranding = pulumi.BoolPtr(true)
	}
	if s.Method != "" {
		args.Method = pulumi.StringPtr(s.Method)
	}
	if s.Wildcard {
		args.Wildcard = pulumi.BoolPtr(true)
	}
	if s.CustomCertificate != "" {
		args.CustomCertificate = pulumi.StringPtr(s.CustomCertificate)
	}
	if s.CustomCsrId != "" {
		args.CustomCsrId = pulumi.StringPtr(s.CustomCsrId)
	}
	if s.CustomKey != "" {
		args.CustomKey = pulumi.StringPtr(s.CustomKey)
	}

	if len(s.CustomCertBundle) > 0 {
		bundles := make(cloudflare.CustomHostnameSslCustomCertBundleArray, 0, len(s.CustomCertBundle))
		for _, b := range s.CustomCertBundle {
			bundles = append(bundles, &cloudflare.CustomHostnameSslCustomCertBundleArgs{
				CustomCertificate: pulumi.String(b.CustomCertificate),
				CustomKey:         pulumi.String(b.CustomKey),
			})
		}
		args.CustomCertBundles = bundles
	}

	if s.Settings != nil {
		settings := &cloudflare.CustomHostnameSslSettingsArgs{}
		if len(s.Settings.Ciphers) > 0 {
			settings.Ciphers = pulumi.ToStringArray(s.Settings.Ciphers)
		}
		if s.Settings.EarlyHints != "" {
			settings.EarlyHints = pulumi.StringPtr(s.Settings.EarlyHints)
		}
		if s.Settings.Http2 != "" {
			settings.Http2 = pulumi.StringPtr(s.Settings.Http2)
		}
		if s.Settings.MinTlsVersion != "" {
			settings.MinTlsVersion = pulumi.StringPtr(s.Settings.MinTlsVersion)
		}
		if s.Settings.GetTls_1_3() != "" {
			settings.Tls13 = pulumi.StringPtr(s.Settings.GetTls_1_3())
		}
		args.Settings = settings
	}

	return args
}
