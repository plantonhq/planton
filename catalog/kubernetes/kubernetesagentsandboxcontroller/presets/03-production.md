# Production

A shape to run agents on for real.
- **Controller:** two replicas, so a hot standby takes over at once if the leader's node dies, with requests and limits the cluster's quotas can count.
- **Router:** three replicas behind a disruption budget of two, so a node drain never takes the router down.
- **Authorization:** every request must carry a Kubernetes bearer token, which the router checks through the TokenReview API.
- **Network policy:** only the `agents` namespace may call the router.

Your product's pods send their ServiceAccount token (`Authorization: Bearer ...`), and the router refuses anything without one.

Change first:
- Set `clientNamespaceLabels` to your callers' namespaces.
- Set `tokenReview.audiences` if your callers send projected tokens minted for an audience.
- Size the resources from your fleet: the controller's work grows with the number of sandboxes, and the router's with requests per second.
- Add `router.tls` (see the TLS preset) when traffic must be encrypted inside the cluster.

See [03-production.yaml](./03-production.yaml) for the manifest.
