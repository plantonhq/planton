# Dev preset — KubernetesExecutor

The smallest useful Airflow 3: API server (UI + REST API), scheduler,
DAG processor and triggerer against a composed KubernetesPostgres named
`airflow-db` in the same Kubernetes namespace, running tasks with the
KubernetesExecutor — every task is its own ephemeral pod, so there is
no broker to operate and no idle worker fleet. Log in over the
port-forward command in the outputs with the admin user
(password in the `dev-airflow-admin-auth` Secret).

The database references do all the wiring: the host resolves to the
Postgres read-write Service and the credential to the
operator-maintained application Secret, and the module composes the
SQLAlchemy connection Secret from them at deploy time — nothing
password-shaped is ever written into this manifest or into rendered
chart values. What the preset expects from the database side: the
KubernetesPostgres declares `airflow` as its bootstrap database
(owner `airflow`).

DAGs come from the container image here (Airflow's default layout).
For real teams, switch to the production preset's git-sync block —
pipelines deploy on push with no image rebuilds.

Change first: `components` sizing before real load (the scheduler is
the heart — give it CPU), and `logging.persistence` so task logs
survive pod restarts.

See [01-dev-kubernetes-executor.yaml](./01-dev-kubernetes-executor.yaml)
for the manifest.
