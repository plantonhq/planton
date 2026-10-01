package module

import (
	"github.com/pkg/errors"
	auth0resourceserverv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0resourceserver/v1alpha1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// defaultForThirdPartyClients is the only group Auth0 applies a default grant
// to: every third-party application in the tenant.
const defaultForThirdPartyClients = "third_party_clients"

// createDefaultGrants declares one client grant per
// spec.third_party_client_default_grants entry -- the twin of iac/tf/main.tf's
// auth0_client_grant.third_party_client_default_grants. A default grant names
// no application (default_for replaces client_id): every third-party
// application gets its scopes on this API without a grant of its own. Each is
// keyed by its subject type, the key its id is exported and imported under.
// The grants follow the scopes, since a grant may only name scopes the API
// defines. Destroy deletes them.
func createDefaultGrants(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider,
	resourceServer *auth0.ResourceServer, scopes *auth0.ResourceServerScopes) (pulumi.StringMap, error) {
	ids := pulumi.StringMap{}
	if len(locals.DefaultGrants) == 0 {
		return ids, nil
	}

	dependsOn := []pulumi.Resource{resourceServer}
	if scopes != nil {
		dependsOn = append(dependsOn, scopes)
	}

	for _, grant := range locals.DefaultGrants {
		created, err := auth0.NewClientGrant(ctx,
			locals.ResourceName+"-default-grant-"+grant.SubjectType,
			defaultGrantArgs(grant, resourceServer.Identifier),
			pulumi.Provider(provider), pulumi.DependsOn(dependsOn))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create the %s default grant for third-party applications on %s",
				grant.SubjectType, locals.ResourceName)
		}
		ids[grant.SubjectType] = created.ID().ToStringOutput()
	}
	return ids, nil
}

// defaultGrantArgs builds one default grant's arguments. It is a pure function
// of the entry (locals_test.go pins it). scopes is sent unless allow_all_scopes
// is true (the provider refuses both); every other optional setting is sent
// only as declared.
func defaultGrantArgs(grant *auth0resourceserverv1alpha1.Auth0ResourceServerThirdPartyClientDefaultGrant,
	audience pulumi.StringInput) *auth0.ClientGrantArgs {
	args := &auth0.ClientGrantArgs{
		Audience:             audience,
		DefaultFor:           pulumi.String(defaultForThirdPartyClients),
		SubjectType:          pulumi.String(grant.SubjectType),
		AllowAllScopes:       pulumi.BoolPtrFromPtr(grant.AllowAllScopes),
		OrganizationUsage:    pulumi.StringPtrFromPtr(grant.OrganizationUsage),
		AllowAnyOrganization: pulumi.BoolPtrFromPtr(grant.AllowAnyOrganization),
	}
	if !grant.GetAllowAllScopes() {
		args.Scopes = pulumi.ToStringArray(grant.Scopes)
	}
	if len(grant.AuthorizationDetailsTypes) > 0 {
		args.AuthorizationDetailsTypes = pulumi.ToStringArray(grant.AuthorizationDetailsTypes)
	}
	return args
}
