# Strict Customer Chat

## Use Case

A screen for a customer-facing assistant: every content category, strict injection detection, personal data found and redacted through your own Sensitive Data Protection templates, malicious links blocked, and a friendly message instead of the blocked text.

## When to Use

- Public chat, support, or sales assistants
- Assistants whose users may type personal or payment data
- Any flow where a harmful answer is a brand or compliance incident

## What This Creates

- A template in the `us` multi-region running all four Responsible AI filters, prompt-injection detection at `LOW_AND_ABOVE`, Sensitive Data Protection through a `customer-pii` inspect template and a `redact-pii` de-identify template, and malicious URL detection
- Blocking enforcement with custom messages, template and sanitize logging, multi-language detection, and Google's stable filter version
- `PREVENT` on destroy, so no apply removes the screen by accident

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `filterConfig.sdpSettings.advancedConfig` | your templates | Point at your own Sensitive Data Protection templates, or use `basicConfig` for Google's detectors. |
| `raiSettings.raiFilters[].confidenceLevel` | per category | Loosen a category your product legitimately discusses. |
| `templateMetadata.custom*ErrorMessage` | generic apology | Your product's voice. |
| `filterVersionSelector` | stable alias | `version: v2` to pin verdicts for audits. |
