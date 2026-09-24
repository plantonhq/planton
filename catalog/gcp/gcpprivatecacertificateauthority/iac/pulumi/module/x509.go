package module

import (
	gcpprivatecacertificateauthorityv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecacertificateauthority/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificateauthority"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// x509Parameters maps the spec's X.509 parameters onto the provider's
// AuthorityConfigX509Config block. key_usage is sent empty when the spec omits it (the provider requires
// it); ca_options always carries is_ca (the spec's rule).
func x509Parameters(p *gcpprivatecacertificateauthorityv1alpha1.GcpPrivateCaCertificateAuthorityX509Parameters) *certificateauthority.AuthorityConfigX509ConfigArgs {
	args := &certificateauthority.AuthorityConfigX509ConfigArgs{
		CaOptions: caOptions(p.GetCaOptions()),
		KeyUsage:  keyUsage(p.GetKeyUsage()),
	}
	if len(p.GetAiaOcspServers()) > 0 {
		args.AiaOcspServers = pulumi.ToStringArray(p.GetAiaOcspServers())
	}
	if len(p.GetPolicyIds()) > 0 {
		policyIds := certificateauthority.AuthorityConfigX509ConfigPolicyIdArray{}
		for _, oid := range p.GetPolicyIds() {
			policyIds = append(policyIds, &certificateauthority.AuthorityConfigX509ConfigPolicyIdArgs{ObjectIdPaths: objectIdPath(oid)})
		}
		args.PolicyIds = policyIds
	}
	if len(p.GetAdditionalExtensions()) > 0 {
		extensions := certificateauthority.AuthorityConfigX509ConfigAdditionalExtensionArray{}
		for _, extension := range p.GetAdditionalExtensions() {
			extensions = append(extensions, &certificateauthority.AuthorityConfigX509ConfigAdditionalExtensionArgs{
				Critical: pulumi.Bool(extension.GetCritical()),
				ObjectId: &certificateauthority.AuthorityConfigX509ConfigAdditionalExtensionObjectIdArgs{ObjectIdPaths: objectIdPath(extension.GetObjectId())},
				Value:    pulumi.String(extension.GetValue()),
			})
		}
		args.AdditionalExtensions = extensions
	}
	if nc := p.GetNameConstraints(); nc != nil {
		args.NameConstraints = &certificateauthority.AuthorityConfigX509ConfigNameConstraintsArgs{
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
func caOptions(o *gcpprivatecacertificateauthorityv1alpha1.GcpPrivateCaCertificateAuthorityCaOptions) *certificateauthority.AuthorityConfigX509ConfigCaOptionsArgs {
	// is_ca is required on an authority (the spec's rule); false is sent
	// with non_ca, the provider's way to state CA:FALSE.
	args := &certificateauthority.AuthorityConfigX509ConfigCaOptionsArgs{IsCa: pulumi.Bool(o.GetIsCa())}
	if o.IsCa != nil && !o.GetIsCa() {
		args.NonCa = pulumi.Bool(true)
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

// keyUsage maps the key usage groups. The provider requires both groups
// whenever key usage is sent; an omitted group is sent with every bit false,
// which states nothing.
func keyUsage(k *gcpprivatecacertificateauthorityv1alpha1.GcpPrivateCaCertificateAuthorityKeyUsage) *certificateauthority.AuthorityConfigX509ConfigKeyUsageArgs {
	base := k.GetBaseKeyUsage()
	extended := k.GetExtendedKeyUsage()
	args := &certificateauthority.AuthorityConfigX509ConfigKeyUsageArgs{
		BaseKeyUsage: &certificateauthority.AuthorityConfigX509ConfigKeyUsageBaseKeyUsageArgs{
			DigitalSignature:  pulumi.Bool(base.GetDigitalSignature()),
			ContentCommitment: pulumi.Bool(base.GetContentCommitment()),
			KeyEncipherment:   pulumi.Bool(base.GetKeyEncipherment()),
			DataEncipherment:  pulumi.Bool(base.GetDataEncipherment()),
			KeyAgreement:      pulumi.Bool(base.GetKeyAgreement()),
			CertSign:          pulumi.Bool(base.GetCertSign()),
			CrlSign:           pulumi.Bool(base.GetCrlSign()),
			EncipherOnly:      pulumi.Bool(base.GetEncipherOnly()),
			DecipherOnly:      pulumi.Bool(base.GetDecipherOnly()),
		},
		ExtendedKeyUsage: &certificateauthority.AuthorityConfigX509ConfigKeyUsageExtendedKeyUsageArgs{
			ServerAuth:      pulumi.Bool(extended.GetServerAuth()),
			ClientAuth:      pulumi.Bool(extended.GetClientAuth()),
			CodeSigning:     pulumi.Bool(extended.GetCodeSigning()),
			EmailProtection: pulumi.Bool(extended.GetEmailProtection()),
			TimeStamping:    pulumi.Bool(extended.GetTimeStamping()),
			OcspSigning:     pulumi.Bool(extended.GetOcspSigning()),
		},
	}
	if len(k.GetUnknownExtendedKeyUsages()) > 0 {
		unknown := certificateauthority.AuthorityConfigX509ConfigKeyUsageUnknownExtendedKeyUsageArray{}
		for _, oid := range k.GetUnknownExtendedKeyUsages() {
			unknown = append(unknown, &certificateauthority.AuthorityConfigX509ConfigKeyUsageUnknownExtendedKeyUsageArgs{ObjectIdPaths: objectIdPath(oid)})
		}
		args.UnknownExtendedKeyUsages = unknown
	}
	return args
}

// objectIdPath converts an OID's arcs to the provider's integer list.
func objectIdPath(oid *gcpprivatecacertificateauthorityv1alpha1.GcpPrivateCaCertificateAuthorityObjectId) pulumi.IntArray {
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

// subjectConfig maps the subject and subject alternative names.
func subjectConfig(s *gcpprivatecacertificateauthorityv1alpha1.GcpPrivateCaCertificateAuthoritySubjectConfig) *certificateauthority.AuthorityConfigSubjectConfigArgs {
	subject := s.GetSubject()
	args := &certificateauthority.AuthorityConfigSubjectConfigArgs{
		Subject: &certificateauthority.AuthorityConfigSubjectConfigSubjectArgs{
			CommonName:         pulumi.String(subject.GetCommonName()),
			CountryCode:        stringPtr(subject.GetCountryCode()),
			Organization:       stringPtr(subject.GetOrganization()),
			OrganizationalUnit: stringPtr(subject.GetOrganizationalUnit()),
			Locality:           stringPtr(subject.GetLocality()),
			Province:           stringPtr(subject.GetProvince()),
			StreetAddress:      stringPtr(subject.GetStreetAddress()),
			PostalCode:         stringPtr(subject.GetPostalCode()),
		},
	}
	if san := s.GetSubjectAltName(); san != nil {
		args.SubjectAltName = &certificateauthority.AuthorityConfigSubjectConfigSubjectAltNameArgs{
			DnsNames:       stringArray(san.GetDnsNames()),
			Uris:           stringArray(san.GetUris()),
			EmailAddresses: stringArray(san.GetEmailAddresses()),
			IpAddresses:    stringArray(san.GetIpAddresses()),
		}
	}
	return args
}
