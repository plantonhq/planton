package module

import (
	"fmt"
	"strings"
)

// renderBaoConfigHcl synthesizes the OpenBao server configuration from
// the typed spec. This module OWNS config rendering — the chart takes
// config as a raw HCL string it writes into a ConfigMap, so every typed
// field (TLS listener material, storage backend, Raft peers, seal
// stanzas, telemetry) converges here. The Terraform twin builds the
// byte-identical document in locals.tf; keep them in lockstep.
//
// SENSITIVE-MATERIAL RULE: this document lands in a ConfigMap. Seal
// stanzas carry only NON-credential parameters (regions, key ids,
// addresses); credential material rides environment variables from the
// module-owned seal-credentials Secret (seal_secret.go) — the seal
// wrappers read their cloud SDKs' standard env vars. The PostgreSQL
// stanza carries NO connection_url at all: the driver reads the whole
// connection from its environment (postgres_env.go), the password from
// the referenced Secret.
//
// Dev mode renders NO config: `bao server -dev` ignores it (in-memory,
// auto-unsealed).
func renderBaoConfigHcl(locals *Locals) string {
	if locals.Dev {
		return ""
	}

	var b strings.Builder

	// The web UI is served by the same listener; the ui_enabled spec
	// field drives BOTH this config line and the chart's `<name>-ui`
	// Service (two switches upstream, one field here).
	uiEnabled := true
	if locals.Spec.UiEnabled != nil {
		uiEnabled = locals.Spec.GetUiEnabled()
	}
	fmt.Fprintf(&b, "ui = %t\n\n", uiEnabled)

	// ------------------------------ listener ------------------------------
	b.WriteString("listener \"tcp\" {\n")
	if locals.TlsEnabled {
		// The certificate Secret is mounted at the TLS mount path
		// (values.go wires server.volumes/volumeMounts); cert-manager
		// Secrets carry exactly these file names.
		fmt.Fprintf(&b, "  tls_cert_file = \"%s/tls.crt\"\n", vars.TlsMountPath)
		fmt.Fprintf(&b, "  tls_key_file = \"%s/tls.key\"\n", vars.TlsMountPath)
	} else {
		b.WriteString("  tls_disable = 1\n")
	}
	fmt.Fprintf(&b, "  address = \"[::]:%d\"\n", vars.ApiPort)
	fmt.Fprintf(&b, "  cluster_address = \"[::]:%d\"\n", vars.ClusterPort)
	if locals.Spec.GetMetrics().GetEnabled() {
		// Prometheus scrapes /v1/sys/metrics without an OpenBao token
		// only when the listener allows it — scoped to the metrics
		// endpoint, everything else stays authenticated.
		b.WriteString("  telemetry {\n")
		b.WriteString("    unauthenticated_metrics_access = \"true\"\n")
		b.WriteString("  }\n")
	}
	b.WriteString("}\n\n")

	// ------------------------------ storage -------------------------------
	// Both engines run in the chart's HA mode, so both end with the
	// service_registration stanza: the server patches
	// openbao-active/openbao-sealed labels onto its own pod, which is what
	// the chart's active/standby Services select on — without it those
	// Services would select nothing.
	switch locals.Storage {
	case storagePostgresql:
		b.WriteString("storage \"postgresql\" {\n")
		// ha_enabled is UNCONDITIONAL, one replica included: the server
		// labels its pod active only from its HA leader path, and that
		// path runs only when the backend reports HA enabled. Without it
		// a one-replica PostgreSQL vault would never carry the
		// openbao-active label and the `-active` Service (and the
		// active_service output) would select nothing. One holder of the
		// lock table costs nothing.
		b.WriteString("  ha_enabled = \"true\"\n")
		if pg := locals.Spec.GetServer().GetPostgresql(); pg.MaxParallel != nil && pg.GetMaxParallel() > 0 {
			// The per-server connection ceiling; the backend's own default
			// is 128, more than a default PostgreSQL's max_connections.
			fmt.Fprintf(&b, "  max_parallel = \"%d\"\n", pg.GetMaxParallel())
		}
		// No connection_url, deliberately: this document is a ConfigMap,
		// and the driver reads PGHOST/PGPORT/PGDATABASE/PGUSER/PGSSLMODE
		// and PGPASSWORD from the pod's environment when the URL is blank.
		b.WriteString("}\n\n")
		b.WriteString("service_registration \"kubernetes\" {}\n\n")
	case storageRaft:
		fmt.Fprintf(&b, "storage \"raft\" {\n")
		fmt.Fprintf(&b, "  path = \"%s\"\n", vars.DataMountPath)
		// THE RETRY_JOIN SYNTHESIS: the chart ships NO retry_join —
		// without these blocks a multi-replica Raft install never forms
		// a cluster (each pod sits uninitialized and independent,
		// verified at chart 0.28.6). Peer addresses are the StatefulSet
		// pods' stable DNS names through the headless `-internal`
		// Service; fullnameOverride is pinned to metadata.name, so the
		// names are deterministic. Joins are idempotent — a node
		// already in the cluster ignores its own entry.
		for i := 0; i < locals.Replicas; i++ {
			b.WriteString("  retry_join {\n")
			fmt.Fprintf(&b, "    leader_api_addr = \"%s://%s-%d.%s-internal:%d\"\n",
				locals.Scheme, locals.ReleaseName, i, locals.ReleaseName, vars.ApiPort)
			if locals.TlsEnabled {
				// Peers verify each other against the certificate's own
				// CA — cert-manager Secrets include ca.crt; the
				// certificate must cover the pod DNS names
				// (*.<name>-internal or explicit SANs).
				fmt.Fprintf(&b, "    leader_ca_cert_file = \"%s/ca.crt\"\n", vars.TlsMountPath)
			}
			b.WriteString("  }\n")
		}
		b.WriteString("}\n\n")
		b.WriteString("service_registration \"kubernetes\" {}\n\n")
	}

	// ------------------------------ seal -----------------------------------
	if seal := locals.Spec.GetAutoUnseal(); seal != nil {
		switch {
		case seal.GetAwsKms() != nil:
			aws := seal.GetAwsKms()
			b.WriteString("seal \"awskms\" {\n")
			fmt.Fprintf(&b, "  region = \"%s\"\n", aws.GetRegion())
			fmt.Fprintf(&b, "  kms_key_id = \"%s\"\n", aws.GetKmsKeyId())
			b.WriteString("}\n\n")
		case seal.GetGcpKms() != nil:
			gcp := seal.GetGcpKms()
			b.WriteString("seal \"gcpckms\" {\n")
			fmt.Fprintf(&b, "  project = \"%s\"\n", gcp.GetProject().GetValue())
			fmt.Fprintf(&b, "  region = \"%s\"\n", gcp.GetRegion())
			fmt.Fprintf(&b, "  key_ring = \"%s\"\n", gcp.GetKeyRing().GetValue())
			fmt.Fprintf(&b, "  crypto_key = \"%s\"\n", gcp.GetCryptoKey().GetValue())
			b.WriteString("}\n\n")
		case seal.GetAzureKeyVault() != nil:
			az := seal.GetAzureKeyVault()
			b.WriteString("seal \"azurekeyvault\" {\n")
			fmt.Fprintf(&b, "  vault_name = \"%s\"\n", az.GetVaultName())
			fmt.Fprintf(&b, "  key_name = \"%s\"\n", az.GetKeyName())
			fmt.Fprintf(&b, "  tenant_id = \"%s\"\n", az.GetTenantId())
			if az.GetClientId() != "" {
				fmt.Fprintf(&b, "  client_id = \"%s\"\n", az.GetClientId())
			}
			b.WriteString("}\n\n")
		case seal.GetTransit() != nil:
			tr := seal.GetTransit()
			b.WriteString("seal \"transit\" {\n")
			fmt.Fprintf(&b, "  address = \"%s\"\n", tr.GetAddress())
			fmt.Fprintf(&b, "  key_name = \"%s\"\n", tr.GetKeyName())
			mountPath := "transit/"
			if tr.MountPath != nil && tr.GetMountPath() != "" {
				mountPath = tr.GetMountPath()
			}
			fmt.Fprintf(&b, "  mount_path = \"%s\"\n", mountPath)
			b.WriteString("}\n\n")
		}
	}

	// ------------------------------ telemetry ------------------------------
	if locals.Spec.GetMetrics().GetEnabled() {
		b.WriteString("telemetry {\n")
		b.WriteString("  prometheus_retention_time = \"30s\"\n")
		b.WriteString("  disable_hostname = true\n")
		b.WriteString("}\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}
