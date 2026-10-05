# AWS LB Listener

Deploys an ELBv2 listener — the port/protocol entry point on a load balancer, owning the TLS material clients see and the default action taken when no listener rule matches. A listener is a first-class node in the routing graph: one load balancer carries many listeners (80 and 443 at minimum on most ALBs), and listener rules attach to a specific listener as services deploy. The same kind serves both families — ALB listeners (HTTP/HTTPS) take the full action set including authentication; NLB listeners (TCP/UDP/TCP_UDP/TLS/QUIC/TCP_QUIC) forward only.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Listener** -- on the referenced load balancer, with its port, protocol, and default-action chain
- **Certificate attachments** -- the default certificate plus any SNI certificates (HTTPS/TLS listeners)
- **Mutual TLS configuration** -- when configured on an ALB HTTPS listener
- **Header policy attributes** -- edge-served CORS/security response headers and TLS/mTLS request-header injection (ALB listeners)

The load balancer, target groups, and listener rules are separate components — rules attach through this listener's `listener_arn` output.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Load Balancer** -- an AwsAlb (or AwsNlb) referenced by its `load_balancer_arn` output; the wizard's family toggle swaps the reference target.
- **Certificates** -- AwsCertManagerCert resources referenced by their `cert_arn` outputs for HTTPS/TLS listeners.
- **Target Groups** -- AwsLbTargetGroup resources for the forward actions, referenced by their `target_group_arn` outputs.

### AWS Account

- **ELB permissions** -- the credentials used by the Provider Connection must have `elasticloadbalancing:CreateListener`, `ModifyListener`, `AddListenerCertificates`, and `DeleteListener`, plus `iam:CreateServiceLinkedRole` on first use.
- **Certificate region** -- ACM certificates must live in the listener's own region.
- **OIDC secret** -- an authenticate-oidc action's client secret is a managed secret reference (`$secret/<slug>`); create the org secret before deploying.

## Deploy

### Console

Open the deployment store, find **AWS LB Listener**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **HTTPS Forward** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLbListener
metadata:
  name: https-listener
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  loadBalancerArn:
    valueFrom:
      kind: AwsAlb
      name: api-gateway
      fieldPath: status.outputs.load_balancer_arn
  port: 443
  protocol: HTTPS
  certificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: wildcard-cert
      fieldPath: status.outputs.cert_arn
  defaultActions:
    - type: forward
      forward:
        targetGroups:
          - arn:
              valueFrom:
                kind: AwsLbTargetGroup
                name: web-servers
                fieldPath: status.outputs.target_group_arn
```

```shell
planton apply -f listener.yaml
```

This attaches a TLS-terminating HTTPS listener to the ALB, forwarding everything to the web-servers group. An Infra Job tracks the provisioning in real time.

### InfraChart

When the listener deploys alongside its load balancer, certificate, and target group in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  loadBalancerArn:
    valueFrom:
      kind: AwsAlb
      name: api-gateway
      fieldPath: status.outputs.load_balancer_arn
  port: 443
  protocol: HTTPS
  certificateArn:
    valueFrom:
      kind: AwsCertManagerCert
      name: wildcard-cert
      fieldPath: status.outputs.cert_arn
  defaultActions:
    - type: forward
      forward:
        targetGroups:
          - arn:
              valueFrom:
                kind: AwsLbTargetGroup
                name: web-servers
                fieldPath: status.outputs.target_group_arn
```

The InfraPipeline resolves the dependency graph, deploys the load balancer, certificate, and target group first, then attaches the listener.

## Key Configuration

These are the most important decisions when configuring a listener. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The attachment is a one-way door** -- moving a listener to another load balancer replaces it. Port, protocol, certificates, and actions all update in place.

**The protocol decides everything** -- HTTP/HTTPS listeners take the full action set (forward, redirect, fixed-response, Cognito, OIDC, JWT validation) plus mTLS and header policy; NLB protocols (TCP/UDP/TCP_UDP/TLS, and QUIC/TCP_QUIC for HTTP/3 with TCP fallback) forward to exactly one target group — AWS rejects everything else at Layer 4.

**Certificates live here, not on the load balancer** -- the default certificate serves clients matching no SNI certificate; additional certificates multiplex several domains on one listener. Reference AwsCertManagerCert outputs so ACM renewals never go stale.

**Default actions are the unmatched-request answer** -- a chain may front the terminal action with authentication, and every chain ends in exactly one forward, redirect, or fixed response. A listener whose real traffic flows through rules often defaults to an explicit fixed-response 404.

**Authentication without application code** -- authenticate-oidc puts corporate SSO in front of anything behind the listener; jwt-validation rejects requests without a valid bearer token at the edge. The OIDC client secret is reference-only, resolved just-in-time at deploy.

**Weighted forwards enable canaries** -- up to 5 target groups with weights (90/10 …), plus group stickiness so sessions do not flap between blue and green.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsAlb** (default) or **AwsNlb** | `loadBalancerArn` | `status.outputs.load_balancer_arn` |
| **AwsCertManagerCert** | `certificateArn` / `additionalCertificateArns[]` | `status.outputs.cert_arn` |
| **AwsLbTargetGroup** | `defaultActions[].forward.targetGroups[].arn` | `status.outputs.target_group_arn` |
| **AwsCognitoUserPool** | `defaultActions[].authenticateCognito.userPoolArn` / `.userPoolDomain` | `status.outputs.user_pool_arn` / `status.outputs.user_pool_domain` |
| **AwsCognitoUserPoolClient** | `defaultActions[].authenticateCognito.userPoolClientId` | `status.outputs.client_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `listener_arn` | ARN of the listener | AwsLbListenerRule resources attach through it |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**TLS termination + forward** -- the production 443 listener. Start from the **HTTPS Forward** preset.

**HTTP→HTTPS redirect** -- the one-action port-80 listener every HTTPS site pairs with. Start from the **HTTP Redirect to HTTPS** preset.

**SSO in front of an internal tool** -- an authenticate-oidc action ahead of the forward. Start from the **OIDC-Protected HTTPS** preset.

## Works With

- [**AWS ALB**](/infra-catalog/aws-alb) -- the Layer-7 load balancer this listener attaches to, referenced by `loadBalancerArn`.
- [**AWS NLB**](/infra-catalog/aws-nlb) -- the Layer-4 alternative attachment, taking forward-only protocols.
- [**AWS LB Listener Rule**](/infra-catalog/aws-lb-listener-rule) -- per-service routing attached through this listener's `listener_arn` output.
- [**AWS LB Target Group**](/infra-catalog/aws-lb-target-group) -- the forward destinations, referenced per action.
- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) -- the TLS material, referenced by `certificateArn`.
- [**AWS Cognito User Pool**](/infra-catalog/aws-cognito-user-pool) -- the user pool behind authenticate-cognito actions.
