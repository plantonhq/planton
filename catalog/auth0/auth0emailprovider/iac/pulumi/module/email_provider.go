package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applyEmailProvider manages the ONE email provider of the tenant the
// provider's credential belongs to -- the twin of iac/tf/main.tf. The service
// arm the spec sets names it (locals.Name) and fills exactly the credentials
// and settings that service uses; the custom arm sends the empty credentials
// block the provider requires, and the Action bound to the tenant's
// custom-email-provider trigger does the sending. When the tenant already has
// a provider, the create takes it over and rewrites it. Every credential is
// sent as a Pulumi secret. Destroy deletes the provider, and the tenant falls
// back to Auth0's built-in test provider.
func applyEmailProvider(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.EmailProvider, error) {
	creds := locals.Credentials
	args := &auth0.EmailProviderArgs{
		Name:               pulumi.String(locals.Name),
		Enabled:            pulumi.Bool(locals.Enabled),
		DefaultFromAddress: pulumi.String(locals.DefaultFromAddress),
		Credentials: &auth0.EmailProviderCredentialsArgs{
			SmtpHost:                pulumi.StringPtrFromPtr(creds.SmtpHost),
			SmtpPort:                pulumi.IntPtrFromPtr(creds.SmtpPort),
			SmtpUser:                pulumi.StringPtrFromPtr(creds.SmtpUser),
			SmtpPass:                secret(creds.SmtpPass),
			AccessKeyId:             secret(creds.AccessKeyId),
			SecretAccessKey:         secret(creds.SecretAccessKey),
			ApiKey:                  secret(creds.ApiKey),
			Region:                  pulumi.StringPtrFromPtr(creds.Region),
			Domain:                  pulumi.StringPtrFromPtr(creds.Domain),
			AzureCsConnectionString: secret(creds.AzureCsConnectionString),
			Ms365TenantId:           secret(creds.Ms365TenantId),
			Ms365ClientId:           secret(creds.Ms365ClientId),
			Ms365ClientSecret:       secret(creds.Ms365ClientSecret),
		},
	}

	// The settings block is declared only when it carries headers or a
	// message; otherwise the provider keeps what Auth0 reports.
	if locals.Headers != nil || locals.Message != nil {
		settings := &auth0.EmailProviderSettingsArgs{}
		if locals.Headers != nil {
			settings.Headers = &auth0.EmailProviderSettingsHeadersArgs{
				XMcViewContentLink:   pulumi.StringPtrFromPtr(locals.Headers.XMcViewContentLink),
				XSesConfigurationSet: pulumi.StringPtrFromPtr(locals.Headers.XSesConfigurationSet),
			}
		}
		if locals.Message != nil {
			settings.Message = &auth0.EmailProviderSettingsMessageArgs{
				ConfigurationSetName: pulumi.StringPtrFromPtr(locals.Message.ConfigurationSetName),
				ViewContentLink:      pulumi.BoolPtrFromPtr(locals.Message.ViewContentLink),
			}
		}
		args.Settings = settings
	}

	emailProvider, err := auth0.NewEmailProvider(ctx, locals.ResourceName, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to apply the %s email provider of the Auth0 tenant for %s", locals.Name, locals.ResourceName)
	}
	return emailProvider, nil
}

// secret sends a credential as a Pulumi secret, and nothing when the arm does
// not carry it.
func secret(value *string) pulumi.StringPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.ToSecret(pulumi.String(*value)).(pulumi.StringOutput)
}
