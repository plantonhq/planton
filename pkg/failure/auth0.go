package failure

import (
	"fmt"
	"regexp"
)

// signatureAuth0OrganizationsEntitlement is the fragment of
// Auth0OrganizationsEntitlement's observation Explain looks for before
// speaking.
const signatureAuth0OrganizationsEntitlement = "because the tenant's plan does not include machine-to-machine access to Organizations"

// Auth0OrganizationsEntitlement: Auth0 refused a client grant on a tenant
// whose plan lacks machine-to-machine access to Organizations (every Free
// and Essentials tenant). Auth0 reads the PRESENCE of either organization
// attribute on a grant as use of the feature, so the refusal names the plan
// and never the field; the field is what the reader can change. resource is
// the grant's OpenTofu address when the output carries one ("" otherwise:
// the runner's one-line summary names only the Auth0Client); raw
// is Auth0's text, kept whole.
//
// Both catalog modules send allow_any_organization only when it is true and
// organization_usage only when it is set, so a manifest meets this refusal
// only when it asked for organization access on a grant.
func Auth0OrganizationsEntitlement(resource, raw string) *Failure {
	grant := "a client grant"
	if resource != "" {
		grant = "the client grant " + resource
	}
	return &Failure{
		Observed: fmt.Sprintf("Auth0 refused %s %s (Auth0 answered: %s)", grant, signatureAuth0OrganizationsEntitlement, raw),
		Meaning:  "the manifest asks for organization access on an API grant: spec.apiGrants[].allowAnyOrganization is true or spec.apiGrants[].organizationUsage is set, and Auth0 treats either on a grant as use of Organizations for machine-to-machine access, a feature of paid plans",
		NextStep: "remove allowAnyOrganization and organizationUsage from every entry of spec.apiGrants and re-apply, or upgrade the Auth0 tenant's plan to one that includes machine-to-machine access to Organizations and re-apply",
	}
}

var (
	// 403 Forbidden: Please upgrade your subscription to use Machine to Machine access to Organizations.
	auth0OrganizationsEntitlementPattern = regexp.MustCompile(`(?:\d{3} Forbidden: )?Please upgrade your subscription to use Machine to Machine access to Organizations\.?`)
	// OpenTofu names the failing resource on the diagnostic's "with" line:
	// with auth0_client_grant.api_grants["0"],
	tofuAuth0ClientGrantAddressPattern = regexp.MustCompile(`with (auth0_client_grant\.[\w-]+(?:\["[^"]*"\]|\[\d+\])?),`)
)

func explainAuth0OrganizationsEntitlement(text string) *Failure {
	raw := auth0OrganizationsEntitlementPattern.FindString(text)
	if raw == "" {
		return nil
	}
	resource := ""
	if m := tofuAuth0ClientGrantAddressPattern.FindStringSubmatch(text); m != nil {
		resource = m[1]
	}
	return Auth0OrganizationsEntitlement(resource, raw)
}
