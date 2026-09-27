package module

import (
	"strings"

	"github.com/pkg/errors"
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetestemporalv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestemporal/v1alpha1"
	"sigs.k8s.io/yaml"
)

// buildHelmValues renders the typed spec into the chart's values map, then
// merges the spec's helm_values escape hatch over it with Helm `-f`
// semantics (maps deep-merge with the later document winning, lists
// replace).
//
// SECRET DISCIPLINE (load-bearing): nothing rendered here carries
// credential material. Every database password rides the chart's
// existingSecret contract — the chart wires a secretKeyRef and STRIPS the
// Helm-side keys before writing the server config, and because
// existingSecret is always set, the chart's own per-store password Secret
// (which would embed an inline password) is never created.
//
// PARITY: the Terraform module reaches the same result natively — its
// helm_release passes values = [yamlencode(typed values), helm_values] and
// the provider merges the documents in exactly this order. Keep every typed
// mapping below in lockstep with the Terraform module's locals.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec
	values := map[string]interface{}{}

	// fullnameOverride pins every child name (`<name>-frontend`,
	// `<name>-web`, ...) to the resource name; the exported outputs are
	// built from that contract.
	values["fullnameOverride"] = locals.ReleaseName

	// ---- the server block ------------------------------------------------
	server := map[string]interface{}{}

	// config: log level + persistence. createDatabase/manageSchema are
	// ALWAYS rendered explicitly — the chart's getStore helper silently
	// defaults BOTH to true when unset, and an unintended
	// create-database attempt fails against least-privilege users.
	logLevel := spec.GetLogLevel()
	if logLevel == "" {
		logLevel = "info"
	}

	numHistoryShards := int32(512)
	if spec.NumHistoryShards != nil {
		numHistoryShards = spec.GetNumHistoryShards()
	}

	datastores := map[string]interface{}{
		"default":    defaultStoreBlock(spec.GetDatabase()),
		"visibility": visibilityStoreBlock(spec.GetDatabase()),
	}

	config := map[string]interface{}{
		"logLevel": logLevel,
		"persistence": map[string]interface{}{
			"defaultStore":     "default",
			"visibilityStore":  "visibility",
			"numHistoryShards": int(numHistoryShards),
			"datastores":       datastores,
		},
	}

	// Declarative Temporal namespaces: a post-install Job runs
	// `temporal operator namespace describe || create` per entry.
	if len(spec.GetTemporalNamespaces()) > 0 {
		namespaceList := make([]interface{}, 0, len(spec.GetTemporalNamespaces()))
		for _, ns := range spec.GetTemporalNamespaces() {
			entry := map[string]interface{}{"name": ns.GetName()}
			retention := ns.GetRetention()
			if retention == "" {
				retention = "3d"
			}
			entry["retention"] = retention
			namespaceList = append(namespaceList, entry)
		}
		config["namespaces"] = map[string]interface{}{
			"create":    true,
			"namespace": namespaceList,
		}
	}

	server["config"] = config

	// Per-service sizing (the chart reads replicaCount per service,
	// falling back to server.replicaCount).
	if services := spec.GetServices(); services != nil {
		serviceBlocks := map[string]*kubernetestemporalv1alpha1.KubernetesTemporalServiceConfig{
			"frontend": services.GetFrontend(),
			"history":  services.GetHistory(),
			"matching": services.GetMatching(),
			"worker":   services.GetWorker(),
		}
		for name, sc := range serviceBlocks {
			if sc == nil {
				continue
			}
			block := map[string]interface{}{}
			if sc.Replicas != nil {
				block["replicaCount"] = int(sc.GetReplicas())
			}
			if resources := resourcesBlock(sc.GetResources()); resources != nil {
				block["resources"] = resources
			}
			if len(block) > 0 {
				server[name] = block
			}
		}
	}

	// The internal-frontend service (NOTE the chart key carries a dash).
	if spec.GetInternalFrontendEnabled() {
		server["internal-frontend"] = map[string]interface{}{"enabled": true}
	}

	// Dynamic-config limits (keys verified against the server source at
	// the pin). Each key takes a list of {value, constraints} entries;
	// an empty constraints object applies the value globally.
	if dc := spec.GetDynamicConfig(); dc != nil {
		dynamicConfig := map[string]interface{}{}
		addDynamicConfigInt(dynamicConfig, vars.DcHistorySizeLimitError, dc.HistorySizeLimitError)
		addDynamicConfigInt(dynamicConfig, vars.DcHistorySizeLimitWarn, dc.HistorySizeLimitWarn)
		addDynamicConfigInt(dynamicConfig, vars.DcHistoryCountLimitError, dc.HistoryCountLimitError)
		addDynamicConfigInt(dynamicConfig, vars.DcHistoryCountLimitWarn, dc.HistoryCountLimitWarn)
		addDynamicConfigInt(dynamicConfig, vars.DcBlobSizeLimitError, dc.BlobSizeLimitError)
		addDynamicConfigInt(dynamicConfig, vars.DcBlobSizeLimitWarn, dc.BlobSizeLimitWarn)
		if len(dynamicConfig) > 0 {
			server["dynamicConfig"] = dynamicConfig
		}
	}

	// Archival: the provider block enables the capability; the
	// namespaceDefaults URIs make every Temporal namespace archive by
	// default. Cloud credentials are ambient (IRSA / workload identity)
	// — nothing credential-bearing renders.
	if archival := spec.GetArchival(); archival != nil {
		var provider map[string]interface{}
		switch archival.GetProvider().(type) {
		case *kubernetestemporalv1alpha1.KubernetesTemporalArchival_S3:
			provider = map[string]interface{}{
				"s3store": map[string]interface{}{
					"region": archival.GetS3().GetRegion(),
				},
			}
		case *kubernetestemporalv1alpha1.KubernetesTemporalArchival_Gcs:
			provider = map[string]interface{}{
				"gstorage": map[string]interface{}{},
			}
		case *kubernetestemporalv1alpha1.KubernetesTemporalArchival_Filestore:
			provider = map[string]interface{}{
				"filestore": map[string]interface{}{
					"fileMode": "0666",
					"dirMode":  "0766",
				},
			}
		}
		archivalState := func() map[string]interface{} {
			return map[string]interface{}{
				"state":      "enabled",
				"enableRead": true,
				"provider":   provider,
			}
		}
		server["archival"] = map[string]interface{}{
			"history":    archivalState(),
			"visibility": archivalState(),
		}
		server["namespaceDefaults"] = map[string]interface{}{
			"archival": map[string]interface{}{
				"history": map[string]interface{}{
					"state": "enabled",
					"URI":   archival.GetHistoryUri(),
				},
				"visibility": map[string]interface{}{
					"state": "enabled",
					"URI":   archival.GetVisibilityUri(),
				},
			},
		}
	}

	// ServiceMonitor resources for the Prometheus Operator (one per
	// server service via the chart's global metrics block).
	if spec.GetServiceMonitorEnabled() {
		server["metrics"] = map[string]interface{}{
			"serviceMonitor": map[string]interface{}{"enabled": true},
		}
	}

	// Scheduling: the chart applies server.nodeSelector/tolerations to
	// all four services; web and admintools carry their own copies (the
	// schema Jobs schedule with admintools' values).
	scheduling := map[string]interface{}{}
	if sched := spec.GetScheduling(); sched != nil {
		if len(sched.GetNodeSelector()) > 0 {
			scheduling["nodeSelector"] = stringMapToInterface(sched.GetNodeSelector())
		}
		if len(sched.GetTolerations()) > 0 {
			scheduling["tolerations"] = tolerationsSlice(sched.GetTolerations())
		}
	}
	mergeInto(server, scheduling)

	// Server image override (all four services run the same image).
	if img := imageBlock(spec.GetImages().GetServer()); img != nil {
		server["image"] = img
	}
	// Every server service's pods (frontend, history, matching, worker)
	// carry the resource's identity labels.
	server["podLabels"] = stringMapToInterface(locals.Labels)

	values["server"] = server

	// ---- the web UI --------------------------------------------------------
	web := map[string]interface{}{}
	if !locals.WebUiEnabled {
		web["enabled"] = false
	} else {
		if w := spec.GetWebUi(); w != nil {
			if w.Replicas != nil {
				web["replicaCount"] = int(w.GetReplicas())
			}
			if resources := resourcesBlock(w.GetResources()); resources != nil {
				web["resources"] = resources
			}
		}
		if img := imageBlock(spec.GetImages().GetWebUi()); img != nil {
			web["image"] = img
		}
		mergeInto(web, scheduling)
		web["podLabels"] = stringMapToInterface(locals.Labels)
	}
	values["web"] = web

	// ---- admin tools --------------------------------------------------------
	// The image is needed even with the pod disabled — the schema and
	// namespace Jobs run it.
	adminToolsEnabled := true
	if spec.AdminToolsEnabled != nil {
		adminToolsEnabled = spec.GetAdminToolsEnabled()
	}
	admintools := map[string]interface{}{}
	if !adminToolsEnabled {
		admintools["enabled"] = false
	} else {
		admintools["podLabels"] = stringMapToInterface(locals.Labels)
	}
	if img := imageBlock(spec.GetImages().GetAdminTools()); img != nil {
		admintools["image"] = img
	}
	mergeInto(admintools, scheduling)
	if len(admintools) > 0 {
		values["admintools"] = admintools
	}

	// ---- the schema-setup and namespace Jobs' pods ----------------------------
	// The chart reads only podLabels from this block here; the Jobs
	// themselves are enabled by the datastores' manageSchema.
	values["schema"] = map[string]interface{}{"podLabels": stringMapToInterface(locals.Labels)}

	// ---- 1.29-image compatibility shims: OFF at this pin --------------------
	// The chart defaults both shims ON for Temporal 1.29 images; our pin
	// runs 1.31+, where they only add a ConfigMap and mounts.
	values["shims"] = map[string]interface{}{
		"dockerize":         false,
		"elasticsearchTool": false,
	}

	// ---- image pull secrets (global — the chart has one list) ----------------
	pullSecrets := []interface{}{}
	seen := map[string]bool{}
	for _, img := range []*kubernetesprovider.ContainerImage{
		spec.GetImages().GetServer(), spec.GetImages().GetWebUi(), spec.GetImages().GetAdminTools(),
	} {
		name := img.GetPullSecretName()
		if name != "" && !seen[name] {
			seen[name] = true
			pullSecrets = append(pullSecrets, map[string]interface{}{"name": name})
		}
	}
	if len(pullSecrets) > 0 {
		values["imagePullSecrets"] = pullSecrets
	}

	// ---- escape hatch (merged LAST, helm -f semantics) ------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	// fullnameOverride re-pinned AFTER the merge — the one deliberate
	// exception to the escape hatch's last-word contract (twin of the
	// Terraform module's third values document). Every child name — and
	// the exported outputs built from them — derive from the fullname;
	// letting an override move it would break every output.
	values["fullnameOverride"] = locals.ReleaseName

	return values, nil
}

// defaultStoreBlock renders the DEFAULT (workflow state) store from the
// database oneof.
func defaultStoreBlock(db *kubernetestemporalv1alpha1.KubernetesTemporalDatabase) map[string]interface{} {
	databaseName := db.GetDatabaseName()
	if databaseName == "" {
		databaseName = "temporal"
	}

	switch db.GetBackend().(type) {
	case *kubernetestemporalv1alpha1.KubernetesTemporalDatabase_Postgres:
		return map[string]interface{}{
			"sql": sqlStoreConfig(vars.PostgresPlugin, 5432, databaseName, db.GetPostgres(), db),
		}
	case *kubernetestemporalv1alpha1.KubernetesTemporalDatabase_Mysql:
		return map[string]interface{}{
			"sql": mysqlStoreConfig(databaseName, db.GetMysql(), db),
		}
	case *kubernetestemporalv1alpha1.KubernetesTemporalDatabase_Cassandra:
		return map[string]interface{}{
			"cassandra": cassandraStoreConfig(databaseName, db.GetCassandra(), db),
		}
	}
	return nil
}

// visibilityStoreBlock renders the VISIBILITY store: the dedicated
// `visibility` block when declared (required for cassandra), otherwise the
// default store's own connection pointed at the visibility database.
func visibilityStoreBlock(db *kubernetestemporalv1alpha1.KubernetesTemporalDatabase) map[string]interface{} {
	visibilityDatabase := db.GetVisibilityDatabaseName()
	if visibilityDatabase == "" {
		visibilityDatabase = "temporal_visibility"
	}

	if vis := db.GetVisibility(); vis != nil {
		if vis.GetDatabaseName() != "" {
			visibilityDatabase = vis.GetDatabaseName()
		}
		switch vis.GetBackend().(type) {
		case *kubernetestemporalv1alpha1.KubernetesTemporalVisibility_Postgres:
			return map[string]interface{}{
				"sql": sqlStoreConfig(vars.PostgresPlugin, 5432, visibilityDatabase, vis.GetPostgres(), db),
			}
		case *kubernetestemporalv1alpha1.KubernetesTemporalVisibility_Mysql:
			return map[string]interface{}{
				"sql": mysqlStoreConfig(visibilityDatabase, vis.GetMysql(), db),
			}
		}
	}

	// Derived: the same SQL server as the default store (CEL guarantees
	// the cassandra arm always declares a visibility block).
	switch db.GetBackend().(type) {
	case *kubernetestemporalv1alpha1.KubernetesTemporalDatabase_Postgres:
		return map[string]interface{}{
			"sql": sqlStoreConfig(vars.PostgresPlugin, 5432, visibilityDatabase, db.GetPostgres(), db),
		}
	case *kubernetestemporalv1alpha1.KubernetesTemporalDatabase_Mysql:
		return map[string]interface{}{
			"sql": mysqlStoreConfig(visibilityDatabase, db.GetMysql(), db),
		}
	}
	return nil
}

// sqlStoreConfig renders one PostgreSQL store config in the chart's raw
// server-config format. connectAddr carries host:port — the admintools/
// schema-Job env template REQUIRES that form (it parses SQL_HOST and
// SQL_PORT out of it).
func sqlStoreConfig(plugin string, defaultPort int32, databaseName string,
	pg *kubernetestemporalv1alpha1.KubernetesTemporalPostgres,
	db *kubernetestemporalv1alpha1.KubernetesTemporalDatabase,
) map[string]interface{} {
	port := defaultPort
	if pg.Port != nil {
		port = pg.GetPort()
	}
	secretKey := pg.GetPasswordSecret().GetSecretKey()
	if secretKey == "" {
		secretKey = "password"
	}
	block := sqlCommon(plugin, pg.GetHost().GetValue(), port, databaseName,
		pg.GetUsername(), pg.GetPasswordSecret().GetSecretName().GetValue(), secretKey,
		pg.MaxConns, pg.MaxIdleConns, pg.GetMaxConnLifetime(), db)
	if tls := tlsBlock(pg.GetTls()); tls != nil {
		block["tls"] = tls
	}
	return block
}

// mysqlStoreConfig renders one MySQL store config.
func mysqlStoreConfig(databaseName string,
	my *kubernetestemporalv1alpha1.KubernetesTemporalMysql,
	db *kubernetestemporalv1alpha1.KubernetesTemporalDatabase,
) map[string]interface{} {
	port := int32(3306)
	if my.Port != nil {
		port = my.GetPort()
	}
	block := sqlCommon(vars.MysqlPlugin, my.GetHost().GetValue(), port, databaseName,
		my.GetUsername(), my.GetPasswordSecret().GetSecretName().GetValue(), my.GetPasswordSecret().GetSecretKey(),
		my.MaxConns, my.MaxIdleConns, my.GetMaxConnLifetime(), db)
	if tls := tlsBlock(my.GetTls()); tls != nil {
		block["tls"] = tls
	}
	return block
}

// sqlCommon is the shared SQL store shape. existingSecret/secretKey are
// Helm-side keys: the chart wires the secretKeyRef and strips them from
// the rendered server config.
func sqlCommon(plugin, host string, port int32, databaseName, user, secretName, secretKey string,
	maxConns, maxIdleConns *int32, maxConnLifetime string,
	db *kubernetestemporalv1alpha1.KubernetesTemporalDatabase,
) map[string]interface{} {
	mc := int32(20)
	if maxConns != nil {
		mc = *maxConns
	}
	mic := int32(20)
	if maxIdleConns != nil {
		mic = *maxIdleConns
	}
	lifetime := maxConnLifetime
	if lifetime == "" {
		lifetime = "1h"
	}
	return map[string]interface{}{
		"pluginName":      plugin,
		"driverName":      plugin,
		"databaseName":    databaseName,
		"connectAddr":     joinHostPort(host, port),
		"connectProtocol": "tcp",
		"user":            user,
		"existingSecret":  secretName,
		"secretKey":       secretKey,
		"maxConns":        int(mc),
		"maxIdleConns":    int(mic),
		"maxConnLifetime": lifetime,
		"createDatabase":  db.GetCreateDatabases(),
		"manageSchema":    !db.GetSkipSchemaSetup(),
	}
}

// cassandraStoreConfig renders the Cassandra default-store config. hosts
// is a comma-joined string (the chart's own documented form; its env
// template takes the first for the schema tools).
func cassandraStoreConfig(keyspace string,
	cass *kubernetestemporalv1alpha1.KubernetesTemporalCassandra,
	db *kubernetestemporalv1alpha1.KubernetesTemporalDatabase,
) map[string]interface{} {
	port := int32(9042)
	if cass.Port != nil {
		port = cass.GetPort()
	}
	replicationFactor := int32(3)
	if cass.ReplicationFactor != nil {
		replicationFactor = cass.GetReplicationFactor()
	}
	secretKey := cass.GetPasswordSecret().GetSecretKey()
	if secretKey == "" {
		secretKey = "password"
	}
	block := map[string]interface{}{
		"hosts":             strings.Join(cass.GetHosts(), ","),
		"port":              int(port),
		"keyspace":          keyspace,
		"user":              cass.GetUsername(),
		"existingSecret":    cass.GetPasswordSecret().GetSecretName().GetValue(),
		"secretKey":         secretKey,
		"replicationFactor": int(replicationFactor),
		"createDatabase":    db.GetCreateDatabases(),
		"manageSchema":      !db.GetSkipSchemaSetup(),
	}
	if cass.GetDatacenter() != "" {
		block["datacenter"] = cass.GetDatacenter()
	}
	if tls := tlsBlock(cass.GetTls()); tls != nil {
		block["tls"] = tls
	}
	return block
}

// tlsBlock renders a store TLS block (nil when not declared).
func tlsBlock(tls *kubernetestemporalv1alpha1.KubernetesTemporalDatabaseTls) map[string]interface{} {
	if tls == nil {
		return nil
	}
	block := map[string]interface{}{
		"enabled":                tls.GetEnabled(),
		"enableHostVerification": tls.GetHostVerification(),
	}
	if tls.GetServerName() != "" {
		block["serverName"] = tls.GetServerName()
	}
	return block
}

// addDynamicConfigInt appends one global {value, constraints:{}} entry for
// a dynamic-config key when the spec sets it.
func addDynamicConfigInt(dst map[string]interface{}, key string, value *int64) {
	if value == nil {
		return
	}
	dst[key] = []interface{}{
		map[string]interface{}{
			"value":       *value,
			"constraints": map[string]interface{}{},
		},
	}
}
