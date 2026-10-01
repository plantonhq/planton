package module

import (
	"testing"

	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
)

func verifications(methods ...map[string]interface{}) []auth0.CustomDomainVerificationType {
	list := make([]interface{}, 0, len(methods))
	for _, m := range methods {
		list = append(list, m)
	}
	return []auth0.CustomDomainVerificationType{{Methods: list}}
}

func TestVerificationRecord(t *testing.T) {
	cases := []struct {
		name          string
		verifications []auth0.CustomDomainVerificationType
		want          dnsRecord
	}{
		{
			name: "an Auth0-managed domain publishes its CNAME at the domain's own name",
			verifications: verifications(map[string]interface{}{
				"name": "cname", "record": "example-cd-abc123.edge.tenants.eu.auth0.com",
			}),
			want: dnsRecord{Name: "id.example.com", Type: "CNAME", Value: "example-cd-abc123.edge.tenants.eu.auth0.com"},
		},
		{
			name: "the CNAME is preferred when Auth0 offers a TXT method beside it",
			verifications: verifications(
				map[string]interface{}{"name": "txt", "record": "auth0-domain-verification=xyz", "domain": "_cf-custom-hostname.id.example.com"},
				map[string]interface{}{"name": "cname", "record": "example-cd-abc123.edge.tenants.eu.auth0.com"},
			),
			want: dnsRecord{Name: "id.example.com", Type: "CNAME", Value: "example-cd-abc123.edge.tenants.eu.auth0.com"},
		},
		{
			name: "a self-managed domain publishes the TXT record at the name Auth0 assigns",
			verifications: verifications(map[string]interface{}{
				"name": "txt", "record": "auth0-domain-verification=xyz", "domain": "_cf-custom-hostname.id.example.com",
			}),
			want: dnsRecord{Name: "_cf-custom-hostname.id.example.com", Type: "TXT", Value: "auth0-domain-verification=xyz"},
		},
		{
			name:          "no verification methods yield no record",
			verifications: nil,
			want:          dnsRecord{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := verificationRecord(tc.verifications, "id.example.com"); got != tc.want {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
