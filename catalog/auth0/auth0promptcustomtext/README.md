# Auth0PromptCustomText

Manages the words one [Universal Login](https://auth0.com/docs/customize/login-pages/universal-login/customize-text-elements) prompt shows in one language: every title, description, button and error message on that prompt's screens, in place of Auth0's defaults.

## When to Use

- **Your product's voice**: "Log in to Planton to continue to Planton Console." instead of Auth0's default sentence, and error messages that say what to do next.
- **Another language**: translate the login and sign-up screens for a language you serve.
- **Words for your partials**: define `var-<name>` keys that the fields added by Auth0 Prompt Screen Partials read, translated per language.
- **Infrastructure as code**: the words people read at sign-in are reviewed like the rest of the product's copy.

## Quick Start

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0PromptCustomText
metadata:
  name: login-en
  org: acme-corp
  env: production
spec:
  prompt: login
  language: en
  screens:
    login:
      texts:
        title: Welcome back
        description: Log in to Acme to continue to ${clientName}.
```

## Fields

| Field | Description |
|---|---|
| `prompt` | The step of the login flow the words belong to (`login`, `login-id`, `signup`, `reset-password`, the multi-factor challenges, ...) |
| `language` | The language the words are shown in, as Auth0 names it (`en`, `fr-CA`, `pt-BR`, `zh-TW`, ...) |
| `screens` | Each screen of the prompt, mapped to its `texts`: a map of text key (`title`, `description`, `buttonText`, `wrong-credentials`, ...) to the words it shows |

The screens each prompt has, and the keys each screen offers, come from Auth0's customization reference: [Prompt values](https://auth0.com/docs/customize/login-pages/universal-login/customize-text-elements#prompt-values). A key may use Universal Login's variables, such as `${clientName}` and `${companyName}`.

## Key Behaviors

- **One resource per prompt and language**: Auth0 stores one custom text per prompt and language, and replaces it whole on every write. This resource owns all of it: a screen or key the spec leaves out shows Auth0's default words, and a key Auth0 adds later shows its default until you declare it.
- **The body is rendered for you**: both modules render `screens` into the provider's one JSON document, `{"<screen>": {"<key>": "<text>"}}`, keys sorted, the same bytes in either engine.
- **Destroy returns the defaults**: destroying the resource writes an empty custom text, so the prompt shows Auth0's default words in that language again.
- **Prompt and language are the identity**: they name the custom text in Auth0 (the import id is `<prompt>::<language>`). Treat them as fixed; to re-word another prompt or language, declare another resource. Changing either in place writes the words to the new pair and leaves the old pair's words as they were.
- **Universal Login only**: custom text applies to the `new` experience (Auth0 Prompt's `universalLoginExperience`); the Classic pages never show it.
- **The language must be enabled**: Universal Login shows a language only when the tenant lists it among its enabled languages.
- **Plans**: custom text is available on every plan, the Free plan included, and needs no custom domain.
- **Permissions**: the credential needs `read:prompts` and `update:prompts` on the tenant's Management API (`iac/permissions.yaml`).

## Outputs

| Output | Description |
|---|---|
| `prompt` | The prompt the words belong to |
| `language` | The language the words are shown in |
| `id` | The custom text's identifier, `<prompt>::<language>` (for example `login::en`), also its import id |

## Auth0 Documentation

- [Customize Universal Login text elements](https://auth0.com/docs/customize/login-pages/universal-login/customize-text-elements)
- [Universal Login internationalization](https://auth0.com/docs/customize/internationalization-and-localization/universal-login-internationalization)
- [Terraform auth0_prompt_custom_text](https://registry.terraform.io/providers/auth0/auth0/latest/docs/resources/prompt_custom_text)
- [Pulumi auth0.PromptCustomText](https://www.pulumi.com/registry/packages/auth0/api-docs/promptcustomtext/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
