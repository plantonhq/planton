# AwsCognitoResourceServer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsCognitoResourceServerSpec defines the desired configuration for a Cognito
resource server -- the OAuth 2.0 resource (an API) a user pool mints custom
access-token scopes for.

A resource server is deliberately its own resource rather than a field on
the pool: a pool protects many APIs, each with its own identifier and scope
vocabulary, and the scopes it mints ("{identifier}/{scope_name}") are what
app clients request in `allowed_oauth_scopes` -- most importantly
machine-to-machine clients using the client_credentials grant, which can
ONLY request custom scopes.

Key design notes:
- `identifier` is **ForceNew**: it is the resource server's identity within
  the pool (conventionally the API's audience URI, e.g.
  "https://api.example.com").
- Scopes update in place. Removing a scope invalidates it for future
  tokens; already-issued tokens carry it until they expire.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsCognitoResourceServer
metadata:
  name: test-orders-api
  org: test-org
  env: dev
  id: awscogrs-test-001
spec:
  region: us-west-2
  userPoolId:
    value: us-west-2_TestPool123
  identifier: https://api.example.com
  name: orders-api
  scopes:
    - scopeName: read
      scopeDescription: Read access to orders
    - scopeName: write
      scopeDescription: Write access to orders
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.userPoolId` | `string \| valueFrom` | yes |  | AwsCognitoUserPool (`status.outputs.user_pool_id`) |
| `spec.identifier` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.scopes` | `[]AwsCognitoResourceServerScope` |  |  |  |
| `spec.scopes[].scopeName` | `string` | yes |  |  |
| `spec.scopes[].scopeDescription` | `string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.userPoolId

`string | valueFrom` · required

The Cognito User Pool this resource server belongs to.
Format: "{region}_{poolId}" (e.g., "us-east-1_Ab1Cd2EfG").
ForceNew -- a resource server cannot be moved between pools.
Accepts a direct pool ID or a reference to an AwsCognitoUserPool resource.

- references: AwsCognitoUserPool (`status.outputs.user_pool_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCognitoUserPool, name: <that resource's name>, fieldPath: status.outputs.user_pool_id}} -- a bare string does not parse

### spec.identifier

`string` · required

The resource server's unique identifier within the pool -- the prefix of
every scope it mints ("{identifier}/{scope_name}") and the value access
tokens carry. Conventionally the API's audience URI (e.g.
"https://api.example.com"). 1-256 characters. ForceNew.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.name

`string` · required

Human-readable display name shown in the Cognito console. 1-256
characters.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.scopes

`[]AwsCognitoResourceServerScope`

The custom scopes this resource server defines. Each becomes requestable
by app clients as "{identifier}/{scope_name}". Maximum 100 scopes.

- rule: {"repeated":{"maxItems":"100"}}
- rule: scope_name cannot contain spaces, '/', '"', or '\' -- '/' is reserved as the identifier/scope separator

### spec.scopes[].scopeName

`string` · required

The scope name (e.g. "read", "orders:write"). 1-256 characters; letters,
digits, and most punctuation are allowed, but not spaces, '"', '/', or
'\\' (the '/' is reserved as the identifier/scope separator).

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.scopes[].scopeDescription

`string` · required

What granting this scope means, shown on consent screens and in the
console. 1-256 characters.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

## Validation Rules

- `scope_names_unique`: scope names must be unique within the resource server

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsCognitoResourceServer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.resource_server_identifier` | `string` | The resource server's identifier within its pool -- the scope prefix access tokens carry (e.g. "https://api.example.com"). |
| `status.outputs.scope_identifiers` | `[]string` | The fully-qualified scope identifiers this resource server mints, in "{identifier}/{scope_name}" form (e.g. "https://api.example.com/read"). These are the exact strings app clients list in allowed_oauth_scopes. |
| `status.outputs.user_pool_id` | `string` | The user pool this resource server belongs to, resolved from the spec reference. Resource servers are keyed by (pool id, identifier) in AWS, and a consumer holding only this resource gets both halves of that key from its outputs. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.userPoolId` | AwsCognitoUserPool | `status.outputs.user_pool_id` |

## See Also

- [Overview](../README.md)
