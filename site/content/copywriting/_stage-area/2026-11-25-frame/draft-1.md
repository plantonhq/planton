# The frame, draft 1 (2026-11-25)

The words of the Security page, the header, the footer and the 404. Every sentence sits beside its source: a claims-ledger row (`company/sales/_ledger/platform-claims.2026-10-30.md` in planton-platform, "Security answers" and "Governance") or a story chapter (`src/data/story.ts`).

## Security (`/security`)

**Title:** Security

**Lede:** Planton works in your cloud through one role you create, and in your GitHub through one App you install. Delete either, and Planton is out.
- ledger: "No stored cloud keys…", "They cut it off in their own cloud, any time"; chapter 8 (runs where you decide)

### 1. What Planton can touch

Only what the infrastructure you chose needs. When you connect a cloud, you create one role in your own account: an IAM role on AWS, a service account on Google Cloud, an app registration on Azure. Its permissions come from the published permission list of each kind you use, and it is never Owner, Editor or Contributor.
- ledger: "The role can do only what the kinds they choose need"; `wiki/architecture.security.keyless-connections-and-oidc-issuer.md`

Your coding agent never holds cloud credentials. It writes a manifest and asks Planton to deploy it. Before anything is created, you see the plan: what it will cost each month and the exact permissions it will use.
- ledger: "Every infrastructure change is a plan…", "Before anything exists: a bill preview, the runner's exact permission policy"; chapter 3 (verified before it exists)

**Artifact:** one statement from the S3 bucket kind's permission list (`catalog/aws/awss3bucket/iac/permissions.yaml`, `ManageBucket`).
**Check it yourself:** every kind's permission list is in the open-source catalog.

### 2. No stored keys

Planton stores no cloud keys. Your cloud trusts one issuer, `https://planton.ai`, and lets a token in only when it names this one connection of your organization. Each token lasts 15 minutes, and the access it buys lasts an hour at most.
- ledger: "No stored cloud keys…", "The role trusts only one named connection of one organization"; docs `security/what-your-cloud-trusts`

**Artifact:** the AWS trust policy the setup script writes (`AwsConnectionScripts.trustPolicy`), for the example connection `acme-aws-prod`.
**Check it yourself:** `curl https://planton.ai/.well-known/openid-configuration`

### 3. What the GitHub App sees

Only the repositories you choose on GitHub's own install screen. GitHub gives Planton tokens that last about an hour, so there is nothing to paste or rotate.
- `wiki/product.integrations.github.md`; ledger "A GitHub connection's Verify…" (evidence: `gh api /apps/planton-ai`, 2026-11-25)

**Artifact:** what the App asks GitHub for, and why (the product's own capability words, `GithubConnectionCapability`):
| GitHub permission | What Planton does with it |
|---|---|
| Contents | clone your repository to build it, and open pull requests for you |
| Pull requests | build a preview for each pull request |
| Checks | report each build and preview as a check on its commit |
| Deployments | record each deploy as a GitHub deployment |
| Actions | show your GitHub Actions runs, and re-run or cancel them for you |
| Metadata (read only) | list the repositories you installed it on |

Events it hears: pushes and pull requests, to start builds and previews; GitHub Actions runs and jobs, to show them as they happen.
- `GithubConnectionCapability` (the product's one list); the production App also asks for repository webhooks, which Planton never uses: removed from the App before deploy (issues ledger)

**Check it yourself:** github.com/apps/planton-ai lists the same permissions.

### 4. Who approves what

You choose which environments are protected. In a protected environment, every infrastructure change and every deploy waits for an approval from someone other than the person who asked. Protected environments come with the Team plan.
- ledger: "In a protected environment…", "Production can require a second person's approval…", "Seven built-in roles…"; chapter 4 (your rules hold)

**Artifact:** who can approve a change in a protected environment:
| Who | Can approve? |
|---|---|
| The person who asked | No |
| Their coding agent | No |
| Planton's assistant | No |
| A teammate you gave approval on that environment | Yes |
- ledger: "Neither the Assistant nor a coding agent can approve a run for its person"

### 5. The record

Every infrastructure change is kept for the life of your organization: its exact configuration, who asked for it, the coding agent that made it, what it cost and who approved it. Every change to who can do what is kept too, and you can export it. If Planton's staff ever need to look inside your organization, your owners get an email with the reason.
- ledger: "Every change names the person and the coding agent…", "They are told every time Planton staff reach into their organization"; chapter 5 (every deployment leaves a record)

**Check it yourself:** `planton activity --agent claude --since 7d`

### 6. Cut Planton off

Delete the role in your cloud and Planton holds nothing it can use. Uninstall the App in GitHub. Your infrastructure keeps running, and the manifests and the open-source modules behind it stay yours.
- ledger: "They cut it off in their own cloud…", "Leaving never touches their cloud…"

**Artifact:** what to delete, per cloud (docs `security/what-your-cloud-trusts`, "How to Revoke"), and `planton connect aws cleanup acme-aws-prod`, which prints the exact commands.

**Read more:** What Your Cloud Trusts · Keyless Cloud Connections · Planton-Hosted Runners · Audit Trails (docs).

## Header

Docs · Pricing · Security · Sign In · **Start Free**

## Footer

- **Planton:** Pricing · Security · Changelog · Planton Desktop
- **Get Started:** Start Free · Docs
- **Open Source:** Planton on GitHub · The Catalog
- **Help:** Status · Discord Community
- Bottom line: Privacy · Terms · Refunds

## 404

- 404 · There's no page here.
- The page you followed may have moved or been retired.
- Run your cloud without a DevOps team. (the stand, read from chapter 1)
- Doors: as today.
