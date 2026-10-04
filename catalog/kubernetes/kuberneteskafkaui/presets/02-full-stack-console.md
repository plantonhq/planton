# Full stack console preset

The whole Kafka family in one pane: a TLS + SCRAM cluster connection,
schema browsing through the registry, Connect pipe monitoring, and a
login gate on the console itself. This is the shape an Infra Chart
composes from siblings — the three addresses are foreign keys to
KubernetesKafka (bootstrap endpoint), KubernetesKarapace (registry
endpoint) and KubernetesKafkaConnect (REST endpoint), and the two
credential Secrets come from KubernetesKafkaUser and the module.

The credential story is the part worth trusting: no password in this
manifest survives into rendered configuration. The SASL password is a
reference to an existing Secret; the console login password is the
one literal, and the module materializes it into the
`<name>-secrets` Secret — both reach the app as Secret-backed
environment variables that Spring resolves inside the container.

The login is deliberately ONE account: kafbat UI's form login
authenticates against Spring's single default security user (verified
in the app source — there is no multi-user store behind LOGIN_FORM).
Teams needing per-person accounts wire OAuth2/OIDC or LDAP through
`helm_values`. With auth in place, exposing the console is a
composed KubernetesIngress or Gateway route against the exported
`service_name`.

See [02-full-stack-console.yaml](./02-full-stack-console.yaml) for
the manifest.
