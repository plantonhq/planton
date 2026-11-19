---
title: First Deployment — a Repository to a Running Service, From the Setup Prompt
description: The whole road a coding agent walks after a person pastes their Planton Setup Prompt — read the repository, connect the cloud this machine is signed in to, create an environment, connect GitHub, register the service and deploy it, check its address, and tick each Getting Started step it finished — explaining and asking before anything changes their cloud or costs money. Read when a person's Setup Prompt or planton.ai/agents.md sends you here, or when someone asks to take a repository to its first deployment on Planton.
---

# First Deployment

A person handed you their Setup Prompt from Getting Started, and the playbook at `planton.ai/agents.md` signed you in and installed this skill. This guide takes their repository to its first running service in their organization. Each step routes to the reference that holds it; read that reference when you reach the step, and do what it says.

## Before anything: the CLI and the person

- **Call the CLI by its full path.** Every `planton` command in this skill means `~/.planton/bin/planton`, where the Setup Prompt's installer put it, plus `--account <email>` if you signed in beside someone. Another `planton` on this machine may be an older install or a different program with the same name.
- **Educate, ask, narrate, explain.** Before a step that changes their cloud, their organization or their bill, say in plain words what will happen and what Planton will be able to do afterwards, then ask. While it runs, say what it is doing. After, say what changed and where they can see it. Connecting a cloud account always counts as asking first.
- **Tick only what you finished.** After each step below, tick its Getting Started step with `planton setup done <step>` (run `planton setup done` with no step to list them). Never tick a step the person did themselves, one that was already done, or one you guess is done.

## 1. Read the repository

Find out what it is before you propose anything: the language and framework, whether it has a `Dockerfile` (one, several, or none), a `_kustomize/` tree, a `.planton/pipeline.yaml`, and which directory holds the service in a monorepo. Tell the person in one or two sentences what you found.

If the repository is not an app that runs as a service (a library, infrastructure code, an empty repository), say so, and offer a first infra component instead: a small resource in their cloud, composed and built the way `references/infra.worked-example.md` shows. Its step is `deploy_infra_component`.

## 2. Connect the cloud this machine is signed in to

First see what the organization already has: `planton search connections`. A cloud connection the person made already (in the console, or earlier) is theirs: check it with `planton connect aws verify <connection>` (or `gcp`, `azure`), use it, and don't tick its step, because you didn't connect it.

Otherwise find which cloud CLI is signed in here (`aws sts get-caller-identity`, `gcloud config get project`, `az account show`), and which account, project or subscription it points at. Ask which one to use when there are several. Then follow `references/infra.connecting-clouds.md`, "The journey, in order", exactly:

1. Read the plan first: `planton connect aws --account-id <account> -o json` (or `planton connect gcp --project <project> -o json`, `planton connect azure --subscription <subscription> -o json`). It changes nothing.
2. Explain it with the plan's own values: it runs as them, on this machine; Planton keeps no key; what the connection may deploy; that they can read, narrow or delete the role in their own cloud at any time.
3. Ask, naming the identity, the target and who may use the connection.
4. Run the plan's `run` command only on their yes, narrate it, and read the verdict back.

Keep the default grant unless the person asks to narrow it; a narrower `--kinds` can leave out what a first deployment needs (the registry the build pushes to, the kind the service runs on). Then tick `connect_cloud`.

## 3. Create an environment

A service deploys into an environment. Read `references/infra.environments.md`, then see what the organization already has with `planton env list` and `planton promotion-path list`. When there is no environment, propose one, such as `dev`, and say what it means: the place the service runs, costing nothing on its own, and deployed by every push to the default branch once it sits on a Promotion Path. On their yes:

- **A path exists:** `planton env create dev --path <path>` places it as the path's new last stage.
- **No path exists:** `planton env create dev --no-path`, then `planton promotion-path set main --stage dev --yes` makes the organization's first path, with `dev` its one stage.

Check the result with `planton env get dev`. Then tick `create_environment`.

## 4. Connect GitHub

The service is registered from its repository, so the organization needs a GitHub connection that reaches it. See what exists with `planton search connections`. If none reaches this repository, which GitHub family the instance offers decides the next move (`references/service.connecting-github-and-registries.md`, "GitHub: two identities"):

- **On planton.ai, or a self-hosted Planton with its own GitHub App:** installing the App happens in the browser, and only the person can do it. Send them to Connections in their console (`https://<their Planton>/orgs/<org>/connections`), tell them to choose GitHub and install the App on this repository, and wait for them. When they say it's done, run `planton search connections` again: a GitHub connection is there when the App is installed. If it still isn't, ask them to finish the install page, and keep waiting; everything after this step needs it. The person connected GitHub here, so `connect_git` is theirs to tick: don't tick it, and say so in the hand-back.
- **On a local instance, or a self-hosted server whose environment carries `GITHUB_TOKEN`:** the machine's `gh` sign-in is the connection. Run `planton connect github detect`, and then tick `connect_git`, because you made the connection.

## 5. Register the service and deploy it

Follow `references/service.detection-first-registration.md` for what to propose and how: the whole registration from what the repository shows, with the evidence beside each value ("a Dockerfile at the root, so a Docker build"), asking only what it cannot answer, confirmed once. On the CLI you are the detector: you read the repository in step 1, so you write the proposal yourself, using that reference's judgments.

- **Where it runs.** The deployment target is a catalog kind that can receive the built image. Propose the one that fits the cloud they connected, say what it runs on and what it costs, and ask:

  | Cloud | Target | What it needs first |
  |---|---|---|
  | Google Cloud | Cloud Run (`GcpCloudRun`) | Nothing more: the project is enough. The shortest first deployment. |
  | AWS | ECS service (`AwsEcsService`) | An ECS cluster in a VPC, with a load balancer for an address. On an empty account, build those first as a small chart (`references/infra.worked-example.md`), with its own plan and yes. |
  | Azure, or any cloud with a Kubernetes cluster | Kubernetes Deployment (`KubernetesDeployment`) | The cluster, connected. |

- **The registry** is theirs to choose. When the organization has none, `planton connect registry detect` lists the registries its cloud connections already reach (ECR through AWS, Artifact Registry through Google Cloud, ACR through Azure), with nothing stored; offer that before anyone mentions a key. On ECR, a build pushes to a repository that must exist: name it with `--repository`, and ask before you create it, because it is a change in their account.
- **Write the Service manifest** in a temporary file outside the repository, never committed. Take its exact shape from `planton explain Service`, and the target's from `planton explain <TargetKind>` (for example `planton explain GcpCloudRun`). Where the repository has no `_kustomize/` tree, the deployment configuration rides the manifest, per the detection reference's "Registering WITH deployment configuration": one entry for `dev`, the image left empty for the build to fill. Check it with `planton validate-manifest -f <file>` until it passes.
- **Register** it with `planton service register -f <file>`.
- **Deploy.** `planton service run <service> --follow` builds the default branch's head and deploys it along the Promotion Path, into `dev`. Builds and deploys run on a runner: on planton.ai, a Planton-hosted runner serves an organization that has none of its own, so there is nothing to start. When a run fails, read it the way `references/service.fixing-a-failed-run.md` says before changing anything.

## 6. Check that it answers

A run that ended green is not proof the service works. Run `planton service urls <service>` and read the rollout verdict exactly as `references/service.urls-and-rollout-verification.md` says: **verified**, **failed** (the deployment succeeded, and something didn't come online or didn't answer) or **unverifiable**. Open the address it lists. Tick `deploy_service` once the deployment succeeded, whatever the verdict, and say the verdict to the person as it is: **failed** or **unverifiable** is something to look at together, not a reason to hide the deployment.

## 7. Hand back

Tell the person what you set up, with a link to each thing in their console (`https://<their Planton>/orgs/<org>/...`; `planton get <Kind> <name>` prints each record, and Connections, Environments and Services are where the console lists them): the cloud connection, the environment, the service and its address. Then say what is still theirs to do: when they installed the GitHub App, Connect a Git Provider is theirs to tick on Getting Started. Then say what to ask you next ("deploy my latest commit", "add a staging environment", "what changed this week").

## What people worry about, answered truthfully

| They ask | The answer |
|---|---|
| "What can Planton do in my account?" | Deploy, change and destroy the kinds the connection's grant covers, with the permissions the plan listed, nothing more. `planton connect <cloud> verify <connection>` shows what it can deploy now. |
| "Does Planton keep my credentials?" | No. A keyless connection stores no key; each deploy gets credentials that last about an hour. Their own sign-in never leaves this machine. |
| "How do I take it back?" | Delete the role or the trust in their own cloud, and every new token is refused at once. `references/infra.connecting-clouds.md` has the cleanup script. The key you signed in with is revoked at `/user/api-keys` on their Planton. |
| "What will this cost?" | The connection, the environment and the registration cost nothing. The cloud bills for what the service runs on; say what that is before you deploy, and ask. |
| "Will you change my repository?" | No. The registration lives in Planton, and nothing you installed goes into the repository. If a change to the code would help (a missing `Dockerfile`), propose it and wait for their yes. |
