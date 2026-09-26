package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/iam"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// oauthClient provisions the workforce OAuth client plus its managed
// credentials.
//
// Each spec.credentials entry fans out to one OauthClientCredential whose
// secret GCP generates server-side — the FIRST credential's secret is the
// kind's client_secret output (provider-computed secrets arrive already
// secret-marked from the SDK, so the export needs no extra wrapping).
//
// credential.disabled is sent EXPLICITLY: GCP requires a credential to be
// DISABLED before it can be deleted, so a spec transition false -> true is
// exactly the pre-removal step and must reach the API (the
// send-true-or-omit class).
//
// The spec's ONE deletion_policy fans out to the client AND every
// credential — the credentials have no life apart from the client, so one
// switch governs all of them.
func oauthClient(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpIamOauthClient.Spec

	// Enable the IAM API, which serves the workforce OAuth surface.
	// disable_on_destroy stays false: tearing down one client must never
	// disable IAM for everything else in the project. Matches the
	// Terraform module.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("iam.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"client-iam.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable iam.googleapis.com api")
	}

	allowedRedirectUris := pulumi.StringArray{}
	for _, uri := range spec.AllowedRedirectUris {
		allowedRedirectUris = append(allowedRedirectUris, pulumi.String(uri.GetValue()))
	}

	args := &iam.OauthClientArgs{
		OauthClientId:       pulumi.String(locals.OauthClientId),
		Location:            pulumi.String(locals.Location),
		AllowedGrantTypes:   pulumi.ToStringArray(spec.AllowedGrantTypes),
		AllowedScopes:       pulumi.ToStringArray(spec.AllowedScopes),
		AllowedRedirectUris: allowedRedirectUris,
	}
	if spec.ClientType != "" {
		args.ClientType = pulumi.StringPtr(spec.ClientType)
	}
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.Disabled {
		args.Disabled = pulumi.BoolPtr(true)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	createdClient, err := iam.NewOauthClient(ctx, "oauth-client", args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create oauth client")
	}

	ctx.Export(OpClientId, createdClient.ClientId)
	ctx.Export(OpClientName, createdClient.Name)
	ctx.Export(OpState, createdClient.State)

	firstSecretExported := false
	for _, credential := range spec.Credentials {
		credentialArgs := &iam.OauthClientCredentialArgs{
			Oauthclient:             createdClient.OauthClientId,
			Location:                pulumi.String(locals.Location),
			OauthClientCredentialId: pulumi.String(credential.CredentialId),
			// Explicit send — see the function comment.
			Disabled: pulumi.Bool(credential.Disabled),
		}
		if credential.DisplayName != "" {
			credentialArgs.DisplayName = pulumi.StringPtr(credential.DisplayName)
		}
		if spec.DeletionPolicy != "" {
			credentialArgs.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
		}
		if spec.ProjectId.GetValue() != "" {
			credentialArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		createdCredential, err := iam.NewOauthClientCredential(ctx,
			"credential-"+credential.CredentialId, credentialArgs, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create oauth client credential %s", credential.CredentialId)
		}
		if !firstSecretExported {
			ctx.Export(OpClientSecret, pulumi.ToSecret(createdCredential.ClientSecret))
			firstSecretExported = true
		}
	}
	if !firstSecretExported {
		// The output key must exist either way so the outputs transformer
		// maps a stable shape; empty means "no credentials configured".
		ctx.Export(OpClientSecret, pulumi.ToSecret(pulumi.String("")))
	}

	return nil
}
