# Open source on your own Kubernetes cluster (draft 2)

Draft 1's recommendation (v1, below as it stands on the after) was shown to the founder on 2026-11-26. He asked whether the section should explain why people want this or only what Planton does, whether to say "Kubernetes" rather than "cluster", and for another round. The lead's answer: the why stays, as one clause (the bill); the section shows the what, payoff before mechanism; "Kubernetes" is said, because "cluster" alone lost the non-technical reader, and the self-hosting section below stays apart by its subject (Planton itself), not by avoiding the word.

## v1 (as approved for review)
- Heading: Postgres, Valkey and Keycloak on your own cluster, instead of three managed bills.
- Line: One ask for a dev environment: your coding agent writes the files from Planton's catalog, already wired together, and the open-source CLI deploys them in order. No account, and no password in any file.
- Left: the ask (quoted), "Just the CLI: nothing of Planton is installed on your cluster.", then the table "35+ open-source projects the catalog runs on your cluster, each proven by its own test".
- Right, in order: keycloak.yaml's db: block (16 lines); the deploy (command, the open-source line, the deploy-order line, six deployed lines); "Your cluster, after" (Instead of | On your cluster | Answered).

## v2 (the lead's recommendation)
- **Heading:** Run open source on your own Kubernetes cluster, instead of paying for managed services. *(the founder's value 8; ch 8 proof)*
- **Line:** Ask your coding agent for the database, cache or sign-in your app needs. It writes them from Planton's catalog, already wired together, and the open-source CLI deploys them in order, with no account and no password in any file. *(ch 8 proof; ledger set-deploy row)*
- **Left:** the ask, labelled "One developer's ask, for a dev environment, to Claude Code:"; then "Just the CLI: nothing of Planton is installed on your cluster."
- **Right, in order (the story a film would tell):**
  1. "One command, as run with the open-source CLI": the command, the open-source line, ⋯, the six deployed lines in order, "All 6 resources deployed". The deploy-order line is cut: the six lines already show the order.
  2. "Your cluster, after": Instead of | On your cluster | Answered (a managed Postgres → Postgres → Cluster in healthy state, with Keycloak's 100 tables; a managed cache → Valkey → NOAUTH without its password, PONG with it; a hosted sign-in service → Keycloak → Ready, and its master realm answers 200).
  3. "How they're wired: keycloak.yaml reads its database from Postgres": the db: block cut to the host (7 lines) and the password's source (6 lines).
- **Under the left column:** the table, now labelled "Postgres, Valkey and Keycloak are three of 35+ open-source projects the catalog runs on Kubernetes, each proven by its own test", then the two checks.
