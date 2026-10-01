# GcpBinaryAuthorizationPolicy Guide

The judgment this guide protects: a policy is the last gate before an image runs, and destroying it opens the gate. Roll out in dry run, keep Google's system images admitted, and protect production policies from destroy.

## How a policy is evaluated

For every pod creation in an enforcing GKE cluster (`binaryAuthorizationEvaluationMode: PROJECT_SINGLETON_POLICY_ENFORCE`), Binary Authorization checks, in order:

1. Google's global policy for its own system images, when `globalPolicyEvaluationMode` is `ENABLE`.
2. `admissionWhitelistPatterns` -- matching images are admitted.
3. The cluster's own rule in `clusterAdmissionRules`, keyed `{location}.{cluster_name}` (zone or region), else `defaultAdmissionRule`.

A rule's `evaluationMode` is `ALWAYS_ALLOW`, `ALWAYS_DENY`, or `REQUIRE_ATTESTATION` -- every attestor in `requireAttestationsBy` must have signed the image's digest. Its `enforcementMode` decides whether a denial blocks the pod (`ENFORCED_BLOCK_AND_AUDIT_LOG`) or is only logged (`DRYRUN_AUDIT_LOG_ONLY`).

## Rolling out

1. Create the attestors (`GcpBinaryAuthorizationAttestor`) and teach the build pipeline to sign.
2. Apply the policy with `REQUIRE_ATTESTATION`, `DRYRUN_AUDIT_LOG_ONLY`, and `globalPolicyEvaluationMode: ENABLE`.
3. Read the dry-run denials in Cloud Audit Logs; fix pipelines, add patterns for images nobody signs.
4. Switch to `ENFORCED_BLOCK_AND_AUDIT_LOG`, cluster by cluster if needed.

## Attestor references

`requireAttestationsBy` takes `GcpBinaryAuthorizationAttestor` references (the full `projects/{project}/attestors/{name}`). A bare attestor name resolves to the policy's project. An attestor in another project needs `roles/binaryauthorization.attestorsVerifier` granted to this project's Binary Authorization service agent on the attestor.

## The lifecycle

Google keeps exactly one policy per project. Applying replaces it completely -- every apply sends the whole policy, so rules set outside this block are removed. Destroy under the default `deletionPolicy` writes Google's default back: allow every image, enforced, `gcr.io/google_containers/*` exempt. Production policies should set `PREVENT`. The rule maps Google's API has beyond clusters (Kubernetes namespace, service account, and Istio identity rules) are not exposed by Google's Terraform provider, and an apply removes them if set elsewhere.
