output "nodegroup_name" {
  description = "The name of the EKS node group."
  value       = aws_eks_node_group.this.node_group_name
}

# The ASG name is not directly exposed by the aws_eks_node_group resource.
# Keep placeholder for schema alignment.
output "asg_name" {
  description = "The Auto Scaling Group name backing the node group (if available)."
  value       = ""
}

output "remote_access_sg_id" {
  description = "Security group ID used for SSH remote access when SSH key is provided."
  value       = local.ssh_key_name != null && local.ssh_key_name != "" ? "" : ""
}

output "instance_profile_arn" {
  description = "Instance profile ARN associated with the node group instances."
  value       = ""
}



