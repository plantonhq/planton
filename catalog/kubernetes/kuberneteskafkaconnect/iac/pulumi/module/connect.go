package module

import (
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kuberneteskafkaconnectv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkaconnect/v1alpha1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createKafkaConnect renders the kafka.strimzi.io/v1 KafkaConnect resource
// as an UNTYPED CustomResource — same posture as the Kafka family: the CRD
// types spec.config (and the custom authentication's config block) with
// x-kubernetes-preserve-unknown-fields, which crd2pulumi flattens into
// shapes that cannot hold the free-typed bodies, so no typed Strimzi
// package is shipped at all.
//
// The spec body built here is the exact twin of the Terraform module's
// local.connect_manifest (locals.tf) — same keys rendered and omitted,
// numbers as ints, booleans as booleans. An unset optional is simply never
// inserted into the map (the Go twin of TF's null-prune), so the apiserver
// applies the CRD's own defaults.
//
// No await machinery, deliberately: worker readiness depends on the
// operator (image pulls or an operator-driven image BUILD, group
// formation) that is not part of applying the resource — the same
// never-block-on-a-controller posture as every operator-CR kind.
func createKafkaConnect(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependencies []pulumi.ResourceOption,
) (pulumi.Resource, error) {
	spec := locals.Spec

	// Replicas defaults to 1 through the platform middleware; the
	// fallback keeps a raw iac-input (offline proofs, hand runs)
	// rendering the same value the middleware would have injected.
	replicas := 1
	if spec.Replicas != nil {
		replicas = int(spec.GetReplicas())
	}

	specBody := map[string]interface{}{
		"replicas":         replicas,
		"bootstrapServers": spec.GetBootstrapServers().GetValue(),
		// The group identity quartet always renders (defaults resolved
		// in locals) — leaving them to the CRD's defaults would derive
		// them from the CR name anyway, but rendering them explicit
		// keeps the uniqueness contract visible in the applied object.
		"groupId":            locals.GroupId,
		"configStorageTopic": locals.ConfigStorageTopic,
		"statusStorageTopic": locals.StatusStorageTopic,
		"offsetStorageTopic": locals.OffsetStorageTopic,
	}
	if spec.GetVersion() != "" {
		specBody["version"] = spec.GetVersion()
	}
	if spec.GetImage() != "" {
		// image and build are mutually exclusive at validation (the
		// operator would deploy the BUILT image and silently override a
		// declared one), so this arm only renders for prebuilt-image
		// clusters.
		specBody["image"] = spec.GetImage()
	}
	if tls := tlsBody(spec.GetTls()); tls != nil {
		specBody["tls"] = tls
	}
	if authentication := authenticationBody(spec.GetAuthentication()); authentication != nil {
		specBody["authentication"] = authentication
	}
	if len(spec.GetConfig()) > 0 {
		specBody["config"] = stringMapToInterface(spec.GetConfig())
	}
	if plugins := ociPluginsBody(spec.GetPlugins()); plugins != nil {
		specBody["plugins"] = plugins
	}
	if build := buildBody(spec.GetBuild()); build != nil {
		specBody["build"] = build
	}
	if resources := resourcesMap(spec.GetResources()); resources != nil {
		specBody["resources"] = resources
	}
	if jvm := jvmBody(spec.GetJvm()); jvm != nil {
		specBody["jvmOptions"] = jvm
	}
	if spec.GetRack().GetTopologyKey() != "" {
		specBody["rack"] = map[string]interface{}{
			"topologyKey": spec.GetRack().GetTopologyKey(),
		}
	}
	if spec.GetMetrics().GetEnabled() {
		// The module owns the rules ConfigMap (metrics_rules.go); the CR
		// only points at it.
		specBody["metricsConfig"] = map[string]interface{}{
			"type": "jmxPrometheusExporter",
			"valueFrom": map[string]interface{}{
				"configMapKeyRef": map[string]interface{}{
					"name": locals.MetricsConfigMapName,
					"key":  vars.MetricsConfigKey,
				},
			},
		}
	}
	if template := templateBody(spec); template != nil {
		specBody["template"] = template
	}

	return apiextensions.NewCustomResource(ctx, locals.ConnectName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String(vars.ApiVersion),
			Kind:       pulumi.String("KafkaConnect"),
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.ConnectName),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
				// Module-owned and unconditional: this annotation is what
				// makes KubernetesKafkaConnector declarations work — the
				// operator reconciles KafkaConnector resources against
				// this cluster and reverts REST-API-made changes it does
				// not own.
				Annotations: pulumi.StringMap{
					vars.UseConnectorResourcesAnnotationKey: pulumi.String("true"),
				},
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": specBody,
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependencies...)...)
}

// tlsBody renders the client-side TLS trust block: presence of spec.tls
// enables TLS on the Kafka connection, and each trusted certificate names
// a Secret plus EXACTLY ONE of certificate (a single file) or pattern (a
// glob over the Secret's files) — the proto's XOR validation guarantees
// one and only one is set, so the body renders whichever arm is present.
func tlsBody(tls *kubernetesprovider.StrimziKafkaClientTls) map[string]interface{} {
	if tls == nil {
		return nil
	}
	certificates := make([]interface{}, 0, len(tls.GetTrustedCertificates()))
	for _, cert := range tls.GetTrustedCertificates() {
		body := map[string]interface{}{
			"secretName": cert.GetSecretName().GetValue(),
		}
		if cert.GetCertificate() != "" {
			body["certificate"] = cert.GetCertificate()
		}
		if cert.GetPattern() != "" {
			body["pattern"] = cert.GetPattern()
		}
		certificates = append(certificates, body)
	}
	return map[string]interface{}{
		"trustedCertificates": certificates,
	}
}

// authenticationBody renders the client authentication block. The type is
// verbatim; each arm renders ONLY its own credential fields (the proto's
// CEL rules guarantee the arm's fields are present):
//
//	tls                       -> certificateAndKey {secretName, certificate, key}
//	scram-sha-512/-256, plain -> username + passwordSecret {secretName, password}
//	custom                    -> sasl + config
//
// KubernetesKafkaUser credential Secrets carry user.crt/user.key/password;
// the spec defaults mirror that (twin of the TF module's coalesce
// fallbacks — optional-field getters return "" when unset).
func authenticationBody(auth *kubernetesprovider.StrimziKafkaClientAuthentication) map[string]interface{} {
	if auth == nil {
		return nil
	}
	body := map[string]interface{}{
		"type": auth.GetType(),
	}
	switch auth.GetType() {
	case "tls":
		certAndKey := auth.GetCertificateAndKey()
		certificate := certAndKey.GetCertificate()
		if certificate == "" {
			certificate = "user.crt"
		}
		key := certAndKey.GetKey()
		if key == "" {
			key = "user.key"
		}
		body["certificateAndKey"] = map[string]interface{}{
			"secretName":  certAndKey.GetSecretName().GetValue(),
			"certificate": certificate,
			"key":         key,
		}
	case "scram-sha-512", "scram-sha-256", "plain":
		password := auth.GetPasswordSecret().GetPassword()
		if password == "" {
			password = "password"
		}
		body["username"] = auth.GetUsername()
		body["passwordSecret"] = map[string]interface{}{
			"secretName": auth.GetPasswordSecret().GetSecretName().GetValue(),
			"password":   password,
		}
	case "custom":
		body["sasl"] = auth.GetSasl()
		if len(auth.GetConfig()) > 0 {
			body["config"] = stringMapToInterface(auth.GetConfig())
		}
	}
	return body
}

// ociPluginsBody renders the image-volume plugin arm: each plugin's
// artifacts are OCI references mounted directly into the workers. The
// artifact type is ALWAYS the literal "image" — it is the only artifact
// type this arm supports, so the module owns it rather than asking the
// author to repeat it.
func ociPluginsBody(plugins []*kuberneteskafkaconnectv1alpha1.KubernetesKafkaConnectOciPlugin) []interface{} {
	if len(plugins) == 0 {
		return nil
	}
	out := make([]interface{}, 0, len(plugins))
	for _, plugin := range plugins {
		artifacts := make([]interface{}, 0, len(plugin.GetArtifacts()))
		for _, artifact := range plugin.GetArtifacts() {
			body := map[string]interface{}{
				"type":      "image",
				"reference": artifact.GetReference(),
			}
			if artifact.GetPullPolicy() != "" {
				body["pullPolicy"] = artifact.GetPullPolicy()
			}
			artifacts = append(artifacts, body)
		}
		out = append(out, map[string]interface{}{
			"name":      plugin.GetName(),
			"artifacts": artifacts,
		})
	}
	return out
}

// buildBody renders the operator-driven image build arm: output names the
// registry destination (type defaults to docker — the Kubernetes path;
// imagestream is OpenShift-only), plugins declare the artifacts baked into
// the image.
func buildBody(build *kuberneteskafkaconnectv1alpha1.KubernetesKafkaConnectBuild) map[string]interface{} {
	if build == nil {
		return nil
	}

	output := build.GetOutput()
	outputType := output.GetType()
	if outputType == "" {
		outputType = "docker"
	}
	outputBody := map[string]interface{}{
		"type":  outputType,
		"image": output.GetImage(),
	}
	if output.GetPushSecret() != "" {
		outputBody["pushSecret"] = output.GetPushSecret()
	}
	if len(output.GetAdditionalBuildOptions()) > 0 {
		outputBody["additionalBuildOptions"] = stringSliceToInterface(output.GetAdditionalBuildOptions())
	}
	if len(output.GetAdditionalPushOptions()) > 0 {
		outputBody["additionalPushOptions"] = stringSliceToInterface(output.GetAdditionalPushOptions())
	}

	plugins := make([]interface{}, 0, len(build.GetPlugins()))
	for _, plugin := range build.GetPlugins() {
		artifacts := make([]interface{}, 0, len(plugin.GetArtifacts()))
		for _, artifact := range plugin.GetArtifacts() {
			artifacts = append(artifacts, buildArtifactBody(artifact))
		}
		plugins = append(plugins, map[string]interface{}{
			"name":      plugin.GetName(),
			"artifacts": artifacts,
		})
	}

	return map[string]interface{}{
		"output":  outputBody,
		"plugins": plugins,
	}
}

// buildArtifactBody renders exactly the declared type's fields — a
// url-family artifact (jar/tgz/zip/other) never carries Maven coordinates
// and a maven artifact never carries a url (the proto's CEL rules enforce
// the same partition), so the operator's schema validation sees only the
// keys its per-type sub-schema allows.
func buildArtifactBody(artifact *kuberneteskafkaconnectv1alpha1.KubernetesKafkaConnectBuildArtifact) map[string]interface{} {
	body := map[string]interface{}{
		"type": artifact.GetType(),
	}
	switch artifact.GetType() {
	case "maven":
		if artifact.GetRepository() != "" {
			body["repository"] = artifact.GetRepository()
		}
		body["group"] = artifact.GetGroup()
		body["artifact"] = artifact.GetArtifact()
		body["version"] = artifact.GetVersion()
	default: // jar, tgz, zip, other
		body["url"] = artifact.GetUrl()
		if artifact.GetSha512Sum() != "" {
			body["sha512sum"] = artifact.GetSha512Sum()
		}
		if artifact.GetInsecure() {
			body["insecure"] = true
		}
		if artifact.GetType() == "other" && artifact.GetFileName() != "" {
			body["fileName"] = artifact.GetFileName()
		}
	}
	return body
}

// templateBody renders the worker pods' scheduling knobs into Strimzi's
// pod template. The Strimzi pod template carries affinity and tolerations
// but NO nodeSelector — a node_selector map therefore translates to a
// requiredDuringSchedulingIgnoredDuringExecution nodeAffinity with one
// matchExpressions entry per label (semantically identical for exact-match
// selection; the Terraform module renders the same translation).
func templateBody(spec *kuberneteskafkaconnectv1alpha1.KubernetesKafkaConnectSpec) map[string]interface{} {
	podBody := map[string]interface{}{}

	if len(spec.GetTolerations()) > 0 {
		podBody["tolerations"] = tolerationsSlice(spec.GetTolerations())
	}

	if len(spec.GetNodeSelector()) > 0 {
		matchExpressions := make([]interface{}, 0, len(spec.GetNodeSelector()))
		// Sorted iteration keeps the rendered CR deterministic across
		// runs (Go map order is random; TF for-expressions sort keys).
		for _, key := range sortedKeys(spec.GetNodeSelector()) {
			matchExpressions = append(matchExpressions, map[string]interface{}{
				"key":      key,
				"operator": "In",
				"values":   []interface{}{spec.GetNodeSelector()[key]},
			})
		}
		podBody["affinity"] = map[string]interface{}{
			"nodeAffinity": map[string]interface{}{
				"requiredDuringSchedulingIgnoredDuringExecution": map[string]interface{}{
					"nodeSelectorTerms": []interface{}{
						map[string]interface{}{
							"matchExpressions": matchExpressions,
						},
					},
				},
			},
		}
	}

	if len(podBody) == 0 {
		return nil
	}
	return map[string]interface{}{
		"pod": podBody,
	}
}

func tolerationsSlice(tolerations []*kubernetesprovider.WorkloadToleration) []interface{} {
	out := make([]interface{}, 0, len(tolerations))
	for _, t := range tolerations {
		tol := map[string]interface{}{}
		if t.GetKey() != "" {
			tol["key"] = t.GetKey()
		}
		if t.GetOperator() != "" {
			tol["operator"] = t.GetOperator()
		}
		if t.GetValue() != "" {
			tol["value"] = t.GetValue()
		}
		if t.GetEffect() != "" {
			tol["effect"] = t.GetEffect()
		}
		if t.TolerationSeconds != nil {
			tol["tolerationSeconds"] = t.GetTolerationSeconds()
		}
		out = append(out, tol)
	}
	return out
}

func jvmBody(jvm *kuberneteskafkaconnectv1alpha1.KubernetesKafkaConnectJvm) map[string]interface{} {
	if jvm == nil {
		return nil
	}
	body := map[string]interface{}{}
	if jvm.GetXms() != "" {
		body["-Xms"] = jvm.GetXms()
	}
	if jvm.GetXmx() != "" {
		body["-Xmx"] = jvm.GetXmx()
	}
	if len(body) == 0 {
		return nil
	}
	return body
}

// resourcesMap renders the shared ContainerResources message into the
// standard Kubernetes limits/requests shape. Returns nil when nothing is
// set.
func resourcesMap(r *kubernetesprovider.ContainerResources) map[string]interface{} {
	if r == nil {
		return nil
	}
	out := map[string]interface{}{}
	if l := r.GetLimits(); l != nil && (l.GetCpu() != "" || l.GetMemory() != "") {
		limits := map[string]interface{}{}
		if l.GetCpu() != "" {
			limits["cpu"] = l.GetCpu()
		}
		if l.GetMemory() != "" {
			limits["memory"] = l.GetMemory()
		}
		out["limits"] = limits
	}
	if q := r.GetRequests(); q != nil && (q.GetCpu() != "" || q.GetMemory() != "") {
		requests := map[string]interface{}{}
		if q.GetCpu() != "" {
			requests["cpu"] = q.GetCpu()
		}
		if q.GetMemory() != "" {
			requests["memory"] = q.GetMemory()
		}
		out["requests"] = requests
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stringMapToInterface(in map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func stringSliceToInterface(in []string) []interface{} {
	out := make([]interface{}, 0, len(in))
	for _, v := range in {
		out = append(out, v)
	}
	return out
}

func sortedKeys(in map[string]string) []string {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
