# Private Repo With Secrets

This preset is the production shape for a chart from a private repository: repository credentials for the pull, `set_sensitive` for a secret chart value, and the two lifecycle knobs (`atomic`, `cleanup_on_fail`) that keep a failed deploy from leaving the release half-installed. Every value specific to your chart is a placeholder.

**Before reaching for this kind at all:** if the catalog has a first-class kind for what you're deploying, use it instead. KubernetesHelmRelease is the intentional passthrough for charts no kind covers.

## When to Use

- Vendor or internal charts served from a credential-protected HTTPS repository or OCI registry
- Charts that take secrets (API keys, passwords) as chart values — `set_sensitive` keeps them out of rendered plans and state where each engine supports it
- Production installs where a failed upgrade must roll back cleanly rather than strand resources

## Key Configuration Choices

- **`repository_username` / `repository_password`** — HTTP basic auth for a private HTTPS repository, or the registry login for a private OCI registry. The spec requires them set together (both or neither); the password is marked sensitive and masked in state
- **`set_sensitive`** — secret chart values as dotted-path overrides, always literal strings (Helm `--set-string` semantics, no type coercion) and the highest-precedence values layer. On Terraform the entries are masked individually; on Pulumi the presence of any `set_sensitive` entry marks the whole merged values map secret in state — coarser, but safe, and the installed release is identical
- **`atomic: true`** — a failed install or upgrade rolls everything back and purges new resources; the release is never left half-deployed. Atomic implies waiting for readiness, so it cannot be combined with `skip_await` (the spec rejects that pairing)
- **`cleanup_on_fail: true`** — on a failed upgrade, resources that upgrade newly created are deleted. Redundant-sounding next to `atomic`, but it covers the rollback path itself: cleanup happens even as atomic rolls back
- **`version`** — required and pinned, as always; private charts are no exception

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `https://charts.example.com` | Your private chart repository URL (`https://` or `oci://`) | Vendor documentation or your registry |
| `<your-chart-name>` | Chart name within the repository | Vendor documentation |
| `<your-chart-version>` | Exact chart version to install (e.g. `1.4.2`) | `helm search repo` or vendor release notes |
| `<your-repository-username>` | Username for the repository / registry login | Your credential store |
| `<your-repository-password>` | Password or token for the repository | Your credential store |
| `auth.apiKey` / `<your-api-key>` | Replace with your chart's actual secret value path and the secret itself | The chart's values documentation |

## Related Presets

- **01-https-repo-chart** — the no-credentials baseline with a `values_yaml` block
- **02-oci-registry-chart** — the OCI registry form (private OCI registries use the same credential fields shown here)
