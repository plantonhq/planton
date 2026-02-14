# Alias Record to ALB

This preset creates a Route53 alias record pointing to an Application Load Balancer. Alias records are Route53's most powerful feature -- they work at the zone apex (e.g., `example.com`), incur no query charges for AWS targets, and automatically track the ALB's changing IP addresses. This is the most common DNS record pattern for web applications.

## When to Use

- Pointing a domain or subdomain to an Application Load Balancer
- Zone apex records (e.g., `example.com`) where CNAME records are not allowed by DNS specification
- Any ALB-backed application that needs a friendly DNS name

## Key Configuration Choices

- **A record type** (`type: A`) -- Alias A records resolve to the ALB's IPv4 addresses; Route53 handles the IP tracking automatically
- **Target health evaluation** (`evaluateTargetHealth: true`) -- Route53 checks whether the ALB has healthy targets before returning the alias; prevents routing to unhealthy endpoints
- **No TTL** -- Alias records inherit the target resource's TTL; the `ttl` field is not needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<route53-hosted-zone-id>` | ID of your Route53 hosted zone | AWS Route53 console or `AwsRoute53Zone` status outputs |
| `<your-domain.com>` | Domain or subdomain to point to the ALB (e.g., `example.com` or `app.example.com`) | Your domain registrar or DNS provider |
| `<alb-dns-name>` | DNS name of the ALB (e.g., `my-alb-123456.us-east-1.elb.amazonaws.com`) | AWS EC2 console or `AwsAlb` status outputs (`load_balancer_dns_name`) |
| `<alb-hosted-zone-id>` | Hosted zone ID of the ALB (AWS service zone, not your Route53 zone) | AWS EC2 console or `AwsAlb` status outputs (`load_balancer_hosted_zone_id`) |

## Related Presets

- **02-a-record** -- Use instead for simple A records pointing to static IP addresses
