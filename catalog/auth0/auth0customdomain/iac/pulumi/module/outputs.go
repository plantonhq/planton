package module

import (
	"strings"

	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsRecord is the DNS record that proves control of a custom domain.
type dnsRecord struct {
	Name  string
	Type  string
	Value string
}

// verificationRecord picks the record to publish from the verification methods
// Auth0 returns. The CNAME is preferred when offered -- an Auth0-managed domain
// is served through it, so it is the one record that both proves control and
// carries traffic -- and otherwise the first method (the TXT record of a
// self-managed domain). A method's "domain" names a TXT record; a CNAME's name
// is the custom domain itself. The Terraform module's locals.tf applies the same
// rule -- keep them in lockstep.
func verificationRecord(verifications []auth0.CustomDomainVerificationType, domain string) dnsRecord {
	var methods []map[string]interface{}
	for _, verification := range verifications {
		for _, method := range verification.Methods {
			if m, ok := method.(map[string]interface{}); ok {
				methods = append(methods, m)
			}
		}
	}
	if len(methods) == 0 {
		return dnsRecord{}
	}

	chosen := methods[0]
	for _, m := range methods {
		if strings.EqualFold(stringOf(m, "name"), "cname") {
			chosen = m
			break
		}
	}

	name := stringOf(chosen, "domain")
	if name == "" {
		name = domain
	}
	return dnsRecord{
		Name:  name,
		Type:  strings.ToUpper(stringOf(chosen, "name")),
		Value: stringOf(chosen, "record"),
	}
}

// stringOf reads one string value of a verification method.
func stringOf(method map[string]interface{}, key string) string {
	if value, ok := method[key].(string); ok {
		return value
	}
	return ""
}

// exportOutputs exports the custom domain as Auth0 created it, and the DNS record
// that proves control of it.
func exportOutputs(ctx *pulumi.Context, customDomain *auth0.CustomDomain) error {
	// Each record output is derived straight from the provider's values: the
	// output of pulumi.All is untyped, so chaining a typed applier onto it
	// panics at run time.
	recordField := func(pick func(dnsRecord) string) pulumi.StringOutput {
		return pulumi.All(customDomain.Verifications, customDomain.Domain).ApplyT(func(args []interface{}) string {
			return pick(verificationRecord(args[0].([]auth0.CustomDomainVerificationType), args[1].(string)))
		}).(pulumi.StringOutput)
	}

	ctx.Export("id", customDomain.ID())
	ctx.Export("domain", customDomain.Domain)
	ctx.Export("status", customDomain.Status)
	ctx.Export("origin_domain_name", customDomain.OriginDomainName)
	ctx.Export("dns_record_name", recordField(func(r dnsRecord) string { return r.Name }))
	ctx.Export("dns_record_type", recordField(func(r dnsRecord) string { return r.Type }))
	ctx.Export("dns_record_value", recordField(func(r dnsRecord) string { return r.Value }))

	return nil
}
