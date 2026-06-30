output "database_name" {
  description = "Name of the Glue Data Catalog database"
  value       = aws_glue_catalog_database.this.name
}

output "database_arn" {
  description = "ARN of the Glue Data Catalog database"
  value       = aws_glue_catalog_database.this.arn
}

output "catalog_id" {
  description = "ID of the Glue Data Catalog (AWS Account ID)"
  value       = aws_glue_catalog_database.this.catalog_id
}
