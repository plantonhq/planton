# AWS Secrets Manager Examples

Below are several examples demonstrating how to define an AWS Secrets Manager component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: my-aws-secrets
spec:
  region: us-west-2
  secretNames:
    - "database-password"
    - "api-key"
```

This example creates basic secrets:
• Database password secret for application authentication.
• API key secret for external service integration.
• Placeholder values for initial secret creation.
• Secure storage with AWS KMS encryption.

---

## Application Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: app-secrets
spec:
  region: us-west-2
  secretNames:
    - "database-password"
    - "database-username"
    - "redis-password"
    - "jwt-secret"
    - "stripe-api-key"
    - "sendgrid-api-key"
```

This example creates application-specific secrets:
• Database credentials for application connectivity.
• Redis password for caching layer.
• JWT secret for authentication tokens.
• External API keys for third-party services.
• Comprehensive application security management.

---

## Database Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: database-secrets
spec:
  region: us-west-2
  secretNames:
    - "postgres-password"
    - "postgres-username"
    - "mysql-password"
    - "mysql-username"
    - "mongodb-connection-string"
    - "redis-auth-token"
```

This example creates database-specific secrets:
• PostgreSQL credentials for primary database.
• MySQL credentials for secondary database.
• MongoDB connection string for NoSQL database.
• Redis authentication for caching.
• Centralized database credential management.

---

## API Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: api-secrets
spec:
  region: us-west-2
  secretNames:
    - "github-token"
    - "slack-webhook-url"
    - "aws-access-key-id"
    - "aws-secret-access-key"
    - "google-api-key"
    - "twilio-auth-token"
```

This example creates API-specific secrets:
• GitHub token for repository access.
• Slack webhook for notifications.
• AWS credentials for service integration.
• Google API key for external services.
• Twilio token for SMS services.
• External service integration management.

---

## Production Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: production-secrets
spec:
  region: us-west-2
  secretNames:
    - "prod-database-password"
    - "prod-api-key"
    - "prod-jwt-secret"
    - "prod-stripe-secret-key"
    - "prod-sendgrid-api-key"
    - "prod-aws-access-key"
    - "prod-aws-secret-key"
    - "prod-redis-password"
```

This example creates production environment secrets:
• Production database credentials.
• Production API keys for external services.
• JWT secret for production authentication.
• Payment processing secrets.
• Email service credentials.
• AWS production credentials.
• Production Redis authentication.
• Comprehensive production security.

---

## Development Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: development-secrets
spec:
  region: us-west-2
  secretNames:
    - "dev-database-password"
    - "dev-api-key"
    - "dev-jwt-secret"
    - "dev-stripe-test-key"
    - "dev-redis-password"
```

This example creates development environment secrets:
• Development database credentials.
• Development API keys for testing.
• JWT secret for development authentication.
• Stripe test keys for payment testing.
• Development Redis authentication.
• Safe development environment management.

---

## Microservice Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: microservice-secrets
spec:
  region: us-west-2
  secretNames:
    - "user-service-db-password"
    - "auth-service-jwt-secret"
    - "payment-service-stripe-key"
    - "notification-service-sendgrid-key"
    - "analytics-service-google-key"
    - "storage-service-aws-key"
```

This example creates microservice-specific secrets:
• User service database credentials.
• Authentication service JWT secret.
• Payment service Stripe integration.
• Notification service email credentials.
• Analytics service Google integration.
• Storage service AWS credentials.
• Microservice architecture security.

---

## Security Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: security-secrets
spec:
  region: us-west-2
  secretNames:
    - "encryption-key"
    - "ssl-certificate-password"
    - "vpn-password"
    - "admin-password"
    - "backup-encryption-key"
    - "monitoring-api-key"
```

This example creates security-focused secrets:
• Encryption keys for data protection.
• SSL certificate passwords.
• VPN authentication credentials.
• Administrative access credentials.
• Backup encryption keys.
• Monitoring service API keys.
• Comprehensive security management.

---

## Monitoring Secrets Manager

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: monitoring-secrets
spec:
  region: us-west-2
  secretNames:
    - "datadog-api-key"
    - "newrelic-license-key"
    - "splunk-token"
    - "grafana-admin-password"
    - "prometheus-basic-auth"
    - "alertmanager-webhook-url"
```

This example creates monitoring-specific secrets:
• Datadog API key for metrics and logs.
• New Relic license for application monitoring.
• Splunk token for log aggregation.
• Grafana admin credentials.
• Prometheus authentication.
• Alert manager webhook configuration.
• Observability and monitoring security.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the secrets are active via the AWS console or by
using the AWS CLI:

```shell
aws secretsmanager list-secrets
```

For detailed secret information:

```shell
aws secretsmanager describe-secret --secret-id <your-secret-name>
```

To list secret versions:

```shell
aws secretsmanager list-secret-version-ids --secret-id <your-secret-name>
```

This will show the Secrets Manager details including secret names, ARNs, and configuration information.
