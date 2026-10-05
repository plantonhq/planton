# AzureWebApplicationFirewallPolicy

A regional Web Application Firewall (WAF) policy -- the rule set an Azure
Application Gateway enforces on HTTP traffic. One policy carries three
layers: custom rules (IP/geo allowlists, rate limits, bot challenges),
Microsoft's managed rule sets (OWASP core rule set, bot manager) with
per-rule tuning and scoped exclusions, and policy settings (Prevention vs
Detection, body-inspection limits, log scrubbing).

This is the Application Gateway policy type. Azure Front Door's WAF policy
is a different ARM resource with a different vocabulary.

## When to Use

Use AzureWebApplicationFirewallPolicy when you need:

- **OWASP protection** (SQL injection, XSS, RCE, LFI, protocol attacks) in
  front of web applications behind an Application Gateway (WAF_v2 SKU)
- **Rate limiting** per client address, XFF address, or geography
- **Geo fencing** and IP allowlisting ahead of the managed rules
- **Bot mitigation** via Microsoft's bot-manager rule set and JavaScript
  challenges
- **One org-standard policy shared across gateways**, with per-listener or
  per-route overrides where specific applications need stricter or looser
  rules

## How It Attaches

The policy is referenced by ARM ID (the `policy_id` output) at three
levels of an `AzureApplicationGateway`, most specific wins:

1. `firewall_policy_id` -- the gateway-wide default
2. `http_listeners[].firewall_policy_id` -- per listener (per site)
3. `url_path_map[].path_rules[].firewall_policy_id` -- per route

## Rule Evaluation Order

1. Custom rules, ascending priority (1-100); first terminal action wins
2. Managed rule sets (OWASP anomaly scoring / bot classification)
3. Policy settings govern what "block" means (Prevention vs Detection)

## Key Configuration

- **`managed_rules`** (required): at least one rule set -- OWASP `"3.2"`
  is the production standard; add `MICROSOFT_BOT_MANAGER_RULE_SET` `"1.1"`
  for bot classification. Tune with `rule_group_overrides` (a rule listed
  without `enabled: true` is DISABLED) and scoped `exclusions`.
- **`custom_rules`**: MATCH_RULE or RATE_LIMIT_RULE; actions ALLOW / BLOCK
  / LOG / JS_CHALLENGE (rate-limit rules cannot ALLOW). Conditions AND
  together; values within a condition OR together.
- **`policy_settings`**: `mode` (PREVENTION default / DETECTION for
  tuning), body-inspection dials, file-upload limits, JS-challenge cookie
  lifetime, and `log_scrubbing` to redact secrets from WAF logs.

## Outputs

| Output | Description |
|--------|-------------|
| `policy_id` | The policy's ARM ID -- what gateways, listeners, and path rules reference |
| `policy_name` | The policy's name |

## Related Resources

- **AzureApplicationGateway** -- the L7 load balancer that enforces the policy
- **AzureResourceGroup** -- the policy's container

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
