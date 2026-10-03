# Auth0 Prompt Custom Text

Sets the words one Universal Login prompt shows in one language -- every title, description, button and error message on its screens -- in place of Auth0's defaults. One Infra Component per prompt and language.

## What Gets Created

When you deploy this Infra Component, the IaC module sets the custom text of one prompt in one language on the tenant your Auth0 connection's credential belongs to:

- **The words of each screen you declare** -- rendered into the one document Auth0 stores for the prompt and language
- **Auth0's defaults everywhere else** -- a screen or key you leave out shows Auth0's words

## Before You Deploy

### Planton Setup

- **Auth0 Provider Connection** -- an active connection in the Connect module with the tenant's domain, client ID, and client secret. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Auth0 Account

- **The connection's Machine-to-Machine application** must hold `read:prompts` and `update:prompts` on the tenant's Management API (Auth0 dashboard: Applications, APIs, Auth0 Management API, Machine To Machine Applications).
- **The Universal Login experience "new"** on the tenant (the Auth0 Prompt Infra Component); the Classic pages never show custom text.
- **The language enabled** on the tenant (Auth0 dashboard: Settings, Languages), for Universal Login to show it.

## Deploy

### Console

Open the deployment store, find **Auth0 Prompt Custom Text**, and click **Deploy**. Start from the **Login in Your Voice** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: auth0.planton.dev/v1alpha1
kind: Auth0PromptCustomText
metadata:
  name: login-en
  org: acme-corp
  env: prod
spec:
  prompt: login
  language: en
  screens:
    login:
      texts:
        title: Welcome back
        description: Log in to Acme to continue to ${clientName}.
```

```shell
planton apply -f auth0-prompt-custom-text.yaml
```

The login page reads "Welcome back" and your sentence. An Infra Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Screens and keys come from Auth0's reference** -- Each prompt has one or more screens, and each screen a fixed set of text keys. Look them up in Auth0's [customization reference](https://auth0.com/docs/customize/login-pages/universal-login/customize-text-elements#prompt-values). A key may use Universal Login's variables, such as `${clientName}` and `${companyName}`.

**This resource owns the prompt's words in the language** -- Auth0 replaces the whole custom text on every write. Declare every screen of the prompt you re-word in one resource; anything you leave out shows Auth0's default.

**Prompt and language are fixed** -- They are the custom text's identity. Declare another resource for another prompt or language.

**Destroy returns the defaults** -- Destroying the Infra Component returns the prompt to Auth0's default words in that language.

**Identifier-first tenants use other prompts** -- With identifier-first login, people sign in on the `login-id` and `login-password` prompts and sign up on `signup-id` and `signup-password`, not on `login` and `signup`.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The tenant is the one the Auth0 connection's credential belongs to.

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `prompt` | The prompt the words belong to | Audits |
| `language` | The language the words are shown in | Audits |
| `id` | The custom text's identifier, `<prompt>::<language>` | Importing the custom text into another stack |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Login in your voice** -- The login screen's welcome, sentence and error messages in your product's words. Start from the **Login in Your Voice** preset.

**Signup in your voice** -- The sign-up screen in the same voice. Start from the **Signup in Your Voice** preset.

**Another language** -- Copy a resource, change `language`, and translate each text.

## Works With

- [**Auth0 Prompt**](/infra-catalog/auth0-prompt) -- the Universal Login experience custom text needs, and the identifier-first flow that decides which prompts people see.
- [**Auth0 Tenant Settings**](/infra-catalog/auth0-tenant-settings) -- the friendly name `${companyName}` reads.
- [**Auth0 Client (Application)**](/infra-catalog/auth0-client) -- the application name `${clientName}` reads.
- [**Auth0 Prompt Screen Partials**](/infra-catalog/auth0-prompt-screen-partials) -- extra fields whose labels read `var-<name>` keys defined here.
