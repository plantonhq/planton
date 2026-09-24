package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpdialogflowcxagentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdialogflowcxagent/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/diagflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// webhooks creates one webhook per spec.webhooks[] entry, keyed by display
// name (Google requires it unique within the agent), and exports their
// names in manifest order.
func webhooks(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, createdAgent *diagflow.CxAgent) error {
	resourceName := locals.GcpDialogflowCxAgent.Metadata.Name

	names := pulumi.StringArray{}
	for _, webhook := range locals.GcpDialogflowCxAgent.Spec.Webhooks {
		args := &diagflow.CxWebhookArgs{
			Parent:         createdAgent.ID().ToStringOutput(),
			DisplayName:    pulumi.String(webhook.DisplayName),
			Disabled:       optionalTrue(webhook.Disabled),
			Timeout:        optionalString(webhook.Timeout),
			DeletionPolicy: deletionPolicy(locals),
		}
		if service := webhook.GenericWebService; service != nil {
			args.GenericWebService = genericWebServiceArgs(service)
		}
		if directory := webhook.ServiceDirectory; directory != nil {
			directoryArgs := &diagflow.CxWebhookServiceDirectoryArgs{
				Service: pulumi.String(directory.Service),
			}
			if service := directory.GenericWebService; service != nil {
				directoryArgs.GenericWebService = serviceDirectoryGenericWebServiceArgs(service)
			}
			args.ServiceDirectory = directoryArgs
		}

		created, err := diagflow.NewCxWebhook(ctx,
			fmt.Sprintf("%s-webhook-%s", resourceName, webhook.DisplayName), args,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdAgent))
		if err != nil {
			return errors.Wrapf(err, "failed to create webhook %s", webhook.DisplayName)
		}
		names = append(names, created.ID().ToStringOutput())
	}

	ctx.Export(OpWebhookNames, names.ToStringArrayOutput())
	return nil
}

// genericWebServiceArgs maps a directly called endpoint. The SDK types the
// Service Directory endpoint separately (serviceDirectoryGenericWebServiceArgs
// below) with the same fields, so the two builders mirror each other.
func genericWebServiceArgs(service *gcpdialogflowcxagentv1alpha1.GcpDialogflowCxAgentGenericWebService) *diagflow.CxWebhookGenericWebServiceArgs {
	args := &diagflow.CxWebhookGenericWebServiceArgs{
		Uri:                              pulumi.String(service.Uri),
		WebhookType:                      optionalString(service.WebhookType),
		HttpMethod:                       optionalString(service.HttpMethod),
		RequestBody:                      optionalString(service.RequestBody),
		ParameterMapping:                 optionalStringMap(service.ParameterMapping),
		RequestHeaders:                   optionalStringMap(service.RequestHeaders),
		SecretVersionForUsernamePassword: optionalString(service.SecretVersionForUsernamePassword),
		ServiceAgentAuth:                 optionalString(service.ServiceAgentAuth),
		AllowedCaCerts:                   optionalStringArray(service.AllowedCaCerts),
	}
	if len(service.SecretVersionsForRequestHeaders) > 0 {
		headers := diagflow.CxWebhookGenericWebServiceSecretVersionsForRequestHeaderArray{}
		for _, header := range service.SecretVersionsForRequestHeaders {
			headers = append(headers, &diagflow.CxWebhookGenericWebServiceSecretVersionsForRequestHeaderArgs{
				Key:           pulumi.String(header.Key),
				SecretVersion: pulumi.String(header.SecretVersion),
			})
		}
		args.SecretVersionsForRequestHeaders = headers
	}
	if oauth := service.OauthConfig; oauth != nil {
		args.OauthConfig = &diagflow.CxWebhookGenericWebServiceOauthConfigArgs{
			ClientId:                     pulumi.String(oauth.ClientId),
			TokenEndpoint:                pulumi.String(oauth.TokenEndpoint),
			ClientSecret:                 optionalSecret(oauth.ClientSecret),
			Scopes:                       optionalStringArray(oauth.Scopes),
			SecretVersionForClientSecret: optionalString(oauth.SecretVersionForClientSecret),
		}
	}
	// The spec lifts the block's one field.
	if service.ServiceAccount.GetValue() != "" {
		args.ServiceAccountAuthConfig = &diagflow.CxWebhookGenericWebServiceServiceAccountAuthConfigArgs{
			ServiceAccount: pulumi.String(service.ServiceAccount.GetValue()),
		}
	}
	return args
}

// serviceDirectoryGenericWebServiceArgs maps the endpoint behind a Service
// Directory service -- genericWebServiceArgs field for field.
func serviceDirectoryGenericWebServiceArgs(service *gcpdialogflowcxagentv1alpha1.GcpDialogflowCxAgentGenericWebService) *diagflow.CxWebhookServiceDirectoryGenericWebServiceArgs {
	args := &diagflow.CxWebhookServiceDirectoryGenericWebServiceArgs{
		Uri:                              pulumi.String(service.Uri),
		WebhookType:                      optionalString(service.WebhookType),
		HttpMethod:                       optionalString(service.HttpMethod),
		RequestBody:                      optionalString(service.RequestBody),
		ParameterMapping:                 optionalStringMap(service.ParameterMapping),
		RequestHeaders:                   optionalStringMap(service.RequestHeaders),
		SecretVersionForUsernamePassword: optionalString(service.SecretVersionForUsernamePassword),
		ServiceAgentAuth:                 optionalString(service.ServiceAgentAuth),
		AllowedCaCerts:                   optionalStringArray(service.AllowedCaCerts),
	}
	if len(service.SecretVersionsForRequestHeaders) > 0 {
		headers := diagflow.CxWebhookServiceDirectoryGenericWebServiceSecretVersionsForRequestHeaderArray{}
		for _, header := range service.SecretVersionsForRequestHeaders {
			headers = append(headers, &diagflow.CxWebhookServiceDirectoryGenericWebServiceSecretVersionsForRequestHeaderArgs{
				Key:           pulumi.String(header.Key),
				SecretVersion: pulumi.String(header.SecretVersion),
			})
		}
		args.SecretVersionsForRequestHeaders = headers
	}
	if oauth := service.OauthConfig; oauth != nil {
		args.OauthConfig = &diagflow.CxWebhookServiceDirectoryGenericWebServiceOauthConfigArgs{
			ClientId:                     pulumi.String(oauth.ClientId),
			TokenEndpoint:                pulumi.String(oauth.TokenEndpoint),
			ClientSecret:                 optionalSecret(oauth.ClientSecret),
			Scopes:                       optionalStringArray(oauth.Scopes),
			SecretVersionForClientSecret: optionalString(oauth.SecretVersionForClientSecret),
		}
	}
	if service.ServiceAccount.GetValue() != "" {
		args.ServiceAccountAuthConfig = &diagflow.CxWebhookServiceDirectoryGenericWebServiceServiceAccountAuthConfigArgs{
			ServiceAccount: pulumi.String(service.ServiceAccount.GetValue()),
		}
	}
	return args
}
