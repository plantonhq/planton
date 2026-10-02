package module

import (
	gcpprivatecacertificatetemplatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecacertificatetemplate/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificateauthority"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// x509Parameters maps the spec's X.509 parameters onto the provider's
// CertificateTemplatePredefinedValues block. Every block is sent only when the spec sets it.
func x509Parameters(p *gcpprivatecacertificatetemplatev1alpha1.GcpPrivateCaCertificateTemplateX509Parameters) *certificateauthority.CertificateTemplatePredefinedValuesArgs {
	args := &certificateauthority.CertificateTemplatePredefinedValuesArgs{}
	if p.GetCaOptions() != nil {
		args.CaOptions = caOptions(p.GetCaOptions())
	}
	if p.GetKeyUsage() != nil {
		args.KeyUsage = keyUsage(p.GetKeyUsage())
	}
	if len(p.GetAiaOcspServers()) > 0 {
		args.AiaOcspServers = pulumi.ToStringArray(p.GetAiaOcspServers())
	}
	if len(p.GetPolicyIds()) > 0 {
		policyIds := certificateauthority.CertificateTemplatePredefinedValuesPolicyIdArray{}
		for _, oid := range p.GetPolicyIds() {
			policyIds = append(policyIds, &certificateauthority.CertificateTemplatePredefinedValuesPolicyIdArgs{ObjectIdPaths: objectIdPath(oid)})
		}
		args.PolicyIds = policyIds
	}
	if len(p.GetAdditionalExtensions()) > 0 {
		extensions := certificateauthority.CertificateTemplatePredefinedValuesAdditionalExtensionArray{}
		for _, extension := range p.GetAdditionalExtensions() {
			extensions = append(extensions, &certificateauthority.CertificateTemplatePredefinedValuesAdditionalExtensionArgs{
				Critical: pulumi.Bool(extension.GetCritical()),
				ObjectId: &certificateauthority.CertificateTemplatePredefinedValuesAdditionalExtensionObjectIdArgs{ObjectIdPaths: objectIdPath(extension.GetObjectId())},
				Value:    pulumi.String(extension.GetValue()),
			})
		}
		args.AdditionalExtensions = extensions
	}
	if nc := p.GetNameConstraints(); nc != nil {
		args.NameConstraints = &certificateauthority.CertificateTemplatePredefinedValuesNameConstraintsArgs{
			Critical:                pulumi.Bool(nc.GetCritical()),
			PermittedDnsNames:       stringArray(nc.PermittedDnsNames),
			ExcludedDnsNames:        stringArray(nc.ExcludedDnsNames),
			PermittedIpRanges:       stringArray(nc.PermittedIpRanges),
			ExcludedIpRanges:        stringArray(nc.ExcludedIpRanges),
			PermittedEmailAddresses: stringArray(nc.PermittedEmailAddresses),
			ExcludedEmailAddresses:  stringArray(nc.ExcludedEmailAddresses),
			PermittedUris:           stringArray(nc.PermittedUris),
			ExcludedUris:            stringArray(nc.ExcludedUris),
		}
	}
	return args
}

// caOptions maps the presence-based basic constraints onto the provider's
// flags. A path length of 0 is sent through zero_max_issuer_path_length,
// because the provider reads a plain 0 as unset.
func caOptions(o *gcpprivatecacertificatetemplatev1alpha1.GcpPrivateCaCertificateTemplateCaOptions) *certificateauthority.CertificateTemplatePredefinedValuesCaOptionsArgs {
	args := &certificateauthority.CertificateTemplatePredefinedValuesCaOptionsArgs{}
	// Unset is_ca omits the CA flag; the provider's template resource sends
	// false unless null_ca is set, so unset maps to null_ca.
	if o.IsCa != nil {
		args.IsCa = pulumi.Bool(o.GetIsCa())
	} else {
		args.NullCa = pulumi.Bool(true)
	}
	if o.MaxIssuerPathLength != nil {
		if o.GetMaxIssuerPathLength() == 0 {
			args.ZeroMaxIssuerPathLength = pulumi.Bool(true)
		} else {
			args.MaxIssuerPathLength = pulumi.Int(int(o.GetMaxIssuerPathLength()))
		}
	}
	return args
}

// keyUsage maps the key usage groups, each sent only when the spec sets it.
func keyUsage(k *gcpprivatecacertificatetemplatev1alpha1.GcpPrivateCaCertificateTemplateKeyUsage) *certificateauthority.CertificateTemplatePredefinedValuesKeyUsageArgs {
	args := &certificateauthority.CertificateTemplatePredefinedValuesKeyUsageArgs{}
	if base := k.GetBaseKeyUsage(); base != nil {
		args.BaseKeyUsage = &certificateauthority.CertificateTemplatePredefinedValuesKeyUsageBaseKeyUsageArgs{
			DigitalSignature:  pulumi.Bool(base.GetDigitalSignature()),
			ContentCommitment: pulumi.Bool(base.GetContentCommitment()),
			KeyEncipherment:   pulumi.Bool(base.GetKeyEncipherment()),
			DataEncipherment:  pulumi.Bool(base.GetDataEncipherment()),
			KeyAgreement:      pulumi.Bool(base.GetKeyAgreement()),
			CertSign:          pulumi.Bool(base.GetCertSign()),
			CrlSign:           pulumi.Bool(base.GetCrlSign()),
			EncipherOnly:      pulumi.Bool(base.GetEncipherOnly()),
			DecipherOnly:      pulumi.Bool(base.GetDecipherOnly()),
		}
	}
	if extended := k.GetExtendedKeyUsage(); extended != nil {
		args.ExtendedKeyUsage = &certificateauthority.CertificateTemplatePredefinedValuesKeyUsageExtendedKeyUsageArgs{
			ServerAuth:      pulumi.Bool(extended.GetServerAuth()),
			ClientAuth:      pulumi.Bool(extended.GetClientAuth()),
			CodeSigning:     pulumi.Bool(extended.GetCodeSigning()),
			EmailProtection: pulumi.Bool(extended.GetEmailProtection()),
			TimeStamping:    pulumi.Bool(extended.GetTimeStamping()),
			OcspSigning:     pulumi.Bool(extended.GetOcspSigning()),
		}
	}
	if len(k.GetUnknownExtendedKeyUsages()) > 0 {
		unknown := certificateauthority.CertificateTemplatePredefinedValuesKeyUsageUnknownExtendedKeyUsageArray{}
		for _, oid := range k.GetUnknownExtendedKeyUsages() {
			unknown = append(unknown, &certificateauthority.CertificateTemplatePredefinedValuesKeyUsageUnknownExtendedKeyUsageArgs{ObjectIdPaths: objectIdPath(oid)})
		}
		args.UnknownExtendedKeyUsages = unknown
	}
	return args
}

// objectIdPath converts an OID's arcs to the provider's integer list.
func objectIdPath(oid *gcpprivatecacertificatetemplatev1alpha1.GcpPrivateCaCertificateTemplateObjectId) pulumi.IntArray {
	path := pulumi.IntArray{}
	for _, arc := range oid.GetObjectIdPath() {
		path = append(path, pulumi.Int(int(arc)))
	}
	return path
}

// stringArray sends a list only when it has entries, so an empty spec list
// stays unset instead of an empty value the provider would diff on.
func stringArray(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}

// stringPtr sends a string only when it is set.
func stringPtr(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.String(value)
}
