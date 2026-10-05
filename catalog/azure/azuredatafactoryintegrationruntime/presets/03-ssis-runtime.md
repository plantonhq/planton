# SSIS Runtime

This preset creates the managed SSIS package runtime with an SSISDB catalog on your Azure SQL server -- the lift-and-shift home for existing SQL Server Integration Services projects.

## When to Use

- Running existing SSIS packages in Azure without rewriting them as pipelines
- Deploying SSIS projects from SSIS Studio / Visual Studio to an infra catalog
- Retiring an on-premises SQL Server that exists only to host SSIS

## Key Configuration Choices

- **Created STOPPED and unbilled** -- this definition costs nothing until you start the runtime (Studio: Manage -> Integration runtimes -> Start); node-hours bill from start to stop, so stop it when packages are not running
- **`catalogInfo` with managed-identity auth** -- omitting `administratorLogin`/`administratorPassword` authenticates as the factory's managed identity (grant it access on the SQL server); no password travels through the manifest. Azure creates the SSISDB database at the chosen `pricingTier` on first start
- **`nodeSize: Standard_D4_v3`** -- a balanced starting size; scale out with `numberOfNodes` and pack with `maxParallelExecutionsPerNode` before scaling the node size up

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-data-factory>` | The Planton name of your `AzureDataFactory` resource | Planton console (or replace `valueFrom` with `value:` and the factory's ARM ID) |
| `<your-region>` | The Azure region the SSIS nodes run in, e.g. `eastus` | Your factory's region |
| `<your-sql-server>` | The Azure SQL server hosting SSISDB | Azure portal -> SQL servers (the server's full endpoint) |

## Related Presets

- **Self-Hosted Bridge** -- wire it into this runtime's `proxy` block to reach on-premises data from SSIS packages.
- **Data Flow Compute** -- the modern transformation engine for new work.
