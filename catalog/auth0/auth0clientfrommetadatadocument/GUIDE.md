# Auth0 Client From Metadata Document Guide

This guide protects you from treating a document-registered application like one you own: the document, not the manifest, decides who the application is, and Auth0 reads it only when you ask.

## Security
## Platform Security Posture

The certifications below are Auth0's own published claims about their hosted platform (verify current status on Auth0's compliance page). They describe the vendor's service — never this Planton component, and never your deployment: configuring this resource does not make your application certified, authorized, or compliant with any framework.

Auth0's published certifications and security standards:

- SOC 2 Type II (annual audit)
- ISO 27001, ISO 27018 (privacy controls)
- HIPAA BAA available on enterprise plans
- PCI DSS Level 1 Service Provider
- FedRAMP Authorized (moderate baseline)
- CSA STAR Level 2
- GDPR compliant with Data Processing Agreement

## Data Protection

- **Data residency**: US, EU, AU regions
- **Encryption in transit**: TLS 1.2+
- **Encryption at rest**: AES-256
- **Penetration testing**: Annual third-party assessments

## Document-Registration Security Notes

### Whoever Controls the Document Controls the Application

The application's name on the consent screen, its redirect URIs and its signing keys live in a file on someone else's host. A change there reaches your tenant the next time `externalClientIdVersion` moves -- so bump it only after reviewing what changed (the `validation_warnings` and `callbacks` outputs show what Auth0 took), and treat a version bump in a pull request like a change to the application itself.

### Keys Rotate Under a New kid

For a confidential client (`private_key_jwt`), Auth0 keeps the old key when new material arrives under a `kid` it already holds. The document's owner rotates by publishing a new key with a new `kid`; you then raise `externalClientIdVersion` so Auth0 registers it.

### Keep Redirect Protection On

These applications are third-party by construction. `redirectionPolicy: open_redirect_protection`, Auth0's default for them, is what keeps a failed sign-in from bouncing a person to a callback nobody on your side controls. Declare it so a review sees it.

## Working With Other Auth0 Kinds

### The Order to Apply

1. **Auth0 Tenant Settings** -- with Client ID Metadata Document registration on; without it Auth0 refuses the registration.
2. **Auth0 Connection** -- the connections people sign in through, promoted to the domain level (`isDomainConnection: true`): third-party applications sign in through nothing else.
3. **Auth0 Client From Metadata Document** -- this kind.
4. **The API grant** -- a client grant naming this kind's `client_id` output for each API the client calls; a third-party application reaches no API without one, even an API whose policy allows all applications, and never the Management API.

### Which Id to Use Where

The application has two identifiers. `external_client_id` (the document's URL) is what it sends to `/authorize` and `/oauth/token` and what its access tokens carry as `client_id`. `client_id` (`tpc_...`) is its Management API id, and it is what grants, connections and an import name. Reference `status.outputs.client_id` from other kinds; give `external_client_id` to the APIs that check who is calling.

### Auth0 Client or This Kind

Choose **Auth0 Client** when you own the application and want every setting, a client secret, first-party status or the client-credentials grant. Choose this kind when the application publishes its own document -- an MCP client, a partner integration -- and should stay a third-party application whose identity its owner maintains. The two render as different nodes on the diagram, and that difference is real: this one has no secret to leak and no settings you can make it lie about.

## Traps

### Adopting a Registration Resets What You Did Not Declare

The provider cannot leave some settings alone once Auth0 holds a value for them. On an application registered in the dashboard (or by the client itself) and adopted here, declare the live value of each, or the first apply changes it:

- `allowedOrigins`, `webOrigins`, `clientMetadata`, `organizationDiscoveryMethods`, `skipNonVerifiableCallbackUriConfirmationPrompt` -- reset when unset
- `defaultOrganization`, `tokenQuota` -- removed when unset
- `description`, `requireProofOfPossession` -- proposed for clearing on every plan while Auth0 holds a value; `description` is seeded from the document, so declare it on a fresh registration too

### The Same URL Is the Same Application

Auth0 keys registrations by the document's URL. Two resources with one URL manage one application and fight over it, and destroying either deletes it for both. Two environments that must stay apart need two documents.

### Rules Break Sign-In

A tenant with active Rules fails every login for these applications. Move Rules to Actions before the first registration.

### Settings That Have Nothing to Act On Yet

`tokenQuota` and `defaultOrganization` concern the client-credentials flow, which Auth0 does not give document-registered applications, and `organizationDiscoveryMethods` applies only while the application's organization behavior is the pre-login prompt, which this kind does not set. They are here so an adopted application's values can be declared, not to switch features on.

## Permissions
## Management API Scopes

| Operation | Scope | Description |
|-----------|-------|-------------|
| Register | `create:clients`, `update:clients` | Register from the document; re-register an existing URL or on a version change |
| Read | `read:clients` | Read the application and preview the document on every refresh |
| Update | `update:clients` | Apply the settings the spec declares |
| Delete | `delete:clients` | Delete the application |

## Compliance
## Document-Registration Compliance Notes

### The Registration as Auditable Configuration

Which documents your tenant trusts, when each was last fetched (`externalClientIdVersion`) and the token policy set over it are version-controlled with the rest of the tenant's identity configuration. Auth0's tenant logs name the application by its document URL, a human-readable record of who registered what.

## Cost
## Pricing Model

A registered application is a configuration object; Auth0 bills the tenant by plan and monthly active users, not per application. Registration from a metadata document is on every plan; a confidential client (`private_key_jwt`) needs the Enterprise plan.

## Cost Impact

The people who sign in through the application count toward the tenant's monthly active users. Mutual-TLS proof of possession needs the Enterprise plan with the Highly Regulated Identity add-on, and fine-grained token quotas are an Early Access feature Auth0 enables on request.
