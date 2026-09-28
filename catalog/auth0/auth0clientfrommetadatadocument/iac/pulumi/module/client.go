package module

import (
	"github.com/pkg/errors"
	auth0clientfrommetadatadocumentv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0clientfrommetadatadocument/v1alpha1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createClient registers the application from its metadata document.
//
// What the provider does with it, and why the module sends so little:
//   - Create POSTs the document's URL to /api/v2/clients/cimd/register. Auth0
//     fetches the document and registers the application from it (name,
//     redirect URIs, logo, keys, application type, grant types, description).
//     The registration is an upsert keyed by the URL, so a URL the tenant has
//     already registered is taken over rather than refused.
//   - The provider then PATCHes the client with the settings the spec declares
//     -- only those. Everything the spec leaves unset is omitted from the args
//     (clientArgs), so Auth0 keeps what the document or the tenant set, and an
//     adopted application previews no change for a setting it never declared.
//   - A change of external_client_id_version makes the provider register again
//     (Auth0 fetches the document anew) before the PATCH, so declared settings
//     win over the refreshed document every time.
//   - external_client_id is ForceNew: a new URL is a new application (a new
//     client_id). Destroy deletes the client.
//
// Every read also asks Auth0 to preview the document, which is where the
// validation outputs come from (outputs.go).
func createClient(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.ClientCimd, error) {
	client, err := auth0.NewClientCimd(ctx, locals.ResourceName, clientArgs(locals.Spec), pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to register the application from %s", locals.Spec.ExternalClientId)
	}
	return client, nil
}

// clientArgs maps the spec onto the provider's arguments, setting an argument
// only when the spec declares it: a nil pointer, a nil message, or an empty list
// or map is left out entirely (a nil Input), never sent as a zero value. The
// Terraform module's main.tf and locals.tf apply the same rule -- keep them in
// lockstep. A pure function of the spec, so its unset-means-unmanaged contract
// is tested without an engine (client_test.go).
func clientArgs(spec *auth0clientfrommetadatadocumentv1alpha1.Auth0ClientFromMetadataDocumentSpec) *auth0.ClientCimdArgs {
	args := &auth0.ClientCimdArgs{
		ExternalClientId:         pulumi.String(spec.ExternalClientId),
		ExternalClientIdVersion:  intPtr(spec.ExternalClientIdVersion),
		AppType:                  pulumi.StringPtrFromPtr(spec.AppType),
		Description:              pulumi.StringPtrFromPtr(spec.Description),
		OidcConformant:           pulumi.BoolPtrFromPtr(spec.OidcConformant),
		RequireProofOfPossession: pulumi.BoolPtrFromPtr(spec.RequireProofOfPossession),
		RedirectionPolicy:        pulumi.StringPtrFromPtr(spec.RedirectionPolicy),
	}
	args.SkipNonVerifiableCallbackUriConfirmationPrompt = pulumi.BoolPtrFromPtr(spec.SkipNonVerifiableCallbackUriConfirmationPrompt)

	// Lists and the map: empty is "not managed", so they are sent only when
	// they carry something. grant_types is computed by the provider (the
	// document seeds it), so leaving it out keeps the document's grants;
	// the others are not, which is why their spec comments ask an adopter to
	// declare the live value.
	if len(spec.GrantTypes) > 0 {
		args.GrantTypes = pulumi.ToStringArray(spec.GrantTypes)
	}
	if len(spec.AllowedOrigins) > 0 {
		args.AllowedOrigins = pulumi.ToStringArray(spec.AllowedOrigins)
	}
	if len(spec.WebOrigins) > 0 {
		args.WebOrigins = pulumi.ToStringArray(spec.WebOrigins)
	}
	if len(spec.OrganizationDiscoveryMethods) > 0 {
		args.OrganizationDiscoveryMethods = pulumi.ToStringArray(spec.OrganizationDiscoveryMethods)
	}
	if len(spec.ClientMetadata) > 0 {
		args.ClientMetadata = pulumi.ToStringMap(spec.ClientMetadata)
	}

	// Blocks: a declared message is one block, with only its declared fields.
	if org := spec.DefaultOrganization; org != nil {
		args.DefaultOrganization = &auth0.ClientCimdDefaultOrganizationArgs{
			OrganizationId: pulumi.String(org.OrganizationId),
			Flows:          pulumi.ToStringArray(org.Flows),
		}
	}
	if jwt := spec.JwtConfiguration; jwt != nil {
		args.JwtConfiguration = &auth0.ClientCimdJwtConfigurationArgs{
			Alg:               pulumi.StringPtrFromPtr(jwt.Alg),
			LifetimeInSeconds: intPtr(jwt.LifetimeInSeconds),
		}
	}
	if rt := spec.RefreshToken; rt != nil {
		args.RefreshToken = &auth0.ClientCimdRefreshTokenArgs{
			RotationType:              pulumi.StringPtrFromPtr(rt.RotationType),
			ExpirationType:            pulumi.StringPtrFromPtr(rt.ExpirationType),
			Leeway:                    intPtr(rt.Leeway),
			TokenLifetime:             intPtr(rt.TokenLifetime),
			InfiniteTokenLifetime:     pulumi.BoolPtrFromPtr(rt.InfiniteTokenLifetime),
			IdleTokenLifetime:         intPtr(rt.IdleTokenLifetime),
			InfiniteIdleTokenLifetime: pulumi.BoolPtrFromPtr(rt.InfiniteIdleTokenLifetime),
		}
	}
	if quota := spec.TokenQuota; quota != nil {
		// client_credentials is required inside the block (spec validation), so
		// it is always present here. An unset enforce is left to the provider,
		// which defaults it to true on both engines.
		credentials := quota.ClientCredentials
		args.TokenQuota = &auth0.ClientCimdTokenQuotaArgs{
			ClientCredentials: auth0.ClientCimdTokenQuotaClientCredentialsArgs{
				Enforce: pulumi.BoolPtrFromPtr(credentials.Enforce),
				PerDay:  intPtr(credentials.PerDay),
				PerHour: intPtr(credentials.PerHour),
			},
		}
	}

	return args
}

// intPtr carries an optional proto int32 into the provider's optional int,
// leaving it out (nil) when unset.
func intPtr(value *int32) pulumi.IntPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.IntPtr(int(*value))
}
