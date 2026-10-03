# AWS AppSync API

Deploys an AWS AppSync API — a GraphQL API that resolves fields straight out of DynamoDB, Lambda, OpenSearch, EventBridge, Aurora, or any HTTPS endpoint, or an Events API giving browsers and mobile apps real-time pub/sub over WebSockets with nothing to operate. Everything that belongs to the API is managed in-line: data sources, resolvers and functions, schema types, caching, API keys, channel namespaces, and a custom domain. The spec pivots on exactly one of the two AppSync models — an API is GraphQL or Events for life.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **GraphQL API** — created only when the `graphql` arm is configured: the API with its primary and additional auth providers, the SDL schema (applied through AppSync's async schema creation), visibility, introspection and query-limit settings, X-Ray, enhanced metrics, logging, and optionally a WAF web-ACL association and the MERGED federation variant with its source-API associations
- **Events API** — created only when the `events` arm is configured: per-phase authorization (connect / publish / subscribe) against declared auth providers, plus channel namespaces with inline APPSYNC_JS handlers or DIRECT data-source integrations
- **Data sources** — one per `datasources` entry: DynamoDB, Lambda, HTTP(S), OpenSearch, EventBridge, Aurora Data API, Bedrock runtime, or NONE (local logic), each with its AppSync-trusting service role
- **Resolvers and functions** — GraphQL arm only: UNIT resolvers over one data source or PIPELINE resolvers chaining named functions, in APPSYNC_JS or legacy VTL
- **Server-side cache** — created only when `graphql.cache` is set: a managed Redis-backed instance billed per hour while it exists
- **API keys** — one per `apiKeys` entry, for the API_KEY auth mode, with expiry (AWS's maximum is 365 days)
- **Custom domain** — created only when `customDomain` is set: the domain, its us-east-1 ACM certificate binding, and the association to this API

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AppSync permissions, `iam:PassRole` for data-source service roles, and WAF permissions when attaching a web ACL. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Backends for real data sources** — the table, function, bus, domain, or cluster the API resolves against, plus an IAM role trusting `appsync.amazonaws.com` that can reach each one. NONE and unsigned HTTP data sources need neither.
- **For a custom domain** — an ACM certificate in us-east-1 covering the domain, regardless of the API's own region (AppSync's CloudFront-class contract).
- **For Aurora data sources** — the cluster's Data API enabled and its credentials in a Secrets Manager secret.

## Deploy

### Console

Open the deployment store, find **AWS AppSync API**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the API: the GraphQL or Events arm, data sources, and keys or domains as needed. Start from the **GraphQL API over DynamoDB** preset in the [Presets](#presets) tab for the serverless CRUD backbone, or the **Real-time Events API with channel namespaces** preset for WebSocket pub/sub.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppSyncApi
metadata:
  name: orders-graphql-api
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  graphql:
    apiName: orders_api
    auth:
      type: AMAZON_COGNITO_USER_POOLS
      userPool:
        userPoolId:
          valueFrom:
            kind: AwsCognitoUserPool
            name: customer-users
            fieldPath: status.outputs.user_pool_id
        defaultAction: ALLOW
    schema: |
      type Order {
        id: ID!
        total: Float
      }
      type Query {
        getOrder(id: ID!): Order
      }
      schema {
        query: Query
      }
    resolvers:
      - type: Query
        field: getOrder
        dataSourceName: orders_table
        code: |
          import { util } from "@aws-appsync/utils";
          export function request(ctx) {
            return { operation: "GetItem", key: util.dynamodb.toMapValues({ id: ctx.args.id }) };
          }
          export function response(ctx) {
            return ctx.result;
          }
  datasources:
    - name: orders_table
      type: AMAZON_DYNAMODB
      serviceRoleArn:
        valueFrom:
          kind: AwsIamRole
          name: appsync-orders-table-role
          fieldPath: status.outputs.role_arn
      dynamodb:
        tableName:
          valueFrom:
            kind: AwsDynamodb
            name: orders
            fieldPath: status.outputs.table_name
```

```shell
planton apply -f appsync-api.yaml
```

This creates a Cognito-authorized GraphQL API whose `getOrder` resolver reads the referenced DynamoDB table directly — no Lambda in the middle. An Infra Job tracks the provisioning in real time.

### InfraChart

When the API deploys alongside its user pool, table, and data-source role in one chart, the same reference fields wire via ValueFromRef:

```yaml
spec:
  region: us-west-2
  graphql:
    apiName: orders_api
    auth:
      type: AMAZON_COGNITO_USER_POOLS
      userPool:
        userPoolId:
          valueFrom:
            kind: AwsCognitoUserPool
            name: customer-users
            fieldPath: status.outputs.user_pool_id
        defaultAction: ALLOW
  datasources:
    - name: orders_table
      type: AMAZON_DYNAMODB
      serviceRoleArn:
        valueFrom:
          kind: AwsIamRole
          name: appsync-orders-table-role
          fieldPath: status.outputs.role_arn
      dynamodb:
        tableName:
          valueFrom:
            kind: AwsDynamodb
            name: orders
            fieldPath: status.outputs.table_name
```

The InfraPipeline resolves the dependency graph, deploys the pool, table, and role first, then creates the API wired across them.

## Key Configuration

These are the most important decisions when configuring an AppSync API. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pick the arm by the client conversation** — GraphQL is request/response with subscriptions bolted on; Events is pub/sub first. If clients ask questions about data, use the `graphql` arm; if clients broadcast and listen on channels (chat, presence, live dashboards), the `events` arm is simpler and has no schema to maintain. Migrating between them is a replacement — they are different AWS objects.

**The schema has one writer: the manifest** — AppSync applies the schema asynchronously and the provider performs no drift detection on it, so an out-of-band console edit stays invisible until the next in-band change. Schema errors surface at apply, not plan — a failed apply names the offending SDL line. Resolver, type, and schema mutations serialize behind a per-API lock: a manifest changing thirty resolvers applies them one at a time — slow, not stuck.

**GraphQL names have no hyphens anywhere** — `apiName`, data source names, function names, and type/field names all follow GraphQL's `[A-Za-z_][0-9A-Za-z_]*` charset; that is also what keeps the provider's hyphen-joined import IDs unambiguous. The Events arm differs: its API name comes from `metadata.name`, and channel namespace names allow interior hyphens.

**API keys show their secret once** — AWS returns a key's secret only at creation; every read after that returns the key ID (the `api_key_ids` output carries IDs, never secrets). Fetch the secret from the AWS console or CLI at creation, store it in your secret manager, and rotate by overlap: add a key, roll clients, delete the old. Maximum key life is 365 days.

**The cache is a billed instance, not a flag** — `graphql.cache` provisions a Redis-backed instance billed per hour while it exists; SMALL is the sensible starting size. Both encryption flags are decided at creation — changing either replaces the cache (a cold cache, not an outage) — and PER_RESOLVER_CACHING caches only resolvers whose `caching` block opts in.

**Visibility is a one-way door** — `graphql.visibility` (GLOBAL or PRIVATE-behind-VPC-endpoints) is fixed for life; changing it replaces the API and its endpoint URLs. So is the MERGED flag, the custom domain's name, and each resolver's type.field position.

**EventBridge data sources: replace, don't edit** — At this provider pin, updating an EventBridge data source in place silently drops its bus configuration (an upstream defect). Rename the entry to force recreation instead of editing it. Other data source types update in place.

**MERGED APIs own nothing** — A merged API serves its source APIs' schemas; defining schema, types, functions, or resolvers on the merged surface is rejected at validation before AWS rejects it. AUTO_MERGE propagates source changes automatically; MANUAL_MERGE waits for you. The execution role needs `appsync:SourceGraphQL` plus start-merge permissions on the sources.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsCognitoUserPool** | `graphql.auth.userPool.userPoolId` | `status.outputs.user_pool_id` |
| **AwsIamRole** | `datasources[].serviceRoleArn` | `status.outputs.role_arn` |
| **AwsDynamodb** | `datasources[].dynamodb.tableName` | `status.outputs.table_name` |
| **AwsLambda** | `datasources[].lambda.functionArn` | `status.outputs.function_arn` |
| **AwsEventBridgeBus** | `datasources[].eventbridge.eventBusArn` | `status.outputs.bus_arn` |
| **AwsOpenSearchDomain** | `datasources[].opensearch.endpoint` | `status.outputs.endpoint` |
| **AwsRdsCluster** | `datasources[].relationalDatabase.dbClusterIdentifier` | `status.outputs.cluster_identifier` |
| **AwsWafWebAcl** | `graphql.webAclArn` | `status.outputs.web_acl_arn` |
| **AwsCertManagerCert** | `customDomain.certificateArn` | `status.outputs.cert_arn` |
| **AwsAppSyncApi** | `graphql.merged.sourceApis[].sourceApiId` | `status.outputs.api_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `api_id` | The API's id | The join key MERGED APIs use to reference their sources |
| `api_arn` | The API's ARN | IAM policies scoping AppSync permissions |
| `graphql_url` | The GraphQL endpoint clients query (GraphQL arm) | Application configuration |
| `realtime_url` | The subscriptions endpoint (GraphQL arm) | Application configuration for WebSocket subscriptions |
| `events_http_endpoint` | The publish endpoint (Events arm) | Backend services publishing events |
| `events_realtime_endpoint` | The WebSocket subscribe endpoint (Events arm) | Browser and mobile clients |
| `appsync_domain_name` | The AppSync-managed domain to point DNS at (custom domain) | The CNAME target or Route53 alias for your domain record |
| `domain_hosted_zone_id` | The hosted zone id for Route53 alias records (custom domain) | Route53 alias record configuration |

The map outputs (`datasource_arns`, `function_ids`, `api_key_ids`, `channel_namespace_arns`, `source_api_association_ids`, `type_formats`) exist for import-ID derivation and audit rather than composition — and `api_key_ids` carries key IDs, never key secrets.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**GraphQL over DynamoDB, zero hops** — APPSYNC_JS resolvers call GetItem/PutItem straight from the API layer: latency is the table's, and there is no function cold start to pay. Cognito signs the callers; the table, pool, and role all wire by reference. Start from the **GraphQL API over DynamoDB** preset.

**Real-time channels for chat and presence** — An Events API where anyone with the key connects and subscribes, publishing defaults to IAM so only your backend writes, and per-namespace overrides let clients broadcast their own presence. Inline handlers validate events without a Lambda. Start from the **Real-time Events API with channel namespaces** preset.

**Public fields on a private API** — Keep Cognito as the primary provider and add API_KEY as an additional one; schema directives mark the read-only fields the key may reach. One API serves both audiences instead of two APIs sharing resolvers.

**Federation with a MERGED API** — Teams own source GraphQL APIs; a merged API serves the combined schema under one endpoint. AUTO_MERGE keeps it current automatically, at the cost of source changes propagating without a human gate; MANUAL_MERGE is the reviewed alternative.

## Works With

- [**AWS DynamoDB**](/infra-catalog/aws-dynamodb) — the zero-hop resolver backend via `datasources[].dynamodb`
- [**AWS Lambda**](/infra-catalog/aws-lambda) — data source, Lambda authorizer, and sync-conflict handler
- [**AWS Cognito User Pool**](/infra-catalog/aws-cognito-user-pool) — user-pool authorization for either arm
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the appsync-trusting service roles data sources assume
- [**AWS EventBridge Bus**](/infra-catalog/aws-event-bridge-bus) — data source for publishing events from resolvers and channel handlers
- [**AWS OpenSearch Domain**](/infra-catalog/aws-open-search-domain) — search-backed resolvers via `datasources[].opensearch`
- [**AWS RDS Cluster**](/infra-catalog/aws-rds-cluster) — Aurora Data API resolvers via `datasources[].relationalDatabase`
- [**AWS WAF Web ACL**](/infra-catalog/aws-waf-web-acl) — request filtering on GraphQL APIs via `graphql.webAclArn`
- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) — the us-east-1 certificate behind a custom domain
