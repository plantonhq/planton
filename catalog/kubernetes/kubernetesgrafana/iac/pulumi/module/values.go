package module

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	kubernetesgrafanav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgrafana/v1alpha1"
	"sigs.k8s.io/yaml"
)

// buildHelmValues renders the typed spec into the chart's values map, then
// merges the spec's helm_values escape hatch over it with Helm `-f`
// semantics (maps deep-merge with the later document winning, lists
// replace).
//
// SECRET DISCIPLINE (load-bearing): the chart renders grafana.ini AND the
// datasource provisioning file into a ConfigMap, and its own
// assertNoLeakedSecrets helper REFUSES to render known secret paths into
// it. Every credential below therefore travels as an environment variable
// sourced from a Secret (envValueFrom / the chart's smtp existingSecret
// wiring), with the provisioning file carrying only a $__env{VAR}
// placeholder that Grafana expands at runtime. Never map a typed field
// into ini/provisioning text directly if it can carry a credential.
//
// PARITY: the Terraform module reaches the same result natively — its
// helm_release passes values = [yamlencode(typed values), helm_values] and
// the provider merges the documents in exactly this order. Keep every typed
// mapping below in lockstep with the Terraform module's locals.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec
	values := map[string]interface{}{}

	// fullnameOverride pins grafana.fullname to the resource name: the
	// Service, the chart-generated admin Secret, the ConfigMap and the
	// ServiceAccount all derive deterministically, and the exported
	// outputs are built from that contract.
	values["fullnameOverride"] = locals.ReleaseName

	// ---- replicas ----------------------------------------------------------
	// replicas > 1 is CEL-fenced to require the external database: the
	// embedded SQLite state cannot be shared between pods.
	replicas := int32(1)
	if spec.Replicas != nil {
		replicas = spec.GetReplicas()
	}
	values["replicas"] = int(replicas)

	// ---- container resources ----------------------------------------------
	if r := spec.GetResources(); r != nil {
		resources := map[string]interface{}{}
		if q := r.GetRequests(); q != nil && (q.GetCpu() != "" || q.GetMemory() != "") {
			requests := map[string]interface{}{}
			if q.GetCpu() != "" {
				requests["cpu"] = q.GetCpu()
			}
			if q.GetMemory() != "" {
				requests["memory"] = q.GetMemory()
			}
			resources["requests"] = requests
		}
		if l := r.GetLimits(); l != nil && (l.GetCpu() != "" || l.GetMemory() != "") {
			limits := map[string]interface{}{}
			if l.GetCpu() != "" {
				limits["cpu"] = l.GetCpu()
			}
			if l.GetMemory() != "" {
				limits["memory"] = l.GetMemory()
			}
			resources["limits"] = limits
		}
		if len(resources) > 0 {
			values["resources"] = resources
		}
	}

	// ---- admin credentials ---------------------------------------------------
	// Existing-secret arm: the chart reads the referenced Secret at pod
	// start (env wiring) — it must exist BEFORE the install. Generate arm
	// (admin_secret absent): the chart creates its own `<fullname>` Secret
	// ONCE — its lookup keeps the password stable across upgrades — so
	// nothing is rendered here at all. Credential material never appears
	// in these values either way.
	if existing := spec.GetAdminSecret(); existing != nil {
		userKey := existing.GetUserKey()
		if userKey == "" {
			userKey = vars.AdminUserKey
		}
		passwordKey := existing.GetPasswordKey()
		if passwordKey == "" {
			passwordKey = vars.AdminPasswordKey
		}
		values["admin"] = map[string]interface{}{
			"existingSecret": existing.GetName(),
			"userKey":        userKey,
			"passwordKey":    passwordKey,
		}
	}

	// ---- persistence -----------------------------------------------------------
	// Rendered only when declared: the chart default is ephemeral, which
	// the spec documents honestly (provisioned-as-code state needs no
	// volume; hand-authored dashboards do).
	if storage := spec.GetStorage(); storage != nil {
		size := storage.GetSize()
		if size == "" {
			size = vars.DefaultStorageSize
		}
		persistence := map[string]interface{}{
			"enabled": true,
			"size":    size,
		}
		if sc := storage.GetStorageClass().GetValue(); sc != "" {
			persistence["storageClassName"] = sc
		}
		values["persistence"] = persistence
	}

	// ---- external database --------------------------------------------------------
	// Rendered as GF_DATABASE_* environment variables — Grafana's
	// env-based configuration is first-class and keeps the password out
	// of grafana.ini (which lands in a ConfigMap). The password rides
	// envValueFrom (a secretKeyRef the kubelet resolves), so it never
	// appears in rendered values or in any config text.
	env := map[string]interface{}{}
	envValueFrom := map[string]interface{}{}
	if db := spec.GetDatabase(); db != nil {
		env["GF_DATABASE_TYPE"] = databaseEngineValue(db.GetEngine())
		env["GF_DATABASE_HOST"] = db.GetHost().GetValue()
		env["GF_DATABASE_NAME"] = db.GetName()
		env["GF_DATABASE_USER"] = db.GetUser()
		if db.GetSslMode() != "" {
			env["GF_DATABASE_SSL_MODE"] = db.GetSslMode()
		}
		envValueFrom["GF_DATABASE_PASSWORD"] = map[string]interface{}{
			"secretKeyRef": map[string]interface{}{
				"name": db.GetPasswordSecret().GetName(),
				"key":  db.GetPasswordSecret().GetKey(),
			},
		}
	}

	// ---- datasources -----------------------------------------------------------------
	// Rendered into the chart's datasources provisioning value
	// (datasources.yaml). Basic-auth passwords ride $__env{VAR}
	// placeholders expanded by Grafana at runtime from envValueFrom
	// secretKeyRefs — the provisioning file itself stays credential-free
	// (it lands in the chart's ConfigMap).
	if len(spec.GetDatasources()) > 0 {
		entries := make([]interface{}, 0, len(spec.GetDatasources()))
		for _, ds := range spec.GetDatasources() {
			dsType := ds.GetType()
			if dsType == "" {
				dsType = "prometheus"
			}
			entry := map[string]interface{}{
				"name":   ds.GetName(),
				"type":   dsType,
				"access": "proxy",
				"url":    ds.GetUrl().GetValue(),
			}
			if ds.GetIsDefault() {
				entry["isDefault"] = true
			}
			if ds.GetUid() != "" {
				entry["uid"] = ds.GetUid()
			}
			if ds.GetJsonData() != "" {
				jsonData := map[string]interface{}{}
				if err := yaml.Unmarshal([]byte(ds.GetJsonData()), &jsonData); err != nil {
					return nil, errors.Wrapf(err, "failed to parse json_data of datasource %q as a YAML document", ds.GetName())
				}
				entry["jsonData"] = jsonData
			}
			if ba := ds.GetBasicAuth(); ba != nil {
				envVar := datasourcePasswordEnvVar(ds.GetName())
				entry["basicAuth"] = true
				entry["basicAuthUser"] = ba.GetUsername()
				entry["secureJsonData"] = map[string]interface{}{
					"basicAuthPassword": fmt.Sprintf("$__env{%s}", envVar),
				}
				envValueFrom[envVar] = map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": ba.GetPasswordSecret().GetName(),
						"key":  ba.GetPasswordSecret().GetKey(),
					},
				}
			}
			entries = append(entries, entry)
		}
		values["datasources"] = map[string]interface{}{
			"datasources.yaml": map[string]interface{}{
				"apiVersion":  1,
				"datasources": entries,
			},
		}
	}

	// ---- dashboard discovery sidecar ------------------------------------------------------
	// The composition contract: any ConfigMap labeled grafana_dashboard:
	// "1" in ANY namespace becomes a dashboard here — other components
	// and teams ship dashboards without touching this resource. Default
	// ON (proto default true); when disabled the block is omitted, which
	// is the chart's own default (sidecar.dashboards.enabled: false).
	dashboardSidecar := true
	if spec.DashboardSidecarEnabled != nil {
		dashboardSidecar = spec.GetDashboardSidecarEnabled()
	}
	if dashboardSidecar {
		values["sidecar"] = map[string]interface{}{
			"dashboards": map[string]interface{}{
				"enabled":         true,
				"searchNamespace": "ALL",
			},
		}
	}

	// ---- community dashboards ---------------------------------------------------------------
	// gnetId imports need a file dashboard provider; the chart's
	// dashboards value keys entries under the provider's name.
	if len(spec.GetCommunityDashboards()) > 0 {
		values["dashboardProviders"] = map[string]interface{}{
			"dashboardproviders.yaml": map[string]interface{}{
				"apiVersion": 1,
				"providers": []interface{}{
					map[string]interface{}{
						"name":            "default",
						"orgId":           1,
						"folder":          "",
						"type":            "file",
						"disableDeletion": false,
						"options": map[string]interface{}{
							"path": "/var/lib/grafana/dashboards/default",
						},
					},
				},
			},
		}
		dashboards := map[string]interface{}{}
		for _, cd := range spec.GetCommunityDashboards() {
			entry := map[string]interface{}{
				"gnetId":     int(cd.GetGnetId()),
				"datasource": cd.GetDatasource(),
			}
			if cd.GetRevision() > 0 {
				entry["revision"] = int(cd.GetRevision())
			}
			dashboards[fmt.Sprintf("gnet-%d", cd.GetGnetId())] = entry
		}
		values["dashboards"] = map[string]interface{}{"default": dashboards}
	}

	// ---- plugins ---------------------------------------------------------------------------------
	// Grafana 13 ships its once-core datasource plugins (elasticsearch,
	// cloudwatch) outside the image, and the image's bundled-plugin
	// directory is read-only — the chart's shadowBundledPlugins empties
	// it into an emptyDir so listed plugins install cleanly. Rendered
	// together so a plugin list can never hit the read-only-directory
	// failure.
	if len(spec.GetPlugins()) > 0 {
		plugins := make([]interface{}, 0, len(spec.GetPlugins()))
		for _, p := range spec.GetPlugins() {
			plugins = append(plugins, p)
		}
		values["plugins"] = plugins
		values["shadowBundledPlugins"] = true
	}

	// ---- grafana.ini (non-secret settings ONLY) ----------------------------------------------------
	ini := map[string]interface{}{}
	if server := spec.GetServer(); server != nil && server.GetRootUrl() != "" {
		ini["server"] = map[string]interface{}{"root_url": server.GetRootUrl()}
	}
	if auth := spec.GetAuth(); auth != nil {
		if auth.GetAnonymousEnabled() {
			orgRole := auth.GetAnonymousOrgRole()
			if orgRole == "" {
				orgRole = "Viewer"
			}
			ini["auth.anonymous"] = map[string]interface{}{
				"enabled":  true,
				"org_role": orgRole,
			}
		}
		if auth.GetDisableLoginForm() {
			ini["auth"] = map[string]interface{}{"disable_login_form": true}
		}
	}
	if smtp := spec.GetSmtp(); smtp != nil {
		smtpIni := map[string]interface{}{
			"enabled": true,
			"host":    smtp.GetHost(),
		}
		if smtp.GetFromAddress() != "" {
			smtpIni["from_address"] = smtp.GetFromAddress()
		}
		if smtp.GetFromName() != "" {
			smtpIni["from_name"] = smtp.GetFromName()
		}
		if smtp.GetSkipVerify() {
			smtpIni["skip_verify"] = true
		}
		ini["smtp"] = smtpIni
		// Credentials ride the chart's own existingSecret wiring — it
		// injects GF_SMTP_USER / GF_SMTP_PASSWORD from the referenced
		// Secret (keys user / password), which override the ini section
		// at runtime. Nothing secret lands in the ini text.
		if smtp.GetCredentialsSecretName() != "" {
			values["smtp"] = map[string]interface{}{
				"existingSecret": smtp.GetCredentialsSecretName(),
			}
		}
	}
	// ---- sign-in (auth.google / auth.generic_oauth) -------------------------------------------------
	// Non-secret provider settings join grafana.ini; each client secret
	// rides envValueFrom into the module-owned Secret (signin.go). The
	// pod carries the Secret's fingerprint because Grafana reads it only
	// at start: a rotated secret changes the pod template, so the next
	// apply rolls Grafana onto it.
	if s := locals.SignIn; s != nil {
		for section, settings := range s.Ini {
			ini[section] = settings
		}
		for name, source := range s.EnvValueFrom {
			envValueFrom[name] = source
		}
		values["podAnnotations"] = map[string]interface{}{
			vars.CredentialsChecksumAnnotation: credentialsChecksum(s.SecretData),
		}
	}
	if len(ini) > 0 {
		values["grafana.ini"] = ini
	}

	// ---- observability -----------------------------------------------------------------------------
	if spec.GetServiceMonitorEnabled() {
		values["serviceMonitor"] = map[string]interface{}{"enabled": true}
	}

	// ---- image ---------------------------------------------------------------------------------------
	// The spec's repository carries the registry
	// ("my.registry.com/grafana/grafana") but the chart keeps them as
	// SEPARATE values composed {registry}/{repository}:{tag} with
	// registry defaulting to docker.io — so the declared value is split
	// (see splitImageRepository) or a mirror override would render a
	// docker.io/-prefixed broken reference. pullSecrets is the
	// private-mirror credential fold — the Terraform twin renders the
	// same split and list (a silent single-engine ImagePullBackOff class
	// otherwise).
	if img := spec.GetImage(); img != nil &&
		(img.GetRepository() != "" || img.GetTag() != "" || img.GetPullSecretName() != "") {
		image := map[string]interface{}{}
		if img.GetRepository() != "" {
			registry, repository := splitImageRepository(img.GetRepository())
			if registry != "" {
				image["registry"] = registry
			}
			image["repository"] = repository
		}
		if img.GetTag() != "" {
			image["tag"] = img.GetTag()
		}
		if img.GetPullSecretName() != "" {
			image["pullSecrets"] = []interface{}{img.GetPullSecretName()}
		}
		values["image"] = image
	}

	// ---- scheduling -------------------------------------------------------------------------------------
	if sched := spec.GetScheduling(); sched != nil {
		if len(sched.GetNodeSelector()) > 0 {
			values["nodeSelector"] = stringMapToInterface(sched.GetNodeSelector())
		}
		if len(sched.GetTolerations()) > 0 {
			values["tolerations"] = tolerationsSlice(sched.GetTolerations())
		}
		if sched.GetPriorityClassName() != "" {
			values["priorityClassName"] = sched.GetPriorityClassName()
		}
	}

	// ---- env wiring (database, datasource and sign-in credentials) ---------------------------------------
	if len(env) > 0 {
		values["env"] = env
	}
	if len(envValueFrom) > 0 {
		values["envValueFrom"] = envValueFrom
	}

	// ---- escape hatch (merged LAST, helm -f semantics) ----------------------------------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	// fullnameOverride re-pinned AFTER the merge — the one deliberate
	// exception to the escape hatch's last-word contract (twin of the
	// Terraform module's third values document). The Service and the
	// chart-generated admin Secret — and the exported outputs built from
	// them — derive from the fullname; letting an override move it would
	// break every output.
	values["fullnameOverride"] = locals.ReleaseName

	return values, nil
}

// databaseEngineValue maps the spec's engine enum onto Grafana's
// GF_DATABASE_TYPE vocabulary.
func databaseEngineValue(engine kubernetesgrafanav1alpha1.KubernetesGrafanaDatabaseEngine) string {
	if engine == kubernetesgrafanav1alpha1.KubernetesGrafanaDatabaseEngine_mysql {
		return "mysql"
	}
	return "postgres"
}

var nonEnvChars = regexp.MustCompile(`[^A-Z0-9]+`)

// datasourcePasswordEnvVar derives the deterministic env-var name carrying
// one datasource's basic-auth password (e.g. "External Mimir" →
// GF_DS_EXTERNAL_MIMIR_PASSWORD). Both engines derive identically — the
// name appears in the provisioning file AND the envValueFrom key, and the
// pair must agree.
func datasourcePasswordEnvVar(datasourceName string) string {
	sanitized := nonEnvChars.ReplaceAllString(strings.ToUpper(datasourceName), "_")
	sanitized = strings.Trim(sanitized, "_")
	return fmt.Sprintf("GF_DS_%s_PASSWORD", sanitized)
}
