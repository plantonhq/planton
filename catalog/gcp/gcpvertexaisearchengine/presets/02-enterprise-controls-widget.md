# Enterprise Controls and Widget

## Use Case

A complete documentation search app: the Enterprise tier with generated answers, five serving controls that shape results (synonyms, a boost for the latest release, a filter hiding archived pages, a promoted quickstart, a redirect for status queries) wired through the serving config, and Google's embeddable widget on the public docs site with a conversational answer above the results, branded and encrypted under your own key.

## When to Use

- A public documentation or support site that wants Google-quality search with answers
- When editors need to tune results without redeploying an application
- When the search UI should be Google's widget, not custom code

## What This Creates

- A SEARCH engine in `global` over the referenced website store, `SEARCH_TIER_ENTERPRISE` with `SEARCH_ADD_ON_LLM`, encrypted under a referenced `GcpKmsKey`, with `PREVENT` as its destroy policy
- Five controls (synonyms, boost, filter, promote, redirect), each keyed by id and applied by the engine's default serving config
- The default widget config PATCHed for public access from `docs.example.com`, follow-up questions, extractive answers, autocomplete, safe search, quality feedback, and a tuned answer prompt

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `controls[]` | five examples | Your synonyms, boosts, and filters; every id you list in `servingConfig` must be a control here with the matching action. |
| `widgetConfig.accessSettings.allowPublicAccess` | `true` | `false` for an internal site; add `workforceIdentityPoolProvider` to sign users in through your identity provider. |
| `widgetConfig.uiSettings.interactionType` | follow-ups | `SEARCH_WITH_ANSWER` for one answer, `SEARCH_ONLY` for results alone (and no LLM add-on charges). |
| `widgetConfig.uiSettings.generativeAnswerConfig.modelPromptPreamble` | Acme's | Your voice and citation rules for the generated answer. |
| `kmsKeyName` | a `GcpKmsKey` reference | Remove for Google-managed encryption. |

The widget config is Google's own: this block PATCHes it and destroy leaves it in place. Enterprise-tier queries and generated answers bill per thousand at higher rates than Standard search.
