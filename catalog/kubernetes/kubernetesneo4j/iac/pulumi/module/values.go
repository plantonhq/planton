package module

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// buildHelmValues renders the typed spec into the chart's values map, then
// merges the spec's helm_values escape hatch over it with Helm `-f`
// semantics (maps deep-merge with the later document winning, lists
// replace).
//
// PARITY: the Terraform module reaches the same result natively — its
// helm_release passes values = [yamlencode(typed values), helm_values] and
// the provider merges the documents in exactly this order. Keep every typed
// mapping below in lockstep with the Terraform module's locals.
func buildHelmValues(locals *Locals) (map[string]interface{}, error) {
	spec := locals.Spec
	values := map[string]interface{}{}

	// ---- neo4j (server identity, credentials, sizing) ---------------------
	// neo4j.name is REQUIRED by the chart (its neo4j.name helper fails the
	// install when empty — nothing defaults it to the release name), so the
	// module always renders it: cluster_name when set (Enterprise members
	// sharing it form one cluster), else metadata.name.
	neo4j := map[string]interface{}{
		"name":    locals.Neo4JName,
		"edition": locals.Edition,
	}

	// The chart's own shape for license acceptance is the STRING "yes"/"no"
	// (not a bool); its default is "no", so the module renders only the
	// affirmative.
	if spec.GetAcceptLicenseAgreement() {
		neo4j["acceptLicenseAgreement"] = "yes"
	}

	// The chart contract: passwordFromSecret names a Secret carrying key
	// NEO4J_AUTH with value "neo4j/<password>", and the chart LOOKS IT UP
	// at template time — the Secret must exist BEFORE the release (main.go
	// wires the explicit dependency on the module-materialized Secret).
	// Always set: the module materializes the Secret for every arm but
	// existing_secret, which names one the user owns. The password itself
	// NEVER appears in rendered chart values.
	neo4j["passwordFromSecret"] = locals.AuthSecretName

	// The chart's primary resources shape is flat {cpu, memory} applied to
	// BOTH requests and limits; it also accepts a full-format limits
	// sub-map (limits can differ from requests, or be omitted by the chart
	// only in full format). The module renders requests into the flat keys
	// and declared limits into the limits sub-map. NOTE the chart REJECTS
	// installs below its floor (500m CPU / 2Gi memory) — the module never
	// defaults below that; when spec.resources is empty nothing renders
	// and the chart's own defaults (1000m/2Gi) apply.
	if r := spec.GetResources(); r != nil {
		resources := map[string]interface{}{}
		if q := r.GetRequests(); q != nil {
			if q.GetCpu() != "" {
				resources["cpu"] = q.GetCpu()
			}
			if q.GetMemory() != "" {
				resources["memory"] = q.GetMemory()
			}
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
			neo4j["resources"] = resources
		}
	}

	values["neo4j"] = neo4j

	// ---- data volume -------------------------------------------------------
	// The chart REQUIRES volumes.data.mode (its values.yaml ships mode: ""
	// and the templates fail without one), so the module ALWAYS renders the
	// data volume: "dynamic" with the declared StorageClass, else
	// "defaultStorageClass" for the cluster default — size resolved to the
	// spec default (10Gi) either way.
	size := spec.GetDataVolume().GetSize()
	if size == "" {
		size = vars.DefaultDataVolumeSize
	}
	storageClass := spec.GetDataVolume().GetStorageClass().GetValue()

	dataVolume := map[string]interface{}{}
	if storageClass != "" {
		dataVolume["mode"] = "dynamic"
		dataVolume["dynamic"] = map[string]interface{}{
			"storageClassName": storageClass,
			"accessModes":      []interface{}{"ReadWriteOnce"},
			"requests":         map[string]interface{}{"storage": size},
		}
	} else {
		dataVolume["mode"] = "defaultStorageClass"
		dataVolume["defaultStorageClass"] = map[string]interface{}{
			"accessModes": []interface{}{"ReadWriteOnce"},
			"requests":    map[string]interface{}{"storage": size},
		}
	}
	values["volumes"] = map[string]interface{}{"data": dataVolume}

	// ---- neo4j.conf --------------------------------------------------------
	// The typed memory block renders as the three neo4j.conf memory keys,
	// merged over the free-form config map — TYPED KEYS WIN on collision
	// (the memory block is the declared interface for those keys; a user
	// config duplicate is silently overridden). The chart merges this map
	// over its own config defaults with Helm -f semantics.
	config := map[string]interface{}{}
	for key, value := range spec.GetConfig() {
		config[key] = value
	}
	if m := spec.GetMemory(); m != nil {
		if m.GetHeapInitial() != "" {
			config["server.memory.heap.initial_size"] = m.GetHeapInitial()
		}
		if m.GetHeapMax() != "" {
			config["server.memory.heap.max_size"] = m.GetHeapMax()
		}
		if m.GetPageCache() != "" {
			config["server.memory.pagecache.size"] = m.GetPageCache()
		}
	}
	if len(config) > 0 {
		values["config"] = config
	}

	// ---- apoc.conf -----------------------------------------------------------
	if len(spec.GetApocConfig()) > 0 {
		values["apoc_config"] = stringMapToInterface(spec.GetApocConfig())
	}

	// ---- jvm -----------------------------------------------------------------
	// useNeo4jDefaultJvmArguments renders only when explicitly declared
	// (the chart default is already true); additional arguments render as
	// the chart's list shape.
	jvm := map[string]interface{}{}
	if len(spec.GetAdditionalJvmArguments()) > 0 {
		jvm["additionalJvmArguments"] = toInterfaceSlice(spec.GetAdditionalJvmArguments())
	}
	if spec.UseDefaultJvmArguments != nil {
		jvm["useNeo4jDefaultJvmArguments"] = spec.GetUseDefaultJvmArguments()
	}
	if len(jvm) > 0 {
		values["jvm"] = jvm
	}

	// ---- the neo4j service --------------------------------------------------
	// DELIBERATE OVERRIDE OF THE CHART DEFAULT: the chart ships
	// services.neo4j.spec.type: LoadBalancer, which would provision a cloud
	// load balancer (or hang Pending) on every install. This component pins
	// it to ClusterIP unless spec.service.type says otherwise — exposure
	// composes from first-class kinds (KubernetesIngress, Gateway API)
	// over the exported service handle instead.
	serviceType := spec.GetService().GetType()
	if serviceType == "" {
		serviceType = "ClusterIP"
	}
	neo4jService := map[string]interface{}{
		"spec": map[string]interface{}{"type": serviceType},
	}
	if len(spec.GetService().GetAnnotations()) > 0 {
		neo4jService["annotations"] = stringMapToInterface(spec.GetService().GetAnnotations())
	}
	values["services"] = map[string]interface{}{"neo4j": neo4jService}

	// ---- ssl -----------------------------------------------------------------
	// Both privateKey.secretName and publicCertificate.secretName point at
	// the ONE resolved scope Secret: the chart mounts private.key and
	// public.crt from it (its subPath defaults). cert-manager Secrets carry
	// tls.key/tls.crt instead — the README documents the key bridge; the
	// module does NOT silently rewrite key names.
	ssl := map[string]interface{}{}
	if boltSecret := spec.GetSsl().GetBolt().GetSecret().GetValue(); boltSecret != "" {
		ssl["bolt"] = sslScopeMap(boltSecret)
	}
	if httpsSecret := spec.GetSsl().GetHttps().GetSecret().GetValue(); httpsSecret != "" {
		ssl["https"] = sslScopeMap(httpsSecret)
	}
	if len(ssl) > 0 {
		values["ssl"] = ssl
	}

	// ---- scheduling ------------------------------------------------------------
	// The chart reads nodeSelector at the TOP level; tolerations,
	// podAntiAffinity, and priorityClassName live under podSpec.
	// podAntiAffinity renders only when explicitly declared (chart default
	// is already true).
	if sched := spec.GetScheduling(); sched != nil {
		if len(sched.GetNodeSelector()) > 0 {
			values["nodeSelector"] = stringMapToInterface(sched.GetNodeSelector())
		}
		podSpec := map[string]interface{}{}
		if len(sched.GetTolerations()) > 0 {
			podSpec["tolerations"] = tolerationsSlice(sched.GetTolerations())
		}
		if sched.PodAntiAffinity != nil {
			podSpec["podAntiAffinity"] = sched.GetPodAntiAffinity()
		}
		if sched.GetPriorityClassName() != "" {
			podSpec["priorityClassName"] = sched.GetPriorityClassName()
		}
		if len(podSpec) > 0 {
			values["podSpec"] = podSpec
		}
	}

	// ---- observability ----------------------------------------------------------
	if spec.GetServiceMonitorEnabled() {
		values["serviceMonitor"] = map[string]interface{}{"enabled": true}
	}

	// ---- image ---------------------------------------------------------------------
	// The chart's separated image fields are image.registry / repository /
	// tag. It FAILS when any separated field is set while repository is
	// empty, so the module resolves the spec's documented repository
	// default ("neo4j") whenever the block renders anything.
	if img := spec.GetImage(); img != nil &&
		(img.GetRegistry() != "" || img.GetRepository() != "" || img.GetTag() != "") {
		repository := img.GetRepository()
		if repository == "" {
			repository = "neo4j"
		}
		image := map[string]interface{}{"repository": repository}
		if img.GetRegistry() != "" {
			image["registry"] = img.GetRegistry()
		}
		if img.GetTag() != "" {
			image["tag"] = img.GetTag()
		}
		values["image"] = image
	}

	// ---- escape hatch (merged LAST, helm -f semantics) --------------------------
	if spec.GetHelmValues() != "" {
		overrides := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(spec.GetHelmValues()), &overrides); err != nil {
			return nil, errors.Wrap(err, "failed to parse helm_values as a YAML document")
		}
		values = mergeMaps(values, overrides)
	}

	return values, nil
}

// sslScopeMap renders one ssl scope: both key and certificate come from the
// same Secret, read with the chart's default subPaths (private.key /
// public.crt).
func sslScopeMap(secretName string) map[string]interface{} {
	return map[string]interface{}{
		"privateKey":        map[string]interface{}{"secretName": secretName},
		"publicCertificate": map[string]interface{}{"secretName": secretName},
	}
}
