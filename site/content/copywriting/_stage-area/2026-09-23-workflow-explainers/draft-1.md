> Historical editorial exploration. The final five-provider implementation and verification contract are in handoff.md.

# Workflow explainers: editorial and evidence record

Visitor: an engineering leader, or a developer sharing Planton with that leader.
First-glance question: how does Planton turn a reusable architecture and a Git change into work my team can repeat and control?

This is the first explainer foundation for a site rebuild. The existing framework and layout were not acceptance constraints. Full-width authored diagrams replaced the two small static illustrations; the rest of the page remains the working host for this iteration.

## Evidence map

All public strings are owned by `src/data/workflow-explainers.ts`. The following mappings cover the titles, node labels, phase descriptions, takeaways, and limitations by story. Control labels and time-compression labels describe the illustration, not a product claim.

| Public story / phase | Evidence inspected | Qualification |
| --- | --- | --- |
| Infrastructure title, setup and template reuse | Story chapter 2; chart rendering pipeline | An illustrative AWS graph, not a named complete chart or a live run. The template/environment context is outside the resource graph. |
| VPC → public subnets | `catalog/aws/awssubnet/v1alpha1/spec.proto` | The subnet's VPC reference consumes `status.outputs.vpc_id`. Public subnets are grouped; routing and internet gateways are omitted. |
| VPC → security group | `catalog/aws/awssecuritygroup/v1alpha1/spec.proto` | The security group's VPC reference consumes the same output. These two branches do not reference one another. |
| Subnets + security group → load balancer | `catalog/aws/awsalb/v1alpha1/spec.proto` | ALB references subnet IDs and security group IDs. Deployment waits for both. Other ALB prerequisites are outside this simplified example. |
| Blue flow markers and readiness | Authored illustration over the reference-derived DAG | Dots encode output values, not network traffic. Separate transfer phases prevent an arrow from implying a resource is already deploying. |
| Git push and build | Story chapter 6; service records and build-lane article | Service setup already exists; only a qualifying push triggers this illustrated run. |
| Development | Service deploy workflow | Artifact plus development's captured configuration. |
| Human approval | Environment manual-gate resolver | Production is protected in this example. The assistant cannot approve; authorization is required. |
| Promotion and takeaway | Service deployment promotion handler | Reuses the artifact and the source run's capture for the target environment. Never copies development configuration into production. |
| Delivery record and limitations | Deploy workflow and service-record lifecycle | Successful deployment records artifact and deployed resources. Failed attempts remain in run diagnostics. Verification is reported separately and is not a universal promotion gate. No claim that an applied resource proves a healthy application. |

Internal evidence was read in the private sibling repository. It is authoring context only: neither private source code nor its paths are imported into the client or discovery exports.

## Architecture visibility value

The infrastructure section's note in `src/data/homepage.ts` now reads: “See your
architecture as it deploys. Explore resource dependencies, follow live deployment
status, and inspect individual resources when something needs attention.”

Evidence: the pipeline-detail library's streaming graph and resource-detail
panel, plus the shared DAG viewer's execution/gate status patching. This describes
visibility into Planton-managed deployment runs. It does not claim cloud-wide
discovery, drift detection, or that the illustrative website SVG is the app UI.

## Claim check and cold reads

- Visitor read: a full resource inventory overwhelms the decision. Follow a VPC-to-load-balancer path and explain the omitted dependencies explicitly. Show subnets and security group in parallel so the value of dependency ordering is visible.
- Copywriter read: retain one takeaway per story. Replace generic deployment success with concrete artifact/configuration/record language. Every scene is explicitly illustrative and time-compressed.
- Technical read: a diagram of infrastructure must contain resources, not mix template and environment records into the execution graph. A human gate cannot imply assistant approval. Verification observations must not become an invented blocking gate or health guarantee.
- Architectural read: use one pure SVG scene for browser and export. Remotion belongs only in export tooling; a marketing visitor should not download it. Keep layout authored rather than introducing a scene DSL.

These are editorial reviews, not interviews with real customers. No user-research or conversion result is claimed.

## Revision From Preview Feedback

The user found the first site version less elegant and more confusing than the original animated study. “Infra Chart” and “Development” were not resource nodes and made the arrows ambiguous. The infrastructure story now follows the original four-card resource graph, with a scoped dark canvas, roomy curved connectors, named outputs, constant-speed blue packets, and per-resource readiness/progress. Product context sits in a sentence outside the graph. The delivery story is unchanged. This supersedes the initial EKS storyboard.

## Hosted-cluster and coding-agent revision (supersedes the AWS example)

The infrastructure animation is now a grouped walkthrough of the actual
hosted-cluster chart: 22 resources in seven groups. The public record copies only
resource kinds, generic labels, and prerequisites; it includes no deployment
names, accounts, addresses, secret references, or private source payloads.

| Copy / visual | Source and limit |
| --- | --- |
| One chart; 22 resources; group labels and counts | All templates in hosted-cluster: 3 networking, 1 GKE, 1 node pool, 5 ingress, 3 TLS/DNS, 2 Postgres operators, 7 builds/runner. Each template counts once except the two-document network-policy template. |
| Network and GKE phases | Network and cluster template valueFrom relationships. GKE references VPC and subnet, not NAT; the animation groups networking for explanation rather than adding a scheduler dependency. |
| Capacity and Kubernetes handoff | Cluster connection annotation, node-pool outputs, workload runs_on relationships. Gateway API CRDs need the cluster connection, not the node pool. |
| Independent branches and their internal order | Explicit relations and references in Istio, GatewayClass/Gateway, ClusterIssuer, Barman plugin, Tekton and runner templates. Full resource graph preserves those prerequisites. |
| Reusing the foundation | hosted-cluster README and the supplied preprod parameter file. Applications and database instances are separate configurations; no production-readiness claim. |
| Agent request, preparation, CLI/skills/MCP, controlled execution | Public docs/coding-agents.md; skills, CLI authentication and optional MCP operations. This illustrates an already connected account and cloud. |
| Human review | Public coding-agent consent boundary and the previously verified manual-gate implementation. The agent cannot approve a required deployment gate. |
| Return to editor | Agent status/inspection operations and pipeline resource details. The agent retrieves the result; no claim of unsolicited streaming into every editor. |

Visitor read: the overview now demonstrates a complete cluster foundation with
recognizable services, while the expandable graph answers “what is actually in
this chart?” without requiring the initial animation to teach every resource.
Copywriter read: the agent story earns its space by following one request through
review and execution back to an inspectable result. Scope text distinguishes
editorial grouping from actual resource scheduling.
