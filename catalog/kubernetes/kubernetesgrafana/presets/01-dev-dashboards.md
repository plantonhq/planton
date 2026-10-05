# Dev dashboards preset

The smallest useful Grafana: one ephemeral replica, chart-generated
admin credentials, and a single provisioned Prometheus datasource —
a real dashboard endpoint for a dev loop without any production
ceremony. Sign in with the credentials from the chart-owned
`dev-grafana` Secret (name and a port-forward command land in the
outputs); the datasource is present from first boot because it
is provisioned as code, not clicked together.

Know what ephemeral means here: Grafana keeps hand-made dashboards,
users and preferences in an embedded SQLite database on the pod's
local disk, and this preset gives that database no volume. A pod
restart — a node drain, an upgrade, an eviction — erases everything
built by hand. The declared datasource survives (it is configuration,
re-rendered on every boot); nothing else does. That trade is correct
exactly as long as the UI is a window, not a workshop.

Change first: declare `storage` (a 10Gi PVC is the spec default) the
moment anyone builds a dashboard worth keeping, and switch the
datasource `url` from the literal to a `value_from` reference at a
KubernetesKubePrometheusStack resource so the wiring follows the
stack instead of hard-coding its endpoint.

See [01-dev-dashboards.yaml](./01-dev-dashboards.yaml) for the
manifest.
