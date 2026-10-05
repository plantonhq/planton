# Team notebooks — shared-password sign-in

The fastest way to a real multi-user notebook platform: every teammate
signs in with their own username and ONE shared password
(module-generated into the `team-notebooks-auth` Secret — the chart's
own default of "any username, no password" never ships), gets their own
JupyterLab server pod and a private 10Gi home volume that survives
restarts, and idle servers stop themselves after an hour.

Reach it over the port-forward command in the outputs, or compose
a KubernetesService/Gateway route over the exported `proxy-public`
service handle for a real URL.

Usernames matter even with a shared password: the username IS the
identity — it names the home volume and carries admin rights — so agree
on consistent names (or pin the roster with `authentication.allowed_users`).

Change first: the notebook image (`single_user.image`) to the stack
your team actually works in (PyTorch, R, Julia — any docker-stacks
image), and `single_user.memory_limit` to your real per-user budget.
Growing team? Move sign-in to the `oidc` arm (Keycloak/Okta) and the
hub database to the production preset's PostgreSQL arm.

See [01-team-notebooks.yaml](./01-team-notebooks.yaml) for the manifest.
