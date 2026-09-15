# AzureFrontDoorRuleSet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureFrontDoorRuleSetSpec** defines the configuration for creating a
rule set inside an Azure Front Door (Standard/Premium) profile -- the
edge delivery policy a route attaches to. A rule set is an ordered list
of rules; each rule pairs match conditions (what traffic it applies to)
with actions (what happens: redirect, rewrite, header edits, or a
per-request override of the route's caching and forwarding).

The rules live INSIDE the rule set rather than as standalone resources:
they form one ordered policy document (evaluation order is a property
of the set), nothing references an individual rule, and a rule is
meaningless outside its set. Routes attach the whole set by ARM ID
(see the rule_set_id output); one set is commonly shared by many
routes, which is exactly why the set is its own first-class resource
rather than settings folded into a route.

**ForceNew fields**: `profile_id`, `rule_set_name`, and each rule's
`name` -- they fix ARM identities at creation. Everything else
(order, conditions, actions) updates in place.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorRuleSet
metadata:
  name: test-front-door-rule-set
spec:
  profileId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Cdn/profiles/test-frontdoor
  ruleSetName: deliverypolicy
  rules:
    # Exercises: STOP behavior, host-name condition with a transform,
    # the redirect action with the redirect protocol dialect (HTTPS_ONLY
    # must render "Https" here, not "HttpsOnly").
    - name: apexredirect
      order: 1
      behaviorOnMatch: STOP
      conditions:
        hostName:
          - operator: EQUAL
            matchValues: [example.com]
            transforms: [LOWERCASE]
      actions:
        urlRedirect:
          redirectType: PERMANENT_REDIRECT
          redirectProtocol: HTTPS_ONLY
          destinationHostname: www.example.com
    # Exercises: order 0 (the tfvars zero-drop seam), geo + scheme +
    # method + device + version + port + TLS conditions (the
    # closed-vocabulary pass-throughs), and both header action kinds
    # incl. DELETE without a value.
    - name: securityheaders
      order: 0
      conditions:
        remoteAddress:
          - operator: GEO_MATCH
            negateCondition: true
            matchValues: [US, DE]
        requestMethod:
          - matchValues: [GET, POST]
        requestScheme:
          - matchValue: HTTPS
        isDevice:
          - matchValue: Mobile
        httpVersion:
          - matchValues: ["2.0", "1.1"]
        serverPort:
          - operator: EQUAL
            matchValues: ["443"]
        sslProtocol:
          - matchValues: [TLSv1.2]
      actions:
        requestHeaders:
          - headerAction: APPEND
            headerName: X-Forwarded-Planton
            value: "1"
          - headerAction: DELETE
            headerName: X-Debug
        responseHeaders:
          - headerAction: OVERWRITE
            headerName: Strict-Transport-Security
            value: max-age=31536000; includeSubDomains
    # Exercises: url-path WILDCARD, the rewrite action, and the
    # unspecified-operator IP_MATCH default on socket addresses.
    - name: apirewrite
      order: 2
      conditions:
        urlPath:
          - operator: WILDCARD
            matchValues: [api/*/v1]
        socketAddress:
          - matchValues: [203.0.113.0/24]
      actions:
        urlRewrite:
          sourcePattern: /api
          destination: /internal/api
          preserveUnmatchedPath: true
    # Exercises: the route-configuration override's full caching matrix
    # (OVERRIDE_ALWAYS + duration + specified query strings), the origin
    # override with the OVERRIDE forwarding dialect (HTTPS_ONLY must
    # render "HttpsOnly" here), and compression.
    - name: staticcache
      order: 3
      conditions:
        urlFileExtension:
          - operator: EQUAL
            matchValues: [css, js, woff2]
            transforms: [LOWERCASE, TRIM]
      actions:
        routeConfigurationOverride:
          originGroupId:
            value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Cdn/profiles/test-frontdoor/originGroups/static-backends
          forwardingProtocol: HTTPS_ONLY
          cacheBehavior: OVERRIDE_ALWAYS
          cacheDuration: 1.12:00:00
          queryStringCachingBehavior: INCLUDE_SPECIFIED_QUERY_STRINGS
          queryStringParameters: [page, lang]
          compressionEnabled: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.profileId` | `string \| valueFrom` | yes |  | AzureFrontDoorProfile (`status.outputs.profile_id`) |
| `spec.ruleSetName` | `string` | yes |  |  |
| `spec.rules` | `[]AzureFrontDoorRule` |  |  |  |
| `spec.rules[].name` | `string` | yes |  |  |
| `spec.rules[].order` | `int32` |  |  |  |
| `spec.rules[].behaviorOnMatch` | `enum` |  |  |  |
| `spec.rules[].conditions` | `AzureFrontDoorRuleConditions` |  |  |  |
| `spec.rules[].conditions.remoteAddress` | `[]AzureFrontDoorRuleRemoteAddressCondition` |  |  |  |
| `spec.rules[].conditions.remoteAddress[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.remoteAddress[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.remoteAddress[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.requestMethod` | `[]AzureFrontDoorRuleRequestMethodCondition` |  |  |  |
| `spec.rules[].conditions.requestMethod[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.requestMethod[].matchValues` | `[]string` | yes |  |  |
| `spec.rules[].conditions.queryString` | `[]AzureFrontDoorRuleQueryStringCondition` |  |  |  |
| `spec.rules[].conditions.queryString[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.queryString[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.queryString[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.queryString[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.postArgs` | `[]AzureFrontDoorRulePostArgsCondition` |  |  |  |
| `spec.rules[].conditions.postArgs[].postArgsName` | `string` | yes |  |  |
| `spec.rules[].conditions.postArgs[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.postArgs[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.postArgs[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.postArgs[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.requestUri` | `[]AzureFrontDoorRuleRequestUriCondition` |  |  |  |
| `spec.rules[].conditions.requestUri[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.requestUri[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.requestUri[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.requestUri[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.requestHeader` | `[]AzureFrontDoorRuleRequestHeaderCondition` |  |  |  |
| `spec.rules[].conditions.requestHeader[].headerName` | `string` | yes |  |  |
| `spec.rules[].conditions.requestHeader[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.requestHeader[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.requestHeader[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.requestHeader[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.requestBody` | `[]AzureFrontDoorRuleRequestBodyCondition` |  |  |  |
| `spec.rules[].conditions.requestBody[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.requestBody[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.requestBody[].matchValues` | `[]string` | yes |  |  |
| `spec.rules[].conditions.requestBody[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.requestScheme` | `[]AzureFrontDoorRuleRequestSchemeCondition` |  |  |  |
| `spec.rules[].conditions.requestScheme[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.requestScheme[].matchValue` | `string` |  | `HTTP` |  |
| `spec.rules[].conditions.urlPath` | `[]AzureFrontDoorRuleUrlPathCondition` |  |  |  |
| `spec.rules[].conditions.urlPath[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.urlPath[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.urlPath[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.urlPath[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.urlFileExtension` | `[]AzureFrontDoorRuleUrlFileExtensionCondition` |  |  |  |
| `spec.rules[].conditions.urlFileExtension[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.urlFileExtension[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.urlFileExtension[].matchValues` | `[]string` | yes |  |  |
| `spec.rules[].conditions.urlFileExtension[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.urlFilename` | `[]AzureFrontDoorRuleUrlFilenameCondition` |  |  |  |
| `spec.rules[].conditions.urlFilename[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.urlFilename[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.urlFilename[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.urlFilename[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.httpVersion` | `[]AzureFrontDoorRuleHttpVersionCondition` |  |  |  |
| `spec.rules[].conditions.httpVersion[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.httpVersion[].matchValues` | `[]string` | yes |  |  |
| `spec.rules[].conditions.cookies` | `[]AzureFrontDoorRuleCookiesCondition` |  |  |  |
| `spec.rules[].conditions.cookies[].cookieName` | `string` | yes |  |  |
| `spec.rules[].conditions.cookies[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.cookies[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.cookies[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.cookies[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.isDevice` | `[]AzureFrontDoorRuleIsDeviceCondition` |  |  |  |
| `spec.rules[].conditions.isDevice[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.isDevice[].matchValue` | `string` | yes |  |  |
| `spec.rules[].conditions.socketAddress` | `[]AzureFrontDoorRuleSocketAddressCondition` |  |  |  |
| `spec.rules[].conditions.socketAddress[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.socketAddress[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.socketAddress[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.clientPort` | `[]AzureFrontDoorRuleClientPortCondition` |  |  |  |
| `spec.rules[].conditions.clientPort[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.clientPort[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.clientPort[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.serverPort` | `[]AzureFrontDoorRuleServerPortCondition` |  |  |  |
| `spec.rules[].conditions.serverPort[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.serverPort[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.serverPort[].matchValues` | `[]string` | yes |  |  |
| `spec.rules[].conditions.hostName` | `[]AzureFrontDoorRuleHostNameCondition` |  |  |  |
| `spec.rules[].conditions.hostName[].operator` | `enum` |  |  |  |
| `spec.rules[].conditions.hostName[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.hostName[].matchValues` | `[]string` |  |  |  |
| `spec.rules[].conditions.hostName[].transforms` | `[]enum` |  |  |  |
| `spec.rules[].conditions.sslProtocol` | `[]AzureFrontDoorRuleSslProtocolCondition` |  |  |  |
| `spec.rules[].conditions.sslProtocol[].negateCondition` | `bool` |  |  |  |
| `spec.rules[].conditions.sslProtocol[].matchValues` | `[]string` | yes |  |  |
| `spec.rules[].actions` | `AzureFrontDoorRuleActions` | yes |  |  |
| `spec.rules[].actions.urlRedirect` | `AzureFrontDoorRuleUrlRedirectAction` |  |  |  |
| `spec.rules[].actions.urlRedirect.redirectType` | `enum` |  |  |  |
| `spec.rules[].actions.urlRedirect.redirectProtocol` | `enum` |  |  |  |
| `spec.rules[].actions.urlRedirect.destinationHostname` | `string` |  |  |  |
| `spec.rules[].actions.urlRedirect.destinationPath` | `string` |  |  |  |
| `spec.rules[].actions.urlRedirect.queryString` | `string` |  |  |  |
| `spec.rules[].actions.urlRedirect.destinationFragment` | `string` |  |  |  |
| `spec.rules[].actions.urlRewrite` | `AzureFrontDoorRuleUrlRewriteAction` |  |  |  |
| `spec.rules[].actions.urlRewrite.sourcePattern` | `string` | yes |  |  |
| `spec.rules[].actions.urlRewrite.destination` | `string` | yes |  |  |
| `spec.rules[].actions.urlRewrite.preserveUnmatchedPath` | `bool` |  |  |  |
| `spec.rules[].actions.requestHeaders` | `[]AzureFrontDoorRuleHeaderAction` |  |  |  |
| `spec.rules[].actions.requestHeaders[].headerAction` | `enum` |  |  |  |
| `spec.rules[].actions.requestHeaders[].headerName` | `string` | yes |  |  |
| `spec.rules[].actions.requestHeaders[].value` | `string` |  |  |  |
| `spec.rules[].actions.responseHeaders` | `[]AzureFrontDoorRuleHeaderAction` |  |  |  |
| `spec.rules[].actions.responseHeaders[].headerAction` | `enum` |  |  |  |
| `spec.rules[].actions.responseHeaders[].headerName` | `string` | yes |  |  |
| `spec.rules[].actions.responseHeaders[].value` | `string` |  |  |  |
| `spec.rules[].actions.routeConfigurationOverride` | `AzureFrontDoorRuleRouteConfigurationOverrideAction` |  |  |  |
| `spec.rules[].actions.routeConfigurationOverride.originGroupId` | `string \| valueFrom` |  |  | AzureFrontDoorOriginGroup (`status.outputs.origin_group_id`) |
| `spec.rules[].actions.routeConfigurationOverride.forwardingProtocol` | `enum` |  |  |  |
| `spec.rules[].actions.routeConfigurationOverride.cacheBehavior` | `enum` |  |  |  |
| `spec.rules[].actions.routeConfigurationOverride.cacheDuration` | `string` |  |  |  |
| `spec.rules[].actions.routeConfigurationOverride.queryStringCachingBehavior` | `enum` |  |  |  |
| `spec.rules[].actions.routeConfigurationOverride.queryStringParameters` | `[]string` |  |  |  |
| `spec.rules[].actions.routeConfigurationOverride.compressionEnabled` | `bool` |  | `false` |  |

## Field Details

### spec.profileId

`string | valueFrom` · required

The Front Door profile the rule set lives in, by ARM ID. References
an AzureFrontDoorProfile's profile_id output so the profile and its
rule sets compose in one manifest set. Fixed at creation.

- references: AzureFrontDoorProfile (`status.outputs.profile_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorProfile, name: <that resource's name>, fieldPath: status.outputs.profile_id}} -- a bare string does not parse

### spec.ruleSetName

`string` · required

The rule set's name -- unique within the profile. Routes attach the
set by ARM ID, so the name is mostly a human-facing label; it is
also the segment under which each rule's ARM ID nests.

1-60 characters; must begin with a letter and may contain only
letters and numbers (Azure allows no hyphens here, unlike most
Front Door names).

**ForceNew**: changing the name replaces the rule set AND every
rule inside it.

- rule: rule_set_name must be 1-60 characters, begin with a letter, and contain only letters and numbers (no hyphens)
- rule: {"required":true}

### spec.rules

`[]AzureFrontDoorRule`

The rules that make up this delivery policy, evaluated in ascending
`order`. An empty set is legal (a placeholder routes can already
attach); add rules as the policy grows. Azure evaluates a request
against each rule in order: when a rule's conditions match, its
actions apply, and evaluation continues or stops per the rule's
behavior_on_match.

### spec.rules[].name

`string` · required

The rule's name -- unique within the rule set (each rule is an ARM
child resource of the set, keyed by this name). 1-260 characters;
must begin with a letter and may contain only letters and numbers.

**ForceNew**: renaming a rule replaces it (the ARM identity changes).

- rule: rule name must be 1-260 characters, begin with a letter, and contain only letters and numbers (no hyphens)
- rule: {"required":true}

### spec.rules[].order

`int32`

The rule's evaluation position within the set -- Azure evaluates
rules in ascending order. Must be 0 or greater. Convention: number
rules 1, 2, 3, ... and leave gaps (10, 20, 30) when you expect to
insert rules later; order only needs to be consistent, not
contiguous.

- rule: {"int32":{"gte":0}}

### spec.rules[].behaviorOnMatch

`enum`

What happens after this rule's actions apply: keep evaluating the
remaining rules (CONTINUE, Azure's default) or stop so later rules
never see the request (STOP). Use STOP for terminal actions like
redirects, where applying further rules would be surprising.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_behavior_on_match_unspecified` -- Not specified -- deploys CONTINUE, Azure's default.
- `CONTINUE` -- Keep evaluating the remaining rules in the set.
- `STOP` -- Stop evaluating -- later rules never see the request.

### spec.rules[].conditions

`AzureFrontDoorRuleConditions`

The match conditions. ALL conditions must match for the rule's
actions to apply (conditions AND together; the values WITHIN one
condition OR together). Omit entirely for a rule that applies to
every request -- e.g. a set-wide security-headers rule.

- rule: a rule may carry at most 10 match conditions across all condition types -- split the policy into additional rules if you need more

### spec.rules[].conditions.remoteAddress

`[]AzureFrontDoorRuleRemoteAddressCondition`

Match on the client's IP address (IP_MATCH against CIDR ranges) or
the country Azure geo-locates it to (GEO_MATCH against two-letter
ISO 3166-1 codes).

- rule: match_values must carry at least one CIDR range or country code
- rule: with the GEO_MATCH operator every match value must be a two-letter uppercase ISO 3166-1 country code, e.g. "US" or "DE"

### spec.rules[].conditions.remoteAddress[].operator

`enum`

IP_MATCH (default when unspecified) or GEO_MATCH -- the only
comparisons Azure's delivery-rule engine supports on client
addresses. IP_MATCH compares against CIDR ranges; GEO_MATCH against
two-letter country codes.

- rule: {"enum":{"definedOnly":true,"in":[0,12,13]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.remoteAddress[].negateCondition

`bool`

Invert the condition -- match when the address does NOT satisfy the
operator.

### spec.rules[].conditions.remoteAddress[].matchValues

`[]string`

1-25 values: IPv4/IPv6 CIDR ranges for IP_MATCH (e.g.
"203.0.113.0/24"), two-letter uppercase ISO 3166-1 country codes
for GEO_MATCH (e.g. "US"). Values OR together. Azure additionally
rejects duplicate and overlapping CIDRs at apply time.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.requestMethod

`[]AzureFrontDoorRuleRequestMethodCondition`

Match on the HTTP request method (GET, POST, ...).

### spec.rules[].conditions.requestMethod[].negateCondition

`bool`

Invert the condition -- match when the method is NOT in the list.

### spec.rules[].conditions.requestMethod[].matchValues

`[]string` · required

The methods to match, 1-7 unique values from: GET, HEAD, OPTIONS,
TRACE, POST, PUT, DELETE.

- rule: {"repeated":{"minItems":"1","maxItems":"7","unique":true,"items":{"cel":[{"id":"front_door_request_method_value","message":"each value must be one of: GET, HEAD, OPTIONS, TRACE, POST, PUT, DELETE","expression":"this in ['GET', 'HEAD', 'OPTIONS', 'TRACE', 'POST', 'PUT', 'DELETE']"}]}}}

### spec.rules[].conditions.queryString

`[]AzureFrontDoorRuleQueryStringCondition`

Match on the full query string (everything after '?').

- rule: match_values must be empty when the operator is ANY, and non-empty otherwise

### spec.rules[].conditions.queryString[].operator

`enum`

The comparison operator (the standard set; ANY matches whenever a
query string is present).

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.queryString[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.queryString[].matchValues

`[]string`

Up to 25 values, OR'd together. Empty strings are legal values.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.queryString[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.postArgs

`[]AzureFrontDoorRulePostArgsCondition`

Match on a named argument in a POST body (application/x-www-form-urlencoded).

- rule: match_values must be empty when the operator is ANY, and non-empty otherwise

### spec.rules[].conditions.postArgs[].postArgsName

`string` · required

The name of the POST argument to inspect.

- rule: {"required":true}

### spec.rules[].conditions.postArgs[].operator

`enum`

The comparison operator (the standard set).

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.postArgs[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.postArgs[].matchValues

`[]string`

Up to 25 values, OR'd together.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.postArgs[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.requestUri

`[]AzureFrontDoorRuleRequestUriCondition`

Match on the full request URI (path + query string).

- rule: match_values must be empty when the operator is ANY, and non-empty otherwise

### spec.rules[].conditions.requestUri[].operator

`enum`

The comparison operator (the standard set).

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.requestUri[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.requestUri[].matchValues

`[]string`

Up to 25 values, OR'd together.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.requestUri[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.requestHeader

`[]AzureFrontDoorRuleRequestHeaderCondition`

Match on the value of a named request header.

- rule: match_values must be empty when the operator is ANY, and non-empty otherwise

### spec.rules[].conditions.requestHeader[].headerName

`string` · required

The header to inspect (case-insensitive per HTTP semantics).

- rule: {"required":true}

### spec.rules[].conditions.requestHeader[].operator

`enum`

The comparison operator (the standard set; ANY matches whenever the
header is present).

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.requestHeader[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.requestHeader[].matchValues

`[]string`

Up to 25 values, OR'd together.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.requestHeader[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.requestBody

`[]AzureFrontDoorRuleRequestBodyCondition`

Match on the request body's content.

### spec.rules[].conditions.requestBody[].operator

`enum`

The comparison operator (the standard set minus ANY).

- rule: {"enum":{"definedOnly":true,"in":[2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.requestBody[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.requestBody[].matchValues

`[]string` · required

1-25 non-empty values, OR'd together.

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.rules[].conditions.requestBody[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.requestScheme

`[]AzureFrontDoorRuleRequestSchemeCondition`

Match on the request scheme (HTTP vs HTTPS).

### spec.rules[].conditions.requestScheme[].negateCondition

`bool`

Invert the condition -- match the OTHER scheme.

### spec.rules[].conditions.requestScheme[].matchValue

`string` · optional (explicit presence)

The scheme to match: "HTTP" or "HTTPS". Unset matches HTTP (Azure's
default) -- though stating it explicitly reads better.

- default: `HTTP`
- rule: match_value must be HTTP or HTTPS

### spec.rules[].conditions.urlPath

`[]AzureFrontDoorRuleUrlPathCondition`

Match on the URL path (everything between the host and the query
string). The only condition type that supports the WILDCARD
operator.

- rule: match_values must be empty when the operator is ANY, and non-empty otherwise

### spec.rules[].conditions.urlPath[].operator

`enum`

The comparison operator: the standard set plus WILDCARD.

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10,11]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.urlPath[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.urlPath[].matchValues

`[]string`

Up to 25 values, OR'd together. Write paths WITHOUT a leading '/'
(Azure matches against the path with the leading slash stripped,
e.g. "images/products" not "/images/products").

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.urlPath[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.urlFileExtension

`[]AzureFrontDoorRuleUrlFileExtensionCondition`

Match on the requested file's extension (the text after the last
'.' in the path -- write values WITHOUT the leading dot, e.g. "pdf").

### spec.rules[].conditions.urlFileExtension[].operator

`enum`

The comparison operator (the standard set minus ANY).

- rule: {"enum":{"definedOnly":true,"in":[2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.urlFileExtension[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.urlFileExtension[].matchValues

`[]string` · required

1-25 non-empty extensions WITHOUT the leading dot (e.g. "pdf",
"woff2"), OR'd together.

- rule: {"repeated":{"minItems":"1","maxItems":"25","items":{"string":{"minLen":"1"}}}}

### spec.rules[].conditions.urlFileExtension[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.urlFilename

`[]AzureFrontDoorRuleUrlFilenameCondition`

Match on the requested file's name (the last path segment).

- rule: match_values must be empty when the operator is ANY, and non-empty otherwise

### spec.rules[].conditions.urlFilename[].operator

`enum`

The comparison operator (the standard set).

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.urlFilename[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.urlFilename[].matchValues

`[]string`

Up to 25 values, OR'd together.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.urlFilename[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.httpVersion

`[]AzureFrontDoorRuleHttpVersionCondition`

Match on the HTTP protocol version the client negotiated.

### spec.rules[].conditions.httpVersion[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.httpVersion[].matchValues

`[]string` · required

The versions to match, 1-4 unique values from: "2.0", "1.1", "1.0",
"0.9".

- rule: {"repeated":{"minItems":"1","maxItems":"4","unique":true,"items":{"cel":[{"id":"front_door_http_version_value","message":"each value must be one of: 2.0, 1.1, 1.0, 0.9","expression":"this in ['2.0', '1.1', '1.0', '0.9']"}]}}}

### spec.rules[].conditions.cookies

`[]AzureFrontDoorRuleCookiesCondition`

Match on the value of a named cookie.

- rule: match_values must be empty when the operator is ANY, and non-empty otherwise

### spec.rules[].conditions.cookies[].cookieName

`string` · required

The cookie to inspect.

- rule: {"required":true}

### spec.rules[].conditions.cookies[].operator

`enum`

The comparison operator (the standard set; ANY matches whenever the
cookie is present).

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.cookies[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.cookies[].matchValues

`[]string`

Up to 25 values, OR'd together.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.cookies[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.isDevice

`[]AzureFrontDoorRuleIsDeviceCondition`

Match on the device class (Desktop vs Mobile) Azure derives from
the User-Agent.

### spec.rules[].conditions.isDevice[].negateCondition

`bool`

Invert the condition -- match the OTHER device class.

### spec.rules[].conditions.isDevice[].matchValue

`string` · required

The device class to match: "Desktop" or "Mobile".

- rule: match_value must be Desktop or Mobile
- rule: {"required":true}

### spec.rules[].conditions.socketAddress

`[]AzureFrontDoorRuleSocketAddressCondition`

Match on the socket address the connection actually arrived from
(the TCP peer). Differs from remote_address when a proxy sits in
front of Front Door: remote_address honors X-Forwarded-For; the
socket address is the proxy itself.

- rule: match_values must carry at least one CIDR range

### spec.rules[].conditions.socketAddress[].operator

`enum`

IP_MATCH (the default when unspecified) is the only comparison
Azure supports on socket addresses -- no geo-matching, no other
operators.

- rule: {"enum":{"definedOnly":true,"in":[0,13]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.socketAddress[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.socketAddress[].matchValues

`[]string`

1-25 IPv4/IPv6 CIDR ranges, OR'd together. Azure rejects
duplicate and overlapping CIDRs at apply time.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.clientPort

`[]AzureFrontDoorRuleClientPortCondition`

Match on the client's TCP port.

- rule: match_values must be empty when the operator is ANY, and non-empty otherwise

### spec.rules[].conditions.clientPort[].operator

`enum`

The comparison operator (the standard set).

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.clientPort[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.clientPort[].matchValues

`[]string`

Up to 25 port numbers (as strings), OR'd together.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.serverPort

`[]AzureFrontDoorRuleServerPortCondition`

Match on the port the request arrived on at Front Door (80 or 443).

### spec.rules[].conditions.serverPort[].operator

`enum`

The comparison operator (the standard set).

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.serverPort[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.serverPort[].matchValues

`[]string` · required

The ports to match, 1-2 unique values from: "80", "443" (the only
ports Front Door serves).

- rule: {"repeated":{"minItems":"1","maxItems":"2","unique":true,"items":{"cel":[{"id":"front_door_server_port_value","message":"each value must be 80 or 443 -- the only ports Front Door listens on","expression":"this in ['80', '443']"}]}}}

### spec.rules[].conditions.hostName

`[]AzureFrontDoorRuleHostNameCondition`

Match on the Host header (the hostname the client requested) --
how one route serving several custom domains applies per-domain
policy.

- rule: match_values must be empty when the operator is ANY, and non-empty otherwise

### spec.rules[].conditions.hostName[].operator

`enum`

The comparison operator (the standard set).

- rule: {"enum":{"definedOnly":true,"in":[1,2,3,4,5,6,7,8,9,10]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_operator_unspecified` -- Not specified -- invalid for most conditions; the address conditions treat it as IP_MATCH (their documented default).
- `ANY` -- Match any value -- the condition matches whenever the attribute is present at all. match_values must be empty with this operator.
- `EQUAL` -- Exact (case-sensitive) equality; combine with transforms like LOWERCASE for case-insensitive matching.
- `CONTAINS` -- The attribute contains the match value as a substring.
- `BEGINS_WITH` -- The attribute starts with the match value.
- `ENDS_WITH` -- The attribute ends with the match value.
- `GREATER_THAN` -- Numeric greater-than (values compared as numbers).
- `GREATER_THAN_OR_EQUAL` -- Numeric greater-than-or-equal.
- `LESS_THAN` -- Numeric less-than.
- `LESS_THAN_OR_EQUAL` -- Numeric less-than-or-equal.
- `REG_EX` -- Regular-expression match (RE2 syntax; Premium-tier profiles only -- Azure rejects RegEx on Standard at apply time).
- `WILDCARD` -- Wildcard match with '*' placeholders -- URL path conditions only.
- `GEO_MATCH` -- Geo-match: the client's country equals one of the two-letter ISO 3166-1 match values -- address conditions only.
- `IP_MATCH` -- IP match: the client's address falls inside one of the IPv4/IPv6 CIDR match values -- address conditions only.

### spec.rules[].conditions.hostName[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.hostName[].matchValues

`[]string`

Up to 25 hostnames, OR'd together.

- rule: {"repeated":{"maxItems":"25"}}

### spec.rules[].conditions.hostName[].transforms

`[]enum`

Normalizations applied before comparing, up to 4.

- rule: {"repeated":{"maxItems":"4","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_transform_unspecified` -- Not specified -- invalid; list concrete transforms.
- `LOWERCASE` -- Lowercase the attribute before comparing.
- `UPPERCASE` -- Uppercase the attribute before comparing.
- `TRIM` -- Trim leading/trailing whitespace before comparing.
- `URL_DECODE` -- URL-decode the attribute before comparing.
- `URL_ENCODE` -- URL-encode the attribute before comparing.
- `REMOVE_NULLS` -- Remove null bytes before comparing.

### spec.rules[].conditions.sslProtocol

`[]AzureFrontDoorRuleSslProtocolCondition`

Match on the TLS protocol version of the client connection.

### spec.rules[].conditions.sslProtocol[].negateCondition

`bool`

Invert the condition.

### spec.rules[].conditions.sslProtocol[].matchValues

`[]string` · required

The TLS versions to match, 1-3 unique values from: "TLSv1",
"TLSv1.1", "TLSv1.2".

- rule: {"repeated":{"minItems":"1","maxItems":"3","unique":true,"items":{"cel":[{"id":"front_door_ssl_protocol_value","message":"each value must be one of: TLSv1, TLSv1.1, TLSv1.2","expression":"this in ['TLSv1', 'TLSv1.1', 'TLSv1.2']"}]}}}

### spec.rules[].actions

`AzureFrontDoorRuleActions` · required

The actions applied when the conditions match. At least one; at
most five per rule (Azure's cap).

- rule: {"required":true}
- rule: a rule must carry at least one action -- a rule with conditions but no actions does nothing and Azure rejects it
- rule: a rule may carry at most 5 actions in total -- split the policy into additional rules if you need more
- rule: url_redirect and url_rewrite cannot both be set on one rule -- a redirect answers the client directly while a rewrite forwards to the origin, so the two contradict
- rule: url_redirect and route_configuration_override cannot both be set on one rule -- a redirected request is answered at the edge and never forwards, so there is nothing for the override to configure

### spec.rules[].actions.urlRedirect

`AzureFrontDoorRuleUrlRedirectAction`

Answer the request with an HTTP redirect instead of forwarding it
to the origin. Mutually exclusive with url_rewrite and with
route_configuration_override.

### spec.rules[].actions.urlRedirect.redirectType

`enum`

The redirect's HTTP status: MOVED (301), FOUND (302),
TEMPORARY_REDIRECT (307), or PERMANENT_REDIRECT (308). The 30x
codes preserve the request method; 301/302 may downgrade POST to
GET in older clients.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_redirect_type_unspecified` -- Not specified -- invalid; pick a concrete redirect status.
- `MOVED` -- 301 Moved Permanently -- cached by clients; may downgrade POST to GET in older clients.
- `FOUND` -- 302 Found -- temporary; may downgrade POST to GET in older clients.
- `TEMPORARY_REDIRECT` -- 307 Temporary Redirect -- preserves the request method.
- `PERMANENT_REDIRECT` -- 308 Permanent Redirect -- preserves the request method; cached.

### spec.rules[].actions.urlRedirect.redirectProtocol

`enum`

The scheme of the redirect target: MATCH_REQUEST (default,
preserve the incoming scheme), HTTP_ONLY, or HTTPS_ONLY. An
HTTPS_ONLY redirect with everything else preserved is the classic
HTTP-to-HTTPS upgrade.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_forwarding_protocol_unspecified` -- Not specified -- deploys MATCH_REQUEST, Azure's default, for redirects; for the route-configuration override the field is only sent when a value is chosen.
- `MATCH_REQUEST` -- Mirror the incoming request's protocol.
- `HTTP_ONLY` -- Always HTTP.
- `HTTPS_ONLY` -- Always HTTPS.

### spec.rules[].actions.urlRedirect.destinationHostname

`string`

The hostname to redirect to, up to 2048 characters. Leave empty to
preserve the incoming host (e.g. for a scheme- or path-only
redirect).

- rule: {"string":{"maxLen":"2048"}}

### spec.rules[].actions.urlRedirect.destinationPath

`string`

The path to redirect to; must start with '/'. Leave empty to
preserve the incoming path.

- rule: destination_path must begin with '/' -- leave it empty to preserve the incoming path

### spec.rules[].actions.urlRedirect.queryString

`string`

The query string of the redirect target in key=value&key2=value2
form, WITHOUT the leading '?' (Front Door adds it), up to 2048
characters. Leave empty to preserve the incoming query string.

- rule: query_string must not start with '?' -- Front Door adds the separator automatically
- rule: {"string":{"maxLen":"2048"}}

### spec.rules[].actions.urlRedirect.destinationFragment

`string`

The fragment (the part after '#') of the redirect target, without
the '#', up to 1024 characters. Leave empty for no fragment.

- rule: {"string":{"maxLen":"1024"}}

### spec.rules[].actions.urlRewrite

`AzureFrontDoorRuleUrlRewriteAction`

Rewrite the request path before forwarding it to the origin (the
client never sees the rewritten path). Mutually exclusive with
url_redirect.

### spec.rules[].actions.urlRewrite.sourcePattern

`string` · required

The path prefix to match and replace, e.g. "/v1". Use "/" to
replace the whole path.

- rule: {"required":true}

### spec.rules[].actions.urlRewrite.destination

`string` · required

What the matched prefix is replaced with, e.g. "/api/v1".

- rule: {"required":true}

### spec.rules[].actions.urlRewrite.preserveUnmatchedPath

`bool`

Keep the path remainder after the matched prefix and append it to
the destination (true), or forward exactly the destination
(false, Azure's default). With source "/v1", destination "/api"
and preserve on, "/v1/users" becomes "/api/users".

### spec.rules[].actions.requestHeaders

`[]AzureFrontDoorRuleHeaderAction`

Add, overwrite, or delete headers on the request Front Door
forwards to the origin.

- rule: value is required when header_action is APPEND or OVERWRITE, and must be empty when it is DELETE

### spec.rules[].actions.requestHeaders[].headerAction

`enum`

APPEND adds the header (joining an existing value), OVERWRITE
replaces it, DELETE removes it.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_header_action_type_unspecified` -- Not specified -- invalid; pick APPEND, OVERWRITE, or DELETE.
- `APPEND` -- Add the header, joining any existing value.
- `OVERWRITE` -- Replace the header's value (or add it if absent).
- `DELETE` -- Remove the header.

### spec.rules[].actions.requestHeaders[].headerName

`string` · required

The header to act on.

- rule: {"required":true}

### spec.rules[].actions.requestHeaders[].value

`string`

The value to append or overwrite with. Required for APPEND and
OVERWRITE; must be empty for DELETE (there is nothing to set when
removing a header).

### spec.rules[].actions.responseHeaders

`[]AzureFrontDoorRuleHeaderAction`

Add, overwrite, or delete headers on the response Front Door
returns to the client (e.g. security headers).

- rule: value is required when header_action is APPEND or OVERWRITE, and must be empty when it is DELETE

### spec.rules[].actions.responseHeaders[].headerAction

`enum`

APPEND adds the header (joining an existing value), OVERWRITE
replaces it, DELETE removes it.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_header_action_type_unspecified` -- Not specified -- invalid; pick APPEND, OVERWRITE, or DELETE.
- `APPEND` -- Add the header, joining any existing value.
- `OVERWRITE` -- Replace the header's value (or add it if absent).
- `DELETE` -- Remove the header.

### spec.rules[].actions.responseHeaders[].headerName

`string` · required

The header to act on.

- rule: {"required":true}

### spec.rules[].actions.responseHeaders[].value

`string`

The value to append or overwrite with. Required for APPEND and
OVERWRITE; must be empty for DELETE (there is nothing to set when
removing a header).

### spec.rules[].actions.routeConfigurationOverride

`AzureFrontDoorRuleRouteConfigurationOverrideAction`

Override the route's own forwarding and caching configuration for
requests this rule matches -- e.g. send matched traffic to a
different origin group, or disable caching for authenticated paths.
Mutually exclusive with url_redirect: a redirected request is
answered at the edge and never forwards, so there is nothing for
the override to configure.

- rule: forwarding_protocol must be set when origin_group_id is set (it qualifies the overriding origin), and must be left unspecified when there is no origin override
- rule: with cache_behavior DISABLED nothing caches, so cache_duration, query_string_caching_behavior, and query_string_parameters must stay unset
- rule: query_string_caching_behavior is required when cache_behavior enables caching (HONOR_ORIGIN or the OVERRIDE_* behaviors)
- rule: cache_duration is required with OVERRIDE_ALWAYS and OVERRIDE_IF_ORIGIN_MISSING, and must stay unset with HONOR_ORIGIN (the origin's Cache-Control decides)
- rule: query_string_parameters is required with the *_SPECIFIED caching behaviors and must stay unset with IGNORE_QUERY_STRING and USE_QUERY_STRING

### spec.rules[].actions.routeConfigurationOverride.originGroupId

`string | valueFrom`

Send matched requests to a different origin group than the route's
own, by ARM ID. References an AzureFrontDoorOriginGroup's
origin_group_id output. When set, forwarding_protocol must be
chosen too; when unset, the route's own origin group keeps serving.
The rule set lives in its profile and steers TO the origin group, so
the reference is access, not placement, on a diagram.

- references: AzureFrontDoorOriginGroup (`status.outputs.origin_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorOriginGroup, name: <that resource's name>, fieldPath: status.outputs.origin_group_id}} -- a bare string does not parse

### spec.rules[].actions.routeConfigurationOverride.forwardingProtocol

`enum`

The protocol toward the overriding origin group. Required when
origin_group_id is set; must be left unspecified otherwise (it
qualifies the override's origin, not the route's).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_forwarding_protocol_unspecified` -- Not specified -- deploys MATCH_REQUEST, Azure's default, for redirects; for the route-configuration override the field is only sent when a value is chosen.
- `MATCH_REQUEST` -- Mirror the incoming request's protocol.
- `HTTP_ONLY` -- Always HTTP.
- `HTTPS_ONLY` -- Always HTTPS.

### spec.rules[].actions.routeConfigurationOverride.cacheBehavior

`enum`

How matched responses cache at the edge: HONOR_ORIGIN follows the
origin's Cache-Control, the OVERRIDE_* behaviors force the
cache_duration, and DISABLED turns caching off for matched
requests. Required -- every override makes an explicit cache
decision (DISABLED is the "no caching" choice; there is no
leave-the-route-alone value on this action).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_cache_behavior_unspecified` -- Not specified -- invalid; every override makes an explicit cache decision (use DISABLED to turn caching off for matched requests).
- `HONOR_ORIGIN` -- Follow the origin's Cache-Control headers.
- `OVERRIDE_ALWAYS` -- Cache for cache_duration regardless of the origin's headers.
- `OVERRIDE_IF_ORIGIN_MISSING` -- Cache for cache_duration only when the origin sends no caching headers of its own.
- `DISABLED` -- Turn caching off for matched requests.

### spec.rules[].actions.routeConfigurationOverride.cacheDuration

`string`

How long forced caching lasts, in "d.HH:MM:SS" (days 1-365) or
"HH:MM:SS" format, e.g. "1.12:00:00" or "00:05:00". Required with
the OVERRIDE_* behaviors; forbidden with HONOR_ORIGIN (the origin
decides) and DISABLED (nothing caches).

- rule: cache_duration must be d.HH:MM:SS (days 1-365) or HH:MM:SS, e.g. "1.12:00:00" or "00:05:00"

### spec.rules[].actions.routeConfigurationOverride.queryStringCachingBehavior

`enum`

How query strings participate in the cache key for matched
requests. Required whenever cache_behavior enables caching
(HONOR_ORIGIN or OVERRIDE_*); forbidden with DISABLED.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_front_door_rule_query_string_caching_behavior_unspecified` -- Not specified -- only legal when cache_behavior is unspecified or DISABLED.
- `IGNORE_QUERY_STRING` -- Strip query strings from the cache key -- every variant shares one entry.
- `USE_QUERY_STRING` -- Key the cache on the full query string -- each variant caches separately.
- `IGNORE_SPECIFIED_QUERY_STRINGS` -- Ignore ONLY the parameters named in query_string_parameters.
- `INCLUDE_SPECIFIED_QUERY_STRINGS` -- Include ONLY the parameters named in query_string_parameters.

### spec.rules[].actions.routeConfigurationOverride.queryStringParameters

`[]string`

The query-string parameter NAMES the *_SPECIFIED caching behaviors
operate on, up to 100. Required with those two behaviors; forbidden
with the other two.

- rule: {"repeated":{"maxItems":"100"}}

### spec.rules[].actions.routeConfigurationOverride.compressionEnabled

`bool` · optional (explicit presence)

Compress eligible matched responses at the edge. Only meaningful
when cache_behavior enables caching (compression is part of the
cache configuration); with DISABLED it is ignored. Default false.

- default: `false`

## Validation Rules

- `front_door_rule_set_rule_names_unique`: every rule in the set must have a unique name -- each rule is its own Azure resource keyed by name, so two rules with the same name would silently target one resource

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureFrontDoorRuleSet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.rule_set_id` | `string` | The Azure Resource Manager ID of the rule set -- what AzureFrontDoorRoute's rule_set_ids references to attach this delivery policy. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Cdn/profiles/{profile}/ruleSets/{name} |
| `status.outputs.rule_set_name` | `string` | The rule set's name -- unique within its profile. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.profileId` | AzureFrontDoorProfile | `status.outputs.profile_id` |
| `spec.rules[].actions.routeConfigurationOverride.originGroupId` | AzureFrontDoorOriginGroup | `status.outputs.origin_group_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureFrontDoorRoute | `spec.ruleSetIds` | `status.outputs.rule_set_id` |

## See Also

- [Overview](../README.md)
