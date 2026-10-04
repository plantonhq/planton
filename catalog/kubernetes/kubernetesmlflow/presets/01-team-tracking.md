# Team tracking server — zero dependencies

The one-manifest MLflow: experiments, runs, metrics and the model
registry on a sqlite volume, artifacts on a second volume served
through the tracking server itself — nothing else to operate, and still
secured (basic auth is on by default with a module-generated admin
password in the `mlflow-admin-auth` Secret; MLflow's own default is an
OPEN server, which never ships from here).

Point your training code at the tracking endpoint from the
outputs:

    MLFLOW_TRACKING_URI=http://mlflow.mlflow.svc.cluster.local:5000
    MLFLOW_TRACKING_USERNAME=admin
    MLFLOW_TRACKING_PASSWORD=<from the mlflow-admin-auth Secret>

Because the server proxies artifact traffic, that is the ONLY
configuration clients need — no S3 keys in notebooks or pipelines.

Honest limits of this shape: sqlite is a single-writer file, so the
server stays at one replica, and metric-heavy experiment fleets will
outgrow it. When that day comes, move to the production preset —
PostgreSQL backend, object-store artifacts, multiple replicas.

Change first: `auth.default_permission` — the default READ lets every
authenticated user see all experiments; NO_PERMISSIONS makes each
experiment private to its creator until shared.

See [01-team-tracking.yaml](./01-team-tracking.yaml) for the manifest.
