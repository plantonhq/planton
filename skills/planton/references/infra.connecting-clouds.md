---
title: Connecting a Cloud — Keyless With the Person's Own Sign-In, Explained the Whole Way
description: How an agent connects AWS, Google Cloud, Azure, DigitalOcean or Cloudflare to Planton for the person -- keyless with the sign-in already on their machine (`planton connect gcp`, `aws`, `azure`), a token for DigitalOcean and Cloudflare -- and keeps them informed and in control throughout -- what each instance offers, the plan to show before anything changes, the facts that answer every worry about access, the grant and its two opt-in tiers, reading the verdict, resuming, verify and cleanup, and every refusal with its fix. Read when the person asks to connect a cloud account, asks what access Planton needs or has, or a connection cannot deploy something.
---

# Connecting a Cloud

Read this when a person asks you to connect their cloud to Planton, asks what Planton can do in their account, or a connection cannot deploy something they asked for. Connecting writes no files in the workspace, so the chart workflow's folder check does not apply.

## The doctrine in one sentence

The person decides, fully informed: you read the plan, explain what Planton will be able to do and that the setup runs as them in their own cloud, ask, run it only on their yes, narrate each step, read the verdict back -- and every reassurance you give is a fact the product makes true, never a comfort you made up.

## Why the person is right to ask questions

A connection gives Planton a way into their cloud. A careful engineer, and every security lead, will want to know exactly what that way allows and how to take it back. Treat each question as the most important part of the job, not an obstacle to it. Connecting is half the work; the person understanding what they connected is the other half.

## What each Planton offers

| Where Planton runs | What the bare `planton connect gcp` (or `aws`, `azure`) does |
|---|---|
| planton.ai | Keyless: the trust is written in the person's cloud with their sign-in, Planton stores no key |
| A self-hosted Planton with a public address | Keyless, the same way |
| Planton Desktop's own instance, or a self-hosted Planton with a private address | Refuses keyless with the instance's reason, and names `planton connect gcp detect` (the Runner Identity method: deploys run with this machine's sign-in directly, with that sign-in's full access) |

DigitalOcean and Cloudflare have no keyless trust: the connection holds an API token (below). The four methods are named the same everywhere: Keyless, Stored Key, Runner Identity, Vault Broker. `aws` and `azure` take the same `--print-documents`, `verify`, `cleanup` and `-o json` as `gcp`; this page shows `gcp` where all three behave alike.

## The journey, in order

1. **Read the plan first; it changes nothing.** `planton connect gcp --project <project> -o json` prints the plan (`"written": false`). Pass every choice the person already made -- `--kinds`, `--organization-access`, `--access-granting`, `--no-authorize`, `--no-default`, `--name`, `--configuration` -- because the plan's `run` command carries exactly what the plan was asked with. Its keys: `runsAs` (the person's sign-in the setup acts as), `grantedTo` (the identity that gains the grant -- a service account, a role, an app), `target` and `targetNoun` (the project, account or subscription), `connection` (the name and slug it will create), `grant` and `kindsGranted` of `kindsInCatalog` (what Planton may deploy), `granted` (each document applied: its `title`, `permissions` count and `predefinedRoles`), `resumes` (an existing keyless connection by that name is picked up), and per script `changes` (every step it takes in their cloud, in order), `keeps` (what other connections share, which it leaves in place) `reachesOthers` (what it also changes for other connections) `reachedConnections` (those connections the person may open, each `{slug, name}` -- say each to the person; the slug is what every command takes) and `unlistedReachedCount` (how many more it reaches that the person may not open: counted, never named -- say the number, and that someone who can open them should know). `readScript` prints the exact script.
2. **Explain, in plain words, with the plan's real values:** the facts in "What to tell the person" below.
3. **Ask, naming the identity, the target and who may use it:** "This runs as dev@acme.com in project acme-prod and creates a role only Planton's connection can use, with what you see in the plan. The connection will be usable by every environment in your organization and becomes its default GCP connection when you have none. Go ahead?"
4. **Run on their yes:** the plan's `run` command (it carries `--yes`). The script runs on this machine, as them; narrate as it goes -- the trust, the role, waiting for the cloud to accept the trust (a new trust can take a few minutes to spread), the connection, Verify.
5. **Read the verdict back** (below), and say what they can deploy now.

A person at the terminal can simply run `planton connect gcp`: it shows the same plan as a card and asks the same question.

- **A request to connect is not a yes to a plan the person has not seen.** When they said outright to go ahead without asking, still show the plan as you run it: this changes who can act in their cloud.
- **`--yes` is the person's yes, given ahead.** Pass it only after they agreed to this plan.
- **`--accept-shared-changes`** is needed with `--yes` when `reachesOthers` is not empty: the role is bound to an identity other connections also sign in as, so the change reaches them too. Pass it only after the person accepted that consequence by name.
- **Target and sign-in flags:** GCP `--project` and `--configuration` (a gcloud named configuration); AWS `--account-id` and `--profile`; Azure `--subscription` (by id or name). Without them the setup uses the cloud CLI's active sign-in and its default project, account or subscription -- never another sign-in it happens to find. The default name is `gcp.project.<project>`, `aws.account.<id>`, `azure.subscription.<id>`; the slug is that name with dots as dashes.

## What to tell the person -- every sentence is a product fact

- **It runs as you, on your machine.** The setup script runs here with the sign-in you already have. Your sign-in never leaves this machine; Planton receives only what the script reports back (on GCP the trust's provider name and the service account, on AWS the role's ARN, on Azure the app's client id, tenant and subscription).
- **Planton keeps no key.** A keyless connection stores no secret. When it deploys, Planton's issuer signs a token for exactly this connection, and the cloud exchanges it for short-lived credentials (about an hour).
- **Only this connection can use the trust.** The trust accepts a token only from this Planton's issuer (planton.ai's, or a self-hosted Planton's own) and only with this connection's subject (`planton:v1:org:<org>:conn:<slug>:provider:<cloud>`); no other connection, organization or Planton can present it. `planton connect gcp verify <slug>` prints the issuer and subject.
- **Say plainly when the identity is shared.** By default the role is granted to an identity every Planton connection in the account may sign in as -- GCP's service account `planton-provisioner@<project>`, Azure's app `planton-provisioner` -- so each connection there can do what all of their roles allow together; the plan's `reachesOthers` names the connections that share it. To keep a connection's role apart, name one of its own: `--service-account <name>@<project>.iam.gserviceaccount.com` (GCP) or `--app-name <name>` (Azure). AWS gives every connection its own role.
- **It can do what the grant says, nothing more.** The setup creates the connection's own role with the permissions the catalog lists for the kinds granted, plus any predefined roles a custom role cannot carry -- the plan's `granted` names them; read each one to the person, never summarize them away. The default grant never includes Owner or Editor, and never permissions above the target (an organization, a folder, a billing account); power over who has access in the account (IAM policies, roles, users, keys) and the tier above the target are separate opt-ins, `--access-granting` and `--organization-access`, each off unless the person asks. When a predefined role in `granted` touches access, say so plainly.
- **What else the setup creates or turns on, and what stays.** The plan's `changes` lists every step in the server's own words; read them to the person. GCP: turns on the APIs Planton itself operates on in the project (the trust's, Cloud Resource Manager, Service Usage), and creates or adopts the workload identity pool `planton` and the service account `planton-provisioner`; every kind turns on the APIs it needs at its first deploy. AWS: registers the account's identity provider for Planton's issuer. Azure: creates or adopts the app registration `planton-provisioner`. These are shared by every Planton connection in the account, so the cleanup leaves them in place; the plan's `keeps` says so.
- **You stay in control, in your own cloud.** The role is theirs: they can read it, narrow it or delete it at any time in their cloud's console. Verify always says what the connection can actually do after a change.
- **Taking it back is one script and one delete** (below). Deleting the trust or the role refuses every new token at once; credentials already issued expire on their own, in about an hour.
- **The script is readable before it runs.** `planton connect gcp --project <project> --print-documents` prints it exactly as the command would run it with the same name and grant, every permission written out.

## The common worries, answered truthfully

| The person asks | The answer, and where they see it |
|---|---|
| "What exactly can Planton do in my account?" | Deploy, change and destroy the kinds the grant covers, with the permissions the catalog lists for them and the predefined roles the plan's `granted` names; `grantedTo` is the identity that holds them. The plan's `grant` and `kindsGranted` say which kinds; `--print-documents` shows every permission. |
| "Does Planton keep my credentials?" | No. Keyless stores nothing; each deploy gets short-lived credentials (about an hour). Their own sign-in never leaves the machine. (DigitalOcean and Cloudflare are the exception: their token is stored as an organization secret, because those clouds have no keyless trust.) |
| "Can Planton read my data?" | Answer from the documents, never with a blanket no: the grant includes what the kinds need to manage their resources, which for some kinds includes reading that resource's contents (a secret store's versions, for example). Show the permissions with `--print-documents`, name any that read data, and offer to grant only the kinds they need. |
| "How do I limit it?" | Grant only some kinds (`--kinds GcpGkeCluster,GcpGcsBucket`), or narrow the role in their cloud later; Verify then says which kinds can no longer deploy. A recorded grant is changed with Update the Grant in the connection's Permissions section in the console (below). |
| "How do I revoke it?" | The cleanup script, then the delete (below). |
| "Does this touch my other roles or connections?" | The plan's `keeps` lists what stays untouched; `reachesOthers` lists what the change does to other connections that sign in as the same identity -- read it to them before asking. |
| "Why does it need this permission?" | Because a kind in the grant needs it to deploy: the catalog's permission files list each kind's permissions by purpose, and the connection's Permissions section in the console shows which kind needs what. |
| "What does Planton itself use it for?" | Beyond deploying kinds: the pickers that list what exists in the account, Verify, and -- when the organization uses them -- a state backend, a state encryption key, a secret backend, a container registry and Kubernetes access. "Everything" includes them; a grant of picked kinds keeps the pickers and Verify. |
| "Who in my organization can use this connection?" | By default it is authorized for every environment of the organization, so the people who can deploy into those environments can deploy through it; it becomes the organization's default for that cloud when none exists (never replacing one). `--no-authorize` and `--no-default` leave both alone; then `planton connection auth create --provider gcp --connection <slug> --scope environment --environments <env>` authorizes the environments the person names. |

## The grant

- **Default: everything Planton can deploy** for that cloud -- every catalog kind that needs neither opt-in tier, and every way Planton itself uses a connection.
- **`--kinds`** narrows it to the kinds named, as the catalog spells them (`GcpGkeCluster`, `AwsVpc`). Name every kind the person's architecture needs -- a GKE cluster usually also wants its node pool, network and subnetwork kinds; `catalog.kind-grounding.md` finds them. A grant of picked kinds keeps the pickers and Verify; the other uses (state backend, state key, secret backend, registry, Kubernetes access) come with "everything".
- **`--organization-access`** adds the permissions above the target that organization-level kinds need. **`--access-granting`** adds power over who has access in the account. Each is the person's explicit choice; say what it gives before passing it. The kinds that need a tier the grant leaves out are outside the grant.
- **Changing a recorded grant** is Update the Grant in the connection's Permissions section: on GCP and Azure one consent applies it, everywhere the page also gives the update script to run with the person's own sign-in. Planton needs no power over access for this -- the person's consent or sign-in applies it. Running `planton connect` again keeps the recorded grant and refuses a different one.

## Reading the verdict

`-o json --yes` ends with `verdict`; a person sees the same as a headline and lists.

- `headline`: "171 of 186 GCP Kinds Can Deploy" -- 186 is the kinds the grant covers, 171 of them hold what they need.
- `blocked`: kinds the grant covers that cannot deploy, each with its permission groups (`title` as the person reads it, e.g. "Manage GKE Cluster", the `missing` permissions and a `note`). Right after a setup, the cloud may still be spreading the new role: verify again in a minute. Still blocked, the role lacks what the grant includes (a newer catalog needs more, or someone narrowed it): apply the grant again with Update the Grant, never by building a second role by hand.
- `outsideGrant`: kinds the person did not grant, including those needing an opt-in tier. Not a fault -- what they chose.
- `canDeploy`, `usesNotReady` (Planton's own uses the identity cannot serve yet, each with its `missing` permissions), `catalogRelease`.
- `severity`: `success`, `warning` (a granted kind or one of Planton's own uses can't work yet -- the connection signed in, but don't report it as fully working), `info` (not judged) or `error` (no sign-in).
- `blockedFix` and `outsideGrantFix`: each a `sentence` and the `command` to run or the `url` to open -- the one way forward for that list. Kinds outside the grant are added on the connection's grant page; the sentence names the opt-in tier any of them needs.
- `healthy: false` with `reason` and `providerReason`: Planton's sentence and the cloud's own words, and `fix`: the cause as the console titles it (`refusal`, e.g. "GCP Refused This Connection's Identity"), its `door` and its `sentence` with the `command` or `url`. `door: "retry"` (severity `warning`) means nothing is wrong yet -- the cloud is still spreading the trust, or rate-limited Verify: wait a minute and run the `command`. `set-up-trust` runs the setup again, which re-creates the trust and adopts what exists; the other doors are changes the person makes. A refused sign-in exits 1, so a chained `verify && deploy` stops there.

`planton connect gcp verify <slug>` re-runs Verify any time.

## Taking it back

1. **Decide what happens to what it deployed.** Deleting a connection leaves everything it deployed running; destroy those first if the person wants them gone.
2. **The removal script, while the connection exists:** `planton connect gcp cleanup <slug>` prints what deleting the connection leaves in the cloud (to stderr) and the script that removes what the setup made for this connection (to stdout): on GCP its OIDC provider, its subject's binding on the service account and its own role with its bindings; on AWS its role and policies; on Azure its federated credential, its role's assignment and the role. The shared pool, service account, app, identity provider and the APIs stay, because other connections may use them. It is bash too: without bash, run it in the cloud's own shell. Run it on the person's yes with their sign-in, or hand it to them; it is safe to run again, and safe on a setup that made only part of what it would. When the removal reaches other connections, the output ends with `planton connect <cloud> verify <slug>` for each one: run them after the script, so the person sees what those connections can still deploy.
3. **Then delete the connection:** `planton delete GcpProviderConnection <slug>` (`AwsProviderConnection`, `AzureProviderConnection`, `DigitalOceanProviderConnection`, `CloudflareProviderConnection`). Planton refuses while a state or secret backend signs in through the connection, and names it; it does not check what the connection deployed, which is why step 1 comes first. It is a mutation like any other: the person's yes first.

A setup that stopped before the connection existed (the script failed or was interrupted, or the trust check refused) prints what it may have left and the command that prints its removal script: the same `planton connect` with `--print-cleanup` (in JSON, `cleanup.remains` and `cleanup.print`). Offer both ways forward: run the same command again to finish it (the script adopts what it made), or print and run the removal script to take it back. Planton refuses `--print-cleanup` once a connection holds that name: take that connection back as above instead.

## When something stops

Every failure after the script ran names the command that resumes; running the same command again picks up where it stopped: the script adopts what it already made, and an existing keyless connection with the same trust goes straight to Verify.

| The refusal | What to do |
|---|---|
| Keyless Isn't Offered Here | This Planton cannot offer keyless (its reason is in the sentence). Offer `planton connect gcp detect` (Runner Identity) and say what differs: deploys run with this machine's sign-in, with that sign-in's full access. |
| A lapsed or missing sign-in (detection fails) | The sentence names the sign-in it tried. The person runs `gcloud auth login`, `aws sso login --profile <profile>` or `az login`, then the same command. |
| Run the Setup in Your Cloud's Shell (no bash) | The machine has no bash (Windows without Git Bash). See "Without bash" below. |
| Install gcloud / aws / az | The cloud's CLI is missing. Name it; installing it is the person's choice -- or use the cloud's own shell (below). |
| The Setup Stopped | The first sentence is Planton's plain reading of the cloud's refusal; the cloud's own line follows (`providerReason` in JSON). A permission refusal means the person's sign-in cannot create the trust: on GCP it must be able to create workload identity pools, service accounts and custom roles, bind roles on the project and turn on APIs (a project Owner can); on AWS create IAM identity providers, roles and policies; on Azure register an app in Entra ID and create and assign roles in the subscription (Owner or User Access Administrator). Nothing was created in Planton. |
| The script says the AWS CLI is signed in to another account | The profile belongs to a different account than the target. Pick the right one with `--profile`. |
| The script says the Azure CLI works in another tenant | Switching is the person's choice: `az account set --subscription <id>`, then run again. |
| The Trust Isn't Working | The cloud refused the connection's token; the sentence and the cloud's reason say why. The trust is set up; after fixing, the same command finishes. |
| Choose Another Name | The name belongs to a different connection. Pass `--name`. |
| Change the Grant From the Connection | A re-run keeps the recorded grant. Use Update the Grant in the connection's Permissions section. |
| This Setup Changes Other Connections | `--yes` without `--accept-shared-changes` while the change reaches other connections. Tell the person which; the refusal's `next` gives this connection its own identity (fill in the name with them), and accepting the shared change is `--accept-shared-changes` -- only on their yes to that consequence by name. |

## Without bash: the cloud's own shell

`planton connect gcp --project <project> --name <name> --print-documents` prints the script; the person runs it in Google Cloud Shell, AWS CloudShell or Azure Cloud Shell, which have bash and are already signed in. The script ends by printing what Planton needs. The connection is then created from those values in the console's connection wizard (the setup step takes them), or with `planton apply -f` on a manifest naming them -- for GCP:

```yaml
apiVersion: connect.planton.ai/v1alpha1
kind: GcpProviderConnection
metadata:
  name: <name>
  slug: <slug>      # exactly the slug the script was printed for: the trust pins it
  org: <org>
spec:
  authMode: oidc
  workloadIdentityProvider: //iam.googleapis.com/projects/<number>/locations/global/workloadIdentityPools/planton/providers/<provider>
  serviceAccountEmail: planton-provisioner@<project>.iam.gserviceaccount.com
  projectId:
    value: <project>
  grant:
    coverage: everything   # or: coverage: chosen, with kinds: [GcpGkeCluster, GcpGcsBucket] -- the grant the script was printed for
```

Then `planton connect gcp verify <slug>`. `--print-documents` takes the same `--kinds`, `--name` and identity flags as the run, so the script matches the manifest's grant.

## DigitalOcean and Cloudflare: a token

1. `planton connect digitalocean --print-documents` (or `cloudflare`) prints the scopes the token needs and where to create it.
2. The person creates the token and runs the connect command themselves, with the token in their own shell -- `DIGITALOCEAN_TOKEN=<token> planton connect digitalocean --region <region>` (the region is required: where resources deploy when they name none) or `CLOUDFLARE_API_TOKEN=<token> planton connect cloudflare`; in PowerShell, set `$env:DIGITALOCEAN_TOKEN` first -- so the token is never pasted into the conversation and never a command-line flag you write. The command can also read it from stdin. `--no-authorize` and `--no-default` work as for the keyless clouds.
3. It stores the token as an organization secret (the one Planton Desktop uses too), connects, makes the connection ready and verifies it; `-o json` with `--yes` prints the receipt and the verdict.

The connection is named `digitalocean` (or `cloudflare`), the same one Planton Desktop makes, so an organization has one of each from these commands; connecting again with a new token rotates it. Revoking the token on the provider's token page cuts Planton off; `planton delete DigitalOceanProviderConnection digitalocean` removes the connection, and the stored secret (`do-api-token`, `cf-api-token`) stays until deleted with `planton secret delete do-api-token`.

Without the planton CLI (the platform-tools arm), connecting a cloud happens in the console's connection wizard; the CLI is the only agent path.

## What never to do

- Never pass `--yes` or `--accept-shared-changes` without the person's yes to that exact plan.
- Never edit the server's script, and never compose a provider name, an ARN or a subject yourself: the script reports them.
- Never put a token in a command-line flag, the conversation or a file in the repository.
- Never tell the person Planton "can't read anything" or "only has read access": answer from the permissions the documents list.
- Never switch the person's default subscription, project or profile for them; name the command and let them choose.
