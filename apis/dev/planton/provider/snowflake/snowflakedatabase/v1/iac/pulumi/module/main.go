package module

import (
	"github.com/pkg/errors"
	snowflakedatabasev1 "github.com/plantonhq/planton/apis/dev/planton/provider/snowflake/snowflakedatabase/v1"
	"github.com/pulumi/pulumi-snowflake/sdk/v2/go/snowflake"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates a Snowflake database with all configured parameters
func Resources(ctx *pulumi.Context, stackInput *snowflakedatabasev1.SnowflakeDatabaseStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Setup Snowflake provider with credentials from provider config
	var provider *snowflake.Provider
	var err error
	providerConfig := stackInput.ProviderConfig

	if providerConfig == nil {
		// Use default provider (assumes credentials from environment)
		provider, err = snowflake.NewProvider(ctx, "snowflake-provider", &snowflake.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default Snowflake provider")
		}
	} else {
		// Create provider with explicit credentials
		// Note: The proto config has 'account' and 'region' fields, but the Pulumi provider
		// expects 'AccountName'. We'll use 'account' directly as AccountName.
		// The 'region' field is not directly used by the provider but may be part of account identifier.
		provider, err = snowflake.NewProvider(ctx, "snowflake-provider", &snowflake.ProviderArgs{
			AccountName: pulumi.String(providerConfig.Account),
			User:        pulumi.String(providerConfig.Username),
			Password:    pulumi.String(providerConfig.Password),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create Snowflake provider with credentials")
		}
	}

	// Create the Snowflake database resource
	createdDatabase, err := createDatabase(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Snowflake database")
	}

	// Export stack outputs
	ctx.Export(OpId, createdDatabase.ID())
	ctx.Export(OpName, createdDatabase.Name)
	ctx.Export(OpIsTransient, pulumi.Bool(locals.SnowflakeDatabase.Spec.IsTransient))
	ctx.Export(OpDataRetentionDays, pulumi.Int(locals.SnowflakeDatabase.Spec.DataRetentionTimeInDays))

	return nil
}

// createDatabase creates the Snowflake database resource with all spec parameters
func createDatabase(ctx *pulumi.Context, locals *Locals, provider *snowflake.Provider) (*snowflake.Database, error) {
	spec := locals.SnowflakeDatabase.Spec
	metadata := locals.SnowflakeDatabase.Metadata

	// Build database arguments
	databaseArgs := &snowflake.DatabaseArgs{
		Name: pulumi.String(spec.Name),
	}

	// Set optional string parameters
	if spec.Catalog != "" {
		databaseArgs.Catalog = pulumi.String(spec.Catalog)
	}
	if spec.Comment != "" {
		databaseArgs.Comment = pulumi.String(spec.Comment)
	}
	if spec.DefaultDdlCollation != "" {
		databaseArgs.DefaultDdlCollation = pulumi.String(spec.DefaultDdlCollation)
	}
	if spec.ExternalVolume != "" {
		databaseArgs.ExternalVolume = pulumi.String(spec.ExternalVolume)
	}
	if spec.LogLevel != "" {
		databaseArgs.LogLevel = pulumi.String(spec.LogLevel)
	}
	if spec.StorageSerializationPolicy != "" {
		databaseArgs.StorageSerializationPolicy = pulumi.String(spec.StorageSerializationPolicy)
	}
	if spec.TraceLevel != "" {
		databaseArgs.TraceLevel = pulumi.String(spec.TraceLevel)
	}

	// Set integer parameters (only if > 0)
	if spec.DataRetentionTimeInDays > 0 {
		databaseArgs.DataRetentionTimeInDays = pulumi.Int(int(spec.DataRetentionTimeInDays))
	}
	if spec.MaxDataExtensionTimeInDays > 0 {
		databaseArgs.MaxDataExtensionTimeInDays = pulumi.Int(int(spec.MaxDataExtensionTimeInDays))
	}
	if spec.SuspendTaskAfterNumFailures >= 0 {
		databaseArgs.SuspendTaskAfterNumFailures = pulumi.Int(int(spec.SuspendTaskAfterNumFailures))
	}
	if spec.TaskAutoRetryAttempts >= 0 {
		databaseArgs.TaskAutoRetryAttempts = pulumi.Int(int(spec.TaskAutoRetryAttempts))
	}

	// Set boolean parameters
	databaseArgs.DropPublicSchemaOnCreation = pulumi.Bool(spec.DropPublicSchemaOnCreation)
	databaseArgs.EnableConsoleOutput = pulumi.Bool(spec.EnableConsoleOutput)
	databaseArgs.IsTransient = pulumi.Bool(spec.IsTransient)
	databaseArgs.QuotedIdentifiersIgnoreCase = pulumi.Bool(spec.QuotedIdentifiersIgnoreCase)
	databaseArgs.ReplaceInvalidCharacters = pulumi.Bool(spec.ReplaceInvalidCharacters)

	// Set user task parameters
	if spec.UserTask != nil {
		if spec.UserTask.ManagedInitialWarehouseSize != "" {
			databaseArgs.UserTaskManagedInitialWarehouseSize = pulumi.String(spec.UserTask.ManagedInitialWarehouseSize)
		}
		if spec.UserTask.MinimumTriggerIntervalInSeconds > 0 {
			databaseArgs.UserTaskMinimumTriggerIntervalInSeconds = pulumi.Int(int(spec.UserTask.MinimumTriggerIntervalInSeconds))
		}
		if spec.UserTask.TimeoutMs > 0 {
			databaseArgs.UserTaskTimeoutMs = pulumi.Int(int(spec.UserTask.TimeoutMs))
		}
	}

	// Create the database resource
	resourceName := metadata.Name
	database, err := snowflake.NewDatabase(ctx, resourceName, databaseArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create database %s", spec.Name)
	}

	return database, nil
}
