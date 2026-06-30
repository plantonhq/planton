
# **Managing SSL Provisioning for Cloud Run Custom Domains**

## **Section 1: Executive Summary and Architectural Overview of Cloud Run Custom Domains**

### **1.1 The Default Cloud Run Endpoint vs. The Custom Domain Imperative**

By default, every service deployed on Google Cloud Run is assigned a unique, stable, and fully-managed HTTPS endpoint. This endpoint is a subdomain of \*.run.app.1 While this default URL is functional for development and testing, it is unsuitable for production applications. The use of a custom domain is a fundamental requirement for establishing brand identity, user trust, and integration with existing enterprise domain structures.

### **1.2 The Core Mechanism: Google-Managed SSL Provisioning**

The primary value of mapping a custom domain to Cloud Run is not merely DNS routing; it is the automated provisioning and, critically, the automated renewal of Google-managed SSL/TLS certificates.2 This mechanism provides a zero-maintenance, secure endpoint for the service.

Google automatically provisions these certificates from one of two Certificate Authorities (CAs): **Google Trust Services (GTS)** or **Let's Encrypt**.4 The selection of the CA is automatic and can be influenced by the domain's existing Certification Authority Authorization (CAA) DNS records.3

The underlying provisioning process relies on the **ACME HTTP-01 challenge**. This is evidenced by documentation warning that proxy services (like Cloudflare) with "Always use HTTPS" features can break certificate provisioning.5 These features interfere by redirecting the unencrypted HTTP request that the ACME protocol uses to validate domain control. For a certificate to be issued or renewed, the CA must successfully send a challenge to http://\<your-domain\>/.well-known/acme-challenge/\<token\>. The Google Front End (GFE) handling the Cloud Run traffic must be able to receive and respond to this unencrypted request over Port 80\. This reliance on an unencrypted HTTP challenge is a foundational architectural detail that creates a common point of failure when proxies or CDNs are misconfigured.

### **1.3 Comparative Analysis: The Three Patterns for Domain Mapping**

While the query focuses on a specific method, there are three distinct architectural patterns available for mapping a custom domain to a Cloud Run service.5 The choice of pattern has significant implications for cost, performance, security, and available features.

1. Pattern 1: Native Cloud Run Domain Mapping (The "Preview" Feature)  
   This is the most direct method, creating a DomainMapping resource that links a domain directly to a Cloud Run service.5 This feature, however, is in "Limited availability and Preview" and is not recommended for production workloads.5 Its most significant limitations are that it cannot be used for wildcard certificates (e.g., \*.example.com) and it only maps to the root path (/) of a service.5  
2. Pattern 2: Global External Application Load Balancer (GEALB) (The "Recommended" Feature)  
   This is Google's recommended, enterprise-grade method.5 In this architecture, the domain's DNS records point to the static IP address of a Global Application Load Balancer.8 The load balancer terminates the SSL (HTTPS) traffic and uses a Serverless Network Endpoint Group (NEG) as its backend to route requests to the appropriate Cloud Run service.9 This pattern overcomes all limitations of the native method, supporting wildcard certificates, Google Cloud Armor (WAF), Cloud CDN, and Identity-Aware Proxy (IAP).7  
3. Pattern 3: Firebase Hosting (The "Simplified" Feature)  
   This pattern uses the Firebase Hosting platform to act as a simple, CDN-fronted proxy for Cloud Run services.5 It is extremely simple to configure, has a generous free tier, and provides a global CDN by default.7 It is an excellent alternative for simple web applications or for users who find the GEALB too complex or costly.

### **1.4 Table 1: Architectural Feature Comparison (Domain Mapping Patterns)**

The following table provides a high-level comparison of the three available architectural patterns.

| Feature | Native Cloud Run Domain Mapping | Global Application LB (GEALB) | Firebase Hosting |
| :---- | :---- | :---- | :---- |
| **Production Ready?** | No (Preview, Limited Availability) 5 | **Yes (Recommended)** 5 | Yes |
| **SSL Type** | Google-managed 4 | Google-managed, Self-managed | Google-managed \[11\] |
| **Wildcard SSL Support** | No 5 | **Yes** | Yes (via Firebase) |
| **Global CDN Integration** | No | **Yes (Cloud CDN)** 7 | Yes (Built-in) 7 |
| **Google Cloud Armor (WAF)** | No | **Yes** 7 | No (Basic DDoS protection) |
| **Identity-Aware Proxy (IAP)** | No | **Yes** 7 | No (Use Firebase Auth) |
| **Cost Model** | No Cost (Feature of Cloud Run) | Per-rule, per-GB, static IP 7 | Free tier, then per-GB 7 |
| **Setup Complexity** | Low | High | Very Low |

## **Section 2: Phase 1: Prerequisite \- Domain Ownership Verification**

### **2.1 The "Why": Proving Control to Google**

Before Google will provision an SSL certificate or serve traffic from a custom domain, the user must prove to Google that they control that domain.12 This is a critical security step to prevent unauthorized domain use.

This verification process is handled via the **Google Search Console** (formerly Webmaster Tools) and is a distinct, one-time action separate from the *mapping* process.13 Verification establishes ownership of a base domain. For example, a user must verify ownership of example.com even if their ultimate goal is to map a subdomain like www.example.com or api.example.com.12 Once verified, the domain is available for use across multiple Google Cloud services.

### **2.2 Implementation: Verifying via DNS TXT Record**

The most common and robust method for domain verification is by adding a DNS TXT record.14

1. **Step 1:** In the Google Cloud Console, navigate to the Cloud Run "Domain mappings" page. When adding a new mapping, select the option to "Verify a new domain".5  
2. **Step 2:** This action will redirect to the Google Search Console. In the Search Console interface, add a new "Domain property" (e.g., example.com).14  
3. **Step 3:** Search Console will provide a unique verification string, typically formatted as google-site-verification=....14  
4. **Step 4:** Log in to the administrative console of the domain's DNS provider (e.g., GoDaddy, Cloudflare, or Google Cloud DNS).  
5. **Step 5:** Create a new TXT record for the apex (root) domain. The **Host** (or **Name**) will be @ or the bare domain example.com. The **Value** will be the entire google-site-verification=... string provided by Search Console.14  
6. **Step 6:** Save the record and wait for it to propagate. This can take anywhere from a few minutes to several hours.  
7. **Step 7:** Return to the Google Search Console and click the "Verify" button. Once Google's systems detect the new TXT record, the domain is marked as verified.14

## **Section 3: Core Implementation Flow: DNS Managed Externally (e.g., GoDaddy, Cloudflare)**

This section details the complete workflow for a user whose DNS is managed by a third-party provider, using the **Native Cloud Run Domain Mapping** feature.

### **3.1 Creating the DomainMapping Resource**

1. **Step 1:** Ensure the Cloud Run service that will receive the traffic is successfully deployed.16  
2. **Step 2:** In the GCP Console, navigate to the Cloud Run page and select the "Manage custom domains" (or "Domain mappings") tab.5  
3. **Step 3:** Click "Add Mapping".12  
4. **Step 4:** From the dropdown menu, select the target Cloud Run service.5  
5. **Step 5:** Select the previously verified domain (from Section 2\) and specify the exact hostname to be mapped (e.g., the apex example.com or the subdomain www.example.com).  
6. **Step 6:** Click "Continue." The console will now process the request and display a list of required DNS records. This is the critical handoff. The records displayed will differ depending on whether an apex domain or a subdomain is being mapped.

### **3.2 Scenario A: Mapping an Apex Domain (example.com)**

* **DNS Theory:** According to DNS specifications (RFC 1034), the "naked" or "apex" domain (e.g., example.com without a www or other prefix) cannot use a CNAME record if it also has other records, such as MX (mail) or NS (name server) records. To circumvent this, Google provides static IP addresses.  
* **Required Records:** The Cloud Run UI will display a list of A and AAAA records.5 This is typically a set of four IPv4 addresses (A records) and four IPv6 addresses (AAAA records) to ensure high availability and reliability.17  
* **Implementation Guide (External Registrar):**  
  1. Log in to the external DNS provider's management console.  
  2. Navigate to the DNS record settings for the example.com zone.  
  3. Create the four A records. For each, the **Host** (or **Name**) will be @ (which signifies the apex domain).5 The **Value** (or **Points To**) will be one of the provided IPv4 addresses. Set a reasonable TTL (Time-To-Live), such as 3600 seconds (1 hour).18  
  4. Repeat this process for the other three IPv4 addresses.17  
  5. Create the four AAAA records, again setting the **Host** to @, and using the provided IPv6 addresses as the **Value**.17

### **3.3 Scenario B: Mapping a Subdomain (www.example.com)**

* **DNS Theory:** Subdomains (such as www, api, or shop) are not subject to the same CNAME restrictions as apex domains. The best practice is to use a CNAME (Canonical Name) record. This allows Google to manage and change the underlying IP addresses of its GFE infrastructure over time without requiring any future DNS changes from the user.  
* **Required Record:** The Cloud Run UI will display a single CNAME record.5  
* **Implementation Guide (External Registrar):**  
  1. Log in to the external DNS provider's management console.  
  2. Navigate to the DNS record settings for the example.com zone.  
  3. Create a new CNAME record.  
  4. Set the **Host** (or **Name**) to the desired subdomain (e.g., www).5  
  5. Set the **Value** (or **Points To**) to the canonical name provided by Google: ghs.googlehosted.com..18 The trailing dot is important as it signifies a Fully Qualified Domain Name (FQDN).

### **3.4 Phase 3: The Automated SSL Provisioning Sequence**

After the DNS records have been added at the registrar, return to the Cloud Run UI and click "Done".5 The domain mapping will now appear in the list with a "Pending" or "Provisioning certificate" status.1

The following automated sequence begins:

1. **DNS Propagation:** The world's DNS resolvers must first learn about the new records. This propagation delay is dependent on the DNS provider and the TTL of any previous records. It can take minutes, or in some cases, up to 72 hours, though a few hours is typical.21  
2. **Google DNS Validation:** Google's systems will periodically poll the public DNS for the new records. Once it can verify that the hostname (example.com or www.example.com) correctly resolves to Google's infrastructure (either the A/AAAA IPs or the ghs.googlehosted.com CNAME), it proceeds to the next step.  
3. **SSL HTTP-01 Challenge:** As identified in Section 1.2, Google's chosen CA (GTS or Let's Encrypt) 4 performs an ACME HTTP-01 challenge to prove domain control.  
4. **Challenge-Response:** The Google Front End, which is now configured to accept traffic for the domain, intercepts this unencrypted HTTP challenge and provides the correct cryptographic response.  
5. **Certificate Issuance:** Upon successful validation, the CA issues the SSL/TLS certificate.  
6. **Binding and Activation:** Google's infrastructure binds this new certificate to the frontend serving the Cloud Run service. The status in the Cloud Run console will update to "Active".21

This end-to-end process (steps 2-6) typically takes about 15 minutes *after* the DNS changes have fully propagated, but the documentation advises it can take up to 24 hours.1

## **Section 4: Core Implementation Flow: DNS Managed via Google Cloud DNS**

This section addresses the scenario where the domain's DNS is authoritatively managed by Google's own Cloud DNS service.

### **4.1 Prerequisites**

Before starting, two conditions must be met:

1. A Public Managed Zone must exist within the Google Cloud DNS service for the domain (e.g., example.com).22  
2. The domain's registrar (which could be Google Domains or a third party like GoDaddy) must be configured to use the Google Cloud DNS name servers (e.g., ns-cloud-a1.googledomains.com, ns-cloud-a2.googledomains.com, etc.).24

### **4.2 The Workflow: Expectation vs. Reality**

A common expectation is that because both Cloud Run and Cloud DNS are Google Cloud services, the domain mapping process will be fully automated. The assumption is that creating a DomainMapping in Cloud Run would automatically create the necessary A, AAAA, or CNAME records in the corresponding Cloud DNS zone.

The available documentation and service behavior indicate **this is not the case**. The Native Cloud Run Domain Mapping feature is not deeply integrated with Cloud DNS. The official documentation explicitly states, "If you're using Cloud DNS as your DNS provider, see Adding a record".5 This link directs to the standard, *manual* process for creating DNS records in a Cloud DNS zone. Other tutorials confirm this manual workflow.22 The desire for such automation is a common request 25, which further implies its current absence.

Therefore, Cloud DNS is treated by the Cloud Run mapping feature as just another external DNS provider. The only tangible benefits of using Cloud DNS in this context are consolidated billing, a single console for management, and potentially faster DNS propagation within Google's own network, which may slightly accelerate the validation step.

### **4.3 Implementation Guide (Cloud DNS)**

The workflow is identical to the external DNS flow (Section 3.1) up to the point where the Cloud Run UI displays the required DNS records.

1. **Step 1:** In the Cloud Run "Add Mapping" UI, note the required A/AAAA or CNAME records.  
2. **Step 2:** In a separate browser tab or window, navigate within the GCP Console to "Network Services" \-\> "Cloud DNS".22  
3. **Step 3:** Click on the name of the managed zone for the domain (e.g., example.com).  
4. **Step 4 (For Apex Domain \- example.com):**  
   * Click "Add Standard".22  
   * For "DNS Name," leave it blank (this defaults to the apex).  
   * For "Resource Record Type," select A.  
   * In "IPv4 Address," enter the first IP address provided by the Cloud Run UI. Click "Add item" to add the other three IPs.22  
   * Click "Create."  
   * Repeat this process, selecting AAAA for the "Resource Record Type" and adding the four IPv6 addresses.  
5. **Step 4 (For Subdomain \- www.example.com):**  
   * Click "Add Standard".22  
   * For "DNS Name," enter www.22  
   * For "Resource Record Type," select CNAME.22  
   * For "Canonical name," enter ghs.googlehosted.com..22  
   * Click "Create."  
6. **Step 5:** Return to the Cloud Run "Domain mappings" UI and click "Done."

### **4.4 SSL Provisioning**

The subsequent SSL provisioning sequence is **absolutely identical** to the external DNS flow described in Section 3.4. Google's systems will wait for its *own* Cloud DNS records to propagate, then initiate the same HTTP-01 challenge to validate control and issue the certificate.

## **Section 5: Advanced Scenario: Conflicts with Proxies and CDNs (e.g., Cloudflare)**

One of the most common and difficult-to-diagnose failure modes occurs when a proxy or CDN, such as Cloudflare, is placed in front of a Cloud Run service that uses native domain mapping.

### **5.1 The Root Cause of SSL Provisioning and Renewal Failures**

A service like Cloudflare, when "proxied" (i.e., the "orange cloud" is active), breaks the automated SSL process in two distinct ways:

1. **DNS Obfuscation:** When proxied, a public DNS query for www.example.com no longer returns the ghs.googlehosted.com CNAME. Instead, it returns Cloudflare's IP addresses. This can cause Google's initial DNS validation (Step 2 in Section 3.4) to fail, as the record does not match what it expects.  
2. **HTTP-01 Challenge Interception:** This is the more insidious problem. Common Cloudflare features like "Always Use HTTPS" or "Automatic HTTPS Rewrites" intercept the *unencrypted* HTTP-01 challenge from the CA and issue a 301 or 302 redirect to the HTTPS version of the URL.5 The ACME protocol standard requires that the challenge be served over HTTP to be valid. This redirection causes the challenge to fail, and the certificate cannot be issued.

### **5.2 The Vicious Renewal Cycle**

This issue often manifests in a catastrophic way. A user might successfully set up their domain mapping *before* enabling the Cloudflare proxy. The initial setup works, the certificate is provisioned, and the status becomes ACTIVE.6 The user then enables the Cloudflare proxy ("orange cloud") to gain CDN and WAF benefits. The site runs perfectly for 60-80 days.

However, Google-managed certificates attempt to automatically renew approximately 30 days before their expiration.3 This renewal process uses the *exact same* HTTP-01 challenge mechanism. But now, Cloudflare is active and intercepting the HTTP challenge, causing the renewal to fail. The console status will change to RENEWAL\_FAILED.2 When the original certificate finally expires, the custom domain will begin serving SSL errors, and the service will become inaccessible.

### **5.3 Correct Configuration Guide for Proxies**

To prevent both initial provisioning failure and subsequent renewal failures, an explicit exception must be created for the ACME challenge path.

1. **Step 1:** In the proxy/CDN provider's dashboard (e.g., Cloudflare), navigate to the "Page Rules" (or equivalent) section.6  
2. **Step 2:** Create a new rule with the highest priority.  
   * **URL Match:** \*.example.com/.well-known/acme-challenge/\* 6  
   * **Settings:**  
     * SSL: **Off** 6  
     * Automatic HTTPS Rewrites: **Off** 6  
3. **Step 3:** Create a *second* Page Rule with a lower priority to secure all other site traffic.  
   * **URL Match:** \*.example.com/\*  
   * **Setting:** Always Use HTTPS 6  
4. **Step 4:** In the Cloudflare "SSL/TLS" \-\> "Edge Certificates" tab, *disable* the global "Always use HTTPS" toggle.5 This is critical, as the Page Rules will now handle this logic correctly, allowing the unencrypted challenge to pass through while forcing encryption for all user-facing traffic.

## **Section 6: Troubleshooting and Diagnostics**

### **6.1 Analyzing Mapping and Certificate Statuses**

The status of the domain mapping in the GCP Console is the primary diagnostic tool.

* PROVISIONING 21: Normal state during initial setup or renewal. Google is actively working to validate DNS and provision the certificate. This can take up to 24 hours.1  
* ACTIVE 21: Success. The domain is correctly mapped, and a valid SSL certificate is bound and serving traffic.  
* PROVISIONING\_FAILED 21: A persistent failure. Google's retries have failed. This is almost always due to incorrect DNS records that never propagated or a configuration (like a proxy) that is actively blocking the challenge.21  
* RENEWAL\_FAILED 2: The certificate was once ACTIVE but could not be renewed. This strongly implies a configuration change *after* the initial setup, such as a new proxy, a firewall rule, or a change to the domain's CAA records.

### **6.2 Common Problem: Stuck in "Certificate Provisioning Pending"**

If the status remains "Pending" or PROVISIONING for more than an hour, it is likely stuck.1

* **Primary Cause:** Incorrect DNS records or slow DNS propagation.26  
* **Diagnostic Steps:**  
  1. **Wait:** First, wait at least one hour. DNS propagation is not instant.21  
  2. **Verify DNS Externally:** Use a third-party tool to check what the world sees. Do not rely on your local machine, which may have cached records.  
     * **Google Admin Toolbox (Dig):** Use this tool to query for your A/AAAA or CNAME records.1  
     * **MxToolbox:** This tool can check DNS propagation from multiple locations around the world.26  
  3. **Check for Typos:** The most common errors are mistyping ghs.googlehosted.com, using the wrong IP addresses, or creating an A record for www when a CNAME is required.  
* **Remediation:**  
  1. Delete the Domain Mapping in the Cloud Run UI.26  
  2. Go to the DNS provider and *fix* the incorrect record(s).  
  3. Wait for the *correct* records to propagate (verify with an external dig tool).  
  4. Re-add the Domain Mapping in the Cloud Run UI. This forces the validation process to restart from a clean state.26

### **6.3 Common Problem: "PROVISIONING\_FAILED" or "RENEWAL\_FAILED"**

* **Cause 1 (Proxy):** The most likely cause is a newly added proxy or a feature like "Always Use HTTPS" that is blocking the HTTP-01 challenge. **See Section 5 for the solution**.5  
* **Cause 2 (CAA Records):** The domain's DNS may have a CAA (Certificate Authority Authorization) record that restricts which CAs can issue certificates. If this record does not permit pki.goog (for Google Trust Services) and letsencrypt.org (for Let's Encrypt), the issuance will fail.  
  * **Diagnostic:** Use a dig command: dig example.com CAA  
  * **Remediation:** Add CAA records to explicitly permit Google's CAs, or remove the overly restrictive CAA record.3  
* **Cause 3 (DNS Record Changed):** The required A/AAAA/CNAME records were altered or deleted after the initial setup.

### **6.4 Table 2: Common Domain Mapping Error States and Remediation**

| Error Status | Likely Cause(s) | Diagnostic Command(s) | Remediation Steps |
| :---- | :---- | :---- | :---- |
| **PROVISIONING** (Stuck \> 24h) | 1\. Incorrect DNS records. 2\. Slow DNS propagation. | dig \<domain\> A dig \<domain\> AAAA dig \<subdomain\> CNAME (Use Google Admin Toolbox 1) | 1\. Delete mapping.26 2\. Fix DNS records. 3\. Wait for propagation. 4\. Re-add mapping.26 |
| **PROVISIONING\_FAILED** | 1\. Persistently incorrect DNS. 2\. Proxy (e.g., Cloudflare) blocking HTTP-01 challenge.5 | dig \<domain\> CNAME (check for proxy IPs) curl \-v http://\<domain\>/.well-known/acme-challenge/test (check for redirects) | 1\. See steps for PROVISIONING (Stuck). 2\. Configure proxy exceptions (See Section 5.3).6 |
| **RENEWAL\_FAILED** | 1\. Proxy/CDN was added *after* setup.6 2\. Restrictive CAA record added.3 3\. DNS records were changed. | dig \<domain\> CAA curl \-v http://\<domain\>/.well-known/acme-challenge/test | 1\. Configure proxy exceptions (See Section 5.3). 2\. Add CAA records for pki.goog and letsencrypt.org. 3\. Restore correct DNS records. |

## **Section 7: Expert Recommendations and Final Architectural Considerations**

### **7.1 The Case Against Native Domain Mapping for Production**

The "Native Cloud Run Domain Mapping" feature, while simple, carries significant risks for production environments.

* Its official status as "Preview" and "Limited availability" makes it ineligible for most enterprise-grade services with SLAs.5  
* The lack of wildcard certificate support (\*.example.com) is a major limitation for applications that serve dynamic subdomains.5  
* The fragility of the HTTP-01 challenge mechanism, which is easily broken by common and recommended CDN/WAF proxy configurations, creates an unacceptable operational risk of sudden, hard-to-diagnose outages, especially during the automated renewal cycle.6  
* The lack of integration with Cloud DNS demonstrates it is not a fully mature, "batteries-included" Google Cloud feature.5

### **7.2 Authoritative Recommendation: Use a Global Application Load Balancer**

For any serious production workload, the **Global Application Load Balancer (GEALB) with a Serverless NEG backend** is the correct and recommended architecture.5

* **Superior SSL Management:** The GEALB natively supports Google-managed certificates, **including wildcards**. It also allows operators to upload and manage their own custom (Organization Validated or Extended Validation) certificates. Its certificate provisioning is more robust.  
* **Production-Grade Security:** This is the only pattern that allows the service to be protected by **Google Cloud Armor (WAF)**, providing essential defense against DDoS and OWASP Top 10 web attacks.7  
* **Performance and Identity:** The GEALB pattern seamlessly integrates with **Cloud CDN** for caching and **Identity-Aware Proxy (IAP)** for zero-trust, user-based authentication.7

### **7.3 Final Verdict: A Tool for the Right Job**

The choice of domain mapping architecture should be driven by the use case.

* **Use Native Domain Mapping** for: Personal projects, internal tools, and non-critical development/testing environments where simplicity and zero cost are the primary drivers.7  
* **Use Firebase Hosting** for: Static sites, Jamstack applications, and simple web applications or APIs that need a simple, free, and globally distributed CDN without the complexity of a full load balancer.7  
* **Use a Global Application Load Balancer** for: All production, enterprise-grade, or revenue-generating applications that require high-availability, robust security (WAF), and advanced features like wildcard certificates, CDN, or IAP.5

#### **Works cited**

1. Troubleshoot Cloud Run issues | Google Cloud Documentation, accessed on November 5, 2025, [https://docs.cloud.google.com/run/docs/troubleshooting](https://docs.cloud.google.com/run/docs/troubleshooting)  
2. Google Cloud Run \+ CloudFlare \- Auto Cert Renewal, accessed on November 5, 2025, [https://discuss.google.dev/t/google-cloud-run-cloudflare-auto-cert-renewal/177609](https://discuss.google.dev/t/google-cloud-run-cloudflare-auto-cert-renewal/177609)  
3. Use Google-managed SSL certificates | Cloud Load Balancing, accessed on November 5, 2025, [https://docs.cloud.google.com/load-balancing/docs/ssl-certificates/google-managed-certs](https://docs.cloud.google.com/load-balancing/docs/ssl-certificates/google-managed-certs)  
4. Securing custom domains with SSL | App Engine standard environment \- Google Cloud, accessed on November 5, 2025, [https://cloud.google.com/appengine/docs/standard/securing-custom-domains-with-ssl](https://cloud.google.com/appengine/docs/standard/securing-custom-domains-with-ssl)  
5. Mapping custom domains | Cloud Run | Google Cloud Documentation, accessed on November 5, 2025, [https://docs.cloud.google.com/run/docs/mapping-custom-domains](https://docs.cloud.google.com/run/docs/mapping-custom-domains)  
6. Cloud Run Custom Domain with Cloudflare : r/googlecloud \- Reddit, accessed on November 5, 2025, [https://www.reddit.com/r/googlecloud/comments/kvj2ss/cloud\_run\_custom\_domain\_with\_cloudflare/](https://www.reddit.com/r/googlecloud/comments/kvj2ss/cloud_run_custom_domain_with_cloudflare/)  
7. Any easy way to set up a custom domain for a cloud run service? \- Reddit, accessed on November 5, 2025, [https://www.reddit.com/r/googlecloud/comments/1f9icea/any\_easy\_way\_to\_set\_up\_a\_custom\_domain\_for\_a/](https://www.reddit.com/r/googlecloud/comments/1f9icea/any_easy_way_to_set_up_a_custom_domain_for_a/)  
8. GTM Server-side (sGTM) \- How to map a custom domain now (after Cloud Run Integrations end)? : r/googlecloud \- Reddit, accessed on November 5, 2025, [https://www.reddit.com/r/googlecloud/comments/1k9zpoc/gtm\_serverside\_sgtm\_how\_to\_map\_a\_custom\_domain/](https://www.reddit.com/r/googlecloud/comments/1k9zpoc/gtm_serverside_sgtm_how_to_map_a_custom_domain/)  
9. Set up a global external Application Load Balancer with Cloud Run, App Engine, or Cloud Run functions \- Google Cloud Documentation, accessed on November 5, 2025, [https://docs.cloud.google.com/load-balancing/docs/https/setup-global-ext-https-serverless](https://docs.cloud.google.com/load-balancing/docs/https/setup-global-ext-https-serverless)  
10. Google cloud run domain mapping and custom domains-google cloud load balancing can it have two same domain name but in only one cloud run service \- Stack Overflow, accessed on November 5, 2025, [https://stackoverflow.com/questions/77927656/google-cloud-run-domain-mapping-and-custom-domains-google-cloud-load-balancing-c](https://stackoverflow.com/questions/77927656/google-cloud-run-domain-mapping-and-custom-domains-google-cloud-load-balancing-c)  
11. How to map Google Domains domain name to Google Cloud Run project ? I can't make 'www' work \- Server Fault, accessed on November 5, 2025, [https://serverfault.com/questions/1053154/how-to-map-google-domains-domain-name-to-google-cloud-run-project-i-cant-make](https://serverfault.com/questions/1053154/how-to-map-google-domains-domain-name-to-google-cloud-run-project-i-cant-make)  
12. Mapping Custom Domains | App Engine standard environment for Go 1.11 docs, accessed on November 5, 2025, [https://cloud.google.com/appengine/docs/legacy/standard/go111/mapping-custom-domains](https://cloud.google.com/appengine/docs/legacy/standard/go111/mapping-custom-domains)  
13. Verify your site ownership \- Search Console Help \- Google Help, accessed on November 5, 2025, [https://support.google.com/webmasters/answer/9008080](https://support.google.com/webmasters/answer/9008080)  
14. Verifying your domain | Cloud Identity, accessed on November 5, 2025, [https://cloud.google.com/identity/docs/verify-domain](https://cloud.google.com/identity/docs/verify-domain)  
15. Create a Cloud Run service for a custom domain, accessed on November 5, 2025, [https://cloud.google.com/run/docs/samples/cloudrun-custom-domain-mapping-run-service](https://cloud.google.com/run/docs/samples/cloudrun-custom-domain-mapping-run-service)  
16. How to Set Up a Custom Domain for Your Google Cloud Run service \- David Muraya, accessed on November 5, 2025, [https://davidmuraya.com/blog/custom-domain-google-cloud-run/](https://davidmuraya.com/blog/custom-domain-google-cloud-run/)  
17. Boost Your Website's Performance with Cloud Run Autoscaling | by Rubens Zimbres, accessed on November 5, 2025, [https://medium.com/@rubenszimbres/boost-your-websites-performance-with-cloud-run-autoscaling-844e0b7e4e63](https://medium.com/@rubenszimbres/boost-your-websites-performance-with-cloud-run-autoscaling-844e0b7e4e63)  
18. How to Map a Custom Domain to Google Cloud Run Service \- YouTube, accessed on November 5, 2025, [https://www.youtube.com/watch?v=lDtvpUYAFzA](https://www.youtube.com/watch?v=lDtvpUYAFzA)  
19. Troubleshooting and escalating a systemic failure in Google Cloud's automated domain mapping and SSL certificate provisioning process for a Cloud Run service \- Serverless Applications, accessed on November 5, 2025, [https://discuss.google.dev/t/troubleshooting-and-escalating-a-systemic-failure-in-google-clouds-automated-domain-mapping-and-ssl-certificate-provisioning-process-for-a-cloud-run-service/274378](https://discuss.google.dev/t/troubleshooting-and-escalating-a-systemic-failure-in-google-clouds-automated-domain-mapping-and-ssl-certificate-provisioning-process-for-a-cloud-run-service/274378)  
20. Troubleshoot SSL certificates | Cloud Load Balancing, accessed on November 5, 2025, [https://docs.cloud.google.com/load-balancing/docs/ssl-certificates/troubleshooting](https://docs.cloud.google.com/load-balancing/docs/ssl-certificates/troubleshooting)  
21. Quickstart: Set up DNS records for a domain name with Cloud DNS, accessed on November 5, 2025, [https://docs.cloud.google.com/dns/docs/set-up-dns-records-domain-name](https://docs.cloud.google.com/dns/docs/set-up-dns-records-domain-name)  
22. Cloud DNS, accessed on November 5, 2025, [https://cloud.google.com/dns](https://cloud.google.com/dns)  
23. Tutorial: Set up a domain by using Cloud DNS, accessed on November 5, 2025, [https://docs.cloud.google.com/dns/docs/tutorials/create-domain-tutorial](https://docs.cloud.google.com/dns/docs/tutorials/create-domain-tutorial)  
24. How to automatically create a DNS zone and DNS records in google cloud for a CLoud Run hosted API using terraform \- Stack Overflow, accessed on November 5, 2025, [https://stackoverflow.com/questions/76503912/how-to-automatically-create-a-dns-zone-and-dns-records-in-google-cloud-for-a-clo](https://stackoverflow.com/questions/76503912/how-to-automatically-create-a-dns-zone-and-dns-records-in-google-cloud-for-a-clo)  
25. Google Cloud Run \- Domain Mapping stuck at Certificate Provisioning \- Stack Overflow, accessed on November 5, 2025, [https://stackoverflow.com/questions/57789565/google-cloud-run-domain-mapping-stuck-at-certificate-provisioning](https://stackoverflow.com/questions/57789565/google-cloud-run-domain-mapping-stuck-at-certificate-provisioning)
