# GcpWorkloadIdentityPool Guide

The judgment this guide protects: the pool is a TRUST BOUNDARY with a
30-day memory. Its ID is baked into every principal and grant that
federates through it, deletion reserves the ID for a month, and its
operating mode cannot change — so pools are created deliberately and
almost never destroyed.

## One pool, many providers

The pool holds no issuer configuration — attach one
GcpWorkloadIdentityPoolProvider per external issuer (GitHub, GitLab,
AWS, on-prem SAML) and grant the pool's principals on service accounts
or directly in IAM. One pool per trust boundary (org CI, partner
workloads), not one per repo: grants reference the pool's stable
resource name, and boundary sprawl multiplies audit surface.

## Disable, don't delete

`disabled: true` is the kill switch — token exchanges stop instantly,
existing tokens stop granting, and flipping it back restores everything.
Deletion starts a ~30-day soft-delete clock during which the ID cannot
be reused, and there is NO undelete-on-create (recreating with the same
ID fails outright). For rotation, incident response, or investigations:
disable. Delete only when the boundary itself is retired.

## The mode is set in stone — by the API, not the plan

FEDERATION_ONLY (default) federates external identities;
TRUST_DOMAIN turns the pool into a managed-identity issuer for Google
Cloud workloads. The provider will happily PLAN a mode change and the
API will fail the apply ("Attempted to update an immutable field") —
a documented plan/apply mismatch. Changing modes is a new pool.

## The managed-identity half (TRUST_DOMAIN pools)

`attestationRules` decides WHICH workloads may receive an identity (max
50 rules; GCP applies them in a second API call after create, so a
failed apply can leave a pool without rules — re-apply converges).
`inlineCertificateIssuanceConfig` decides who SIGNS their mTLS
certificates: exactly one of your own CA Service pools (`caPools`,
region-keyed, each value a `GcpPrivateCaPool` reference or the pool's
full name) or the zero-setup GCP shared CA
(`useDefaultSharedCa`). `inlineTrustConfig` extends trust to foreign
trust domains — note GCP requires at least one PEM anchor per bundle
even when `trustDefaultSharedCa` is on: the shared CA is added to a
bundle, never a substitute for one.

## Certificate lifetimes are a blast-radius dial

Shorter `lifetime` (floor 86400s) shrinks what a leaked workload
certificate is worth, at the cost of more rotation traffic;
`rotationWindowPercentage` 50 (rotate at half-life) is the right default
— raise it only if workloads tolerate very tight rotation windows.

## Teardown discipline

`PREVENT` is the right policy for any pool with live grants — losing the
boundary invalidates every principal built from its name. `ABANDON`
keeps federation running while dropping management (the handoff path).
`DELETE` is for pools that never left development — and even then the ID
stays reserved for ~30 days.
