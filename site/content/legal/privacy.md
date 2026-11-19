# Privacy Policy

*Effective date: March 25, 2026*

*Last updated: November 19, 2026*

## Introduction

We at Planton Cloud, Inc. ("**Planton**", "**we**", or "**us**") are committed to protecting your privacy and safeguarding the information you share with us. This Privacy Policy ("**Policy**") explains how we collect, use, disclose, and process your personal data when you use Planton's platform, APIs, documentation, and related tools, including the website at [planton.ai](https://planton.ai) and all related software made available by Planton to deploy infrastructure, manage services, and automate DevOps workflows (collectively, the "**Service**").

This Policy does not apply where Planton acts as a data processor on behalf of commercial customers using our enterprise or self-hosted offerings. In those cases, our use of your data is governed by the applicable customer agreement.

## 1. Personal Data We Collect

### A. Data You Provide Directly

- **Account Information.** When you create a Planton account, we collect identifiers such as your name, email address, profile picture, and organization name. Accounts are provisioned through our identity provider, which supports sign-in via email/password, Google, GitHub, and Microsoft.

- **Payment Information.** If you subscribe to a paid plan, we collect billing details including your name, billing address, and payment method. Payment processing is handled by Stripe, Inc. We do not store full credit card numbers on our servers.

- **Infrastructure Configuration.** You may submit infrastructure deployment manifests, service configurations, environment variables, and related YAML or form-based inputs ("**Inputs**") to the Service. The Service processes these Inputs to provision infrastructure, build container images, and deploy services, producing deployment outputs, logs, and status information ("**Outputs**"). Inputs and Outputs are collectively referred to as "**Content**."

- **Communication Information.** If you contact us for support or other inquiries, we collect your name, contact information, and the contents of any messages you send.

- **Status Page Subscriptions.** When you subscribe to updates on status.planton.ai, we collect what you give us for that purpose: an email address; a webhook address and a contact email; or, when you add a Slack or Discord channel through that app's own approval screen, the address the app gives us for posting to that channel, the channel's name and, for Slack, the name of its workspace. We use it only to send status updates and messages about your subscription. Every email and every channel message links to a page where you can change or end the subscription, and removing Planton Status from a Slack or Discord channel ends it too. The Subscribe page uses Cloudflare Turnstile to check that a request comes from a person; Turnstile processes your IP address and signals from your browser for that check. To limit how often one visitor can subscribe, we keep a keyed one-way fingerprint of your IP address for an hour, never the address itself.

- **Feedback.** If you provide feedback, feature requests, bug reports, or rate any aspect of the Service, we may store that information along with relevant context.

### B. Data We Collect Automatically

- **Device Information.** Your device or browser automatically sends us information when you access or use the Service, including device type, operating system, browser type, and mobile network or ISP.

- **Log Information.** We collect IP addresses, browser settings, error logs, request timestamps, and information about how you interact with the Service.

- **Usage Data.** We collect information about how you use the Service, such as pages visited, features used, infrastructure resources deployed, pipeline runs triggered, and time spent on various sections.

- **Cookies and Similar Technologies.** We use cookies and similar technologies to maintain sessions, remember your preferences, and analyze usage patterns. See Section 10 for details.

### C. Data Received from Third-Party Sign-In Providers

When you sign in to Planton using a third-party identity provider such as Google, GitHub, or Microsoft, we receive certain profile information from that provider. The specific data depends on the provider and the permissions you grant during sign-in:

- **Google Sign-In.** We receive your name, email address, and profile picture from your Google Account. We use this information solely to create and manage your Planton account, display your identity within the Service, and communicate with you. Signing in requests no access to your Google Drive, Gmail, Calendar, or any other Google service data beyond basic profile information. Connecting a Google Cloud account is a separate step you choose, described in Section 6.

- **GitHub Sign-In.** We receive your name, email address, profile picture, and GitHub username. Your GitHub username is also used to attribute Git commits to your Planton identity for CI/CD pipeline authorization and audit trails.

- **Microsoft Sign-In.** We receive your name, email address, and profile picture from your Microsoft account. We use this information solely to create and manage your Planton account, display your identity within the Service, and communicate with you. We do not request access to Microsoft 365, OneDrive, Outlook, or any other Microsoft service data beyond basic profile information.

### D. Data We Do Not Collect

Planton does not knowingly collect sensitive or special-category personal information such as health data, biometric data, religious beliefs, or genetic data. The Service is not directed at children under the age of 18. If we learn that a user is under 18, we will take appropriate steps to delete the associated account and personal data.

## 2. How Your Infrastructure Credentials Are Handled

This section describes an architecturally significant aspect of Planton's data handling that distinguishes us from many SaaS platforms.

**Planton never stores plaintext infrastructure credentials in the platform database.** When you configure a cloud provider connection (AWS, GCP, Azure, or others), sensitive fields — access keys, service account keys, tokens — are stored as references to entries in an organization-level secrets manager. The platform database contains only the reference identifier (a slug), never the secret value itself.

At execution time, secrets are resolved just-in-time by the subsystem that needs them and are discarded after use. They are not cached, logged, or persisted in any intermediate state.

For customers using the **customer-hosted runner** security model, secret resolution occurs entirely within the customer's infrastructure boundary. The Planton control plane receives only deployment status — it never sees or transmits plaintext credentials.

This architecture means that:

- A database breach would expose only secret reference slugs, not usable credentials.
- Credential rotation requires updating only the secrets manager entry — no connection reconfiguration needed.
- Audit trails track every secret access event independently.

## 3. Your Source Code and Repositories

To build and deploy your services, Planton needs access to the Git repositories you connect. This section explains exactly what that access is, what we do with it, and what we keep.

### A. How Access Is Granted

- **You choose the repositories.** You connect GitHub by installing Planton's GitHub App on the repositories you select. You can change that selection, or uninstall the app, at any time in GitHub.
- **Or you bring your own app.** With the "Your Own GitHub App" option, you create a GitHub App in your own GitHub organization, decide its permissions and repositories, and hold its private key. Planton features that need a permission you did not grant will not work.
- **No long-lived tokens.** Planton stores no GitHub token. Each time it needs access, it asks GitHub for a short-lived installation token, which expires within an hour. After you uninstall the app, GitHub issues no new tokens, so access ends once the last token expires.

### B. What We Read, and Why

Planton reads repository contents only to provide the Service:

- **To build your code.** A build copies the repository at the commit being built into an isolated build workspace. If you run builds on your own cluster, that copy is made in your cluster, not ours.
- **To keep your deployment settings current.** When you push, Planton reads the deployment files that the push changed (such as `service.yaml` and the configuration for each environment) and its own pipeline files.
- **To set up a service.** Planton reads a repository's metadata, languages, and list of files to suggest how to build and deploy it.
- **When you ask.** A member of your organization, or the Planton Assistant acting for one, can list files or read specific files at a commit.
- **For infrastructure modules.** If you point an infrastructure deployment at your own Terraform or Pulumi module repository, Planton copies that module to run it.

We do not scan, index, or analyze your repositories for any other purpose.

### C. What We Write

- **Build and deployment status.** Planton reports check runs, statuses, and deployments on your commits so you can see results in GitHub.
- **Pull requests, on new branches only.** When you or an assistant acting for you asks Planton to propose a change, it creates a new branch and opens a pull request for your review. It never writes to your default branch or to any branch that already exists. The Planton Assistant opens one only after the person has seen the change and confirmed it.
- **Merges, only of its own pull requests.** Planton merges a pull request only when a member of your organization asks, only if Planton opened it, and only at the exact commit that person reviewed.

### D. What We Keep

- **Commit details.** For each build and deployment we keep the repository, branch, commit, commit message, author name and email, pull request title and number, and the paths of the files a commit changed. These make up your deployment history and audit trail.
- **Your deployment configuration.** The deployment files Planton reads, and the configuration rendered from them for each environment, are stored as part of your service's settings.
- **Build logs.** Logs from each build step are kept as part of your deployment history. They contain whatever your build prints, which can include fragments of your code.
- **Not your source code.** We do not copy your repositories into our databases or storage. A build workspace is deleted when the build succeeds. When a build fails, its workspace is kept for up to two days for troubleshooting, then deleted. Files read when you ask are shown to you and not stored. Notifications GitHub sends us about pushes and pull requests are kept for up to 72 hours while they are processed.

### E. Who Can See It

- **Members of your organization**, according to the roles you grant them.
- **Planton staff, only by time-limited access.** No one at Planton has standing access to your organization. A staff member can enter it only for a stated reason and a limited time. Read access lets them see what your organization's readers see, which can include repository files and build logs. Anything beyond read access needs a second Planton operator's approval. Every such access is recorded in your organization's access history.

### F. AI Features

If your organization uses the Planton Assistant, it reads repository files only when a conversation or an automation you have allowed needs them. Those files are sent to the AI service providers that run the Assistant (see Section 5). Your source code is covered by the same commitment on model training as the rest of your Content (Section 4).

## 4. How We Use Personal Data

We use personal data for the following purposes:

- To provide, operate, and maintain the Service, including infrastructure deployment, CI/CD pipeline execution, and platform features.
- To create, manage, and administer your account, including processing payments and responding to support requests.
- To improve and develop the Service, including debugging, performance analysis, and feature development.
- To communicate with you about updates, changes to the Service, security alerts, and administrative messages.
- To prevent, detect, and investigate fraud, abuse, security incidents, and violations of our [Terms of Service](/legal/terms).
- To comply with legal obligations and protect the rights, safety, and property of users, Planton, and third parties.
- To enforce our [Terms of Service](/legal/terms) and other applicable agreements.

**We do not use your Inputs or Outputs to train machine learning models**, or permit third parties to use them for training, unless: (1) they are flagged for security review, (2) you explicitly report them to us as feedback, or (3) you have explicitly consented to such use.

We may aggregate or de-identify data so that it no longer identifies you, and use such data for analytics, research, and Service improvement. We will not attempt to re-identify de-identified data except as required by law.

## 5. How We Share Personal Data

We may disclose personal data in the following circumstances:

- **Service Providers.** We share data with third-party vendors who help us operate the Service, including cloud hosting providers (Google Cloud Platform, and Cloudflare, which runs status.planton.ai and its bot check), identity providers (Auth0), payment processors (Stripe), email delivery (Resend, which sends the Service's email, including status updates), Slack and Discord (which deliver status messages to the channels you add), AI service providers that run the Planton Assistant (Stigmer and Cursor, with Anthropic's Claude models), analytics services, and customer support tools. These parties process data only as necessary to perform services on our behalf.

- **Business Transfers.** In the event of a merger, acquisition, or similar corporate transaction, personal data may be disclosed to counterparties and advisers as part of due diligence or transferred as part of the transaction.

- **Legal Compliance.** We may disclose personal data to government authorities or other third parties when we believe it is necessary to comply with applicable law, respond to lawful requests, protect safety or rights, prevent fraud, or enforce our Terms of Service.

- **Affiliates.** We may share personal data with our corporate affiliates, who will use it consistent with this Policy.

- **Organization Administrators.** If you create an account using an email associated with an organization on Planton, administrators of that organization may access account-related information such as your email address, account status, and activity within the organization's resources.

- **With Your Consent.** We may disclose personal data when you direct us to do so or consent to such disclosure.

We do not sell your personal data to third parties.

## 6. Third-Party Sign-In Provider Data

### Google

Planton's use and transfer to any other app of information received from Google APIs will adhere to the [Google API Services User Data Policy](https://developers.google.com/terms/api-services-user-data-policy), including the Limited Use requirements.

**Signing in with Google.**

When you sign in with Google, we access only your basic profile information (name, email address, and profile picture). We use this data to:

- Create and maintain your Planton account.
- Display your name and avatar within the Service.
- Send you account-related communications.

**Connecting a Google Cloud account.**

Separately from signing in, a member of your organization can connect a Google Cloud project so Planton can deploy in it. One way to connect uses your Google account in the browser: the one-click keyless setup, **Set It Up for Me**. It asks Google for the `cloud-platform` permission, which lets Planton act in Google Cloud as the person who approves it, within what that person's own Google Cloud roles allow, and for your email address, to record who approved the setup. Planton asks for it only when you start the setup, and you see Google's consent screen before anything happens.

You approve once, and Planton uses that approval only to set up a keyless trust in the Google Cloud project you chose. It reads that project (its number, and the organization and billing account it belongs to) to confirm your account can see it, and checks that your account may make every change before it makes any. It then turns on the Google Cloud APIs the trust and Planton need in that project (IAM, IAM Service Account Credentials, Security Token Service, Cloud Resource Manager and Service Usage), creates a workload identity pool and an OIDC provider that trust only this connection's tokens from Planton's identity issuer, creates a service account named "Planton Provisioner" (or uses the one you named), and creates this connection's own custom role, holding exactly the permissions of what you chose to let Planton deploy, and grants it to that service account. If you turned on organization access, the same role is also created in the project's organization, and billing roles are granted on the project's billing account. When you later change what Planton may deploy and choose **Update the Role for Me**, Planton asks for the same approval once more and uses it only to run the same setup again, which keeps what already exists and rewrites that role and its grants.

Your Google access token is used only during the setup request it was issued for and is never stored. Planton requests online access only, so Google issues no refresh token, and no service account key is created. From then on, Planton reaches the project through the keyless trust, which you can remove in your own project at any time.

The email address of the Google account that approved the setup may appear in our service logs. To end the permission on Google's side, remove Planton at your [Google Account permissions page](https://myaccount.google.com/permissions). To remove the trust itself, delete the OIDC provider, the custom role and the service account Planton created in your project; when you delete the connection in Planton, it shows exactly what stays in your project and a script that removes it.

We use data received through this Google Cloud permission only to set up, and later update, the trust you asked for.

**What we do not do with Google user data.**

We do not:

- Use Google user data for serving advertisements.
- Sell Google user data to third parties.
- Use Google user data for purposes unrelated to providing and improving the Service.
- Allow humans to read your Google user data unless we have your affirmative consent, it is necessary for security purposes, to comply with applicable law, or the data is aggregated and anonymized for internal operations.

You can revoke Planton's access to your Google account at any time through your [Google Account permissions page](https://myaccount.google.com/permissions).

### Microsoft

When you authenticate with Microsoft, we access only your basic profile information (name, email address, and profile picture). We use this data to:

- Create and maintain your Planton account.
- Display your name and avatar within the Service.
- Send you account-related communications.

**Connecting an Azure subscription.**

Separately from signing in, a member of your organization can connect an Azure subscription so Planton can deploy in it. One way to connect uses your Microsoft account in the browser: the one-click keyless setup, **Set It Up for Me**. On one consent screen it asks Microsoft for two delegated permissions: `Application.ReadWrite.All` on Microsoft Graph, to create the app registration Planton signs in as and its federated credential, and `user_impersonation` on Azure Resource Manager, to create this connection's role in your subscription and assign it. It also asks for your basic profile, to record who approved the setup and in which tenant. Both permissions act as the person who approves them, within what that person's own Microsoft Entra and Azure roles allow; Microsoft requires an Entra administrator to consent to `Application.ReadWrite.All`. Planton asks for them only when you start the setup, and you see Microsoft's consent screen before anything happens.

You approve once, and Planton uses that approval only to set up a keyless trust for the subscription you chose. It creates an app registration and its service principal (or uses the ones that already exist under the name you gave), adds a federated identity credential that trusts only this connection's tokens from Planton's identity issuer, confirms your account can see the subscription, creates this connection's own role, "Planton Deployer (*organization*/*connection*)", holding exactly the actions of what you chose to let Planton deploy, and assigns it to the app's service principal on the subscription. No client secret is created. Planton stores the app's client ID on the connection: an identifier that names the app and signs nothing. When you later change what Planton may deploy and choose **Update the Role for Me**, Planton asks Microsoft once more, for `user_impersonation` on Azure Resource Manager alone, and uses it only to rewrite that role.

Each Microsoft access token is used only during the setup request it was issued for and is never stored. Planton never requests offline access, so Microsoft issues no refresh token. From then on, Planton reaches the subscription through the keyless trust, which you can remove in your own tenant at any time.

The email address of the Microsoft account that approved the setup may appear in our service logs. To end the permission on Microsoft's side, a tenant administrator removes Planton in the Microsoft Entra admin center under **Enterprise applications** (revoke its permissions, or delete it), and anyone can review the apps they have consented to at [My Apps](https://myapps.microsoft.com). To remove the trust itself, delete the federated credential, the role assignment and the role Planton created (and the app registration, if no other Planton connection uses it); when you delete the connection in Planton, it shows exactly what stays in your tenant and a script that removes it.

We use data received through these Microsoft permissions only to set up, and later update, the trust you asked for.

We do not:

- Use Microsoft user data for serving advertisements.
- Sell Microsoft user data to third parties.
- Use Microsoft user data for purposes unrelated to providing and improving the Service.
- Allow humans to read your Microsoft user data unless we have your affirmative consent, it is necessary for security purposes, to comply with applicable law, or the data is aggregated and anonymized for internal operations.

You can revoke Planton's access to your Microsoft account at any time through your [Microsoft account permissions page](https://account.microsoft.com/privacy/app-access).

## 7. Data Retention

We retain personal data for as long as necessary to provide the Service and fulfill the purposes described in this Policy, including legal compliance, dispute resolution, safety, and enforcement of agreements.

Infrastructure deployment history, audit trails, and configuration versions are retained as part of the Service's core functionality (versioning and auditability). You may request deletion of your account and associated personal data, subject to our legal obligations.

Status page subscriptions are kept until you unsubscribe or remove Planton Status from the channel. The status page's database, which includes subscribers' addresses, is copied every hour for recovery, and each copy is deleted after 30 days, so an address you remove is gone from every copy within 30 days. When an email address bounces or draws a spam complaint, we keep a one-way fingerprint of it so that nothing more is sent to it.

When personal data is no longer needed, we follow procedures to delete, de-identify, or anonymize it in compliance with applicable law.

## 8. Security

We implement commercially reasonable technical and organizational measures to protect personal data, including:

- Encryption in transit (TLS) and at rest.
- Reference-based secret management — plaintext credentials are never stored in the platform database.
- Role-based access control using OpenFGA (Zanzibar-based authorization).
- Audit trails for all resource mutations.
- Scoped cloud provider permissions (per-module, not blanket account access).
- Optional customer-hosted deployment runners that keep infrastructure operations within the customer's cloud boundary.

No method of transmission or storage is completely secure. We cannot guarantee absolute security, but we continuously evaluate and improve our security posture.

## 9. Your Rights and Choices

Depending on your location and applicable law, you may have rights regarding your personal data, including:

- **Access** — request a copy of the personal data we hold about you.
- **Correction** — request that we correct inaccurate personal data.
- **Deletion** — request that we delete your personal data, subject to certain exceptions.
- **Portability** — request your data in a structured, machine-readable format.
- **Restriction** — request that we limit processing of your data in certain circumstances.
- **Objection** — object to processing based on legitimate interests.
- **Withdrawal of Consent** — where processing is based on consent, withdraw that consent at any time.

To exercise any of these rights, contact us at [legal@planton.ai](mailto:legal@planton.ai). We may request information to verify your identity before processing your request. We will respond to verified requests within the timeframes required by applicable law.

**No sale or targeted advertising.** We do not sell personal data or use it for cross-contextual behavioral advertising as those terms are defined under applicable US state privacy laws.

## 10. Cookies and Tracking Technologies

We use cookies and similar technologies to:

- Maintain your session and authentication state.
- Remember your preferences and settings.
- Analyze usage patterns and improve the Service.
- Provide security protections.

| Cookie Type | Purpose | Duration |
|---|---|---|
| **Essential** | Authentication, session management, security | Session or short-lived |
| **Functional** | User preferences, language settings | Persistent (up to 1 year) |
| **Analytics** | Usage patterns, feature adoption, performance | Persistent (up to 2 years) |

status.planton.ai sets no cookie unless you choose a theme (a preference cookie) or add a Slack or Discord channel (an essential cookie that lasts ten minutes and ties the app's approval to your browser). We do not use advertising or tracking cookies. You can control cookies through your browser settings. Disabling essential cookies may prevent you from using the Service.

## 11. International Data Transfers

Planton processes personal data on servers located in various jurisdictions, including the United States. If you are located in the European Economic Area (EEA), United Kingdom, or Switzerland, your personal data may be transferred to countries that have not been recognized as having an adequate level of data protection. When we engage in such transfers, we rely on legally valid transfer mechanisms, including standard contractual clauses published by the European Commission, to protect your data.

## 12. US State-Specific Disclosures

### California (CCPA/CPRA)

California residents have additional rights under the California Consumer Privacy Act and the California Privacy Rights Act, including the right to know what personal information is collected, the right to request deletion, the right to opt out of sales (we do not sell personal data), and the right to non-discrimination. To exercise these rights, contact [legal@planton.ai](mailto:legal@planton.ai).

In the preceding 12 months, we have collected the categories of personal information described in Section 1 of this Policy. We collect this information for the business purposes described in Section 4. We do not sell personal information and do not share personal information for cross-context behavioral advertising.

### Other US States

Residents of Colorado, Connecticut, Virginia, Utah, Texas, Oregon, Montana, and other states with consumer privacy laws may have similar rights, including the right to access, correct, and delete personal data, and the right to opt out of the sale of personal data. Contact [legal@planton.ai](mailto:legal@planton.ai) to exercise your rights under applicable state law.

## 13. Changes to This Policy

We may update this Policy from time to time. When we make material changes, we will update the effective date at the top of this page and notify you through the Service or by email. Your continued use of the Service after changes take effect constitutes acceptance of the updated Policy.

## 14. Contact Us

For questions about this Privacy Policy or to exercise your privacy rights, contact us at:

**Email:** [legal@planton.ai](mailto:legal@planton.ai)

**Planton Cloud, Inc.**
4902 North 9th Street, Apt 215
Fresno, CA 93726
United States
