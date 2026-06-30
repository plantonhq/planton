output "load_balancer_arn" {
  description = "The ARN of the Application Load Balancer"
  value       = aws_lb.this.arn
}

output "load_balancer_name" {
  description = "The name of the Application Load Balancer"
  value       = aws_lb.this.name
}

output "load_balancer_dns_name" {
  description = "The DNS name of the Application Load Balancer"
  value       = aws_lb.this.dns_name
}

output "load_balancer_hosted_zone_id" {
  description = "The Route53 hosted zone ID of the ALB"
  value       = aws_lb.this.zone_id
}


