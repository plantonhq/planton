package module

import (
	"github.com/pkg/errors"
	gcpdatastreamstreamv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdatastreamstream/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/datastream"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// optionalString returns nil for an empty value so the provider omits it.
func optionalString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.String(value)
}

// optionalCount returns nil for zero so Google's default applies -- the
// Terraform module's null-for-zero rule.
func optionalCount(count int32) pulumi.IntPtrInput {
	if count <= 0 {
		return nil
	}
	return pulumi.Int(int(count))
}

// stream creates the Datastream stream: one source arm, one destination
// arm, and one backfill mode (the spec's rules), with desired_state always
// declared.
func stream(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDatastreamStream.Spec
	resourceName := locals.GcpDatastreamStream.Metadata.Name

	args := &datastream.StreamArgs{
		Location:          pulumi.String(spec.Location),
		StreamId:          pulumi.String(locals.StreamId),
		DisplayName:       pulumi.String(locals.DisplayName),
		DesiredState:      pulumi.String(locals.DesiredState),
		SourceConfig:      sourceConfig(spec.SourceConfig),
		DestinationConfig: destinationConfig(spec.DestinationConfig),
		Labels:            pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.CreateWithoutValidation {
		args.CreateWithoutValidation = pulumi.BoolPtr(true)
	}
	args.CustomerManagedEncryptionKey = optionalString(spec.CustomerManagedEncryptionKey.GetValue())

	if backfill := spec.BackfillAll; backfill != nil {
		args.BackfillAll = &datastream.StreamBackfillAllArgs{
			MysqlExcludedObjects:      mysqlBackfillExcludedObjects(backfill.MysqlExcludedObjects),
			PostgresqlExcludedObjects: postgresqlBackfillExcludedObjects(backfill.PostgresqlExcludedObjects),
			OracleExcludedObjects:     oracleBackfillExcludedObjects(backfill.OracleExcludedObjects),
			SqlServerExcludedObjects:  sqlServerBackfillExcludedObjects(backfill.SqlServerExcludedObjects),
			MongodbExcludedObjects:    mongodbBackfillExcludedObjects(backfill.MongodbExcludedObjects),
			SalesforceExcludedObjects: salesforceBackfillExcludedObjects(backfill.SalesforceExcludedObjects),
			SpannerExcludedObjects:    spannerBackfillExcludedObjects(backfill.SpannerExcludedObjects),
		}
	}
	// Google's backfill_none block has no settings; the spec's bool emits
	// the empty marker block.
	if spec.BackfillNone {
		args.BackfillNone = &datastream.StreamBackfillNoneArgs{}
	}

	if len(spec.RuleSets) > 0 {
		args.RuleSets = ruleSets(spec.RuleSets)
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := datastream.NewStream(ctx, resourceName, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create datastream stream")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpStreamId, created.StreamId)
	return nil
}

// sourceConfig maps the one declared source arm. The either-or enums
// select which of Google's empty blocks is sent.
func sourceConfig(source *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSourceConfig) *datastream.StreamSourceConfigArgs {
	args := &datastream.StreamSourceConfigArgs{
		SourceConnectionProfile: pulumi.String(source.SourceConnectionProfile.GetValue()),
	}

	if mysql := source.MysqlSourceConfig; mysql != nil {
		mysqlArgs := &datastream.StreamSourceConfigMysqlSourceConfigArgs{
			IncludeObjects:             mysqlIncludeObjects(mysql.IncludeObjects),
			ExcludeObjects:             mysqlExcludeObjects(mysql.ExcludeObjects),
			MaxConcurrentBackfillTasks: optionalCount(mysql.MaxConcurrentBackfillTasks),
			MaxConcurrentCdcTasks:      optionalCount(mysql.MaxConcurrentCdcTasks),
		}
		switch mysql.CdcMethod {
		case "GTID":
			mysqlArgs.Gtid = &datastream.StreamSourceConfigMysqlSourceConfigGtidArgs{}
		case "BINARY_LOG_POSITION":
			mysqlArgs.BinaryLogPosition = &datastream.StreamSourceConfigMysqlSourceConfigBinaryLogPositionArgs{}
		}
		args.MysqlSourceConfig = mysqlArgs
	}

	if postgresql := source.PostgresqlSourceConfig; postgresql != nil {
		args.PostgresqlSourceConfig = &datastream.StreamSourceConfigPostgresqlSourceConfigArgs{
			IncludeObjects:             postgresqlIncludeObjects(postgresql.IncludeObjects),
			ExcludeObjects:             postgresqlExcludeObjects(postgresql.ExcludeObjects),
			ReplicationSlot:            pulumi.String(postgresql.ReplicationSlot),
			Publication:                pulumi.String(postgresql.Publication),
			MaxConcurrentBackfillTasks: optionalCount(postgresql.MaxConcurrentBackfillTasks),
		}
	}

	if oracle := source.OracleSourceConfig; oracle != nil {
		oracleArgs := &datastream.StreamSourceConfigOracleSourceConfigArgs{
			IncludeObjects:             oracleIncludeObjects(oracle.IncludeObjects),
			ExcludeObjects:             oracleExcludeObjects(oracle.ExcludeObjects),
			MaxConcurrentBackfillTasks: optionalCount(oracle.MaxConcurrentBackfillTasks),
			MaxConcurrentCdcTasks:      optionalCount(oracle.MaxConcurrentCdcTasks),
		}
		switch oracle.LargeObjectsHandling {
		case "DROP":
			oracleArgs.DropLargeObjects = &datastream.StreamSourceConfigOracleSourceConfigDropLargeObjectsArgs{}
		case "STREAM":
			oracleArgs.StreamLargeObjects = &datastream.StreamSourceConfigOracleSourceConfigStreamLargeObjectsArgs{}
		}
		args.OracleSourceConfig = oracleArgs
	}

	if sqlServer := source.SqlServerSourceConfig; sqlServer != nil {
		sqlServerArgs := &datastream.StreamSourceConfigSqlServerSourceConfigArgs{
			IncludeObjects:             sqlServerIncludeObjects(sqlServer.IncludeObjects),
			ExcludeObjects:             sqlServerExcludeObjects(sqlServer.ExcludeObjects),
			MaxConcurrentBackfillTasks: optionalCount(sqlServer.MaxConcurrentBackfillTasks),
			MaxConcurrentCdcTasks:      optionalCount(sqlServer.MaxConcurrentCdcTasks),
		}
		switch sqlServer.CdcMethod {
		case "CHANGE_TABLES":
			sqlServerArgs.ChangeTables = &datastream.StreamSourceConfigSqlServerSourceConfigChangeTablesArgs{}
		case "TRANSACTION_LOGS":
			sqlServerArgs.TransactionLogs = &datastream.StreamSourceConfigSqlServerSourceConfigTransactionLogsArgs{}
		}
		args.SqlServerSourceConfig = sqlServerArgs
	}

	if mongodb := source.MongodbSourceConfig; mongodb != nil {
		args.MongodbSourceConfig = &datastream.StreamSourceConfigMongodbSourceConfigArgs{
			IncludeObjects:             mongodbIncludeObjects(mongodb.IncludeObjects),
			ExcludeObjects:             mongodbExcludeObjects(mongodb.ExcludeObjects),
			MaxConcurrentBackfillTasks: optionalCount(mongodb.MaxConcurrentBackfillTasks),
		}
	}

	if salesforce := source.SalesforceSourceConfig; salesforce != nil {
		args.SalesforceSourceConfig = &datastream.StreamSourceConfigSalesforceSourceConfigArgs{
			IncludeObjects:  salesforceIncludeObjects(salesforce.IncludeObjects),
			ExcludeObjects:  salesforceExcludeObjects(salesforce.ExcludeObjects),
			PollingInterval: pulumi.String(salesforce.PollingInterval),
		}
	}

	if spanner := source.SpannerSourceConfig; spanner != nil {
		spannerArgs := &datastream.StreamSourceConfigSpannerSourceConfigArgs{
			IncludeObjects:             spannerIncludeObjects(spanner.IncludeObjects),
			ExcludeObjects:             spannerExcludeObjects(spanner.ExcludeObjects),
			ChangeStreamName:           pulumi.String(spanner.ChangeStreamName),
			FgacRole:                   optionalString(spanner.FgacRole),
			MaxConcurrentBackfillTasks: optionalCount(spanner.MaxConcurrentBackfillTasks),
			MaxConcurrentCdcTasks:      optionalCount(spanner.MaxConcurrentCdcTasks),
			SpannerRpcPriority:         optionalString(spanner.SpannerRpcPriority),
		}
		if spanner.BackfillDataBoostEnabled {
			spannerArgs.BackfillDataBoostEnabled = pulumi.BoolPtr(true)
		}
		args.SpannerSourceConfig = spannerArgs
	}

	return args
}

// destinationConfig maps the one declared destination arm, converting the
// dataset and BigQuery-connection references to the forms Google wants.
func destinationConfig(destination *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamDestinationConfig) *datastream.StreamDestinationConfigArgs {
	args := &datastream.StreamDestinationConfigArgs{
		DestinationConnectionProfile: pulumi.String(destination.DestinationConnectionProfile.GetValue()),
	}

	if bigquery := destination.BigqueryDestinationConfig; bigquery != nil {
		bigqueryArgs := &datastream.StreamDestinationConfigBigqueryDestinationConfigArgs{
			DataFreshness: optionalString(bigquery.DataFreshness),
		}
		if single := bigquery.SingleTargetDataset; single != nil {
			bigqueryArgs.SingleTargetDataset = &datastream.StreamDestinationConfigBigqueryDestinationConfigSingleTargetDatasetArgs{
				DatasetId: pulumi.String(datasetId(single.DatasetId.GetValue())),
			}
		}
		if hierarchy := bigquery.SourceHierarchyDatasets; hierarchy != nil {
			template := hierarchy.DatasetTemplate
			bigqueryArgs.SourceHierarchyDatasets = &datastream.StreamDestinationConfigBigqueryDestinationConfigSourceHierarchyDatasetsArgs{
				ProjectId: optionalString(hierarchy.ProjectId.GetValue()),
				DatasetTemplate: &datastream.StreamDestinationConfigBigqueryDestinationConfigSourceHierarchyDatasetsDatasetTemplateArgs{
					Location:        pulumi.String(template.Location),
					DatasetIdPrefix: optionalString(template.DatasetIdPrefix),
					KmsKeyName:      optionalString(template.KmsKeyName.GetValue()),
				},
			}
		}
		switch bigquery.WriteMode {
		case "MERGE":
			bigqueryArgs.Merge = &datastream.StreamDestinationConfigBigqueryDestinationConfigMergeArgs{}
		case "APPEND_ONLY":
			bigqueryArgs.AppendOnly = &datastream.StreamDestinationConfigBigqueryDestinationConfigAppendOnlyArgs{}
		}
		if blmt := bigquery.BlmtConfig; blmt != nil {
			bigqueryArgs.BlmtConfig = &datastream.StreamDestinationConfigBigqueryDestinationConfigBlmtConfigArgs{
				Bucket:         pulumi.String(blmt.Bucket.GetValue()),
				RootPath:       optionalString(blmt.RootPath),
				ConnectionName: pulumi.String(bigQueryConnectionName(blmt.ConnectionName.GetValue())),
				FileFormat:     pulumi.String(blmt.FileFormat),
				TableFormat:    pulumi.String(blmt.TableFormat),
			}
		}
		args.BigqueryDestinationConfig = bigqueryArgs
	}

	if gcs := destination.GcsDestinationConfig; gcs != nil {
		gcsArgs := &datastream.StreamDestinationConfigGcsDestinationConfigArgs{
			Path:                 optionalString(gcs.Path),
			FileRotationInterval: optionalString(gcs.FileRotationInterval),
			FileRotationMb:       optionalCount(gcs.FileRotationMb),
		}
		// Google's Avro block has no settings; the spec's bool emits the
		// empty marker block.
		if gcs.AvroFileFormat {
			gcsArgs.AvroFileFormat = &datastream.StreamDestinationConfigGcsDestinationConfigAvroFileFormatArgs{}
		}
		if json := gcs.JsonFileFormat; json != nil {
			gcsArgs.JsonFileFormat = &datastream.StreamDestinationConfigGcsDestinationConfigJsonFileFormatArgs{
				Compression:      optionalString(json.Compression),
				SchemaFileFormat: optionalString(json.SchemaFileFormat),
			}
		}
		args.GcsDestinationConfig = gcsArgs
	}

	return args
}

// ruleSets maps the per-object BigQuery table customizations.
func ruleSets(sets []*gcpdatastreamstreamv1alpha1.GcpDatastreamStreamRuleSet) datastream.StreamRuleSetArray {
	out := datastream.StreamRuleSetArray{}
	for _, set := range sets {
		filterArgs := &datastream.StreamRuleSetObjectFilterArgs{}
		if identifier := set.ObjectFilter.SourceObjectIdentifier; identifier != nil {
			filterArgs.SourceObjectIdentifier = sourceObjectIdentifier(identifier)
		}

		rules := datastream.StreamRuleSetCustomizationRuleArray{}
		for _, rule := range set.CustomizationRules {
			ruleArgs := &datastream.StreamRuleSetCustomizationRuleArgs{}
			if clustering := rule.BigqueryClustering; clustering != nil {
				ruleArgs.BigqueryClustering = &datastream.StreamRuleSetCustomizationRuleBigqueryClusteringArgs{
					Columns: pulumi.ToStringArray(clustering.Columns),
				}
			}
			if partitioning := rule.BigqueryPartitioning; partitioning != nil {
				ruleArgs.BigqueryPartitioning = bigqueryPartitioning(partitioning)
			}
			rules = append(rules, ruleArgs)
		}

		out = append(out, &datastream.StreamRuleSetArgs{
			ObjectFilter:       filterArgs,
			CustomizationRules: rules,
		})
	}
	return out
}

func sourceObjectIdentifier(identifier *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSourceObjectIdentifier) *datastream.StreamRuleSetObjectFilterSourceObjectIdentifierArgs {
	args := &datastream.StreamRuleSetObjectFilterSourceObjectIdentifierArgs{}
	if id := identifier.MysqlIdentifier; id != nil {
		args.MysqlIdentifier = &datastream.StreamRuleSetObjectFilterSourceObjectIdentifierMysqlIdentifierArgs{
			Database: pulumi.String(id.Database),
			Table:    pulumi.String(id.Table),
		}
	}
	if id := identifier.PostgresqlIdentifier; id != nil {
		args.PostgresqlIdentifier = &datastream.StreamRuleSetObjectFilterSourceObjectIdentifierPostgresqlIdentifierArgs{
			Schema: pulumi.String(id.Schema),
			Table:  pulumi.String(id.Table),
		}
	}
	if id := identifier.OracleIdentifier; id != nil {
		args.OracleIdentifier = &datastream.StreamRuleSetObjectFilterSourceObjectIdentifierOracleIdentifierArgs{
			Schema: pulumi.String(id.Schema),
			Table:  pulumi.String(id.Table),
		}
	}
	if id := identifier.SqlServerIdentifier; id != nil {
		args.SqlServerIdentifier = &datastream.StreamRuleSetObjectFilterSourceObjectIdentifierSqlServerIdentifierArgs{
			Schema: pulumi.String(id.Schema),
			Table:  pulumi.String(id.Table),
		}
	}
	if id := identifier.MongodbIdentifier; id != nil {
		args.MongodbIdentifier = &datastream.StreamRuleSetObjectFilterSourceObjectIdentifierMongodbIdentifierArgs{
			Database:   pulumi.String(id.Database),
			Collection: pulumi.String(id.Collection),
		}
	}
	if id := identifier.SalesforceIdentifier; id != nil {
		args.SalesforceIdentifier = &datastream.StreamRuleSetObjectFilterSourceObjectIdentifierSalesforceIdentifierArgs{
			ObjectName: pulumi.String(id.ObjectName),
		}
	}
	if id := identifier.SpannerIdentifier; id != nil {
		args.SpannerIdentifier = &datastream.StreamRuleSetObjectFilterSourceObjectIdentifierSpannerIdentifierArgs{
			Schema: optionalString(id.Schema),
			Table:  pulumi.String(id.Table),
		}
	}
	return args
}

func bigqueryPartitioning(partitioning *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamBigqueryPartitioning) *datastream.StreamRuleSetCustomizationRuleBigqueryPartitioningArgs {
	args := &datastream.StreamRuleSetCustomizationRuleBigqueryPartitioningArgs{}
	if partitioning.RequirePartitionFilter {
		args.RequirePartitionFilter = pulumi.BoolPtr(true)
	}
	if ingestion := partitioning.IngestionTimePartition; ingestion != nil {
		args.IngestionTimePartition = &datastream.StreamRuleSetCustomizationRuleBigqueryPartitioningIngestionTimePartitionArgs{
			PartitioningTimeGranularity: optionalString(ingestion.PartitioningTimeGranularity),
		}
	}
	if timeUnit := partitioning.TimeUnitPartition; timeUnit != nil {
		args.TimeUnitPartition = &datastream.StreamRuleSetCustomizationRuleBigqueryPartitioningTimeUnitPartitionArgs{
			Column:                      pulumi.String(timeUnit.Column),
			PartitioningTimeGranularity: optionalString(timeUnit.PartitioningTimeGranularity),
		}
	}
	if integerRange := partitioning.IntegerRangePartition; integerRange != nil {
		args.IntegerRangePartition = &datastream.StreamRuleSetCustomizationRuleBigqueryPartitioningIntegerRangePartitionArgs{
			Column:   pulumi.String(integerRange.Column),
			Start:    pulumi.Int(int(integerRange.Start)),
			End:      pulumi.Int(int(integerRange.End)),
			Interval: pulumi.Int(int(integerRange.Interval)),
		}
	}
	return args
}
