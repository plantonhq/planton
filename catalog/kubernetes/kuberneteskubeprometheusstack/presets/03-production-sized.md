# Production sized preset

The full production posture: an HA Prometheus pair (each replica
scrapes and stores the complete target set — duplication, not
sharding), a quorum-safe three-replica Alertmanager gossip cluster,
30-day/80GiB dual-bounded retention on 100Gi volumes, persistent
bundled Grafana, and explicit resources on every component so the
stack cannot be evicted by the workloads it is supposed to watch.

Retention is deliberately two-dimensional: `retention` trims by age,
`retention_size` trims by bytes BELOW the volume size — when the
volume itself fills, Prometheus crash-loops instead of trimming.
Memory is the number to watch: it scales with active series, and an
undersized limit is the most common cause of a crash-looping
Prometheus on busy clusters. Treat the 2Gi/4Gi shape here as a
starting point and raise it as target count grows.

The control-plane scraper posture matches the managed-cloud preset
because most production clusters are managed — on a self-hosted
control plane (kubeadm, datacenter) DROP the `control_plane_scrapers`
and `default_rules` blocks so the controller-manager, scheduler,
etcd and kube-proxy are scraped and their curated alerts stay armed.
The overcommit alerts are off because managed node pools usually
autoscale; on fixed-size pools remove `disabled_alerts`, where they
are true.

Alert delivery is declared, not left for later: every alert posts to
the team's Discord channel, `severity=page` alerts also ring the
on-call phone through Pushover (emergency priority, repeating until
acknowledged), and the Watchdog heartbeat lets an outside monitor page
when the cluster goes quiet. The pager receiver carries the channel's
webhook too, because the root receiver only takes alerts no child
route claims.

Change first: `external_labels.cluster` to the cluster's real name
(multi-cluster backends and federation key on it), then the
`$secret/` placeholders under `alertmanager.notifications` to your
organization's secrets and the heartbeat URL to your dead-man's-switch
monitor. Fire one test alert (the guide shows how) before trusting it.
Add `prometheus.remote_write` when a long-term or managed backend
enters the picture.

See [03-production-sized.yaml](./03-production-sized.yaml) for the
manifest.
