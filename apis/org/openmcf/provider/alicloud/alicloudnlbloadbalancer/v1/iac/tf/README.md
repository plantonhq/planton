# AlicloudNlbLoadBalancer Terraform Module

This Terraform module provisions an Alibaba Cloud Network Load Balancer (NLB) with server groups and listeners.

## Usage

```hcl
module "nlb" {
  source = "./path/to/module"

  metadata = {
    name = "my-nlb"
  }

  spec = {
    region = "cn-hangzhou"
    vpc_id = "vpc-abc123"
    zone_mappings = [
      { zone_id = "cn-hangzhou-a", vswitch_id = "vsw-aaa" },
      { zone_id = "cn-hangzhou-b", vswitch_id = "vsw-bbb" },
    ]
    server_groups = [
      {
        name = "tcp-backend"
        health_check = {
          health_check_enabled = true
        }
      }
    ]
    listeners = [
      {
        listener_port     = 80
        listener_protocol = "TCP"
        server_group_name = "tcp-backend"
      }
    ]
  }
}
```

## Outputs

| Output | Description |
|---|---|
| `load_balancer_id` | The NLB instance ID |
| `dns_name` | DNS name assigned to the NLB |
| `server_group_ids` | Map of server group names to IDs |
