# Application Secrets

This preset creates a set of secrets in AWS Secrets Manager for a typical application. It provisions empty secret placeholders for database credentials, API keys, and TLS certificates. Secret values are managed separately after creation, either through the AWS console, CLI, or a secrets rotation workflow.

## When to Use

- Bootstrapping secret storage for a new application or microservice
- Declaring secrets as infrastructure alongside your deployment components
- Any workload that needs secure, auditable storage for credentials and keys

## Key Configuration Choices

- **Three secrets** -- Covers the most common secret categories for a production application: database credentials, API keys, and TLS certificates
- **Path-prefixed names** (`<app-name>/...`) -- Organizes secrets by application for easy IAM policy scoping and discoverability
- **Empty values** -- Secrets are created as empty placeholders; populate values post-creation via the AWS console, CLI, or automated rotation

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<app-name>` | Application or service name used as a path prefix (e.g., `payment-service`) | Your application naming convention |
