# Auth0PromptScreenPartials

Manages the HTML fragments one [Universal Login](https://auth0.com/docs/customize/login-pages/universal-login/customize-signup-and-login-prompts) prompt inserts at its named insertion points: extra form fields, consent checkboxes, a note above the form or below the buttons, on each of the prompt's screens.

## When to Use

- **Collect more at sign-up**: a company name, a terms-of-service checkbox, a marketing opt-in -- submitted with the form and read by Actions.
- **Say something on the login page**: a maintenance notice, which account to use, where to get help.
- **Rebuild a consent form**: replace the customized-consent screen's form with your own.
- **Infrastructure as code**: what the login screens show and collect is reviewed like the rest of the product.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0PromptScreenPartials
metadata:
  name: signup-partials
  org: acme-corp
  env: production
spec:
  promptType: signup
  screenPartials:
    - screenName: signup
      insertionPoints:
        formContentEnd: <div class="terms">By signing up you accept the terms.</div>
```

## Fields

| Field | Description |
|---|---|
| `promptType` | The prompt whose screens the partials extend: `login-id`, `login`, `login-password`, `signup`, `signup-id`, `signup-password`, `login-passwordless`, `customized-consent`, `passkeys`, `confirmation` |
| `screenPartials[].screenName` | One of the prompt's screens (for example `login`, `signup-id`, `login-passwordless-email-code`) |
| `screenPartials[].insertionPoints` | The fragments of the screen: `formContent` (replaces the form's own fields -- only to rebuild the form), `formContentStart`, `formContentEnd`, `formFooterStart`, `formFooterEnd`, `secondaryActionsStart`, `secondaryActionsEnd` |

The screens each prompt has, and the insertion points each screen offers, come from Auth0's customization reference: [Customize Signup and Login Prompts](https://auth0.com/docs/customize/login-pages/universal-login/customize-signup-and-login-prompts). A fragment is HTML, CSS and JavaScript, may use Liquid and every variable of the page template, and is at most 10,000 characters. Form fields named with the `ulp-` prefix are submitted with the form and reach Actions on `event.request.body`.

## Key Behaviors

- **Partials need a page template and a custom domain**: they render only inside a Universal Login page template -- Auth0 Branding's `universal_login_template` -- and a page template needs a custom domain on the tenant (Auth0 Custom Domain). They also need the Universal Login experience `new` (Auth0 Prompt).
- **One resource per prompt, owning its whole set**: Auth0 stores one set of partials per prompt and replaces it whole on every write. A screen or insertion point the spec leaves out renders nothing.
- **Destroy removes every partial of the prompt**: the resource's delete writes an empty set.
- **The prompt is the identity**: `promptType` names the set in Auth0 and is its import id. Treat it as fixed, and declare a second resource for a second prompt.
- **Screen order does not matter**: the modules declare the screens in screen-name order, the order Auth0 answers them in, so listing them differently plans no change.
- **The single-screen resource is composed here**: the provider's one-screen `auth0_prompt_screen_partial` adds or replaces one screen of a prompt; every one of its arguments is one `screenPartials` entry of this kind. Never mix the two on one prompt: this kind replaces the prompt's whole set on every apply, so a one-screen resource beside it is overwritten, and overwrites it in turn.
- **Plans**: partials cost nothing themselves; the custom domain they need is included on every plan (the Free plan has one, once a card is on file).
- **Permissions**: the credential needs `read:prompts` and `update:prompts` on the tenant's Management API (`iac/permissions.yaml`).

## Outputs

| Output | Description |
|---|---|
| `prompt_type` | The prompt whose screens the partials extend |

## Auth0 Documentation

- [Customize signup and login prompts](https://auth0.com/docs/customize/login-pages/universal-login/customize-signup-and-login-prompts)
- [Universal Login page templates](https://auth0.com/docs/customize/login-pages/universal-login/customize-templates)
- [Terraform auth0_prompt_screen_partials](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/prompt_screen_partials)
- [Pulumi auth0.PromptScreenPartials](https://www.pulumi.com/registry/packages/auth0/api-docs/promptscreenpartials/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
