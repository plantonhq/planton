package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/bigquery"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// connection creates the BigQuery connection with its one declared arm.
func connection(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpBigQueryConnection.Spec
	resourceName := locals.GcpBigQueryConnection.Metadata.Name

	// Enable the BigQuery Connection API first so a fresh project works on
	// the first deploy. DisableOnDestroy stays false: tearing down one
	// connection must never disable the API for every other connection.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("bigqueryconnection.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpbqcn-bigqueryconnection.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable bigqueryconnection.googleapis.com api")
	}

	args := &bigquery.ConnectionArgs{
		ConnectionId: pulumi.String(locals.ConnectionId),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors. Empty optional
	// strings stay out of the payload.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Location != "" {
		args.Location = pulumi.String(spec.Location)
	}
	if spec.FriendlyName != "" {
		args.FriendlyName = pulumi.String(spec.FriendlyName)
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.KmsKeyName = pulumi.String(spec.KmsKeyName.GetValue())
	}

	// Exactly one arm is declared -- the spec's CEL rule guarantees it.
	// The spec lifts the access_role wrapper's one leaf.
	if aws := spec.Aws; aws != nil {
		args.Aws = &bigquery.ConnectionAwsArgs{
			AccessRole: &bigquery.ConnectionAwsAccessRoleArgs{IamRoleId: pulumi.String(aws.IamRoleId)},
		}
	}
	if azure := spec.Azure; azure != nil {
		azureArgs := &bigquery.ConnectionAzureArgs{CustomerTenantId: pulumi.String(azure.CustomerTenantId)}
		if azure.FederatedApplicationClientId != "" {
			azureArgs.FederatedApplicationClientId = pulumi.String(azure.FederatedApplicationClientId)
		}
		args.Azure = azureArgs
	}
	// Google's arm has no settings -- the spec's bool emits the empty
	// marker block, and Google answers with the service account it created.
	if spec.CloudResource {
		args.CloudResource = &bigquery.ConnectionCloudResourceArgs{}
	}
	// The flags are sent only when true, so Google's defaults stay in
	// charge otherwise -- the Terraform module's rule.
	if spanner := spec.CloudSpanner; spanner != nil {
		spannerArgs := &bigquery.ConnectionCloudSpannerArgs{Database: pulumi.String(spanner.Database)}
		if spanner.DatabaseRole != "" {
			spannerArgs.DatabaseRole = pulumi.String(spanner.DatabaseRole)
		}
		if spanner.UseParallelism {
			spannerArgs.UseParallelism = pulumi.Bool(true)
		}
		if spanner.UseDataBoost {
			spannerArgs.UseDataBoost = pulumi.Bool(true)
		}
		if spanner.MaxParallelism > 0 {
			spannerArgs.MaxParallelism = pulumi.Int(int(spanner.MaxParallelism))
		}
		args.CloudSpanner = spannerArgs
	}
	if cloudSql := spec.CloudSql; cloudSql != nil {
		args.CloudSql = &bigquery.ConnectionCloudSqlArgs{
			InstanceId: pulumi.String(cloudSql.InstanceId.GetValue()),
			Database:   pulumi.String(cloudSql.Database),
			Type:       pulumi.String(cloudSql.Type),
			Credential: &bigquery.ConnectionCloudSqlCredentialArgs{
				Username: pulumi.String(cloudSql.Credential.Username),
				Password: pulumi.ToSecret(pulumi.String(cloudSql.Credential.Password)).(pulumi.StringOutput),
			},
		}
	}
	// The spec lifts the authentication, password, endpoint, and network
	// wrappers; each nested block is sent only when its value is declared.
	if configuration := spec.Configuration; configuration != nil {
		assetArgs := &bigquery.ConnectionConfigurationAssetArgs{}
		if configuration.Asset.GetDatabase() != "" {
			assetArgs.Database = pulumi.String(configuration.Asset.GetDatabase())
		}
		if configuration.Asset.GetGoogleCloudResource() != "" {
			assetArgs.GoogleCloudResource = pulumi.String(configuration.Asset.GetGoogleCloudResource())
		}
		configurationArgs := &bigquery.ConnectionConfigurationArgs{
			ConnectorId: pulumi.String(configuration.ConnectorId),
			Asset:       assetArgs,
		}
		if usernamePassword := configuration.UsernamePassword; usernamePassword != nil {
			configurationArgs.Authentication = &bigquery.ConnectionConfigurationAuthenticationArgs{
				UsernamePassword: &bigquery.ConnectionConfigurationAuthenticationUsernamePasswordArgs{
					Username: pulumi.String(usernamePassword.Username),
					Password: &bigquery.ConnectionConfigurationAuthenticationUsernamePasswordPasswordArgs{
						Plaintext: pulumi.ToSecret(pulumi.String(usernamePassword.Password)).(pulumi.StringOutput),
					},
				},
			}
		}
		if configuration.HostPort != "" {
			configurationArgs.Endpoint = &bigquery.ConnectionConfigurationEndpointArgs{
				HostPort: pulumi.String(configuration.HostPort),
			}
		}
		if configuration.NetworkAttachment != "" {
			configurationArgs.Network = &bigquery.ConnectionConfigurationNetworkArgs{
				PrivateServiceConnect: &bigquery.ConnectionConfigurationNetworkPrivateServiceConnectArgs{
					NetworkAttachment: pulumi.String(configuration.NetworkAttachment),
				},
			}
		}
		args.Configuration = configurationArgs
	}
	// The spec lifts both one-leaf wrappers; each is sent only when set.
	if spark := spec.Spark; spark != nil {
		sparkArgs := &bigquery.ConnectionSparkArgs{}
		if spark.MetastoreService != "" {
			sparkArgs.MetastoreServiceConfig = &bigquery.ConnectionSparkMetastoreServiceConfigArgs{
				MetastoreService: pulumi.String(spark.MetastoreService),
			}
		}
		if spark.HistoryServerDataprocCluster != "" {
			sparkArgs.SparkHistoryServerConfig = &bigquery.ConnectionSparkSparkHistoryServerConfigArgs{
				DataprocCluster: pulumi.String(spark.HistoryServerDataprocCluster),
			}
		}
		args.Spark = sparkArgs
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := bigquery.NewConnection(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create bigquery connection")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpConnectionId, created.ConnectionId)
	ctx.Export(OpLocation, created.Location)

	// The Google-owned identities exist only for the declared arm; the
	// others export empty strings, as the Terraform module's try() does.
	ctx.Export(OpCloudResourceServiceAccountId, created.CloudResource.ApplyT(func(v *bigquery.ConnectionCloudResource) string {
		if v == nil || v.ServiceAccountId == nil {
			return ""
		}
		return *v.ServiceAccountId
	}).(pulumi.StringOutput))
	ctx.Export(OpSparkServiceAccountId, created.Spark.ApplyT(func(v *bigquery.ConnectionSpark) string {
		if v == nil || v.ServiceAccountId == nil {
			return ""
		}
		return *v.ServiceAccountId
	}).(pulumi.StringOutput))
	ctx.Export(OpCloudSqlServiceAccountId, created.CloudSql.ApplyT(func(v *bigquery.ConnectionCloudSql) string {
		if v == nil || v.ServiceAccountId == nil {
			return ""
		}
		return *v.ServiceAccountId
	}).(pulumi.StringOutput))
	ctx.Export(OpConnectorServiceAccount, created.Configuration.ApplyT(func(v *bigquery.ConnectionConfiguration) string {
		if v == nil || v.Authentication == nil || v.Authentication.ServiceAccount == nil {
			return ""
		}
		return *v.Authentication.ServiceAccount
	}).(pulumi.StringOutput))
	ctx.Export(OpAwsIdentity, created.Aws.ApplyT(func(v *bigquery.ConnectionAws) string {
		if v == nil || v.AccessRole.Identity == nil {
			return ""
		}
		return *v.AccessRole.Identity
	}).(pulumi.StringOutput))
	azureField := func(pick func(*bigquery.ConnectionAzure) *string) pulumi.StringOutput {
		return created.Azure.ApplyT(func(v *bigquery.ConnectionAzure) string {
			if v == nil || pick(v) == nil {
				return ""
			}
			return *pick(v)
		}).(pulumi.StringOutput)
	}
	ctx.Export(OpAzureIdentity, azureField(func(v *bigquery.ConnectionAzure) *string { return v.Identity }))
	ctx.Export(OpAzureApplication, azureField(func(v *bigquery.ConnectionAzure) *string { return v.Application }))
	ctx.Export(OpAzureClientId, azureField(func(v *bigquery.ConnectionAzure) *string { return v.ClientId }))
	ctx.Export(OpAzureObjectId, azureField(func(v *bigquery.ConnectionAzure) *string { return v.ObjectId }))
	ctx.Export(OpAzureRedirectUri, azureField(func(v *bigquery.ConnectionAzure) *string { return v.RedirectUri }))
	return nil
}
