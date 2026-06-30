output "endpoint" {
  description = "The URL of the Kubernetes API server for the EKS cluster."
  value       = aws_eks_cluster.this.endpoint
}

output "cluster_ca_certificate" {
  description = "The Base64-encoded certificate authority for the cluster."
  value       = aws_eks_cluster.this.certificate_authority[0].data
}

output "cluster_security_group_id" {
  description = "The ID of the security group created by EKS for the cluster control plane."
  value       = aws_eks_cluster.this.vpc_config[0].cluster_security_group_id
}

output "oidc_issuer_url" {
  description = "The URL of the OpenID Connect issuer for the cluster (used for IAM Roles for Service Accounts)."
  value       = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

output "cluster_arn" {
  description = "The Amazon Resource Name of the EKS cluster."
  value       = aws_eks_cluster.this.arn
}

output "name" {
  description = "The EKS cluster name."
  value       = aws_eks_cluster.this.name
}


