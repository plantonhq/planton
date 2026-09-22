# Privacy Policy

*Effective date: March 25, 2026*

*Last updated: March 25, 2026*

## Introduction

We at Planton Cloud, Inc. ("**Planton**", "**we**", or "**us**") are committed to protecting your privacy and safeguarding the information you share with us. This Privacy Policy ("**Policy**") explains how we collect, use, disclose, and process your personal data when you use Planton's platform, APIs, documentation, and related tools, including the website at [planton.ai](https://planton.ai) and all related software made available by Planton to deploy infrastructure, manage services, and automate DevOps workflows (collectively, the "**Service**").

This Policy does not apply where Planton acts as a data processor on behalf of commercial customers using our enterprise or self-hosted offerings. In those cases, our use of your data is governed by the applicable customer agreement.

## 1. Personal Data We Collect

### A. Data You Provide Directly

- **Account Information.** When you create a Planton account, we collect identifiers such as your name, email address, profile picture, and organization name. Accounts are provisioned through our identity provider, which supports sign-in via email/password, Google, GitHub, and Microsoft.

- **Payment Information.** If you subscribe to a paid plan, we collect billing details including your name, billing address, and payment method. Payment processing is handled by Stripe, Inc. We do not store full credit card numbers on our servers.

- **Infrastructure Configuration.** You may submit infrastructure deployment manifests, service configurations, environment variables, and related YAML or form-based inputs ("**Inputs**") to the Service. The Service processes these Inputs to provision infrastructure, build container images, and deploy services, producing deployment outputs, logs, and status information ("**Outputs**"). Inputs and Outputs are collectively referred to as "**Content**."

- **Communication Information.** If you contact us for support or other inquiries, we collect your name, contact information, and the contents of any messages you send.

- **Feedback.** If you provide feedback, feature requests, bug reports, or rate any aspect of the Service, we may store that information along with relevant context.

### B. Data We Collect Automatically

- **Device Information.** Your device or browser automatically sends us information when you access or use the Service, including device type, operating system, browser type, and mobile network or ISP.

- **Log Information.** We collect IP addresses, browser settings, error logs, request timestamps, and information about how you interact with the Service.

- **Usage Data.** We collect information about how you use the Service, such as pages visited, features used, infrastructure resources deployed, pipeline runs triggered, and time spent on various sections.

- **Cookies and Similar Technologies.** We use cookies and similar technologies to maintain sessions, remember your preferences, and analyze usage patterns. See Section 9 for details.

### C. Data Received from Third-Party Sign-In Providers

When you sign in to Planton using a third-party identity provider such as Google, GitHub, or Microsoft, we receive certain profile information from that provider. The specific data depends on the provider and the permissions you grant during sign-in:

- **Google Sign-In.** We receive your name, email address, and profile picture from your Google Account. We use this information solely to create and manage your Planton account, display your identity within the Service, and communicate with you. We do not request access to your Google Drive, Gmail, Calendar, or any other Google service data beyond basic profile information.

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

## 3. How We Use Personal Data

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

## 4. How We Share Personal Data

We may disclose personal data in the following circumstances:

- **Service Providers.** We share data with third-party vendors who help us operate the Service, including cloud hosting providers (Google Cloud Platform), identity providers (Auth0), payment processors (Stripe), analytics services, and customer support tools. These parties process data only as necessary to perform services on our behalf.

- **Business Transfers.** In the event of a merger, acquisition, or similar corporate transaction, personal data may be disclosed to counterparties and advisers as part of due diligence or transferred as part of the transaction.

- **Legal Compliance.** We may disclose personal data to government authorities or other third parties when we believe it is necessary to comply with applicable law, respond to lawful requests, protect safety or rights, prevent fraud, or enforce our Terms of Service.

- **Affiliates.** We may share personal data with our corporate affiliates, who will use it consistent with this Policy.

- **Organization Administrators.** If you create an account using an email associated with an organization on Planton, administrators of that organization may access account-related information such as your email address, account status, and activity within the organization's resources.

- **With Your Consent.** We may disclose personal data when you direct us to do so or consent to such disclosure.

We do not sell your personal data to third parties.

## 5. Third-Party Sign-In Provider Data

### Google

Planton's use and transfer to any other app of information received from Google APIs will adhere to the [Google API Services User Data Policy](https://developers.google.com/terms/api-services-user-data-policy), including the Limited Use requirements.

When you authenticate with Google, we access only your basic profile information (name, email address, and profile picture). We use this data to:

- Create and maintain your Planton account.
- Display your name and avatar within the Service.
- Send you account-related communications.

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

We do not:

- Use Microsoft user data for serving advertisements.
- Sell Microsoft user data to third parties.
- Use Microsoft user data for purposes unrelated to providing and improving the Service.
- Allow humans to read your Microsoft user data unless we have your affirmative consent, it is necessary for security purposes, to comply with applicable law, or the data is aggregated and anonymized for internal operations.

You can revoke Planton's access to your Microsoft account at any time through your [Microsoft account permissions page](https://account.microsoft.com/privacy/app-access).

## 6. Data Retention

We retain personal data for as long as necessary to provide the Service and fulfill the purposes described in this Policy, including legal compliance, dispute resolution, safety, and enforcement of agreements.

Infrastructure deployment history, audit trails, and configuration versions are retained as part of the Service's core functionality (versioning and auditability). You may request deletion of your account and associated personal data, subject to our legal obligations.

When personal data is no longer needed, we follow procedures to delete, de-identify, or anonymize it in compliance with applicable law.

## 7. Security

We implement commercially reasonable technical and organizational measures to protect personal data, including:

- Encryption in transit (TLS) and at rest.
- Reference-based secret management — plaintext credentials are never stored in the platform database.
- Role-based access control using OpenFGA (Zanzibar-based authorization).
- Audit trails for all resource mutations.
- Scoped cloud provider permissions (per-module, not blanket account access).
- Optional customer-hosted deployment runners that keep infrastructure operations within the customer's cloud boundary.

No method of transmission or storage is completely secure. We cannot guarantee absolute security, but we continuously evaluate and improve our security posture.

## 8. Your Rights and Choices

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

## 9. Cookies and Tracking Technologies

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

We do not use advertising or tracking cookies. You can control cookies through your browser settings. Disabling essential cookies may prevent you from using the Service.

## 10. International Data Transfers

Planton processes personal data on servers located in various jurisdictions, including the United States. If you are located in the European Economic Area (EEA), United Kingdom, or Switzerland, your personal data may be transferred to countries that have not been recognized as having an adequate level of data protection. When we engage in such transfers, we rely on legally valid transfer mechanisms, including standard contractual clauses published by the European Commission, to protect your data.

## 11. US State-Specific Disclosures

### California (CCPA/CPRA)

California residents have additional rights under the California Consumer Privacy Act and the California Privacy Rights Act, including the right to know what personal information is collected, the right to request deletion, the right to opt out of sales (we do not sell personal data), and the right to non-discrimination. To exercise these rights, contact [legal@planton.ai](mailto:legal@planton.ai).

In the preceding 12 months, we have collected the categories of personal information described in Section 1 of this Policy. We collect this information for the business purposes described in Section 3. We do not sell personal information and do not share personal information for cross-context behavioral advertising.

### Other US States

Residents of Colorado, Connecticut, Virginia, Utah, Texas, Oregon, Montana, and other states with consumer privacy laws may have similar rights, including the right to access, correct, and delete personal data, and the right to opt out of the sale of personal data. Contact [legal@planton.ai](mailto:legal@planton.ai) to exercise your rights under applicable state law.

## 12. Changes to This Policy

We may update this Policy from time to time. When we make material changes, we will update the effective date at the top of this page and notify you through the Service or by email. Your continued use of the Service after changes take effect constitutes acceptance of the updated Policy.

## 13. Contact Us

For questions about this Privacy Policy or to exercise your privacy rights, contact us at:

**Email:** [legal@planton.ai](mailto:legal@planton.ai)

**Planton Cloud, Inc.**
4902 North 9th Street, Apt 215
Fresno, CA 93726
United States
