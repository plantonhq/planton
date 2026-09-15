# AwsAppSyncApi

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsAppSyncApiSpec defines one AWS AppSync API - AWS's managed API
service - with everything that belongs to the API managed in-line:
data sources, resolvers, functions, schema types, caching, API keys,
channel namespaces, and a custom domain.

AppSync ships two API models, and this spec pivots on exactly one:

GRAPHQL (the graphql arm): a GraphQL API over your data sources -
you provide an SDL schema and resolvers that map fields to data
source operations. Also covers MERGED APIs (a federation surface
that merges other GraphQL APIs' schemas).

EVENTS (the events arm): serverless real-time pub/sub over
WebSockets - clients publish and subscribe on channels grouped into
namespaces; no schema and no resolvers.

Data sources, API keys, and the custom domain sit at the top level
because AWS attaches them to either API model. Resolvers, functions,
types, and caching are GraphQL concepts and live inside the graphql
arm; channel namespaces are the Events concept and live inside the
events arm.

## Example

```yaml
# Canonical AwsAppSyncApi example (hack/dev manifest and refgen Example
# source): a GraphQL API with API_KEY primary auth plus an AWS_IAM
# additional provider, an HTTP data source and a local NONE data
# source, a pipeline function, a pipeline resolver and a unit
# resolver (both APPSYNC_JS), and one API key.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAppSyncApi
metadata:
  name: orders-api
  id: orders-api
  org: test-org
  env: dev
spec:
  region: us-west-2
  graphql:
    apiName: orders_api
    auth:
      type: API_KEY
    additionalAuthProviders:
      - type: AWS_IAM
    schema: |
      type Order {
        id: ID!
        total: Float
      }
      type Query {
        getOrder(id: ID!): Order
      }
      type Mutation {
        putOrder(id: ID!, total: Float): Order
      }
      schema {
        query: Query
        mutation: Mutation
      }
    queryDepthLimit: "10"
    xrayEnabled: true
    functions:
      - name: fetch_order
        dataSourceName: orders_backend
        code: |
          export function request(ctx) {
            return { method: "GET", resourcePath: `/orders/${ctx.args.id}` };
          }
          export function response(ctx) {
            return JSON.parse(ctx.result.body);
          }
    resolvers:
      - type: Query
        field: getOrder
        pipelineFunctions:
          - fetch_order
        code: |
          export function request(ctx) { return {}; }
          export function response(ctx) { return ctx.prev.result; }
      - type: Mutation
        field: putOrder
        dataSourceName: local_passthrough
        code: |
          export function request(ctx) {
            return { payload: ctx.args };
          }
          export function response(ctx) {
            return ctx.result;
          }
  datasources:
    - name: orders_backend
      type: HTTP
      description: The orders service's HTTPS endpoint
      http:
        endpoint: https://orders.internal.example.com
    - name: local_passthrough
      type: NONE
      description: Local resolver logic with no backend
  apiKeys:
    - name: web_client
      description: The web client's key
      expires: "2027-02-01T00:00:00Z"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.graphql` | `AwsAppSyncGraphqlApi` |  |  |  |
| `spec.graphql.apiName` | `string` | yes |  |  |
| `spec.graphql.auth` | `AwsAppSyncGraphqlAuthProvider` | yes |  |  |
| `spec.graphql.auth.type` | `string` |  |  |  |
| `spec.graphql.auth.userPool` | `AwsAppSyncCognitoUserPoolAuth` |  |  |  |
| `spec.graphql.auth.userPool.userPoolId` | `string \| valueFrom` |  |  | AwsCognitoUserPool (`status.outputs.user_pool_id`) |
| `spec.graphql.auth.userPool.appIdClientRegex` | `string` |  |  |  |
| `spec.graphql.auth.userPool.awsRegion` | `string` |  |  |  |
| `spec.graphql.auth.userPool.defaultAction` | `string` |  |  |  |
| `spec.graphql.auth.openidConnect` | `AwsAppSyncOpenidConnectAuth` |  |  |  |
| `spec.graphql.auth.openidConnect.issuer` | `string` | yes |  |  |
| `spec.graphql.auth.openidConnect.clientId` | `string` |  |  |  |
| `spec.graphql.auth.openidConnect.iatTtl` | `int64` |  |  |  |
| `spec.graphql.auth.openidConnect.authTtl` | `int64` |  |  |  |
| `spec.graphql.auth.lambda` | `AwsAppSyncLambdaAuth` |  |  |  |
| `spec.graphql.auth.lambda.authorizerUri` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.graphql.auth.lambda.authorizerResultTtlInSeconds` | `int64` |  |  |  |
| `spec.graphql.auth.lambda.identityValidationExpression` | `string` |  |  |  |
| `spec.graphql.additionalAuthProviders` | `[]AwsAppSyncGraphqlAuthProvider` |  |  |  |
| `spec.graphql.additionalAuthProviders[].type` | `string` |  |  |  |
| `spec.graphql.additionalAuthProviders[].userPool` | `AwsAppSyncCognitoUserPoolAuth` |  |  |  |
| `spec.graphql.additionalAuthProviders[].userPool.userPoolId` | `string \| valueFrom` |  |  | AwsCognitoUserPool (`status.outputs.user_pool_id`) |
| `spec.graphql.additionalAuthProviders[].userPool.appIdClientRegex` | `string` |  |  |  |
| `spec.graphql.additionalAuthProviders[].userPool.awsRegion` | `string` |  |  |  |
| `spec.graphql.additionalAuthProviders[].userPool.defaultAction` | `string` |  |  |  |
| `spec.graphql.additionalAuthProviders[].openidConnect` | `AwsAppSyncOpenidConnectAuth` |  |  |  |
| `spec.graphql.additionalAuthProviders[].openidConnect.issuer` | `string` | yes |  |  |
| `spec.graphql.additionalAuthProviders[].openidConnect.clientId` | `string` |  |  |  |
| `spec.graphql.additionalAuthProviders[].openidConnect.iatTtl` | `int64` |  |  |  |
| `spec.graphql.additionalAuthProviders[].openidConnect.authTtl` | `int64` |  |  |  |
| `spec.graphql.additionalAuthProviders[].lambda` | `AwsAppSyncLambdaAuth` |  |  |  |
| `spec.graphql.additionalAuthProviders[].lambda.authorizerUri` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.graphql.additionalAuthProviders[].lambda.authorizerResultTtlInSeconds` | `int64` |  |  |  |
| `spec.graphql.additionalAuthProviders[].lambda.identityValidationExpression` | `string` |  |  |  |
| `spec.graphql.schema` | `string` |  |  |  |
| `spec.graphql.visibility` | `string` |  |  |  |
| `spec.graphql.disableIntrospection` | `bool` |  |  |  |
| `spec.graphql.queryDepthLimit` | `int64` |  |  |  |
| `spec.graphql.resolverCountLimit` | `int64` |  |  |  |
| `spec.graphql.xrayEnabled` | `bool` |  |  |  |
| `spec.graphql.logConfig` | `AwsAppSyncGraphqlLogConfig` |  |  |  |
| `spec.graphql.logConfig.cloudwatchLogsRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.graphql.logConfig.fieldLogLevel` | `string` |  |  |  |
| `spec.graphql.logConfig.excludeVerboseContent` | `bool` |  |  |  |
| `spec.graphql.enhancedMetrics` | `AwsAppSyncGraphqlEnhancedMetrics` |  |  |  |
| `spec.graphql.enhancedMetrics.dataSourceLevelMetricsBehavior` | `string` |  |  |  |
| `spec.graphql.enhancedMetrics.operationLevelMetricsConfig` | `string` |  |  |  |
| `spec.graphql.enhancedMetrics.resolverLevelMetricsBehavior` | `string` |  |  |  |
| `spec.graphql.webAclArn` | `string \| valueFrom` |  |  | AwsWafWebAcl (`status.outputs.web_acl_arn`) |
| `spec.graphql.cache` | `AwsAppSyncGraphqlCache` |  |  |  |
| `spec.graphql.cache.apiCachingBehavior` | `string` |  |  |  |
| `spec.graphql.cache.ttl` | `int64` |  |  |  |
| `spec.graphql.cache.type` | `string` |  |  |  |
| `spec.graphql.cache.atRestEncryptionEnabled` | `bool` |  |  |  |
| `spec.graphql.cache.transitEncryptionEnabled` | `bool` |  |  |  |
| `spec.graphql.types` | `[]AwsAppSyncGraphqlType` |  |  |  |
| `spec.graphql.types[].name` | `string` | yes |  |  |
| `spec.graphql.types[].definition` | `string` | yes |  |  |
| `spec.graphql.types[].format` | `string` |  |  |  |
| `spec.graphql.functions` | `[]AwsAppSyncGraphqlFunction` |  |  |  |
| `spec.graphql.functions[].name` | `string` | yes |  |  |
| `spec.graphql.functions[].dataSourceName` | `string` | yes |  |  |
| `spec.graphql.functions[].description` | `string` |  |  |  |
| `spec.graphql.functions[].code` | `string` |  |  |  |
| `spec.graphql.functions[].runtimeVersion` | `string` |  |  |  |
| `spec.graphql.functions[].requestMappingTemplate` | `string` |  |  |  |
| `spec.graphql.functions[].responseMappingTemplate` | `string` |  |  |  |
| `spec.graphql.functions[].maxBatchSize` | `int64` |  |  |  |
| `spec.graphql.functions[].syncConfig` | `AwsAppSyncSyncConfig` |  |  |  |
| `spec.graphql.functions[].syncConfig.conflictDetection` | `string` |  |  |  |
| `spec.graphql.functions[].syncConfig.conflictHandler` | `string` |  |  |  |
| `spec.graphql.functions[].syncConfig.lambdaConflictHandlerArn` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.graphql.resolvers` | `[]AwsAppSyncGraphqlResolver` |  |  |  |
| `spec.graphql.resolvers[].type` | `string` | yes |  |  |
| `spec.graphql.resolvers[].field` | `string` | yes |  |  |
| `spec.graphql.resolvers[].dataSourceName` | `string` |  |  |  |
| `spec.graphql.resolvers[].pipelineFunctions` | `[]string` |  |  |  |
| `spec.graphql.resolvers[].code` | `string` |  |  |  |
| `spec.graphql.resolvers[].runtimeVersion` | `string` |  |  |  |
| `spec.graphql.resolvers[].requestTemplate` | `string` |  |  |  |
| `spec.graphql.resolvers[].responseTemplate` | `string` |  |  |  |
| `spec.graphql.resolvers[].maxBatchSize` | `int64` |  |  |  |
| `spec.graphql.resolvers[].caching` | `AwsAppSyncResolverCaching` |  |  |  |
| `spec.graphql.resolvers[].caching.cachingKeys` | `[]string` |  |  |  |
| `spec.graphql.resolvers[].caching.ttl` | `int64` |  |  |  |
| `spec.graphql.resolvers[].syncConfig` | `AwsAppSyncSyncConfig` |  |  |  |
| `spec.graphql.resolvers[].syncConfig.conflictDetection` | `string` |  |  |  |
| `spec.graphql.resolvers[].syncConfig.conflictHandler` | `string` |  |  |  |
| `spec.graphql.resolvers[].syncConfig.lambdaConflictHandlerArn` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.graphql.merged` | `AwsAppSyncGraphqlMerged` |  |  |  |
| `spec.graphql.merged.executionRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.graphql.merged.sourceApis` | `[]AwsAppSyncSourceApi` |  |  |  |
| `spec.graphql.merged.sourceApis[].name` | `string` | yes |  |  |
| `spec.graphql.merged.sourceApis[].sourceApiId` | `string \| valueFrom` | yes |  | AwsAppSyncApi (`status.outputs.api_id`) |
| `spec.graphql.merged.sourceApis[].mergeType` | `string` |  |  |  |
| `spec.graphql.merged.sourceApis[].description` | `string` |  |  |  |
| `spec.events` | `AwsAppSyncEventsApi` |  |  |  |
| `spec.events.ownerContact` | `string` |  |  |  |
| `spec.events.authProviders` | `[]AwsAppSyncEventsAuthProvider` | yes |  |  |
| `spec.events.authProviders[].type` | `string` |  |  |  |
| `spec.events.authProviders[].cognito` | `AwsAppSyncEventsCognitoAuth` |  |  |  |
| `spec.events.authProviders[].cognito.userPoolId` | `string \| valueFrom` |  |  | AwsCognitoUserPool (`status.outputs.user_pool_id`) |
| `spec.events.authProviders[].cognito.awsRegion` | `string` | yes |  |  |
| `spec.events.authProviders[].cognito.appIdClientRegex` | `string` |  |  |  |
| `spec.events.authProviders[].openidConnect` | `AwsAppSyncOpenidConnectAuth` |  |  |  |
| `spec.events.authProviders[].openidConnect.issuer` | `string` | yes |  |  |
| `spec.events.authProviders[].openidConnect.clientId` | `string` |  |  |  |
| `spec.events.authProviders[].openidConnect.iatTtl` | `int64` |  |  |  |
| `spec.events.authProviders[].openidConnect.authTtl` | `int64` |  |  |  |
| `spec.events.authProviders[].lambda` | `AwsAppSyncLambdaAuth` |  |  |  |
| `spec.events.authProviders[].lambda.authorizerUri` | `string \| valueFrom` |  |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.events.authProviders[].lambda.authorizerResultTtlInSeconds` | `int64` |  |  |  |
| `spec.events.authProviders[].lambda.identityValidationExpression` | `string` |  |  |  |
| `spec.events.connectionAuthModes` | `[]string` | yes |  |  |
| `spec.events.defaultPublishAuthModes` | `[]string` | yes |  |  |
| `spec.events.defaultSubscribeAuthModes` | `[]string` | yes |  |  |
| `spec.events.logConfig` | `AwsAppSyncEventsLogConfig` |  |  |  |
| `spec.events.logConfig.cloudwatchLogsRoleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.events.logConfig.logLevel` | `string` |  |  |  |
| `spec.events.channelNamespaces` | `[]AwsAppSyncChannelNamespace` |  |  |  |
| `spec.events.channelNamespaces[].name` | `string` | yes |  |  |
| `spec.events.channelNamespaces[].codeHandlers` | `string` |  |  |  |
| `spec.events.channelNamespaces[].publishAuthModes` | `[]string` |  |  |  |
| `spec.events.channelNamespaces[].subscribeAuthModes` | `[]string` |  |  |  |
| `spec.events.channelNamespaces[].handlerConfigs` | `AwsAppSyncChannelHandlerConfigs` |  |  |  |
| `spec.events.channelNamespaces[].handlerConfigs.onPublish` | `AwsAppSyncChannelHandler` |  |  |  |
| `spec.events.channelNamespaces[].handlerConfigs.onPublish.behavior` | `string` |  |  |  |
| `spec.events.channelNamespaces[].handlerConfigs.onPublish.dataSourceName` | `string` | yes |  |  |
| `spec.events.channelNamespaces[].handlerConfigs.onPublish.lambdaInvokeType` | `string` |  |  |  |
| `spec.events.channelNamespaces[].handlerConfigs.onSubscribe` | `AwsAppSyncChannelHandler` |  |  |  |
| `spec.events.channelNamespaces[].handlerConfigs.onSubscribe.behavior` | `string` |  |  |  |
| `spec.events.channelNamespaces[].handlerConfigs.onSubscribe.dataSourceName` | `string` | yes |  |  |
| `spec.events.channelNamespaces[].handlerConfigs.onSubscribe.lambdaInvokeType` | `string` |  |  |  |
| `spec.datasources` | `[]AwsAppSyncDatasource` |  |  |  |
| `spec.datasources[].name` | `string` | yes |  |  |
| `spec.datasources[].description` | `string` |  |  |  |
| `spec.datasources[].type` | `string` |  |  |  |
| `spec.datasources[].serviceRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.datasources[].dynamodb` | `AwsAppSyncDatasourceDynamodb` |  |  |  |
| `spec.datasources[].dynamodb.tableName` | `string \| valueFrom` | yes |  | AwsDynamodb (`status.outputs.table_name`) |
| `spec.datasources[].dynamodb.region` | `string` |  |  |  |
| `spec.datasources[].dynamodb.useCallerCredentials` | `bool` |  |  |  |
| `spec.datasources[].dynamodb.versioned` | `bool` |  |  |  |
| `spec.datasources[].dynamodb.deltaSync` | `AwsAppSyncDatasourceDeltaSync` |  |  |  |
| `spec.datasources[].dynamodb.deltaSync.deltaSyncTableName` | `string` | yes |  |  |
| `spec.datasources[].dynamodb.deltaSync.baseTableTtl` | `int64` |  |  |  |
| `spec.datasources[].dynamodb.deltaSync.deltaSyncTableTtl` | `int64` |  |  |  |
| `spec.datasources[].lambda` | `AwsAppSyncDatasourceLambda` |  |  |  |
| `spec.datasources[].lambda.functionArn` | `string \| valueFrom` | yes |  | AwsLambda (`status.outputs.function_arn`) |
| `spec.datasources[].http` | `AwsAppSyncDatasourceHttp` |  |  |  |
| `spec.datasources[].http.endpoint` | `string` | yes |  |  |
| `spec.datasources[].http.sigv4` | `AwsAppSyncDatasourceHttpSigning` |  |  |  |
| `spec.datasources[].http.sigv4.signingRegion` | `string` |  |  |  |
| `spec.datasources[].http.sigv4.signingServiceName` | `string` |  |  |  |
| `spec.datasources[].opensearch` | `AwsAppSyncDatasourceOpensearch` |  |  |  |
| `spec.datasources[].opensearch.endpoint` | `string \| valueFrom` | yes |  | AwsOpenSearchDomain (`status.outputs.endpoint`) |
| `spec.datasources[].opensearch.region` | `string` |  |  |  |
| `spec.datasources[].elasticsearch` | `AwsAppSyncDatasourceElasticsearch` |  |  |  |
| `spec.datasources[].elasticsearch.endpoint` | `string` | yes |  |  |
| `spec.datasources[].elasticsearch.region` | `string` |  |  |  |
| `spec.datasources[].eventbridge` | `AwsAppSyncDatasourceEventbridge` |  |  |  |
| `spec.datasources[].eventbridge.eventBusArn` | `string \| valueFrom` | yes |  | AwsEventBridgeBus (`status.outputs.bus_arn`) |
| `spec.datasources[].relationalDatabase` | `AwsAppSyncDatasourceRelationalDatabase` |  |  |  |
| `spec.datasources[].relationalDatabase.dbClusterIdentifier` | `string \| valueFrom` | yes |  | AwsRdsCluster (`status.outputs.cluster_identifier`) |
| `spec.datasources[].relationalDatabase.awsSecretStoreArn` | `string \| valueFrom` | yes |  | AwsSecretsManagerSecret (`status.outputs.secret_arn`) |
| `spec.datasources[].relationalDatabase.databaseName` | `string` |  |  |  |
| `spec.datasources[].relationalDatabase.schema` | `string` |  |  |  |
| `spec.datasources[].relationalDatabase.region` | `string` |  |  |  |
| `spec.apiKeys` | `[]AwsAppSyncApiKey` |  |  |  |
| `spec.apiKeys[].name` | `string` | yes |  |  |
| `spec.apiKeys[].description` | `string` |  |  |  |
| `spec.apiKeys[].expires` | `string` |  |  |  |
| `spec.customDomain` | `AwsAppSyncCustomDomain` |  |  |  |
| `spec.customDomain.domainName` | `string` | yes |  |  |
| `spec.customDomain.certificateArn` | `string \| valueFrom` | yes |  | AwsCertManagerCert (`status.outputs.cert_arn`) |
| `spec.customDomain.description` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the API lives in. Example: "us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.graphql

`AwsAppSyncGraphqlApi`

The GraphQL API model: an SDL schema, resolvers over data
sources, optional caching, and the MERGED federation variant.

- rule: the auth provider's config must match its type - user_pool with AMAZON_COGNITO_USER_POOLS, openid_connect with OPENID_CONNECT, lambda with AWS_LAMBDA, and none of the three with API_KEY or AWS_IAM
- rule: the primary auth provider's user_pool requires default_action (ALLOW or DENY)
- rule: each additional auth provider's config must match its type - user_pool with AMAZON_COGNITO_USER_POOLS, openid_connect with OPENID_CONNECT, lambda with AWS_LAMBDA, and none of the three with API_KEY or AWS_IAM
- rule: default_action applies only to the primary auth provider - remove it from additional providers' user_pool blocks
- rule: each authorization type may appear at most once across the primary and additional auth providers
- rule: a MERGED API merges its source APIs' schemas - remove schema, types, functions, resolvers, and cache (they belong on the source APIs)
- rule: function names must be unique within the API
- rule: resolvers must be unique per type.field position
- rule: type names must be unique within the API

### spec.graphql.apiName

`string` · required

The API's name in AWS - an explicit field rather than
metadata.name because GraphQL API names allow only letters,
digits, and underscores (NO hyphens), starting with a letter or
underscore.

- rule: {"string":{"minLen":"1","pattern":"^[A-Za-z_][0-9A-Za-z_]*$"}}

### spec.graphql.auth

`AwsAppSyncGraphqlAuthProvider` · required

The primary authorization provider - every GraphQL API has
exactly one, plus optional additionals for multi-auth schemas
(fields pick providers with @aws_auth-style directives).

- rule: {"required":true}

### spec.graphql.auth.type

`string`

How callers authenticate: API_KEY (keys from api_keys), AWS_IAM
(SigV4), AMAZON_COGNITO_USER_POOLS (user_pool), OPENID_CONNECT
(openid_connect), or AWS_LAMBDA (lambda authorizer).

- rule: {"string":{"in":["API_KEY","AWS_IAM","AMAZON_COGNITO_USER_POOLS","OPENID_CONNECT","AWS_LAMBDA"]}}

### spec.graphql.auth.userPool

`AwsAppSyncCognitoUserPoolAuth`

Cognito user pool authorization (type AMAZON_COGNITO_USER_POOLS).

### spec.graphql.auth.userPool.userPoolId

`string | valueFrom`

The user pool. Can reference an AwsCognitoUserPool resource or
pass a literal pool id.

Containment-exempt: the API authorizes callers AGAINST the pool; it is
not deployed into it. On a diagram the pool is a room its clients and
identity providers live in, and the API stands outside with a line in.

- references: AwsCognitoUserPool (`status.outputs.user_pool_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_id}} -- a bare string does not parse

### spec.graphql.auth.userPool.appIdClientRegex

`string`

A regex the caller's app client id must match. Unset means any
client in the pool.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.auth.userPool.awsRegion

`string`

The user pool's region when it differs from the API's region.
Unset means the API's own region.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.auth.userPool.defaultAction

`string`

What happens to requests that match no OTHER auth provider: ALLOW
passes them to this pool, DENY rejects them. REQUIRED on the
primary auth provider; forbidden on additionals (AWS's
asymmetry, enforced by CEL on the arm).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ALLOW","DENY"]}}

### spec.graphql.auth.openidConnect

`AwsAppSyncOpenidConnectAuth`

OpenID Connect authorization (type OPENID_CONNECT).

### spec.graphql.auth.openidConnect.issuer

`string` · required

The OIDC issuer URL the tokens must come from (matches the
token's iss claim).

- rule: {"string":{"minLen":"1"}}

### spec.graphql.auth.openidConnect.clientId

`string`

The client id the token's aud claim must match. Unset means any.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.auth.openidConnect.iatTtl

`int64`

Milliseconds a token stays valid after its iat claim. Unset (0)
defers to the token's own exp.

### spec.graphql.auth.openidConnect.authTtl

`int64`

Milliseconds a token stays valid after authentication. Unset (0)
defers to the token's own exp.

### spec.graphql.auth.lambda

`AwsAppSyncLambdaAuth`

Lambda authorizer (type AWS_LAMBDA).

### spec.graphql.auth.lambda.authorizerUri

`string | valueFrom`

The authorizer function. AppSync must be allowed to invoke it
(lambda:InvokeFunction from appsync.amazonaws.com). Can reference
an AwsLambda resource or pass a literal function/alias ARN.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.graphql.auth.lambda.authorizerResultTtlInSeconds

`int64`

Seconds AppSync caches an authorizer response, 0-3600. Unset (0)
means no caching... except AWS's schema default is 300 when the
field is omitted entirely; the modules send the value only when
set, so 0 keeps AWS's default. Set 1-3600 to tune.

- rule: authorizer_result_ttl_in_seconds must be between 0 and 3600

### spec.graphql.auth.lambda.identityValidationExpression

`string`

A regex the token must match before AppSync bothers invoking the
authorizer (a cheap pre-filter).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.additionalAuthProviders

`[]AwsAppSyncGraphqlAuthProvider`

Additional authorization providers for multi-auth APIs.

### spec.graphql.additionalAuthProviders[].type

`string`

How callers authenticate: API_KEY (keys from api_keys), AWS_IAM
(SigV4), AMAZON_COGNITO_USER_POOLS (user_pool), OPENID_CONNECT
(openid_connect), or AWS_LAMBDA (lambda authorizer).

- rule: {"string":{"in":["API_KEY","AWS_IAM","AMAZON_COGNITO_USER_POOLS","OPENID_CONNECT","AWS_LAMBDA"]}}

### spec.graphql.additionalAuthProviders[].userPool

`AwsAppSyncCognitoUserPoolAuth`

Cognito user pool authorization (type AMAZON_COGNITO_USER_POOLS).

### spec.graphql.additionalAuthProviders[].userPool.userPoolId

`string | valueFrom`

The user pool. Can reference an AwsCognitoUserPool resource or
pass a literal pool id.

Containment-exempt: the API authorizes callers AGAINST the pool; it is
not deployed into it. On a diagram the pool is a room its clients and
identity providers live in, and the API stands outside with a line in.

- references: AwsCognitoUserPool (`status.outputs.user_pool_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_id}} -- a bare string does not parse

### spec.graphql.additionalAuthProviders[].userPool.appIdClientRegex

`string`

A regex the caller's app client id must match. Unset means any
client in the pool.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.additionalAuthProviders[].userPool.awsRegion

`string`

The user pool's region when it differs from the API's region.
Unset means the API's own region.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.additionalAuthProviders[].userPool.defaultAction

`string`

What happens to requests that match no OTHER auth provider: ALLOW
passes them to this pool, DENY rejects them. REQUIRED on the
primary auth provider; forbidden on additionals (AWS's
asymmetry, enforced by CEL on the arm).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ALLOW","DENY"]}}

### spec.graphql.additionalAuthProviders[].openidConnect

`AwsAppSyncOpenidConnectAuth`

OpenID Connect authorization (type OPENID_CONNECT).

### spec.graphql.additionalAuthProviders[].openidConnect.issuer

`string` · required

The OIDC issuer URL the tokens must come from (matches the
token's iss claim).

- rule: {"string":{"minLen":"1"}}

### spec.graphql.additionalAuthProviders[].openidConnect.clientId

`string`

The client id the token's aud claim must match. Unset means any.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.additionalAuthProviders[].openidConnect.iatTtl

`int64`

Milliseconds a token stays valid after its iat claim. Unset (0)
defers to the token's own exp.

### spec.graphql.additionalAuthProviders[].openidConnect.authTtl

`int64`

Milliseconds a token stays valid after authentication. Unset (0)
defers to the token's own exp.

### spec.graphql.additionalAuthProviders[].lambda

`AwsAppSyncLambdaAuth`

Lambda authorizer (type AWS_LAMBDA).

### spec.graphql.additionalAuthProviders[].lambda.authorizerUri

`string | valueFrom`

The authorizer function. AppSync must be allowed to invoke it
(lambda:InvokeFunction from appsync.amazonaws.com). Can reference
an AwsLambda resource or pass a literal function/alias ARN.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.graphql.additionalAuthProviders[].lambda.authorizerResultTtlInSeconds

`int64`

Seconds AppSync caches an authorizer response, 0-3600. Unset (0)
means no caching... except AWS's schema default is 300 when the
field is omitted entirely; the modules send the value only when
set, so 0 keeps AWS's default. Set 1-3600 to tune.

- rule: authorizer_result_ttl_in_seconds must be between 0 and 3600

### spec.graphql.additionalAuthProviders[].lambda.identityValidationExpression

`string`

A regex the token must match before AppSync bothers invoking the
authorizer (a cheap pre-filter).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.schema

`string`

The GraphQL schema in SDL. Applied via AppSync's async schema
creation; AWS reports schema errors at apply, not at plan. NOTE:
the provider performs no drift detection on the schema -
out-of-band schema edits are invisible until the next in-band
change (recorded in the import catalog as config-only).

### spec.graphql.visibility

`string`

Who can call the API. GLOBAL (AWS default when unset) serves the
public AppSync endpoint; PRIVATE serves only through VPC
endpoints. Fixed for life - changing it replaces the API.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["GLOBAL","PRIVATE"]}}

### spec.graphql.disableIntrospection

`bool`

Turn OFF schema introspection (__schema queries). AWS's default
is enabled; disable it to keep the schema shape private on
public-facing APIs.

### spec.graphql.queryDepthLimit

`int64`

The maximum nesting depth a query may reach, 1-75. Unset (0)
means AWS's default: unlimited depth.

- rule: query_depth_limit must be between 1 and 75 (0 = unlimited)

### spec.graphql.resolverCountLimit

`int64`

The maximum resolvers one query may invoke, 1-10000. Unset (0)
means AWS's default: 10000.

- rule: resolver_count_limit must be between 1 and 10000 (0 = AWS default)

### spec.graphql.xrayEnabled

`bool`

Trace requests with AWS X-Ray.

### spec.graphql.logConfig

`AwsAppSyncGraphqlLogConfig`

Where AppSync writes request logs. Unset means no logging.

### spec.graphql.logConfig.cloudwatchLogsRoleArn

`string | valueFrom`

The role AppSync assumes to write logs. Can reference an
AwsIamRole resource or pass a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.graphql.logConfig.fieldLogLevel

`string`

How much of each request to log: NONE, ERROR, ALL, INFO, or
DEBUG.

- rule: {"string":{"in":["NONE","ERROR","ALL","INFO","DEBUG"]}}

### spec.graphql.logConfig.excludeVerboseContent

`bool`

Drop request/response bodies and headers from the logs (log
metadata only).

### spec.graphql.enhancedMetrics

`AwsAppSyncGraphqlEnhancedMetrics`

Per-data-source, per-operation, and per-resolver CloudWatch
metrics granularity. Unset means AWS's defaults.

### spec.graphql.enhancedMetrics.dataSourceLevelMetricsBehavior

`string`

Per-data-source metrics: FULL_REQUEST_DATA_SOURCE_METRICS (all
data sources) or PER_DATA_SOURCE_METRICS (only data sources whose
metrics_enabled is set - the provider exposes no per-data-source
flag at this pin, so FULL is the useful value).

- rule: {"string":{"in":["FULL_REQUEST_DATA_SOURCE_METRICS","PER_DATA_SOURCE_METRICS"]}}

### spec.graphql.enhancedMetrics.operationLevelMetricsConfig

`string`

Operation-level metrics: ENABLED or DISABLED.

- rule: {"string":{"in":["ENABLED","DISABLED"]}}

### spec.graphql.enhancedMetrics.resolverLevelMetricsBehavior

`string`

Per-resolver metrics: FULL_REQUEST_RESOLVER_METRICS (all
resolvers) or PER_RESOLVER_METRICS.

- rule: {"string":{"in":["FULL_REQUEST_RESOLVER_METRICS","PER_RESOLVER_METRICS"]}}

### spec.graphql.webAclArn

`string | valueFrom`

The REGIONAL-scope WAFv2 web ACL protecting this GraphQL API, by
ARN - the modules create the web-ACL association alongside the
API (the AwsAlb.spec.web_acl_arn pattern). A GraphQL API has at
most one web ACL; the web ACL must live in the API's region.
(AWS's WAF association supports GraphQL APIs, not Events APIs.)
Can reference an AwsWafWebAcl resource.

- references: AwsWafWebAcl (`status.outputs.web_acl_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsWafWebAcl, name: <that resource's name>, fieldPath: status.outputs.web_acl_arn}} -- a bare string does not parse

### spec.graphql.cache

`AwsAppSyncGraphqlCache`

Server-side response caching - one cache per API (AWS's own
model: the cache's identity IS the API).

### spec.graphql.cache.apiCachingBehavior

`string`

What gets cached: FULL_REQUEST_CACHING (everything),
PER_RESOLVER_CACHING (only resolvers whose caching_config opts
in), or OPERATION_LEVEL_CACHING.

- rule: {"string":{"in":["FULL_REQUEST_CACHING","PER_RESOLVER_CACHING","OPERATION_LEVEL_CACHING"]}}

### spec.graphql.cache.ttl

`int64`

Seconds cached entries live, 1-3600.

- rule: cache ttl must be between 1 and 3600 seconds

### spec.graphql.cache.type

`string`

The cache instance size. SMALL through LARGE_12X are the current
generation; T2_/R4_ names are the legacy generation AWS still
accepts.

- rule: {"string":{"in":["SMALL","MEDIUM","LARGE","XLARGE","LARGE_2X","LARGE_4X","LARGE_8X","LARGE_12X","T2_SMALL","T2_MEDIUM","R4_LARGE","R4_XLARGE","R4_2XLARGE","R4_4XLARGE","R4_8XLARGE"]}}

### spec.graphql.cache.atRestEncryptionEnabled

`bool`

Encrypt cached responses at rest. Decided at cache creation -
changing it REPLACES the cache (a cold cache, not an outage).

### spec.graphql.cache.transitEncryptionEnabled

`bool`

Encrypt cache traffic in transit. Also replace-on-change.

### spec.graphql.types

`[]AwsAppSyncGraphqlType`

Schema types managed individually (outside the schema document),
keyed by their declared name.

### spec.graphql.types[].name

`string` · required

The type's name as declared inside definition - the for_each key
on both engines and part of the import ID. AWS derives the real
name from the definition; keep this field in sync with it (a
mismatch surfaces as a perpetual diff).

- rule: {"string":{"minLen":"1","pattern":"^[A-Za-z_][0-9A-Za-z_]*$"}}

### spec.graphql.types[].definition

`string` · required

The type definition - SDL (format SDL) or introspection JSON
(format JSON). The text is CREATE-TIME: AWS rewrites it
server-side into its own whitespace form (indentation stripped,
braces re-placed, blank lines injected), so both engines ignore
the textual echo to keep plans converging. Editing the definition
body in place therefore does NOT propagate - to change a managed
type's definition, remove the entry, apply, and re-add it with
the new text (the same replace-the-entry choreography the format
field requires).

- rule: {"string":{"minLen":"1"}}

### spec.graphql.types[].format

`string`

The definition's format. Fixed for life at AWS: the provider
ignores format changes on update (a perpetual-diff trap) -
replace the type entry instead.

- rule: {"string":{"in":["SDL","JSON"]}}

### spec.graphql.functions

`[]AwsAppSyncGraphqlFunction`

Pipeline functions, keyed by name - reusable units of resolver
logic bound to a data source. Resolvers reference them by name in
pipeline_functions (the modules join names to AWS function ids).

- rule: a function uses code (JavaScript) or request/response mapping templates (VTL), not both

### spec.graphql.functions[].name

`string` · required

The function's name - the for_each key, the function_ids output
key, and what resolvers reference in pipeline_functions. AWS
allows only letters, digits, and underscores.

- rule: {"string":{"minLen":"1","pattern":"^[A-Za-z_][0-9A-Za-z_]*$"}}

### spec.graphql.functions[].dataSourceName

`string` · required

The data source this function operates on, by spec datasource
name (or the name of a data source created outside this
manifest).

- rule: {"string":{"minLen":"1"}}

### spec.graphql.functions[].description

`string`

What this function does.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.functions[].code

`string`

The function body as APPSYNC_JS JavaScript (the modern form).
When set, the modules pin the runtime to APPSYNC_JS at
runtime_version. 1-32768 characters.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"32768"}}

### spec.graphql.functions[].runtimeVersion

`string`

The APPSYNC_JS runtime version when code is set. Unset means
"1.0.0" (the only version AWS ships today).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+\\.[0-9]+\\.[0-9]+$"}}

### spec.graphql.functions[].requestMappingTemplate

`string`

The VTL request mapping template (the legacy form; prefer code).
When templates are used, the modules pin AWS's required
functionVersion 2018-05-29.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.functions[].responseMappingTemplate

`string`

The VTL response mapping template.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.functions[].maxBatchSize

`int64`

Maximum batch size for batched Lambda operations, 1-2000. Unset
(0) means no batching.

- rule: max_batch_size must be between 1 and 2000 (0 = no batching)

### spec.graphql.functions[].syncConfig

`AwsAppSyncSyncConfig`

Versioned-data-source conflict detection and resolution.

- rule: conflict_handler LAMBDA requires lambda_conflict_handler_arn; other handlers must not set it

### spec.graphql.functions[].syncConfig.conflictDetection

`string`

How conflicts are detected: VERSION (compare item versions) or
NONE.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["VERSION","NONE"]}}

### spec.graphql.functions[].syncConfig.conflictHandler

`string`

How detected conflicts resolve: OPTIMISTIC_CONCURRENCY (reject
stale writes), AUTOMERGE, LAMBDA (your function decides), or
NONE.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["OPTIMISTIC_CONCURRENCY","LAMBDA","AUTOMERGE","NONE"]}}

### spec.graphql.functions[].syncConfig.lambdaConflictHandlerArn

`string | valueFrom`

The conflict-resolution function (handler LAMBDA). Can reference
an AwsLambda resource or pass a literal function ARN.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.graphql.resolvers

`[]AwsAppSyncGraphqlResolver`

Resolvers mapping schema fields to data source operations, keyed
by their type.field position.

- rule: set exactly one of data_source_name (a UNIT resolver) and pipeline_functions (a PIPELINE resolver)
- rule: a resolver uses code (JavaScript) or request/response mapping templates (VTL), not both

### spec.graphql.resolvers[].type

`string` · required

The schema type the resolved field lives on (e.g. "Query",
"Mutation", or an object type). Fixed for life - the type.field
position IS the resolver's identity.

- rule: {"string":{"minLen":"1","pattern":"^[A-Za-z_][0-9A-Za-z_]*$"}}

### spec.graphql.resolvers[].field

`string` · required

The field this resolver resolves. Fixed for life.

- rule: {"string":{"minLen":"1","pattern":"^[A-Za-z_][0-9A-Za-z_]*$"}}

### spec.graphql.resolvers[].dataSourceName

`string`

UNIT form: the data source this resolver calls, by spec
datasource name (or an externally created data source's name).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.resolvers[].pipelineFunctions

`[]string`

PIPELINE form: the functions to run in order, by spec function
name - the modules join names to AWS function ids.

### spec.graphql.resolvers[].code

`string`

The resolver body as APPSYNC_JS JavaScript. Same runtime pinning
as functions. 1-32768 characters.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"32768"}}

### spec.graphql.resolvers[].runtimeVersion

`string`

The APPSYNC_JS runtime version when code is set. Unset means
"1.0.0".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+\\.[0-9]+\\.[0-9]+$"}}

### spec.graphql.resolvers[].requestTemplate

`string`

The VTL request mapping template (legacy form).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.resolvers[].responseTemplate

`string`

The VTL response mapping template.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.graphql.resolvers[].maxBatchSize

`int64`

Maximum batch size for batched Lambda operations, 1-2000.

- rule: max_batch_size must be between 1 and 2000 (0 = no batching)

### spec.graphql.resolvers[].caching

`AwsAppSyncResolverCaching`

Per-resolver caching (the cache's PER_RESOLVER_CACHING mode).

### spec.graphql.resolvers[].caching.cachingKeys

`[]string`

The request values that compose the cache key (e.g.
"$context.arguments.id", "$context.identity.sub"). Empty caches
per full request.

### spec.graphql.resolvers[].caching.ttl

`int64`

Seconds this resolver's entries live, 1-3600.

- rule: caching ttl must be between 1 and 3600 seconds

### spec.graphql.resolvers[].syncConfig

`AwsAppSyncSyncConfig`

Versioned-data-source conflict detection and resolution.

- rule: conflict_handler LAMBDA requires lambda_conflict_handler_arn; other handlers must not set it

### spec.graphql.resolvers[].syncConfig.conflictDetection

`string`

How conflicts are detected: VERSION (compare item versions) or
NONE.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["VERSION","NONE"]}}

### spec.graphql.resolvers[].syncConfig.conflictHandler

`string`

How detected conflicts resolve: OPTIMISTIC_CONCURRENCY (reject
stale writes), AUTOMERGE, LAMBDA (your function decides), or
NONE.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["OPTIMISTIC_CONCURRENCY","LAMBDA","AUTOMERGE","NONE"]}}

### spec.graphql.resolvers[].syncConfig.lambdaConflictHandlerArn

`string | valueFrom`

The conflict-resolution function (handler LAMBDA). Can reference
an AwsLambda resource or pass a literal function ARN.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.graphql.merged

`AwsAppSyncGraphqlMerged`

Makes this a MERGED API - a federation surface serving the merged
schemas of its source APIs. Set at create, for life (changing the
API type replaces the API).

- rule: source API entry names must be unique

### spec.graphql.merged.executionRoleArn

`string | valueFrom` · required

The role AppSync assumes to read source APIs during merges. Can
reference an AwsIamRole resource or pass a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.graphql.merged.sourceApis

`[]AwsAppSyncSourceApi`

The GraphQL APIs whose schemas merge into this one, keyed by a
spec-side name.

### spec.graphql.merged.sourceApis[].name

`string` · required

The entry's key in the for_each and the
source_api_association_ids output map.

- rule: {"string":{"minLen":"1","pattern":"^[A-Za-z_][0-9A-Za-z_]*$"}}

### spec.graphql.merged.sourceApis[].sourceApiId

`string | valueFrom` · required

The source GraphQL API. Can reference another AwsAppSyncApi
resource or pass a literal API id. Fixed for life.

- references: AwsAppSyncApi (`status.outputs.api_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsAppSyncApi, name: <that resource's name>, fieldPath: status.outputs.api_id}} -- a bare string does not parse

### spec.graphql.merged.sourceApis[].mergeType

`string`

How source schema changes reach the merged API: AUTO_MERGE
(propagate automatically) or MANUAL_MERGE (AWS's default when
unset - you trigger merges).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AUTO_MERGE","MANUAL_MERGE"]}}

### spec.graphql.merged.sourceApis[].description

`string`

What this source contributes.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.events

`AwsAppSyncEventsApi`

The Events API model: real-time pub/sub over channel namespaces
with per-phase (connect/publish/subscribe) authorization.

- rule: every connection_auth_modes entry must match a declared auth provider's type
- rule: every default_publish_auth_modes entry must match a declared auth provider's type
- rule: every default_subscribe_auth_modes entry must match a declared auth provider's type
- rule: each auth provider's config must match its type - cognito with AMAZON_COGNITO_USER_POOLS, openid_connect with OPENID_CONNECT, lambda with AWS_LAMBDA, and none of the three with API_KEY or AWS_IAM
- rule: each authorization type may appear at most once in auth_providers
- rule: channel namespace names must be unique within the API

### spec.events.ownerContact

`string`

A contact for the API's owner (surfaces in the AWS console).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.events.authProviders

`[]AwsAppSyncEventsAuthProvider` · required

The auth providers clients may use, by phase (below). At least
one.

- rule: {"repeated":{"minItems":"1"}}

### spec.events.authProviders[].type

`string`

The authorization type this provider serves.

- rule: {"string":{"in":["API_KEY","AWS_IAM","AMAZON_COGNITO_USER_POOLS","OPENID_CONNECT","AWS_LAMBDA"]}}

### spec.events.authProviders[].cognito

`AwsAppSyncEventsCognitoAuth`

Cognito user pool authorization (type AMAZON_COGNITO_USER_POOLS).

### spec.events.authProviders[].cognito.userPoolId

`string | valueFrom`

The user pool. Can reference an AwsCognitoUserPool resource or
pass a literal pool id.

Containment-exempt: the API authorizes callers AGAINST the pool; it is
not deployed into it (the GraphQL arm carries the same verdict).

- references: AwsCognitoUserPool (`status.outputs.user_pool_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_id}} -- a bare string does not parse

### spec.events.authProviders[].cognito.awsRegion

`string` · required

The user pool's region. Required by AWS on Events APIs.

- rule: {"string":{"minLen":"1"}}

### spec.events.authProviders[].cognito.appIdClientRegex

`string`

A regex the caller's app client id must match. Unset means any
client in the pool.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.events.authProviders[].openidConnect

`AwsAppSyncOpenidConnectAuth`

OpenID Connect authorization (type OPENID_CONNECT).

### spec.events.authProviders[].openidConnect.issuer

`string` · required

The OIDC issuer URL the tokens must come from (matches the
token's iss claim).

- rule: {"string":{"minLen":"1"}}

### spec.events.authProviders[].openidConnect.clientId

`string`

The client id the token's aud claim must match. Unset means any.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.events.authProviders[].openidConnect.iatTtl

`int64`

Milliseconds a token stays valid after its iat claim. Unset (0)
defers to the token's own exp.

### spec.events.authProviders[].openidConnect.authTtl

`int64`

Milliseconds a token stays valid after authentication. Unset (0)
defers to the token's own exp.

### spec.events.authProviders[].lambda

`AwsAppSyncLambdaAuth`

Lambda authorizer (type AWS_LAMBDA).

### spec.events.authProviders[].lambda.authorizerUri

`string | valueFrom`

The authorizer function. AppSync must be allowed to invoke it
(lambda:InvokeFunction from appsync.amazonaws.com). Can reference
an AwsLambda resource or pass a literal function/alias ARN.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.events.authProviders[].lambda.authorizerResultTtlInSeconds

`int64`

Seconds AppSync caches an authorizer response, 0-3600. Unset (0)
means no caching... except AWS's schema default is 300 when the
field is omitted entirely; the modules send the value only when
set, so 0 keeps AWS's default. Set 1-3600 to tune.

- rule: authorizer_result_ttl_in_seconds must be between 0 and 3600

### spec.events.authProviders[].lambda.identityValidationExpression

`string`

A regex the token must match before AppSync bothers invoking the
authorizer (a cheap pre-filter).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.events.connectionAuthModes

`[]string` · required

Who may open a WebSocket connection - one or more declared
provider types.

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["API_KEY","AWS_IAM","AMAZON_COGNITO_USER_POOLS","OPENID_CONNECT","AWS_LAMBDA"]}}}}

### spec.events.defaultPublishAuthModes

`[]string` · required

The default publish authorization for namespaces that do not
override it.

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["API_KEY","AWS_IAM","AMAZON_COGNITO_USER_POOLS","OPENID_CONNECT","AWS_LAMBDA"]}}}}

### spec.events.defaultSubscribeAuthModes

`[]string` · required

The default subscribe authorization for namespaces that do not
override it.

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["API_KEY","AWS_IAM","AMAZON_COGNITO_USER_POOLS","OPENID_CONNECT","AWS_LAMBDA"]}}}}

### spec.events.logConfig

`AwsAppSyncEventsLogConfig`

Where AppSync writes event logs. Unset means no logging.

### spec.events.logConfig.cloudwatchLogsRoleArn

`string | valueFrom` · required

The role AppSync assumes to write logs. Can reference an
AwsIamRole resource or pass a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.events.logConfig.logLevel

`string`

How much to log: NONE, ERROR, ALL, INFO, or DEBUG.

- rule: {"string":{"in":["NONE","ERROR","ALL","INFO","DEBUG"]}}

### spec.events.channelNamespaces

`[]AwsAppSyncChannelNamespace`

The channel namespaces clients publish and subscribe in, keyed by
name.

### spec.events.channelNamespaces[].name

`string` · required

The namespace's name - the for_each key and half its import ID.
Letters, digits, and single interior hyphens; 1-50 characters.
Fixed for life.

- rule: {"string":{"minLen":"1","maxLen":"50","pattern":"^[A-Za-z0-9](?:[A-Za-z0-9-]{0,48}[A-Za-z0-9])?$"}}

### spec.events.channelNamespaces[].codeHandlers

`string`

Event handler code (APPSYNC_JS) running on publish/subscribe for
this namespace's channels - the inline alternative to DIRECT
data-source handlers. 1-32768 characters.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"32768"}}

### spec.events.channelNamespaces[].publishAuthModes

`[]string`

Publish auth modes overriding the API's
default_publish_auth_modes for this namespace.

- rule: {"repeated":{"items":{"string":{"in":["API_KEY","AWS_IAM","AMAZON_COGNITO_USER_POOLS","OPENID_CONNECT","AWS_LAMBDA"]}}}}

### spec.events.channelNamespaces[].subscribeAuthModes

`[]string`

Subscribe auth modes overriding the API's
default_subscribe_auth_modes for this namespace.

- rule: {"repeated":{"items":{"string":{"in":["API_KEY","AWS_IAM","AMAZON_COGNITO_USER_POOLS","OPENID_CONNECT","AWS_LAMBDA"]}}}}

### spec.events.channelNamespaces[].handlerConfigs

`AwsAppSyncChannelHandlerConfigs`

Integration handlers routing publish/subscribe events to data
sources.

### spec.events.channelNamespaces[].handlerConfigs.onPublish

`AwsAppSyncChannelHandler`

The handler for publish events.

### spec.events.channelNamespaces[].handlerConfigs.onPublish.behavior

`string`

How the handler runs: CODE (the namespace's code_handlers
JavaScript decides) or DIRECT (events go straight to the
integration's data source).

- rule: {"string":{"in":["CODE","DIRECT"]}}

### spec.events.channelNamespaces[].handlerConfigs.onPublish.dataSourceName

`string` · required

The data source receiving the events, by spec datasource name
(or an externally created data source's name).

- rule: {"string":{"minLen":"1"}}

### spec.events.channelNamespaces[].handlerConfigs.onPublish.lambdaInvokeType

`string`

For Lambda data sources: REQUEST_RESPONSE (synchronous - the
function's return feeds the event flow) or EVENT (asynchronous
fire-and-forget). Unset means AWS's default (REQUEST_RESPONSE).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["REQUEST_RESPONSE","EVENT"]}}

### spec.events.channelNamespaces[].handlerConfigs.onSubscribe

`AwsAppSyncChannelHandler`

The handler for subscribe events.

### spec.events.channelNamespaces[].handlerConfigs.onSubscribe.behavior

`string`

How the handler runs: CODE (the namespace's code_handlers
JavaScript decides) or DIRECT (events go straight to the
integration's data source).

- rule: {"string":{"in":["CODE","DIRECT"]}}

### spec.events.channelNamespaces[].handlerConfigs.onSubscribe.dataSourceName

`string` · required

The data source receiving the events, by spec datasource name
(or an externally created data source's name).

- rule: {"string":{"minLen":"1"}}

### spec.events.channelNamespaces[].handlerConfigs.onSubscribe.lambdaInvokeType

`string`

For Lambda data sources: REQUEST_RESPONSE (synchronous - the
function's return feeds the event flow) or EVENT (asynchronous
fire-and-forget). Unset means AWS's default (REQUEST_RESPONSE).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["REQUEST_RESPONSE","EVENT"]}}

### spec.datasources

`[]AwsAppSyncDatasource`

The API's data sources, keyed by name - what resolvers (GraphQL)
and channel-namespace handlers (Events) call into. A data source
belongs to exactly one API for life.

- rule: the data source's config block must match its type - dynamodb, lambda, http, opensearch, elasticsearch, eventbridge, or relational_database; NONE and AMAZON_BEDROCK_RUNTIME take no config block

### spec.datasources[].name

`string` · required

The data source's name - the for_each key, what resolvers,
functions, and channel handlers reference, and half its import
ID. AWS allows only letters, digits, and underscores.

- rule: {"string":{"minLen":"1","pattern":"^[A-Za-z_][0-9A-Za-z_]*$"}}

### spec.datasources[].description

`string`

What this data source connects to.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.datasources[].type

`string`

What kind of backend this is. NONE runs resolver logic with no
backend (local pub/sub, transformations);
AMAZON_BEDROCK_RUNTIME invokes Bedrock models directly.
AMAZON_ELASTICSEARCH is the legacy Elasticsearch type - prefer
AMAZON_OPENSEARCH_SERVICE for new work.

- rule: {"string":{"in":["AWS_LAMBDA","AMAZON_DYNAMODB","HTTP","AMAZON_OPENSEARCH_SERVICE","AMAZON_ELASTICSEARCH","AMAZON_EVENTBRIDGE","RELATIONAL_DATABASE","AMAZON_BEDROCK_RUNTIME","NONE"]}}

### spec.datasources[].serviceRoleArn

`string | valueFrom`

The role AppSync assumes to reach the backend. Required by AWS
for every type except NONE and HTTP-without-signing. Can
reference an AwsIamRole resource or pass a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.datasources[].dynamodb

`AwsAppSyncDatasourceDynamodb`

DynamoDB backend (type AMAZON_DYNAMODB).

### spec.datasources[].dynamodb.tableName

`string | valueFrom` · required

The table. Can reference an AwsDynamodb resource or pass a
literal table name.

- references: AwsDynamodb (`status.outputs.table_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsDynamodb, name: <that resource's name>, fieldPath: status.outputs.table_name}} -- a bare string does not parse

### spec.datasources[].dynamodb.region

`string`

The table's region when it differs from the API's region.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.datasources[].dynamodb.useCallerCredentials

`bool`

Use the CALLER's IAM credentials (from an AWS_IAM-authorized
request) instead of the data source's service role.

### spec.datasources[].dynamodb.versioned

`bool`

Mark the table as versioned for conflict detection (pairs with
resolvers' sync_config).

### spec.datasources[].dynamodb.deltaSync

`AwsAppSyncDatasourceDeltaSync`

Delta Sync: journal changes into a companion table so offline
clients can catch up incrementally.

### spec.datasources[].dynamodb.deltaSync.deltaSyncTableName

`string` · required

The companion table holding the change journal.

- rule: {"string":{"minLen":"1"}}

### spec.datasources[].dynamodb.deltaSync.baseTableTtl

`int64`

Minutes an item lives in the base table's TTL attribute. Unset
(0) means no base-table TTL.

### spec.datasources[].dynamodb.deltaSync.deltaSyncTableTtl

`int64`

Minutes a journal entry lives in the delta table.

### spec.datasources[].lambda

`AwsAppSyncDatasourceLambda`

Lambda backend (type AWS_LAMBDA).

### spec.datasources[].lambda.functionArn

`string | valueFrom` · required

The function. Can reference an AwsLambda resource or pass a
literal function ARN.

- references: AwsLambda (`status.outputs.function_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsLambda, name: <that resource's name>, fieldPath: status.outputs.function_arn}} -- a bare string does not parse

### spec.datasources[].http

`AwsAppSyncDatasourceHttp`

HTTP backend (type HTTP).

### spec.datasources[].http.endpoint

`string` · required

The endpoint URL (e.g. "https://api.example.com").

- rule: {"string":{"minLen":"1"}}

### spec.datasources[].http.sigv4

`AwsAppSyncDatasourceHttpSigning`

Sign requests with SigV4 (for calling AWS services directly).
Unset sends unsigned requests. (AWS_IAM is the only
authorization type - the modules pin it when this block is set.)

### spec.datasources[].http.sigv4.signingRegion

`string`

The region to sign for (the target service's region).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.datasources[].http.sigv4.signingServiceName

`string`

The service name to sign for (e.g. "states" for Step Functions).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.datasources[].opensearch

`AwsAppSyncDatasourceOpensearch`

OpenSearch Service backend (type AMAZON_OPENSEARCH_SERVICE).

### spec.datasources[].opensearch.endpoint

`string | valueFrom` · required

The domain endpoint (e.g. "https://search-...es.amazonaws.com").
Can reference an AwsOpenSearchDomain resource or pass a literal
endpoint URL.

- references: AwsOpenSearchDomain (`status.outputs.endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsOpenSearchDomain, name: <that resource's name>, fieldPath: status.outputs.endpoint}} -- a bare string does not parse

### spec.datasources[].opensearch.region

`string`

The domain's region when it differs from the API's region.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.datasources[].elasticsearch

`AwsAppSyncDatasourceElasticsearch`

Legacy Elasticsearch backend (type AMAZON_ELASTICSEARCH).

### spec.datasources[].elasticsearch.endpoint

`string` · required

The domain endpoint.

- rule: {"string":{"minLen":"1"}}

### spec.datasources[].elasticsearch.region

`string`

The domain's region when it differs from the API's region.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.datasources[].eventbridge

`AwsAppSyncDatasourceEventbridge`

EventBridge backend (type AMAZON_EVENTBRIDGE). NOTE: the
provider's update path drops this block (an upstream defect at
the pin) - treat EventBridge data sources as replace-to-change:
rename the entry instead of editing it in place.

### spec.datasources[].eventbridge.eventBusArn

`string | valueFrom` · required

The event bus. Can reference an AwsEventBridgeBus resource or
pass a literal bus ARN.

Containment-exempt: the API PUBLISHES events onto the bus; it is not
deployed into it. On a diagram the bus is a room its rules live in,
and the API stands outside with a line in -- the verdict an SES
configuration set's event-bus destination already carries.

- references: AwsEventBridgeBus (`status.outputs.bus_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEventBridgeBus, name: <that resource's name>, fieldPath: status.outputs.bus_arn}} -- a bare string does not parse

### spec.datasources[].relationalDatabase

`AwsAppSyncDatasourceRelationalDatabase`

Aurora Serverless Data API backend (type RELATIONAL_DATABASE).

### spec.datasources[].relationalDatabase.dbClusterIdentifier

`string | valueFrom` · required

The Aurora cluster (Data API enabled). Can reference an
AwsRdsCluster resource or pass a literal cluster identifier.

- references: AwsRdsCluster (`status.outputs.cluster_identifier`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRdsCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_identifier}} -- a bare string does not parse

### spec.datasources[].relationalDatabase.awsSecretStoreArn

`string | valueFrom` · required

The Secrets Manager secret holding the database credentials. Can
reference an AwsSecretsManagerSecret resource or pass a literal
secret ARN.

- references: AwsSecretsManagerSecret (`status.outputs.secret_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecretsManagerSecret, name: <that resource's name>, fieldPath: status.outputs.secret_arn}} -- a bare string does not parse

### spec.datasources[].relationalDatabase.databaseName

`string`

The database to use within the cluster.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.datasources[].relationalDatabase.schema

`string`

The logical schema to use.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.datasources[].relationalDatabase.region

`string`

The cluster's region when it differs from the API's region.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.apiKeys

`[]AwsAppSyncApiKey`

API keys for the API_KEY authorization mode, keyed by a spec-side
name (AWS assigns the real key id). Only useful when the chosen
arm actually enables API_KEY auth - a key on an API without that
mode authenticates nothing.

### spec.apiKeys[].name

`string` · required

The entry's key in the for_each and the api_key_ids output map
(spec-side only - AWS keys have no name).

- rule: {"string":{"minLen":"1","pattern":"^[A-Za-z_][0-9A-Za-z_]*$"}}

### spec.apiKeys[].description

`string`

What this key is for. Unset means AWS's default description.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.apiKeys[].expires

`string`

When the key expires, RFC 3339 (e.g. "2027-02-01T00:00:00Z").
AWS rounds to the nearest hour; minimum 1 day, maximum 365 days
from creation. Unset means AWS's default (7 days).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}(\\.\\d+)?(Z|[+-]\\d{2}:\\d{2})$"}}

### spec.customDomain

`AwsAppSyncCustomDomain`

A custom domain fronting the API instead of the AWS-generated
endpoint. AppSync custom domains require an ACM certificate in
us-east-1 regardless of the API's own region (the CloudFront
class).

- rule: AppSync custom domains require an ACM certificate in us-east-1 regardless of the API's region

### spec.customDomain.domainName

`string` · required

The domain name (e.g. "api.example.com"). Fixed for life. After
apply, point your DNS at the appsync_domain_name output (a
CNAME, or a Route53 alias using domain_hosted_zone_id).

- rule: {"string":{"minLen":"1","maxLen":"253","pattern":"^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$"}}

### spec.customDomain.certificateArn

`string | valueFrom` · required

The ACM certificate covering the domain - must be in us-east-1.
Fixed for life. Can reference an AwsCertManagerCert resource or
pass a literal certificate ARN.

- references: AwsCertManagerCert (`status.outputs.cert_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCertManagerCert, name: <that resource's name>, fieldPath: status.outputs.cert_arn}} -- a bare string does not parse

### spec.customDomain.description

`string`

What this domain serves.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

## Validation Rules

- `spec.exactly_one_api_arm`: set exactly one of graphql and events - an AppSync API is either a GraphQL API or an Events API
- `spec.datasource_names_unique`: data source names must be unique within the API
- `spec.api_key_names_unique`: api key names must be unique within the API

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsAppSyncApi, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.api_id` | `string` | The API's id - the provider's import ID for the pivot, the prefix of every satellite's composite import ID, and the join key MERGED APIs use to reference their sources. |
| `status.outputs.api_arn` | `string` | The API's ARN. |
| `status.outputs.graphql_url` | `string` | GraphQL arm: the GraphQL endpoint URL clients query - the chart-ready join key for application configuration. |
| `status.outputs.realtime_url` | `string` | GraphQL arm: the real-time (subscriptions) endpoint URL. |
| `status.outputs.events_http_endpoint` | `string` | Events arm: the HTTP endpoint domain clients publish through. |
| `status.outputs.events_realtime_endpoint` | `string` | Events arm: the real-time (WebSocket) endpoint domain clients subscribe through. |
| `status.outputs.appsync_domain_name` | `string` | The AppSync-managed domain to point DNS at when custom_domain is set (a CNAME target, or a Route53 alias with domain_hosted_zone_id). |
| `status.outputs.domain_hosted_zone_id` | `string` | The Route53 hosted zone id for alias records at the custom domain. |
| `status.outputs.datasource_arns` | `map<string, string>` | Data source ARNs keyed by spec datasource name. |
| `status.outputs.function_ids` | `map<string, string>` | Function ids keyed by spec function name - part of each function's composite import ID. |
| `status.outputs.api_key_ids` | `map<string, string>` | API key ids keyed by spec key name - part of each key's composite import ID. NOTE: these are the key IDs, not the secrets; AWS returns a key's secret only at creation (fetch it from the console/CLI). |
| `status.outputs.channel_namespace_arns` | `map<string, string>` | Channel namespace ARNs keyed by namespace name (Events arm). |
| `status.outputs.source_api_association_ids` | `map<string, string>` | Source API association ids keyed by spec source-API entry name (MERGED APIs) - part of each association's composite import ID. |
| `status.outputs.type_formats` | `map<string, string>` | Import-derivation echo map: each managed type's format (SDL/JSON) keyed by type name - part of the type's composite import ID ("{api_id}:{format}:{name}"). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.graphql.auth.userPool.userPoolId` | AwsCognitoUserPool | `status.outputs.user_pool_id` |
| `spec.graphql.auth.lambda.authorizerUri` | AwsLambda | `status.outputs.function_arn` |
| `spec.graphql.additionalAuthProviders[].userPool.userPoolId` | AwsCognitoUserPool | `status.outputs.user_pool_id` |
| `spec.graphql.additionalAuthProviders[].lambda.authorizerUri` | AwsLambda | `status.outputs.function_arn` |
| `spec.graphql.logConfig.cloudwatchLogsRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.graphql.webAclArn` | AwsWafWebAcl | `status.outputs.web_acl_arn` |
| `spec.graphql.functions[].syncConfig.lambdaConflictHandlerArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.graphql.resolvers[].syncConfig.lambdaConflictHandlerArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.graphql.merged.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.graphql.merged.sourceApis[].sourceApiId` | AwsAppSyncApi | `status.outputs.api_id` |
| `spec.events.authProviders[].cognito.userPoolId` | AwsCognitoUserPool | `status.outputs.user_pool_id` |
| `spec.events.authProviders[].lambda.authorizerUri` | AwsLambda | `status.outputs.function_arn` |
| `spec.events.logConfig.cloudwatchLogsRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.datasources[].serviceRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.datasources[].dynamodb.tableName` | AwsDynamodb | `status.outputs.table_name` |
| `spec.datasources[].lambda.functionArn` | AwsLambda | `status.outputs.function_arn` |
| `spec.datasources[].opensearch.endpoint` | AwsOpenSearchDomain | `status.outputs.endpoint` |
| `spec.datasources[].eventbridge.eventBusArn` | AwsEventBridgeBus | `status.outputs.bus_arn` |
| `spec.datasources[].relationalDatabase.dbClusterIdentifier` | AwsRdsCluster | `status.outputs.cluster_identifier` |
| `spec.datasources[].relationalDatabase.awsSecretStoreArn` | AwsSecretsManagerSecret | `status.outputs.secret_arn` |
| `spec.customDomain.certificateArn` | AwsCertManagerCert | `status.outputs.cert_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppSyncApi | `spec.graphql.merged.sourceApis[].sourceApiId` | `status.outputs.api_id` |

## See Also

- [Overview](../README.md)
