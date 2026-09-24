# GcpDatastreamConnectionProfile Guide

The judgment this guide protects: a profile is a shared address book entry. Create one per database and one per destination, give it a credential that can only replicate, and let every stream reuse it.

## Connectivity

- **Public** (neither option set) -- Datastream connects from its regional static IPs. Allowlist them on the database: for Cloud SQL, in the instance's authorized networks (Google's documented path, and why a `GcpCloudSql` reference defaults to `public_ip`). Read the list from Datastream's console or its `fetchStaticIps` API for your region.
- **`privateConnection`** -- through a `GcpDatastreamPrivateConnection` into your VPC. Private Cloud SQL and AlloyDB need a proxy VM in that VPC as the hostname, because peering is not transitive.
- **`forwardSshConnectivity`** -- through an SSH bastion; prefer `privateKey` to `password`.

## Credentials

Use a database user that can only replicate: REPLICATION SLAVE, REPLICATION CLIENT, and SELECT on MySQL; the REPLICATION attribute on PostgreSQL. Keep its password in Secret Manager and name the version in `secretManagerStoredPassword`. Datastream reads it as its service agent, `service-{project-number}@gcp-sa-datastream.iam.gserviceaccount.com`, so grant that identity `roles/secretmanager.secretAccessor` in the secret's `iamMembers`. A `GcpSecretManagerSecret` reference resolves to the version it seeded (`latest_version_name`); rotate by adding a version and pointing the profile at it.

## Destinations

`bigqueryProfile: true` has no settings: each stream chooses datasets. Datastream's service agent needs BigQuery Data Editor where it writes. A `gcsProfile` names a bucket and an optional root path; grant the service agent object write access through the bucket's `iamMembers`.

## Validation

Google tests connectivity at create. `createWithoutValidation: true` skips the test -- useful while a database or network is still being built -- but a wrong host or password then surfaces only when a stream starts.
