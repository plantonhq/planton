# Homepage draft and claim trace

Audience: engineering leaders. First question: what does Planton do, and is it worth a demo?

Approved direction: white-on-black, outcome-led, existing form and calendar.

## Hero

Set up your cloud. Let your team ship.

Planton brings infrastructure setup and application delivery into one self-service platform. Create reusable environments, deploy applications from Git, and manage changes—all in your own cloud.

Sources: overview, infrastructure and delivery sources below; positioning language from story chapter 2.

## One path from infrastructure to application.

Chapter: what-planton-is. Sources: `wiki/product.connect.connection-verification.md`, `wiki/product.infra-hub.infra-chart.rendering-pipeline.md`, `wiki/product.service-hub.console-creating-a-service.md`.

Your cloud account is the foundation. Planton connects the work of setting it up with the work of shipping software onto it. Platform engineers define the starting points and controls. Developers use them to get their applications running.


- Connect your cloud: Connect the cloud accounts your team already uses. Choose the connection method that fits your deployment and verify that it can obtain the access it needs.
- Create an environment: Compose cloud resources into reusable templates. Deploy the network, runtime, and supporting services in dependency order, with a configuration you can inspect.
- Ship your application: Connect your repository, configure where the service runs, and follow its journey from build to deployment. See what is live in each environment.

## Design it once. Reuse it across environments.

Chapter: what-planton-is. Sources: `wiki/product.infra-hub.infra-chart.rendering-pipeline.md`, `wiki/product.infra-hub.infra-chart.one-run-cluster-composition.md`, `wiki/product.infra-hub.infra-chart.platform-catalog.md`.

A working environment should become a starting point your team can use again. With Planton, a reusable infrastructure template—an Infra Chart—describes the resources, their configuration, and how they connect.

Use the assistant to help compose the architecture, or work directly with its manifests. Review the resulting configuration before you deploy. The template stays readable and editable as your requirements change.
Give development and production their own configuration without rebuilding the architecture from scratch. Planton resolves dependencies and deploys the resources in order, including the connections that let workloads reach a newly created cluster.
Platform engineers define the pattern. Developers reuse it.

## From a Git push to a running service.

Chapter: services-ship-from-git. Sources: `wiki/product.service-hub.console-creating-a-service.md`, `wiki/product.service-hub.service-records-and-build-lane.md`, `wiki/product.service-hub.console-run-view.md`, `wiki/product.service-hub.assistant-repository-access.md`.

Choose a repository. Planton reads it and proposes a service configuration, showing the evidence behind its choices. Confirm the setup, choose your environments, and give the next push a path to deployment.

Use a detected Dockerfile, a supported Buildpacks build, or a custom pipeline. Keep deployment configuration in your repository or configure it in Planton. Your team chooses the level of control it needs.
Follow builds and logs, promote an artifact through your environments, and pause for approval where required. Each delivery records what shipped. When something fails, open the run and ask the assistant for help with the actual configuration and repository context.
Know what shipped, where it landed, and what needs attention.

## Give developers freedom. Keep changes accountable.

Chapter: your-rules-hold. Sources: `wiki/product.infra-hub.catalog-curation.md`, `wiki/product.infra-hub.control-posture.md`, `wiki/product.infra-hub.stack-job.md`, `wiki/product.service-hub.assistant-repository-access.md`.

Self-service works when the boundaries are clear. Define which cloud components your organization can create, require approval for protected environments, and keep a record of the changes that run.

The assistant works within the access of the person asking. Repository changes go through a reviewed pull request.
- A catalog shaped by your team: Make approved component kinds available for new resources. The platform enforces creation restrictions at the API boundary, including requests from the CLI and agents.
- Evidence with its limits visible: Review available cost, permission, and technical-control information. Proven, declared, and not-evaluable claims remain distinct; estimates are not your cloud bill.
- A record you can return to: Inspect the captured configuration, execution outcome, and approval history. Later edits do not rewrite what an earlier deployment recorded.

## Fits the cloud— and the way—you work.

Chapter: runs-where-you-decide. Sources: `wiki/product.infra-hub.cloud-resource-import.md`, `wiki/product.infra-hub.planton.md`, `wiki/product.desktop.feature-availability.md`, `wiki/product.self-hosted.install-and-upgrade-lifecycle.md`, `wiki/product.connect.connection-methods-by-deployment.md`.

You can begin with an existing cloud account. Supported resources can be imported into infrastructure state without recreating them. Check the supported kind and import path before bringing a resource under management.

Choose hosted Planton for a shared platform, self-host it on your Kubernetes cluster, or use Desktop for an individual workspace. Connection methods and capabilities vary by deployment; the underlying resource model stays consistent.
The infrastructure modules are open source. Your manifests remain readable, and you can deploy them with the standalone open-source CLI. Your adoption path can start with one environment or service.


## Customer proof

Verbatim: `src/data/testimonials.ts`, Sai Saketh and Rohit Reddy Gopu. No quote is tightened or paraphrased.

## Claim check and cold reads

- Removed the older hero’s blanket claim that all costs, permissions, and controls are verified. Control-posture documentation distinguishes proven, declared, and not evaluable.
- Catalog restrictions govern creation, not all operations on existing resources.
- Registration does not start a build; the next qualifying push does.
- Supported build paths include Dockerfiles, Buildpacks, and custom pipelines. No universal zero-configuration promise.
- Hosted, self-hosted, and Desktop have different connection capabilities; Desktop is single-owner.
- No guaranteed savings, certifications, live illustrative prices, or automatic approvals.
- Visitor read: define the jobs before introducing Infra Hub and Service Hub.
- Copywriter read: one primary conversion, two customer accounts, direct demo expectations.

## FAQ trace

Integration → service-records-and-build-lane; import → cloud-resource-import; deployment → feature-availability/install-and-upgrade; credentials → connection-methods and broker-issued-credentials; AI → assistant-repository-access and rooms; demo → existing BookDemoForm and BookDemoScheduler.

## Approved launch revision

Light-only homepage and demo journey. Hero supporting paragraph and the new agent workflow read directly from HOMEPAGE.intro and HOMEPAGE.agents in src/data/homepage.ts. The agent section explains ask → prepare → review → deploy, sourced to public/docs/coding-agents.md, skills/planton/SKILL.md, and the wiki references on the record. Do not describe approvals as optional, claim setup-free use, or imply the agent can self-approve a gate.
