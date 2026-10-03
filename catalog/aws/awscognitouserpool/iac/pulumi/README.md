# AwsCognitoUserPool — Pulumi Module

## Overview

This Pulumi (Go) module provisions an Amazon Cognito User Pool at full provider depth, plus the two pool-scoped folded satellites: the hosted-UI domain (one per pool) and log delivery. It is behaviorally identical to the Terraform module — same resources, same names, same outputs.

## Module Structure

```
main.go                — entrypoint (loads IaC input, runs module.Resources)
module/main.go         — orchestration: provider → pool → domain → log delivery
module/user_pool.go    — cognito.UserPool with the full spec surface
module/domain.go       — cognito.UserPoolDomain + domain join-key exports
module/log_delivery.go — cognito.LogDeliveryConfiguration (single pool-scoped resource)
module/locals.go       — identity tags
module/outputs.go      — output name constants
```

## Deploy

```bash
# from iac/pulumi/ with an iac-input.yaml present
make preview
make up
make destroy
```

## Notes

- The pool's cloud name is `metadata.name` — the cross-engine naming basis.
- `issuer` is exported with the `https://` scheme (the JWT-authorizer join key); `user_pool_domain` is the RAW domain string (the ALB authenticate-cognito join key); `hosted_ui_url` is the full sign-in URL.
- CloudFront alias outputs are populated only for custom domains — for prefix domains they are honestly empty.
