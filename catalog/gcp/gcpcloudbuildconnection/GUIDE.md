# GcpCloudBuildConnection Guide

The judgment this guide protects: a connection is the one place a code host's credentials meet Google Cloud, so the credentials never enter a manifest -- every token is a Secret Manager version the Cloud Build service agent reads.

## Credentials

Store each token and webhook secret as a `GcpSecretManagerSecret` with an initial version, and reference its `latest_version_name` (a literal `.../versions/latest` follows rotation instead of pinning one version). Grant `service-{PROJECT_NUMBER}@gcp-sa-cloudbuild.iam.gserviceaccount.com` `roles/secretmanager.secretAccessor` on each secret through the secret's `iamMembers` before the connection is created; the agent reads them at create time. GitLab and Bitbucket take two tokens: a read-write one Cloud Build uses to install webhooks and a read-only one for everyday reads.

## GitHub

github.com connections go through Cloud Build's GitHub App. Install the app on the account or organization, then set `githubConfig.appInstallationId`; `authorizerCredential` names the OAuth token of the account that authorized the app (a robot account, not a person). Until the installation is complete, the connection exports `installation_stage` and the `installation_action_uri` a person follows to finish it. GitHub Enterprise servers use a GitHub App created on the server; its ID, slug, installation ID, private key, and webhook secret go into `githubEnterpriseConfig`.

## Private servers

For an on-premises GitHub Enterprise, GitLab Enterprise, or Bitbucket Data Center server, `serviceDirectoryConfig.service` routes Cloud Build to it through Service Directory instead of the internet, and `sslCa` trusts its private certificate authority.

## Lifecycle

The location and ID are permanent. Credentials, the installation ID, annotations, and `disabled` update in place, except that GitLab and both Bitbucket hosts replace the connection when the webhook secret changes. Destroy the linked repositories before the connection.
