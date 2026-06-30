output "rds_instance_id" {
  value = aws_db_instance.this.id
}

output "rds_instance_arn" {
  value = aws_db_instance.this.arn
}

output "rds_instance_endpoint" {
  value = aws_db_instance.this.address
}

output "rds_instance_port" {
  value = aws_db_instance.this.port
}

output "rds_subnet_group" {
  value = coalesce(try(var.spec.db_subnet_group_name.value, null), try(aws_db_subnet_group.this[0].name, null))
}

output "rds_security_group" {
  value = try(element(local.ingress_sg_ids, 0), "")
}

output "rds_parameter_group" {
  value = try(var.spec.parameter_group_name, "")
}
