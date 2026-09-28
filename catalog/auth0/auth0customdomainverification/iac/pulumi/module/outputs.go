package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the verified custom domain. The domain's name is read
// back from Auth0 by id once verification has finished, so the name exported is
// one Auth0 has verified.
func exportOutputs(ctx *pulumi.Context, verification *auth0.CustomDomainVerification, provider *auth0.Provider) error {
	verified := auth0.LookupCustomDomainOutput(ctx, auth0.LookupCustomDomainOutputArgs{
		CustomDomainId: verification.CustomDomainId,
	}, pulumi.Provider(provider))

	ctx.Export("custom_domain_id", verification.CustomDomainId)
	ctx.Export("domain", verified.Domain())
	ctx.Export("origin_domain_name", verification.OriginDomainName)
	ctx.Export("cname_api_key", pulumi.ToSecret(verification.CnameApiKey))

	return nil
}
