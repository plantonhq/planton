# Preset: Basic SQL Workgroup

A minimal Athena workgroup for interactive SQL analytics with query results
stored in S3. All governance defaults apply — configuration enforcement is
enabled, CloudWatch metrics are published.

## What This Configures

- Query results written to the specified S3 location.
- Workgroup configuration enforcement enabled (default).
- CloudWatch metrics enabled (default).
- No cost limits — queries can scan any amount of data.
- Latest Athena engine version (AUTO).

## When to Use

- Development or early-stage analytics teams.
- Ad-hoc query workloads where cost controls aren't yet needed.
- Quick-start setup before adding governance features.

## Customization Points

- Add `bytesScannedCutoffPerQuery` for cost control.
- Add `encryptionOption: SSE_S3` for encrypted query results.
- Change `enforceWorkgroupConfiguration` to `false` for development flexibility.
