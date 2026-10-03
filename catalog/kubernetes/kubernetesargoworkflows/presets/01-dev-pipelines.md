# Dev pipelines preset

The smallest useful Argo Workflows: the engine, the UI, and a runner
identity — submit a Workflow CR (or use the UI over a port-forward;
the command lands in the outputs) and it runs. The server's
`client` auth mode means everyone acts with their own Kubernetes
permissions from day one, which is the right default even in dev:
there is no anonymous power to unlearn later.

Know what the empty seams mean: without an `artifact_repository`,
steps cannot pass files to each other and archived logs have nowhere
to live; without an `archive`, a workflow's record IS its CR — delete
or retention-prune the CR and the history is gone. Both are deliberate
dev-loop trades, not oversights.

Change first: declare `artifact_repository` (an in-cluster
KubernetesSeaweedFs pairs naturally through the endpoint's reference
field) the moment a pipeline has more than one step touching the same
file, and `archive` (a KubernetesPostgres reference) when run history
starts answering questions — annotate the runner ServiceAccount for
IRSA/workload identity when workflows need cloud APIs.

See [01-dev-pipelines.yaml](./01-dev-pipelines.yaml) for the manifest.
