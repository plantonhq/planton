# AWS RDS Proxy

Deploys an RDS Proxy — the managed connection pool that multiplexes thousands of application connections onto the handful a database can actually hold, the standard cure for Lambda-to-RDS connection storms and the graceful-failover front for Aurora. The proxy authenticates clients with IAM tokens or native credentials, signs in to the database with credentials read from Secrets Manager under an IAM role you provide, and fronts exactly one RDS instance or one Aurora cluster. Three things are fixed for life at the provider — the engine family, the proxy's subnets, and both network-type dials — and changing any of them replaces the proxy along with its endpoint DNS names.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DB Proxy** — the proxy itself: engine family, network placement, Secrets Manager sign-ins with per-secret IAM auth posture, TLS enforcement, idle timeout, and debug logging. The proxy name is wired from `metadata.name`
- **Default Target Group** — the proxy's built-in pool, tuned by `connectionPool` (max connections percent, idle ceiling, borrow timeout, init query, pinning filters). It has no delete of its own; destroying the proxy takes it along
- **Proxy Endpoints** — one per `endpoints` entry; read-only endpoints for Aurora reader farms, or endpoints in other network configurations
- **Proxy Target** — created only when `target` is set; registers the RDS instance or Aurora cluster the proxy fronts

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with RDS, IAM, and Secrets Manager permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **A Secrets Manager secret** holding the database credentials in the standard RDS shape (username and password JSON) — each `auth` entry references one via `secretArn`.
- **An IAM role the proxy can assume** — its trust policy must allow rds.amazonaws.com, and its policy needs `secretsmanager:GetSecretValue` on every auth secret, plus `kms:Decrypt` when the secrets use a customer-managed key. Referenced via `roleArn`.
- **At least two subnets in different availability zones** — AWS rejects fewer; referenced via `vpcSubnetIds`.
- **The database** — the RDS instance or Aurora cluster the proxy will front, in the same VPC reachability domain.

## Deploy

### Console

Open the deployment store, find **AWS RDS Proxy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: engine family and network placement, the Secrets Manager sign-ins, then pool tuning, additional endpoints, and the database target. Start from the **Lambda-to-PostgreSQL Pool** preset in the [Presets](#presets) tab for the classic shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRdsProxy
metadata:
  name: orders-db-proxy
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  engineFamily: POSTGRESQL
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: orders-proxy-secrets
      fieldPath: status.outputs.role_arn
  vpcSubnetIds:
    - value: subnet-0123456789abcdef0
    - value: subnet-0123456789abcdef1
  vpcSecurityGroupIds:
    - value: sg-0123456789abcdef0
  auth:
    - secretArn:
        valueFrom:
          kind: AwsSecretsManagerSecret
          name: orders-db-creds
          fieldPath: status.outputs.secret_arn
      iamAuth: REQUIRED
  requireTls: true
  connectionPool:
    maxConnectionsPercent: 90
    maxIdleConnectionsPercent: 10
  target:
    dbInstanceIdentifier:
      valueFrom:
        kind: AwsRdsInstance
        name: orders-db
        fieldPath: status.outputs.instance_identifier
```

```shell
planton apply -f aws-rds-proxy.yaml
```

This creates a PostgreSQL proxy fronting the referenced instance, with IAM-token client auth over enforced TLS and a pool capped at 90% of the database's connections. An Infra Job tracks the provisioning in real time.

### InfraChart

When the proxy deploys alongside its database in one chart, wire the target via ValueFromRef:

```yaml
spec:
  region: us-east-1
  engineFamily: POSTGRESQL
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: orders-proxy-secrets
      fieldPath: status.outputs.role_arn
  vpcSubnetIds:
    - value: subnet-0123456789abcdef0
    - value: subnet-0123456789abcdef1
  auth:
    - secretArn:
        valueFrom:
          kind: AwsSecretsManagerSecret
          name: orders-db-creds
          fieldPath: status.outputs.secret_arn
  target:
    dbInstanceIdentifier:
      valueFrom:
        kind: AwsRdsInstance
        name: orders-db
        fieldPath: status.outputs.instance_identifier
```

The InfraPipeline resolves the dependency graph — role, secret, and database first — then registers the target on the freshly created proxy.

## Key Configuration

These are the most important decisions when configuring a proxy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Know the replacement boundary** — `engineFamily`, `vpcSubnetIds`, `endpointNetworkType`, and `targetConnectionNetworkType` are fixed for life; changing any of them replaces the proxy, and the endpoint DNS names change with it. Roll application configuration in the same change. Everything else — pool tuning, sign-ins, endpoints, the target — updates in place or re-registers.

**The role is the usual first failure** — a proxy stuck in INCOMPATIBLE_AUTH or never reaching AVAILABLE almost always has a role problem: the trust policy must name rds.amazonaws.com, and the policy needs `secretsmanager:GetSecretValue` on every auth secret plus `kms:Decrypt` for customer-managed keys. Check the role before anything else.

**IAM auth is a TLS decision too** — `iamAuth: REQUIRED` only works over TLS, so pair it with `requireTls`. Clients then present short-lived auth tokens (`rds generate-db-auth-token`) instead of passwords; the proxy still signs in to the database with the secret's credentials either way.

**Session pinning is the silent multiplexing killer** — the proxy multiplexes only while connections stay interchangeable. Prepared statements, session variables, and temp tables pin a connection to one client and quietly turn the pool into 1:1 passthrough. `sessionPinningFilters: [EXCLUDE_VARIABLE_SETS]` relaxes the variable-set trigger for the MySQL family; beyond that, pinning is an application-behavior fix, not a proxy dial.

**Watch the idle ceiling on small databases** — the pool defaults (`maxIdleConnectionsPercent` 50) are tuned for big instances. On a db.t-class instance with low max_connections, a 50% idle ceiling can starve the database's own headroom — drop it to 10-20% so the proxy returns connections aggressively.

**Endpoints multiply networks, not capacity** — additional `endpoints` exist for network topology (a second VPC's subnets, IPv6) and Aurora read-only routing; they add no throughput. `READ_ONLY` endpoints distribute across Aurora readers — against a plain instance target they have nothing to distribute to.

**Debug logging captures query text** — `debugLogging` logs the SQL statements the proxy forwards, sensitive values included. Turn it on for active troubleshooting and back off after.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsSubnet** | `vpcSubnetIds[]` / `endpoints[].vpcSubnetIds[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `vpcSecurityGroupIds[]` / `endpoints[].vpcSecurityGroupIds[]` | `status.outputs.security_group_id` |
| **AwsSecretsManagerSecret** | `auth[].secretArn` | `status.outputs.secret_arn` |
| **AwsRdsInstance** | `target.dbInstanceIdentifier` | `status.outputs.instance_identifier` |
| **AwsRdsCluster** | `target.dbClusterIdentifier` | `status.outputs.cluster_identifier` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint` | The proxy's default endpoint DNS name | Application database-host configuration — point apps here instead of at the database |
| `endpoint_addresses` | Additional endpoints' DNS names keyed by endpoint name | Read-path configuration for Aurora reader endpoints |
| `proxy_arn` | The proxy's ARN | IAM policies scoping proxy management and `rds-db:connect` grants |

`proxy_name`, `default_target_group_arn`, `default_target_group_name`, `endpoint_arns`, `target_type`, and `target_rds_resource_id` are also present — import addresses and audit echoes rather than composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Lambda connection storm control** — functions storming a PostgreSQL instance get pooled through the proxy with IAM-token auth over enforced TLS: no passwords in function config, no connection exhaustion under burst. Wire the functions' database host to the `endpoint` output and grant their execution role `rds-db:connect`. Start from the **Lambda-to-PostgreSQL Pool** preset.

**Aurora fronted end to end** — writes through the default endpoint, reads through a `READ_ONLY` endpoint that distributes across the cluster's replicas and rides failovers without dropping clients. Relax MySQL's variable-set pinning so multiplexing stays effective. Start from the **Aurora Cluster with a Reader Endpoint** preset.

**One proxy, many database users** — multiple `auth` entries let different applications sign in as different database users through the same proxy, each backed by its own secret with its own `iamAuth` posture. Trades a shared blast radius for one pooling point and one endpoint to secure.

## Works With

- [**AWS RDS Instance**](/infra-catalog/aws-rds-instance) — the instance target, wired via `target.dbInstanceIdentifier`
- [**AWS RDS Cluster**](/infra-catalog/aws-rds-cluster) — the Aurora target whose writer/reader topology the proxy tracks
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the secrets-reading role the proxy assumes, wired via `roleArn`
- [**AWS Secrets Manager Secret**](/infra-catalog/aws-secrets-manager-secret) — the database credentials behind each `auth` entry
- [**AWS Subnet**](/infra-catalog/aws-subnet) — the proxy's network placement (at least two, in different availability zones)
- [**AWS Security Group**](/infra-catalog/aws-security-group) — must allow applications in and the database out
- [**AWS Lambda**](/infra-catalog/aws-lambda) — the archetypal client whose connection bursts the proxy absorbs
