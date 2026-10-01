# KubernetesClusterIssuer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesClusterIssuerSpec** creates one cert-manager ClusterIssuer — a
cluster-scoped certificate authority front-end that Certificate resources in
ANY namespace can request signed certificates from. The ClusterIssuer is
named after this resource (`metadata.name`); Certificates select it by that
name, and ingress-shim annotations (`cert-manager.io/cluster-issuer`) use
the same name.

ClusterIssuer vs Issuer: identical signing capabilities (the config message
is shared), different scope. Use a ClusterIssuer for the platform-wide CA
(one "letsencrypt-production" serving every team); use a namespace-scoped
KubernetesIssuer when a namespace needs its own CA or its own credentials
blast radius.

Requires cert-manager (KubernetesCertManager) on the cluster — the
ClusterIssuer is a cert-manager custom resource, and credential Secrets
created for it must live in cert-manager's cluster-resource namespace,
which is where the controller reads secrets for cluster-scoped resources.

## Example

```yaml
# Full-surface manifest for offline module proofs (tofu validate/plan and
# pulumi preview). Exercises the ACME backend with EAB, selector-scoped
# solvers across DNS-01 providers, and an HTTP-01 ingress solver; both
# engines must render identical CR + credential Secrets from it.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClusterIssuer
metadata:
  name: test-cluster-issuer
spec:
  certManagerNamespace:
    value: cert-manager
  config:
    acme:
      email: platform@example.com
      server: https://acme-staging-v02.api.letsencrypt.org/directory
      profile: tlsserver
      preferredChain: ISRG Root X1
      skipTlsVerify: false
      externalAccountBinding:
        keyId: eab-key-id-123
        hmacKey: dGVzdC1obWFjLWtleQ
      disableAccountKeyGeneration: false
      enableDurationFeature: true
      solvers:
        - selector:
            dnsZones:
              - example.com
          dns01:
            cnameStrategy: follow
            cloudflare:
              apiToken:
                token: test-cloudflare-token
        - selector:
            dnsZones:
              - example.org
            matchLabels:
              use-route53: "true"
          dns01:
            route53:
              region: us-east-1
              hostedZoneId: Z2E9THH2AEXAMPLE
              assumeRoleArn: arn:aws:iam::123456789012:role/dns01-cross-account
              staticCredentials:
                accessKeyId: AKIAEXAMPLE
                secretAccessKey: test-secret-access-key
        - selector:
            dnsNames:
              - www.example.net
          http01:
            ingress:
              ingressClassName: nginx
              serviceType: ClusterIP
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.certManagerNamespace` | `string \| valueFrom` | yes |  | KubernetesCertManager (`status.outputs.cluster_resource_namespace`) |
| `spec.config` | `CertManagerIssuerConfig` | yes |  |  |
| `spec.config.acme` | `CertManagerAcmeConfig` |  |  |  |
| `spec.config.acme.email` | `string` | yes |  |  |
| `spec.config.acme.server` | `string` |  | `https://acme-v02.api.letsencrypt.org/directory` |  |
| `spec.config.acme.profile` | `string` |  |  |  |
| `spec.config.acme.preferredChain` | `string` |  |  |  |
| `spec.config.acme.caBundle` | `string` |  |  |  |
| `spec.config.acme.skipTlsVerify` | `bool` |  |  |  |
| `spec.config.acme.externalAccountBinding` | `CertManagerAcmeExternalAccountBinding` |  |  |  |
| `spec.config.acme.externalAccountBinding.keyId` | `string` | yes |  |  |
| `spec.config.acme.externalAccountBinding.hmacKey` | `string` (sensitive) | yes |  |  |
| `spec.config.acme.disableAccountKeyGeneration` | `bool` |  |  |  |
| `spec.config.acme.enableDurationFeature` | `bool` |  |  |  |
| `spec.config.acme.solvers` | `[]CertManagerAcmeSolver` | yes |  |  |
| `spec.config.acme.solvers[].selector` | `CertManagerAcmeSolverSelector` |  |  |  |
| `spec.config.acme.solvers[].selector.dnsZones` | `[]string` |  |  |  |
| `spec.config.acme.solvers[].selector.dnsNames` | `[]string` |  |  |  |
| `spec.config.acme.solvers[].selector.matchLabels` | `map<string, string>` |  |  |  |
| `spec.config.acme.solvers[].http01` | `CertManagerAcmeHttp01Solver` |  |  |  |
| `spec.config.acme.solvers[].http01.ingress` | `CertManagerAcmeHttp01IngressSolver` |  |  |  |
| `spec.config.acme.solvers[].http01.ingress.ingressClassName` | `string` |  |  |  |
| `spec.config.acme.solvers[].http01.ingress.name` | `string` |  |  |  |
| `spec.config.acme.solvers[].http01.ingress.serviceType` | `string` |  |  |  |
| `spec.config.acme.solvers[].http01.gatewayHttpRoute` | `CertManagerAcmeHttp01GatewaySolver` |  |  |  |
| `spec.config.acme.solvers[].http01.gatewayHttpRoute.parentRefs` | `[]CertManagerGatewayParentRef` | yes |  |  |
| `spec.config.acme.solvers[].http01.gatewayHttpRoute.parentRefs[].name` | `string` | yes |  |  |
| `spec.config.acme.solvers[].http01.gatewayHttpRoute.parentRefs[].namespace` | `string` |  |  |  |
| `spec.config.acme.solvers[].http01.gatewayHttpRoute.parentRefs[].sectionName` | `string` |  |  |  |
| `spec.config.acme.solvers[].http01.gatewayHttpRoute.labels` | `map<string, string>` |  |  |  |
| `spec.config.acme.solvers[].http01.gatewayHttpRoute.serviceType` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01` | `CertManagerAcmeDns01Solver` |  |  |  |
| `spec.config.acme.solvers[].dns01.cnameStrategy` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.cloudflare` | `CertManagerDns01Cloudflare` |  |  |  |
| `spec.config.acme.solvers[].dns01.cloudflare.apiToken` | `CertManagerCloudflareApiToken` |  |  |  |
| `spec.config.acme.solvers[].dns01.cloudflare.apiToken.token` | `string` (sensitive) | yes |  |  |
| `spec.config.acme.solvers[].dns01.cloudflare.apiKey` | `CertManagerCloudflareApiKey` |  |  |  |
| `spec.config.acme.solvers[].dns01.cloudflare.apiKey.email` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.cloudflare.apiKey.key` | `string` (sensitive) | yes |  |  |
| `spec.config.acme.solvers[].dns01.route53` | `CertManagerDns01Route53` |  |  |  |
| `spec.config.acme.solvers[].dns01.route53.region` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.route53.hostedZoneId` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.route53.assumeRoleArn` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.route53.staticCredentials` | `CertManagerRoute53StaticCredentials` |  |  |  |
| `spec.config.acme.solvers[].dns01.route53.staticCredentials.accessKeyId` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.route53.staticCredentials.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.config.acme.solvers[].dns01.route53.serviceAccount` | `CertManagerRoute53ServiceAccountAuth` |  |  |  |
| `spec.config.acme.solvers[].dns01.route53.serviceAccount.serviceAccountName` | `string \| valueFrom` | yes |  | KubernetesServiceAccount (`status.outputs.service_account_name`) |
| `spec.config.acme.solvers[].dns01.route53.serviceAccount.audiences` | `[]string` |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns` | `CertManagerDns01AzureDns` |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.subscriptionId` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.resourceGroupName` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.hostedZoneName` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.zoneType` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.environment` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.clientId` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.tenantId` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.managedIdentity` | `CertManagerAzureManagedIdentity` |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.managedIdentity.clientId` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.azureDns.managedIdentity.resourceId` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.gcpCloudDns` | `CertManagerDns01GcpCloudDns` |  |  |  |
| `spec.config.acme.solvers[].dns01.gcpCloudDns.projectId` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.gcpCloudDns.hostedZoneName` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.gcpCloudDns.serviceAccountKeyJson` | `string` (sensitive) |  |  |  |
| `spec.config.acme.solvers[].dns01.digitalocean` | `CertManagerDns01DigitalOcean` |  |  |  |
| `spec.config.acme.solvers[].dns01.digitalocean.token` | `string` (sensitive) | yes |  |  |
| `spec.config.acme.solvers[].dns01.rfc2136` | `CertManagerDns01Rfc2136` |  |  |  |
| `spec.config.acme.solvers[].dns01.rfc2136.nameserver` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.rfc2136.tsigKeyName` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.rfc2136.tsigAlgorithm` | `string` |  |  |  |
| `spec.config.acme.solvers[].dns01.rfc2136.tsigSecret` | `string` (sensitive) |  |  |  |
| `spec.config.acme.solvers[].dns01.acmeDns` | `CertManagerDns01AcmeDns` |  |  |  |
| `spec.config.acme.solvers[].dns01.acmeDns.host` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.acmeDns.accountJson` | `string` (sensitive) | yes |  |  |
| `spec.config.acme.solvers[].dns01.akamai` | `CertManagerDns01Akamai` |  |  |  |
| `spec.config.acme.solvers[].dns01.akamai.serviceConsumerDomain` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.akamai.clientToken` | `string` (sensitive) | yes |  |  |
| `spec.config.acme.solvers[].dns01.akamai.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.config.acme.solvers[].dns01.akamai.accessToken` | `string` (sensitive) | yes |  |  |
| `spec.config.acme.solvers[].dns01.webhook` | `CertManagerDns01Webhook` |  |  |  |
| `spec.config.acme.solvers[].dns01.webhook.groupName` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.webhook.solverName` | `string` | yes |  |  |
| `spec.config.acme.solvers[].dns01.webhook.configYaml` | `string` |  |  |  |
| `spec.config.ca` | `CertManagerCaConfig` |  |  |  |
| `spec.config.ca.caSecretName` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.config.ca.crlDistributionPoints` | `[]string` |  |  |  |
| `spec.config.ca.ocspServers` | `[]string` |  |  |  |
| `spec.config.ca.issuingCertificateUrls` | `[]string` |  |  |  |
| `spec.config.selfSigned` | `CertManagerSelfSignedConfig` |  |  |  |
| `spec.config.selfSigned.crlDistributionPoints` | `[]string` |  |  |  |
| `spec.config.vault` | `CertManagerVaultConfig` |  |  |  |
| `spec.config.vault.server` | `string` | yes |  |  |
| `spec.config.vault.path` | `string` | yes |  |  |
| `spec.config.vault.vaultNamespace` | `string` |  |  |  |
| `spec.config.vault.caBundle` | `string` |  |  |  |
| `spec.config.vault.serverName` | `string` |  |  |  |
| `spec.config.vault.tokenAuth` | `CertManagerVaultTokenAuth` |  |  |  |
| `spec.config.vault.tokenAuth.token` | `string` (sensitive) | yes |  |  |
| `spec.config.vault.appRoleAuth` | `CertManagerVaultAppRoleAuth` |  |  |  |
| `spec.config.vault.appRoleAuth.path` | `string` | yes |  |  |
| `spec.config.vault.appRoleAuth.roleId` | `string` | yes |  |  |
| `spec.config.vault.appRoleAuth.secretId` | `string` (sensitive) | yes |  |  |
| `spec.config.vault.kubernetesAuth` | `CertManagerVaultKubernetesAuth` |  |  |  |
| `spec.config.vault.kubernetesAuth.role` | `string` | yes |  |  |
| `spec.config.vault.kubernetesAuth.mountPath` | `string` |  | `kubernetes` |  |
| `spec.config.vault.kubernetesAuth.serviceAccountName` | `string \| valueFrom` | yes |  | KubernetesServiceAccount (`status.outputs.service_account_name`) |
| `spec.config.vault.kubernetesAuth.audiences` | `[]string` |  |  |  |

## Field Details

### spec.certManagerNamespace

`string | valueFrom` · required

Namespace where cert-manager reads Secrets for cluster-scoped resources
(its "cluster resource namespace" — by default the namespace cert-manager
is installed in). Credential Secrets this resource materializes (API
tokens, static keys) are created here, and cert-manager stores ACME
account keys here. Accepts a literal namespace name or a reference to a
KubernetesCertManager resource's output.

- references: KubernetesCertManager (`status.outputs.cluster_resource_namespace`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertManager, name: <that resource's name>, fieldPath: status.outputs.cluster_resource_namespace}} -- a bare string does not parse

### spec.config

`CertManagerIssuerConfig` · required

The signing backend and its configuration — ACME (public CAs like
Let's Encrypt), CA (sign with a Secret-held keypair), SelfSigned
(bootstrap/dev), or Vault (Vault/OpenBao PKI).

- rule: {"required":true}
- rule: An issuer backend must be selected — choose 'acme' for public TLS certificates (e.g. Let's Encrypt), 'ca' to sign with a CA keypair from a Kubernetes Secret, 'self_signed' to bootstrap a CA chain or for development, or 'vault' for a Vault/OpenBao PKI engine

### spec.config.acme

`CertManagerAcmeConfig`

ACME (RFC 8555): obtain certificates from a public CA such as
Let's Encrypt by proving control of the requested names.

### spec.config.acme.email

`string` · required

Email registered with the ACME account — the CA sends certificate expiry
notices and account notifications here.

- rule: {"required":true}

### spec.config.acme.server

`string` · optional (explicit presence)

ACME directory URL. Defaults to Let's Encrypt production. Use the
Let's Encrypt staging endpoint
(https://acme-staging-v02.api.letsencrypt.org/directory) while testing —
production enforces strict rate limits that a misconfigured setup can
exhaust for a whole domain.

- default: `https://acme-v02.api.letsencrypt.org/directory`

### spec.config.acme.profile

`string`

ACME certificate profile, for CAs that offer them (e.g. Let's Encrypt
"classic", "tlsserver", "shortlived"). Leave empty for the CA default.

### spec.config.acme.preferredChain

`string`

Preferred certificate chain to request from the CA, matched against the
Common Name of the chain's root certificate (e.g. "ISRG Root X1").
Relevant when a CA offers multiple chains; the CA falls back to its
default chain when no match exists.

### spec.config.acme.caBundle

`string`

PEM bundle used to validate the ACME server's TLS certificate — for
private ACME CAs (e.g. an internal Smallstep/step-ca) whose certificate
is not in the system trust store. Mutually exclusive with skip_tls_verify.

### spec.config.acme.skipTlsVerify

`bool`

Disable TLS verification when talking to the ACME server. INSECURE —
only for lab environments; prefer ca_bundle for private CAs.

### spec.config.acme.externalAccountBinding

`CertManagerAcmeExternalAccountBinding`

External Account Binding — required by CAs that tie ACME accounts to an
existing customer account (ZeroSSL, Google Trust Services, Sectigo).

### spec.config.acme.externalAccountBinding.keyId

`string` · required

Key identifier issued by the CA.

- rule: {"required":true}

### spec.config.acme.externalAccountBinding.hmacKey

`string` · required · sensitive

Symmetric MAC key issued by the CA (base64url-encoded, as handed out by
the CA's dashboard). Materialized as a Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.acme.disableAccountKeyGeneration

`bool`

When true, cert-manager will NOT generate a new ACME account key and
instead requires the account key Secret to already exist. Used when
migrating an existing ACME account between clusters.

### spec.config.acme.enableDurationFeature

`bool`

When true, request the certificate validity period from the CA via the
ACME "notAfter" field (only where Certificate resources set a duration;
support varies by CA).

### spec.config.acme.solvers

`[]CertManagerAcmeSolver` · required

Challenge solvers, tried in order of selector specificity. At least one
is required: an ACME issuer without solvers can never satisfy a
challenge, so every issuance would hang — cert-manager accepts such an
issuer but it is always a misconfiguration, rejected here instead.
A single catch-all solver (no selector) is the common case; add
selector-scoped solvers when different DNS zones need different
providers or when specific certificates must use HTTP-01.

- rule: {"repeated":{"minItems":"1"}}
- rule: Each solver must configure exactly one challenge type — 'http01' (proves control of a domain by serving a token over HTTP; requires public reachability on port 80) or 'dns01' (proves control by publishing a DNS TXT record; required for wildcard certificates)

### spec.config.acme.solvers[].selector

`CertManagerAcmeSolverSelector`

Restricts which certificates this solver applies to. Empty selector =
catch-all. When multiple solvers match, the most specific selector wins
(dns_names beats dns_zones beats match_labels beats catch-all).

### spec.config.acme.solvers[].selector.dnsZones

`[]string`

Match certificates whose names fall under any of these zones
(e.g. "example.com" matches "www.example.com").

### spec.config.acme.solvers[].selector.dnsNames

`[]string`

Match exact DNS names (e.g. "api.example.com").

### spec.config.acme.solvers[].selector.matchLabels

`map<string, string>`

Match Certificate resources carrying all of these labels.

### spec.config.acme.solvers[].http01

`CertManagerAcmeHttp01Solver`

HTTP-01: cert-manager provisions an ephemeral challenge endpoint that
the CA fetches over HTTP on port 80. Cannot issue wildcards.

- rule: http01 must configure exactly one exposure — 'ingress' (challenge routed through an Ingress controller) or 'gateway_http_route' (challenge routed through a Gateway API gateway)

### spec.config.acme.solvers[].http01.ingress

`CertManagerAcmeHttp01IngressSolver`

Route challenge traffic through an Ingress controller.

- rule: Set at most one of ingress_class_name (create ephemeral challenge Ingresses with that class) or name (modify an existing Ingress) — they select mutually exclusive strategies

### spec.config.acme.solvers[].http01.ingress.ingressClassName

`string`

IngressClass used for the ephemeral challenge Ingress (e.g. "nginx").
cert-manager creates new challenge Ingresses with this class. Set either
this or name — not both.

### spec.config.acme.solvers[].http01.ingress.name

`string`

Name of an EXISTING Ingress to temporarily modify with challenge routes
instead of creating new ones — for controllers (or DNS setups) that
misbehave with ephemeral Ingresses. Set either this or ingress_class_name.

### spec.config.acme.solvers[].http01.ingress.serviceType

`string` · optional (explicit presence)

Service type for the challenge Service. Defaults to NodePort upstream;
ClusterIP suffices when the Ingress controller routes in-cluster.

- rule: service_type must be NodePort or ClusterIP

### spec.config.acme.solvers[].http01.gatewayHttpRoute

`CertManagerAcmeHttp01GatewaySolver`

Route challenge traffic through a Gateway API HTTPRoute.

### spec.config.acme.solvers[].http01.gatewayHttpRoute.parentRefs

`[]CertManagerGatewayParentRef` · required

Gateways the challenge HTTPRoute attaches to. At least one required.

- rule: {"repeated":{"minItems":"1"}}

### spec.config.acme.solvers[].http01.gatewayHttpRoute.parentRefs[].name

`string` · required

Gateway name.

- rule: {"required":true}

### spec.config.acme.solvers[].http01.gatewayHttpRoute.parentRefs[].namespace

`string`

Gateway namespace. Empty = same namespace as the challenge route.

### spec.config.acme.solvers[].http01.gatewayHttpRoute.parentRefs[].sectionName

`string`

Specific listener section name on the Gateway (empty = all listeners).

### spec.config.acme.solvers[].http01.gatewayHttpRoute.labels

`map<string, string>`

Labels stamped on the ephemeral challenge HTTPRoute.

### spec.config.acme.solvers[].http01.gatewayHttpRoute.serviceType

`string` · optional (explicit presence)

Service type for the challenge Service (NodePort default upstream).

- rule: service_type must be NodePort or ClusterIP

### spec.config.acme.solvers[].dns01

`CertManagerAcmeDns01Solver`

DNS-01: cert-manager publishes a TXT record through the configured DNS
provider. The only challenge type that can issue wildcard certificates,
and the one that works for cluster-internal or non-public services.

- rule: dns01 must configure exactly one DNS provider (cloudflare, route53, azure_dns, gcp_cloud_dns, digitalocean, rfc2136, acme_dns, akamai, or webhook for third-party solvers)

### spec.config.acme.solvers[].dns01.cnameStrategy

`string` · optional (explicit presence)

How cert-manager follows CNAME records when locating the TXT record
zone: "none" (default — set the record where the name lives) or "follow"
(chase CNAMEs to the canonical zone; the pattern for delegated
_acme-challenge subdomains).

- rule: cname_strategy must be 'none' or 'follow'

### spec.config.acme.solvers[].dns01.cloudflare

`CertManagerDns01Cloudflare`

- rule: Cloudflare needs exactly one credential — api_token (recommended: scoped to the zone) or the legacy api_key + email pair

### spec.config.acme.solvers[].dns01.cloudflare.apiToken

`CertManagerCloudflareApiToken`

Scoped API token (recommended). Created in the Cloudflare dashboard
with Zone:Zone:Read and Zone:DNS:Edit on the zones this issuer manages.

### spec.config.acme.solvers[].dns01.cloudflare.apiToken.token

`string` · required · sensitive

The API token value. Materialized as a Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.cloudflare.apiKey

`CertManagerCloudflareApiKey`

Legacy global API key — account-wide power; use only where tokens are
unavailable. Requires the account email alongside.

### spec.config.acme.solvers[].dns01.cloudflare.apiKey.email

`string` · required

Cloudflare account email the key belongs to.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.cloudflare.apiKey.key

`string` · required · sensitive

The global API key value. Materialized as a Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.route53

`CertManagerDns01Route53`

- rule: Configure at most one Route53 authentication path — static_credentials or service_account; leave both empty for ambient (IRSA on the cert-manager controller) authentication

### spec.config.acme.solvers[].dns01.route53.region

`string` · required

AWS region hosting Route53 API calls (e.g. "us-east-1").

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.route53.hostedZoneId

`string`

Route53 hosted zone ID (e.g. "Z2E9THH2A..."). Optional: when empty,
cert-manager discovers the zone by name — set it to disambiguate when
multiple zones share a domain name (public + private split-horizon).

### spec.config.acme.solvers[].dns01.route53.assumeRoleArn

`string`

IAM role ARN to assume (via STS) before touching Route53 — the
cross-account pattern: the base identity (ambient or static) must be
allowed to sts:AssumeRole this role.

### spec.config.acme.solvers[].dns01.route53.staticCredentials

`CertManagerRoute53StaticCredentials`

Static AWS access keys. Leave unset for keyless (IRSA / ambient) auth.

### spec.config.acme.solvers[].dns01.route53.staticCredentials.accessKeyId

`string` · required

AWS access key ID.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.route53.staticCredentials.secretAccessKey

`string` · required · sensitive

AWS secret access key. Materialized as a Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.route53.serviceAccount

`CertManagerRoute53ServiceAccountAuth`

Dedicated ServiceAccount whose projected token authenticates to AWS —
per-issuer IRSA. The ServiceAccount must exist in the cert-manager
namespace and carry the eks.amazonaws.com/role-arn annotation
(compose it with a KubernetesServiceAccount resource).

### spec.config.acme.solvers[].dns01.route53.serviceAccount.serviceAccountName

`string | valueFrom` · required

Name of the ServiceAccount (in the cert-manager namespace) whose token
is exchanged for AWS credentials. Accepts a literal name or a reference
to a KubernetesServiceAccount resource.

- references: KubernetesServiceAccount (`status.outputs.service_account_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: status.outputs.service_account_name}} -- a bare string does not parse

### spec.config.acme.solvers[].dns01.route53.serviceAccount.audiences

`[]string`

TokenRequest audiences for the projected token. Empty = cert-manager's
default audience set.

### spec.config.acme.solvers[].dns01.azureDns

`CertManagerDns01AzureDns`

- rule: Service-principal authentication needs client_id, client_secret, and tenant_id together; leave all three empty for keyless (workload identity / managed identity) authentication
- rule: Configure either service-principal credentials or managed_identity, not both

### spec.config.acme.solvers[].dns01.azureDns.subscriptionId

`string` · required

Azure subscription ID containing the DNS zone.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.azureDns.resourceGroupName

`string` · required

Resource group containing the DNS zone.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.azureDns.hostedZoneName

`string`

DNS zone name (e.g. "example.com"). Optional: when empty, cert-manager
discovers the zone from the record name.

### spec.config.acme.solvers[].dns01.azureDns.zoneType

`string` · optional (explicit presence)

"public" (default) or "private" — which zone type to manage.

- rule: zone_type must be 'public' or 'private'

### spec.config.acme.solvers[].dns01.azureDns.environment

`string` · optional (explicit presence)

Azure cloud environment. Default: the public cloud.

- rule: environment must be one of AzurePublicCloud, AzureChinaCloud, AzureGermanCloud, AzureUSGovernmentCloud

### spec.config.acme.solvers[].dns01.azureDns.clientId

`string`

Service principal (static) auth: the app registration's client ID.
Requires client_secret and tenant_id.

### spec.config.acme.solvers[].dns01.azureDns.clientSecret

`string` · sensitive

Service principal client secret. Materialized as a Kubernetes Secret by
the module. Leave empty for keyless (workload identity) auth.

### spec.config.acme.solvers[].dns01.azureDns.tenantId

`string`

Entra tenant ID for service principal auth.

### spec.config.acme.solvers[].dns01.azureDns.managedIdentity

`CertManagerAzureManagedIdentity`

Per-issuer managed identity (keyless): use this user-assigned managed
identity instead of the controller's ambient identity. Set client_id OR
resource_id of the identity.

- rule: Select the managed identity by exactly one of client_id or resource_id

### spec.config.acme.solvers[].dns01.azureDns.managedIdentity.clientId

`string`

Client ID (GUID) of the user-assigned managed identity.

### spec.config.acme.solvers[].dns01.azureDns.managedIdentity.resourceId

`string`

Full Azure resource ID of the user-assigned managed identity.

### spec.config.acme.solvers[].dns01.gcpCloudDns

`CertManagerDns01GcpCloudDns`

### spec.config.acme.solvers[].dns01.gcpCloudDns.projectId

`string` · required

GCP project ID containing the Cloud DNS zone.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.gcpCloudDns.hostedZoneName

`string`

Cloud DNS zone name. Optional: when empty, cert-manager discovers the
zone from the record name.

### spec.config.acme.solvers[].dns01.gcpCloudDns.serviceAccountKeyJson

`string` · sensitive

GCP service account key JSON (static auth). Materialized as a Kubernetes
Secret by the module. Leave empty for keyless (Workload Identity) auth —
the strongly preferred path.

### spec.config.acme.solvers[].dns01.digitalocean

`CertManagerDns01DigitalOcean`

### spec.config.acme.solvers[].dns01.digitalocean.token

`string` · required · sensitive

DigitalOcean personal access token with write scope. Materialized as a
Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.rfc2136

`CertManagerDns01Rfc2136`

- rule: tsig_key_name and tsig_secret must be set together (both or neither) — a TSIG key name without its secret (or vice versa) cannot authenticate updates

### spec.config.acme.solvers[].dns01.rfc2136.nameserver

`string` · required

Authoritative DNS server address ("host" or "host:port"; IP required by
upstream for the host part).

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.rfc2136.tsigKeyName

`string`

TSIG key name, as configured on the DNS server. Required together with
tsig_secret for authenticated updates.

### spec.config.acme.solvers[].dns01.rfc2136.tsigAlgorithm

`string`

TSIG algorithm (e.g. "HMACSHA256", "HMACSHA512"). Defaults upstream to
HMACMD5; set explicitly to match the server key.

### spec.config.acme.solvers[].dns01.rfc2136.tsigSecret

`string` · sensitive

TSIG shared secret. Materialized as a Kubernetes Secret by the module.

### spec.config.acme.solvers[].dns01.acmeDns

`CertManagerDns01AcmeDns`

### spec.config.acme.solvers[].dns01.acmeDns.host

`string` · required

URL of the acme-dns server.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.acmeDns.accountJson

`string` · required · sensitive

acme-dns account registration JSON (the response from the /register
endpoint, keyed by domain). Materialized as a Kubernetes Secret by the
module.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.akamai

`CertManagerDns01Akamai`

### spec.config.acme.solvers[].dns01.akamai.serviceConsumerDomain

`string` · required

Akamai API service consumer domain (from the API client credentials).

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.akamai.clientToken

`string` · required · sensitive

Akamai API client token. Materialized as a Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.akamai.clientSecret

`string` · required · sensitive

Akamai API client secret. Materialized as a Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.akamai.accessToken

`string` · required · sensitive

Akamai API access token. Materialized as a Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.webhook

`CertManagerDns01Webhook`

### spec.config.acme.solvers[].dns01.webhook.groupName

`string` · required

API group of the webhook solver (e.g. "acme.hetzner.cloud"), as
documented by the webhook project.

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.webhook.solverName

`string` · required

Solver name within the group (e.g. "hetzner").

- rule: {"required":true}

### spec.config.acme.solvers[].dns01.webhook.configYaml

`string`

Webhook-specific configuration as a YAML document — the schema belongs
to the webhook project (typically credentials secret references and
zone settings). Parsed and passed through as the solver's config.

### spec.config.ca

`CertManagerCaConfig`

CA: sign certificates with a CA certificate + private key stored in a
Kubernetes Secret (typically created by a KubernetesCertificate with
is_ca=true and a self_signed issuer — the standard internal-PKI bootstrap).

### spec.config.ca.caSecretName

`string | valueFrom` · required

Name of the Secret holding the CA certificate and key. Typically the
output of a KubernetesCertificate with is_ca=true — reference its
status.outputs.secret_name to compose the standard CA-chain bootstrap.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.config.ca.crlDistributionPoints

`[]string`

CRL distribution point URLs stamped into issued certificates.

### spec.config.ca.ocspServers

`[]string`

OCSP responder URLs stamped into issued certificates.

### spec.config.ca.issuingCertificateUrls

`[]string`

CA-issuers URLs (Authority Information Access) stamped into issued
certificates, pointing at the CA certificate for chain building.

### spec.config.selfSigned

`CertManagerSelfSignedConfig`

SelfSigned: certificates are signed by their own private key. No
external authority involved — browsers will not trust these. Used to
create root CAs and for development.

### spec.config.selfSigned.crlDistributionPoints

`[]string`

CRL distribution point URLs stamped into issued certificates. Rarely
needed for self-signed use.

### spec.config.vault

`CertManagerVaultConfig`

Vault: certificates are signed by a HashiCorp Vault (or OpenBao) PKI
secrets engine.

- rule: Vault authentication is required — choose token (static Vault token), app_role (AppRole role_id + secret_id), or kubernetes (the cluster's ServiceAccount token via Vault's Kubernetes auth method, the keyless path)

### spec.config.vault.server

`string` · required

Vault server URL (e.g. "https://vault.example.com:8200").

- rule: {"required":true}

### spec.config.vault.path

`string` · required

PKI signing path (e.g. "pki_int/sign/example-dot-com").

- rule: {"required":true}

### spec.config.vault.vaultNamespace

`string`

Vault Enterprise namespace (empty for OSS Vault / OpenBao root).

### spec.config.vault.caBundle

`string`

PEM bundle to validate the Vault server's TLS certificate, for private
CAs. Empty = system trust store.

### spec.config.vault.serverName

`string`

SNI host name used when validating the Vault server certificate, when it
differs from the server URL host.

### spec.config.vault.tokenAuth

`CertManagerVaultTokenAuth`

Static Vault token auth.

### spec.config.vault.tokenAuth.token

`string` · required · sensitive

The Vault token. Materialized as a Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.vault.appRoleAuth

`CertManagerVaultAppRoleAuth`

AppRole auth.

### spec.config.vault.appRoleAuth.path

`string` · required

Mount path of the AppRole auth method (e.g. "approle").

- rule: {"required":true}

### spec.config.vault.appRoleAuth.roleId

`string` · required

AppRole role ID.

- rule: {"required":true}

### spec.config.vault.appRoleAuth.secretId

`string` · required · sensitive

AppRole secret ID. Materialized as a Kubernetes Secret by the module.

- rule: {"required":true}

### spec.config.vault.kubernetesAuth

`CertManagerVaultKubernetesAuth`

Kubernetes auth method — Vault validates the cluster ServiceAccount
token; no static secret involved (the keyless path).

### spec.config.vault.kubernetesAuth.role

`string` · required

Vault role bound to the ServiceAccount (configured on the Vault side).

- rule: {"required":true}

### spec.config.vault.kubernetesAuth.mountPath

`string` · optional (explicit presence)

Mount path of the Kubernetes auth method. Defaults to "kubernetes".

- default: `kubernetes`

### spec.config.vault.kubernetesAuth.serviceAccountName

`string | valueFrom` · required

ServiceAccount (in the cert-manager namespace for a ClusterIssuer; the
Issuer's namespace for an Issuer) whose token authenticates to Vault.
Accepts a literal name or a reference to a KubernetesServiceAccount.

- references: KubernetesServiceAccount (`status.outputs.service_account_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesServiceAccount, name: <that resource's name>, fieldPath: status.outputs.service_account_name}} -- a bare string does not parse

### spec.config.vault.kubernetesAuth.audiences

`[]string`

TokenRequest audiences for the ServiceAccount token. Empty = Vault default.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesClusterIssuer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_issuer_name` | `string` | Name of the created ClusterIssuer (equals metadata.name). Use it in Certificate issuerRef.name (kind ClusterIssuer) and in the cert-manager.io/cluster-issuer ingress-shim annotation. |
| `status.outputs.secrets_namespace` | `string` | Namespace where credential Secrets for this issuer were materialized (cert-manager's cluster-resource namespace). |
| `status.outputs.acme_account_key_secret_name` | `string` | Name of the ACME account private key Secret cert-manager creates in the cluster-resource namespace. Empty for non-ACME backends. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.certManagerNamespace` | KubernetesCertManager | `status.outputs.cluster_resource_namespace` |
| `spec.config.acme.solvers[].dns01.route53.serviceAccount.serviceAccountName` | KubernetesServiceAccount | `status.outputs.service_account_name` |
| `spec.config.ca.caSecretName` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.config.vault.kubernetesAuth.serviceAccountName` | KubernetesServiceAccount | `status.outputs.service_account_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesCertificate | `spec.issuerRef.clusterIssuer.name` | `status.outputs.cluster_issuer_name` |
| KubernetesKeda | `spec.certificates.certManagerIssuer.clusterIssuer.name` | `status.outputs.cluster_issuer_name` |
| KubernetesKyverno | `spec.certificates.certManager.issuerName` | `metadata.name` |
| KubernetesMetricsServer | `spec.tls.certManagerIssuer.clusterIssuer.name` | `status.outputs.cluster_issuer_name` |
| KubernetesMongodb | `spec.tls.issuer` | `metadata.name` |
| KubernetesMysql | `spec.tls.issuer` | `metadata.name` |

## See Also

- [Overview](../README.md)
