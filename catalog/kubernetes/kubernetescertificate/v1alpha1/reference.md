# KubernetesCertificate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesCertificateSpec** requests one signed X.509 certificate from a
cert-manager issuer and keeps it renewed for as long as the resource
exists. The signed certificate, its private key, and the CA chain land in
a Kubernetes TLS Secret (`secret_name`) that consumers — Ingress TLS
blocks, Gateway listeners, workload volume mounts, CA issuers — reference
by name.

The issuer decides WHO signs (Let's Encrypt via ACME, an internal CA, a
self-signed root, Vault); this resource decides WHAT is requested: the
names, lifetime, key parameters, usages, and output formats. Covers the
complete cert-manager.io/v1 Certificate surface.

Requires cert-manager (KubernetesCertManager) on the cluster.

## Example

```yaml
# Full-surface manifest for offline module proofs (tofu validate/plan and
# pulumi preview). Exercises names, subject, key parameters, keystores,
# output formats, and the secret template; both engines must render an
# identical CR from it. (The is_ca + name_constraints arm is exercised by
# the CA-chain E2E scenario.)
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCertificate
metadata:
  name: test-certificate
spec:
  namespace:
    value: team-a
  secretName: test-certificate-tls
  issuerRef:
    clusterIssuer:
      name:
        value: letsencrypt-staging
  dnsNames:
    - api.example.com
    - "*.apps.example.com"
  ipAddresses:
    - 10.0.0.10
  uris:
    - spiffe://cluster.local/ns/team-a/sa/api
  emailAddresses:
    - platform@example.com
  commonName: api.example.com
  subject:
    organizations:
      - Example Corp
    organizationalUnits:
      - Platform
    countries:
      - US
    provinces:
      - CA
    localities:
      - San Francisco
    streetAddresses:
      - 1 Example Way
    postalCodes:
      - "94105"
    serialNumber: "0042"
  otherNames:
    - oid: 1.3.6.1.4.1.311.20.2.3
      utf8Value: upn@example.com
  duration: 2160h
  renewBefore: 360h
  privateKey:
    algorithm: rsa
    size: 4096
    encoding: pkcs8
    rotationPolicy: always
  usages:
    - digital signature
    - key encipherment
    - server auth
  encodeUsagesInRequest: true
  isCa: false
  signatureAlgorithm: SHA384WithRSA
  keystores:
    jks:
      create: true
      alias: certificate
      password: test-jks-password
    pkcs12:
      create: true
      password: test-p12-password
      profile: modern2023
  additionalOutputFormats:
    - type: der
    - type: combined_pem
  secretTemplate:
    labels:
      team: team-a
    annotations:
      example.com/rotation-owner: platform
  revisionHistoryLimit: 3
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.secretName` | `string` | yes |  |  |
| `spec.issuerRef` | `KubernetesCertificateIssuerRef` | yes |  |  |
| `spec.issuerRef.clusterIssuer` | `KubernetesCertificateClusterIssuerRef` |  |  |  |
| `spec.issuerRef.clusterIssuer.name` | `string \| valueFrom` | yes |  | KubernetesClusterIssuer (`status.outputs.cluster_issuer_name`) |
| `spec.issuerRef.issuer` | `KubernetesCertificateNamespacedIssuerRef` |  |  |  |
| `spec.issuerRef.issuer.name` | `string \| valueFrom` | yes |  | KubernetesIssuer (`status.outputs.issuer_name`) |
| `spec.issuerRef.external` | `KubernetesCertificateExternalIssuerRef` |  |  |  |
| `spec.issuerRef.external.group` | `string` | yes |  |  |
| `spec.issuerRef.external.kind` | `string` | yes |  |  |
| `spec.issuerRef.external.name` | `string` | yes |  |  |
| `spec.dnsNames` | `[]string` |  |  |  |
| `spec.ipAddresses` | `[]string` |  |  |  |
| `spec.uris` | `[]string` |  |  |  |
| `spec.emailAddresses` | `[]string` |  |  |  |
| `spec.commonName` | `string` |  |  |  |
| `spec.subject` | `KubernetesCertificateSubject` |  |  |  |
| `spec.subject.organizations` | `[]string` |  |  |  |
| `spec.subject.organizationalUnits` | `[]string` |  |  |  |
| `spec.subject.countries` | `[]string` |  |  |  |
| `spec.subject.provinces` | `[]string` |  |  |  |
| `spec.subject.localities` | `[]string` |  |  |  |
| `spec.subject.streetAddresses` | `[]string` |  |  |  |
| `spec.subject.postalCodes` | `[]string` |  |  |  |
| `spec.subject.serialNumber` | `string` |  |  |  |
| `spec.literalSubject` | `string` |  |  |  |
| `spec.otherNames` | `[]KubernetesCertificateOtherName` |  |  |  |
| `spec.otherNames[].oid` | `string` | yes |  |  |
| `spec.otherNames[].utf8Value` | `string` | yes |  |  |
| `spec.duration` | `string` |  | `2160h` |  |
| `spec.renewBefore` | `string` |  |  |  |
| `spec.renewBeforePercentage` | `int32` |  |  |  |
| `spec.privateKey` | `KubernetesCertificatePrivateKey` |  |  |  |
| `spec.privateKey.algorithm` | `enum` |  | `rsa` |  |
| `spec.privateKey.size` | `int32` |  |  |  |
| `spec.privateKey.encoding` | `enum` |  | `pkcs1` |  |
| `spec.privateKey.rotationPolicy` | `enum` |  | `always` |  |
| `spec.usages` | `[]string` |  |  |  |
| `spec.encodeUsagesInRequest` | `bool` |  |  |  |
| `spec.isCa` | `bool` |  |  |  |
| `spec.signatureAlgorithm` | `string` |  |  |  |
| `spec.keystores` | `KubernetesCertificateKeystores` |  |  |  |
| `spec.keystores.jks` | `KubernetesCertificateJksKeystore` |  |  |  |
| `spec.keystores.jks.create` | `bool` |  |  |  |
| `spec.keystores.jks.alias` | `string` |  | `certificate` |  |
| `spec.keystores.jks.password` | `string` (sensitive) | yes |  |  |
| `spec.keystores.pkcs12` | `KubernetesCertificatePkcs12Keystore` |  |  |  |
| `spec.keystores.pkcs12.create` | `bool` |  |  |  |
| `spec.keystores.pkcs12.password` | `string` (sensitive) | yes |  |  |
| `spec.keystores.pkcs12.profile` | `string` |  |  |  |
| `spec.additionalOutputFormats` | `[]KubernetesCertificateAdditionalOutputFormat` |  |  |  |
| `spec.additionalOutputFormats[].type` | `string` | yes |  |  |
| `spec.nameConstraints` | `KubernetesCertificateNameConstraints` |  |  |  |
| `spec.nameConstraints.critical` | `bool` |  |  |  |
| `spec.nameConstraints.permitted` | `KubernetesCertificateNameConstraintSet` |  |  |  |
| `spec.nameConstraints.permitted.dnsDomains` | `[]string` |  |  |  |
| `spec.nameConstraints.permitted.ipRanges` | `[]string` |  |  |  |
| `spec.nameConstraints.permitted.emailAddresses` | `[]string` |  |  |  |
| `spec.nameConstraints.permitted.uriDomains` | `[]string` |  |  |  |
| `spec.nameConstraints.excluded` | `KubernetesCertificateNameConstraintSet` |  |  |  |
| `spec.nameConstraints.excluded.dnsDomains` | `[]string` |  |  |  |
| `spec.nameConstraints.excluded.ipRanges` | `[]string` |  |  |  |
| `spec.nameConstraints.excluded.emailAddresses` | `[]string` |  |  |  |
| `spec.nameConstraints.excluded.uriDomains` | `[]string` |  |  |  |
| `spec.secretTemplate` | `KubernetesCertificateSecretTemplate` |  |  |  |
| `spec.secretTemplate.labels` | `map<string, string>` |  |  |  |
| `spec.secretTemplate.annotations` | `map<string, string>` |  |  |  |
| `spec.revisionHistoryLimit` | `int32` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace the Certificate (and its output Secret) is created in.
A namespace-scoped Issuer must live in this same namespace; a
ClusterIssuer serves any namespace. Accepts a literal namespace name or
a reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.secretName

`string` · required

Name of the Kubernetes Secret the signed certificate is written to
(keys: tls.crt, tls.key, ca.crt). This is the handle every consumer
references — exported as status.outputs.secret_name.

- rule: {"required":true}

### spec.issuerRef

`KubernetesCertificateIssuerRef` · required

The issuer that signs this certificate.

- rule: {"required":true}
- rule: An issuer reference is required — choose 'cluster_issuer' (cluster-scoped, serves any namespace), 'issuer' (namespace-scoped, same namespace as the certificate), or 'external' for a third-party issuer kind (e.g. AWS Private CA)

### spec.issuerRef.clusterIssuer

`KubernetesCertificateClusterIssuerRef`

Cluster-scoped ClusterIssuer, by name.

### spec.issuerRef.clusterIssuer.name

`string | valueFrom` · required

ClusterIssuer name. Accepts a literal name or a reference to a
KubernetesClusterIssuer resource.

- references: KubernetesClusterIssuer (`status.outputs.cluster_issuer_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClusterIssuer, name: <that resource's name>, fieldPath: status.outputs.cluster_issuer_name}} -- a bare string does not parse

### spec.issuerRef.issuer

`KubernetesCertificateNamespacedIssuerRef`

Namespace-scoped Issuer (must live in the certificate's namespace), by name.

### spec.issuerRef.issuer.name

`string | valueFrom` · required

Issuer name. Accepts a literal name or a reference to a
KubernetesIssuer resource.

- references: KubernetesIssuer (`status.outputs.issuer_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesIssuer, name: <that resource's name>, fieldPath: status.outputs.issuer_name}} -- a bare string does not parse

### spec.issuerRef.external

`KubernetesCertificateExternalIssuerRef`

Third-party issuer kind (external issuer controller).

### spec.issuerRef.external.group

`string` · required

API group of the external issuer kind (e.g. "awspca.cert-manager.io").

- rule: {"required":true}

### spec.issuerRef.external.kind

`string` · required

Kind of the external issuer (e.g. "AWSPCAClusterIssuer").

- rule: {"required":true}

### spec.issuerRef.external.name

`string` · required

Name of the external issuer resource.

- rule: {"required":true}

### spec.dnsNames

`[]string`

Requested DNS Subject Alternative Names (e.g. "api.example.com",
"*.apps.example.com"). Wildcards require a DNS-01-capable ACME issuer or
a private CA. At least one name field (dns_names, ip_addresses, uris,
email_addresses, common_name, or literal_subject) must be set.

### spec.ipAddresses

`[]string`

Requested IP address SANs — for certificates presented on bare IPs
(internal load balancers, node endpoints). Public ACME CAs generally
refuse IP SANs; use with internal issuers.

### spec.uris

`[]string`

Requested URI SANs (e.g. SPIFFE IDs like
"spiffe://cluster.local/ns/team-a/sa/api"). The mTLS workload-identity
pattern with internal issuers.

### spec.emailAddresses

`[]string`

Requested email address SANs — for S/MIME client certificates.

### spec.commonName

`string`

Requested X.509 common name (CN). Modern TLS validation ignores CN in
favor of SANs, and CNs are limited to 64 characters — set it only when a
consumer requires it (legacy clients, LDAP, some private CAs).

- rule: {"string":{"maxLen":"64"}}

### spec.subject

`KubernetesCertificateSubject`

Requested subject attributes (organization, country, ...). Most public
CAs (Let's Encrypt) strip these; they matter for internal PKI. Mutually
exclusive with literal_subject.

### spec.subject.organizations

`[]string`

Organizations (O).

### spec.subject.organizationalUnits

`[]string`

Organizational units (OU).

### spec.subject.countries

`[]string`

Countries (C), two-letter codes.

### spec.subject.provinces

`[]string`

State/province names (ST).

### spec.subject.localities

`[]string`

Locality/city names (L).

### spec.subject.streetAddresses

`[]string`

Street addresses.

### spec.subject.postalCodes

`[]string`

Postal codes.

### spec.subject.serialNumber

`string`

Serial number RDN (NOT the certificate serial — the subject attribute).

### spec.literalSubject

`string`

The FULL subject as an LDAP RFC 4514 distinguished-name string (e.g.
"CN=api,OU=platform,O=acme,C=US"). Preserves attribute ORDER — required
for LDAP-authentication certificates where order is semantic. Mutually
exclusive with subject and common_name (it embeds both).

### spec.otherNames

`[]KubernetesCertificateOtherName`

OtherName SANs (e.g. Microsoft User Principal Names for smart-card /
AD-joined client certificates).

### spec.otherNames[].oid

`string` · required

Object identifier in dotted form (e.g. "1.3.6.1.4.1.311.20.2.3").

- rule: {"required":true}

### spec.otherNames[].utf8Value

`string` · required

UTF-8 value for the name (e.g. "upn@domain.local").

- rule: {"required":true}

### spec.duration

`string` · optional (explicit presence)

Requested certificate lifetime as a Go duration (e.g. "2160h" = 90
days). The issuer may override (ACME CAs set their own policy; Let's
Encrypt always issues 90 days). Default upstream: 90 days.

- default: `2160h`
- rule: duration must be a Go duration string (e.g. '2160h', '720h30m')

### spec.renewBefore

`string`

Renew this long before expiry, as a Go duration (e.g. "360h" = 15 days).
Mutually exclusive with renew_before_percentage. Default upstream:
a third of the certificate lifetime.

- rule: renew_before must be a Go duration string (e.g. '360h')

### spec.renewBeforePercentage

`int32` · optional (explicit presence)

Renew when this percentage of the certificate's lifetime REMAINS (e.g.
25 renews a 60-minute certificate 45 minutes after issuance). Scales
with issuer-decided lifetimes, unlike the absolute renew_before.
Mutually exclusive with renew_before.

- rule: {"int32":{"lte":99,"gte":1}}

### spec.privateKey

`KubernetesCertificatePrivateKey`

Private key algorithm, size, encoding, and rotation policy.

- rule: Key size must match the algorithm family — RSA: 2048/3072/4096/8192, ECDSA: 256/384/521 (Ed25519 takes no size)

### spec.privateKey.algorithm

`enum` · optional (explicit presence)

Key algorithm. Default upstream: RSA.

- default: `rsa`

Allowed values (use exactly as shown):

- `kubernetes_certificate_private_key_algorithm_unspecified`
- `rsa`
- `ecdsa`
- `ed25519`

### spec.privateKey.size

`int32` · optional (explicit presence)

Key size: RSA 2048/3072/4096/8192 (default 2048); ECDSA 256/384/521
(default 256). Ignored for Ed25519.

- rule: size must be 2048, 3072, 4096, or 8192 for RSA; 256, 384, or 521 for ECDSA

### spec.privateKey.encoding

`enum` · optional (explicit presence)

Serialization format in the Secret. Default upstream: PKCS#1.

- default: `pkcs1`

Allowed values (use exactly as shown):

- `kubernetes_certificate_private_key_encoding_unspecified`
- `pkcs1`
- `pkcs8`

### spec.privateKey.rotationPolicy

`enum` · optional (explicit presence)

Key rotation at renewal. Default upstream (v1.18+): always.

- default: `always`

Allowed values (use exactly as shown):

- `kubernetes_certificate_private_key_rotation_policy_unspecified`
- `always`
- `never`

### spec.usages

`[]string`

Requested key usages and extended key usages. Empty = upstream default
("digital signature" + "key encipherment"). Add "server auth" / "client
auth" explicitly when an issuer requires exact usages (Vault, some
private CAs); public ACME CAs typically override.

- rule: {"repeated":{"items":{"cel":[{"id":"usages.vocabulary","message":"Each usage must be one of cert-manager's x509 usage names, e.g. 'digital signature', 'key encipherment', 'server auth', 'client auth', 'cert sign', 'any' (full list: signing, digital signature, content commitment, key encipherment, key agreement, data encipherment, cert sign, crl sign, encipher only, decipher only, any, server auth, client auth, code signing, email protection, s/mime, ipsec end system, ipsec tunnel, ipsec user, timestamping, ocsp signing, microsoft sgc, netscape sgc)","expression":"this in ['signing', 'digital signature', 'content commitment', 'key encipherment', 'key agreement', 'data encipherment', 'cert sign', 'crl sign', 'encipher only', 'decipher only', 'any', 'server auth', 'client auth', 'code signing', 'email protection', 's/mime', 'ipsec end system', 'ipsec tunnel', 'ipsec user', 'timestamping', 'ocsp signing', 'microsoft sgc', 'netscape sgc']"}]}}}

### spec.encodeUsagesInRequest

`bool`

When true, the encoded usages are included in the CertificateRequest
sent to the issuer (some issuers require it, most ignore it).

### spec.isCa

`bool`

When true, request a CA certificate (adds the cert-sign usage and the CA
basic constraint). THE internal-PKI bootstrap: a self_signed issuer +
is_ca=true produces a root CA Secret, which a CA-backed issuer then
signs leaf certificates with.

### spec.signatureAlgorithm

`string` · optional (explicit presence)

Requested signature algorithm (e.g. "SHA256WithRSA", "ECDSAWithSHA384",
"PureEd25519"). Empty = issuer default. Must match the private key
algorithm family.

- rule: signature_algorithm must be one of SHA256WithRSA, SHA384WithRSA, SHA512WithRSA, ECDSAWithSHA256, ECDSAWithSHA384, ECDSAWithSHA512, PureEd25519

### spec.keystores

`KubernetesCertificateKeystores`

Java/PKCS#12 keystore outputs added to the Secret alongside the PEM
data — for JVM and Windows consumers that cannot read PEM.

### spec.keystores.jks

`KubernetesCertificateJksKeystore`

Java keystore (keystore.jks + truststore.jks entries in the Secret).

### spec.keystores.jks.create

`bool`

Create the JKS entries.

### spec.keystores.jks.alias

`string` · optional (explicit presence)

Certificate alias inside the keystore. Upstream default: "certificate".

- default: `certificate`

### spec.keystores.jks.password

`string` · required · sensitive

Password encrypting the keystore.

- rule: {"required":true}

### spec.keystores.pkcs12

`KubernetesCertificatePkcs12Keystore`

PKCS#12 keystore (keystore.p12 + truststore.p12 entries in the Secret).

### spec.keystores.pkcs12.create

`bool`

Create the PKCS#12 entries.

### spec.keystores.pkcs12.password

`string` · required · sensitive

Password encrypting the keystore.

- rule: {"required":true}

### spec.keystores.pkcs12.profile

`string` · optional (explicit presence)

Encryption/MAC profile: "legacy_rc2" (upstream default — maximum
compatibility, weak crypto), "legacy_des", or "modern2023" (AES/SHA256;
requires consumers newer than ~2013).

- rule: profile must be one of legacy_rc2, legacy_des, modern2023

### spec.additionalOutputFormats

`[]KubernetesCertificateAdditionalOutputFormat`

Additional encodings of the certificate added to the Secret:
"der" (tls-combined.der... binary DER of the key) and "combined_pem"
(tls-combined.pem: key + certificate concatenated — HAProxy-style
consumers).

### spec.additionalOutputFormats[].type

`string` · required

"der" (adds key.der — binary DER private key) or "combined_pem" (adds
tls-combined.pem — private key and certificate concatenated).

- rule: type must be 'der' or 'combined_pem'
- rule: {"required":true}

### spec.nameConstraints

`KubernetesCertificateNameConstraints`

X.509 name constraints stamped into CA certificates (is_ca=true) —
restricts what names the CA is allowed to sign. Defense-in-depth for
delegated internal CAs.

### spec.nameConstraints.critical

`bool`

Mark the extension critical: validators that do not understand name
constraints must reject the chain (the strict, recommended posture).

### spec.nameConstraints.permitted

`KubernetesCertificateNameConstraintSet`

Names the CA is permitted to sign.

### spec.nameConstraints.permitted.dnsDomains

`[]string`

DNS domains (e.g. "internal.example.com" constrains all subdomains).

### spec.nameConstraints.permitted.ipRanges

`[]string`

IP ranges in CIDR form.

### spec.nameConstraints.permitted.emailAddresses

`[]string`

Email addresses or domains.

### spec.nameConstraints.permitted.uriDomains

`[]string`

URI domains.

### spec.nameConstraints.excluded

`KubernetesCertificateNameConstraintSet`

Names the CA is forbidden to sign.

### spec.nameConstraints.excluded.dnsDomains

`[]string`

DNS domains (e.g. "internal.example.com" constrains all subdomains).

### spec.nameConstraints.excluded.ipRanges

`[]string`

IP ranges in CIDR form.

### spec.nameConstraints.excluded.emailAddresses

`[]string`

Email addresses or domains.

### spec.nameConstraints.excluded.uriDomains

`[]string`

URI domains.

### spec.secretTemplate

`KubernetesCertificateSecretTemplate`

Labels and annotations copied onto the output Secret — for consumers
that discover certificates by label (e.g. reflector/replicator tooling,
Gateway API implementations with Secret label selectors).

### spec.secretTemplate.labels

`map<string, string>`

Labels for the Secret.

### spec.secretTemplate.annotations

`map<string, string>`

Annotations for the Secret.

### spec.revisionHistoryLimit

`int32` · optional (explicit presence)

How many issued-certificate revisions (CertificateRequest resources)
cert-manager retains. Upstream default: 1. Raise it to keep an audit
trail of recent issuances.

- rule: {"int32":{"gte":1}}

## Validation Rules

- `spec.names_required`: At least one name must be requested — set dns_names, ip_addresses, uris, email_addresses, common_name, literal_subject, or other_names
- `spec.literal_subject_exclusive`: literal_subject embeds the full distinguished name (including CN and subject attributes) — it cannot be combined with subject or common_name
- `spec.renew_before_exclusive`: renew_before (absolute duration) and renew_before_percentage (relative) are two ways to express the same renewal window — set at most one

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesCertificate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace where the Certificate resource was created. |
| `status.outputs.certificate_name` | `string` | Name of the created Certificate resource. |
| `status.outputs.secret_name` | `string` | TLS Secret name. Consumers use this for Gateway certificateRefs, Ingress tls.secretName, or as ca_secret_name for a CA Issuer. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.issuerRef.clusterIssuer.name` | KubernetesClusterIssuer | `status.outputs.cluster_issuer_name` |
| `spec.issuerRef.issuer.name` | KubernetesIssuer | `status.outputs.issuer_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesClusterIssuer | `spec.config.ca.caSecretName` | `status.outputs.secret_name` |
| KubernetesGatekeeper | `spec.externalCert.secretName` | `spec.secret_name` |
| KubernetesHarbor | `spec.expose.tls.certSecretName` | `status.outputs.secret_name` |
| KubernetesIngressNginx | `spec.defaultTlsCertificate.secretName` | `status.outputs.secret_name` |
| KubernetesIssuer | `spec.config.ca.caSecretName` | `status.outputs.secret_name` |
| KubernetesKafka | `spec.listeners[].configuration.brokerCertChainAndKey.secretName` | `status.outputs.secret_name` |
| KubernetesKarapace | `spec.serverTls.secretName` | `status.outputs.secret_name` |
| KubernetesKeycloak | `spec.http.tlsSecretName` | `status.outputs.secret_name` |
| KubernetesNats | `spec.tls.secretName` | `status.outputs.secret_name` |
| KubernetesNeo4j | `spec.ssl.bolt.secret` | `status.outputs.secret_name` |
| KubernetesNeo4j | `spec.ssl.https.secret` | `status.outputs.secret_name` |
| KubernetesOpenBao | `spec.tls.certSecretName` | `status.outputs.secret_name` |
| KubernetesOpenSearch | `spec.security.transportTls.secret` | `status.outputs.secret_name` |
| KubernetesOpenSearch | `spec.security.httpTls.secret` | `status.outputs.secret_name` |
| KubernetesOpenSearch | `spec.dashboards.tls.secret` | `status.outputs.secret_name` |
| KubernetesPostgres | `spec.certificates.serverTlsSecret` | `status.outputs.secret_name` |
| KubernetesQdrant | `spec.tls.secret` | `status.outputs.secret_name` |
| KubernetesRabbitMq | `spec.tls.secretName` | `status.outputs.secret_name` |
| KubernetesValkey | `spec.tls.certificateSecret` | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
