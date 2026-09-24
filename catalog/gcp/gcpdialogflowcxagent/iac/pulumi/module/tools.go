package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpdialogflowcxagentv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdialogflowcxagent/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/diagflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// toolDefinition is what a tool and a tool version's frozen snapshot share:
// exactly one of the three specifications.
type toolDefinition interface {
	GetOpenApiSpec() *gcpdialogflowcxagentv1alpha1.GcpDialogflowCxAgentToolOpenApiSpec
	GetDataStoreSpec() *gcpdialogflowcxagentv1alpha1.GcpDialogflowCxAgentToolDataStoreSpec
	GetFunctionSpec() *gcpdialogflowcxagentv1alpha1.GcpDialogflowCxAgentToolFunctionSpec
}

// tools creates one tool per spec.tools[] entry, keyed by display name
// (Google requires it unique within the agent), and one tool version per
// tools[].versions[] entry parented to its tool; it exports both name lists
// in manifest order.
func tools(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, createdAgent *diagflow.CxAgent) error {
	resourceName := locals.GcpDialogflowCxAgent.Metadata.Name

	toolNames := pulumi.StringArray{}
	toolVersionNames := pulumi.StringArray{}
	for _, tool := range locals.GcpDialogflowCxAgent.Spec.Tools {
		args := &diagflow.CxToolArgs{
			Parent:         createdAgent.ID().ToStringOutput(),
			DisplayName:    pulumi.String(tool.DisplayName),
			Description:    pulumi.String(tool.Description),
			DeletionPolicy: deletionPolicy(locals),
		}
		setToolSpecification(args, tool)

		createdTool, err := diagflow.NewCxTool(ctx,
			fmt.Sprintf("%s-tool-%s", resourceName, tool.DisplayName), args,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(createdAgent))
		if err != nil {
			return errors.Wrapf(err, "failed to create tool %s", tool.DisplayName)
		}
		toolNames = append(toolNames, createdTool.ID().ToStringOutput())

		// Every argument of a version is immutable: any change replaces it.
		for _, version := range tool.Versions {
			snapshot := version.Tool
			snapshotArgs := diagflow.CxToolVersionToolArgs{
				DisplayName: pulumi.String(snapshot.DisplayName),
				Description: pulumi.String(snapshot.Description),
			}
			setToolVersionSpecification(&snapshotArgs, snapshot)

			createdVersion, err := diagflow.NewCxToolVersion(ctx,
				fmt.Sprintf("%s-tool-%s-version-%s", resourceName, tool.DisplayName, version.DisplayName),
				&diagflow.CxToolVersionArgs{
					Parent:         createdTool.ID().ToStringOutput(),
					DisplayName:    pulumi.String(version.DisplayName),
					Tool:           snapshotArgs,
					DeletionPolicy: deletionPolicy(locals),
				},
				pulumi.Provider(gcpProvider),
				pulumi.Parent(createdTool))
			if err != nil {
				return errors.Wrapf(err, "failed to create tool version %s/%s", tool.DisplayName, version.DisplayName)
			}
			toolVersionNames = append(toolVersionNames, createdVersion.ID().ToStringOutput())
		}
	}

	ctx.Export(OpToolNames, toolNames.ToStringArrayOutput())
	ctx.Export(OpToolVersionNames, toolVersionNames.ToStringArrayOutput())
	return nil
}

// setToolSpecification maps a tool's one specification onto the tool's
// Args. setToolVersionSpecification below mirrors it field for field onto
// the separately typed tool-version snapshot.
func setToolSpecification(args *diagflow.CxToolArgs, definition toolDefinition) {
	if spec := definition.GetOpenApiSpec(); spec != nil {
		openApi := &diagflow.CxToolOpenApiSpecArgs{
			TextSchema: pulumi.String(spec.TextSchema),
		}
		if auth := spec.Authentication; auth != nil {
			authArgs := &diagflow.CxToolOpenApiSpecAuthenticationArgs{}
			if apiKey := auth.ApiKeyConfig; apiKey != nil {
				authArgs.ApiKeyConfig = &diagflow.CxToolOpenApiSpecAuthenticationApiKeyConfigArgs{
					KeyName:                pulumi.String(apiKey.KeyName),
					RequestLocation:        pulumi.String(apiKey.RequestLocation),
					ApiKey:                 optionalSecret(apiKey.ApiKey),
					SecretVersionForApiKey: optionalString(apiKey.SecretVersionForApiKey),
				}
			}
			if bearer := auth.BearerTokenConfig; bearer != nil {
				authArgs.BearerTokenConfig = &diagflow.CxToolOpenApiSpecAuthenticationBearerTokenConfigArgs{
					Token:                 optionalSecret(bearer.Token),
					SecretVersionForToken: optionalString(bearer.SecretVersionForToken),
				}
			}
			if oauth := auth.OauthConfig; oauth != nil {
				authArgs.OauthConfig = &diagflow.CxToolOpenApiSpecAuthenticationOauthConfigArgs{
					ClientId:                     pulumi.String(oauth.ClientId),
					OauthGrantType:               pulumi.String(oauth.OauthGrantType),
					TokenEndpoint:                pulumi.String(oauth.TokenEndpoint),
					ClientSecret:                 optionalSecret(oauth.ClientSecret),
					Scopes:                       optionalStringArray(oauth.Scopes),
					SecretVersionForClientSecret: optionalString(oauth.SecretVersionForClientSecret),
				}
			}
			if serviceAgent := auth.ServiceAgentAuthConfig; serviceAgent != nil {
				authArgs.ServiceAgentAuthConfig = &diagflow.CxToolOpenApiSpecAuthenticationServiceAgentAuthConfigArgs{
					ServiceAgentAuth: optionalString(serviceAgent.ServiceAgentAuth),
				}
			}
			openApi.Authentication = authArgs
		}
		if directory := spec.ServiceDirectoryConfig; directory != nil {
			openApi.ServiceDirectoryConfig = &diagflow.CxToolOpenApiSpecServiceDirectoryConfigArgs{
				Service: pulumi.String(directory.Service),
			}
		}
		if tls := spec.TlsConfig; tls != nil {
			certs := diagflow.CxToolOpenApiSpecTlsConfigCaCertArray{}
			for _, cert := range tls.CaCerts {
				certs = append(certs, &diagflow.CxToolOpenApiSpecTlsConfigCaCertArgs{
					DisplayName: pulumi.String(cert.DisplayName),
					Cert:        pulumi.String(cert.Cert),
				})
			}
			openApi.TlsConfig = &diagflow.CxToolOpenApiSpecTlsConfigArgs{CaCerts: certs}
		}
		args.OpenApiSpec = openApi
	}
	if spec := definition.GetDataStoreSpec(); spec != nil {
		connections := diagflow.CxToolDataStoreSpecDataStoreConnectionArray{}
		for _, connection := range spec.DataStoreConnections {
			connections = append(connections, &diagflow.CxToolDataStoreSpecDataStoreConnectionArgs{
				DataStore:              optionalString(connection.DataStore.GetValue()),
				DataStoreType:          optionalString(connection.DataStoreType),
				DocumentProcessingMode: optionalString(connection.DocumentProcessingMode),
			})
		}
		args.DataStoreSpec = &diagflow.CxToolDataStoreSpecArgs{
			DataStoreConnections: connections,
			// Required by Google and carries no settings.
			FallbackPrompt: diagflow.CxToolDataStoreSpecFallbackPromptArgs{},
		}
	}
	if spec := definition.GetFunctionSpec(); spec != nil {
		args.FunctionSpec = &diagflow.CxToolFunctionSpecArgs{
			InputSchema:  optionalString(spec.InputSchema),
			OutputSchema: optionalString(spec.OutputSchema),
		}
	}
}

// setToolVersionSpecification maps a frozen snapshot's one specification --
// setToolSpecification field for field.
func setToolVersionSpecification(args *diagflow.CxToolVersionToolArgs, definition toolDefinition) {
	if spec := definition.GetOpenApiSpec(); spec != nil {
		openApi := &diagflow.CxToolVersionToolOpenApiSpecArgs{
			TextSchema: pulumi.String(spec.TextSchema),
		}
		if auth := spec.Authentication; auth != nil {
			authArgs := &diagflow.CxToolVersionToolOpenApiSpecAuthenticationArgs{}
			if apiKey := auth.ApiKeyConfig; apiKey != nil {
				authArgs.ApiKeyConfig = &diagflow.CxToolVersionToolOpenApiSpecAuthenticationApiKeyConfigArgs{
					KeyName:                pulumi.String(apiKey.KeyName),
					RequestLocation:        pulumi.String(apiKey.RequestLocation),
					ApiKey:                 optionalSecret(apiKey.ApiKey),
					SecretVersionForApiKey: optionalString(apiKey.SecretVersionForApiKey),
				}
			}
			if bearer := auth.BearerTokenConfig; bearer != nil {
				authArgs.BearerTokenConfig = &diagflow.CxToolVersionToolOpenApiSpecAuthenticationBearerTokenConfigArgs{
					Token:                 optionalSecret(bearer.Token),
					SecretVersionForToken: optionalString(bearer.SecretVersionForToken),
				}
			}
			if oauth := auth.OauthConfig; oauth != nil {
				authArgs.OauthConfig = &diagflow.CxToolVersionToolOpenApiSpecAuthenticationOauthConfigArgs{
					ClientId:                     pulumi.String(oauth.ClientId),
					OauthGrantType:               pulumi.String(oauth.OauthGrantType),
					TokenEndpoint:                pulumi.String(oauth.TokenEndpoint),
					ClientSecret:                 optionalSecret(oauth.ClientSecret),
					Scopes:                       optionalStringArray(oauth.Scopes),
					SecretVersionForClientSecret: optionalString(oauth.SecretVersionForClientSecret),
				}
			}
			if serviceAgent := auth.ServiceAgentAuthConfig; serviceAgent != nil {
				authArgs.ServiceAgentAuthConfig = &diagflow.CxToolVersionToolOpenApiSpecAuthenticationServiceAgentAuthConfigArgs{
					ServiceAgentAuth: optionalString(serviceAgent.ServiceAgentAuth),
				}
			}
			openApi.Authentication = authArgs
		}
		if directory := spec.ServiceDirectoryConfig; directory != nil {
			openApi.ServiceDirectoryConfig = &diagflow.CxToolVersionToolOpenApiSpecServiceDirectoryConfigArgs{
				Service: pulumi.String(directory.Service),
			}
		}
		if tls := spec.TlsConfig; tls != nil {
			certs := diagflow.CxToolVersionToolOpenApiSpecTlsConfigCaCertArray{}
			for _, cert := range tls.CaCerts {
				certs = append(certs, &diagflow.CxToolVersionToolOpenApiSpecTlsConfigCaCertArgs{
					DisplayName: pulumi.String(cert.DisplayName),
					Cert:        pulumi.String(cert.Cert),
				})
			}
			openApi.TlsConfig = &diagflow.CxToolVersionToolOpenApiSpecTlsConfigArgs{CaCerts: certs}
		}
		args.OpenApiSpec = openApi
	}
	if spec := definition.GetDataStoreSpec(); spec != nil {
		connections := diagflow.CxToolVersionToolDataStoreSpecDataStoreConnectionArray{}
		for _, connection := range spec.DataStoreConnections {
			connections = append(connections, &diagflow.CxToolVersionToolDataStoreSpecDataStoreConnectionArgs{
				DataStore:              optionalString(connection.DataStore.GetValue()),
				DataStoreType:          optionalString(connection.DataStoreType),
				DocumentProcessingMode: optionalString(connection.DocumentProcessingMode),
			})
		}
		args.DataStoreSpec = &diagflow.CxToolVersionToolDataStoreSpecArgs{
			DataStoreConnections: connections,
			FallbackPrompt:       diagflow.CxToolVersionToolDataStoreSpecFallbackPromptArgs{},
		}
	}
	if spec := definition.GetFunctionSpec(); spec != nil {
		args.FunctionSpec = &diagflow.CxToolVersionToolFunctionSpecArgs{
			InputSchema:  optionalString(spec.InputSchema),
			OutputSchema: optionalString(spec.OutputSchema),
		}
	}
}
