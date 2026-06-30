package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createConnection creates an Auth0 connection based on the configuration
func createConnection(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.Connection, error) {
	// Build connection arguments
	connectionArgs := &auth0.ConnectionArgs{
		Name:               pulumi.String(locals.ConnectionName),
		Strategy:           pulumi.String(locals.Strategy),
		DisplayName:        pulumi.String(locals.DisplayName),
		IsDomainConnection: pulumi.Bool(locals.IsDomainConnection),
		ShowAsButton:       pulumi.Bool(locals.ShowAsButton),
	}

	// Add realms if specified
	if len(locals.Realms) > 0 {
		realmsArray := pulumi.StringArray{}
		for _, realm := range locals.Realms {
			realmsArray = append(realmsArray, pulumi.String(realm))
		}
		connectionArgs.Realms = realmsArray
	}

	// Add metadata if specified
	if len(locals.Metadata) > 0 {
		metadataMap := pulumi.StringMap{}
		for k, v := range locals.Metadata {
			metadataMap[k] = pulumi.String(v)
		}
		connectionArgs.Metadata = metadataMap
	}

	// Build options based on strategy
	options := buildConnectionOptions(locals)
	if options != nil {
		connectionArgs.Options = options
	}

	// Create the connection resource
	connection, err := auth0.NewConnection(ctx, locals.ConnectionName, connectionArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Auth0 connection %s", locals.ConnectionName)
	}

	// Enable clients for this connection if specified
	if len(locals.EnabledClients) > 0 {
		clientsArray := pulumi.StringArray{}
		for _, client := range locals.EnabledClients {
			clientsArray = append(clientsArray, pulumi.String(client))
		}

		_, err = auth0.NewConnectionClients(ctx, locals.ConnectionName+"-clients", &auth0.ConnectionClientsArgs{
			ConnectionId:   connection.ID(),
			EnabledClients: clientsArray,
		}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{connection}))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to enable clients for Auth0 connection %s", locals.ConnectionName)
		}
	}

	return connection, nil
}

// buildConnectionOptions builds the strategy-specific options for the connection
func buildConnectionOptions(locals *Locals) *auth0.ConnectionOptionsArgs {
	options := &auth0.ConnectionOptionsArgs{}
	hasOptions := false

	switch locals.Strategy {
	case "auth0":
		// Database connection options
		if locals.DatabaseOptions != nil {
			hasOptions = true
			db := locals.DatabaseOptions

			if db.PasswordPolicy != "" {
				options.PasswordPolicy = pulumi.String(db.PasswordPolicy)
			}

			options.RequiresUsername = pulumi.Bool(db.RequiresUsername)
			options.DisableSignup = pulumi.Bool(db.DisableSignup)
			options.BruteForceProtection = pulumi.Bool(db.BruteForceProtection)

			if db.PasswordHistorySize > 0 {
				options.PasswordHistories = auth0.ConnectionOptionsPasswordHistoryArray{
					&auth0.ConnectionOptionsPasswordHistoryArgs{
						Enable: pulumi.Bool(true),
						Size:   pulumi.Int(int(db.PasswordHistorySize)),
					},
				}
			}

			if db.PasswordNoPersonalInfo {
				options.PasswordNoPersonalInfo = &auth0.ConnectionOptionsPasswordNoPersonalInfoArgs{
					Enable: pulumi.Bool(true),
				}
			}

			if db.PasswordDictionary {
				options.PasswordDictionary = &auth0.ConnectionOptionsPasswordDictionaryArgs{
					Enable: pulumi.Bool(true),
				}
			}

			if db.MfaEnabled {
				options.Mfa = &auth0.ConnectionOptionsMfaArgs{
					Active:               pulumi.Bool(true),
					ReturnEnrollSettings: pulumi.Bool(true),
				}
			}
		}

	case "google-oauth2", "facebook", "github", "linkedin", "twitter", "microsoft-account", "apple":
		// Social connection options
		if locals.SocialOptions != nil {
			hasOptions = true
			social := locals.SocialOptions

			options.ClientId = pulumi.String(social.ClientId)
			options.ClientSecret = pulumi.String(social.ClientSecret)

			if len(social.Scopes) > 0 {
				scopesArray := pulumi.StringArray{}
				for _, scope := range social.Scopes {
					scopesArray = append(scopesArray, pulumi.String(scope))
				}
				options.Scopes = scopesArray
			}

			if len(social.AllowedAudiences) > 0 {
				audiencesArray := pulumi.StringArray{}
				for _, audience := range social.AllowedAudiences {
					audiencesArray = append(audiencesArray, pulumi.String(audience))
				}
				options.AllowedAudiences = audiencesArray
			}

			if len(social.UpstreamParams) > 0 {
				// Convert map to JSON string for upstream params
				jsonBytes, _ := json.Marshal(social.UpstreamParams)
				options.UpstreamParams = pulumi.String(string(jsonBytes))
			}
		}

	case "samlp":
		// SAML connection options
		if locals.SamlOptions != nil {
			hasOptions = true
			saml := locals.SamlOptions

			options.SignInEndpoint = pulumi.String(saml.SignInEndpoint)
			options.SigningCert = pulumi.String(saml.SigningCert)

			if saml.SignOutEndpoint != "" {
				options.SignOutEndpoint = pulumi.String(saml.SignOutEndpoint)
			}

			if saml.EntityId != "" {
				options.EntityId = pulumi.String(saml.EntityId)
			}

			if saml.ProtocolBinding != "" {
				options.ProtocolBinding = pulumi.String(saml.ProtocolBinding)
			}

			options.SignSamlRequest = pulumi.Bool(saml.SignRequest)

			if saml.SignatureAlgorithm != "" {
				options.SignatureAlgorithm = pulumi.String(saml.SignatureAlgorithm)
			}

			if saml.DigestAlgorithm != "" {
				options.DigestAlgorithm = pulumi.String(saml.DigestAlgorithm)
			}

			if len(saml.AttributeMappings) > 0 {
				// Convert attribute mappings to JSON string
				jsonBytes, _ := json.Marshal(saml.AttributeMappings)
				options.FieldsMap = pulumi.String(string(jsonBytes))
			}
		}

	case "oidc":
		// OIDC connection options
		if locals.OidcOptions != nil {
			hasOptions = true
			oidc := locals.OidcOptions

			options.Issuer = pulumi.String(oidc.Issuer)
			options.ClientId = pulumi.String(oidc.ClientId)

			if oidc.ClientSecret != "" {
				options.ClientSecret = pulumi.String(oidc.ClientSecret)
			}

			if len(oidc.Scopes) > 0 {
				scopesArray := pulumi.StringArray{}
				for _, scope := range oidc.Scopes {
					scopesArray = append(scopesArray, pulumi.String(scope))
				}
				options.Scopes = scopesArray
			}

			if oidc.Type != "" {
				options.Type = pulumi.String(oidc.Type)
			}

			if oidc.AuthorizationEndpoint != "" {
				options.AuthorizationEndpoint = pulumi.String(oidc.AuthorizationEndpoint)
			}

			if oidc.TokenEndpoint != "" {
				options.TokenEndpoint = pulumi.String(oidc.TokenEndpoint)
			}

			if oidc.JwksUri != "" {
				options.JwksUri = pulumi.String(oidc.JwksUri)
			}
		}

	case "waad":
		// Azure AD connection options
		if locals.AzureAdOptions != nil {
			hasOptions = true
			azuread := locals.AzureAdOptions

			options.ClientId = pulumi.String(azuread.ClientId)
			options.ClientSecret = pulumi.String(azuread.ClientSecret)
			options.Domain = pulumi.String(azuread.Domain)

			if azuread.TenantId != "" {
				options.TenantDomain = pulumi.String(azuread.TenantId)
			}

			if azuread.MaxGroupsToRetrieve > 0 {
				options.MaxGroupsToRetrieve = pulumi.Sprintf("%d", azuread.MaxGroupsToRetrieve)
			}

			options.ApiEnableUsers = pulumi.Bool(azuread.ApiEnableUsers)
		}
	}

	if hasOptions {
		return options
	}
	return nil
}
