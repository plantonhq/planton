package module

import (
	gcpprivatecapoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecapool/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificateauthority"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// x509Parameters maps the spec's X.509 parameters onto the provider's
// CaPoolIssuancePolicyBaselineValues block. The provider requires ca_options and key_usage whenever the baseline
// values are sent; an omitted one is sent empty, which states nothing.
func x509Parameters(p *gcpprivatecapoolv1alpha1.GcpPrivateCaPoolX509Parameters) *certificateauthority.CaPoolIssuancePolicyBaselineValuesArgs {
	args := &certificateauthority.CaPoolIssuancePolicyBaselineValuesArgs{
		CaOptions: caOptions(p.GetCaOptions()),
		KeyUsage:  keyUsage(p.GetKeyUsage()),
	}
	if len(p.GetAiaOcspServers()) > 0 {
		args.AiaOcspServers = pulumi.ToStringArray(p.GetAiaOcspServers())
	}
	if len(p.GetPolicyIds()) > 0 {
		policyIds := certificateauthority.CaPoolIssuancePolicyBaselineValuesPolicyIdArray{}
		for _, oid := range p.GetPolicyIds() {
			policyIds = append(policyIds, &certificateauthority.CaPoolIssuancePolicyBaselineValuesPolicyIdArgs{ObjectIdPaths: objectIdPath(oid)})
		}
		args.PolicyIds = policyIds
	}
	if len(p.GetAdditionalExtensions()) > 0 {
		extensions := certificateauthority.CaPoolIssuancePolicyBaselineValuesAdditionalExtensionArray{}
		for _, extension := range p.GetAdditionalExtensions() {
			extensions = append(extensions, &certificateauthority.CaPoolIssuancePolicyBaselineValuesAdditionalExtensionArgs{
				Critical: pulumi.Bool(extension.GetCritical()),
				ObjectId: &certificateauthority.CaPoolIssuancePolicyBaselineValuesAdditionalExtensionObjectIdArgs{ObjectIdPaths: objectIdPath(extension.GetObjectId())},
				Value:    pulumi.String(extension.GetValue()),
			})
		}
		args.AdditionalExtensions = extensions
	}
	if nc := p.GetNameConstraints(); nc != nil {
		args.NameConstraints = &certificateauthority.CaPoolIssuancePolicyBaselineValuesNameConstraintsArgs{
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
func caOptions(o *gcpprivatecapoolv1alpha1.GcpPrivateCaPoolCaOptions) *certificateauthority.CaPoolIssuancePolicyBaselineValuesCaOptionsArgs {
	args := &certificateauthority.CaPoolIssuancePolicyBaselineValuesCaOptionsArgs{}
	// Unset is_ca omits the CA flag; false is sent with non_ca, the
	// provider's way to state CA:FALSE rather than leave it out.
	if o.IsCa != nil {
		args.IsCa = pulumi.Bool(o.GetIsCa())
		if !o.GetIsCa() {
			args.NonCa = pulumi.Bool(true)
		}
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
func keyUsage(k *gcpprivatecapoolv1alpha1.GcpPrivateCaPoolKeyUsage) *certificateauthority.CaPoolIssuancePolicyBaselineValuesKeyUsageArgs {
	base := k.GetBaseKeyUsage()
	extended := k.GetExtendedKeyUsage()
	args := &certificateauthority.CaPoolIssuancePolicyBaselineValuesKeyUsageArgs{
		BaseKeyUsage: &certificateauthority.CaPoolIssuancePolicyBaselineValuesKeyUsageBaseKeyUsageArgs{
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
		ExtendedKeyUsage: &certificateauthority.CaPoolIssuancePolicyBaselineValuesKeyUsageExtendedKeyUsageArgs{
			ServerAuth:      pulumi.Bool(extended.GetServerAuth()),
			ClientAuth:      pulumi.Bool(extended.GetClientAuth()),
			CodeSigning:     pulumi.Bool(extended.GetCodeSigning()),
			EmailProtection: pulumi.Bool(extended.GetEmailProtection()),
			TimeStamping:    pulumi.Bool(extended.GetTimeStamping()),
			OcspSigning:     pulumi.Bool(extended.GetOcspSigning()),
		},
	}
	if len(k.GetUnknownExtendedKeyUsages()) > 0 {
		unknown := certificateauthority.CaPoolIssuancePolicyBaselineValuesKeyUsageUnknownExtendedKeyUsageArray{}
		for _, oid := range k.GetUnknownExtendedKeyUsages() {
			unknown = append(unknown, &certificateauthority.CaPoolIssuancePolicyBaselineValuesKeyUsageUnknownExtendedKeyUsageArgs{ObjectIdPaths: objectIdPath(oid)})
		}
		args.UnknownExtendedKeyUsages = unknown
	}
	return args
}

// objectIdPath converts an OID's arcs to the provider's integer list.
func objectIdPath(oid *gcpprivatecapoolv1alpha1.GcpPrivateCaPoolObjectId) pulumi.IntArray {
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
