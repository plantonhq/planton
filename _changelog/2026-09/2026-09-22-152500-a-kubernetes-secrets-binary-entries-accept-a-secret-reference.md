# A Kubernetes Secret's binary entries accept a secret reference

**Date**: September 22, 2026
**Type**: Bug fix
**Components**: `catalog/kubernetes/kubernetessecret` (`v1alpha1/spec.proto`, the regenerated stub and reference page); `pkg/protodocs/index.json.gz`

## Summary

`KubernetesSecret`'s `opaque.binaryData` is the one place a base64 payload -- a CA keypair, a keystore, a DER certificate -- enters a Secret in the exact wire form the Kubernetes API stores. Its value rule accepted canonical base64 and nothing else, so a value held in a platform's secret store could never reach it: a manifest carries such a value only as a reference token (`$secret/...`, `$var/...`), and the token's characters are not base64. The rule now has two arms, the literal base64 it always had and a reference token, so a platform resolves the token before the module runs and the module writes the base64 it receives, unchanged. A malformed literal is still refused before any apply, and a string that merely resembles a token (a bare sigil, an unknown prefix) is still a malformed literal.

## What Changed

- **Schema**: the `binary_data` map's value pattern is `^(?:\$(?:secret|var)/.+|<canonical base64>)$`; the field's documentation says why the arm exists and what a platform does with it.
- **Tests**: `spec_test.go` accepts `$secret/@<env>/<slug>`, `$secret/<slug>`, and `$var/<group>/<entry>` values and refuses `$secret/`, `$var/`, `$token/abc`, `secret/abc`; the existing malformed-base64 refusal stands.
- **Generated**: `spec.pb.go`, the kind's `reference.md`, and the proto-docs index.

## Verification

`buf lint` on the proto; `go test ./catalog/kubernetes/kubernetessecret/v1alpha1/` (every case green, the two new ones included); `go test ./pkg/explain/refgen/ ./pkg/protodocs/` (the reference and docs-index freshness gates green after `make generate-reference`). Not run here: the protovalidate-java conformance gate (`make protos` regenerates the gitignored Java stubs it compiles against, a whole-tree lane); the change is one `string.pattern` adding a non-capturing group and an escaped dollar, both plain RE2, and no CEL expression moved. The first platform pinned to this release validates the rule on the Java engine when it accepts a `KubernetesSecret` carrying a reference.
