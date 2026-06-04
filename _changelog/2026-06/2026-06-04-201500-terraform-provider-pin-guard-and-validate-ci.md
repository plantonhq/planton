# Terraform Provider-Pin Guard + Static Validate CI; Close Remaining Unpinned Modules

**Date**: June 4, 2026
**Type**: Enhancement (CI hardening) + Bug Fix (latent unpinned providers)
**Components**: CI/CD, Kubernetes/GCP Terraform modules

## Summary

The helm-provider-v3 incident (ExternalDNS) reached production because the failure is
**static** (`tofu init` floats an unpinned provider to its latest major; `tofu validate`
rejects the now-invalid config) yet nothing static ran in CI: PRs only *compile* e2e tests,
there is no repo-wide `tofu validate`, `.terraform.lock.hcl` is gitignored, and the affected
component was an e2e `skip` (needs real cloud DNS, so it was never `apply`-tested). This
change adds a static Terraform gate to the build lane and closes the remaining unpinned
modules so the gate is green and blocking.

## Why e2e didn't catch it

The forge agent correctly marked `KubernetesExternalDns` and `KubernetesCertManager` e2e
profiles as `status: skip` ("requires cloud DNS provider_config" / "ACME + DnsProvider"),
because a full lifecycle needs real DNS. But the helm v3 break is a config-decode error that
surfaces at `tofu init`/`validate` -- long before any DNS call -- so it was catchable with
**zero infrastructure**. The skip removed both the expensive `apply` (legitimately) and the
cheap static check (the gap this closes).

## What's new

### 1. Provider-pin guard (root cause, static, repo-wide, PR-blocking)

`hack/guards/ensure_tf_provider_pins.sh` fails if any `apis/**/v1/iac/tf` module references a
provider (`resource "<p>_..."` / `data "<p>_..."`) without declaring it in
`required_providers`. Unpinned providers are exactly what let `tofu init` float to a new major.
No network/cluster/creds, so it covers `skip`/`deferred` components too.

### 2. `tofu validate` in CI (`.github/workflows/lint.terraform-modules.yaml`)

- **tf-provider-pins** (PR + push): runs the guard.
- **tf-validate-changed** (PR): `tofu fmt -check` + `tofu init -backend=false` + `tofu validate`
  on the iac/tf modules changed in the PR -- fast feedback, catches malformed HCL at authoring.
- **tf-validate-all** (nightly + manual): init + validate across all modules -- the net for a
  provider-release breaking an UNCHANGED module (the ExternalDns scenario). Reports failures;
  does not gate PRs.

### 3. Closed the remaining unpinned modules (22)

Same latent class as ExternalDns, found while wiring the guard:

- **18 Kubernetes CRD-projection modules** used `kubernetes_manifest` with a bare
  `provider "kubernetes" {}` and no `required_providers` -- added a `hashicorp/kubernetes ~> 2.35`
  pin: authorizationpolicy, certificate, clusterissuer, destinationrule, envoyfilter, gateway,
  gatewayclass, grpcroute, httproute, issuer, peerauthentication, prometheus, referencegrant,
  requestauthentication, serviceentry, tcproute, telemetry, tlsroute.
- **4 modules** used `random_*` without pinning `hashicorp/random` -- added `~> 3.6`:
  `gcp/gcpartifactregistryrepo`, `kubernetes/{jenkins,mongodb,redis}`.

After this, all 377 tofu modules pin every provider they reference (guard verified).

## Known follow-ups (not in this change)

`tf-validate-all` will report two **pre-existing, helm-unrelated** `tofu validate` failures
(conditional type mismatches): `kubernetesrookcephcluster` (`var.spec.cluster.network`) and
`kubernetesgharunnerscaleset` (github auth). Their helm v3 init resolves cleanly; flagged for
the parity sweep.

## Validation

- `bash hack/guards/ensure_tf_provider_pins.sh` passes (377 modules); a negative test
  (a temp unpinned `helm_release`) is correctly flagged and fails the guard.
- Workflow YAML and guard script pass syntax checks.
