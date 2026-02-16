# Basic Public Image

This preset creates an App Runner service from a public ECR image with all default settings. It is the simplest possible App Runner deployment -- no IAM roles, no VPC, no encryption configuration. App Runner provisions instances, handles TLS termination, and auto scales based on request concurrency.

## When to Use

- Quick prototyping or proof-of-concept deployments
- Deploying open-source or community container images from ECR Public Gallery
- Learning App Runner without needing to set up IAM roles or VPC networking
- Internal tools and demos that don't require production-grade security

## Key Configuration Choices

- **ECR Public image** (`imageRepositoryType: ECR_PUBLIC`) -- No authentication or IAM access role required. Anyone can pull public images.
- **Default instance size** (`cpu: 1024`, `memory: 2048`) -- 1 vCPU and 2 GB RAM. Suitable for most lightweight web applications.
- **Default auto scaling** -- 1 minimum instance (always warm), up to 25 maximum, scaling out when any instance exceeds 100 concurrent requests.
- **Default TCP health check** -- Checks that the port is open. Sufficient for simple services; switch to HTTP health checks for production.
- **Auto-deploy enabled** -- New pushes to the same image tag trigger automatic redeployment.
- **No VPC egress** -- Instances have outbound internet access but cannot reach resources in a private VPC.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<ecr-public-image-uri>` | Full ECR Public image URI (e.g., `public.ecr.aws/nginx/nginx:latest`) | [ECR Public Gallery](https://gallery.ecr.aws/) |
| `<application-port>` | Port your application listens on (e.g., `80`, `3000`, `8080`) | Your application's documentation or Dockerfile `EXPOSE` directive |

## Related Presets

- **02-production-vpc-encrypted** -- Use when you need VPC egress, KMS encryption, and tuned auto scaling for production workloads.
- **03-github-code-source** -- Use when deploying directly from a GitHub repository instead of a pre-built container image.
