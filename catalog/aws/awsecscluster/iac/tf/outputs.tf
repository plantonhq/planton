# Outputs flatten onto AwsEcsClusterOutputs field-for-field;
# both engines export the same names so composition never depends on the
# engine.

output "cluster_name" {
  description = "The cluster name (mirrors metadata.name)"
  value       = aws_ecs_cluster.this.name
}

output "cluster_arn" {
  description = "The ARN of the cluster -- the join key AwsEcsService references"
  value       = aws_ecs_cluster.this.arn
}

output "capacity_provider_names" {
  description = "Every associated capacity provider: the Fargate built-ins plus the folded EC2 and managed-instances provider names, in spec order"
  value       = local.all_capacity_provider_names
}

output "capacity_provider_arns" {
  description = "The ARNs of the folded capacity providers this cluster defines (EC2 then managed-instances, each in spec order; empty for Fargate-only clusters)"
  value = concat(
    [for p in var.spec.ec2_capacity_providers : aws_ecs_capacity_provider.this[p.name].arn],
    [for p in var.spec.managed_instances_capacity_providers : aws_ecs_capacity_provider.managed_instances[p.name].arn],
  )
}
