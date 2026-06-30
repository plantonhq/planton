# AliCloudApplicationLoadBalancer Pulumi Module

This module deploys an Alibaba Cloud Application Load Balancer (ALB) with server groups and listeners using Pulumi Go.

## Usage

This module is invoked by the Planton Pulumi entrypoint. It reads an `AliCloudApplicationLoadBalancerStackInput` from the Pulumi stack config and creates all resources.

## Development

```bash
cd module/
go build ./...
go vet ./...
```

## Stack Outputs

| Output | Type | Description |
| --- | --- | --- |
| `load_balancer_id` | string | ALB instance ID |
| `dns_name` | string | ALB DNS name |
| `server_group_ids` | map[string]string | Server group name to ID mapping |
