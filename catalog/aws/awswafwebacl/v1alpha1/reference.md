# AwsWafWebAcl

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsWafWebAclSpec defines the desired configuration for an AWS WAFv2 Web ACL.

A Web ACL (Web Access Control List) is the top-level WAFv2 resource that
protects AWS applications from web exploits, bots, and abuse. It contains an
ordered set of rules that inspect incoming web requests and take actions
(allow, block, count, CAPTCHA, challenge) based on matching conditions.

Web ACLs protect:
- Application Load Balancers, API Gateway REST APIs, AppSync GraphQL APIs,
  Cognito user pools, App Runner services, Verified Access instances
  (REGIONAL scope — associated from the protected resource's side, e.g.
  AwsAlb.spec.web_acl_arn)
- CloudFront distributions (CLOUDFRONT scope — must be created in
  us-east-1; referenced from AwsCloudFront.spec.web_acl_arn)

Rules are evaluated in priority order (lowest number first). When a rule
matches, its action is taken and evaluation stops (count/CAPTCHA-passed
requests continue to later rules). If no rule matches, the default_action
applies.

The full WAFv2 statement language is modeled as a typed, recursive tree
(AwsWafWebAclStatement): managed rule groups (with ATP/ACFP/Bot
Control/anti-DDoS configs), rate limiting (including custom aggregation
keys), IP set / regex pattern set / rule group references, geo, byte /
SQLi / XSS / size / regex / label / ASN matching, and AND/OR/NOT
composition. A raw-JSON custom_statement escape hatch remains for anything
AWS ships before this spec models it.

Every rule consumes Web ACL Capacity Units (WCUs); the default account
limit is 5,000 WCUs per web ACL (the deployed total is exported as the
`capacity` stack output).

Associations are NOT bundled: a web ACL protects at most one scope's worth
of resources, but which resources it protects is each resource's own
setting — the protected resource (ALB, CloudFront distribution) references
this web ACL's ARN output, never the other way around.

Logging IS bundled (the logging field) because a web ACL has at most one
logging configuration and it shares the ACL's lifecycle.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsWafWebAcl
metadata:
  name: test-waf-acl
  org: acme
  env: dev
spec:
  region: us-west-2
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: aws-managed-common
      priority: 1
      overrideAction: none
      statement:
        managedRuleGroup:
          name: AWSManagedRulesCommonRuleSet
          vendorName: AWS
          scopeDownStatement:
            notStatement:
              statement:
                byteMatch:
                  searchString: /health
                  positionalConstraint: STARTS_WITH
                  fieldToMatch:
                    uriPath: true
                  textTransformations:
                    - priority: 0
                      type: NONE
    - name: rate-limit-per-api-key
      priority: 2
      action: block
      statement:
        rateBased:
          limit: 2000
          aggregateKeyType: CUSTOM_KEYS
          customKeys:
            - header:
                name: x-api-key
                textTransformations:
                  - priority: 0
                    type: NONE
            - ip: true
    - name: block-bad-paths
      priority: 3
      action: block
      customResponse:
        responseCode: 403
      statement:
        orStatement:
          statements:
            - byteMatch:
                searchString: /wp-admin
                positionalConstraint: STARTS_WITH
                fieldToMatch:
                  uriPath: true
                textTransformations:
                  - priority: 0
                    type: LOWERCASE
            - sqliMatch:
                fieldToMatch:
                  body:
                    oversizeHandling: MATCH
                textTransformations:
                  - priority: 0
                    type: URL_DECODE
  # Request logging with the full redaction surface and a keep-only-enforcement
  # filter: BLOCK and COUNT records are kept, everything else is dropped.
  logging:
    destinationArn:
      value: arn:aws:logs:us-west-2:111122223333:log-group:aws-waf-logs-test-waf-acl
    redactedHeaderNames:
      - authorization
      - cookie
    redactUriPath: true
    redactQueryString: true
    redactMethod: true
    filter:
      defaultBehavior: DROP
      filters:
        - behavior: KEEP
          requirement: MEETS_ANY
          conditions:
            - action: BLOCK
            - action: COUNT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.scope` | `string` | yes |  |  |
| `spec.defaultAction` | `AwsWafWebAclDefaultAction` | yes |  |  |
| `spec.defaultAction.type` | `string` | yes |  |  |
| `spec.defaultAction.customResponse` | `AwsWafWebAclCustomResponse` |  |  |  |
| `spec.defaultAction.customResponse.responseCode` | `int32` | yes |  |  |
| `spec.defaultAction.customResponse.customResponseBodyKey` | `string` |  |  |  |
| `spec.defaultAction.customResponse.responseHeaders` | `[]AwsWafWebAclCustomHeader` |  |  |  |
| `spec.defaultAction.customResponse.responseHeaders[].name` | `string` | yes |  |  |
| `spec.defaultAction.customResponse.responseHeaders[].value` | `string` | yes |  |  |
| `spec.defaultAction.customRequestHeaders` | `[]AwsWafWebAclCustomHeader` |  |  |  |
| `spec.defaultAction.customRequestHeaders[].name` | `string` | yes |  |  |
| `spec.defaultAction.customRequestHeaders[].value` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.rules` | `[]AwsWafWebAclRule` |  |  |  |
| `spec.rules[].name` | `string` | yes |  |  |
| `spec.rules[].priority` | `int32` | yes |  |  |
| `spec.rules[].statement` | `AwsWafWebAclStatement` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup` | `AwsWafWebAclManagedRuleGroupStatement` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.name` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.vendorName` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.version` | `string` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.ruleActionOverrides` | `[]AwsWafWebAclRuleActionOverride` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.ruleActionOverrides[].name` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.ruleActionOverrides[].action` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.scopeDownStatement` | `AwsWafWebAclStatement` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs` | `AwsWafWebAclManagedRuleGroupConfigs` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.botControl` | `AwsWafWebAclBotControlConfig` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.botControl.inspectionLevel` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.botControl.enableMachineLearning` | `bool` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention` | `AwsWafWebAclAtpConfig` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.loginPath` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.enableRegexInPath` | `bool` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection` | `AwsWafWebAclAtpRequestInspection` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.payloadType` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.usernameField` | `AwsWafWebAclFieldIdentifier` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.usernameField.identifier` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.passwordField` | `AwsWafWebAclFieldIdentifier` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.passwordField.identifier` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection` | `AwsWafWebAclResponseInspection` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.statusCode` | `AwsWafWebAclResponseInspectionStatusCode` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.statusCode.successCodes` | `[]int32` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.statusCode.failureCodes` | `[]int32` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.header` | `AwsWafWebAclResponseInspectionHeader` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.header.name` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.header.successValues` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.header.failureValues` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyJson` | `AwsWafWebAclResponseInspectionJson` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyJson.identifier` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyJson.successValues` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyJson.failureValues` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyContains` | `AwsWafWebAclResponseInspectionBodyContains` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyContains.successStrings` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyContains.failureStrings` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention` | `AwsWafWebAclAcfpConfig` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.creationPath` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.registrationPagePath` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.enableRegexInPath` | `bool` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection` | `AwsWafWebAclAcfpRequestInspection` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.payloadType` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.usernameField` | `AwsWafWebAclFieldIdentifier` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.usernameField.identifier` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.passwordField` | `AwsWafWebAclFieldIdentifier` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.passwordField.identifier` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.emailField` | `AwsWafWebAclFieldIdentifier` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.emailField.identifier` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.phoneNumberFields` | `AwsWafWebAclFieldIdentifiers` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.phoneNumberFields.identifiers` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.addressFields` | `AwsWafWebAclFieldIdentifiers` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.addressFields.identifiers` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection` | `AwsWafWebAclResponseInspection` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.statusCode` | `AwsWafWebAclResponseInspectionStatusCode` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.statusCode.successCodes` | `[]int32` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.statusCode.failureCodes` | `[]int32` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.header` | `AwsWafWebAclResponseInspectionHeader` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.header.name` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.header.successValues` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.header.failureValues` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyJson` | `AwsWafWebAclResponseInspectionJson` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyJson.identifier` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyJson.successValues` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyJson.failureValues` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyContains` | `AwsWafWebAclResponseInspectionBodyContains` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyContains.successStrings` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyContains.failureStrings` | `[]string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos` | `AwsWafWebAclAntiDdosConfig` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.clientSideAction` | `AwsWafWebAclAntiDdosClientSideAction` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.clientSideAction.usageOfAction` | `string` | yes |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.clientSideAction.sensitivity` | `string` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.clientSideAction.exemptUriRegularExpressions` | `[]string` |  |  |  |
| `spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.sensitivityToBlock` | `string` |  |  |  |
| `spec.rules[].statement.rateBased` | `AwsWafWebAclRateBasedStatement` |  |  |  |
| `spec.rules[].statement.rateBased.limit` | `int32` | yes |  |  |
| `spec.rules[].statement.rateBased.evaluationWindowSec` | `int32` |  |  |  |
| `spec.rules[].statement.rateBased.aggregateKeyType` | `string` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys` | `[]AwsWafWebAclRateBasedCustomKey` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].header` | `AwsWafWebAclKeyWithTransformations` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].header.name` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].header.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].header.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].header.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].cookie` | `AwsWafWebAclKeyWithTransformations` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].cookie.name` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].cookie.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].cookie.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].cookie.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].queryArgument` | `AwsWafWebAclKeyWithTransformations` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].queryArgument.name` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].queryArgument.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].queryArgument.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].queryArgument.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].queryString` | `AwsWafWebAclTransformationsOnlyKey` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].queryString.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].queryString.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].queryString.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].uriPath` | `AwsWafWebAclTransformationsOnlyKey` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].uriPath.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].uriPath.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].uriPath.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].httpMethod` | `bool` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].ip` | `bool` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].forwardedIp` | `bool` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].asn` | `bool` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].labelNamespace` | `AwsWafWebAclLabelNamespaceKey` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].labelNamespace.namespace` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].ja3Fingerprint` | `AwsWafWebAclFingerprintKey` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].ja3Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.customKeys[].ja4Fingerprint` | `AwsWafWebAclFingerprintKey` |  |  |  |
| `spec.rules[].statement.rateBased.customKeys[].ja4Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.forwardedIpConfig` | `AwsWafWebAclForwardedIpConfig` |  |  |  |
| `spec.rules[].statement.rateBased.forwardedIpConfig.headerName` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.forwardedIpConfig.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.rateBased.forwardedIpConfig.position` | `string` |  |  |  |
| `spec.rules[].statement.rateBased.scopeDownStatement` | `AwsWafWebAclStatement` |  |  |  |
| `spec.rules[].statement.ruleGroupReference` | `AwsWafWebAclRuleGroupReferenceStatement` |  |  |  |
| `spec.rules[].statement.ruleGroupReference.arn` | `string` | yes |  |  |
| `spec.rules[].statement.ruleGroupReference.ruleActionOverrides` | `[]AwsWafWebAclRuleActionOverride` |  |  |  |
| `spec.rules[].statement.ruleGroupReference.ruleActionOverrides[].name` | `string` | yes |  |  |
| `spec.rules[].statement.ruleGroupReference.ruleActionOverrides[].action` | `string` | yes |  |  |
| `spec.rules[].statement.ipSetReference` | `AwsWafWebAclIpSetReferenceStatement` |  |  |  |
| `spec.rules[].statement.ipSetReference.arn` | `string \| valueFrom` | yes |  | AwsWafIpSet (`status.outputs.ip_set_arn`) |
| `spec.rules[].statement.ipSetReference.forwardedIpConfig` | `AwsWafWebAclForwardedIpConfig` |  |  |  |
| `spec.rules[].statement.ipSetReference.forwardedIpConfig.headerName` | `string` | yes |  |  |
| `spec.rules[].statement.ipSetReference.forwardedIpConfig.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.ipSetReference.forwardedIpConfig.position` | `string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference` | `AwsWafWebAclRegexPatternSetReferenceStatement` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.arn` | `string \| valueFrom` | yes |  | AwsWafRegexPatternSet (`status.outputs.regex_pattern_set_arn`) |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch` | `AwsWafWebAclFieldToMatch` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.uriPath` | `bool` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.queryString` | `bool` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.method` | `bool` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.allQueryArguments` | `bool` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.singleHeader` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.singleHeader.name` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.singleQueryArgument` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.singleQueryArgument.name` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.body` | `AwsWafWebAclBodyMatch` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.body.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody` | `AwsWafWebAclJsonBodyMatch` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody.includedPaths` | `[]string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody.invalidFallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers` | `AwsWafWebAclHeadersMatch` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies` | `AwsWafWebAclCookiesMatch` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.headerOrder` | `AwsWafWebAclHeaderOrderMatch` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.headerOrder.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.ja3Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.ja3Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.ja4Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.ja4Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.uriFragment` | `AwsWafWebAclUriFragmentMatch` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.fieldToMatch.uriFragment.fallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.regexPatternSetReference.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.regexPatternSetReference.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.geoMatch` | `AwsWafWebAclGeoMatchStatement` |  |  |  |
| `spec.rules[].statement.geoMatch.countryCodes` | `[]string` | yes |  |  |
| `spec.rules[].statement.geoMatch.forwardedIpConfig` | `AwsWafWebAclForwardedIpConfig` |  |  |  |
| `spec.rules[].statement.geoMatch.forwardedIpConfig.headerName` | `string` | yes |  |  |
| `spec.rules[].statement.geoMatch.forwardedIpConfig.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.geoMatch.forwardedIpConfig.position` | `string` |  |  |  |
| `spec.rules[].statement.byteMatch` | `AwsWafWebAclByteMatchStatement` |  |  |  |
| `spec.rules[].statement.byteMatch.searchString` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.positionalConstraint` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch` | `AwsWafWebAclFieldToMatch` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.uriPath` | `bool` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.queryString` | `bool` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.method` | `bool` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.allQueryArguments` | `bool` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.singleHeader` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.singleHeader.name` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.singleQueryArgument` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.singleQueryArgument.name` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.body` | `AwsWafWebAclBodyMatch` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.body.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.jsonBody` | `AwsWafWebAclJsonBodyMatch` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.jsonBody.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.jsonBody.includedPaths` | `[]string` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.jsonBody.invalidFallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.jsonBody.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.headers` | `AwsWafWebAclHeadersMatch` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.headers.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.headers.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.headers.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.headers.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.headers.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.headers.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.cookies` | `AwsWafWebAclCookiesMatch` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.cookies.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.headerOrder` | `AwsWafWebAclHeaderOrderMatch` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.headerOrder.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.ja3Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.ja3Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.ja4Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.ja4Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.uriFragment` | `AwsWafWebAclUriFragmentMatch` |  |  |  |
| `spec.rules[].statement.byteMatch.fieldToMatch.uriFragment.fallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.byteMatch.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.byteMatch.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.byteMatch.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch` | `AwsWafWebAclSqliMatchStatement` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch` | `AwsWafWebAclFieldToMatch` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.uriPath` | `bool` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.queryString` | `bool` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.method` | `bool` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.allQueryArguments` | `bool` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.singleHeader` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.singleHeader.name` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.singleQueryArgument` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.singleQueryArgument.name` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.body` | `AwsWafWebAclBodyMatch` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.body.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody` | `AwsWafWebAclJsonBodyMatch` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody.includedPaths` | `[]string` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody.invalidFallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.headers` | `AwsWafWebAclHeadersMatch` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.headers.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.cookies` | `AwsWafWebAclCookiesMatch` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.cookies.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.headerOrder` | `AwsWafWebAclHeaderOrderMatch` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.headerOrder.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.ja3Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.ja3Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.ja4Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.ja4Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.uriFragment` | `AwsWafWebAclUriFragmentMatch` |  |  |  |
| `spec.rules[].statement.sqliMatch.fieldToMatch.uriFragment.fallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.sqliMatch.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.sqliMatch.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.sqliMatch.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.sqliMatch.sensitivityLevel` | `string` |  |  |  |
| `spec.rules[].statement.xssMatch` | `AwsWafWebAclXssMatchStatement` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch` | `AwsWafWebAclFieldToMatch` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.uriPath` | `bool` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.queryString` | `bool` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.method` | `bool` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.allQueryArguments` | `bool` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.singleHeader` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.singleHeader.name` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.singleQueryArgument` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.singleQueryArgument.name` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.body` | `AwsWafWebAclBodyMatch` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.body.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.jsonBody` | `AwsWafWebAclJsonBodyMatch` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.jsonBody.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.jsonBody.includedPaths` | `[]string` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.jsonBody.invalidFallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.jsonBody.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.headers` | `AwsWafWebAclHeadersMatch` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.headers.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.headers.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.headers.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.headers.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.headers.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.headers.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.cookies` | `AwsWafWebAclCookiesMatch` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.cookies.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.headerOrder` | `AwsWafWebAclHeaderOrderMatch` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.headerOrder.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.ja3Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.ja3Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.ja4Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.ja4Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.uriFragment` | `AwsWafWebAclUriFragmentMatch` |  |  |  |
| `spec.rules[].statement.xssMatch.fieldToMatch.uriFragment.fallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.xssMatch.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.xssMatch.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.xssMatch.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint` | `AwsWafWebAclSizeConstraintStatement` |  |  |  |
| `spec.rules[].statement.sizeConstraint.comparisonOperator` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.size` | `int32` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch` | `AwsWafWebAclFieldToMatch` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.uriPath` | `bool` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.queryString` | `bool` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.method` | `bool` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.allQueryArguments` | `bool` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.singleHeader` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.singleHeader.name` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.singleQueryArgument` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.singleQueryArgument.name` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.body` | `AwsWafWebAclBodyMatch` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.body.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody` | `AwsWafWebAclJsonBodyMatch` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody.includedPaths` | `[]string` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody.invalidFallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.headers` | `AwsWafWebAclHeadersMatch` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.headers.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.cookies` | `AwsWafWebAclCookiesMatch` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.headerOrder` | `AwsWafWebAclHeaderOrderMatch` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.headerOrder.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.ja3Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.ja3Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.ja4Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.ja4Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.uriFragment` | `AwsWafWebAclUriFragmentMatch` |  |  |  |
| `spec.rules[].statement.sizeConstraint.fieldToMatch.uriFragment.fallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.sizeConstraint.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.sizeConstraint.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.sizeConstraint.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch` | `AwsWafWebAclRegexMatchStatement` |  |  |  |
| `spec.rules[].statement.regexMatch.regexString` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch` | `AwsWafWebAclFieldToMatch` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.uriPath` | `bool` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.queryString` | `bool` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.method` | `bool` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.allQueryArguments` | `bool` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.singleHeader` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.singleHeader.name` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.singleQueryArgument` | `AwsWafWebAclSingleField` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.singleQueryArgument.name` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.body` | `AwsWafWebAclBodyMatch` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.body.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.jsonBody` | `AwsWafWebAclJsonBodyMatch` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.jsonBody.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.jsonBody.includedPaths` | `[]string` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.jsonBody.invalidFallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.jsonBody.oversizeHandling` | `string` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.headers` | `AwsWafWebAclHeadersMatch` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.headers.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.headers.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.headers.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.headers.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.headers.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.headers.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.cookies` | `AwsWafWebAclCookiesMatch` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchPattern` | `AwsWafWebAclNamePattern` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchPattern.all` | `bool` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchPattern.includedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchPattern.excludedNames` | `[]string` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchScope` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.cookies.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.headerOrder` | `AwsWafWebAclHeaderOrderMatch` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.headerOrder.oversizeHandling` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.ja3Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.ja3Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.ja4Fingerprint` | `AwsWafWebAclFingerprintMatch` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.ja4Fingerprint.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.uriFragment` | `AwsWafWebAclUriFragmentMatch` |  |  |  |
| `spec.rules[].statement.regexMatch.fieldToMatch.uriFragment.fallbackBehavior` | `string` |  |  |  |
| `spec.rules[].statement.regexMatch.textTransformations` | `[]AwsWafWebAclTextTransformation` | yes |  |  |
| `spec.rules[].statement.regexMatch.textTransformations[].priority` | `int32` |  |  |  |
| `spec.rules[].statement.regexMatch.textTransformations[].type` | `string` | yes |  |  |
| `spec.rules[].statement.labelMatch` | `AwsWafWebAclLabelMatchStatement` |  |  |  |
| `spec.rules[].statement.labelMatch.scope` | `string` | yes |  |  |
| `spec.rules[].statement.labelMatch.key` | `string` | yes |  |  |
| `spec.rules[].statement.asnMatch` | `AwsWafWebAclAsnMatchStatement` |  |  |  |
| `spec.rules[].statement.asnMatch.asnList` | `[]uint32` | yes |  |  |
| `spec.rules[].statement.asnMatch.forwardedIpConfig` | `AwsWafWebAclForwardedIpConfig` |  |  |  |
| `spec.rules[].statement.asnMatch.forwardedIpConfig.headerName` | `string` | yes |  |  |
| `spec.rules[].statement.asnMatch.forwardedIpConfig.fallbackBehavior` | `string` | yes |  |  |
| `spec.rules[].statement.asnMatch.forwardedIpConfig.position` | `string` |  |  |  |
| `spec.rules[].statement.andStatement` | `AwsWafWebAclAndStatement` |  |  |  |
| `spec.rules[].statement.andStatement.statements` | `[]AwsWafWebAclStatement` | yes |  |  |
| `spec.rules[].statement.orStatement` | `AwsWafWebAclOrStatement` |  |  |  |
| `spec.rules[].statement.orStatement.statements` | `[]AwsWafWebAclStatement` | yes |  |  |
| `spec.rules[].statement.notStatement` | `AwsWafWebAclNotStatement` |  |  |  |
| `spec.rules[].statement.notStatement.statement` | `AwsWafWebAclStatement` | yes |  |  |
| `spec.rules[].statement.customStatement` | `object` |  |  |  |
| `spec.rules[].action` | `string` |  |  |  |
| `spec.rules[].overrideAction` | `string` |  |  |  |
| `spec.rules[].customResponse` | `AwsWafWebAclCustomResponse` |  |  |  |
| `spec.rules[].customResponse.responseCode` | `int32` | yes |  |  |
| `spec.rules[].customResponse.customResponseBodyKey` | `string` |  |  |  |
| `spec.rules[].customResponse.responseHeaders` | `[]AwsWafWebAclCustomHeader` |  |  |  |
| `spec.rules[].customResponse.responseHeaders[].name` | `string` | yes |  |  |
| `spec.rules[].customResponse.responseHeaders[].value` | `string` | yes |  |  |
| `spec.rules[].customRequestHeaders` | `[]AwsWafWebAclCustomHeader` |  |  |  |
| `spec.rules[].customRequestHeaders[].name` | `string` | yes |  |  |
| `spec.rules[].customRequestHeaders[].value` | `string` | yes |  |  |
| `spec.rules[].ruleLabels` | `[]string` |  |  |  |
| `spec.rules[].visibilityConfig` | `AwsWafWebAclVisibilityConfig` |  |  |  |
| `spec.rules[].visibilityConfig.cloudwatchMetricsEnabled` | `bool` |  |  |  |
| `spec.rules[].visibilityConfig.sampledRequestsEnabled` | `bool` |  |  |  |
| `spec.rules[].visibilityConfig.metricName` | `string` |  |  |  |
| `spec.rules[].captchaConfig` | `AwsWafWebAclImmunityTimeConfig` |  |  |  |
| `spec.rules[].captchaConfig.immunityTimeSec` | `int32` | yes |  |  |
| `spec.rules[].challengeConfig` | `AwsWafWebAclImmunityTimeConfig` |  |  |  |
| `spec.rules[].challengeConfig.immunityTimeSec` | `int32` | yes |  |  |
| `spec.visibilityConfig` | `AwsWafWebAclVisibilityConfig` |  |  |  |
| `spec.visibilityConfig.cloudwatchMetricsEnabled` | `bool` |  |  |  |
| `spec.visibilityConfig.sampledRequestsEnabled` | `bool` |  |  |  |
| `spec.visibilityConfig.metricName` | `string` |  |  |  |
| `spec.customResponseBodies` | `[]AwsWafWebAclCustomResponseBody` |  |  |  |
| `spec.customResponseBodies[].key` | `string` | yes |  |  |
| `spec.customResponseBodies[].content` | `string` | yes |  |  |
| `spec.customResponseBodies[].contentType` | `string` | yes |  |  |
| `spec.tokenDomains` | `[]string` |  |  |  |
| `spec.captchaConfig` | `AwsWafWebAclImmunityTimeConfig` |  |  |  |
| `spec.captchaConfig.immunityTimeSec` | `int32` | yes |  |  |
| `spec.challengeConfig` | `AwsWafWebAclImmunityTimeConfig` |  |  |  |
| `spec.challengeConfig.immunityTimeSec` | `int32` | yes |  |  |
| `spec.associationConfig` | `AwsWafWebAclAssociationConfig` |  |  |  |
| `spec.associationConfig.cloudfrontRequestBodyLimit` | `string` |  |  |  |
| `spec.associationConfig.apiGatewayRequestBodyLimit` | `string` |  |  |  |
| `spec.associationConfig.cognitoUserPoolRequestBodyLimit` | `string` |  |  |  |
| `spec.associationConfig.appRunnerServiceRequestBodyLimit` | `string` |  |  |  |
| `spec.associationConfig.verifiedAccessInstanceRequestBodyLimit` | `string` |  |  |  |
| `spec.dataProtectionConfig` | `AwsWafWebAclDataProtectionConfig` |  |  |  |
| `spec.dataProtectionConfig.dataProtections` | `[]AwsWafWebAclDataProtection` | yes |  |  |
| `spec.dataProtectionConfig.dataProtections[].fieldType` | `string` | yes |  |  |
| `spec.dataProtectionConfig.dataProtections[].fieldKeys` | `[]string` |  |  |  |
| `spec.dataProtectionConfig.dataProtections[].action` | `string` | yes |  |  |
| `spec.dataProtectionConfig.dataProtections[].excludeRuleMatchDetails` | `bool` |  |  |  |
| `spec.dataProtectionConfig.dataProtections[].excludeRateBasedDetails` | `bool` |  |  |  |
| `spec.logging` | `AwsWafWebAclLoggingConfig` |  |  |  |
| `spec.logging.destinationArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`), AwsS3Bucket (`status.outputs.bucket_arn`), AwsKinesisFirehose (`status.outputs.delivery_stream_arn`) |
| `spec.logging.redactedHeaderNames` | `[]string` |  |  |  |
| `spec.logging.redactUriPath` | `bool` |  |  |  |
| `spec.logging.redactQueryString` | `bool` |  |  |  |
| `spec.logging.redactMethod` | `bool` |  |  |  |
| `spec.logging.filter` | `AwsWafWebAclLoggingFilterConfig` |  |  |  |
| `spec.logging.filter.defaultBehavior` | `string` | yes |  |  |
| `spec.logging.filter.filters` | `[]AwsWafWebAclLoggingFilter` | yes |  |  |
| `spec.logging.filter.filters[].behavior` | `string` | yes |  |  |
| `spec.logging.filter.filters[].requirement` | `string` | yes |  |  |
| `spec.logging.filter.filters[].conditions` | `[]AwsWafWebAclLoggingFilterCondition` | yes |  |  |
| `spec.logging.filter.filters[].conditions[].action` | `string` |  |  |  |
| `spec.logging.filter.filters[].conditions[].labelName` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
For CLOUDFRONT scope this must be "us-east-1" (the WAF global region).
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.scope

`string` · required

Scope determines where the Web ACL can be used. Create-time immutable
(ForceNew) — changing it replaces the web ACL.

"REGIONAL" — protects ALB, API Gateway, AppSync, Cognito, App Runner,
and Verified Access resources in the web ACL's own region.

"CLOUDFRONT" — protects CloudFront distributions. The Web ACL MUST be
created in us-east-1 (region must be "us-east-1").

- rule: {"required":true}

### spec.defaultAction

`AwsWafWebAclDefaultAction` · required

Action to take when no rule matches a request. This is the "baseline"
security posture:
- "allow" (permissive): allow all traffic unless a rule blocks it.
  Use when most traffic is legitimate (e.g., public website).
- "block" (restrictive): block all traffic unless a rule allows it.
  Use when most traffic should be denied (e.g., private API).

- rule: {"required":true}
- rule: type must be 'allow' or 'block'
- rule: custom_response is only valid when default action type is 'block'
- rule: custom_request_headers is only valid when default action type is 'allow'

### spec.defaultAction.type

`string` · required

Action type: "allow" or "block".

- rule: {"required":true}

### spec.defaultAction.customResponse

`AwsWafWebAclCustomResponse`

Custom response configuration for block actions. Only valid when
type is "block". Specifies the HTTP response code and optional body
to return to blocked requests.

### spec.defaultAction.customResponse.responseCode

`int32` · required

HTTP response status code to return. Range: 200-600.
Common values: 403 (Forbidden), 429 (Too Many Requests), 503 (Service Unavailable).

- rule: {"required":true,"int32":{"lte":600,"gte":200}}

### spec.defaultAction.customResponse.customResponseBodyKey

`string`

Key referencing a custom_response_body defined at the Web ACL level.
When set, the response body from the matching custom_response_body is
returned with the specified response_code and content type.

### spec.defaultAction.customResponse.responseHeaders

`[]AwsWafWebAclCustomHeader`

Additional HTTP headers to include in the block response.

### spec.defaultAction.customResponse.responseHeaders[].name

`string` · required

HTTP header name (case-insensitive). 1-64 characters: letters, digits,
and _ $ . - only. For INSERTED request headers, WAF prefixes the name
with "x-amzn-waf-" on the wire (name "sample" arrives as
"x-amzn-waf-sample") to avoid clobbering existing request headers;
response headers are sent under the name as given.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_$.-]+$"}}

### spec.defaultAction.customResponse.responseHeaders[].value

`string` · required

HTTP header value. 1-255 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"255"}}

### spec.defaultAction.customRequestHeaders

`[]AwsWafWebAclCustomHeader`

Custom request headers to insert for allow actions. Only valid when
type is "allow". Headers are added to the request before forwarding
to the protected resource.

### spec.defaultAction.customRequestHeaders[].name

`string` · required

HTTP header name (case-insensitive). 1-64 characters: letters, digits,
and _ $ . - only. For INSERTED request headers, WAF prefixes the name
with "x-amzn-waf-" on the wire (name "sample" arrives as
"x-amzn-waf-sample") to avoid clobbering existing request headers;
response headers are sent under the name as given.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_$.-]+$"}}

### spec.defaultAction.customRequestHeaders[].value

`string` · required

HTTP header value. 1-255 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"255"}}

### spec.description

`string`

Human-readable description of the Web ACL. AWS restricts the character
set: letters, digits, whitespace, and _ + = : # @ / - , . only (notably
NO parentheses), 3-256 characters — WAF rejects anything else at create
time, so the constraint is enforced here where the failure is immediate
and readable.

- rule: description may only contain letters, digits, whitespace, and _+=:#@/-,. (no parentheses), and must be at least 3 characters when set
- rule: {"string":{"maxLen":"256"}}

### spec.rules

`[]AwsWafWebAclRule`

Ordered set of rules evaluated against each incoming request. Rules are
evaluated by priority (lowest number first). When a rule matches, its
action is taken and evaluation stops for that request.

When no rules are provided, only the default_action applies. This is
valid but uncommon — most Web ACLs have at least one managed rule group.

- rule: action must be 'allow', 'block', 'count', 'captcha', or 'challenge' when set
- rule: override_action must be 'count' or 'none' when set
- rule: managed_rule_group and rule_group_reference rules must use override_action (not action)
- rule: match rules (everything except managed_rule_group and rule_group_reference) must use action (not override_action)
- rule: custom_response is only valid when action is 'block'

### spec.rules[].name

`string` · required

Unique name for this rule. Used in CloudWatch metrics and log entries.
1-128 characters, must be unique within the Web ACL.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.rules[].priority

`int32` · required · optional (explicit presence)

Evaluation priority. Lower numbers are evaluated first. Must be unique
across all rules in the Web ACL. Range: 0-2147483647. Priority 0 is
legal (AWS-console-generated ACLs commonly start at 0) — the field is
presence-typed exactly so 0 can be expressed while the field stays
required.

- rule: {"required":true,"int32":{"gte":0}}

### spec.rules[].statement

`AwsWafWebAclStatement` · required

The match condition for this rule — exactly one statement type must be
set inside (see AwsWafWebAclStatement for the full statement language,
including AND/OR/NOT composition and the custom_statement escape hatch).

- rule: {"required":true}

### spec.rules[].statement.managedRuleGroup

`AwsWafWebAclManagedRuleGroupStatement`

AWS Managed Rule Group or marketplace rule group — pre-built,
vendor-maintained rule collections (Core rule set, SQLi, bot control,
account takeover prevention, ...). Rule top-level only (AWS rejects it
nested inside AND/OR/NOT or scope-downs); the containing rule must use
override_action.

### spec.rules[].statement.managedRuleGroup.name

`string` · required

Name of the managed rule group (e.g., "AWSManagedRulesCommonRuleSet").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.rules[].statement.managedRuleGroup.vendorName

`string` · required

Vendor that publishes the rule group.
"AWS" for AWS Managed Rules, or the marketplace vendor name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.rules[].statement.managedRuleGroup.version

`string`

Pin to a specific version of the managed rule group (e.g., "Version_1.0").
When omitted, WAF uses the vendor's default (typically the latest) version.
Pinning prevents unexpected behavior changes when the vendor updates rules.

### spec.rules[].statement.managedRuleGroup.ruleActionOverrides

`[]AwsWafWebAclRuleActionOverride`

Override the action for specific rules within the managed group. This is
essential for tuning: you can set individual rules to "count" while the
rest of the group enforces, allowing you to identify false positives
before enabling enforcement.

- rule: action must be 'allow', 'block', 'count', 'captcha', or 'challenge'

### spec.rules[].statement.managedRuleGroup.ruleActionOverrides[].name

`string` · required

Name of the rule within the group to override.
Must match a rule name defined in the rule group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.managedRuleGroup.ruleActionOverrides[].action

`string` · required

Action to use instead of the rule's configured action.
Valid values: "allow", "block", "count", "captcha", "challenge".
Most commonly set to "count" for monitoring without enforcement.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.managedRuleGroup.scopeDownStatement

`AwsWafWebAclStatement`

Optional scope-down statement to narrow which requests the managed rule
group evaluates: only requests matching it are inspected by the group
(e.g. run Bot Control — a paid add-on — only on /api paths). AWS forbids
group references and rate-based statements inside a scope-down.

- rule: a scope-down statement cannot contain managed_rule_group, rule_group_reference, or rate_based statements (AWS restriction)
- rule: recursive: same shape as enclosing AwsWafWebAclStatement

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs

`AwsWafWebAclManagedRuleGroupConfigs`

Additional configuration for the intelligent-threat managed rule groups
(Bot Control, Account Takeover Prevention, Account Creation Fraud
Prevention, anti-DDoS). Required by those groups; meaningless for the
ordinary protection packs (Common, SQLi, KnownBadInputs, ...).

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.botControl

`AwsWafWebAclBotControlConfig`

Configuration for AWSManagedRulesBotControlRuleSet.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.botControl.inspectionLevel

`string` · required

Inspection level:
- "COMMON": self-identifying bots and simple automation (lower WCU/cost).
- "TARGETED": adds ML-based detection of sophisticated bots, browser
  interrogation, and CAPTCHA/challenge actions (higher WCU/cost).

- rule: {"required":true,"string":{"in":["COMMON","TARGETED"]}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.botControl.enableMachineLearning

`bool` · optional (explicit presence)

Whether TARGETED inspection uses machine learning to analyze traffic
statistics for bot patterns. AWS defaults this to true — set false only
to opt out of ML analysis. Ignored for COMMON.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention

`AwsWafWebAclAtpConfig`

Configuration for AWSManagedRulesATPRuleSet (account takeover
prevention — credential stuffing / login abuse detection).

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.loginPath

`string` · required

The path of your login endpoint (e.g. "/api/v1/login"). ATP inspects
only requests to this path.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.enableRegexInPath

`bool`

Treat login_path as a regular expression instead of a literal path.
AWS's default is false, and the AWS API types this as a plain (non-
nullable) boolean — absence and false are identical on the wire, so a
plain bool models it exactly.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection

`AwsWafWebAclAtpRequestInspection`

How to find the username and password in the login request.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.payloadType

`string` · required

Payload encoding of the login request: "JSON" or "FORM_ENCODED".

- rule: {"required":true,"string":{"in":["JSON","FORM_ENCODED"]}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.usernameField

`AwsWafWebAclFieldIdentifier` · required

The username field (JSON pointer like "/email", or form field name).

- rule: {"required":true}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.usernameField.identifier

`string` · required

The field's identifier. 1–512 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"512"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.passwordField

`AwsWafWebAclFieldIdentifier` · required

The password field (JSON pointer like "/password", or form field name).

- rule: {"required":true}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.requestInspection.passwordField.identifier

`string` · required

The field's identifier. 1–512 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"512"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection

`AwsWafWebAclResponseInspection`

How to recognize login success vs failure in your RESPONSES, so ATP can
track failure rates per client. Not available for CLOUDFRONT-scope ACLs.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.statusCode

`AwsWafWebAclResponseInspectionStatusCode`

Outcome signaled by HTTP status codes.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.statusCode.successCodes

`[]int32` · required

Status codes that mean the attempt SUCCEEDED (e.g. [200, 302]).

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.statusCode.failureCodes

`[]int32` · required

Status codes that mean the attempt FAILED (e.g. [401, 403]).

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.header

`AwsWafWebAclResponseInspectionHeader`

Outcome signaled by a response header value.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.header.name

`string` · required

The header to inspect (e.g. "x-login-result").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.header.successValues

`[]string` · required

Header values that mean the attempt SUCCEEDED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.header.failureValues

`[]string` · required

Header values that mean the attempt FAILED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyJson

`AwsWafWebAclResponseInspectionJson`

Outcome signaled by a JSON field in the response body.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyJson.identifier

`string` · required

JSON pointer to the field to inspect (e.g. "/result").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyJson.successValues

`[]string` · required

Field values that mean the attempt SUCCEEDED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyJson.failureValues

`[]string` · required

Field values that mean the attempt FAILED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyContains

`AwsWafWebAclResponseInspectionBodyContains`

Outcome signaled by text appearing in the response body.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyContains.successStrings

`[]string` · required

Body substrings that mean the attempt SUCCEEDED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountTakeoverPrevention.responseInspection.bodyContains.failureStrings

`[]string` · required

Body substrings that mean the attempt FAILED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention

`AwsWafWebAclAcfpConfig`

Configuration for AWSManagedRulesACFPRuleSet (account creation fraud
prevention — fake sign-up detection).

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.creationPath

`string` · required

The path of your account-creation API endpoint (e.g. "/api/signup").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.registrationPagePath

`string` · required

The path of the page presenting the sign-up form (e.g. "/register").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.enableRegexInPath

`bool`

Treat the paths as regular expressions instead of literal paths.
AWS's default is false, and the AWS API types this as a plain (non-
nullable) boolean — absence and false are identical on the wire, so a
plain bool models it exactly.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection

`AwsWafWebAclAcfpRequestInspection` · required

How to find the sign-up fields in the account-creation request.

- rule: {"required":true}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.payloadType

`string` · required

Payload encoding of the sign-up request: "JSON" or "FORM_ENCODED".

- rule: {"required":true,"string":{"in":["JSON","FORM_ENCODED"]}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.usernameField

`AwsWafWebAclFieldIdentifier`

The username field, when the form has one.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.usernameField.identifier

`string` · required

The field's identifier. 1–512 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"512"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.passwordField

`AwsWafWebAclFieldIdentifier`

The password field, when the form has one.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.passwordField.identifier

`string` · required

The field's identifier. 1–512 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"512"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.emailField

`AwsWafWebAclFieldIdentifier`

The email field, when the form has one.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.emailField.identifier

`string` · required

The field's identifier. 1–512 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"512"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.phoneNumberFields

`AwsWafWebAclFieldIdentifiers`

The phone-number fields, when the form has them.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.phoneNumberFields.identifiers

`[]string` · required

The fields' identifiers.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.addressFields

`AwsWafWebAclFieldIdentifiers`

The address fields, when the form has them.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.requestInspection.addressFields.identifiers

`[]string` · required

The fields' identifiers.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection

`AwsWafWebAclResponseInspection`

How to recognize sign-up success vs failure in your RESPONSES. Not
available for CLOUDFRONT-scope ACLs.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.statusCode

`AwsWafWebAclResponseInspectionStatusCode`

Outcome signaled by HTTP status codes.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.statusCode.successCodes

`[]int32` · required

Status codes that mean the attempt SUCCEEDED (e.g. [200, 302]).

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.statusCode.failureCodes

`[]int32` · required

Status codes that mean the attempt FAILED (e.g. [401, 403]).

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.header

`AwsWafWebAclResponseInspectionHeader`

Outcome signaled by a response header value.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.header.name

`string` · required

The header to inspect (e.g. "x-login-result").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.header.successValues

`[]string` · required

Header values that mean the attempt SUCCEEDED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.header.failureValues

`[]string` · required

Header values that mean the attempt FAILED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyJson

`AwsWafWebAclResponseInspectionJson`

Outcome signaled by a JSON field in the response body.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyJson.identifier

`string` · required

JSON pointer to the field to inspect (e.g. "/result").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"256"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyJson.successValues

`[]string` · required

Field values that mean the attempt SUCCEEDED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyJson.failureValues

`[]string` · required

Field values that mean the attempt FAILED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyContains

`AwsWafWebAclResponseInspectionBodyContains`

Outcome signaled by text appearing in the response body.

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyContains.successStrings

`[]string` · required

Body substrings that mean the attempt SUCCEEDED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.accountCreationFraudPrevention.responseInspection.bodyContains.failureStrings

`[]string` · required

Body substrings that mean the attempt FAILED.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos

`AwsWafWebAclAntiDdosConfig`

Configuration for AWSManagedRulesAntiDDoSRuleSet (layer-7 DDoS
mitigation with client-side silent challenges).

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.clientSideAction

`AwsWafWebAclAntiDdosClientSideAction` · required

Client-side action configuration — what the group does to suspicious
clients during a DDoS event.

- rule: {"required":true}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.clientSideAction.usageOfAction

`string` · required

"ENABLED" — serve silent challenges to suspicious clients during an
event (recommended). "DISABLED" — the group only labels traffic.

- rule: {"required":true,"string":{"in":["ENABLED","DISABLED"]}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.clientSideAction.sensitivity

`string`

How readily clients are challenged during an event: "LOW", "MEDIUM", or
"HIGH" (AWS default HIGH — challenge liberally).

- rule: sensitivity must be 'LOW', 'MEDIUM', or 'HIGH' when set

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.clientSideAction.exemptUriRegularExpressions

`[]string`

Regexes for URI paths to EXEMPT from client-side challenges (up to 5) —
paths that must stay reachable by non-browser clients (webhooks, health
checks) even mid-event.

- rule: {"repeated":{"maxItems":"5","items":{"string":{"minLen":"1","maxLen":"512"}}}}

### spec.rules[].statement.managedRuleGroup.managedRuleGroupConfigs.antiDdos.sensitivityToBlock

`string`

How aggressively the group's blocking rule matches during an event:
"LOW" (default — blocks only confident matches), "MEDIUM", or "HIGH"
(blocks more aggressively, more false-positive risk).

- rule: sensitivity_to_block must be 'LOW', 'MEDIUM', or 'HIGH' when set

### spec.rules[].statement.rateBased

`AwsWafWebAclRateBasedStatement`

Rate-based rule that tracks and limits request rates per aggregation
key (source IP by default; forwarded IP, or up to 5 custom keys).
Rule top-level only; the containing rule uses action for what happens
when the limit is exceeded.

- rule: evaluation_window_sec must be 60, 120, 300, or 600 when set
- rule: aggregate_key_type must be 'IP', 'FORWARDED_IP', 'CONSTANT', or 'CUSTOM_KEYS' when set
- rule: forwarded_ip_config is required when aggregate_key_type is 'FORWARDED_IP'
- rule: custom_keys requires aggregate_key_type 'CUSTOM_KEYS', and 'CUSTOM_KEYS' requires at least one custom key

### spec.rules[].statement.rateBased.limit

`int32` · required

Maximum number of requests allowed from a single tracked key within the
evaluation window. When exceeded, the rule's action applies.
Range: 10 to 2,000,000,000.

Common values:
- 100-500: Strict API rate limiting
- 1000-2000: Standard web application protection
- 10000+: DDoS mitigation for high-traffic endpoints

- rule: {"required":true,"int32":{"lte":2000000000,"gte":10}}

### spec.rules[].statement.rateBased.evaluationWindowSec

`int32`

Duration in seconds over which request rates are evaluated.
Valid values: 60, 120, 300 (default), 600.
Shorter windows detect bursts faster but may produce more false positives.

### spec.rules[].statement.rateBased.aggregateKeyType

`string`

How to aggregate requests for rate tracking.

Valid values:
- "IP" (default): Track by source IP address. Simplest and most common.
- "FORWARDED_IP": Track by IP from a forwarded header (e.g., X-Forwarded-For).
  Requires forwarded_ip_config. Use when behind a proxy or CDN.
- "CONSTANT": Count all requests as one group (global rate limit).
  Requires a scope_down_statement to be useful.
- "CUSTOM_KEYS": Track by a composite of up to 5 request properties
  (custom_keys) — e.g. rate-limit per (IP, URI path) pair, per session
  cookie, or per API key header.

### spec.rules[].statement.rateBased.customKeys

`[]AwsWafWebAclRateBasedCustomKey`

The composite aggregation key for CUSTOM_KEYS: each entry contributes
one request property, and requests sharing ALL properties are counted
together. Up to 5 keys. To ALSO include the source IP in the composite,
add an `ip` (or `forwarded_ip`) key entry.

- rule: {"repeated":{"maxItems":"5"}}

### spec.rules[].statement.rateBased.customKeys[].header

`AwsWafWebAclKeyWithTransformations`

Aggregate on a request header's value.

### spec.rules[].statement.rateBased.customKeys[].header.name

`string` · required

The header, cookie, or query-argument name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.rateBased.customKeys[].header.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before the value is used as
an aggregation key. Use one NONE transformation for the raw value.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.rateBased.customKeys[].header.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.rateBased.customKeys[].header.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.rateBased.customKeys[].cookie

`AwsWafWebAclKeyWithTransformations`

Aggregate on a cookie's value.

### spec.rules[].statement.rateBased.customKeys[].cookie.name

`string` · required

The header, cookie, or query-argument name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.rateBased.customKeys[].cookie.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before the value is used as
an aggregation key. Use one NONE transformation for the raw value.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.rateBased.customKeys[].cookie.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.rateBased.customKeys[].cookie.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.rateBased.customKeys[].queryArgument

`AwsWafWebAclKeyWithTransformations`

Aggregate on a query argument's value.

### spec.rules[].statement.rateBased.customKeys[].queryArgument.name

`string` · required

The header, cookie, or query-argument name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.rateBased.customKeys[].queryArgument.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before the value is used as
an aggregation key. Use one NONE transformation for the raw value.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.rateBased.customKeys[].queryArgument.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.rateBased.customKeys[].queryArgument.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.rateBased.customKeys[].queryString

`AwsWafWebAclTransformationsOnlyKey`

Aggregate on the whole query string.

### spec.rules[].statement.rateBased.customKeys[].queryString.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before the component is used
as an aggregation key. Use one NONE transformation for the raw value.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.rateBased.customKeys[].queryString.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.rateBased.customKeys[].queryString.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.rateBased.customKeys[].uriPath

`AwsWafWebAclTransformationsOnlyKey`

Aggregate on the URI path.

### spec.rules[].statement.rateBased.customKeys[].uriPath.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before the component is used
as an aggregation key. Use one NONE transformation for the raw value.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.rateBased.customKeys[].uriPath.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.rateBased.customKeys[].uriPath.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.rateBased.customKeys[].httpMethod

`bool`

Aggregate on the HTTP method (GET/POST/...).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.rateBased.customKeys[].ip

`bool`

Aggregate on the source IP (adds IP to a composite key — with this as
the ONLY key, prefer aggregate_key_type "IP" instead).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.rateBased.customKeys[].forwardedIp

`bool`

Aggregate on the forwarded client IP (requires the statement's
forwarded_ip_config to locate the header).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.rateBased.customKeys[].asn

`bool`

Aggregate on the source address's autonomous system number — one
bucket per network operator.

- rule: {"bool":{"const":true}}

### spec.rules[].statement.rateBased.customKeys[].labelNamespace

`AwsWafWebAclLabelNamespaceKey`

Aggregate on the presence of a label namespace on the request (set by
earlier rules or managed groups).

### spec.rules[].statement.rateBased.customKeys[].labelNamespace.namespace

`string` · required

The label namespace to aggregate on, ending with ':' (e.g.
"awswaf:managed:aws:bot-control:").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.rateBased.customKeys[].ja3Fingerprint

`AwsWafWebAclFingerprintKey`

Aggregate on the JA3 TLS-client fingerprint.

### spec.rules[].statement.rateBased.customKeys[].ja3Fingerprint.fallbackBehavior

`string` · required

How to handle requests without a fingerprint: "MATCH" (aggregate them
together) or "NO_MATCH" (leave them out).

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.rateBased.customKeys[].ja4Fingerprint

`AwsWafWebAclFingerprintKey`

Aggregate on the JA4 TLS-client fingerprint.

### spec.rules[].statement.rateBased.customKeys[].ja4Fingerprint.fallbackBehavior

`string` · required

How to handle requests without a fingerprint: "MATCH" (aggregate them
together) or "NO_MATCH" (leave them out).

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.rateBased.forwardedIpConfig

`AwsWafWebAclForwardedIpConfig`

Forwarded IP configuration. Required when aggregate_key_type is
"FORWARDED_IP". Specifies which header contains the real client IP
and how to handle missing or invalid headers.

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH'
- rule: position must be 'FIRST', 'LAST', or 'ANY' when set

### spec.rules[].statement.rateBased.forwardedIpConfig.headerName

`string` · required

Name of the HTTP header containing the forwarded IP address.
Common values: "X-Forwarded-For", "X-Real-IP", "True-Client-IP".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.rateBased.forwardedIpConfig.fallbackBehavior

`string` · required

How to handle requests with missing or invalid forwarded IP headers.
- "MATCH": Treat as a match (allows the rule to process the request).
- "NO_MATCH": Treat as no match (skips the rule for this request).

- rule: {"required":true}

### spec.rules[].statement.rateBased.forwardedIpConfig.position

`string`

Which IP address to use from multi-value forwarded headers.
Only applicable for IP Set reference rules (ignored by geo_match,
asn_match, and rate_based).

- "FIRST": Use the first IP in the header (closest to the client).
- "LAST": Use the last IP (closest to the server).
- "ANY": Match if any IP in the header matches the IP Set.

When omitted, defaults to "FIRST".

### spec.rules[].statement.rateBased.scopeDownStatement

`AwsWafWebAclStatement`

Optional scope-down statement to narrow which requests are counted
toward the rate limit. Only requests matching this statement contribute
to the rate calculation (e.g. rate-limit only POST /login). AWS forbids
group references and rate-based statements inside a scope-down.

- rule: a scope-down statement cannot contain managed_rule_group, rule_group_reference, or rate_based statements (AWS restriction)
- rule: recursive: same shape as enclosing AwsWafWebAclStatement

### spec.rules[].statement.ruleGroupReference

`AwsWafWebAclRuleGroupReferenceStatement`

Reference to YOUR OWN WAFv2 rule group by ARN (a reusable rule
collection you manage outside this web ACL). Rule top-level only; the
containing rule must use override_action.

### spec.rules[].statement.ruleGroupReference.arn

`string` · required

ARN of the WAFv2 rule group. Must be in the same scope (and for
REGIONAL, the same region) as this web ACL.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.ruleGroupReference.ruleActionOverrides

`[]AwsWafWebAclRuleActionOverride`

Override the action for specific rules within the referenced group
(the managed-rule-group tuning workflow applies identically).

- rule: action must be 'allow', 'block', 'count', 'captcha', or 'challenge'

### spec.rules[].statement.ruleGroupReference.ruleActionOverrides[].name

`string` · required

Name of the rule within the group to override.
Must match a rule name defined in the rule group.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.ruleGroupReference.ruleActionOverrides[].action

`string` · required

Action to use instead of the rule's configured action.
Valid values: "allow", "block", "count", "captcha", "challenge".
Most commonly set to "count" for monitoring without enforcement.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.ipSetReference

`AwsWafWebAclIpSetReferenceStatement`

Match source IPs against a WAFv2 IP set (AwsWafIpSet) — the allow-list
/ deny-list building block.

### spec.rules[].statement.ipSetReference.arn

`string | valueFrom` · required

The IP set to match against — reference an AwsWafIpSet or provide a
literal WAFv2 IP set ARN. The set must be in the same scope (and for
REGIONAL, the same region) as this web ACL.

- references: AwsWafIpSet (`status.outputs.ip_set_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsWafIpSet, name: <that resource's name>, fieldPath: status.outputs.ip_set_arn}} -- a bare string does not parse

### spec.rules[].statement.ipSetReference.forwardedIpConfig

`AwsWafWebAclForwardedIpConfig`

Optional forwarded IP configuration. Use when the Web ACL is behind a
proxy or CDN and the real client IP is in a forwarded header.

When set, the position field controls which IP from the forwarded header
is matched against the IP Set (FIRST, LAST, or ANY).

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH'
- rule: position must be 'FIRST', 'LAST', or 'ANY' when set

### spec.rules[].statement.ipSetReference.forwardedIpConfig.headerName

`string` · required

Name of the HTTP header containing the forwarded IP address.
Common values: "X-Forwarded-For", "X-Real-IP", "True-Client-IP".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.ipSetReference.forwardedIpConfig.fallbackBehavior

`string` · required

How to handle requests with missing or invalid forwarded IP headers.
- "MATCH": Treat as a match (allows the rule to process the request).
- "NO_MATCH": Treat as no match (skips the rule for this request).

- rule: {"required":true}

### spec.rules[].statement.ipSetReference.forwardedIpConfig.position

`string`

Which IP address to use from multi-value forwarded headers.
Only applicable for IP Set reference rules (ignored by geo_match,
asn_match, and rate_based).

- "FIRST": Use the first IP in the header (closest to the client).
- "LAST": Use the last IP (closest to the server).
- "ANY": Match if any IP in the header matches the IP Set.

When omitted, defaults to "FIRST".

### spec.rules[].statement.regexPatternSetReference

`AwsWafWebAclRegexPatternSetReferenceStatement`

Match a request component against a WAFv2 regex pattern set
(AwsWafRegexPatternSet) — matches when ANY regex in the set matches.

### spec.rules[].statement.regexPatternSetReference.arn

`string | valueFrom` · required

The pattern set to match against — reference an AwsWafRegexPatternSet or
provide a literal WAFv2 regex pattern set ARN. The set must be in the
same scope (and for REGIONAL, the same region) as this web ACL.

- references: AwsWafRegexPatternSet (`status.outputs.regex_pattern_set_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsWafRegexPatternSet, name: <that resource's name>, fieldPath: status.outputs.regex_pattern_set_arn}} -- a bare string does not parse

### spec.rules[].statement.regexPatternSetReference.fieldToMatch

`AwsWafWebAclFieldToMatch` · required

The request component to inspect.

- rule: {"required":true}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.uriPath

`bool`

Inspect the URI path (e.g. "/images/daily-ad.jpg").

- rule: {"bool":{"const":true}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.queryString

`bool`

Inspect the whole query string (everything after the '?').

- rule: {"bool":{"const":true}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.method

`bool`

Inspect the HTTP method (GET, POST, ...).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.allQueryArguments

`bool`

Inspect all query arguments' values (max 30 arguments, 10 KB).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.singleHeader

`AwsWafWebAclSingleField`

Inspect a single named header (e.g. "User-Agent").

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.singleHeader.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.singleQueryArgument

`AwsWafWebAclSingleField`

Inspect a single named query argument's value (max 30 characters for
the name).

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.singleQueryArgument.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.body

`AwsWafWebAclBodyMatch`

Inspect the raw request body.

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.body.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default — inspect what fits), "MATCH" (treat oversize as a
rule match — the safe choice for block rules), or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody

`AwsWafWebAclJsonBodyMatch`

Inspect the request body parsed as JSON — match against keys, values,
or both, over all elements or specific JSON pointer paths.

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody.matchScope

`string` · required

Which parts of the JSON to match against: "ALL" (keys and values),
"KEY", or "VALUE".

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody.includedPaths

`[]string`

JSON pointer paths to inspect (e.g. "/user/name"). When empty, ALL
elements are inspected.

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody.invalidFallbackBehavior

`string`

What to do when the body is not valid JSON: "MATCH", "NO_MATCH", or
"EVALUATE_AS_STRING" (fall back to raw-string inspection). When unset,
AWS applies its default host-of-checks behavior.

- rule: invalid_fallback_behavior must be 'MATCH', 'NO_MATCH', or 'EVALUATE_AS_STRING' when set

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.jsonBody.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default), "MATCH", or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers

`AwsWafWebAclHeadersMatch`

Inspect multiple headers by pattern (all, an include list, or an
exclude list).

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchPattern

`AwsWafWebAclNamePattern` · required

Which headers to inspect: set exactly one of all / included_headers /
excluded_headers.

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.headers.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit (200 headers /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies

`AwsWafWebAclCookiesMatch`

Inspect cookies by pattern (all, an include list, or an exclude list).

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchPattern

`AwsWafWebAclNamePattern` · required

Which cookies to inspect: set exactly one of all / included_names
(cookie names) / excluded_names (cookie names).

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.cookies.oversizeHandling

`string` · required

What to do when the cookies exceed WAF's inspection limit (200 cookies /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.headerOrder

`AwsWafWebAclHeaderOrderMatch`

Inspect the comma-delimited list of header NAMES in receive order —
header-order fingerprinting of clients.

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.headerOrder.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit:
"CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.ja3Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA3 TLS-client fingerprint (a 32-character hash of the
TLS ClientHello — fingerprints the client software).

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.ja3Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.ja4Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA4 TLS-client fingerprint (the successor fingerprint
format to JA3).

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.ja4Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.uriFragment

`AwsWafWebAclUriFragmentMatch`

Inspect the URI fragment (the part after '#'). Rarely sent by
browsers; useful for API clients that forward fragments.

### spec.rules[].statement.regexPatternSetReference.fieldToMatch.uriFragment.fallbackBehavior

`string`

What to do when the request has no fragment: "MATCH" or "NO_MATCH".
When unset, AWS treats missing fragments as NO_MATCH.

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH' when set

### spec.rules[].statement.regexPatternSetReference.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before matching. Use one
NONE transformation when no normalization is wanted.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.regexPatternSetReference.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.regexPatternSetReference.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.geoMatch

`AwsWafWebAclGeoMatchStatement`

Match requests by country of origin (from the source IP or a
forwarded header).

### spec.rules[].statement.geoMatch.countryCodes

`[]string` · required

ISO 3166-1 alpha-2 country codes to match against (two uppercase
letters). Examples: "US", "CA", "GB", "DE", "JP", "AU".

At least one country code is required.

- rule: {"required":true,"repeated":{"minItems":"1","items":{"string":{"pattern":"^[A-Z]{2}$"}}}}

### spec.rules[].statement.geoMatch.forwardedIpConfig

`AwsWafWebAclForwardedIpConfig`

Optional forwarded IP configuration. Use when the Web ACL is behind a
proxy or CDN and the real client IP is in a forwarded header.

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH'
- rule: position must be 'FIRST', 'LAST', or 'ANY' when set

### spec.rules[].statement.geoMatch.forwardedIpConfig.headerName

`string` · required

Name of the HTTP header containing the forwarded IP address.
Common values: "X-Forwarded-For", "X-Real-IP", "True-Client-IP".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.geoMatch.forwardedIpConfig.fallbackBehavior

`string` · required

How to handle requests with missing or invalid forwarded IP headers.
- "MATCH": Treat as a match (allows the rule to process the request).
- "NO_MATCH": Treat as no match (skips the rule for this request).

- rule: {"required":true}

### spec.rules[].statement.geoMatch.forwardedIpConfig.position

`string`

Which IP address to use from multi-value forwarded headers.
Only applicable for IP Set reference rules (ignored by geo_match,
asn_match, and rate_based).

- "FIRST": Use the first IP in the header (closest to the client).
- "LAST": Use the last IP (closest to the server).
- "ANY": Match if any IP in the header matches the IP Set.

When omitted, defaults to "FIRST".

### spec.rules[].statement.byteMatch

`AwsWafWebAclByteMatchStatement`

Match a string (or its position) inside a request component — the
workhorse for path prefixes, header values, user-agent substrings.

### spec.rules[].statement.byteMatch.searchString

`string` · required

The string to search for (1–200 bytes). WAF compares raw bytes after
text transformations are applied.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"200"}}

### spec.rules[].statement.byteMatch.positionalConstraint

`string` · required

Where the string must appear in the inspected component:
"EXACTLY", "STARTS_WITH", "ENDS_WITH", "CONTAINS", or "CONTAINS_WORD"
(CONTAINS_WORD requires alphanumeric-boundary separation).

- rule: {"required":true,"string":{"in":["EXACTLY","STARTS_WITH","ENDS_WITH","CONTAINS","CONTAINS_WORD"]}}

### spec.rules[].statement.byteMatch.fieldToMatch

`AwsWafWebAclFieldToMatch` · required

The request component to inspect.

- rule: {"required":true}

### spec.rules[].statement.byteMatch.fieldToMatch.uriPath

`bool`

Inspect the URI path (e.g. "/images/daily-ad.jpg").

- rule: {"bool":{"const":true}}

### spec.rules[].statement.byteMatch.fieldToMatch.queryString

`bool`

Inspect the whole query string (everything after the '?').

- rule: {"bool":{"const":true}}

### spec.rules[].statement.byteMatch.fieldToMatch.method

`bool`

Inspect the HTTP method (GET, POST, ...).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.byteMatch.fieldToMatch.allQueryArguments

`bool`

Inspect all query arguments' values (max 30 arguments, 10 KB).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.byteMatch.fieldToMatch.singleHeader

`AwsWafWebAclSingleField`

Inspect a single named header (e.g. "User-Agent").

### spec.rules[].statement.byteMatch.fieldToMatch.singleHeader.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.byteMatch.fieldToMatch.singleQueryArgument

`AwsWafWebAclSingleField`

Inspect a single named query argument's value (max 30 characters for
the name).

### spec.rules[].statement.byteMatch.fieldToMatch.singleQueryArgument.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.byteMatch.fieldToMatch.body

`AwsWafWebAclBodyMatch`

Inspect the raw request body.

### spec.rules[].statement.byteMatch.fieldToMatch.body.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default — inspect what fits), "MATCH" (treat oversize as a
rule match — the safe choice for block rules), or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.byteMatch.fieldToMatch.jsonBody

`AwsWafWebAclJsonBodyMatch`

Inspect the request body parsed as JSON — match against keys, values,
or both, over all elements or specific JSON pointer paths.

### spec.rules[].statement.byteMatch.fieldToMatch.jsonBody.matchScope

`string` · required

Which parts of the JSON to match against: "ALL" (keys and values),
"KEY", or "VALUE".

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.byteMatch.fieldToMatch.jsonBody.includedPaths

`[]string`

JSON pointer paths to inspect (e.g. "/user/name"). When empty, ALL
elements are inspected.

### spec.rules[].statement.byteMatch.fieldToMatch.jsonBody.invalidFallbackBehavior

`string`

What to do when the body is not valid JSON: "MATCH", "NO_MATCH", or
"EVALUATE_AS_STRING" (fall back to raw-string inspection). When unset,
AWS applies its default host-of-checks behavior.

- rule: invalid_fallback_behavior must be 'MATCH', 'NO_MATCH', or 'EVALUATE_AS_STRING' when set

### spec.rules[].statement.byteMatch.fieldToMatch.jsonBody.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default), "MATCH", or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.byteMatch.fieldToMatch.headers

`AwsWafWebAclHeadersMatch`

Inspect multiple headers by pattern (all, an include list, or an
exclude list).

### spec.rules[].statement.byteMatch.fieldToMatch.headers.matchPattern

`AwsWafWebAclNamePattern` · required

Which headers to inspect: set exactly one of all / included_headers /
excluded_headers.

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.byteMatch.fieldToMatch.headers.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.byteMatch.fieldToMatch.headers.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.byteMatch.fieldToMatch.headers.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.byteMatch.fieldToMatch.headers.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.byteMatch.fieldToMatch.headers.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit (200 headers /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.byteMatch.fieldToMatch.cookies

`AwsWafWebAclCookiesMatch`

Inspect cookies by pattern (all, an include list, or an exclude list).

### spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchPattern

`AwsWafWebAclNamePattern` · required

Which cookies to inspect: set exactly one of all / included_names
(cookie names) / excluded_names (cookie names).

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.byteMatch.fieldToMatch.cookies.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.byteMatch.fieldToMatch.cookies.oversizeHandling

`string` · required

What to do when the cookies exceed WAF's inspection limit (200 cookies /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.byteMatch.fieldToMatch.headerOrder

`AwsWafWebAclHeaderOrderMatch`

Inspect the comma-delimited list of header NAMES in receive order —
header-order fingerprinting of clients.

### spec.rules[].statement.byteMatch.fieldToMatch.headerOrder.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit:
"CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.byteMatch.fieldToMatch.ja3Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA3 TLS-client fingerprint (a 32-character hash of the
TLS ClientHello — fingerprints the client software).

### spec.rules[].statement.byteMatch.fieldToMatch.ja3Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.byteMatch.fieldToMatch.ja4Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA4 TLS-client fingerprint (the successor fingerprint
format to JA3).

### spec.rules[].statement.byteMatch.fieldToMatch.ja4Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.byteMatch.fieldToMatch.uriFragment

`AwsWafWebAclUriFragmentMatch`

Inspect the URI fragment (the part after '#'). Rarely sent by
browsers; useful for API clients that forward fragments.

### spec.rules[].statement.byteMatch.fieldToMatch.uriFragment.fallbackBehavior

`string`

What to do when the request has no fragment: "MATCH" or "NO_MATCH".
When unset, AWS treats missing fragments as NO_MATCH.

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH' when set

### spec.rules[].statement.byteMatch.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before matching. Use one
NONE transformation when no normalization is wanted.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.byteMatch.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.byteMatch.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.sqliMatch

`AwsWafWebAclSqliMatchStatement`

Detect SQL-injection attack patterns in a request component.

### spec.rules[].statement.sqliMatch.fieldToMatch

`AwsWafWebAclFieldToMatch` · required

The request component to inspect.

- rule: {"required":true}

### spec.rules[].statement.sqliMatch.fieldToMatch.uriPath

`bool`

Inspect the URI path (e.g. "/images/daily-ad.jpg").

- rule: {"bool":{"const":true}}

### spec.rules[].statement.sqliMatch.fieldToMatch.queryString

`bool`

Inspect the whole query string (everything after the '?').

- rule: {"bool":{"const":true}}

### spec.rules[].statement.sqliMatch.fieldToMatch.method

`bool`

Inspect the HTTP method (GET, POST, ...).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.sqliMatch.fieldToMatch.allQueryArguments

`bool`

Inspect all query arguments' values (max 30 arguments, 10 KB).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.sqliMatch.fieldToMatch.singleHeader

`AwsWafWebAclSingleField`

Inspect a single named header (e.g. "User-Agent").

### spec.rules[].statement.sqliMatch.fieldToMatch.singleHeader.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.sqliMatch.fieldToMatch.singleQueryArgument

`AwsWafWebAclSingleField`

Inspect a single named query argument's value (max 30 characters for
the name).

### spec.rules[].statement.sqliMatch.fieldToMatch.singleQueryArgument.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.sqliMatch.fieldToMatch.body

`AwsWafWebAclBodyMatch`

Inspect the raw request body.

### spec.rules[].statement.sqliMatch.fieldToMatch.body.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default — inspect what fits), "MATCH" (treat oversize as a
rule match — the safe choice for block rules), or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody

`AwsWafWebAclJsonBodyMatch`

Inspect the request body parsed as JSON — match against keys, values,
or both, over all elements or specific JSON pointer paths.

### spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody.matchScope

`string` · required

Which parts of the JSON to match against: "ALL" (keys and values),
"KEY", or "VALUE".

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody.includedPaths

`[]string`

JSON pointer paths to inspect (e.g. "/user/name"). When empty, ALL
elements are inspected.

### spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody.invalidFallbackBehavior

`string`

What to do when the body is not valid JSON: "MATCH", "NO_MATCH", or
"EVALUATE_AS_STRING" (fall back to raw-string inspection). When unset,
AWS applies its default host-of-checks behavior.

- rule: invalid_fallback_behavior must be 'MATCH', 'NO_MATCH', or 'EVALUATE_AS_STRING' when set

### spec.rules[].statement.sqliMatch.fieldToMatch.jsonBody.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default), "MATCH", or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.sqliMatch.fieldToMatch.headers

`AwsWafWebAclHeadersMatch`

Inspect multiple headers by pattern (all, an include list, or an
exclude list).

### spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchPattern

`AwsWafWebAclNamePattern` · required

Which headers to inspect: set exactly one of all / included_headers /
excluded_headers.

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.sqliMatch.fieldToMatch.headers.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.sqliMatch.fieldToMatch.headers.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit (200 headers /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.sqliMatch.fieldToMatch.cookies

`AwsWafWebAclCookiesMatch`

Inspect cookies by pattern (all, an include list, or an exclude list).

### spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchPattern

`AwsWafWebAclNamePattern` · required

Which cookies to inspect: set exactly one of all / included_names
(cookie names) / excluded_names (cookie names).

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.sqliMatch.fieldToMatch.cookies.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.sqliMatch.fieldToMatch.cookies.oversizeHandling

`string` · required

What to do when the cookies exceed WAF's inspection limit (200 cookies /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.sqliMatch.fieldToMatch.headerOrder

`AwsWafWebAclHeaderOrderMatch`

Inspect the comma-delimited list of header NAMES in receive order —
header-order fingerprinting of clients.

### spec.rules[].statement.sqliMatch.fieldToMatch.headerOrder.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit:
"CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.sqliMatch.fieldToMatch.ja3Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA3 TLS-client fingerprint (a 32-character hash of the
TLS ClientHello — fingerprints the client software).

### spec.rules[].statement.sqliMatch.fieldToMatch.ja3Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.sqliMatch.fieldToMatch.ja4Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA4 TLS-client fingerprint (the successor fingerprint
format to JA3).

### spec.rules[].statement.sqliMatch.fieldToMatch.ja4Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.sqliMatch.fieldToMatch.uriFragment

`AwsWafWebAclUriFragmentMatch`

Inspect the URI fragment (the part after '#'). Rarely sent by
browsers; useful for API clients that forward fragments.

### spec.rules[].statement.sqliMatch.fieldToMatch.uriFragment.fallbackBehavior

`string`

What to do when the request has no fragment: "MATCH" or "NO_MATCH".
When unset, AWS treats missing fragments as NO_MATCH.

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH' when set

### spec.rules[].statement.sqliMatch.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before detection — URL_DECODE
and HTML_ENTITY_DECODE defeat common encoding evasion.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.sqliMatch.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.sqliMatch.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.sqliMatch.sensitivityLevel

`string`

Detection sensitivity: "LOW" (default — fewer false positives) or
"HIGH" (catches more injection styles, more false-positive risk).

- rule: sensitivity_level must be 'LOW' or 'HIGH' when set

### spec.rules[].statement.xssMatch

`AwsWafWebAclXssMatchStatement`

Detect cross-site-scripting attack patterns in a request component.

### spec.rules[].statement.xssMatch.fieldToMatch

`AwsWafWebAclFieldToMatch` · required

The request component to inspect.

- rule: {"required":true}

### spec.rules[].statement.xssMatch.fieldToMatch.uriPath

`bool`

Inspect the URI path (e.g. "/images/daily-ad.jpg").

- rule: {"bool":{"const":true}}

### spec.rules[].statement.xssMatch.fieldToMatch.queryString

`bool`

Inspect the whole query string (everything after the '?').

- rule: {"bool":{"const":true}}

### spec.rules[].statement.xssMatch.fieldToMatch.method

`bool`

Inspect the HTTP method (GET, POST, ...).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.xssMatch.fieldToMatch.allQueryArguments

`bool`

Inspect all query arguments' values (max 30 arguments, 10 KB).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.xssMatch.fieldToMatch.singleHeader

`AwsWafWebAclSingleField`

Inspect a single named header (e.g. "User-Agent").

### spec.rules[].statement.xssMatch.fieldToMatch.singleHeader.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.xssMatch.fieldToMatch.singleQueryArgument

`AwsWafWebAclSingleField`

Inspect a single named query argument's value (max 30 characters for
the name).

### spec.rules[].statement.xssMatch.fieldToMatch.singleQueryArgument.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.xssMatch.fieldToMatch.body

`AwsWafWebAclBodyMatch`

Inspect the raw request body.

### spec.rules[].statement.xssMatch.fieldToMatch.body.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default — inspect what fits), "MATCH" (treat oversize as a
rule match — the safe choice for block rules), or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.xssMatch.fieldToMatch.jsonBody

`AwsWafWebAclJsonBodyMatch`

Inspect the request body parsed as JSON — match against keys, values,
or both, over all elements or specific JSON pointer paths.

### spec.rules[].statement.xssMatch.fieldToMatch.jsonBody.matchScope

`string` · required

Which parts of the JSON to match against: "ALL" (keys and values),
"KEY", or "VALUE".

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.xssMatch.fieldToMatch.jsonBody.includedPaths

`[]string`

JSON pointer paths to inspect (e.g. "/user/name"). When empty, ALL
elements are inspected.

### spec.rules[].statement.xssMatch.fieldToMatch.jsonBody.invalidFallbackBehavior

`string`

What to do when the body is not valid JSON: "MATCH", "NO_MATCH", or
"EVALUATE_AS_STRING" (fall back to raw-string inspection). When unset,
AWS applies its default host-of-checks behavior.

- rule: invalid_fallback_behavior must be 'MATCH', 'NO_MATCH', or 'EVALUATE_AS_STRING' when set

### spec.rules[].statement.xssMatch.fieldToMatch.jsonBody.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default), "MATCH", or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.xssMatch.fieldToMatch.headers

`AwsWafWebAclHeadersMatch`

Inspect multiple headers by pattern (all, an include list, or an
exclude list).

### spec.rules[].statement.xssMatch.fieldToMatch.headers.matchPattern

`AwsWafWebAclNamePattern` · required

Which headers to inspect: set exactly one of all / included_headers /
excluded_headers.

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.xssMatch.fieldToMatch.headers.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.xssMatch.fieldToMatch.headers.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.xssMatch.fieldToMatch.headers.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.xssMatch.fieldToMatch.headers.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.xssMatch.fieldToMatch.headers.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit (200 headers /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.xssMatch.fieldToMatch.cookies

`AwsWafWebAclCookiesMatch`

Inspect cookies by pattern (all, an include list, or an exclude list).

### spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchPattern

`AwsWafWebAclNamePattern` · required

Which cookies to inspect: set exactly one of all / included_names
(cookie names) / excluded_names (cookie names).

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.xssMatch.fieldToMatch.cookies.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.xssMatch.fieldToMatch.cookies.oversizeHandling

`string` · required

What to do when the cookies exceed WAF's inspection limit (200 cookies /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.xssMatch.fieldToMatch.headerOrder

`AwsWafWebAclHeaderOrderMatch`

Inspect the comma-delimited list of header NAMES in receive order —
header-order fingerprinting of clients.

### spec.rules[].statement.xssMatch.fieldToMatch.headerOrder.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit:
"CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.xssMatch.fieldToMatch.ja3Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA3 TLS-client fingerprint (a 32-character hash of the
TLS ClientHello — fingerprints the client software).

### spec.rules[].statement.xssMatch.fieldToMatch.ja3Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.xssMatch.fieldToMatch.ja4Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA4 TLS-client fingerprint (the successor fingerprint
format to JA3).

### spec.rules[].statement.xssMatch.fieldToMatch.ja4Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.xssMatch.fieldToMatch.uriFragment

`AwsWafWebAclUriFragmentMatch`

Inspect the URI fragment (the part after '#'). Rarely sent by
browsers; useful for API clients that forward fragments.

### spec.rules[].statement.xssMatch.fieldToMatch.uriFragment.fallbackBehavior

`string`

What to do when the request has no fragment: "MATCH" or "NO_MATCH".
When unset, AWS treats missing fragments as NO_MATCH.

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH' when set

### spec.rules[].statement.xssMatch.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before detection —
HTML_ENTITY_DECODE and URL_DECODE defeat common encoding evasion.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.xssMatch.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.xssMatch.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.sizeConstraint

`AwsWafWebAclSizeConstraintStatement`

Compare the byte size of a request component against a threshold
(e.g. block bodies over 8 KB, or URI paths longer than 128 bytes).

### spec.rules[].statement.sizeConstraint.comparisonOperator

`string` · required

The comparison: "EQ", "NE", "LE", "LT", "GE", or "GT"
(component_size <op> size).

- rule: {"required":true,"string":{"in":["EQ","NE","LE","LT","GE","GT"]}}

### spec.rules[].statement.sizeConstraint.size

`int32`

The size threshold in bytes. The provider caps this at 2,147,483,647
(int32 max) even though the AWS API documents a 21,474,836,480 ceiling —
the tighter bound is what a deployment can actually carry. Typed int32
deliberately: protojson stringifies 64-bit integers, which would corrupt
the rule-JSON document both engines build from this tree.

- rule: {"int32":{"lte":2147483647,"gte":0}}

### spec.rules[].statement.sizeConstraint.fieldToMatch

`AwsWafWebAclFieldToMatch` · required

The request component to measure.

- rule: {"required":true}

### spec.rules[].statement.sizeConstraint.fieldToMatch.uriPath

`bool`

Inspect the URI path (e.g. "/images/daily-ad.jpg").

- rule: {"bool":{"const":true}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.queryString

`bool`

Inspect the whole query string (everything after the '?').

- rule: {"bool":{"const":true}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.method

`bool`

Inspect the HTTP method (GET, POST, ...).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.allQueryArguments

`bool`

Inspect all query arguments' values (max 30 arguments, 10 KB).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.singleHeader

`AwsWafWebAclSingleField`

Inspect a single named header (e.g. "User-Agent").

### spec.rules[].statement.sizeConstraint.fieldToMatch.singleHeader.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.singleQueryArgument

`AwsWafWebAclSingleField`

Inspect a single named query argument's value (max 30 characters for
the name).

### spec.rules[].statement.sizeConstraint.fieldToMatch.singleQueryArgument.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.body

`AwsWafWebAclBodyMatch`

Inspect the raw request body.

### spec.rules[].statement.sizeConstraint.fieldToMatch.body.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default — inspect what fits), "MATCH" (treat oversize as a
rule match — the safe choice for block rules), or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody

`AwsWafWebAclJsonBodyMatch`

Inspect the request body parsed as JSON — match against keys, values,
or both, over all elements or specific JSON pointer paths.

### spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody.matchScope

`string` · required

Which parts of the JSON to match against: "ALL" (keys and values),
"KEY", or "VALUE".

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody.includedPaths

`[]string`

JSON pointer paths to inspect (e.g. "/user/name"). When empty, ALL
elements are inspected.

### spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody.invalidFallbackBehavior

`string`

What to do when the body is not valid JSON: "MATCH", "NO_MATCH", or
"EVALUATE_AS_STRING" (fall back to raw-string inspection). When unset,
AWS applies its default host-of-checks behavior.

- rule: invalid_fallback_behavior must be 'MATCH', 'NO_MATCH', or 'EVALUATE_AS_STRING' when set

### spec.rules[].statement.sizeConstraint.fieldToMatch.jsonBody.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default), "MATCH", or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.sizeConstraint.fieldToMatch.headers

`AwsWafWebAclHeadersMatch`

Inspect multiple headers by pattern (all, an include list, or an
exclude list).

### spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchPattern

`AwsWafWebAclNamePattern` · required

Which headers to inspect: set exactly one of all / included_headers /
excluded_headers.

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.headers.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.headers.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit (200 headers /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.cookies

`AwsWafWebAclCookiesMatch`

Inspect cookies by pattern (all, an include list, or an exclude list).

### spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchPattern

`AwsWafWebAclNamePattern` · required

Which cookies to inspect: set exactly one of all / included_names
(cookie names) / excluded_names (cookie names).

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.cookies.oversizeHandling

`string` · required

What to do when the cookies exceed WAF's inspection limit (200 cookies /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.headerOrder

`AwsWafWebAclHeaderOrderMatch`

Inspect the comma-delimited list of header NAMES in receive order —
header-order fingerprinting of clients.

### spec.rules[].statement.sizeConstraint.fieldToMatch.headerOrder.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit:
"CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.ja3Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA3 TLS-client fingerprint (a 32-character hash of the
TLS ClientHello — fingerprints the client software).

### spec.rules[].statement.sizeConstraint.fieldToMatch.ja3Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.ja4Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA4 TLS-client fingerprint (the successor fingerprint
format to JA3).

### spec.rules[].statement.sizeConstraint.fieldToMatch.ja4Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.sizeConstraint.fieldToMatch.uriFragment

`AwsWafWebAclUriFragmentMatch`

Inspect the URI fragment (the part after '#'). Rarely sent by
browsers; useful for API clients that forward fragments.

### spec.rules[].statement.sizeConstraint.fieldToMatch.uriFragment.fallbackBehavior

`string`

What to do when the request has no fragment: "MATCH" or "NO_MATCH".
When unset, AWS treats missing fragments as NO_MATCH.

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH' when set

### spec.rules[].statement.sizeConstraint.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before measuring. Use one
NONE transformation to measure the raw component.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.sizeConstraint.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.sizeConstraint.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.regexMatch

`AwsWafWebAclRegexMatchStatement`

Match a request component against a single inline regular expression.
For regexes shared across rules, prefer regex_pattern_set_reference.

### spec.rules[].statement.regexMatch.regexString

`string` · required

The regular expression (1–512 characters).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"512"}}

### spec.rules[].statement.regexMatch.fieldToMatch

`AwsWafWebAclFieldToMatch` · required

The request component to inspect.

- rule: {"required":true}

### spec.rules[].statement.regexMatch.fieldToMatch.uriPath

`bool`

Inspect the URI path (e.g. "/images/daily-ad.jpg").

- rule: {"bool":{"const":true}}

### spec.rules[].statement.regexMatch.fieldToMatch.queryString

`bool`

Inspect the whole query string (everything after the '?').

- rule: {"bool":{"const":true}}

### spec.rules[].statement.regexMatch.fieldToMatch.method

`bool`

Inspect the HTTP method (GET, POST, ...).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.regexMatch.fieldToMatch.allQueryArguments

`bool`

Inspect all query arguments' values (max 30 arguments, 10 KB).

- rule: {"bool":{"const":true}}

### spec.rules[].statement.regexMatch.fieldToMatch.singleHeader

`AwsWafWebAclSingleField`

Inspect a single named header (e.g. "User-Agent").

### spec.rules[].statement.regexMatch.fieldToMatch.singleHeader.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.regexMatch.fieldToMatch.singleQueryArgument

`AwsWafWebAclSingleField`

Inspect a single named query argument's value (max 30 characters for
the name).

### spec.rules[].statement.regexMatch.fieldToMatch.singleQueryArgument.name

`string` · required

The header or query-argument name (case-insensitive for headers).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.regexMatch.fieldToMatch.body

`AwsWafWebAclBodyMatch`

Inspect the raw request body.

### spec.rules[].statement.regexMatch.fieldToMatch.body.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default — inspect what fits), "MATCH" (treat oversize as a
rule match — the safe choice for block rules), or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.regexMatch.fieldToMatch.jsonBody

`AwsWafWebAclJsonBodyMatch`

Inspect the request body parsed as JSON — match against keys, values,
or both, over all elements or specific JSON pointer paths.

### spec.rules[].statement.regexMatch.fieldToMatch.jsonBody.matchScope

`string` · required

Which parts of the JSON to match against: "ALL" (keys and values),
"KEY", or "VALUE".

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.regexMatch.fieldToMatch.jsonBody.includedPaths

`[]string`

JSON pointer paths to inspect (e.g. "/user/name"). When empty, ALL
elements are inspected.

### spec.rules[].statement.regexMatch.fieldToMatch.jsonBody.invalidFallbackBehavior

`string`

What to do when the body is not valid JSON: "MATCH", "NO_MATCH", or
"EVALUATE_AS_STRING" (fall back to raw-string inspection). When unset,
AWS applies its default host-of-checks behavior.

- rule: invalid_fallback_behavior must be 'MATCH', 'NO_MATCH', or 'EVALUATE_AS_STRING' when set

### spec.rules[].statement.regexMatch.fieldToMatch.jsonBody.oversizeHandling

`string`

What to do when the body exceeds the inspection size limit:
"CONTINUE" (default), "MATCH", or "NO_MATCH".

- rule: oversize_handling must be 'CONTINUE', 'MATCH', or 'NO_MATCH' when set

### spec.rules[].statement.regexMatch.fieldToMatch.headers

`AwsWafWebAclHeadersMatch`

Inspect multiple headers by pattern (all, an include list, or an
exclude list).

### spec.rules[].statement.regexMatch.fieldToMatch.headers.matchPattern

`AwsWafWebAclNamePattern` · required

Which headers to inspect: set exactly one of all / included_headers /
excluded_headers.

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.regexMatch.fieldToMatch.headers.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.regexMatch.fieldToMatch.headers.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.regexMatch.fieldToMatch.headers.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.regexMatch.fieldToMatch.headers.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.regexMatch.fieldToMatch.headers.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit (200 headers /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexMatch.fieldToMatch.cookies

`AwsWafWebAclCookiesMatch`

Inspect cookies by pattern (all, an include list, or an exclude list).

### spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchPattern

`AwsWafWebAclNamePattern` · required

Which cookies to inspect: set exactly one of all / included_names
(cookie names) / excluded_names (cookie names).

- rule: {"required":true}
- rule: set exactly one of all, included_names, or excluded_names

### spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchPattern.all

`bool`

Inspect all headers/cookies.

### spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchPattern.includedNames

`[]string`

Inspect only these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchPattern.excludedNames

`[]string`

Inspect all EXCEPT these names (max 199).

- rule: {"repeated":{"maxItems":"199"}}

### spec.rules[].statement.regexMatch.fieldToMatch.cookies.matchScope

`string` · required

Which parts to match against: "ALL" (names and values), "KEY" (names),
or "VALUE" (values).

- rule: {"required":true,"string":{"in":["ALL","KEY","VALUE"]}}

### spec.rules[].statement.regexMatch.fieldToMatch.cookies.oversizeHandling

`string` · required

What to do when the cookies exceed WAF's inspection limit (200 cookies /
8 KB): "CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexMatch.fieldToMatch.headerOrder

`AwsWafWebAclHeaderOrderMatch`

Inspect the comma-delimited list of header NAMES in receive order —
header-order fingerprinting of clients.

### spec.rules[].statement.regexMatch.fieldToMatch.headerOrder.oversizeHandling

`string` · required

What to do when the headers exceed WAF's inspection limit:
"CONTINUE", "MATCH", or "NO_MATCH". Required by AWS.

- rule: {"required":true,"string":{"in":["CONTINUE","MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexMatch.fieldToMatch.ja3Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA3 TLS-client fingerprint (a 32-character hash of the
TLS ClientHello — fingerprints the client software).

### spec.rules[].statement.regexMatch.fieldToMatch.ja3Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexMatch.fieldToMatch.ja4Fingerprint

`AwsWafWebAclFingerprintMatch`

Inspect the JA4 TLS-client fingerprint (the successor fingerprint
format to JA3).

### spec.rules[].statement.regexMatch.fieldToMatch.ja4Fingerprint.fallbackBehavior

`string` · required

What to do when the request has no fingerprint: "MATCH" or "NO_MATCH".

- rule: {"required":true,"string":{"in":["MATCH","NO_MATCH"]}}

### spec.rules[].statement.regexMatch.fieldToMatch.uriFragment

`AwsWafWebAclUriFragmentMatch`

Inspect the URI fragment (the part after '#'). Rarely sent by
browsers; useful for API clients that forward fragments.

### spec.rules[].statement.regexMatch.fieldToMatch.uriFragment.fallbackBehavior

`string`

What to do when the request has no fragment: "MATCH" or "NO_MATCH".
When unset, AWS treats missing fragments as NO_MATCH.

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH' when set

### spec.rules[].statement.regexMatch.textTransformations

`[]AwsWafWebAclTextTransformation` · required

Normalizations applied (in priority order) before matching. Use one
NONE transformation when no normalization is wanted.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].statement.regexMatch.textTransformations[].priority

`int32`

Order of this transformation relative to its siblings (lower first).
Priorities must be unique within one statement.

- rule: {"int32":{"gte":0}}

### spec.rules[].statement.regexMatch.textTransformations[].type

`string` · required

The transformation. "NONE" for raw bytes; common evasion-defeaters are
"URL_DECODE", "HTML_ENTITY_DECODE", "LOWERCASE", "COMPRESS_WHITE_SPACE",
and "CMD_LINE" (for command-injection inspection).

- rule: {"required":true,"string":{"in":["NONE","COMPRESS_WHITE_SPACE","HTML_ENTITY_DECODE","LOWERCASE","CMD_LINE","URL_DECODE","BASE64_DECODE","HEX_DECODE","MD5","REPLACE_COMMENTS","ESCAPE_SEQ_DECODE","SQL_HEX_DECODE","CSS_DECODE","JS_DECODE","NORMALIZE_PATH","NORMALIZE_PATH_WIN","REMOVE_NULLS","REPLACE_NULLS","BASE64_DECODE_EXT","URL_DECODE_UNI","UTF8_TO_UNICODE"]}}

### spec.rules[].statement.labelMatch

`AwsWafWebAclLabelMatchStatement`

Match labels added to the request by EARLIER rules (rule_labels) or by
managed rule groups — the mechanism for multi-stage evaluation
("managed group flagged it as bot AND it hit /checkout → block").

### spec.rules[].statement.labelMatch.scope

`string` · required

Match granularity: "LABEL" (the full label string) or "NAMESPACE" (any
label whose namespace prefix matches key).

- rule: {"required":true,"string":{"in":["LABEL","NAMESPACE"]}}

### spec.rules[].statement.labelMatch.key

`string` · required

The label (for scope LABEL, e.g. "awswaf:111122223333:rulegroup:testRules:label:testLabel")
or namespace prefix ending in ':' (for scope NAMESPACE). 1-1024
characters of letters, digits, hyphen, underscore, and ':' (the WAF
label syntax).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"1024","pattern":"^[0-9A-Za-z_\\-:]+$"}}

### spec.rules[].statement.asnMatch

`AwsWafWebAclAsnMatchStatement`

Match requests by the autonomous system number (ASN) of the source
address — block or rate-limit whole networks/hosting providers.

### spec.rules[].statement.asnMatch.asnList

`[]uint32` · required

The ASNs to match (1–100 entries), e.g. 64496. Typed uint32 — ASNs are
32-bit unsigned (4-byte ASNs reach 4,294,967,295), and protojson
stringifies 64-bit integers, which would corrupt the rule-JSON document
both engines build from this tree.

- rule: {"repeated":{"minItems":"1","maxItems":"100"}}

### spec.rules[].statement.asnMatch.forwardedIpConfig

`AwsWafWebAclForwardedIpConfig`

Optional forwarded IP configuration — derive the ASN from a forwarded
client IP instead of the source address.

- rule: fallback_behavior must be 'MATCH' or 'NO_MATCH'
- rule: position must be 'FIRST', 'LAST', or 'ANY' when set

### spec.rules[].statement.asnMatch.forwardedIpConfig.headerName

`string` · required

Name of the HTTP header containing the forwarded IP address.
Common values: "X-Forwarded-For", "X-Real-IP", "True-Client-IP".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.rules[].statement.asnMatch.forwardedIpConfig.fallbackBehavior

`string` · required

How to handle requests with missing or invalid forwarded IP headers.
- "MATCH": Treat as a match (allows the rule to process the request).
- "NO_MATCH": Treat as no match (skips the rule for this request).

- rule: {"required":true}

### spec.rules[].statement.asnMatch.forwardedIpConfig.position

`string`

Which IP address to use from multi-value forwarded headers.
Only applicable for IP Set reference rules (ignored by geo_match,
asn_match, and rate_based).

- "FIRST": Use the first IP in the header (closest to the client).
- "LAST": Use the last IP (closest to the server).
- "ANY": Match if any IP in the header matches the IP Set.

When omitted, defaults to "FIRST".

### spec.rules[].statement.andStatement

`AwsWafWebAclAndStatement`

Logical AND — matches when ALL nested statements match.

### spec.rules[].statement.andStatement.statements

`[]AwsWafWebAclStatement` · required

The statements to AND together. At least two (AND of one is the
statement itself).

- rule: {"repeated":{"minItems":"2"}}
- rule: recursive: same shape as enclosing AwsWafWebAclStatement

### spec.rules[].statement.orStatement

`AwsWafWebAclOrStatement`

Logical OR — matches when ANY nested statement matches.

### spec.rules[].statement.orStatement.statements

`[]AwsWafWebAclStatement` · required

The statements to OR together. At least two (OR of one is the statement
itself).

- rule: {"repeated":{"minItems":"2"}}
- rule: recursive: same shape as enclosing AwsWafWebAclStatement

### spec.rules[].statement.notStatement

`AwsWafWebAclNotStatement`

Logical NOT — matches when the nested statement does NOT match.

### spec.rules[].statement.notStatement.statement

`AwsWafWebAclStatement` · required

The statement to negate.

- rule: {"required":true}
- rule: recursive: same shape as enclosing AwsWafWebAclStatement

### spec.rules[].statement.customStatement

`object`

Escape hatch for any WAFv2 statement AWS ships before this spec models
it. Provide the raw AWS WAFv2 JSON statement structure as a Struct,
with PascalCase keys exactly as the AWS API expects.

Example:
  customStatement:
    SqliMatchStatement:
      FieldToMatch:
        Body: {}
      TextTransformations:
        - Priority: 0
          Type: URL_DECODE

### spec.rules[].action

`string`

Action for match rules (everything except managed_rule_group and
rule_group_reference). Specifies what happens when the rule matches.

Valid values: "allow", "block", "count", "captcha", "challenge".

Required for match rules. Must NOT be set for group rules.

### spec.rules[].overrideAction

`string`

Override action for group rules (managed_rule_group,
rule_group_reference). Controls whether the group's own actions are used
or overridden.

Valid values:
- "none": Use the rule group's configured actions as-is (most common).
- "count": Override all actions to count (useful for testing/monitoring
  a new rule group before enforcing).

Required for group rules. Must NOT be set for match rules.

### spec.rules[].customResponse

`AwsWafWebAclCustomResponse`

Custom response for block actions. Only valid when action is "block".
Specifies the HTTP response code and optional body.

### spec.rules[].customResponse.responseCode

`int32` · required

HTTP response status code to return. Range: 200-600.
Common values: 403 (Forbidden), 429 (Too Many Requests), 503 (Service Unavailable).

- rule: {"required":true,"int32":{"lte":600,"gte":200}}

### spec.rules[].customResponse.customResponseBodyKey

`string`

Key referencing a custom_response_body defined at the Web ACL level.
When set, the response body from the matching custom_response_body is
returned with the specified response_code and content type.

### spec.rules[].customResponse.responseHeaders

`[]AwsWafWebAclCustomHeader`

Additional HTTP headers to include in the block response.

### spec.rules[].customResponse.responseHeaders[].name

`string` · required

HTTP header name (case-insensitive). 1-64 characters: letters, digits,
and _ $ . - only. For INSERTED request headers, WAF prefixes the name
with "x-amzn-waf-" on the wire (name "sample" arrives as
"x-amzn-waf-sample") to avoid clobbering existing request headers;
response headers are sent under the name as given.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_$.-]+$"}}

### spec.rules[].customResponse.responseHeaders[].value

`string` · required

HTTP header value. 1-255 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"255"}}

### spec.rules[].customRequestHeaders

`[]AwsWafWebAclCustomHeader`

Custom request headers to insert for allow/count/captcha/challenge actions.
Headers are added to the request before forwarding to the protected resource.

### spec.rules[].customRequestHeaders[].name

`string` · required

HTTP header name (case-insensitive). 1-64 characters: letters, digits,
and _ $ . - only. For INSERTED request headers, WAF prefixes the name
with "x-amzn-waf-" on the wire (name "sample" arrives as
"x-amzn-waf-sample") to avoid clobbering existing request headers;
response headers are sent under the name as given.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z_$.-]+$"}}

### spec.rules[].customRequestHeaders[].value

`string` · required

HTTP header value. 1-255 characters.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"255"}}

### spec.rules[].ruleLabels

`[]string`

Labels added to requests that match this rule. Labels are key-value pairs
(namespace:name format) that can be matched by label_match statements in
subsequent rules, enabling multi-stage rule evaluation. Each label is
1-1024 characters of letters, digits, hyphen, underscore, and ':'
namespace separators (the WAF label syntax).

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"1024","pattern":"^[0-9A-Za-z_\\-:]+$"}}}}

### spec.rules[].visibilityConfig

`AwsWafWebAclVisibilityConfig`

CloudWatch metrics configuration for this rule. When omitted, the IaC
module applies sensible defaults:
- cloudwatch_metrics_enabled = true
- sampled_requests_enabled = true
- metric_name = rule name

### spec.rules[].visibilityConfig.cloudwatchMetricsEnabled

`bool`

Enable CloudWatch metrics for this Web ACL or rule.
Default: true (applied by IaC module when omitted).

### spec.rules[].visibilityConfig.sampledRequestsEnabled

`bool`

Enable request sampling for this Web ACL or rule. Sampled requests are
viewable in the AWS WAF console for debugging rule matches.
Default: true (applied by IaC module when omitted).

### spec.rules[].visibilityConfig.metricName

`string`

CloudWatch metric name for this Web ACL or rule. Must be unique within
the Web ACL's rules. 1-128 characters: letters, digits, hyphen,
underscore only; "All" and "Default_Action" are reserved by WAF.
Default: resource name (Web ACL) or rule name (applied by IaC module).

- rule: metric_name must be 1-128 characters of A-Za-z0-9_- and must not be the reserved 'All' or 'Default_Action' when set

### spec.rules[].captchaConfig

`AwsWafWebAclImmunityTimeConfig`

Per-rule CAPTCHA immunity time override (seconds a solved CAPTCHA stays
valid for requests matching THIS rule). Overrides the web ACL's
captcha_config. Only meaningful for rules whose action (or a managed
group's internal rules) can serve a CAPTCHA.

### spec.rules[].captchaConfig.immunityTimeSec

`int32` · required

Immunity time in seconds. CAPTCHA: 60–259,200 (AWS default 300).
Challenge: 300–259,200 (AWS default 300). Typed int32 (the max fits) —
protojson stringifies 64-bit integers, which would corrupt the rule-JSON
document for the per-rule configs living inside the rules subtree.

- rule: {"required":true,"int32":{"lte":259200,"gte":60}}

### spec.rules[].challengeConfig

`AwsWafWebAclImmunityTimeConfig`

Per-rule challenge immunity time override (seconds a passed silent
challenge stays valid for requests matching THIS rule). Overrides the
web ACL's challenge_config.

### spec.rules[].challengeConfig.immunityTimeSec

`int32` · required

Immunity time in seconds. CAPTCHA: 60–259,200 (AWS default 300).
Challenge: 300–259,200 (AWS default 300). Typed int32 (the max fits) —
protojson stringifies 64-bit integers, which would corrupt the rule-JSON
document for the per-rule configs living inside the rules subtree.

- rule: {"required":true,"int32":{"lte":259200,"gte":60}}

### spec.visibilityConfig

`AwsWafWebAclVisibilityConfig`

CloudWatch metrics configuration for the Web ACL itself.

When omitted, the IaC module applies sensible defaults:
- cloudwatch_metrics_enabled = true
- sampled_requests_enabled = true
- metric_name = resource name

### spec.visibilityConfig.cloudwatchMetricsEnabled

`bool`

Enable CloudWatch metrics for this Web ACL or rule.
Default: true (applied by IaC module when omitted).

### spec.visibilityConfig.sampledRequestsEnabled

`bool`

Enable request sampling for this Web ACL or rule. Sampled requests are
viewable in the AWS WAF console for debugging rule matches.
Default: true (applied by IaC module when omitted).

### spec.visibilityConfig.metricName

`string`

CloudWatch metric name for this Web ACL or rule. Must be unique within
the Web ACL's rules. 1-128 characters: letters, digits, hyphen,
underscore only; "All" and "Default_Action" are reserved by WAF.
Default: resource name (Web ACL) or rule name (applied by IaC module).

- rule: metric_name must be 1-128 characters of A-Za-z0-9_- and must not be the reserved 'All' or 'Default_Action' when set

### spec.customResponseBodies

`[]AwsWafWebAclCustomResponseBody`

Reusable response body templates that can be referenced by block actions
via custom_response_body_key. Define branded error pages or structured
error responses here and reference them by key from individual rules.

Each entry requires a unique key (used as the reference), content string,
and content_type.

- rule: content_type must be 'TEXT_PLAIN', 'TEXT_HTML', or 'APPLICATION_JSON'

### spec.customResponseBodies[].key

`string` · required

Unique key used to reference this response body from custom_response
configurations. Must be unique within the Web ACL. 1-128 characters:
letters, digits, underscore, hyphen.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128","pattern":"^[\\w\\-]+$"}}

### spec.customResponseBodies[].content

`string` · required

The response body content (HTML, plain text, or JSON). Max 10,240 bytes.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"10240"}}

### spec.customResponseBodies[].contentType

`string` · required

MIME type of the content.
Valid values: "TEXT_PLAIN", "TEXT_HTML", "APPLICATION_JSON".

- rule: {"required":true}

### spec.tokenDomains

`[]string`

Domains to accept in web request tokens for CAPTCHA and Challenge actions.
Required when using CAPTCHA/Challenge with multiple domains that share
the same Web ACL. When omitted, tokens are scoped to the request domain.
Each entry is a domain name (1-253 characters; letters, digits, dots,
hyphens, slashes). Public suffixes (e.g. "co.uk", "gov.au") are rejected
by AWS.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"253","pattern":"^[\\w\\.\\-/]+$"}}}}

### spec.captchaConfig

`AwsWafWebAclImmunityTimeConfig`

How long a client's successful CAPTCHA solve remains valid, web-ACL-wide
(immunity time in seconds, 60–259,200; AWS default 300). Rules can
override this per rule via the rule's captcha_config.

### spec.captchaConfig.immunityTimeSec

`int32` · required

Immunity time in seconds. CAPTCHA: 60–259,200 (AWS default 300).
Challenge: 300–259,200 (AWS default 300). Typed int32 (the max fits) —
protojson stringifies 64-bit integers, which would corrupt the rule-JSON
document for the per-rule configs living inside the rules subtree.

- rule: {"required":true,"int32":{"lte":259200,"gte":60}}

### spec.challengeConfig

`AwsWafWebAclImmunityTimeConfig`

How long a client's successful silent-challenge response remains valid,
web-ACL-wide (immunity time in seconds, 300–259,200; AWS default 300).
Rules can override this per rule via the rule's challenge_config.

### spec.challengeConfig.immunityTimeSec

`int32` · required

Immunity time in seconds. CAPTCHA: 60–259,200 (AWS default 300).
Challenge: 300–259,200 (AWS default 300). Typed int32 (the max fits) —
protojson stringifies 64-bit integers, which would corrupt the rule-JSON
document for the per-rule configs living inside the rules subtree.

- rule: {"required":true,"int32":{"lte":259200,"gte":60}}

### spec.associationConfig

`AwsWafWebAclAssociationConfig`

Per-resource-type request BODY inspection size limits. By default WAF
inspects only the first 16 KB of a request body; raise the limit (to 32,
48, or 64 KB) for APIs that carry large JSON payloads whose tail must
still be inspected. Larger limits increase WCU cost of body-inspecting
rules. CloudFront also supports raising this; the other resource types
are capped per their entry.

### spec.associationConfig.cloudfrontRequestBodyLimit

`string`

Body inspection limit for CloudFront distributions ("KB_16", "KB_32",
"KB_48", or "KB_64").

- rule: the request body limit must be 'KB_16', 'KB_32', 'KB_48', or 'KB_64' when set

### spec.associationConfig.apiGatewayRequestBodyLimit

`string`

Body inspection limit for API Gateway REST APIs.

- rule: the request body limit must be 'KB_16', 'KB_32', 'KB_48', or 'KB_64' when set

### spec.associationConfig.cognitoUserPoolRequestBodyLimit

`string`

Body inspection limit for Cognito user pools.

- rule: the request body limit must be 'KB_16', 'KB_32', 'KB_48', or 'KB_64' when set

### spec.associationConfig.appRunnerServiceRequestBodyLimit

`string`

Body inspection limit for App Runner services.

- rule: the request body limit must be 'KB_16', 'KB_32', 'KB_48', or 'KB_64' when set

### spec.associationConfig.verifiedAccessInstanceRequestBodyLimit

`string`

Body inspection limit for Verified Access instances.

- rule: the request body limit must be 'KB_16', 'KB_32', 'KB_48', or 'KB_64' when set

### spec.dataProtectionConfig

`AwsWafWebAclDataProtectionConfig`

Field-level data protection: replace or hash specified request fields
(headers, cookies, query strings, body) in ALL WAF outputs — logs,
sampled requests, and rule match details — before they leave WAF. This
is stronger than the logging block's redacted fields (which only affect
the logging destination): use it for PII that must never appear anywhere.

### spec.dataProtectionConfig.dataProtections

`[]AwsWafWebAclDataProtection` · required

The fields to protect (up to 26 entries).

- rule: {"repeated":{"minItems":"1","maxItems":"26"}}
- rule: field_keys is only valid for SINGLE_HEADER/SINGLE_COOKIE/SINGLE_QUERY_ARGUMENT (omit it to protect all keys of the field type)

### spec.dataProtectionConfig.dataProtections[].fieldType

`string` · required

The field class to protect: "SINGLE_HEADER", "SINGLE_COOKIE",
"SINGLE_QUERY_ARGUMENT", "QUERY_STRING", or "BODY".

- rule: {"required":true,"string":{"in":["SINGLE_HEADER","SINGLE_COOKIE","SINGLE_QUERY_ARGUMENT","QUERY_STRING","BODY"]}}

### spec.dataProtectionConfig.dataProtections[].fieldKeys

`[]string`

The specific keys to protect for the SINGLE_* field types (header names,
cookie names, query-argument names; up to 100, each 1-64 characters).
OMITTING keys for a SINGLE_* type protects ALL keys of that type (the
AWS contract: "If you don't specify any key, then all keys for the
field type are protected"). Not applicable to QUERY_STRING/BODY (which
protect the whole component).

- rule: {"repeated":{"maxItems":"100","items":{"string":{"minLen":"1","maxLen":"64"}}}}

### spec.dataProtectionConfig.dataProtections[].action

`string` · required

How to mask: "SUBSTITUTION" (replace with a fixed placeholder) or
"HASH" (replace with a one-way hash, preserving correlate-ability).

- rule: {"required":true,"string":{"in":["SUBSTITUTION","HASH"]}}

### spec.dataProtectionConfig.dataProtections[].excludeRuleMatchDetails

`bool`

Also exclude this field from rule MATCH details in logs.

### spec.dataProtectionConfig.dataProtections[].excludeRateBasedDetails

`bool`

Also exclude this field from rate-based rule details in logs.

### spec.logging

`AwsWafWebAclLoggingConfig`

Optional logging configuration. When provided, WAF sends detailed request
logs to the specified destination (CloudWatch Logs, S3, or Kinesis
Firehose).

Important naming constraint: the destination resource name must start with
"aws-waf-logs-" (enforced by AWS). For example:
- CloudWatch Log Group: "aws-waf-logs-my-acl"
- S3 Bucket: "aws-waf-logs-my-acl"
- Kinesis Firehose: "aws-waf-logs-my-acl"

### spec.logging.destinationArn

`string | valueFrom` · required

ARN of the logging destination. Must be one of:
- CloudWatch Logs log group ARN
- S3 bucket ARN
- Kinesis Firehose delivery stream ARN

The destination resource name must start with "aws-waf-logs-".

Deliberately singular: AWS allows exactly one logging destination per
web ACL (the Terraform provider's log_destination_configs argument
nominally accepts up to 100 entries, but the service contract — SDK:
"You can associate one logging destination to a web ACL" — is one).

No default_kind is set because the destination can be any of three
different resource types; the three are declared as candidates, each
with its ARN output.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`), AwsS3Bucket (`status.outputs.bucket_arn`), AwsKinesisFirehose (`status.outputs.delivery_stream_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.logging.redactedHeaderNames

`[]string`

HTTP header names to redact from logs (each 1-64 characters). Redacted
headers appear as "REDACTED" in log entries instead of their actual
values.

Common headers to redact: "Authorization", "Cookie", "X-Api-Key".

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"64"}}}}

### spec.logging.redactUriPath

`bool`

Redact the URI path from log entries. When true, the URI path appears
as "REDACTED" in logs.

### spec.logging.redactQueryString

`bool`

Redact the query string from log entries. When true, query string
parameters appear as "REDACTED" in logs.

### spec.logging.redactMethod

`bool`

Redact the HTTP method from log entries. When true, the method appears
as "REDACTED" in logs. (Method, query string, URI path, and single
headers are the only components AWS supports redacting.)

### spec.logging.filter

`AwsWafWebAclLoggingFilterConfig`

Optional log filtering: keep or drop log records by the action WAF
applied or by labels on the request, instead of logging every inspected
request. Typical use: keep only BLOCK and COUNT records to cut logging
cost on high-traffic ACLs.

### spec.logging.filter.defaultBehavior

`string` · required

What happens to log records that match NONE of the filters:
"KEEP" (log them) or "DROP" (discard them).

- rule: {"required":true,"string":{"in":["KEEP","DROP"]}}

### spec.logging.filter.filters

`[]AwsWafWebAclLoggingFilter` · required

The filters, each with its own keep/drop behavior. At least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.logging.filter.filters[].behavior

`string` · required

What happens to log records matching this filter: "KEEP" or "DROP".

- rule: {"required":true,"string":{"in":["KEEP","DROP"]}}

### spec.logging.filter.filters[].requirement

`string` · required

How the conditions combine: "MEETS_ALL" (every condition must match)
or "MEETS_ANY" (at least one condition matches).

- rule: {"required":true,"string":{"in":["MEETS_ALL","MEETS_ANY"]}}

### spec.logging.filter.filters[].conditions

`[]AwsWafWebAclLoggingFilterCondition` · required

The match conditions. At least one.

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of action or label_name per condition

### spec.logging.filter.filters[].conditions[].action

`string`

Match log records by the action WAF applied to the request:
"ALLOW", "BLOCK", "COUNT", "CAPTCHA", "CHALLENGE",
"EXCLUDED_AS_COUNT" (rules overridden to count — the tuning-noise
filter), or "MONETIZE" (marketplace rule groups only).

- rule: action must be 'ALLOW', 'BLOCK', 'COUNT', 'CAPTCHA', 'CHALLENGE', 'EXCLUDED_AS_COUNT', or 'MONETIZE' when set

### spec.logging.filter.filters[].conditions[].labelName

`string`

Match log records carrying this label. Must be the FULLY QUALIFIED
label name including the prefix, e.g.
"awswaf:managed:aws:bot-control:bot:category:monitoring".

- rule: {"string":{"maxLen":"1024"}}

## Validation Rules

- `scope_valid`: scope must be 'REGIONAL' or 'CLOUDFRONT'
- `cloudfront_scope_requires_us_east_1`: CloudFront-scoped web ACLs live in the global (us-east-1) region — set region to 'us-east-1' when scope is 'CLOUDFRONT'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsWafWebAcl, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.web_acl_arn` | `string` | The Amazon Resource Name (ARN) of the Web ACL. This is the primary output used to associate the Web ACL with ALB, API Gateway, CloudFront, AppSync, Cognito, or App Runner resources. |
| `status.outputs.web_acl_id` | `string` | The unique identifier of the Web ACL. |
| `status.outputs.web_acl_name` | `string` | The name of the Web ACL as specified in metadata. |
| `status.outputs.capacity` | `int32` | The Web ACL Capacity Units (WCUs) consumed by all rules in this Web ACL. The default account limit is 5,000 WCUs per Web ACL. Use this output to monitor capacity usage and plan rule additions. |
| `status.outputs.application_integration_url` | `string` | The URL to use in your application's client integration when the Web ACL serves CAPTCHA or Challenge actions (the AWS WAF JavaScript integration endpoint). Empty when the ACL uses neither action. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.rules[].statement.ipSetReference.arn` | AwsWafIpSet | `status.outputs.ip_set_arn` |
| `spec.rules[].statement.regexPatternSetReference.arn` | AwsWafRegexPatternSet | `status.outputs.regex_pattern_set_arn` |
| `spec.logging.destinationArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.logging.destinationArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.logging.destinationArn` | AwsKinesisFirehose | `status.outputs.delivery_stream_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAlb | `spec.webAclArn` | `status.outputs.web_acl_arn` |
| AwsAppRunnerService | `spec.webAclArn` | `status.outputs.web_acl_arn` |
| AwsAppSyncApi | `spec.graphql.webAclArn` | `status.outputs.web_acl_arn` |
| AwsCloudFront | `spec.webAclArn` | `status.outputs.web_acl_arn` |

## See Also

- [Overview](../README.md)
