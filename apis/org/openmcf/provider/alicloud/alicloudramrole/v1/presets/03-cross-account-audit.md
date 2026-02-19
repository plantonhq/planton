# Cross-Account Audit Role

This preset creates a RAM role that another Alibaba Cloud account can assume for read-only security auditing. The trusted account gains access to billing data (BSS), centralized logs (SLS), and operational audit trails (ActionTrail). The 12-hour session duration accommodates extended audit sessions without re-authentication. All attached policies are read-only -- the auditor cannot modify any resources.

## When to Use

- Multi-account organizations where a security or compliance team needs read-only visibility into other accounts
- Regulatory audit scenarios requiring access to billing records and operational logs
- Centralized security operations centers (SOC) that monitor multiple Alibaba Cloud accounts

## Key Configuration Choices

- **Cross-account RAM principal** (`acs:ram::<account-id>:root`) -- Trusts the root identity of another Alibaba Cloud account. Replace `<trusted-account-id>` with the 16-digit account ID of the auditing account.
- **12-hour session duration** (`maxSessionDuration: 43200`) -- Maximum allowed value. Extended sessions prevent auditors from needing to re-assume the role during long reviews. Default is 3600 (1 hour).
- **Force deletion enabled** (`force: true`) -- Allows the role to be deleted even with policies attached. Useful when decommissioning audit relationships; prevents orphaned roles.
- **Billing read-only** (`AliyunBSSReadOnlyAccess`) -- Grants read access to billing, cost, and usage data without modification rights
- **Log read-only** (`AliyunLogReadOnlyAccess`) -- Grants read access to all SLS projects and log stores for log analysis
- **ActionTrail read-only** (`AliyunActionTrailReadOnlyAccess`) -- Grants read access to API call audit trails for compliance review

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-audit-role-name>` | RAM role name, unique per Alibaba Cloud account (1-64 chars: letters, digits, `.`, `-`, `_`) | Choose a name following your organization's naming convention |
| `<trusted-account-id>` | 16-digit Alibaba Cloud account ID of the auditing account | Alibaba Cloud console > Account Management, or provided by the audit team |

## Related Presets

- **01-ecs-service-role** -- Use instead for ECS instance or ACK worker node roles
- **02-fc-execution-role** -- Use instead for Function Compute execution roles
