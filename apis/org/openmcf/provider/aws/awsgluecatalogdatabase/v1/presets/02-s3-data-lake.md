# Preset: S3 Data Lake

A Glue Data Catalog database with a default S3 storage location for an organized
data lake. Tables created in this database inherit the base S3 path, keeping
data organized under a consistent prefix.

## What This Configures

- A named database in the Glue Data Catalog.
- A description documenting the database's purpose.
- A default S3 URI so tables inherit a consistent storage prefix.

## When to Use

- Organized data lakes where all tables in a domain share an S3 prefix.
- Production data warehouses with curated ETL pipelines.
- Environments where Glue Crawlers auto-discover data under a common path.

## Customization Points

- Change `locationUri` to point to your actual S3 data lake bucket and prefix.
- Adjust `description` to document your specific data domain.
- The S3 bucket must exist before deploying this database.
