# Planton Changelog

This directory contains detailed changelog entries for Planton. Each significant change is documented in its own file with the following naming convention:

```
YYYY-MM-DD-brief-description.md
```

## Changelog Format

Each changelog entry should include:

1. **Title** - Brief description of the change
2. **Date** - Date of the change (YYYY-MM-DD)
3. **Type** - Feature, Fix, Breaking Change, Performance, Security, etc.
4. **PR** - Link to the pull request (if applicable)
5. **Summary** - High-level overview of the change
6. **Motivation** - Why the change was made
7. **What's New** - Detailed description of new functionality
8. **Implementation Details** - Technical details for maintainers
9. **Migration Guide** - Steps for users to adopt the change (if applicable)
10. **Examples** - Code snippets demonstrating the change
11. **Benefits** - Value provided to users

## Recent Changes

- [2026-06-16: Secure-by-default secret-field annotation sweep](.changelog/2026-06-16-secure-by-default-annotation-sweep.md) - Annotated 62 real secret fields `sensitive` (secret-by-default, fail-closed against plaintext) and exempted 19 heuristic false positives with an auditable `sensitive_exempt_reason`, raising measured secret-coverage from 3.9% to 81.7% across all providers; 20 real secrets deliberately deferred and tracked in the coverage baseline
- [2025-10-17: ExternalDNS Cloudflare Provider Support](.changelog/2025-10-17-external-dns-cloudflare-support.md) - Added comprehensive Cloudflare provider support to ExternalDNS Kubernetes addon, enabling automated DNS record management with Cloudflare proxy support, API rate limiting optimization, and multi-domain deployment capabilities
- [2025-10-16: Temporal Search Attributes Refactor](.changelog/2025-10-16-temporal-search-attributes-refactor.md) - Refactored Temporal Kubernetes search attributes from enum to validated string using Temporal's official naming convention, eliminating programming language keyword conflicts and improving user experience
- [2025-10-11: Percona PostgreSQL Operator Support](.changelog/2025-10-11-percona-postgresql-operator.md) - Added support for deploying and managing the Percona Distribution for PostgreSQL Operator on Kubernetes clusters with automated deployment, high availability, backup, and disaster recovery features
- [2025-10-11: Percona Server MySQL Operator & Improved Error Handling](.changelog/2025-10-11-percona-server-mysql-operator-and-improved-error-handling.md) - Added support for Percona Server for MySQL Operator and significantly improved CLI error messaging with beautiful, actionable guidance for unsupported cloud resource kinds
- [2025-10-10: Percona Server MongoDB Operator Support](.changelog/2025-10-10-percona-server-mongodb-operator.md) - Added support for deploying and managing the Percona Server for MongoDB Operator on Kubernetes clusters with automated deployment, backup, and high availability features
- [2025-10-10: PostgreSQL Backup and Disaster Recovery](.changelog/2025-10-10-postgres-backup-configuration.md) - Added comprehensive backup and disaster recovery capabilities for PostgreSQL databases with operator-level defaults and per-database overrides
- [2025-09-16: Manifest Backend Configuration](.changelog/2025-09-16-manifest-backend-configuration.md) - Added support for embedding Pulumi and Terraform/Tofu backend configuration in manifest labels

## Notes

- Each changelog file represents a single logical change or feature
- Changes are ordered chronologically by filename
- These entries will be aggregated for release notes and web documentation
- Keep entries detailed but user-focused
