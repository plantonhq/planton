# Secret Backends — Where Secrets Live, and How Planton Signs In to Your Store

Read this when a person wants their secrets in their own AWS Secrets Manager, Google Cloud Secret Manager or Azure Key Vault; when they ask where their secrets live, how Planton reaches that store or whether it keeps a key to it; when they are on Planton Desktop or a self-hosted install and ask what backend to use; when they ask where a new secret lands; when a backend's verify fails; when a create is refused because the install cannot serve that backend; when a self-hosted install will not boot because it can serve no backend; when a connection's delete is refused because a backend signs in through it; when deleting a secret is refused because a backend signs in with it; or when they want to move a backend off a pasted key.

## The doctrine in one sentence

A cloud backend signs in through one of the organization's cloud connections — keyless wherever the install offers it, which stores nothing; on Planton Desktop, which offers no keyless connections, through the person's own cloud CLI login — and never through a pasted long-lived key when either would do.

Secret values are read and written by Planton's control plane itself, never by a runner. Everything below follows from that: only connections the control plane can sign in as can serve, and the control plane mints the keyless token itself.

## Where you are decides what is offered

Find out first, before suggesting anything: `planton instance show` reads `deployment_kind` (`hosted`, `self_hosted` or `local` — the desktop), and `planton secret backend list` shows the organization's backends with their TYPE, SIGN-IN and DEFAULT columns. On a self-hosted install the default's type tells you whether the platform vault runs: `platform` means it does, `local` means it is off (`spec.vault.enabled: false` on the PlantonPlatform).

| Where | The default backend | `platform` (OpenBAO) | `local` (built-in) | Ambient Identity | Keyless connections |
|---|---|---|---|---|---|
| Planton's hosted service | `platform-default` — the Planton-operated OpenBAO | Offered | Refused | Refused | Offered |
| Planton Desktop | `local-default` — encrypted in the instance's database, key in the OS keychain | Refused | Offered | The person's cloud CLI login | **Refused** — the instance's issuer is a loopback address no cloud can reach |
| Self-hosted, vault on (the default) | `platform-default` — the install's bundled OpenBAO | Offered | Refused | The control plane's workload identity | Offered when the front door is public and HTTPS |
| Self-hosted, vault off | `local-default` — encrypted in the platform's database, key in a Kubernetes Secret the operator holds | Refused | Offered | The control plane's workload identity | **Refused** — the vault signs keyless tokens |

A self-hosted install may instead declare its first organization's default (`PLANTON_BOOTSTRAP_SECRET_BACKEND_TYPE`): `platform`, or `aws-secrets-manager` — `aws-secrets-manager-default`, signing in with the control plane's workload identity. Every other organization gets the rule above. AWS Secrets Manager, Google Cloud Secret Manager, Azure Key Vault, OpenBAO and HashiCorp Vault backends are offered everywhere. Never offer a cell marked refused: the create is refused with the sentence below, and the console never shows the card.

## The three ways a cloud backend signs in (`spec.auth_mode`)

| Console label | `auth_mode` | What the backend holds | Offer it when |
|---|---|---|---|
| **Cloud Connection** | `connection` | The slug of an AWS, Google Cloud or Azure connection (`aws_secrets_manager.aws_connection`, `gcp_secret_manager.gcp_connection`, `azure_key_vault.azure_connection`) | First on hosted and self-hosted: keyless stores nothing. On Planton Desktop, a stored-key connection — or Ambient Identity, usually simpler there |
| **Inline Credentials** | `inline` (or unset) | A key sent on create/update, kept in Planton's credential store and masked as `***` in every response | The person has no connection and will not make one |
| **Ambient Identity** | `ambient` | Nothing, or a handle that picks an identity on a machine signed in to several (`profile`, `configuration`, `subscription`) | Planton Desktop (the person's own CLI sign-in — the recommended way there) or a self-hosted install (the cluster's workload identity). **Never on Planton's hosted service** — it is refused there |

OpenBAO and HashiCorp Vault backends sign in with a token stored with the backend; the platform and local backends need no sign-in.

### Ambient Identity, by where it runs

- **Planton Desktop** — the control plane signs in with the person's own cloud CLI login: `aws` (the default chain, or the profile named in `aws_secrets_manager.profile`), `gcloud` (Application Default Credentials, or the configuration named in `gcp_secret_manager.configuration`), `az` (the default chain, or the subscription named in `azure_key_vault.subscription`). On a machine signed in to several identities, name the handle — `--aws-profile`, `--gcp-configuration`, `--azure-subscription` on `planton secret backend create`. When verify fails, its sentence names the remedy: `aws sso login`, `gcloud auth application-default login` or `az login`; relay it as written.
- **Self-hosted** — the control plane's own pod identity through the cloud SDK's default chain: IRSA on EKS, Workload Identity on GKE, a managed identity on AKS, or the node's instance profile. That identity — the control plane's Kubernetes service account, mapped to the cloud — needs the store permissions; the handles are ignored there.

Either way, the identity needs to manage secrets in the store: Secrets Manager permissions on AWS (`secretsmanager:*`, or a policy scoped to the backend's name prefix), Secret Manager Admin on Google Cloud (`roles/secretmanager.admin` on the project), Key Vault Secrets Officer on the Azure vault.
- **Planton's hosted service** — refused: the control plane's identity is Planton's own.

## Which connections can serve

| Connection signs in | Serves a secret backend | What happens at a read |
|---|---|---|
| Keyless (`oidc`) | Yes, where the install offers keyless connections (not on Planton Desktop; on self-hosted only with a public HTTPS front door and the vault on) | A new short-lived token is minted for the connection every time the backend's credentials refresh; AWS assumes the role in session `planton-secret-backend-<backend>` (CloudTrail names the backend), Google Cloud exchanges at its token service and impersonates the connection's service account, Azure presents the token as the app registration's client assertion |
| Stored key (`inline`) | Yes | The key is read from the connection's own secrets |
| Runner, Vault broker, browser sign-in | **No** | Refused — the identity lives on a runner, or the sign-in is a person's |

The connection's identity needs permission to manage secrets: the **Secrets Manager** capability on an AWS connection (`secrets_manager`), the **Secret Manager** capability on a Google Cloud connection (`secret_manager`, `roles/secretmanager.admin`). Azure's keyless setup grants no Key Vault role — the person grants one on the vault (Key Vault Secrets Officer, or an equivalent role that manages secrets). When the person has no suitable connection, the console's picker creates one in place with the capability already chosen: keyless where the install offers it, a stored key where it does not. Whether keyless is offered is the install's own answer — the connection wizard shows a keyless card it cannot honor disabled, with one of these sentences, and the same sentence is the refusal of a keyless connection create:

- Planton Desktop: *Keyless connections need cloud providers to fetch this instance's signing keys over the public internet, and this instance runs on your machine at a loopback address they cannot reach. Use the runner or access-key method instead.* — for a secret backend, take the access-key (stored-key) method or Ambient Identity; a runner connection cannot serve one.
- Self-hosted, each naming its one cause: the front door is *a port-forward on your own machine*, is *declared private*, or *serves plain HTTP; cloud providers accept only an HTTPS issuer* (add a certificate to the front door) — each after *Keyless connections need cloud providers to fetch this install's signing keys from its front door over the public internet, …*; or *Keyless connections are signed by the platform vault, which this install runs without. Enable the vault, or use the runner or access-key method instead.*

The person binding a backend to a connection must be able to `get` that connection; environment sharing of the connection is not needed.

## Which backend a secret lands in

A secret is stored in the backend it names when it is created, or in the organization's default when it names none: `planton secret set <slug> --backend <backend>` (the flag is read only when the secret is created). The binding never changes afterward — to move a secret, create it in the target backend and copy the value, then delete the old one. Making another backend the default (`planton secret backend set-default <backend>`) changes where *new* secrets land; every existing secret stays where it is. With no default and no backend named, the create is refused: *No default SecretBackend configured for org '`<org>`'. Either set a default backend or specify one explicitly in spec.backend.* Where a secret's values physically live, under which names and labels, is in the open-source site's `site/public/docs/secrets/where-secrets-live.md`.

## The one-hop rule

A connection's own secrets — a stored key, or an Azure app registration's client ID (Azure reads it even when keyless) — must live in a backend that does **not** itself sign in through a connection. It is checked when a backend is created or updated, when a connection is edited, and again at every read. If an organization's default backend signs in through a connection, the connection's secrets must be created in another backend, named explicitly (a secret with no backend lands in the default).

## Commands

```bash
# Create — the flags say the mode; --auth-mode (connection | inline | ambient) is needed only when they do not
planton secret backend create team-secrets --type aws  --aws-region us-east-1 --aws-connection prod-aws
planton secret backend create team-secrets --type gcp  --gcp-project-id my-project --gcp-connection prod-gcp
planton secret backend create team-secrets --type azure --azure-vault-url https://myvault.vault.azure.net --azure-connection prod-azure
# Ambient (Planton Desktop, self-hosted) — no identity flags, so name the mode; a handle pins one identity on a desktop
planton secret backend create team-secrets --type aws --aws-region us-east-1 --auth-mode ambient [--aws-profile acme-prod]

# Change an existing backend — export it, edit the spec, verify the draft, then apply it
planton get SecretBackend team-secrets -o yaml > SecretBackend.team-secrets.yaml
planton secret backend verify -f SecretBackend.team-secrets.yaml
planton apply -f SecretBackend.team-secrets.yaml

# Where new secrets land
planton secret backend set-default team-secrets

# Prove it end to end (exit 1 when unhealthy) — a saved backend, or a manifest before it exists
planton secret backend verify team-secrets
planton secret backend verify -f SecretBackend.team-secrets.yaml

# How each backend signs in ("Sign-In" row / column)
planton secret backend get team-secrets
planton secret backend list

# The backends that sign in through a connection — what keeps it from being deleted
planton connection get aws prod-aws
```

A manifest is `apiVersion: config-manager.planton.ai/v1alpha1`, `kind: SecretBackend`, and its spec fields read in either spelling (`auth_mode` or `authMode`, `gcp_secret_manager.gcp_connection` or `gcpSecretManager.gcpConnection`). `verify -f` takes a draft of a new backend or a changed one. Contradictory flags (a connection beside stored keys, ambient handles beside a connection) are refused before anything is sent. MCP: `apply_secret_backend`, `verify_secret_backend` (pass the same object you will apply; apply on green), `get_secret_backend`, `list_secret_backends`, `find_connection_backends`.

## Moving a backend off a pasted key

`auth_mode` changes in place; `backend_type` and `name_prefix` never do, and remote names depend only on those two, so every secret stays where it is. Verify the draft spec with the connection first, then apply it with `auth_mode: connection`, the connection field set, and the credential fields removed. The stored key is deleted once the change is saved — Planton's copy only: tell the person to revoke the key itself in their cloud (the IAM access key, the service account key, the app registration's client secret) once the backend verifies through the connection. In the console this is **Switch to Cloud Connection** on the backend's page, which saves only after a green Test Connection. Moving back to inline needs every credential sent fresh.

## The sentences and what to do

Relay these verbatim; each names its remedy.

- *The platform secret backend stores secrets in the platform vault, but this install has no platform vault configured. Use the built-in local backend or a cloud secret backend instead.* — Planton Desktop, or a self-hosted install with its vault off; `local` is the built-in choice there.
- *The built-in local secret backend is not offered on this deployment: it holds no key-encryption-key. Use the platform backend or a cloud secret backend.* — Planton's hosted service.
- *The built-in local secret backend needs its key-encryption-key, but PLANTON_LOCAL_SECRETS_KEK is not set. The Planton operator supplies it when the platform's bundled vault is off (spec.vault.enabled: false); with the vault on, use the platform backend or a cloud secret backend.* — a self-hosted install with its vault on.
- *The built-in local secret backend needs its key-encryption-key, but PLANTON_LOCAL_SECRETS_KEK is not set. The local lifecycle daemon supplies it from the OS keychain when Planton starts: unlock the keychain (sign in to your desktop session) and restart the local instance, then retry -- the operation is safe to repeat.* — Planton Desktop; the remedy is in the sentence.
- *The bootstrap organization '`<org>`' needs a default secret backend and this deployment can serve none: the platform vault is off, no key-encryption-key is set (PLANTON_LOCAL_SECRETS_KEK), and no backend is declared (PLANTON_BOOTSTRAP_SECRET_BACKEND_TYPE).* — a self-hosted control plane that will not finish booting; turn the vault back on, or let the operator mint the local key (it does when `spec.vault.enabled: false`), or declare a backend. Read the PlantonPlatform first (`references/self-hosted.reading-a-platform.md`).
- *cannot mint a web-identity token: no platform vault is configured, so the OIDC issuer has no signing key. Keyless connections require the platform vault (on a self-managed install, enable the vault component).* — a keyless sign-in on an install running without its vault; enable the vault, or choose a stored-key connection.
- *Connection '`<slug>`' signs in as its runner's own identity, which only that runner holds. Secrets are read and written by Planton's control plane, which cannot sign in as it. Point this backend at a keyless (OIDC) or stored-key connection.* — the broker and browser variants read the same way; pick a stored-key connection, or a keyless one where the install offers it.
- *Connection '`<slug>`' keeps its `<field>` in secret '`<secret>`', which is stored in backend '`<backend>`', and that backend also signs in through a connection. Keep a connection's secrets in a backend that uses stored credentials or the platform backend.* — the one-hop rule; "stored credentials or the platform backend" means any backend that does not itself sign in through a connection (inline, ambient, a vault token, `platform`, or `local` where there is no platform vault): recreate the secret in such a backend and point the connection at it, or choose a keyless AWS or Google Cloud connection, which reads no secret, where the install offers keyless.
- *`<Cloud>` connection '`<slug>`' not found in organization '`<org>`'. Create the connection first, then point this backend at it.*
- *Ambient sign-in uses the identity this Planton deployment itself runs as, which Planton's hosted service never lends to an organization. Sign in through one of your cloud connections (auth mode connection), or store a key for this backend (auth mode inline).*
- *`<Cloud>` refused the keyless sign-in through connection '`<slug>`': `<the cloud's refusal>`. `<The trust policy of role … / Workload identity provider … / A federated credential on the app registration …>` must admit subject '`<sub>`' from issuer '`<iss>`'.* — verify's reachability check; the fix is one edit to the trust in the person's cloud, admitting exactly that subject.
- *Connection '`<slug>`' signs in to the store that keeps the secrets of secret backend '`<backend>`'. Deleting it would make those secrets unreadable. Point each backend at another connection first.* — a refused connection delete (state backends add their own clauses); move each backend, then delete.
- *Secret '`<secret>`' signs secret backend '`<backend>`' in to its store through a connection, so deleting it would leave it unable to read or write its secrets. Point the backend at another connection, or change how it signs in, first.* — a refused secret delete.
- *Version '`<id>`' is the latest value of secret '`<secret>`', which signs secret backend '`<backend>`' in to its store through a connection; deleting it would sign it in with an older value. Write the new value as a new version instead.* — rotating a stored key is writing a new version; the backends sign in again with it on their next read.

## What never to do

- Never suggest Ambient Identity on Planton's hosted service.
- Never suggest a keyless connection on Planton Desktop, or on a self-hosted install whose connection wizard shows keyless as not offered; take its sentence's way out.
- Never suggest the `platform` backend where there is no vault, or the `local` backend on Planton's hosted service.
- Never tell a person that changing the default backend moves their secrets — it changes only where new ones land.
- Never suggest a runner, Vault broker or browser connection for a secret backend.
- Never put a connection's key in a backend that signs in through a connection.
- Never tell a person their secrets move when the backend's sign-in changes — they stay exactly where they are.
- Never delete a connection, or the secret a stored-key connection reads, to "reset" a backend; move the backend first.
