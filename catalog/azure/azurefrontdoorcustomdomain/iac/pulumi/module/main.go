package module

import (
	"github.com/pkg/errors"
	azurefrontdoorcustomdomainv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorcustomdomain/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorCustomDomain.Spec

	// TLS is always on for custom domains, so the block is always sent.
	// The minimum TLS version is NOT modeled: Azure retired TLS 1.0/1.1
	// (March 2025) and the provider accepts exactly one value, so the
	// constant TLS12 is sent unconditionally -- a one-value knob is not
	// a choice. (minimum_version, never the deprecated
	// minimum_tls_version alias.)
	tlsArgs := &cdn.FrontdoorCustomDomainTlsArgs{
		MinimumVersion: pulumi.String("TLS12"),
	}

	// Sent only when chosen: ARM defaults to ManagedCertificate. With a
	// customer certificate the referenced Front Door secret carries the
	// key material; with a managed certificate the secret reference is
	// spec-forbidden, so nothing else is sent.
	if spec.Tls.CertificateType != azurefrontdoorcustomdomainv1alpha1.AzureFrontDoorCustomDomainCertificateType_azure_front_door_custom_domain_certificate_type_unspecified {
		tlsArgs.CertificateType = pulumi.String(certificateTypeStrings[spec.Tls.CertificateType])
	}
	if spec.Tls.SecretId != nil {
		tlsArgs.CdnFrontdoorSecretId = pulumi.String(spec.Tls.SecretId.GetValue())
	}

	// A cipher-suite policy is sent only when configured -- absence
	// serves Azure's default suite set. With CUSTOMIZED, the tls13 list
	// is sent only when the user pinned it (empty means Azure's TLS 1.3
	// defaults; when set, the spec guarantees both mandatory suites).
	if spec.Tls.CipherSuite != nil {
		cipherSuiteArgs := &cdn.FrontdoorCustomDomainTlsCipherSuiteArgs{
			Type: pulumi.String(cipherSuiteSetTypeStrings[spec.Tls.CipherSuite.Type]),
		}
		if spec.Tls.CipherSuite.CustomCiphers != nil {
			customCiphersArgs := &cdn.FrontdoorCustomDomainTlsCipherSuiteCustomCiphersArgs{
				Tls12s: pulumi.ToStringArray(spec.Tls.CipherSuite.CustomCiphers.Tls12),
			}
			if len(spec.Tls.CipherSuite.CustomCiphers.Tls13) > 0 {
				customCiphersArgs.Tls13s = pulumi.ToStringArray(spec.Tls.CipherSuite.CustomCiphers.Tls13)
			}
			cipherSuiteArgs.CustomCiphers = customCiphersArgs
		}
		tlsArgs.CipherSuite = cipherSuiteArgs
	}

	domainArgs := &cdn.FrontdoorCustomDomainArgs{
		Name:                  pulumi.String(spec.DomainName),
		CdnFrontdoorProfileId: pulumi.String(locals.ProfileId),
		HostName:              pulumi.String(spec.HostName),
		Tls:                   tlsArgs,
	}

	// Sent only when the hostname's DNS lives in Azure DNS -- Azure then
	// watches the referenced zone for the validation TXT record.
	if spec.DnsZoneId != nil {
		domainArgs.DnsZoneId = pulumi.String(spec.DnsZoneId.GetValue())
	}

	// Creation returns with the domain in a PENDING-VALIDATION state --
	// deployment does not block on DNS proof. The validation_token
	// output is the challenge to publish as a TXT record at
	// _dnsauth.<host_name>; Azure then flips the domain to approved.
	createdCustomDomain, err := cdn.NewFrontdoorCustomDomain(ctx,
		spec.DomainName,
		domainArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door custom domain %s", spec.DomainName)
	}

	// Export outputs. custom_domain_id is what
	// AzureFrontDoorRoute's custom_domain_ids references;
	// validation_token/expiration_date drive the DNS validation
	// workflow; host_name is the name to CNAME to the endpoint.
	ctx.Export(OpCustomDomainId, createdCustomDomain.ID())
	ctx.Export(OpHostName, createdCustomDomain.HostName)
	ctx.Export(OpValidationToken, createdCustomDomain.ValidationToken)
	ctx.Export(OpExpirationDate, createdCustomDomain.ExpirationDate)

	return nil
}
