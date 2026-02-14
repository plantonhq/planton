# Web Service with ALB

This preset deploys a Fargate-based ECS service fronted by an Application Load Balancer using path-based routing. It runs 2 replicas across two Availability Zones with CloudWatch logging and a `/health` endpoint health check. This is the most common pattern for web applications and APIs on ECS.

## When to Use

- Web applications or REST APIs that need to be accessible through a load balancer
- Services that use path-based routing (e.g., `/` for the default service, `/api` for a specific API)
- Standard production deployments with high availability (2+ replicas across AZs)

## Key Configuration Choices

- **2 replicas** (`replicas: 2`) -- Minimum for high availability across AZs
- **512 CPU / 1024 MiB memory** -- Suitable for most web applications; adjust based on actual resource usage
- **Path-based routing** (`routingType: path`, `path: /`) -- Routes all traffic on the ALB to this service
- **Health check** (`/health`) -- ALB checks container health before routing traffic
- **60-second grace period** (`healthCheckGracePeriodSeconds: 60`) -- Allows containers time to start before failing health checks
- **CloudWatch logging** (`logging.enabled: true`) -- Auto-creates a log group named `/ecs/<serviceName>`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<ecs-cluster-arn>` | ARN of the ECS cluster | AWS ECS console or `AwsEcsCluster` status outputs |
| `<ecr-repo-uri>` | ECR repository URI (e.g., `123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app`) | AWS ECR console or `AwsEcrRepo` status outputs |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing traffic from the ALB | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<alb-arn>` | ARN of the Application Load Balancer | AWS EC2 console or `AwsAlb` status outputs |

## Related Presets

- **02-api-service-autoscaling** -- Use instead for APIs with hostname-based routing and CPU-based autoscaling
- **03-background-worker** -- Use instead for services that process messages from queues without ALB
