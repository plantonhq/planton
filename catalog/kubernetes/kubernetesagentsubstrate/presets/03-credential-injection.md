# Credential Injection

Agents call outside services without ever holding the credential. The egress gateway intercepts an agent's https request. When the agent's egress policy says so, it injects a credential read from a Kubernetes Secret, through upstream's bundled Secrets provider. An egress policy names the Secret as `ate-secret://k8s.io/default/<namespace>/<secret>/<key>`.

The provider can read every Secret on the cluster, so it answers only for the namespaces each atespace is granted; with no grants, nothing resolves. Here the `team-a` atespace may use Secrets in `team-a-credentials`.

Pick this when agents call APIs that need keys, and those keys should stay out of every agent's memory and snapshots.

Change first:
- `grants`: one entry per atespace, naming the namespaces whose Secrets its agents may use.
- The platform, database, storage, owner and router clients, as in the **Local or Generic** or **GKE Keyless** preset.

See [03-credential-injection.yaml](./03-credential-injection.yaml) for the manifest.
