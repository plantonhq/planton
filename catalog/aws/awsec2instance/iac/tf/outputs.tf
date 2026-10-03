# Outputs flatten onto AwsEc2InstanceOutputs field-for-field;
# both engines export the same names so composition never depends on the
# engine. Address outputs are empty strings for private-only instances.

output "instance_id" {
  description = "The instance ID -- the primary handle target groups and APIs address"
  value       = aws_instance.this.id
}

output "arn" {
  description = "The ARN of the instance, for IAM policies and EventBridge rules"
  value       = aws_instance.this.arn
}

output "instance_state" {
  description = "The instance lifecycle state as of the last deploy"
  value       = aws_instance.this.instance_state
}

output "availability_zone" {
  description = "The availability zone the instance runs in"
  value       = aws_instance.this.availability_zone
}

output "private_ip" {
  description = "The primary private IPv4 address"
  value       = aws_instance.this.private_ip
}

output "private_dns" {
  description = "The private DNS hostname within the VPC"
  value       = aws_instance.this.private_dns
}

output "public_ip" {
  description = "The public IPv4 address, when one is associated (changes across stop/start; compose an AwsElasticIp for stability)"
  value       = aws_instance.this.public_ip
}

output "public_dns" {
  description = "The public DNS hostname, when a public address is associated"
  value       = aws_instance.this.public_dns
}

output "primary_network_interface_id" {
  description = "The ID of the primary network interface (eth0)"
  value       = aws_instance.this.primary_network_interface_id
}
