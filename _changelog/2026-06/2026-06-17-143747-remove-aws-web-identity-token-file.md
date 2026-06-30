# Remove the AWS web-identity token-file source

**Date**: June 17, 2026
**Type**: Removal (API surface cleanup)
**Components**: AWS Provider, Provider Framework, API Definitions

## Summary

Removed `web_identity_token_file` (field 6) from `AwsWebIdentityProviderConfig`, along with
the `token_xor_file` message-level CEL, the classic builder's file branch, and the
`awswebidentity.Validate` file-rejection. `web_identity_token` is `required` again -- the
inline JWT is once more the single, uniform web-identity token contract across every engine.
The field is removed outright (not reserved): it shipped only in the immediately preceding tag
and has zero consumers, so there is no deployed state to protect against number/name reuse.

## Motivation

The file-based token source was added (in the prior change) to let a long-running stack job
refresh credentials by having the pulumi-aws "classic" provider re-read the token file. The
consuming runner has since adopted a simpler, uniform approach: it **re-mints a fresh inline
JWT before each pulumi operation**. Because a stack job's pulumi operations (`refresh`,
`update`, ...) each re-run the program and re-exchange the token within a single runner
process, re-minting per operation removes any dependency on a single token's TTL -- without a
token ever touching disk.

That makes `web_identity_token_file` not just unused but actively misleading: it was honored
**only** by the classic provider (the builder-side exchange engines -- aws-native, tofu -- are
one-shot and rejected it), so its mere presence in the shared config invited callers and
coding agents to assume a file path works everywhere. Removing it restores a single token
contract and deletes a foot-gun.

## What changed

- **`apis/dev/planton/provider/aws/provider.proto`**: deleted `web_identity_token_file = 6`
  and the `aws.web_identity.token_xor_file` message-level CEL; restored
  `web_identity_token = 1 [(buf.validate.field).required = true]`. Updated the message doc to
  describe the re-mint-per-operation model.
- **`pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider/provider.go`**: the
  web-identity arm requires `role_arn` + `web_identity_token` and always maps the inline token
  to `ProviderAssumeRoleWithWebIdentity.WebIdentityToken`. The `chained_assume_roles` block is
  unchanged, so cross-account-trust still works.
- **`pkg/iac/provider/aws/awswebidentity/exchange.go`**: removed the `web_identity_token_file`
  rejection from `Validate` (the field no longer exists).
- **Tests**: dropped the file-path / both-or-neither builder cases, the `token_xor_file`
  protovalidate test (file deleted -- it had no other cases), and the `Validate` file-rejection
  case.

## Impact

- **Wire/JSON**: removing a field is technically a breaking change. `web_identity_token_file`
  shipped only in the immediately preceding tag and has **zero consumers** (the Planton runner
  only ever set the inline `web_identity_token`), so the practical blast radius is nil and the
  field number is not reserved. (planton's `buf breaking` config exists but is not enforced in
  CI or `make build`/`make release`; this removal is intentional and approved.)
- **No connect-side / runner code change** is required to consume this: the consuming repo
  never referenced the file field.

## Validation

- `make -C apis build` (buf lint + fmt + generate) regenerated the Go + Java stubs.
- `go test` green for `pulumiawsprovider`, `awswebidentity`, and the secret-coverage gate.
- `bazel run //:gazelle` + `bazel build` of the Java stubs and the affected Go packages green.

## Related Work

- Supersedes the file-based token source from
  `2026-06-16-134341-aws-keyless-route53dnsrecord-and-web-identity-token-file.md` (the
  route53dnsrecord builder convergence and overview.md cleanup from that change are retained).

---

**Status**: ✅ Production Ready
