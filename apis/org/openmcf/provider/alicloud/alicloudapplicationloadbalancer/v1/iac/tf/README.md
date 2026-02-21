# AliCloudApplicationLoadBalancer Terraform Module

This module deploys an Alibaba Cloud Application Load Balancer (ALB) with server groups and listeners using Terraform.

## Usage

```hcl
module "alb" {
  source = "./path/to/module"

  metadata = {
    name = "my-alb"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    region = "cn-hangzhou"
    vpc_id = "vpc-abc123"
    zone_mappings = [
      { zone_id = "cn-hangzhou-a", vswitch_id = "vsw-aaa" },
      { zone_id = "cn-hangzhou-b", vswitch_id = "vsw-bbb" },
    ]
    server_groups = [{
      name = "web-backend"
      health_check_config = {
        health_check_enabled = true
        health_check_path    = "/health"
      }
    }]
    listeners = [{
      listener_port                    = 80
      listener_protocol                = "HTTP"
      default_action_server_group_name = "web-backend"
    }]
  }
}
```

## Outputs

| Output | Type | Description |
| --- | --- | --- |
| `load_balancer_id` | string | ALB instance ID |
| `dns_name` | string | ALB DNS name |
| `server_group_ids` | map(string) | Server group name to ID mapping |
