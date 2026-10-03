# Auth0 Prompt Screen Partials

Inserts your own HTML fragments on one Universal Login prompt's screens -- extra form fields, consent checkboxes, a notice above the form -- at the insertion points each screen offers. One Infra Component per prompt. Partials render only inside a page template on a custom domain.

## What Gets Created

When you deploy this Infra Component, the IaC module sets every partial of one prompt on the tenant your Auth0 connection's credential belongs to:

- **The fragments of each screen you declare** -- each at its insertion point: the start or end of the form, the form's footer, around the secondary actions, or in place of the form's fields
- **Nothing anywhere else** -- a screen or insertion point you leave out renders nothing

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `read:prompts` and `update:prompts` on the tenant's Management API (Auth0 dashboard: Applications, APIs, Auth0 Management API, Machine To Machine Applications).
- **A custom domain** on the tenant (the Auth0 Custom Domain Infra Component, verified).
- **A Universal Login page template** (the Auth0 Branding Infra Component's `universalLoginTemplate`), which partials render inside.
- **The Universal Login experience "new"** (the Auth0 Prompt Infra Component).

## Deploy

### Console

Open the deployment store, find **Auth0 Prompt Screen Partials**, and click **Deploy**. Start from the **Terms on Signup** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0PromptScreenPartials
metadata:
  name: signup-partials
  org: acme-corp
  env: prod
spec:
  promptType: signup
  screenPartials:
    - screenName: signup
      insertionPoints:
        formContentEnd: <div class="terms">By signing up you accept the terms.</div>
```

```shell
planton apply -f auth0-prompt-screen-partials.yaml
```

The sign-up screen shows your note below its fields. An Infra Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Screens and insertion points come from Auth0's reference** -- Each prompt has one or more screens, and each screen offers the form, footer and secondary-action insertion points; the secondary actions exist only when a social or enterprise connection is enabled. Look them up in Auth0's [customization reference](https://auth0.com/docs/customize/login-pages/universal-login/customize-signup-and-login-prompts).

**This resource owns the prompt's whole set** -- Auth0 replaces every partial of the prompt on every write. Declare every screen of the prompt in one resource, one entry per screen.

**Fields reach Actions** -- A form field named `ulp-...` is submitted with the form and read by Actions on `event.request.body`: pre-user-registration for sign-up, post-login for login.

**The prompt is fixed** -- `promptType` is the set's identity. Declare another resource for another prompt.

**Destroy removes every partial of the prompt** -- The screens render without them, and a field an Action relies on stops being submitted.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The tenant is the one the Auth0 connection's credential belongs to.

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `prompt_type` | The prompt whose screens the partials extend | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Terms on signup** -- A required terms-of-service checkbox on the sign-up screen, read by an Action. Start from the **Terms on Signup** preset.

**Notice above login** -- A notice at the top of the login form. Start from the **Notice Above Login** preset.

## Works With

- [**Auth0 Branding**](/infra-catalog/auth0-branding) -- the page template partials render inside.
- [**Auth0 Custom Domain**](/infra-catalog/auth0-custom-domain) -- the custom domain a page template needs.
- [**Auth0 Prompt Custom Text**](/infra-catalog/auth0-prompt-custom-text) -- `var-<name>` words a partial reads, per language.
- [**Auth0 Action**](/infra-catalog/auth0-action) -- reads the fields a partial adds.
- [**Auth0 Prompt**](/infra-catalog/auth0-prompt) -- the Universal Login experience partials need, and the flow that decides which prompts people see.
