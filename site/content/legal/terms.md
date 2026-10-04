# Terms of Service

*Effective date: March 25, 2026*

*Last updated: August 20, 2026*

Welcome, and thank you for your interest in Planton Cloud, Inc. ("**Planton**", "**we**", or "**us**"). These Terms of Service ("**Terms**") govern your access to and use of Planton's platform, APIs, documentation, and related tools, including the website at [planton.ai](https://planton.ai) and all related software made available by Planton to deploy infrastructure, manage services, and automate DevOps workflows (collectively, the "**Service**").

By using the Service, you agree to these Terms. Please also read our [Privacy Policy](/legal/privacy), which explains how we collect, use, and process personal data.

If you are using the Service under a Master Services Agreement ("**MSA**") or other enterprise agreement with Planton, that agreement governs to the extent it conflicts with these Terms.

If you are entering into these Terms on behalf of a company or other legal entity, you represent that you have the authority to bind that entity.

## 1. The Service

### 1.1 Description

Planton is a DevOps automation platform that provides infrastructure deployment, service CI/CD, and operational tooling across multiple cloud providers. The Service includes:

- **Infrastructure Hub** — self-service deployment of cloud infrastructure resources across AWS, GCP, Azure, Kubernetes, and other providers using pre-built modules backed by Terraform and Pulumi.
- **Service Hub** — Git-connected CI/CD pipelines for building, containerizing, and deploying backend services using Tekton and Cloud Native BuildPacks.
- **Runner** — the execution engine that performs infrastructure and service deployment operations, available in Planton-hosted or customer-hosted configurations.
- **AI Assistant** (Beta) — AI-powered assistance for operating the platform, troubleshooting pipelines, and debugging deployments. Currently available on Planton-hosted organizations and the Planton desktop app; not yet available on Self-Hosted Platform deployments.
- **Self-Hosted Platform** — the Planton platform deployed in infrastructure you own and control. The free community edition requires no license; org-scale capabilities unlock with a paid license key (Section 3.3).

The Service also leverages **Planton open source**, an open-source multi-cloud infrastructure framework licensed under the Apache License 2.0, which provides the infrastructure-as-code modules used by the platform. Planton open source is available at [github.com/plantonhq/planton](https://github.com/plantonhq/planton) and is governed by its own open-source license, independent of these Terms.

### 1.2 Content

You may provide inputs to the Service, including infrastructure configurations, deployment manifests, service definitions, environment variables, and code repository references ("**Inputs**"). The Service generates deployment outputs, logs, status information, and, in the case of the AI Assistant, diagnostic suggestions ("**Outputs**"). Inputs and Outputs are collectively "**Content**."

We may use Content to provide the Service, comply with applicable law, enforce these Terms, and maintain the security of the Service. By submitting Inputs, you represent that you have all rights, licenses, and permissions necessary for us to process them.

### 1.3 Model Training

**Planton will not use your Content to train, or allow any third party to train, any AI models, unless you have explicitly consented to such use.** The AI Assistant processes your Content in real time to provide its outputs but does not retain it for model training.

### 1.4 Limitations of AI Outputs

The AI Assistant and other AI-powered features generate Outputs automatically using machine learning. You acknowledge that AI-generated Outputs may contain errors, may be incomplete, and should not be relied upon without independent verification. You are responsible for evaluating and bearing all risks associated with the use of AI-generated Outputs.

### 1.5 Use Restrictions

Except to the extent prohibited by applicable law, you may not:

- Reverse engineer, decompile, or disassemble the Service (this restriction does not apply to Planton open source modules, which are governed by the Apache License 2.0).
- Reproduce, modify, or create derivative works of the proprietary portions of the Service.
- Rent, lease, lend, resell, or sublicense the Service.
- Remove proprietary notices from the Service.
- Use the Service or Outputs to develop a competing product, or engage in model extraction attacks against the AI Assistant.
- Probe, scan, or attempt to penetrate the Service's security.
- Harvest or scrape data from the Service.
- Use the Service in any manner that violates applicable law or infringes third-party rights.
- Knowingly submit data subject to heightened regulatory protections (e.g., HIPAA, PCI DSS, GLBA) unless you have a separate agreement with Planton that specifically covers such data.

### 1.6 Beta Services

Features designated as beta, preview, early access, or similar labels ("**Beta Services**") are provided for evaluation purposes, may change or be discontinued without notice, are not fully supported, and are offered "as-is" without warranty. **The AI Assistant is currently a Beta Service.** Planton shall have no liability arising from your use of Beta Services.

## 2. Accounts

### 2.1 Eligibility

You must be at least 18 years old (or the age of majority in your jurisdiction, whichever is higher) to use the Service.

### 2.2 Registration

To access the Service, you must register for an account. The Service supports sign-in via email/password, Google, and GitHub. You agree to provide accurate information and keep it current. You are responsible for maintaining the security of your account credentials and for all activity under your account. If you believe your account has been compromised, notify us immediately at [legal@planton.ai](mailto:legal@planton.ai).

### 2.3 Organizations

The Service supports organizational accounts where multiple users collaborate under shared resources. Organization owners have administrative control over the organization's configuration, members, and Content. If you use the Service through an organization provided by your employer, your employer's administrator may manage your access and Content.

## 3. Payment

### 3.1 Pricing

Certain features of the Service require payment. Current pricing is available on our [pricing page](/pricing). Planton reserves the right to change pricing with reasonable advance notice. Your continued use of the Service after a price change takes effect constitutes acceptance of the new pricing.

### 3.2 Subscriptions

Paid plans are billed in advance on a monthly or annual basis ("**Subscription Period**") and renew automatically unless cancelled before the renewal date. Free-tier organizations carry no payment method and are never charged: when a free-tier limit is reached, the affected action pauses rather than incurring a charge. Refunds are governed by our [Refund Policy](/legal/refund-policy) and Section 8.

### 3.3 Self-Hosted Licenses

A self-hosted license is a yearly purchase that unlocks org-scale capabilities on a Self-Hosted Platform deployment. It is delivered as a signed license key sent to the purchase email address, and it renews automatically each year unless cancelled — each renewal delivers a fresh key. You may cancel renewal at any time by emailing [support@planton.ai](mailto:support@planton.ai) from the purchase email address; the license then runs out its paid term without renewing.

- **Grant.** Subject to payment, Planton grants you a non-exclusive, non-transferable license to use the licensed capabilities on one self-hosted deployment, up to the seat count and for the term stated in the offer you purchased.
- **Verification is offline.** License keys verify locally inside your deployment; they never phone home, and air-gapped deployments are fully supported.
- **Expiry never disables your deployment.** When a license expires, a grace period with full capabilities begins; after it, the deployment steps down to the free community edition. Workloads already deployed keep running, nothing is deleted, and reads and deletes are never blocked.
- **Updates.** The license's update window determines how long the deployment is entitled to product updates. Features you have do not stop working when the update window closes.
- **Revocation.** Planton may revoke a license in the event of a refund, chargeback, or material breach of these Terms. Revocation stops renewals and closes the update window; it never reaches into a running deployment, which degrades gracefully as described above.

### 3.4 Prepaid AI Credits

AI-powered features consume prepaid credits purchased as packs. Credits fund AI usage on Planton-hosted organizations and the Planton desktop app; Self-Hosted Platform deployments do not use credits. Credits are dollar-denominated, drawn down as AI usage occurs, and do not expire while your account exists. Spend protection is on by default: automatic top-up occurs only if you enable it, at the threshold and amount you configure. Credit purchases are refundable as described in the [Refund Policy](/legal/refund-policy); consumed credits are not refundable. Credits carry no cash value and are not transferable between accounts.

### 3.5 Payment Processing

Payments are processed by Stripe, Inc. ("**Stripe**"). By making a purchase, you agree to Stripe's terms of service and, for recurring purchases, authorize recurring charges to your payment method. Planton does not store full payment card numbers.

### 3.6 Taxes

Where Planton is required to collect sales tax, VAT, GST, or similar taxes, they are calculated and collected at checkout based on your location. Depending on your market, displayed prices may be inclusive or exclusive of such taxes; the checkout page always shows the final amount before you pay. You are responsible for any other taxes associated with your use of the Service, other than taxes based on Planton's net income.

### 3.7 Delinquent Accounts

Planton may suspend or terminate access to the Service for any account with overdue payments after reasonable notice. A subscription whose payment fails is warned first and retains access during the retry window; only after payment retries are exhausted is the organization's ability to create new resources suspended — existing workloads, reads, and the ability to fix the payment are never blocked.

## 4. Intellectual Property

### 4.1 Planton's Rights

Planton and its licensors retain all right, title, and interest in the Service, including all intellectual property rights. These Terms do not grant you any implied licenses. The Service's proprietary components — including the platform UI, orchestration engine, and the AI Assistant — are protected by copyright, trade secret, and other intellectual property laws.

### 4.2 Your Content

You retain all right, title, and interest in your Inputs. To the extent Planton may hold any rights in Outputs generated by the Service, Planton assigns those rights to you. This assignment does not extend to the underlying Service, models, or algorithms.

### 4.3 Open Source

The Service incorporates open-source components, including Planton open source (Apache License 2.0), Tekton, and Cloud Native BuildPacks. Your use of these components is governed by their respective open-source licenses. Planton open source modules are available at [github.com/plantonhq/planton](https://github.com/plantonhq/planton).

Nothing in these Terms restricts your rights under applicable open-source licenses or requires you to use the Service to utilize Planton open source.

### 4.4 Feedback

If you provide ideas, suggestions, or other feedback about the Service ("**Feedback**"), you grant Planton a perpetual, irrevocable, royalty-free license to use that Feedback for any purpose without obligation to you.

### 4.5 Usage Data

Planton may collect, analyze, and use technical data about your use of the Service ("**Usage Data**"), including logs, performance metrics, and feature usage patterns, for its business purposes such as analytics, security, and Service improvement. Usage Data may be disclosed to third parties only in aggregated or de-identified form. Usage Data does not include your Content.

## 5. Security and Infrastructure

### 5.1 Customer Cloud Infrastructure

The Service deploys infrastructure and services into cloud accounts that you own and control (AWS, GCP, Azure, or other providers). You are responsible for the security and configuration of your cloud accounts. Planton accesses your cloud accounts only through credentials you provide, scoped to the specific permissions required by each operation.

### 5.2 Security Models

Planton offers multiple security models for cloud provider access:

- **Standard credentials** — you provide access keys or service account credentials stored as encrypted references in Planton's secrets manager.
- **Cross-account roles** — you configure trust policies that allow Planton to assume a role in your account without long-lived credentials.
- **Customer-hosted runner** — you deploy Planton's runner image within your own infrastructure. Infrastructure operations execute entirely within your cloud boundary; the Planton control plane receives only status information.

You are responsible for selecting the security model appropriate for your compliance and security requirements.

### 5.3 Secret Management

Planton's platform stores sensitive credentials as references to entries in an organization-level secrets manager, never as plaintext values. Secrets are resolved just-in-time during execution and discarded after use. For details, see the "How Your Infrastructure Credentials Are Handled" section of our [Privacy Policy](/legal/privacy).

## 6. Acceptable Use

You agree to use the Service in compliance with all applicable laws and regulations and in accordance with these Terms. You will not use the Service to:

- Deploy infrastructure or services for unlawful purposes.
- Store or process data in violation of applicable data protection laws.
- Interfere with or disrupt the Service or its underlying infrastructure.
- Attempt to gain unauthorized access to other users' accounts or resources.
- Circumvent usage limits, rate limits, or security controls.

## 7. Third-Party Services

The Service integrates with third-party services, including cloud providers (AWS, GCP, Azure), source code management platforms (GitHub, GitLab), container registries, identity providers, and payment processors. Your use of third-party services is governed by their respective terms and privacy policies. Planton does not control and is not responsible for third-party services.

## 8. Termination

### 8.1 By You

You may stop using the Service and close your account at any time. If you have an active subscription, cancellation takes effect at the end of the current Subscription Period.

### 8.2 By Planton

Planton may suspend or terminate your access to the Service at any time for cause, including violation of these Terms, with notice where practicable. We may also discontinue the Service or any feature with reasonable advance notice.

If we terminate your paid subscription without cause, we will refund a pro-rata portion of prepaid fees for the remaining Subscription Period. No refund is owed if termination is due to your breach of these Terms.

### 8.3 Effect of Termination

Upon termination, your right to access the Service ceases. Planton may delete Content associated with your account after a reasonable retention period. You should export any Content you wish to retain before termination.

Your infrastructure and services deployed in your own cloud accounts are not affected by termination of your Planton account — they continue to run under your direct control. Planton open source modules remain available under the Apache License 2.0.

### 8.4 Survival

Sections 4 (Intellectual Property), 9 (Disclaimer of Warranties), 10 (Limitation of Liability), 11 (Indemnification), and 14 (General) survive termination.

## 9. DISCLAIMER OF WARRANTIES

THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE." PLANTON DISCLAIMS ALL WARRANTIES, WHETHER EXPRESS, IMPLIED, OR STATUTORY, INCLUDING WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, TITLE, AND NON-INFRINGEMENT.

PLANTON DOES NOT WARRANT THAT THE SERVICE WILL BE UNINTERRUPTED, ERROR-FREE, OR SECURE; THAT DEFECTS WILL BE CORRECTED; OR THAT OUTPUTS (INCLUDING AI-GENERATED OUTPUTS) WILL BE ACCURATE, COMPLETE, OR RELIABLE. YOU USE THE SERVICE AND RELY ON OUTPUTS AT YOUR OWN RISK.

PLANTON DOES NOT WARRANT THE SECURITY, RELIABILITY, OR AVAILABILITY OF YOUR CLOUD PROVIDER ACCOUNTS OR THIRD-PARTY SERVICES INTEGRATED WITH THE SERVICE.

## 10. LIMITATION OF LIABILITY

### 10.1 No Indirect Damages

TO THE FULLEST EXTENT PERMITTED BY LAW, PLANTON WILL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, INCLUDING LOSS OF PROFITS, DATA, GOODWILL, OR BUSINESS OPPORTUNITY, ARISING OUT OF OR RELATED TO THESE TERMS OR THE SERVICE.

### 10.2 Liability Cap

TO THE FULLEST EXTENT PERMITTED BY LAW, PLANTON'S AGGREGATE LIABILITY FOR ALL CLAIMS ARISING OUT OF OR RELATED TO THESE TERMS AND THE SERVICE IS LIMITED TO THE GREATER OF: (A) THE AMOUNT YOU PAID TO PLANTON IN THE SIX (6) MONTHS PRECEDING THE EVENT GIVING RISE TO THE CLAIM, OR (B) ONE HUNDRED DOLLARS ($100).

THESE LIMITATIONS APPLY REGARDLESS OF THE THEORY OF LIABILITY AND EVEN IF PLANTON HAS BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGES.

## 11. Indemnification

You agree to indemnify, defend, and hold harmless Planton and its officers, directors, employees, and agents from any claims, liabilities, damages, and expenses (including reasonable attorneys' fees) arising out of: (a) your use of the Service; (b) your violation of these Terms; (c) your violation of applicable law; or (d) your Inputs infringing third-party rights.

## 12. Dispute Resolution

### 12.1 Binding Arbitration

Any dispute arising out of or relating to these Terms or the Service will be resolved through final and binding arbitration administered by the American Arbitration Association ("**AAA**") under its Consumer Arbitration Rules, except that you may bring individual claims in small claims court if your claims qualify.

By agreeing to these Terms, you and Planton each waive the right to a jury trial and to participate in a class action.

### 12.2 Class Action Waiver

You and Planton agree that disputes will be resolved on an individual basis only. Neither party may bring claims as a plaintiff or class member in any class, consolidated, or representative proceeding.

### 12.3 Opt-Out

You may opt out of arbitration within 30 days of account creation by sending written notice to [legal@planton.ai](mailto:legal@planton.ai) with your name and a clear statement of your intent to opt out.

### 12.4 Pre-Arbitration Resolution

Before initiating arbitration, the claiming party must send a written Notice of Dispute to the other party describing the claim and the relief sought. If the dispute is not resolved within 60 days of receipt, either party may commence arbitration.

## 13. Modifications

We may modify these Terms from time to time. Material changes will be communicated through the Service or by email at least 30 days before they take effect. Your continued use of the Service after the effective date constitutes acceptance of the modified Terms. If you do not agree to the changes, you must discontinue use of the Service.

## 14. General

### 14.1 Governing Law

These Terms are governed by the laws of the State of California, without regard to conflict-of-law principles. Subject to the Dispute Resolution section above, any legal action will be brought exclusively in the federal or state courts located in Fresno County, California.

### 14.2 Entire Agreement

These Terms, together with the Privacy Policy and the Refund Policy, constitute the entire agreement between you and Planton regarding the Service and supersede all prior agreements on the subject matter.

### 14.3 Assignment

You may not assign these Terms without Planton's written consent. Planton may assign these Terms without restriction.

### 14.4 Severability

If any provision of these Terms is held unenforceable, the remaining provisions remain in full force and effect.

### 14.5 No Waiver

Planton's failure to enforce any provision of these Terms does not constitute a waiver of that provision.

### 14.6 Export Controls

You must comply with all applicable trade laws, including US export control and sanctions laws. The Service may not be used in or for the benefit of any US-embargoed country or territory, or by any person on a US government restricted-parties list.

### 14.7 Electronic Communications

By using the Service, you consent to receiving communications from us electronically. You agree that electronic communications satisfy any legal requirement that communications be in writing.

## 15. Contact

For questions about these Terms, contact us at:

**Email:** [legal@planton.ai](mailto:legal@planton.ai)

**Planton Cloud, Inc.**
4902 North 9th Street, Apt 215
Fresno, CA 93726
United States
