# Local or Generic

The control plane on any conformant cluster, kind included. Upstream's bundled PostgreSQL holds Substrate's records on a 20Gi volume, set up for you and reached with certificates rather than passwords. Snapshots of suspended agents go to an S3-compatible bucket, here a **SeaweedFS** by reference. Agents' outbound calls get no injected credentials. One workload of your product owns the API and calls the agents.

Pick this to evaluate Substrate, run it on a development cluster, or anywhere off GKE. The cluster must serve the pod-certificate APIs (Kubernetes 1.37 with the feature gates).

Change first:
- `serviceAccountIssuer` to your cluster's (`kubectl get --raw /.well-known/openid-configuration`); kind and kubeadm clusters use the value shown.
- The S3 store: point `endpoint` at your store and put its access key in a **Kubernetes Secret** named `substrate-s3` in `ate-system`, with the keys `access_key_id` and `secret_access_key`.
- `bootstrapOwners`: your product's service account as its API token names it, `system:serviceaccount:<namespace>:<name>`. It is the API's first owner and grants the rest.
- `allowedClientSpiffeIds`: the workloads that call agents, as their pod certificates name them, `spiffe://cluster.local/ns/<namespace>/sa/<name>`. The router lets no one else through.
- `volumeSize` to the size of your fleet's records.

See [01-local-or-generic.yaml](./01-local-or-generic.yaml) for the manifest.
