# AwsRoute53ResolverFirewall

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsRoute53ResolverFirewallSpec defines one Route 53 Resolver DNS
Firewall rule group - the filtering policy for DNS queries leaving
VPCs - with its domain lists, rules, and VPC associations managed
in-line.

The rule group is the pivot: rules cannot exist without one, and
associating the group to a VPC is what puts the policy into effect
there. Domain lists authored here are owned by this group's
lifecycle; rules can also reference AWS-managed domain lists (or any
external list) by ID.

The per-VPC fail-open setting (what happens when the firewall
cannot evaluate a query) is deliberately NOT here - it is a VPC
setting, not a rule-group setting, and lives with the VPC's other
resolver settings.

## Example

```yaml
# Canonical AwsRoute53ResolverFirewall example (hack/dev manifest and
# refgen Example source): a rule group with an owned blocklist, a BLOCK
# rule sinkholing matches, an ALERT rule on DNS tunneling detection, and
# the group associated to one VPC.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53ResolverFirewall
metadata:
  name: egress-dns-policy
  id: egress-dns-policy
  org: test-org
  env: dev
spec:
  region: us-west-2
  domainLists:
    - name: blocked-domains
      domains:
        - malware.example.
        - phishing.example.
  rules:
    - name: sinkhole-blocked
      priority: 100
      action: BLOCK
      domainListName: blocked-domains
      blockResponse: OVERRIDE
      blockOverrideDomain: sinkhole.example.com
      blockOverrideTtl: 300
    - name: alert-tunneling
      priority: 200
      action: ALERT
      dnsThreatProtection: DNS_TUNNELING
      confidenceThreshold: HIGH
  vpcAssociations:
    - name: main-vpc
      vpcId:
        value: vpc-0123456789abcdef0
      priority: 200
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.domainLists` | `[]AwsRoute53ResolverFirewallDomainList` |  |  |  |
| `spec.domainLists[].name` | `string` | yes |  |  |
| `spec.domainLists[].domains` | `[]string` |  |  |  |
| `spec.rules` | `[]AwsRoute53ResolverFirewallRule` |  |  |  |
| `spec.rules[].name` | `string` | yes |  |  |
| `spec.rules[].priority` | `int64` | yes |  |  |
| `spec.rules[].action` | `string` |  |  |  |
| `spec.rules[].domainListName` | `string` |  |  |  |
| `spec.rules[].domainListId` | `string \| valueFrom` |  |  |  |
| `spec.rules[].dnsThreatProtection` | `string` |  |  |  |
| `spec.rules[].confidenceThreshold` | `string` |  |  |  |
| `spec.rules[].blockResponse` | `string` |  |  |  |
| `spec.rules[].blockOverrideDomain` | `string` | yes |  |  |
| `spec.rules[].blockOverrideTtl` | `int64` |  |  |  |
| `spec.rules[].firewallDomainRedirectionAction` | `string` |  |  |  |
| `spec.rules[].qType` | `string` | yes |  |  |
| `spec.vpcAssociations` | `[]AwsRoute53ResolverFirewallVpcAssociation` |  |  |  |
| `spec.vpcAssociations[].name` | `string` | yes |  |  |
| `spec.vpcAssociations[].vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.vpcAssociations[].priority` | `int64` | yes |  |  |
| `spec.vpcAssociations[].mutationProtection` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the rule group and its satellites live in.
Example: "us-west-2".

- rule: {"string":{"minLen":"1"}}

### spec.domainLists

`[]AwsRoute53ResolverFirewallDomainList`

The domain lists owned by this rule group, keyed by name. Rules
reference them by that name.

### spec.domainLists[].name

`string` · required

List name - the for_each key on both engines, the key in the
domain_list_ids output, and what rules reference via
domain_list_name. Also becomes the list's AWS name.

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_-]+$"}}

### spec.domainLists[].domains

`[]string`

The domains on the list, e.g. "malware.example." - AWS treats each
entry as the domain and all its subdomains. A "*." prefix matches
subdomains only. AWS stores every entry as a trailing-dot FQDN;
both modules append the dot when absent, so either form is fine
here and never causes drift.

- rule: domains must be unique within a list

### spec.rules

`[]AwsRoute53ResolverFirewallRule`

The filtering rules in this group, keyed by name. Evaluated in
ascending priority order; the first matching rule's action
applies.

- rule: configure exactly one of domain_list_name (an owned list), domain_list_id (an external or AWS-managed list), and dns_threat_protection
- rule: dns_threat_protection and confidence_threshold are set together
- rule: block_response applies only when action is BLOCK
- rule: block_response OVERRIDE requires block_override_domain and block_override_ttl
- rule: block_override_domain and block_override_ttl apply only when block_response is OVERRIDE

### spec.rules[].name

`string` · required

Rule name - the for_each key on both engines. Also becomes the
rule's AWS name.

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_-]+$"}}

### spec.rules[].priority

`int64` · required

Where this rule sits in the group's evaluation order (ascending;
first match wins). Must be unique within the group. Leave gaps
(100, 200, ...) so rules can be inserted later without renumbering.

- rule: {"required":true}

### spec.rules[].action

`string`

What to do with a matching query: ALLOW it through, ALERT (allow
but log), or BLOCK it (shape the response with block_response).

- rule: {"string":{"in":["ALLOW","BLOCK","ALERT"]}}

### spec.rules[].domainListName

`string`

Match against a domain list owned by this rule group (an entry in
domain_lists, referenced by name).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"64"}}

### spec.rules[].domainListId

`string | valueFrom`

Match against an external domain list by ID - an AWS-managed list
(e.g. the malware or aggregate threat lists) or one owned by
another rule group. Pass a literal rslvr-fdl-... id.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.rules[].dnsThreatProtection

`string`

Match by DNS threat class instead of a domain list (Advanced
protection): DGA catches algorithmically generated domains,
DICTIONARY_DGA the dictionary-word variant, DNS_TUNNELING data
exfiltration over DNS.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DGA","DICTIONARY_DGA","DNS_TUNNELING"]}}

### spec.rules[].confidenceThreshold

`string`

How confident the threat detection must be before the rule fires.
Set together with dns_threat_protection.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["LOW","MEDIUM","HIGH"]}}

### spec.rules[].blockResponse

`string`

How a BLOCK rule answers the query. NODATA says the domain exists
but has no records; NXDOMAIN says it does not exist; OVERRIDE
substitutes the record configured below. Unset means AWS's
default (NODATA).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NODATA","NXDOMAIN","OVERRIDE"]}}

### spec.rules[].blockOverrideDomain

`string` · required

The record value an OVERRIDE response returns (AWS returns it as
a CNAME - the record type has exactly one legal value, so the
modules pin it). AWS stores it as a trailing-dot FQDN
("sinkhole.example.com."); both modules append the dot when
absent, so either form is fine here and never causes drift.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"255"}}

### spec.rules[].blockOverrideTtl

`int64` · optional (explicit presence)

The TTL (seconds, 0-604800) of the OVERRIDE record.

- rule: block_override_ttl must be between 0 and 604800 seconds

### spec.rules[].firewallDomainRedirectionAction

`string`

How the firewall treats domains a matching CNAME chain redirects
to. Unset means AWS's default (INSPECT_REDIRECTION_DOMAIN - the
redirected-to domain is evaluated against the rules too);
TRUST_REDIRECTION_DOMAIN skips that evaluation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["INSPECT_REDIRECTION_DOMAIN","TRUST_REDIRECTION_DOMAIN"]}}

### spec.rules[].qType

`string` · required

Restrict the rule to one DNS record type (e.g. "TXT", "MX").
Unset matches all query types.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1","maxLen":"16"}}

### spec.vpcAssociations

`[]AwsRoute53ResolverFirewallVpcAssociation`

The VPCs this rule group filters, keyed by association name. A
VPC can carry multiple rule-group associations; AWS evaluates
them in ascending association priority order.

### spec.vpcAssociations[].name

`string` · required

Association name - the for_each key on both engines and the key
in the association_ids output. Also becomes the association's AWS
name.

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_-]+$"}}

### spec.vpcAssociations[].vpcId

`string | valueFrom` · required

The VPC to filter. Reference an AwsVpc vpc_id output or pass a
literal vpc-... id.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.vpcAssociations[].priority

`int64` · required

Where this group sits among the VPC's rule-group associations
(ascending evaluation order - the lowest number filters first).

- rule: {"required":true}

### spec.vpcAssociations[].mutationProtection

`string`

Protect the association from accidental deletion or modification.
Unset leaves AWS's default (DISABLED).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

## Validation Rules

- `spec.domain_list_names_unique`: domain list names must be unique within the rule group
- `spec.rule_names_unique`: rule names must be unique within the rule group
- `spec.association_names_unique`: vpc association names must be unique within the rule group
- `spec.rule_priorities_unique`: rule priorities must be unique within the rule group
- `spec.rule_owned_list_exists`: each rule's domain_list_name must name an entry in domain_lists

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRoute53ResolverFirewall, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.rule_group_id` | `string` | The rule group's id (rslvr-frg-...) - the provider's import ID and half of each rule's composite import ID. |
| `status.outputs.rule_group_arn` | `string` | The rule group's ARN. |
| `status.outputs.share_status` | `string` | Whether the group is shared via RAM (NOT_SHARED / SHARED_BY_ME / SHARED_WITH_ME). |
| `status.outputs.domain_list_ids` | `map<string, string>` | AWS-generated domain list IDs (rslvr-fdl-...) keyed by list name - what rules and imports reference. |
| `status.outputs.association_ids` | `map<string, string>` | AWS-generated VPC association IDs (rslvr-frgassoc-...) keyed by association name. |
| `status.outputs.rule_match_ids` | `map<string, string>` | Each rule's match identity keyed by rule name - the domain list ID for standard rules, the threat-protection ID (rslvr-ftp-...) for advanced rules. The second half of the rule's composite import ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpcAssociations[].vpcId` | AwsVpc | `status.outputs.vpc_id` |

## See Also

- [Overview](../README.md)
