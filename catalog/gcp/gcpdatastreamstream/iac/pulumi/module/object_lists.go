package module

import (
	gcpdatastreamstreamv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdatastreamstream/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/datastream"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// The spec carries ONE object-list message per source family, reused for
// what a stream includes, what it excludes, and what a full backfill skips.
// The Pulumi SDK types each of those three places separately, so every
// family has one builder per place below, all with the Terraform module's
// send posture: empty names, zero positions, and false flags are omitted.

func mysqlIncludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamMysqlRdbms) *datastream.StreamSourceConfigMysqlSourceConfigIncludeObjectsArgs {
	if set == nil {
		return nil
	}
	mysqlDatabases := datastream.StreamSourceConfigMysqlSourceConfigIncludeObjectsMysqlDatabaseArray{}
	for _, mysqlDatabase := range set.MysqlDatabases {
		mysqlDatabaseArgs := &datastream.StreamSourceConfigMysqlSourceConfigIncludeObjectsMysqlDatabaseArgs{}
		if mysqlDatabase.Database != "" {
			mysqlDatabaseArgs.Database = pulumi.String(mysqlDatabase.Database)
		}
		if len(mysqlDatabase.MysqlTables) > 0 {
			mysqlTables := datastream.StreamSourceConfigMysqlSourceConfigIncludeObjectsMysqlDatabaseMysqlTableArray{}
			for _, mysqlTable := range mysqlDatabase.MysqlTables {
				mysqlTableArgs := &datastream.StreamSourceConfigMysqlSourceConfigIncludeObjectsMysqlDatabaseMysqlTableArgs{}
				if mysqlTable.Table != "" {
					mysqlTableArgs.Table = pulumi.String(mysqlTable.Table)
				}
				if len(mysqlTable.MysqlColumns) > 0 {
					mysqlColumns := datastream.StreamSourceConfigMysqlSourceConfigIncludeObjectsMysqlDatabaseMysqlTableMysqlColumnArray{}
					for _, mysqlColumn := range mysqlTable.MysqlColumns {
						mysqlColumnArgs := &datastream.StreamSourceConfigMysqlSourceConfigIncludeObjectsMysqlDatabaseMysqlTableMysqlColumnArgs{}
						if mysqlColumn.Column != "" {
							mysqlColumnArgs.Column = pulumi.String(mysqlColumn.Column)
						}
						if mysqlColumn.Collation != "" {
							mysqlColumnArgs.Collation = pulumi.String(mysqlColumn.Collation)
						}
						if mysqlColumn.DataType != "" {
							mysqlColumnArgs.DataType = pulumi.String(mysqlColumn.DataType)
						}
						if mysqlColumn.Nullable {
							mysqlColumnArgs.Nullable = pulumi.Bool(true)
						}
						if mysqlColumn.OrdinalPosition > 0 {
							mysqlColumnArgs.OrdinalPosition = pulumi.Int(int(mysqlColumn.OrdinalPosition))
						}
						if mysqlColumn.PrimaryKey {
							mysqlColumnArgs.PrimaryKey = pulumi.Bool(true)
						}
						mysqlColumns = append(mysqlColumns, mysqlColumnArgs)
					}
					mysqlTableArgs.MysqlColumns = mysqlColumns
				}
				mysqlTables = append(mysqlTables, mysqlTableArgs)
			}
			mysqlDatabaseArgs.MysqlTables = mysqlTables
		}
		mysqlDatabases = append(mysqlDatabases, mysqlDatabaseArgs)
	}
	return &datastream.StreamSourceConfigMysqlSourceConfigIncludeObjectsArgs{MysqlDatabases: mysqlDatabases}
}

func mysqlExcludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamMysqlRdbms) *datastream.StreamSourceConfigMysqlSourceConfigExcludeObjectsArgs {
	if set == nil {
		return nil
	}
	mysqlDatabases := datastream.StreamSourceConfigMysqlSourceConfigExcludeObjectsMysqlDatabaseArray{}
	for _, mysqlDatabase := range set.MysqlDatabases {
		mysqlDatabaseArgs := &datastream.StreamSourceConfigMysqlSourceConfigExcludeObjectsMysqlDatabaseArgs{}
		if mysqlDatabase.Database != "" {
			mysqlDatabaseArgs.Database = pulumi.String(mysqlDatabase.Database)
		}
		if len(mysqlDatabase.MysqlTables) > 0 {
			mysqlTables := datastream.StreamSourceConfigMysqlSourceConfigExcludeObjectsMysqlDatabaseMysqlTableArray{}
			for _, mysqlTable := range mysqlDatabase.MysqlTables {
				mysqlTableArgs := &datastream.StreamSourceConfigMysqlSourceConfigExcludeObjectsMysqlDatabaseMysqlTableArgs{}
				if mysqlTable.Table != "" {
					mysqlTableArgs.Table = pulumi.String(mysqlTable.Table)
				}
				if len(mysqlTable.MysqlColumns) > 0 {
					mysqlColumns := datastream.StreamSourceConfigMysqlSourceConfigExcludeObjectsMysqlDatabaseMysqlTableMysqlColumnArray{}
					for _, mysqlColumn := range mysqlTable.MysqlColumns {
						mysqlColumnArgs := &datastream.StreamSourceConfigMysqlSourceConfigExcludeObjectsMysqlDatabaseMysqlTableMysqlColumnArgs{}
						if mysqlColumn.Column != "" {
							mysqlColumnArgs.Column = pulumi.String(mysqlColumn.Column)
						}
						if mysqlColumn.Collation != "" {
							mysqlColumnArgs.Collation = pulumi.String(mysqlColumn.Collation)
						}
						if mysqlColumn.DataType != "" {
							mysqlColumnArgs.DataType = pulumi.String(mysqlColumn.DataType)
						}
						if mysqlColumn.Nullable {
							mysqlColumnArgs.Nullable = pulumi.Bool(true)
						}
						if mysqlColumn.OrdinalPosition > 0 {
							mysqlColumnArgs.OrdinalPosition = pulumi.Int(int(mysqlColumn.OrdinalPosition))
						}
						if mysqlColumn.PrimaryKey {
							mysqlColumnArgs.PrimaryKey = pulumi.Bool(true)
						}
						mysqlColumns = append(mysqlColumns, mysqlColumnArgs)
					}
					mysqlTableArgs.MysqlColumns = mysqlColumns
				}
				mysqlTables = append(mysqlTables, mysqlTableArgs)
			}
			mysqlDatabaseArgs.MysqlTables = mysqlTables
		}
		mysqlDatabases = append(mysqlDatabases, mysqlDatabaseArgs)
	}
	return &datastream.StreamSourceConfigMysqlSourceConfigExcludeObjectsArgs{MysqlDatabases: mysqlDatabases}
}

func mysqlBackfillExcludedObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamMysqlRdbms) *datastream.StreamBackfillAllMysqlExcludedObjectsArgs {
	if set == nil {
		return nil
	}
	mysqlDatabases := datastream.StreamBackfillAllMysqlExcludedObjectsMysqlDatabaseArray{}
	for _, mysqlDatabase := range set.MysqlDatabases {
		mysqlDatabaseArgs := &datastream.StreamBackfillAllMysqlExcludedObjectsMysqlDatabaseArgs{}
		if mysqlDatabase.Database != "" {
			mysqlDatabaseArgs.Database = pulumi.String(mysqlDatabase.Database)
		}
		if len(mysqlDatabase.MysqlTables) > 0 {
			mysqlTables := datastream.StreamBackfillAllMysqlExcludedObjectsMysqlDatabaseMysqlTableArray{}
			for _, mysqlTable := range mysqlDatabase.MysqlTables {
				mysqlTableArgs := &datastream.StreamBackfillAllMysqlExcludedObjectsMysqlDatabaseMysqlTableArgs{}
				if mysqlTable.Table != "" {
					mysqlTableArgs.Table = pulumi.String(mysqlTable.Table)
				}
				if len(mysqlTable.MysqlColumns) > 0 {
					mysqlColumns := datastream.StreamBackfillAllMysqlExcludedObjectsMysqlDatabaseMysqlTableMysqlColumnArray{}
					for _, mysqlColumn := range mysqlTable.MysqlColumns {
						mysqlColumnArgs := &datastream.StreamBackfillAllMysqlExcludedObjectsMysqlDatabaseMysqlTableMysqlColumnArgs{}
						if mysqlColumn.Column != "" {
							mysqlColumnArgs.Column = pulumi.String(mysqlColumn.Column)
						}
						if mysqlColumn.Collation != "" {
							mysqlColumnArgs.Collation = pulumi.String(mysqlColumn.Collation)
						}
						if mysqlColumn.DataType != "" {
							mysqlColumnArgs.DataType = pulumi.String(mysqlColumn.DataType)
						}
						if mysqlColumn.Nullable {
							mysqlColumnArgs.Nullable = pulumi.Bool(true)
						}
						if mysqlColumn.OrdinalPosition > 0 {
							mysqlColumnArgs.OrdinalPosition = pulumi.Int(int(mysqlColumn.OrdinalPosition))
						}
						if mysqlColumn.PrimaryKey {
							mysqlColumnArgs.PrimaryKey = pulumi.Bool(true)
						}
						mysqlColumns = append(mysqlColumns, mysqlColumnArgs)
					}
					mysqlTableArgs.MysqlColumns = mysqlColumns
				}
				mysqlTables = append(mysqlTables, mysqlTableArgs)
			}
			mysqlDatabaseArgs.MysqlTables = mysqlTables
		}
		mysqlDatabases = append(mysqlDatabases, mysqlDatabaseArgs)
	}
	return &datastream.StreamBackfillAllMysqlExcludedObjectsArgs{MysqlDatabases: mysqlDatabases}
}

func postgresqlIncludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamPostgresqlRdbms) *datastream.StreamSourceConfigPostgresqlSourceConfigIncludeObjectsArgs {
	if set == nil {
		return nil
	}
	postgresqlSchemas := datastream.StreamSourceConfigPostgresqlSourceConfigIncludeObjectsPostgresqlSchemaArray{}
	for _, postgresqlSchema := range set.PostgresqlSchemas {
		postgresqlSchemaArgs := &datastream.StreamSourceConfigPostgresqlSourceConfigIncludeObjectsPostgresqlSchemaArgs{}
		if postgresqlSchema.Schema != "" {
			postgresqlSchemaArgs.Schema = pulumi.String(postgresqlSchema.Schema)
		}
		if len(postgresqlSchema.PostgresqlTables) > 0 {
			postgresqlTables := datastream.StreamSourceConfigPostgresqlSourceConfigIncludeObjectsPostgresqlSchemaPostgresqlTableArray{}
			for _, postgresqlTable := range postgresqlSchema.PostgresqlTables {
				postgresqlTableArgs := &datastream.StreamSourceConfigPostgresqlSourceConfigIncludeObjectsPostgresqlSchemaPostgresqlTableArgs{}
				if postgresqlTable.Table != "" {
					postgresqlTableArgs.Table = pulumi.String(postgresqlTable.Table)
				}
				if len(postgresqlTable.PostgresqlColumns) > 0 {
					postgresqlColumns := datastream.StreamSourceConfigPostgresqlSourceConfigIncludeObjectsPostgresqlSchemaPostgresqlTablePostgresqlColumnArray{}
					for _, postgresqlColumn := range postgresqlTable.PostgresqlColumns {
						postgresqlColumnArgs := &datastream.StreamSourceConfigPostgresqlSourceConfigIncludeObjectsPostgresqlSchemaPostgresqlTablePostgresqlColumnArgs{}
						if postgresqlColumn.Column != "" {
							postgresqlColumnArgs.Column = pulumi.String(postgresqlColumn.Column)
						}
						if postgresqlColumn.DataType != "" {
							postgresqlColumnArgs.DataType = pulumi.String(postgresqlColumn.DataType)
						}
						if postgresqlColumn.Nullable {
							postgresqlColumnArgs.Nullable = pulumi.Bool(true)
						}
						if postgresqlColumn.OrdinalPosition > 0 {
							postgresqlColumnArgs.OrdinalPosition = pulumi.Int(int(postgresqlColumn.OrdinalPosition))
						}
						if postgresqlColumn.PrimaryKey {
							postgresqlColumnArgs.PrimaryKey = pulumi.Bool(true)
						}
						postgresqlColumns = append(postgresqlColumns, postgresqlColumnArgs)
					}
					postgresqlTableArgs.PostgresqlColumns = postgresqlColumns
				}
				postgresqlTables = append(postgresqlTables, postgresqlTableArgs)
			}
			postgresqlSchemaArgs.PostgresqlTables = postgresqlTables
		}
		postgresqlSchemas = append(postgresqlSchemas, postgresqlSchemaArgs)
	}
	return &datastream.StreamSourceConfigPostgresqlSourceConfigIncludeObjectsArgs{PostgresqlSchemas: postgresqlSchemas}
}

func postgresqlExcludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamPostgresqlRdbms) *datastream.StreamSourceConfigPostgresqlSourceConfigExcludeObjectsArgs {
	if set == nil {
		return nil
	}
	postgresqlSchemas := datastream.StreamSourceConfigPostgresqlSourceConfigExcludeObjectsPostgresqlSchemaArray{}
	for _, postgresqlSchema := range set.PostgresqlSchemas {
		postgresqlSchemaArgs := &datastream.StreamSourceConfigPostgresqlSourceConfigExcludeObjectsPostgresqlSchemaArgs{}
		if postgresqlSchema.Schema != "" {
			postgresqlSchemaArgs.Schema = pulumi.String(postgresqlSchema.Schema)
		}
		if len(postgresqlSchema.PostgresqlTables) > 0 {
			postgresqlTables := datastream.StreamSourceConfigPostgresqlSourceConfigExcludeObjectsPostgresqlSchemaPostgresqlTableArray{}
			for _, postgresqlTable := range postgresqlSchema.PostgresqlTables {
				postgresqlTableArgs := &datastream.StreamSourceConfigPostgresqlSourceConfigExcludeObjectsPostgresqlSchemaPostgresqlTableArgs{}
				if postgresqlTable.Table != "" {
					postgresqlTableArgs.Table = pulumi.String(postgresqlTable.Table)
				}
				if len(postgresqlTable.PostgresqlColumns) > 0 {
					postgresqlColumns := datastream.StreamSourceConfigPostgresqlSourceConfigExcludeObjectsPostgresqlSchemaPostgresqlTablePostgresqlColumnArray{}
					for _, postgresqlColumn := range postgresqlTable.PostgresqlColumns {
						postgresqlColumnArgs := &datastream.StreamSourceConfigPostgresqlSourceConfigExcludeObjectsPostgresqlSchemaPostgresqlTablePostgresqlColumnArgs{}
						if postgresqlColumn.Column != "" {
							postgresqlColumnArgs.Column = pulumi.String(postgresqlColumn.Column)
						}
						if postgresqlColumn.DataType != "" {
							postgresqlColumnArgs.DataType = pulumi.String(postgresqlColumn.DataType)
						}
						if postgresqlColumn.Nullable {
							postgresqlColumnArgs.Nullable = pulumi.Bool(true)
						}
						if postgresqlColumn.OrdinalPosition > 0 {
							postgresqlColumnArgs.OrdinalPosition = pulumi.Int(int(postgresqlColumn.OrdinalPosition))
						}
						if postgresqlColumn.PrimaryKey {
							postgresqlColumnArgs.PrimaryKey = pulumi.Bool(true)
						}
						postgresqlColumns = append(postgresqlColumns, postgresqlColumnArgs)
					}
					postgresqlTableArgs.PostgresqlColumns = postgresqlColumns
				}
				postgresqlTables = append(postgresqlTables, postgresqlTableArgs)
			}
			postgresqlSchemaArgs.PostgresqlTables = postgresqlTables
		}
		postgresqlSchemas = append(postgresqlSchemas, postgresqlSchemaArgs)
	}
	return &datastream.StreamSourceConfigPostgresqlSourceConfigExcludeObjectsArgs{PostgresqlSchemas: postgresqlSchemas}
}

func postgresqlBackfillExcludedObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamPostgresqlRdbms) *datastream.StreamBackfillAllPostgresqlExcludedObjectsArgs {
	if set == nil {
		return nil
	}
	postgresqlSchemas := datastream.StreamBackfillAllPostgresqlExcludedObjectsPostgresqlSchemaArray{}
	for _, postgresqlSchema := range set.PostgresqlSchemas {
		postgresqlSchemaArgs := &datastream.StreamBackfillAllPostgresqlExcludedObjectsPostgresqlSchemaArgs{}
		if postgresqlSchema.Schema != "" {
			postgresqlSchemaArgs.Schema = pulumi.String(postgresqlSchema.Schema)
		}
		if len(postgresqlSchema.PostgresqlTables) > 0 {
			postgresqlTables := datastream.StreamBackfillAllPostgresqlExcludedObjectsPostgresqlSchemaPostgresqlTableArray{}
			for _, postgresqlTable := range postgresqlSchema.PostgresqlTables {
				postgresqlTableArgs := &datastream.StreamBackfillAllPostgresqlExcludedObjectsPostgresqlSchemaPostgresqlTableArgs{}
				if postgresqlTable.Table != "" {
					postgresqlTableArgs.Table = pulumi.String(postgresqlTable.Table)
				}
				if len(postgresqlTable.PostgresqlColumns) > 0 {
					postgresqlColumns := datastream.StreamBackfillAllPostgresqlExcludedObjectsPostgresqlSchemaPostgresqlTablePostgresqlColumnArray{}
					for _, postgresqlColumn := range postgresqlTable.PostgresqlColumns {
						postgresqlColumnArgs := &datastream.StreamBackfillAllPostgresqlExcludedObjectsPostgresqlSchemaPostgresqlTablePostgresqlColumnArgs{}
						if postgresqlColumn.Column != "" {
							postgresqlColumnArgs.Column = pulumi.String(postgresqlColumn.Column)
						}
						if postgresqlColumn.DataType != "" {
							postgresqlColumnArgs.DataType = pulumi.String(postgresqlColumn.DataType)
						}
						if postgresqlColumn.Nullable {
							postgresqlColumnArgs.Nullable = pulumi.Bool(true)
						}
						if postgresqlColumn.OrdinalPosition > 0 {
							postgresqlColumnArgs.OrdinalPosition = pulumi.Int(int(postgresqlColumn.OrdinalPosition))
						}
						if postgresqlColumn.PrimaryKey {
							postgresqlColumnArgs.PrimaryKey = pulumi.Bool(true)
						}
						postgresqlColumns = append(postgresqlColumns, postgresqlColumnArgs)
					}
					postgresqlTableArgs.PostgresqlColumns = postgresqlColumns
				}
				postgresqlTables = append(postgresqlTables, postgresqlTableArgs)
			}
			postgresqlSchemaArgs.PostgresqlTables = postgresqlTables
		}
		postgresqlSchemas = append(postgresqlSchemas, postgresqlSchemaArgs)
	}
	return &datastream.StreamBackfillAllPostgresqlExcludedObjectsArgs{PostgresqlSchemas: postgresqlSchemas}
}

func oracleIncludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamOracleRdbms) *datastream.StreamSourceConfigOracleSourceConfigIncludeObjectsArgs {
	if set == nil {
		return nil
	}
	oracleSchemas := datastream.StreamSourceConfigOracleSourceConfigIncludeObjectsOracleSchemaArray{}
	for _, oracleSchema := range set.OracleSchemas {
		oracleSchemaArgs := &datastream.StreamSourceConfigOracleSourceConfigIncludeObjectsOracleSchemaArgs{}
		if oracleSchema.Schema != "" {
			oracleSchemaArgs.Schema = pulumi.String(oracleSchema.Schema)
		}
		if len(oracleSchema.OracleTables) > 0 {
			oracleTables := datastream.StreamSourceConfigOracleSourceConfigIncludeObjectsOracleSchemaOracleTableArray{}
			for _, oracleTable := range oracleSchema.OracleTables {
				oracleTableArgs := &datastream.StreamSourceConfigOracleSourceConfigIncludeObjectsOracleSchemaOracleTableArgs{}
				if oracleTable.Table != "" {
					oracleTableArgs.Table = pulumi.String(oracleTable.Table)
				}
				if len(oracleTable.OracleColumns) > 0 {
					oracleColumns := datastream.StreamSourceConfigOracleSourceConfigIncludeObjectsOracleSchemaOracleTableOracleColumnArray{}
					for _, oracleColumn := range oracleTable.OracleColumns {
						oracleColumnArgs := &datastream.StreamSourceConfigOracleSourceConfigIncludeObjectsOracleSchemaOracleTableOracleColumnArgs{}
						if oracleColumn.Column != "" {
							oracleColumnArgs.Column = pulumi.String(oracleColumn.Column)
						}
						if oracleColumn.DataType != "" {
							oracleColumnArgs.DataType = pulumi.String(oracleColumn.DataType)
						}
						oracleColumns = append(oracleColumns, oracleColumnArgs)
					}
					oracleTableArgs.OracleColumns = oracleColumns
				}
				oracleTables = append(oracleTables, oracleTableArgs)
			}
			oracleSchemaArgs.OracleTables = oracleTables
		}
		oracleSchemas = append(oracleSchemas, oracleSchemaArgs)
	}
	return &datastream.StreamSourceConfigOracleSourceConfigIncludeObjectsArgs{OracleSchemas: oracleSchemas}
}

func oracleExcludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamOracleRdbms) *datastream.StreamSourceConfigOracleSourceConfigExcludeObjectsArgs {
	if set == nil {
		return nil
	}
	oracleSchemas := datastream.StreamSourceConfigOracleSourceConfigExcludeObjectsOracleSchemaArray{}
	for _, oracleSchema := range set.OracleSchemas {
		oracleSchemaArgs := &datastream.StreamSourceConfigOracleSourceConfigExcludeObjectsOracleSchemaArgs{}
		if oracleSchema.Schema != "" {
			oracleSchemaArgs.Schema = pulumi.String(oracleSchema.Schema)
		}
		if len(oracleSchema.OracleTables) > 0 {
			oracleTables := datastream.StreamSourceConfigOracleSourceConfigExcludeObjectsOracleSchemaOracleTableArray{}
			for _, oracleTable := range oracleSchema.OracleTables {
				oracleTableArgs := &datastream.StreamSourceConfigOracleSourceConfigExcludeObjectsOracleSchemaOracleTableArgs{}
				if oracleTable.Table != "" {
					oracleTableArgs.Table = pulumi.String(oracleTable.Table)
				}
				if len(oracleTable.OracleColumns) > 0 {
					oracleColumns := datastream.StreamSourceConfigOracleSourceConfigExcludeObjectsOracleSchemaOracleTableOracleColumnArray{}
					for _, oracleColumn := range oracleTable.OracleColumns {
						oracleColumnArgs := &datastream.StreamSourceConfigOracleSourceConfigExcludeObjectsOracleSchemaOracleTableOracleColumnArgs{}
						if oracleColumn.Column != "" {
							oracleColumnArgs.Column = pulumi.String(oracleColumn.Column)
						}
						if oracleColumn.DataType != "" {
							oracleColumnArgs.DataType = pulumi.String(oracleColumn.DataType)
						}
						oracleColumns = append(oracleColumns, oracleColumnArgs)
					}
					oracleTableArgs.OracleColumns = oracleColumns
				}
				oracleTables = append(oracleTables, oracleTableArgs)
			}
			oracleSchemaArgs.OracleTables = oracleTables
		}
		oracleSchemas = append(oracleSchemas, oracleSchemaArgs)
	}
	return &datastream.StreamSourceConfigOracleSourceConfigExcludeObjectsArgs{OracleSchemas: oracleSchemas}
}

func oracleBackfillExcludedObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamOracleRdbms) *datastream.StreamBackfillAllOracleExcludedObjectsArgs {
	if set == nil {
		return nil
	}
	oracleSchemas := datastream.StreamBackfillAllOracleExcludedObjectsOracleSchemaArray{}
	for _, oracleSchema := range set.OracleSchemas {
		oracleSchemaArgs := &datastream.StreamBackfillAllOracleExcludedObjectsOracleSchemaArgs{}
		if oracleSchema.Schema != "" {
			oracleSchemaArgs.Schema = pulumi.String(oracleSchema.Schema)
		}
		if len(oracleSchema.OracleTables) > 0 {
			oracleTables := datastream.StreamBackfillAllOracleExcludedObjectsOracleSchemaOracleTableArray{}
			for _, oracleTable := range oracleSchema.OracleTables {
				oracleTableArgs := &datastream.StreamBackfillAllOracleExcludedObjectsOracleSchemaOracleTableArgs{}
				if oracleTable.Table != "" {
					oracleTableArgs.Table = pulumi.String(oracleTable.Table)
				}
				if len(oracleTable.OracleColumns) > 0 {
					oracleColumns := datastream.StreamBackfillAllOracleExcludedObjectsOracleSchemaOracleTableOracleColumnArray{}
					for _, oracleColumn := range oracleTable.OracleColumns {
						oracleColumnArgs := &datastream.StreamBackfillAllOracleExcludedObjectsOracleSchemaOracleTableOracleColumnArgs{}
						if oracleColumn.Column != "" {
							oracleColumnArgs.Column = pulumi.String(oracleColumn.Column)
						}
						if oracleColumn.DataType != "" {
							oracleColumnArgs.DataType = pulumi.String(oracleColumn.DataType)
						}
						oracleColumns = append(oracleColumns, oracleColumnArgs)
					}
					oracleTableArgs.OracleColumns = oracleColumns
				}
				oracleTables = append(oracleTables, oracleTableArgs)
			}
			oracleSchemaArgs.OracleTables = oracleTables
		}
		oracleSchemas = append(oracleSchemas, oracleSchemaArgs)
	}
	return &datastream.StreamBackfillAllOracleExcludedObjectsArgs{OracleSchemas: oracleSchemas}
}

func sqlServerIncludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSqlServerRdbms) *datastream.StreamSourceConfigSqlServerSourceConfigIncludeObjectsArgs {
	if set == nil {
		return nil
	}
	schemas := datastream.StreamSourceConfigSqlServerSourceConfigIncludeObjectsSchemaArray{}
	for _, schema := range set.Schemas {
		schemaArgs := &datastream.StreamSourceConfigSqlServerSourceConfigIncludeObjectsSchemaArgs{}
		if schema.Schema != "" {
			schemaArgs.Schema = pulumi.String(schema.Schema)
		}
		if len(schema.Tables) > 0 {
			tables := datastream.StreamSourceConfigSqlServerSourceConfigIncludeObjectsSchemaTableArray{}
			for _, table := range schema.Tables {
				tableArgs := &datastream.StreamSourceConfigSqlServerSourceConfigIncludeObjectsSchemaTableArgs{}
				if table.Table != "" {
					tableArgs.Table = pulumi.String(table.Table)
				}
				if len(table.Columns) > 0 {
					columns := datastream.StreamSourceConfigSqlServerSourceConfigIncludeObjectsSchemaTableColumnArray{}
					for _, column := range table.Columns {
						columnArgs := &datastream.StreamSourceConfigSqlServerSourceConfigIncludeObjectsSchemaTableColumnArgs{}
						if column.Column != "" {
							columnArgs.Column = pulumi.String(column.Column)
						}
						if column.DataType != "" {
							columnArgs.DataType = pulumi.String(column.DataType)
						}
						columns = append(columns, columnArgs)
					}
					tableArgs.Columns = columns
				}
				tables = append(tables, tableArgs)
			}
			schemaArgs.Tables = tables
		}
		schemas = append(schemas, schemaArgs)
	}
	return &datastream.StreamSourceConfigSqlServerSourceConfigIncludeObjectsArgs{Schemas: schemas}
}

func sqlServerExcludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSqlServerRdbms) *datastream.StreamSourceConfigSqlServerSourceConfigExcludeObjectsArgs {
	if set == nil {
		return nil
	}
	schemas := datastream.StreamSourceConfigSqlServerSourceConfigExcludeObjectsSchemaArray{}
	for _, schema := range set.Schemas {
		schemaArgs := &datastream.StreamSourceConfigSqlServerSourceConfigExcludeObjectsSchemaArgs{}
		if schema.Schema != "" {
			schemaArgs.Schema = pulumi.String(schema.Schema)
		}
		if len(schema.Tables) > 0 {
			tables := datastream.StreamSourceConfigSqlServerSourceConfigExcludeObjectsSchemaTableArray{}
			for _, table := range schema.Tables {
				tableArgs := &datastream.StreamSourceConfigSqlServerSourceConfigExcludeObjectsSchemaTableArgs{}
				if table.Table != "" {
					tableArgs.Table = pulumi.String(table.Table)
				}
				if len(table.Columns) > 0 {
					columns := datastream.StreamSourceConfigSqlServerSourceConfigExcludeObjectsSchemaTableColumnArray{}
					for _, column := range table.Columns {
						columnArgs := &datastream.StreamSourceConfigSqlServerSourceConfigExcludeObjectsSchemaTableColumnArgs{}
						if column.Column != "" {
							columnArgs.Column = pulumi.String(column.Column)
						}
						if column.DataType != "" {
							columnArgs.DataType = pulumi.String(column.DataType)
						}
						columns = append(columns, columnArgs)
					}
					tableArgs.Columns = columns
				}
				tables = append(tables, tableArgs)
			}
			schemaArgs.Tables = tables
		}
		schemas = append(schemas, schemaArgs)
	}
	return &datastream.StreamSourceConfigSqlServerSourceConfigExcludeObjectsArgs{Schemas: schemas}
}

func sqlServerBackfillExcludedObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSqlServerRdbms) *datastream.StreamBackfillAllSqlServerExcludedObjectsArgs {
	if set == nil {
		return nil
	}
	schemas := datastream.StreamBackfillAllSqlServerExcludedObjectsSchemaArray{}
	for _, schema := range set.Schemas {
		schemaArgs := &datastream.StreamBackfillAllSqlServerExcludedObjectsSchemaArgs{}
		if schema.Schema != "" {
			schemaArgs.Schema = pulumi.String(schema.Schema)
		}
		if len(schema.Tables) > 0 {
			tables := datastream.StreamBackfillAllSqlServerExcludedObjectsSchemaTableArray{}
			for _, table := range schema.Tables {
				tableArgs := &datastream.StreamBackfillAllSqlServerExcludedObjectsSchemaTableArgs{}
				if table.Table != "" {
					tableArgs.Table = pulumi.String(table.Table)
				}
				if len(table.Columns) > 0 {
					columns := datastream.StreamBackfillAllSqlServerExcludedObjectsSchemaTableColumnArray{}
					for _, column := range table.Columns {
						columnArgs := &datastream.StreamBackfillAllSqlServerExcludedObjectsSchemaTableColumnArgs{}
						if column.Column != "" {
							columnArgs.Column = pulumi.String(column.Column)
						}
						if column.DataType != "" {
							columnArgs.DataType = pulumi.String(column.DataType)
						}
						columns = append(columns, columnArgs)
					}
					tableArgs.Columns = columns
				}
				tables = append(tables, tableArgs)
			}
			schemaArgs.Tables = tables
		}
		schemas = append(schemas, schemaArgs)
	}
	return &datastream.StreamBackfillAllSqlServerExcludedObjectsArgs{Schemas: schemas}
}

func mongodbIncludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamMongodbCluster) *datastream.StreamSourceConfigMongodbSourceConfigIncludeObjectsArgs {
	if set == nil {
		return nil
	}
	databases := datastream.StreamSourceConfigMongodbSourceConfigIncludeObjectsDatabaseArray{}
	for _, database := range set.Databases {
		databaseArgs := &datastream.StreamSourceConfigMongodbSourceConfigIncludeObjectsDatabaseArgs{}
		if database.Database != "" {
			databaseArgs.Database = pulumi.String(database.Database)
		}
		if len(database.Collections) > 0 {
			collections := datastream.StreamSourceConfigMongodbSourceConfigIncludeObjectsDatabaseCollectionArray{}
			for _, collection := range database.Collections {
				collectionArgs := &datastream.StreamSourceConfigMongodbSourceConfigIncludeObjectsDatabaseCollectionArgs{}
				if collection.Collection != "" {
					collectionArgs.Collection = pulumi.String(collection.Collection)
				}
				if len(collection.Fields) > 0 {
					fields := datastream.StreamSourceConfigMongodbSourceConfigIncludeObjectsDatabaseCollectionFieldArray{}
					for _, field := range collection.Fields {
						fieldArgs := &datastream.StreamSourceConfigMongodbSourceConfigIncludeObjectsDatabaseCollectionFieldArgs{}
						if field.Field != "" {
							fieldArgs.Field = pulumi.String(field.Field)
						}
						fields = append(fields, fieldArgs)
					}
					collectionArgs.Fields = fields
				}
				collections = append(collections, collectionArgs)
			}
			databaseArgs.Collections = collections
		}
		databases = append(databases, databaseArgs)
	}
	return &datastream.StreamSourceConfigMongodbSourceConfigIncludeObjectsArgs{Databases: databases}
}

func mongodbExcludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamMongodbCluster) *datastream.StreamSourceConfigMongodbSourceConfigExcludeObjectsArgs {
	if set == nil {
		return nil
	}
	databases := datastream.StreamSourceConfigMongodbSourceConfigExcludeObjectsDatabaseArray{}
	for _, database := range set.Databases {
		databaseArgs := &datastream.StreamSourceConfigMongodbSourceConfigExcludeObjectsDatabaseArgs{}
		if database.Database != "" {
			databaseArgs.Database = pulumi.String(database.Database)
		}
		if len(database.Collections) > 0 {
			collections := datastream.StreamSourceConfigMongodbSourceConfigExcludeObjectsDatabaseCollectionArray{}
			for _, collection := range database.Collections {
				collectionArgs := &datastream.StreamSourceConfigMongodbSourceConfigExcludeObjectsDatabaseCollectionArgs{}
				if collection.Collection != "" {
					collectionArgs.Collection = pulumi.String(collection.Collection)
				}
				if len(collection.Fields) > 0 {
					fields := datastream.StreamSourceConfigMongodbSourceConfigExcludeObjectsDatabaseCollectionFieldArray{}
					for _, field := range collection.Fields {
						fieldArgs := &datastream.StreamSourceConfigMongodbSourceConfigExcludeObjectsDatabaseCollectionFieldArgs{}
						if field.Field != "" {
							fieldArgs.Field = pulumi.String(field.Field)
						}
						fields = append(fields, fieldArgs)
					}
					collectionArgs.Fields = fields
				}
				collections = append(collections, collectionArgs)
			}
			databaseArgs.Collections = collections
		}
		databases = append(databases, databaseArgs)
	}
	return &datastream.StreamSourceConfigMongodbSourceConfigExcludeObjectsArgs{Databases: databases}
}

func mongodbBackfillExcludedObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamMongodbCluster) *datastream.StreamBackfillAllMongodbExcludedObjectsArgs {
	if set == nil {
		return nil
	}
	databases := datastream.StreamBackfillAllMongodbExcludedObjectsDatabaseArray{}
	for _, database := range set.Databases {
		databaseArgs := &datastream.StreamBackfillAllMongodbExcludedObjectsDatabaseArgs{}
		if database.Database != "" {
			databaseArgs.Database = pulumi.String(database.Database)
		}
		if len(database.Collections) > 0 {
			collections := datastream.StreamBackfillAllMongodbExcludedObjectsDatabaseCollectionArray{}
			for _, collection := range database.Collections {
				collectionArgs := &datastream.StreamBackfillAllMongodbExcludedObjectsDatabaseCollectionArgs{}
				if collection.Collection != "" {
					collectionArgs.Collection = pulumi.String(collection.Collection)
				}
				if len(collection.Fields) > 0 {
					fields := datastream.StreamBackfillAllMongodbExcludedObjectsDatabaseCollectionFieldArray{}
					for _, field := range collection.Fields {
						fieldArgs := &datastream.StreamBackfillAllMongodbExcludedObjectsDatabaseCollectionFieldArgs{}
						if field.Field != "" {
							fieldArgs.Field = pulumi.String(field.Field)
						}
						fields = append(fields, fieldArgs)
					}
					collectionArgs.Fields = fields
				}
				collections = append(collections, collectionArgs)
			}
			databaseArgs.Collections = collections
		}
		databases = append(databases, databaseArgs)
	}
	return &datastream.StreamBackfillAllMongodbExcludedObjectsArgs{Databases: databases}
}

func salesforceIncludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSalesforceOrg) *datastream.StreamSourceConfigSalesforceSourceConfigIncludeObjectsArgs {
	if set == nil {
		return nil
	}
	objects := datastream.StreamSourceConfigSalesforceSourceConfigIncludeObjectsObjectArray{}
	for _, object := range set.Objects {
		objectArgs := &datastream.StreamSourceConfigSalesforceSourceConfigIncludeObjectsObjectArgs{}
		if object.ObjectName != "" {
			objectArgs.ObjectName = pulumi.String(object.ObjectName)
		}
		if len(object.Fields) > 0 {
			fields := datastream.StreamSourceConfigSalesforceSourceConfigIncludeObjectsObjectFieldArray{}
			for _, field := range object.Fields {
				fieldArgs := &datastream.StreamSourceConfigSalesforceSourceConfigIncludeObjectsObjectFieldArgs{}
				if field.Name != "" {
					fieldArgs.Name = pulumi.String(field.Name)
				}
				fields = append(fields, fieldArgs)
			}
			objectArgs.Fields = fields
		}
		objects = append(objects, objectArgs)
	}
	return &datastream.StreamSourceConfigSalesforceSourceConfigIncludeObjectsArgs{Objects: objects}
}

func salesforceExcludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSalesforceOrg) *datastream.StreamSourceConfigSalesforceSourceConfigExcludeObjectsArgs {
	if set == nil {
		return nil
	}
	objects := datastream.StreamSourceConfigSalesforceSourceConfigExcludeObjectsObjectArray{}
	for _, object := range set.Objects {
		objectArgs := &datastream.StreamSourceConfigSalesforceSourceConfigExcludeObjectsObjectArgs{}
		if object.ObjectName != "" {
			objectArgs.ObjectName = pulumi.String(object.ObjectName)
		}
		if len(object.Fields) > 0 {
			fields := datastream.StreamSourceConfigSalesforceSourceConfigExcludeObjectsObjectFieldArray{}
			for _, field := range object.Fields {
				fieldArgs := &datastream.StreamSourceConfigSalesforceSourceConfigExcludeObjectsObjectFieldArgs{}
				if field.Name != "" {
					fieldArgs.Name = pulumi.String(field.Name)
				}
				fields = append(fields, fieldArgs)
			}
			objectArgs.Fields = fields
		}
		objects = append(objects, objectArgs)
	}
	return &datastream.StreamSourceConfigSalesforceSourceConfigExcludeObjectsArgs{Objects: objects}
}

func salesforceBackfillExcludedObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSalesforceOrg) *datastream.StreamBackfillAllSalesforceExcludedObjectsArgs {
	if set == nil {
		return nil
	}
	objects := datastream.StreamBackfillAllSalesforceExcludedObjectsObjectArray{}
	for _, object := range set.Objects {
		objectArgs := &datastream.StreamBackfillAllSalesforceExcludedObjectsObjectArgs{}
		if object.ObjectName != "" {
			objectArgs.ObjectName = pulumi.String(object.ObjectName)
		}
		if len(object.Fields) > 0 {
			fields := datastream.StreamBackfillAllSalesforceExcludedObjectsObjectFieldArray{}
			for _, field := range object.Fields {
				fieldArgs := &datastream.StreamBackfillAllSalesforceExcludedObjectsObjectFieldArgs{}
				if field.Name != "" {
					fieldArgs.Name = pulumi.String(field.Name)
				}
				fields = append(fields, fieldArgs)
			}
			objectArgs.Fields = fields
		}
		objects = append(objects, objectArgs)
	}
	return &datastream.StreamBackfillAllSalesforceExcludedObjectsArgs{Objects: objects}
}

func spannerIncludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSpannerDatabase) *datastream.StreamSourceConfigSpannerSourceConfigIncludeObjectsArgs {
	if set == nil {
		return nil
	}
	schemas := datastream.StreamSourceConfigSpannerSourceConfigIncludeObjectsSchemaArray{}
	for _, schema := range set.Schemas {
		schemaArgs := &datastream.StreamSourceConfigSpannerSourceConfigIncludeObjectsSchemaArgs{}
		if schema.Schema != "" {
			schemaArgs.Schema = pulumi.String(schema.Schema)
		}
		if len(schema.Tables) > 0 {
			tables := datastream.StreamSourceConfigSpannerSourceConfigIncludeObjectsSchemaTableArray{}
			for _, table := range schema.Tables {
				tableArgs := &datastream.StreamSourceConfigSpannerSourceConfigIncludeObjectsSchemaTableArgs{}
				if table.Table != "" {
					tableArgs.Table = pulumi.String(table.Table)
				}
				if len(table.Columns) > 0 {
					columns := datastream.StreamSourceConfigSpannerSourceConfigIncludeObjectsSchemaTableColumnArray{}
					for _, column := range table.Columns {
						columnArgs := &datastream.StreamSourceConfigSpannerSourceConfigIncludeObjectsSchemaTableColumnArgs{}
						if column.Column != "" {
							columnArgs.Column = pulumi.String(column.Column)
						}
						columns = append(columns, columnArgs)
					}
					tableArgs.Columns = columns
				}
				tables = append(tables, tableArgs)
			}
			schemaArgs.Tables = tables
		}
		schemas = append(schemas, schemaArgs)
	}
	return &datastream.StreamSourceConfigSpannerSourceConfigIncludeObjectsArgs{Schemas: schemas}
}

func spannerExcludeObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSpannerDatabase) *datastream.StreamSourceConfigSpannerSourceConfigExcludeObjectsArgs {
	if set == nil {
		return nil
	}
	schemas := datastream.StreamSourceConfigSpannerSourceConfigExcludeObjectsSchemaArray{}
	for _, schema := range set.Schemas {
		schemaArgs := &datastream.StreamSourceConfigSpannerSourceConfigExcludeObjectsSchemaArgs{}
		if schema.Schema != "" {
			schemaArgs.Schema = pulumi.String(schema.Schema)
		}
		if len(schema.Tables) > 0 {
			tables := datastream.StreamSourceConfigSpannerSourceConfigExcludeObjectsSchemaTableArray{}
			for _, table := range schema.Tables {
				tableArgs := &datastream.StreamSourceConfigSpannerSourceConfigExcludeObjectsSchemaTableArgs{}
				if table.Table != "" {
					tableArgs.Table = pulumi.String(table.Table)
				}
				if len(table.Columns) > 0 {
					columns := datastream.StreamSourceConfigSpannerSourceConfigExcludeObjectsSchemaTableColumnArray{}
					for _, column := range table.Columns {
						columnArgs := &datastream.StreamSourceConfigSpannerSourceConfigExcludeObjectsSchemaTableColumnArgs{}
						if column.Column != "" {
							columnArgs.Column = pulumi.String(column.Column)
						}
						columns = append(columns, columnArgs)
					}
					tableArgs.Columns = columns
				}
				tables = append(tables, tableArgs)
			}
			schemaArgs.Tables = tables
		}
		schemas = append(schemas, schemaArgs)
	}
	return &datastream.StreamSourceConfigSpannerSourceConfigExcludeObjectsArgs{Schemas: schemas}
}

func spannerBackfillExcludedObjects(set *gcpdatastreamstreamv1alpha1.GcpDatastreamStreamSpannerDatabase) *datastream.StreamBackfillAllSpannerExcludedObjectsArgs {
	if set == nil {
		return nil
	}
	schemas := datastream.StreamBackfillAllSpannerExcludedObjectsSchemaArray{}
	for _, schema := range set.Schemas {
		schemaArgs := &datastream.StreamBackfillAllSpannerExcludedObjectsSchemaArgs{}
		if schema.Schema != "" {
			schemaArgs.Schema = pulumi.String(schema.Schema)
		}
		if len(schema.Tables) > 0 {
			tables := datastream.StreamBackfillAllSpannerExcludedObjectsSchemaTableArray{}
			for _, table := range schema.Tables {
				tableArgs := &datastream.StreamBackfillAllSpannerExcludedObjectsSchemaTableArgs{}
				if table.Table != "" {
					tableArgs.Table = pulumi.String(table.Table)
				}
				if len(table.Columns) > 0 {
					columns := datastream.StreamBackfillAllSpannerExcludedObjectsSchemaTableColumnArray{}
					for _, column := range table.Columns {
						columnArgs := &datastream.StreamBackfillAllSpannerExcludedObjectsSchemaTableColumnArgs{}
						if column.Column != "" {
							columnArgs.Column = pulumi.String(column.Column)
						}
						columns = append(columns, columnArgs)
					}
					tableArgs.Columns = columns
				}
				tables = append(tables, tableArgs)
			}
			schemaArgs.Tables = tables
		}
		schemas = append(schemas, schemaArgs)
	}
	return &datastream.StreamBackfillAllSpannerExcludedObjectsArgs{Schemas: schemas}
}
