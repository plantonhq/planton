# Secret Backends — Where Secrets Live, and How Planton Signs In to Your Store

Read this when a person wants their secrets in their own AWS Secrets Manager, Google Cloud Secret Manager or Azure Key Vault; when they ask how Planton reaches that store or whether it keeps a key to it; when a backend's verify fails; when a connection's delete is refused because a backend signs in through it; when deleting a secret is refused because a backend signs in with it; or when they want to move a backend off a pasted key.

## The doctrine in one sentence

A cloud backend signs in through one of the organization's cloud connections, and with a keyless connection Planton stores nothing about it — never propose pasting a long-lived key when a keyless connection would do.

Secret values are read and written by Planton's control plane itself, never by a runner. Everything below follows from that: only connections the control plane can sign in as can serve, and the control plane mints the keyless token itself.

## The three ways a cloud backend signs in (`spec.auth_mode`)

| Console label | `auth_mode` | What the backend holds | Offer it when |
|---|---|---|---|
| **Cloud Connection** | `connection` | The slug of an AWS, Google Cloud or Azure connection (`aws_secrets_manager.aws_connection`, `gcp_secret_manager.gcp_connection`, `azure_key_vault.azure_connection`) | Always first. Keyless stores nothing |
| **Inline Credentials** | `inline` (or unset) | A key sent on create/update, kept in Planton's credential store and masked as `***` in every response | The person has no connection and will not make one |
| **Ambient Identity** | `ambient` | Nothing, or a handle that picks an identity on a machine signed in to several (`profile`, `configuration`, `subscription`) | A local instance (the developer's own CLI sign-in) or a self-hosted install (the cluster's workload identity). **Never on Planton's hosted service** — it is refused there |

OpenBAO and HashiCorp Vault backends sign in with a token stored with the backend; the platform and local backends need no sign-in. Check what the install offers before suggesting ambient: `SecretBackendInstallCapabilities.ambient_auth_available` is false on hosted, and the console hides the card.

## Which connections can serve

| Connection signs in | Serves a secret backend | What happens at a read |
|---|---|---|
| Keyless (`oidc`) | Yes | A new short-lived token is minted for the connection every time the backend's credentials refresh; AWS assumes the role in session `planton-secret-backend-<backend>` (CloudTrail names the backend), Google Cloud exchanges at its token service and impersonates the connection's service account, Azure presents the token as the app registration's client assertion |
| Stored key (`inline`) | Yes | The key is read from the connection's own secrets |
| Runner, Vault broker, browser sign-in | **No** | Refused — the identity lives on a runner, or the sign-in is a person's |

The connection's identity needs permission to manage secrets: the **Secrets Manager** capability on an AWS connection (`secrets_manager`), the **Secret Manager** capability on a Google Cloud connection (`secret_manager`, `roles/secretmanager.admin`). Azure's keyless setup grants no Key Vault role — the person grants one on the vault (Key Vault Secrets Officer, or an equivalent role that manages secrets). When the person has no suitable connection, the console's picker creates a keyless one in place with the capability already chosen.

The person binding a backend to a connection must be able to `get` that connection; environment sharing of the connection is not needed.

## The one-hop rule

A connection's own secrets — a stored key, or an Azure app registration's client ID (Azure reads it even when keyless) — must live in a backend that does **not** itself sign in through a connection. It is checked when a backend is created or updated, when a connection is edited, and again at every read. If an organization's default backend signs in through a connection, the connection's secrets must be created in another backend, named explicitly (a secret with no backend lands in the default).

## Commands

```bash
# Create — the flags say the mode; --auth-mode is optional when they do
planton secret backend create team-secrets --type aws  --aws-region us-east-1 --aws-connection prod-aws
planton secret backend create team-secrets --type gcp  --gcp-project-id my-project --gcp-connection prod-gcp
planton secret backend create team-secrets --type azure --azure-vault-url https://myvault.vault.azure.net --azure-connection prod-azure

# Prove it end to end (exit 1 when unhealthy) — a saved backend, or a manifest before it exists
planton secret backend verify team-secrets
planton secret backend verify -f SecretBackend.team-secrets.yaml

# How each backend signs in ("Sign-In" row / column)
planton secret backend get team-secrets
planton secret backend list

# The backends that sign in through a connection — what keeps it from being deleted
planton connection get aws prod-aws
```

Contradictory flags (a connection beside stored keys, ambient handles beside a connection) are refused before anything is sent. MCP: `apply_secret_backend`, `verify_secret_backend` (pass the same object you will apply; apply on green), `get_secret_backend`, `list_secret_backends`, `find_connection_backends`.

## Moving a backend off a pasted key

`auth_mode` changes in place; `backend_type` and `name_prefix` never do, and remote names depend only on those two, so every secret stays where it is. Verify the draft spec with the connection first, then apply it with `auth_mode: connection`, the connection field set, and the credential fields removed. The stored key is deleted once the change is saved. In the console this is **Switch to Cloud Connection** on the backend's page, which saves only after a green Test Connection. Moving back to inline needs every credential sent fresh.

## The sentences and what to do

Relay these verbatim; each names its remedy.

- *Connection '`<slug>`' signs in as its runner's own identity, which only that runner holds. Secrets are read and written by Planton's control plane, which cannot sign in as it. Point this backend at a keyless (OIDC) or stored-key connection.* — the broker and browser variants read the same way; pick a keyless or stored-key connection.
- *Connection '`<slug>`' keeps its `<field>` in secret '`<secret>`', which is stored in backend '`<backend>`', and that backend also signs in through a connection. Keep a connection's secrets in a backend that uses stored credentials or the platform backend.* — the one-hop rule; recreate the secret in such a backend and point the connection at it, or choose a keyless AWS or Google Cloud connection, which reads no secret.
- *`<Cloud>` connection '`<slug>`' not found in organization '`<org>`'. Create the connection first, then point this backend at it.*
- *Ambient sign-in uses the identity this Planton deployment itself runs as, which Planton's hosted service never lends to an organization. Sign in through one of your cloud connections (auth mode connection), or store a key for this backend (auth mode inline).*
- *`<Cloud>` refused the keyless sign-in through connection '`<slug>`': `<the cloud's refusal>`. `<The trust policy of role … / Workload identity provider … / A federated credential on the app registration …>` must admit subject '`<sub>`' from issuer '`<iss>`'.* — verify's reachability check; the fix is one edit to the trust in the person's cloud, admitting exactly that subject.
- *Connection '`<slug>`' signs in to the store that keeps the secrets of secret backend '`<backend>`'. Deleting it would make those secrets unreadable. Point each backend at another connection first.* — a refused connection delete (state backends add their own clauses); move each backend, then delete.
- *Secret '`<secret>`' signs secret backend '`<backend>`' in to its store through a connection, so deleting it would leave it unable to read or write its secrets. Point the backend at another connection, or change how it signs in, first.* — a refused secret delete.
- *Version '`<id>`' is the latest value of secret '`<secret>`', which signs secret backend '`<backend>`' in to its store through a connection; deleting it would sign it in with an older value. Write the new value as a new version instead.* — rotating a stored key is writing a new version; the backends sign in again with it on their next read.

## What never to do

- Never suggest Ambient Identity on Planton's hosted service.
- Never suggest a runner, Vault broker or browser connection for a secret backend.
- Never put a connection's key in a backend that signs in through a connection.
- Never tell a person their secrets move when the backend's sign-in changes — they stay exactly where they are.
- Never delete a connection, or the secret a stored-key connection reads, to "reset" a backend; move the backend first.
