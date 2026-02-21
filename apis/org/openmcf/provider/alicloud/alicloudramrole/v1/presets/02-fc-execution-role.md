# FC Execution Role

This preset creates a RAM role that Alibaba Cloud Function Compute can assume when executing functions. It includes Log Service (SLS) full access so function invocation logs are written to your SLS project. This is the minimal viable execution role -- add policies for OSS, RDS, Redis, or other services your functions access.

## When to Use

- Any Function Compute function that needs to write execution logs to SLS
- Starting point for serverless roles; add service-specific policies based on what the function accesses
- Functions deployed via `AliCloudFcFunction` that reference this role's ARN

## Key Configuration Choices

- **FC service principal** (`fc.aliyuncs.com`) -- Only Function Compute can assume this role
- **Log Service full access** (`AliyunLogFullAccess`) -- Enables FC to write invocation logs, execution traces, and application logs to any SLS project. This is the baseline permission every FC function needs for observability.
- **Minimal policy set** -- Intentionally limited to logging only. Add `AliyunOSSReadOnlyAccess`, `AliyunVPCReadOnlyAccess`, `AliyunECSNetworkInterfaceManagement`, or custom policies depending on what resources your function accesses.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-fc-role-name>` | RAM role name, unique per Alibaba Cloud account (1-64 chars: letters, digits, `.`, `-`, `_`) | Choose a name following your organization's naming convention |

## Related Presets

- **01-ecs-service-role** -- Use instead for ECS instance or ACK worker node roles
- **03-cross-account-audit** -- Use instead for cross-account security audit access
