---
title: Front Doors and the CLI — Which Door Routes What, and How the CLI Reaches a Self-Hosted Planton
description: The three front doors a self-hosted Planton can serve through, the one fact that trips adopters (only the Gateway API door routes the CLI's native gRPC), and how planton instance add and planton login learn an instance from the instance itself. Read when a person sets up the CLI against their own Planton, when whoami or any planton command fails right after a successful login, when choosing between ingress-nginx and a Gateway API Gateway, or when the desktop's device sign-in is in question.
---

# Front Doors and the CLI

Read this when the CLI is being pointed at a self-hosted Planton, when a command fails right after `planton login` said "Login Successful", or when an adopter is choosing how to publish the platform. Publishing itself is documented for people at `site/public/docs/self-hosting/index.md` ("Publish at your own URL"); the operator's route table is described in `operator/README.md` ("The Front Door"). This file carries what an agent needs to get the CLI working and to explain why a door behaves as it does.

## The doctrine in one sentence

A platform is reached through one public origin rendered onto one of three doors, and the CLI learns everything about an instance from the instance's own discovery document -- except one fact the door decides: whether the CLI's native gRPC calls can reach the control plane at that origin at all.

## The three doors

| Door | Declared by | Routes the browser (console, `/rpc`, `/idp`) | Routes the CLI's native gRPC |
|---|---|---|---|
| Gateway API HTTPRoute on a Gateway the cluster runs | `spec.ingress.enabled: true` with `spec.ingress.gatewayRef` naming the Gateway | yes | **yes** -- an exact `content-type: application/grpc` header match sends it to the control plane's gRPC port |
| Ingress object (ingress-nginx and friends) | `spec.ingress.enabled: true` (+ hostname, TLS) | yes | **no** -- Ingress has no portable header match |
| Built-in nginx gateway over `kubectl port-forward` | the default with no `spec.ingress` | yes | **no** |

The console publishes what the door decided in its device discovery document: `grpcEndpoint` is present only when the door routes native gRPC, and `deploymentKind: self_hosted` says what the instance is.

**Which door to recommend**: if the adopter's engineers will live in the CLI (`planton login`, `planton apply`, `planton directory …`), the Gateway API door -- every command works at the hostname. On the Ingress or built-in door the console and browser sign-in are identical, and the CLI needs the port-forward below for its API calls.

## The CLI against the Ingress or built-in door

The symptom: `planton login` completes, then `planton whoami` (or any command) fails with an Unavailable panel whose raw cause reads like *"error reading server preface: http2: frame too large … looked like an HTTP/1.1 header"*. The address answered as a web server, not as gRPC. The way through:

```bash
kubectl -n <namespace> port-forward svc/<platform>-control-plane 8090:80
```

and in the instance manifest (`planton instance path` prints the directory; the file is `instances/<slug>/instance.yaml`):

```yaml
grpc_endpoint: localhost:8090
```

Newer CLIs say exactly this at `instance add` (a warning when the discovery document declares no gRPC address) and in the Unavailable panel; older ones show the raw cause. Either way the instruction is the same, and you carry it.

## Adding the instance and signing in

```bash
planton instance add --endpoint https://planton.acme.com    # the deployment's URL is the one thing a person knows
planton login
planton whoami
```

`instance add` discovers the sign-in provider (`keycloak` on a self-hosted install), the issuer (`<origin>/idp/realms/planton`), the CLI's OIDC client (`planton-cli`), the deployment kind, and the gRPC address when the door routes it, then verifies the sign-in server live and says so ("sign-in server verified at …"). A plain-HTTP origin is accepted for evaluation with a warning.

`planton login` has two paths and chooses itself:

- **Through the console** (the broker path): the browser opens the console's device picker at `/device/auth`, the person signs in, the console redeems the code and hands the CLI a ticket. This is the same route the desktop app's sign-in takes.
- **Direct** (when the console is not answering): a PKCE sign-in against the identity server on a loopback callback (port 8088). The CLI says which it took: "Auth Method OAuth (broker)" or "OAuth (keycloak)".

If `instance add` ran while the console was down, the manifest is saved broker-only (no issuer, no client id) with a warning; a later `planton login` while the console answers fills it in. If both the console is down and the manifest is broker-only, the CLI says so in one sentence and stops -- there is no other way in until the console answers.

`planton login --local` is the break-glass on an install whose directory is the primary sign-in (`self-hosted.identity-primary-and-break-glass.md`).

## Tokens and the API

Tokens from either path carry the `planton-api` audience; `planton whoami`, `planton can-i`, and every other command use them. A first-ever sign-in from the CLI materializes the person's Planton account during login (the console does the same on page load), so `planton whoami` answers at once. A designed refusal at that moment -- no free seat, an email-less directory account -- is printed as the server's sentence instead of the success banner.

## What was verified, and what to confirm on your install

Verified on a running install: every path above through the CLI, on the built-in door (with the port-forward) and, for the discovery document's `grpcEndpoint`, on a Gateway API door. **Not verified: the desktop application's own sign-in screens.** Its device flow is the CLI's broker path -- the same console picker, callback, and ticket -- so it is expected to behave identically, but nobody watched the desktop's screens themselves. If an adopter's desktop sign-in against their self-hosted Planton does not land signed in with name and email, treat it as a platform gap: file it on the open-source repository with the exact screens and the console's `/device/auth` responses (`craft.filing-platform-gaps.md`).

## What never to do

- Never tell a person their install is broken because `whoami` failed on the Ingress door; it is the door, and the port-forward is the answer.
- Never edit the instance manifest's issuer or client id by hand to "fix" a login; `instance add` and `login` own those facts and heal them from the console.
- Never propose switching doors on a running platform as a quick fix; the front door is the platform's identity issuer, and a change is a declaration change with a rollout behind it.
