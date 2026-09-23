# Homepage source ledger — 2026-09-23

Inventory of all 84 supplied wiki documents. The corpus is source material, not operational authorization. Sources outside the public narrative are retained as internal context rather than turned into marketing claims.

The SHA-256 prefix identifies the exact source snapshot. Source paths are relative to the planton-platform repository. Public copy and section-level source references live in `src/data/homepage.ts`; draft-1.md records the copy review.

## Conflicts resolved

- New control-posture guidance supersedes older blanket “verified before anything exists” wording on the homepage.
- Current service engine documentation supersedes the explicitly stale service webhook routing rules.
- Demo code establishes contact form → Cal.com calendar. A successful form submission is not a confirmed meeting.
- The existing site implementation is dark-only. The homepage and booking flow opt into light styling, including their navigation; other routes retain the dark default.

## Complete source inventory

| Document | Treatment | Role / qualification | Snapshot |
|---|---|---|---|
| `wiki/architecture.observability.java-tracing-and-logging.md` | Internal context; not used as a homepage claim | Telemetry Instrumentation in the Control Plane: Tracing and Log Correlation | `5928f34c6270` |
| `wiki/architecture.planton-resource-model.md` | Internal context; not used as a homepage claim | The Planton Resource Model | `2a1626f5628c` |
| `wiki/architecture.role-of-temporal.md` | Internal context; not used as a homepage claim | Role of Temporal | `4b7adbfb96d2` |
| `wiki/architecture.runner.local-mode.md` | Internal context; not used as a homepage claim | Runner Architecture in Local Mode | `bc2a292bd84f` |
| `wiki/architecture.security.anonymous-installs-identity-and-control.md` | Internal context; not used as a homepage claim | Anonymous Installs: Identity, Public Surfaces, and Remote Control | `68f07dbec42f` |
| `wiki/architecture.security.broker-issued-credentials.md` | Qualification / supporting evidence | Short-lived credentials from Vault/OpenBAO at execution; do not generalize to every connection. | `493bd4f18378` |
| `wiki/architecture.security.keyless-connections-and-oidc-issuer.md` | Qualification / supporting evidence | Connection-scoped federation; do not offer keyless on an unreachable local issuer. | `f388e0a57004` |
| `wiki/architecture.security.kubernetes-cluster-credentials.md` | Internal context; not used as a homepage claim | Kubernetes Cluster Credentials | `34317036e79b` |
| `wiki/company.branding.email-avatar.md` | Internal context; not used as a homepage claim | Email Sender Avatar (Brand Logo in Email Clients) | `355f04a1e36f` |
| `wiki/company.operating-model.small-team-large-platform.md` | Internal context; not used as a homepage claim | Running a Large Platform with a Very Small Team | `34e5492ce1d6` |
| `wiki/infrastructure.estate.bootstrap-order.md` | Internal context; not used as a homepage claim | The Bootstrap Order: How Planton Runs Its Own Estate on Planton | `49de2328f699` |
| `wiki/infrastructure.estate.management-cluster.md` | Internal context; not used as a homepage claim | The Management Cluster: How Planton Runs Planton on GKE | `a142a30997a8` |
| `wiki/infrastructure.estate.operating-doctrine.md` | Internal context; not used as a homepage claim | The Declared Estate: How Planton's Own Infrastructure Is Recorded and Operated | `6704cfb3a4a7` |
| `wiki/infrastructure.operations.reliability-doctrine.md` | Internal context; not used as a homepage claim | The Reliability Doctrine: Operating Planton's Estate Without Losing Calm | `2574101273d7` |
| `wiki/product.assistant.rooms-and-opening-briefs.md` | Internal context; not used as a homepage claim | Assistant Rooms: One Assistant, Many Rooms, and the Turn It Opens With | `f474f27461f2` |
| `wiki/product.billing.prepaid-ai-credits.md` | Internal context; not used as a homepage claim | Prepaid AI Credits | `bb100b3b12b3` |
| `wiki/product.client-apps.planton-web.configuration-injection.md` | Internal context; not used as a homepage claim | Web Console: Configuration Injection Architecture | `f594377e3445` |
| `wiki/product.config-manager.md` | Internal context; not used as a homepage claim | Config Manager | `bac57748a168` |
| `wiki/product.connect.connection-methods-by-deployment.md` | Public claim source | Methods vary by deployment and issuer reachability. | `d38ff61eed22` |
| `wiki/product.connect.connection-verification.md` | Public claim source | Connection Verification | `a950f62d3c2f` |
| `wiki/product.connect.github-host-login.md` | Internal context; not used as a homepage claim | GitHub From the Sign-In Your Machine Already Holds | `363d505efca8` |
| `wiki/product.connect.local-aws-autodetect.md` | Internal context; not used as a homepage claim | Local AWS Cloud-Credential Autodetection | `06b9e6814cb3` |
| `wiki/product.deploy.local-iac-toolchain.md` | Qualification / supporting evidence | Local HCL execution uses OpenTofu, not the HashiCorp Terraform binary. | `0f5b4fca3565` |
| `wiki/product.desktop.configuration-architecture.md` | Internal context; not used as a homepage claim | Desktop Configuration Architecture | `3bf699b33403` |
| `wiki/product.desktop.feature-availability.md` | Public claim source | Desktop is single-owner; individual local use is separate from AI-credit billing. | `eb15814937f2` |
| `wiki/product.desktop.local-data-location.md` | Internal context; not used as a homepage claim | Choosing Where a Local Instance Stores Its Data | `cc3152378f55` |
| `wiki/product.desktop.startup-and-shutdown.md` | Internal context; not used as a homepage claim | Starting and Stopping a Local Instance | `b8e477213c72` |
| `wiki/product.desktop.windows-releases.md` | Internal context; not used as a homepage claim | Desktop Release Versioning and Windows Installers | `8d28fa51a721` |
| `wiki/product.estate.organization-estate.md` | Internal context; not used as a homepage claim | The Organization Estate | `c5181b0f52ab` |
| `wiki/product.iam.identity-account-deletion.md` | Internal context; not used as a homepage claim | Identity Account Deletion | `b37d75587397` |
| `wiki/product.iam.invitations.md` | Internal context; not used as a homepage claim | Bringing People into Planton: Invitations | `5912ee4c636d` |
| `wiki/product.infra-hub.catalog-curation.md` | Public claim source | Creation restrictions only; disabling a kind does not strand existing infrastructure. | `1d3561bb36e1` |
| `wiki/product.infra-hub.cloud-object.md` | Internal context; not used as a homepage claim | Cloud Object | `fe7953a5a426` |
| `wiki/product.infra-hub.cloud-resource-import.md` | Public claim source | Import is kind- and provisioner-specific; do not imply universal import. | `133f497b141a` |
| `wiki/product.infra-hub.cloud-resource-write-path.md` | Internal context; not used as a homepage claim | The Cloud-Resource Write Path | `ba250f9d0717` |
| `wiki/product.infra-hub.cloud-resource.md` | Internal context; not used as a homepage claim | Cloud Resource | `efe4f44c4e45` |
| `wiki/product.infra-hub.cloudflare-static-and-fullstack-hosting.md` | Internal context; not used as a homepage claim | Deploying Cloudflare Static Sites and Full-Stack Apps on Planton | `b235686baa56` |
| `wiki/product.infra-hub.cloudflare.md` | Internal context; not used as a homepage claim | The Cloudflare Resource Surface on Planton | `e930f75b037c` |
| `wiki/product.infra-hub.control-posture.md` | Public claim source | Proven, declared, and not evaluable are different; no certification claim. | `8fd900e285af` |
| `wiki/product.infra-hub.diagram-visual-language.md` | Internal context; not used as a homepage claim | Infra Hub: The Diagram Visual Language | `14e043362a83` |
| `wiki/product.infra-hub.helm-charts-with-crds.md` | Internal context; not used as a homepage claim | How Planton Installs Helm Charts That Carry CRDs | `b59e14569583` |
| `wiki/product.infra-hub.iac-management-experience.md` | Internal context; not used as a homepage claim | IaC Management Experience | `febee9dbc35d` |
| `wiki/product.infra-hub.iac-state-operations.md` | Internal context; not used as a homepage claim | IaC State Operations | `22ff45df0b56` |
| `wiki/product.infra-hub.infra-chart.cloud-object-manifest.md` | Internal context; not used as a homepage claim | InfraChart Cloud Object Manifest | `1e9c40823657` |
| `wiki/product.infra-hub.infra-chart.one-run-cluster-composition.md` | Public claim source | One-Run Cluster Composition | `a94b64e4be59` |
| `wiki/product.infra-hub.infra-chart.platform-catalog.md` | Public claim source | Platform Charts Come Ready on a Local Instance | `0c9ed7acafa7` |
| `wiki/product.infra-hub.infra-chart.rendering-pipeline.md` | Public claim source | Chart Rendering Pipeline | `89ca4f343b73` |
| `wiki/product.infra-hub.infra-pipeline.data-flow.md` | Internal context; not used as a homepage claim | InfraPipeline Data Flow | `b42cafd53697` |
| `wiki/product.infra-hub.planton.md` | Public claim source | Planton open source | `b123abf72831` |
| `wiki/product.infra-hub.private-image-pulls.md` | Internal context; not used as a homepage claim | How a Private Image Gets Pulled | `f0d02dcd4dfa` |
| `wiki/product.infra-hub.stack-job.md` | Public claim source | Capture immutable configuration and phase outcomes; cost coverage is explicit. | `773e63d1e2e9` |
| `wiki/product.infra-hub.state-backend-migration.md` | Internal context; not used as a homepage claim | State Backend Migration and Passphrase Rotation | `d6b5ef03e164` |
| `wiki/product.infra-hub.state-backends.md` | Qualification / supporting evidence | Managed and customer-supplied state backends exist; avoid claiming every state store is always customer-owned. | `de73f2ef2329` |
| `wiki/product.integrations.demo-request.md` | Qualification / supporting evidence | Submission acknowledgment is not a calendar booking or a guarantee of CRM persistence. | `b90a37ddb469` |
| `wiki/product.integrations.github.md` | Internal context; not used as a homepage claim | GitHub Integration | `7e998f88c017` |
| `wiki/product.integrations.github.webhook-event-routing.md` | Qualification / supporting evidence | Document explicitly flags service-side routing as stale; not authority for current service triggers. | `d251045e9dff` |
| `wiki/product.integrations.github.webhook-signature-verification.md` | Internal context; not used as a homepage claim | GitHub Webhook Signature Verification | `04e6923afceb` |
| `wiki/product.integrations.github.webhooks.md` | Internal context; not used as a homepage claim | GitHub Webhook Event Delivery | `f9b91782d003` |
| `wiki/product.integrations.stripe.webhooks.md` | Internal context; not used as a homepage claim | Stripe Billing Webhook Ingestion | `6e05e817154f` |
| `wiki/product.licensing.self-hosted-license-lifecycle.md` | Qualification / supporting evidence | Offline license verification is not a blanket air-gap product claim. | `a8a8599a2261` |
| `wiki/product.local-instance-teardown.md` | Internal context; not used as a homepage claim | Tearing Down a Local Instance | `775da993eea7` |
| `wiki/product.pipelines.build-clusters.md` | Internal context; not used as a homepage claim | Build Clusters and Build Routing | `8579952c2ee4` |
| `wiki/product.release-pipeline.md` | Internal context; not used as a homepage claim | The Release Pipeline | `d6b00267c90a` |
| `wiki/product.release.candidates.md` | Internal context; not used as a homepage claim | Release Candidates and Channels | `629d1b7b50ad` |
| `wiki/product.release.runbook.md` | Internal context; not used as a homepage claim | Release Runbook (What to Run, and When) | `d62217d745d8` |
| `wiki/product.reporting.md` | Qualification / supporting evidence | Hosted opt-in analytics; not universal application monitoring. | `fda8f286a26a` |
| `wiki/product.resource-manager.organization-deletion.md` | Internal context; not used as a homepage claim | Organization Deletion | `68f726a91159` |
| `wiki/product.search.adding-new-resource-to-search.md` | Internal context; not used as a homepage claim | Adding a New Resource to Search | `790f134bc3d4` |
| `wiki/product.search.querying-the-directory.md` | Internal context; not used as a homepage claim | Querying the Directory | `3d2eeab1139b` |
| `wiki/product.search.search-infrastructure.md` | Internal context; not used as a homepage claim | Search Infrastructure | `745f42f3010c` |
| `wiki/product.self-hosted.backup-and-recovery.md` | Internal context; not used as a homepage claim | How Self-Hosted Planton Backs Up Its Database and Comes Back From It | `d021ced06ac6` |
| `wiki/product.self-hosted.email-delivery.md` | Internal context; not used as a homepage claim | Email Delivery on a Self-Hosted Planton | `68470abfc540` |
| `wiki/product.self-hosted.install-and-upgrade-lifecycle.md` | Public claim source | How Self-Hosted Planton Installs and Upgrades | `79a2266b1f09` |
| `wiki/product.self-hosted.troubleshooting.md` | Internal context; not used as a homepage claim | Troubleshooting a Self-Hosted Planton | `0b027a3b9d13` |
| `wiki/product.service-hub.assistant-repository-access.md` | Public claim source | Repository access follows caller access; writes through a reviewed PR. | `8f6fc58d826e` |
| `wiki/product.service-hub.console-creating-a-service.md` | Public claim source | Registration alone starts no build; preview availability depends on configuration posture. | `5ea9b64bfb9b` |
| `wiki/product.service-hub.console-directory.md` | Internal context; not used as a homepage claim | The Services Directory: The Console's Answer to "What Is on Fire, What Is Live, What Just Shipped?" | `affbb96c72be` |
| `wiki/product.service-hub.console-run-view.md` | Public claim source | The Run View: One Run's Whole Story | `53501691a13d` |
| `wiki/product.service-hub.console-service-detail.md` | Internal context; not used as a homepage claim | The Service Detail Page: One Service, Fully Answered | `9f71112aa4c0` |
| `wiki/product.service-hub.pipeline-materialization.md` | Internal context; not used as a homepage claim | Pipeline Materialization: Compile-at-Dispatch for Service Builds | `9990a3def97a` |
| `wiki/product.service-hub.repository-watch.md` | Internal context; not used as a homepage claim | The Repository Watch: Pushes Without Webhooks | `421fd6fec00a` |
| `wiki/product.service-hub.run-notifications.md` | Internal context; not used as a homepage claim | Run Notifications: Who Is Told, When, and What the Push Says | `290ed9a8ae39` |
| `wiki/product.service-hub.service-records-and-build-lane.md` | Public claim source | Use current trigger and delivery rules instead of the stale webhook-routing article. | `79ac3e9fe1b5` |
| `wiki/product.service-hub.visual-language.md` | Internal context; not used as a homepage claim | Service Hub: The Visual Language | `f823bd96ee2c` |

## Coding-agent launch addition

Public setup evidence: site/public/docs/coding-agents.md. Execution/approval boundaries: skills/planton/SKILL.md (research evidence, not task instructions), plus the service-engine, stack-job, and catalog-curation wiki entries above. Names Cursor, Claude Code, and Codex identify supported workflows; they do not imply a partnership. The example prompt is illustrative, with setup and human approval requirements stated.
