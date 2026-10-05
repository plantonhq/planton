# AWS Private Certificate Authority

Deploys an AWS Private CA — the managed certificate authority behind internal TLS: mTLS services, MSK and EKS client auth, service meshes, and device fleets — with activation, issued certificates, the ACM renewal permission, and the cross-account resource policy managed in-line. Activation is composed, never user-choreographed: a root CA self-signs at apply and comes up ACTIVE and issuing in one deploy, and a subordinate activates from a parent AwsPrivateCa by reference. The CA bills a flat monthly fee from creation until deletion, prorated hourly, so an idle CA is a real cost decision — and the mode, algorithms, subject, and type are all fixed for life.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Certificate Authority** — root or subordinate, with its key and signing algorithms (post-quantum ML-DSA included), X.500 subject, usage mode, HSM standard, and CRL/OCSP revocation configuration
- **Activation Certificate** — for a root, the self-signed CA certificate issued and installed at apply (the CSR-issue-install dance the raw provider makes you wire); for a subordinate with `subordinateActivation`, a parent-signed CA certificate with the chosen path length. A subordinate without activation sits in PENDING_CERTIFICATE — created and billed, but unable to issue
- **Issued Certificates** — one per `issuedCertificates` entry; end-entity certificates signed from your own CSRs (the private key never touches AWS or the manifest)
- **ACM Renewal Permission** — created only when `acmRenewalPermission` is true; the grant that lets ACM auto-renew certificates it requested from this CA
- **Resource Policy** — created only when `policy` is set; grants other accounts or an organization the right to issue from this CA via RAM sharing

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with ACM PCA permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Nothing for a plain root CA** — it self-activates at apply.
- **A CRL bucket with the service policy in place** (only for CRL publishing) — AWS validates at CA create that the bucket grants ACM PCA `getBucketAcl` and `putObject`; a missing policy fails the deploy, not the first CRL.
- **An ACTIVE parent CA in this account** (only for composed subordinate activation) — referenced via `subordinateActivation.parentCaArn`. External or offline parents sign the `ca_csr` output out of band instead.

## Deploy

### Console

Open the deployment store, find **AWS Private Certificate Authority**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: hierarchy type and algorithms, the X.500 subject, usage mode, then the optional revocation, activation, and issuance sections. Start from the **Internal Root CA** preset in the [Presets](#presets) tab for the one-deploy trust anchor.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsPrivateCa
metadata:
  name: corp-root-ca
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  type: ROOT
  keyAlgorithm: RSA_2048
  signingAlgorithm: SHA256WITHRSA
  subject:
    commonName: Acme Internal Root CA G1
    organization: Acme Corp
    country: US
  rootCaValidity:
    type: YEARS
    value: "10"
  acmRenewalPermission: true
```

```shell
planton apply -f aws-private-ca.yaml
```

This creates a self-activated 10-year root CA, ACTIVE and issuing, with ACM pre-authorized to auto-renew the certificates it requests. An Infra Job tracks the provisioning in real time.

### InfraChart

When a subordinate deploys alongside its root in one chart, wire the parent via ValueFromRef:

```yaml
spec:
  region: us-east-1
  type: SUBORDINATE
  keyAlgorithm: EC_prime256v1
  signingAlgorithm: SHA256WITHECDSA
  usageMode: SHORT_LIVED_CERTIFICATE
  subject:
    commonName: Acme Workload CA G1
  subordinateActivation:
    parentCaArn:
      valueFrom:
        kind: AwsPrivateCa
        name: corp-root-ca
        fieldPath: status.outputs.certificate_authority_arn
    pathLength: 0
    validity:
      type: YEARS
      value: "5"
```

The InfraPipeline resolves the dependency graph, activates the root first, then signs and installs the subordinate's certificate from it.

## Key Configuration

These are the most important decisions when configuring a private CA. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The mode is the dominant cost decision, and it is permanent** — `usageMode: GENERAL_PURPOSE` exists for certificates that live long enough to need revocation; `SHORT_LIVED_CERTIFICATE` caps end-entity validity at 7 days, forbids CRL/OCSP, and bills a fraction of general-purpose — revocation by expiry. High-churn workload identity (mesh mTLS, batch jobs) belongs in short-lived mode; anything a human installs belongs in general-purpose. The mode is fixed for life, and the provider silently ignores an update attempt — get it right at create.

**A CA bills from create to delete, issuing or not** — a subordinate left in PENDING_CERTIFICATE costs the same as an ACTIVE one, and `enabled: false` pauses issuance without pausing billing. Per-certificate issuance fees add on top. Delete CAs you are not using; billing stops the moment the delete lands.

**Almost everything is fixed for life** — type, key and signing algorithms, subject, usage mode, and HSM standard all replace-or-ignore on change. The one deliberately flexible seam is revocation: CRL and OCSP arms toggle in place, as do issued certificates, the ACM permission, and the resource policy.

**Certificates are born, not edited** — every issue request's parameters (CSR, signing algorithm, validity, template, API passthrough) are the certificate's birth certificate: AWS reads them once and never reports them back, so both engines deliberately ignore later edits. The same applies to `rootCaValidity` and `subordinateActivation.validity`. To reissue with different parameters, add a new `issuedCertificates` entry (the name keys the certificate) and retire the old one when consumers have moved.

**ACM renewal needs the permission, and failures are silent** — AwsCertManagerCert certificates issued from this CA auto-renew only while acm.amazonaws.com holds the full three-action grant. Without `acmRenewalPermission` there is no deploy-time symptom — the certificate just expires months later. Set it on any CA that issues ACM certificates, always.

**Templates decide what a certificate can do** — `templateArn` picks the X.509 profile: plain TLS (the default EndEntityCertificate/V1), mTLS client auth, code signing, or the CA-path-length templates that mint sub-CAs. The CSR asks; the template decides — a CSR with CA:TRUE extensions issued under an end-entity template comes out a plain leaf.

**Same-family hierarchies, by design** — composed subordinate activation signs with this spec's `signingAlgorithm` at the parent, which is correct when parent and child share a key family (the norm). A cross-family hierarchy (an RSA root signing EC subordinates) needs out-of-band activation via the `ca_csr` output.

**Decide the restore window at create** — a deleted CA parks restorable for `permanentDeletionTimeInDays` (7-30, default 30), and the value is read at create only: AWS stores nothing, so later edits are ignored and the delete uses the create-time window. Ephemeral environments should set 7 — the shorter window frees the name-and-subject slot sooner.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsPrivateCa** | `subordinateActivation.parentCaArn` | `status.outputs.certificate_authority_arn` |
| **AwsS3Bucket** | `revocation.crl.s3BucketName` | `status.outputs.bucket_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate_authority_arn` | The CA's ARN — the join key every consumer uses | AwsCertManagerCert's issuing CA, MSK's TLS client-auth CA list, a subordinate's `parentCaArn` |
| `ca_certificate` | The CA's certificate, PEM | Relying parties' trust stores |
| `ca_certificate_chain` | The chain up to the root, PEM (empty for a root — it IS the anchor) | Trust-store distribution for subordinate-issued certificates |
| `ca_csr` | The CA's own CSR, PEM | What an external parent signs when activating a subordinate out of band |

`certificate_authority_id`, `issued_certificate_arns` (a name-keyed map), and `activation_certificate_arn` are also present — import addresses and audit echoes rather than composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One-deploy trust anchor** — a self-activating 10-year root with the ACM renewal permission granted, issuing internal TLS through AwsCertManagerCert against `certificate_authority_arn` while trust stores carry `ca_certificate`. One root serves the whole organization. Start from the **Internal Root CA** preset.

**Two-tier hierarchy with a cheap working tier** — an EC subordinate signed by your root by reference, both in one chart, issuing 7-day workload certificates in SHORT_LIVED_CERTIFICATE mode with path length 0. Meshes and batch workloads rotate through the subordinate while the root stays cold. Start from the **Short-Lived Workload mTLS CA** preset.

**Declarative long-lived certificates** — `issuedCertificates` entries sign your own CSRs at apply for the certificates ACM cannot model (MSK client certs, service mTLS pairs with externally held keys). The private key stays wherever you generated it; rotation is a new named entry, not an edit.

## Works With

- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) — issues ACM certificates from this CA via `certificate_authority_arn`; pair with `acmRenewalPermission`
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) — the CRL publishing destination, wired via `revocation.crl.s3BucketName`
- [**AWS MSK Cluster**](/infra-catalog/aws-msk-cluster) — consumes the CA ARN in its TLS client-authentication configuration
