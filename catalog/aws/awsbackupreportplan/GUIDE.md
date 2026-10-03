# AwsBackupReportPlan — Kind Guide

Authored operational judgment for the Backup Audit Manager report plan
component: the design decisions behind the spec's shape, and what to
know before operating report plans in production.

## Design decisions

- **Its own kind, not a framework arm.** A report plan references
  frameworks MANY-TO-MANY (`framework_arns` accepts ARNs from any
  number of framework components) and the job-report templates need no
  framework at all — folding it under one framework would misrepresent
  both facts.
- **The AWS name is an explicit spec field** (`report_plan_name`): AWS
  forbids hyphens in report plan names, stricter than metadata.name
  conventions.
- **`report_template` replacement is taught loudly**: the provider
  marks it ForceNew from INSIDE the nested block — an easy-to-miss
  full-resource replacement the spec comment surfaces.
- **`number_of_frameworks` mirrors the provider's zero-sentinel**: it
  is transmitted only when positive; AWS computes it otherwise.

## Operating report plans in production

- **The bucket policy is the failure point.** Reports are written by
  the AWS Backup report service — the bucket needs a policy granting
  it `s3:PutObject` (AWS documents the exact statement in the Backup
  Audit Manager guide). A missing grant fails report JOBS, not the
  report plan deploy: check `aws backup list-report-jobs` after the
  first cycle.
- **Compliance templates need frameworks in the same region**;
  cross-region evidence takes one report plan per region plus the
  `regions` coverage list.
- **Reports run daily on AWS's schedule** — there is no cron to tune;
  `aws backup start-report-job` produces one on demand.
- **CSV and JSON can both be delivered** — pick both formats when both
  humans and pipelines consume the evidence.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
