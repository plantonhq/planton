# Encrypted Invoice Parser

## Use Case

Parse supplier invoices (vendor, dates, totals, line items) in Europe under your own encryption key, on a pinned model version, protected from accidental destroy.

## When to Use

- Accounts-payable automation
- Regulated finance data that must be encrypted under a customer-managed key
- Pipelines whose downstream systems depend on stable extraction output

## What This Creates

- An invoice parser in the `eu` multi-region, encrypted with a `GcpKmsKey`, serving a pinned pretrained version by default, with `PREVENT` on destroy

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `defaultVersion` | a pinned pretrained release | The version Google lists for your processor; upgrade by changing the id. |
| `kmsKeyName` | `documents-key` reference | Your key in a compatible location. |
| `type` | `INVOICE_PROCESSOR` | `EXPENSE_PROCESSOR` for receipts, other specialized parsers. |
