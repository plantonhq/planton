package module

import (
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesstigmerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstigmer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerenv"
)

// The documents this module applies, rendered as plain data so the cross-engine
// test (parity_test.go) can hold them against what the Terraform module plans
// (locals.tf, chart_values, secrets and stigmer_document). RULE: a key renders
// only when the spec declares it, so the chart's and the operator's own
// defaults stay authoritative for everything else; an empty map or list is
// never written.

// helmValues is the operator chart's values. The chart's values schema is
// closed, so only its own keys appear; the image tag never does, because the
// chart derives it from its version and a second tag would run a different
// Stigmer from the one declared.
func helmValues(spec *kubernetesstigmerv1alpha1.KubernetesStigmerSpec) map[string]interface{} {
	op := spec.GetOperator()
	values := map[string]interface{}{}
	image := map[string]interface{}{}
	putString(image, "repository", op.GetImageRepository())
	putString(image, "pullPolicy", op.GetImagePullPolicy())
	if len(op.GetImagePullSecrets()) > 0 {
		image["pullSecrets"] = nameList(op.GetImagePullSecrets())
	}
	putMap(values, "image", image)
	if op.Replicas != nil {
		values["replicas"] = int(op.GetReplicas())
	}
	resources := map[string]interface{}{}
	putStringMap(resources, "requests", op.GetResources().GetRequests())
	putStringMap(resources, "limits", op.GetResources().GetLimits())
	putMap(values, "resources", resources)
	putString(values, "logLevel", op.GetLogLevel())
	putStringMap(values, "nodeSelector", op.GetNodeSelector())
	if len(op.GetTolerations()) > 0 {
		values["tolerations"] = tolerations(op.GetTolerations())
	}
	return values
}

// materializedSecret is one Secret the module writes: the values the spec
// declares, under the names the operator reads.
type materializedSecret struct {
	Name string
	Data map[string]string
}

// materializedSecrets lists the Secrets the declaration needs, in a fixed
// order. Each holds only the keys the spec gives values for: the runner maps
// every model key as optional, so a key present but empty would reach it as
// an empty setting.
func materializedSecrets(spec *kubernetesstigmerv1alpha1.KubernetesStigmerSpec) []materializedSecret {
	var out []materializedSecret
	add := func(name string, data map[string]string) {
		if len(data) > 0 {
			out = append(out, materializedSecret{Name: name, Data: data})
		}
	}
	if keys := spec.GetRunner().GetModel().GetKeys(); keys != nil {
		data := map[string]string{}
		putData(data, "ANTHROPIC_API_KEY", keys.GetAnthropicApiKey())
		putData(data, "ANTHROPIC_BASE_URL", keys.GetAnthropicBaseUrl())
		putData(data, "OPENAI_API_KEY", keys.GetOpenaiApiKey())
		putData(data, "ANTHROPIC_FOUNDRY_API_KEY", keys.GetAnthropicFoundryApiKey())
		putData(data, "AWS_BEARER_TOKEN_BEDROCK", keys.GetAwsBearerTokenBedrock())
		add(vars.ModelSecret, data)
	}
	if pair := spec.GetObjectStore().GetS3().GetCredentials().GetAccessKey(); pair != nil {
		add(vars.ObjectStoreSecret, map[string]string{
			"ACCESS_KEY_ID":     pair.GetAccessKeyId(),
			"SECRET_ACCESS_KEY": pair.GetSecretAccessKey(),
		})
	}
	if password := spec.GetDatabase().GetExternal().GetPassword().GetValue(); password != "" {
		add(vars.DatabaseSecret, map[string]string{vars.DatabasePasswordKey: password})
	}
	external := spec.GetTemporal().GetExternal()
	if key := external.GetApiKey().GetValue(); key != "" {
		add(vars.TemporalAPIKeySecret, map[string]string{vars.TemporalAPIKeyKey: key})
	}
	if pem := external.GetTls().GetPem(); pem != nil {
		data := map[string]string{}
		putData(data, vars.TemporalCAKey, pem.GetCaCertificate())
		putData(data, vars.TemporalCertKey, pem.GetClientCertificate())
		putData(data, vars.TemporalPrivateKeyKey, pem.GetClientPrivateKey())
		add(vars.TemporalTLSSecret, data)
	}
	if login := spec.GetIdentity().GetBundled().GetSmtp().GetCredentials().GetLogin(); login != nil {
		add(vars.SMTPSecret, map[string]string{"username": login.GetUsername(), "password": login.GetPassword()})
	}
	if license := spec.GetEnterprise().GetLicense().GetValue(); license != "" {
		add(vars.LicenseSecret, map[string]string{vars.LicenseKey: license})
	}
	add(vars.ServerExtraEnvSecret, containerenv.BuildSecretData(spec.GetServer().GetExtraEnv()))
	add(vars.RunnerExtraEnvSecret, containerenv.BuildSecretData(spec.GetRunner().GetExtraEnv()))
	return out
}

// stigmerAnnotations are the Stigmer's annotations: only the one that
// releases a held upgrade, set to the operator's own spelling of the version.
func stigmerAnnotations(spec *kubernetesstigmerv1alpha1.KubernetesStigmerSpec) map[string]string {
	if spec.GetUpgradeWithoutBackupTo() == "" {
		return nil
	}
	return map[string]string{vars.UpgradeWithoutBackupAnnotation: "v" + spec.GetUpgradeWithoutBackupTo()}
}

// stigmerSpec is the Stigmer resource's spec, in the definition's own keys.
func stigmerSpec(spec *kubernetesstigmerv1alpha1.KubernetesStigmerSpec) map[string]interface{} {
	out := map[string]interface{}{}

	image := map[string]interface{}{}
	putString(image, "registry", spec.GetImage().GetRegistry())
	putString(image, "pullPolicy", spec.GetImage().GetPullPolicy())
	if len(spec.GetImage().GetPullSecrets()) > 0 {
		image["pullSecrets"] = nameList(spec.GetImage().GetPullSecrets())
	}
	putMap(out, "image", image)

	if owner := spec.GetOwner(); owner != nil {
		o := map[string]interface{}{"email": owner.GetEmail()}
		putString(o, "name", owner.GetName())
		out["owner"] = o
	}

	putMap(out, "server", serverSpec(spec.GetServer()))
	putMap(out, "runner", runnerSpec(spec.GetRunner()))

	if db := spec.GetDatabase(); db != nil {
		out["database"] = map[string]interface{}{"postgresql": postgresSpec(db)}
	}
	if temporal := spec.GetTemporal(); temporal != nil {
		out["temporal"] = temporalSpec(temporal)
	}
	if s3 := spec.GetObjectStore().GetS3(); s3 != nil {
		out["objectStore"] = map[string]interface{}{"s3": s3Spec(s3)}
	}
	if ingress := spec.GetIngress(); ingress != nil {
		out["ingress"] = ingressSpec(ingress)
	}
	if identity := spec.GetIdentity(); identity != nil {
		out["identity"] = identitySpec(identity)
	}
	if enterprise := spec.GetEnterprise(); enterprise != nil {
		out["enterprise"] = enterpriseSpec(enterprise)
	}
	if sandboxes := spec.GetSandboxes(); sandboxes != nil {
		out["sandboxes"] = sandboxesSpec(sandboxes)
	}
	prerequisites := map[string]interface{}{}
	putString(prerequisites, "postgresOperator", spec.GetPrerequisites().GetPostgresOperator())
	putString(prerequisites, "postgresBackupPlugin", spec.GetPrerequisites().GetPostgresBackupPlugin())
	putString(prerequisites, "agentSandbox", spec.GetPrerequisites().GetAgentSandbox())
	putString(prerequisites, "substrate", spec.GetPrerequisites().GetSubstrate())
	putMap(out, "prerequisites", prerequisites)
	return out
}

func serverSpec(server *kubernetesstigmerv1alpha1.KubernetesStigmerServer) map[string]interface{} {
	out := map[string]interface{}{}
	putMap(out, "resources", resourceRequirements(server.GetResources()))
	putMap(out, "storage", volume(server.GetStorage()))
	putMap(out, "artifacts", volume(server.GetArtifacts()))
	if help := server.GetHelp(); help != nil {
		h := map[string]interface{}{"url": help.GetUrl()}
		putString(h, "label", help.GetLabel())
		out["help"] = h
	}
	putStrings(out, "internalAccounts", server.GetInternalAccounts())
	if server.GetAllowUnauthenticatedExposure() {
		out["allowUnauthenticatedExposure"] = true
	}
	if egress := server.GetEgress(); egress != nil {
		e := map[string]interface{}{}
		if egress.GetFenced() {
			e["fenced"] = true
		}
		putStrings(e, "allowCIDRs", egress.GetAllowCidrs())
		out["egress"] = e
	}
	putList(out, "extraEnv", extraEnv(server.GetExtraEnv(), vars.ServerExtraEnvSecret))
	return out
}

func runnerSpec(runner *kubernetesstigmerv1alpha1.KubernetesStigmerRunner) map[string]interface{} {
	out := map[string]interface{}{}
	if model := runner.GetModel(); model != nil {
		secretName := vars.ModelSecret
		if model.GetKeys() == nil {
			secretName = model.GetExistingSecretName().GetValue()
		}
		out["llm"] = map[string]interface{}{"secretName": secretName}
	}
	putMap(out, "resources", resourceRequirements(runner.GetResources()))
	putMap(out, "storage", volume(runner.GetStorage()))
	putList(out, "extraEnv", extraEnv(runner.GetExtraEnv(), vars.RunnerExtraEnvSecret))
	return out
}

func postgresSpec(db *kubernetesstigmerv1alpha1.KubernetesStigmerDatabase) map[string]interface{} {
	if external := db.GetExternal(); external != nil {
		e := map[string]interface{}{
			"host":           external.GetHost().GetValue(),
			"user":           external.GetUser(),
			"passwordSecret": passwordSecret(external.GetPassword()),
		}
		if external.Port != nil {
			e["port"] = int(external.GetPort())
		}
		putString(e, "database", external.GetDatabase())
		putString(e, "sslMode", external.GetSslMode())
		return map[string]interface{}{"external": e}
	}
	bundled := db.GetBundled()
	b := map[string]interface{}{}
	if bundled.Instances != nil {
		b["instances"] = int(bundled.GetInstances())
	}
	putMap(b, "storage", volume(bundled.GetStorage()))
	putMap(b, "resources", resourceRequirements(bundled.GetResources()))
	if backup := bundled.GetBackup(); backup != nil {
		bk := map[string]interface{}{}
		putString(bk, "schedule", backup.GetSchedule())
		putString(bk, "retention", backup.GetRetention())
		b["backup"] = bk
	}
	if recover := bundled.GetRecoverFrom(); recover != nil {
		r := map[string]interface{}{"serverName": recover.GetServerName()}
		putString(r, "targetTime", recover.GetTargetTime())
		b["recoverFrom"] = r
	}
	return map[string]interface{}{"bundled": b}
}

func passwordSecret(password *kubernetesstigmerv1alpha1.KubernetesStigmerPassword) map[string]interface{} {
	if existing := password.GetExistingSecret(); existing != nil {
		key := existing.GetKey()
		if key == "" {
			key = vars.DatabasePasswordKey
		}
		return map[string]interface{}{"name": existing.GetName().GetValue(), "key": key}
	}
	return map[string]interface{}{"name": vars.DatabaseSecret, "key": vars.DatabasePasswordKey}
}

func temporalSpec(temporal *kubernetesstigmerv1alpha1.KubernetesStigmerTemporal) map[string]interface{} {
	external := temporal.GetExternal()
	if external == nil {
		b := map[string]interface{}{}
		putMap(b, "resources", resourceRequirements(temporal.GetBundled().GetResources()))
		return map[string]interface{}{"bundled": b}
	}
	e := map[string]interface{}{"hostPort": external.GetHostPort().GetValue()}
	putString(e, "namespace", external.GetNamespace())
	if tls := external.GetTls(); tls != nil {
		t := map[string]interface{}{}
		putString(t, "serverName", tls.GetServerName())
		if pem := tls.GetPem(); pem != nil {
			t["secretName"] = vars.TemporalTLSSecret
			if pem.GetCaCertificate() != "" {
				t["caKey"] = vars.TemporalCAKey
			}
			if pem.GetClientCertificate() != "" {
				t["certKey"] = vars.TemporalCertKey
				t["keyKey"] = vars.TemporalPrivateKeyKey
			}
		}
		if existing := tls.GetExistingSecret(); existing != nil {
			t["secretName"] = existing.GetName().GetValue()
			putString(t, "caKey", existing.GetCaKey())
			putString(t, "certKey", existing.GetCertificateKey())
			putString(t, "keyKey", existing.GetPrivateKeyKey())
		}
		e["tls"] = t
	}
	if apiKey := external.GetApiKey(); apiKey != nil {
		e["apiKey"] = secretValueRef(apiKey, vars.TemporalAPIKeySecret, vars.TemporalAPIKeyKey)
	}
	return map[string]interface{}{"external": e}
}

// secretValueRef is the {name, key} the operator reads a secret value from:
// the module's own Secret when the spec gives the value, else the one named.
func secretValueRef(value *kubernetesstigmerv1alpha1.KubernetesStigmerSecretValue, secret, key string) map[string]interface{} {
	if existing := value.GetExistingSecret(); existing != nil {
		return map[string]interface{}{"name": existing.GetName().GetValue(), "key": existing.GetKey()}
	}
	return map[string]interface{}{"name": secret, "key": key}
}

func s3Spec(s3 *kubernetesstigmerv1alpha1.KubernetesStigmerS3) map[string]interface{} {
	credentials := vars.ObjectStoreSecret
	if s3.GetCredentials().GetAccessKey() == nil {
		credentials = s3.GetCredentials().GetExistingSecretName().GetValue()
	}
	out := map[string]interface{}{
		"bucket":            s3.GetBucket().GetValue(),
		"credentialsSecret": map[string]interface{}{"name": credentials},
	}
	putString(out, "region", s3.GetRegion())
	putString(out, "endpoint", s3.GetEndpoint().GetValue())
	putString(out, "runnerEndpoint", s3.GetRunnerEndpoint())
	return out
}

func ingressSpec(ingress *kubernetesstigmerv1alpha1.KubernetesStigmerIngress) map[string]interface{} {
	out := map[string]interface{}{"hostname": ingress.GetHostname()}
	putString(out, "className", ingress.GetClassName())
	putStringMap(out, "annotations", ingress.GetAnnotations())
	putStringMap(out, "serviceAnnotations", ingress.GetServiceAnnotations())
	if tls := ingress.GetTls(); tls != nil {
		t := map[string]interface{}{}
		putString(t, "secretName", tls.GetSecretName())
		if issuer := tls.GetIssuer(); issuer != nil {
			i := map[string]interface{}{"name": issuer.GetName().GetValue()}
			putString(i, "kind", issuer.GetKind())
			t["issuer"] = i
		}
		out["tls"] = t
	}
	if gateway := ingress.GetGateway(); gateway != nil {
		g := map[string]interface{}{"name": gateway.GetName().GetValue()}
		putString(g, "namespace", gateway.GetNamespace().GetValue())
		putString(g, "sectionName", gateway.GetSectionName())
		out["gatewayRef"] = g
	}
	return out
}

func identitySpec(identity *kubernetesstigmerv1alpha1.KubernetesStigmerIdentity) map[string]interface{} {
	if oidc := identity.GetOidc(); oidc != nil {
		o := map[string]interface{}{"issuer": oidc.GetIssuer(), "audience": oidc.GetAudience()}
		putString(o, "consoleClientId", oidc.GetConsoleClientId())
		putString(o, "nativeClientId", oidc.GetNativeClientId())
		putString(o, "discoveryUrl", oidc.GetDiscoveryUrl())
		return map[string]interface{}{"oidc": o}
	}
	bundled := identity.GetBundled()
	b := map[string]interface{}{}
	if domains := bundled.GetSignUpEmailDomains(); len(domains) > 0 {
		b["signUp"] = map[string]interface{}{"emailDomains": stringList(domains)}
	}
	if smtp := bundled.GetSmtp(); smtp != nil {
		s := map[string]interface{}{"host": smtp.GetHost(), "from": smtp.GetFrom()}
		if smtp.Port != nil {
			s["port"] = int(smtp.GetPort())
		}
		if smtp.GetStarttls() {
			s["starttls"] = true
		}
		if credentials := smtp.GetCredentials(); credentials != nil {
			if credentials.GetLogin() != nil {
				s["secretName"] = vars.SMTPSecret
			} else {
				s["secretName"] = credentials.GetExistingSecretName().GetValue()
			}
		}
		b["smtp"] = s
	}
	return map[string]interface{}{"bundled": b}
}

func enterpriseSpec(enterprise *kubernetesstigmerv1alpha1.KubernetesStigmerEnterprise) map[string]interface{} {
	out := map[string]interface{}{
		"license": secretValueRef(enterprise.GetLicense(), vars.LicenseSecret, vars.LicenseKey),
	}
	putStrings(out, "organizationCreators", enterprise.GetOrganizationCreators())
	if enterprise.AccessHistoryRetentionDays != nil {
		out["accessHistoryRetentionDays"] = int(enterprise.GetAccessHistoryRetentionDays())
	}
	putString(out, "outboundEgress", enterprise.GetOutboundEgress())
	return out
}

func sandboxesSpec(sandboxes *kubernetesstigmerv1alpha1.KubernetesStigmerSandboxes) map[string]interface{} {
	out := map[string]interface{}{}
	putString(out, "driver", sandboxes.GetDriver())
	if sandboxes.IdleSeconds != nil {
		out["idleSeconds"] = int(sandboxes.GetIdleSeconds())
	}
	if agentSandbox := sandboxes.GetAgentSandbox(); agentSandbox != nil {
		a := map[string]interface{}{}
		if cidrs := agentSandbox.GetEgressAllowCidrs(); len(cidrs) > 0 {
			a["egress"] = map[string]interface{}{"allowCIDRs": stringList(cidrs)}
		}
		out["agentSandbox"] = a
	}
	if substrate := sandboxes.GetSubstrate(); substrate != nil {
		s := map[string]interface{}{}
		if substrate.Workers != nil {
			s["workers"] = int(substrate.GetWorkers())
		}
		worker := map[string]interface{}{}
		putMap(worker, "resources", resourceRequirements(substrate.GetWorkerResources()))
		putStringMap(worker, "nodeSelector", substrate.GetWorkerNodeSelector())
		if len(substrate.GetWorkerTolerations()) > 0 {
			worker["tolerations"] = tolerations(substrate.GetWorkerTolerations())
		}
		putMap(s, "worker", worker)
		sandbox := map[string]interface{}{}
		if substrate.SandboxCpuMillicores != nil {
			sandbox["cpuMillicores"] = int(substrate.GetSandboxCpuMillicores())
		}
		if substrate.SandboxMemoryMib != nil {
			sandbox["memoryMiB"] = int(substrate.GetSandboxMemoryMib())
		}
		putMap(s, "sandbox", sandbox)
		if http := substrate.GetEgressHttp(); len(http) > 0 {
			s["egress"] = map[string]interface{}{"http": stringList(http)}
		}
		out["substrate"] = s
	}
	return out
}

// extraEnv renders ContainerEnv as the definition's []EnvVar: plain values
// and the Kubernetes sources as written, every secret by reference (to the
// module's Secret for a value, to the user's own for a reference). A
// value_from has already been resolved to a value by the time a module runs.
func extraEnv(env *kubernetes.ContainerEnv, secretName string) []interface{} {
	var out []interface{}
	for _, v := range env.GetVariables() {
		entry := map[string]interface{}{"name": v.GetName()}
		switch source := v.GetSource().(type) {
		case *kubernetes.EnvVar_Value:
			entry["value"] = source.Value
		case *kubernetes.EnvVar_ConfigMapKeyRef:
			ref := map[string]interface{}{"name": source.ConfigMapKeyRef.GetName(), "key": source.ConfigMapKeyRef.GetKey()}
			if source.ConfigMapKeyRef.GetOptional() {
				ref["optional"] = true
			}
			entry["valueFrom"] = map[string]interface{}{"configMapKeyRef": ref}
		case *kubernetes.EnvVar_FieldRef:
			ref := map[string]interface{}{"fieldPath": source.FieldRef.GetFieldPath()}
			putString(ref, "apiVersion", source.FieldRef.GetApiVersion())
			entry["valueFrom"] = map[string]interface{}{"fieldRef": ref}
		case *kubernetes.EnvVar_ResourceFieldRef:
			ref := map[string]interface{}{"resource": source.ResourceFieldRef.GetResource()}
			putString(ref, "containerName", source.ResourceFieldRef.GetContainerName())
			putString(ref, "divisor", source.ResourceFieldRef.GetDivisor())
			entry["valueFrom"] = map[string]interface{}{"resourceFieldRef": ref}
		default:
			continue
		}
		out = append(out, entry)
	}
	for _, s := range env.GetSecrets() {
		var ref map[string]interface{}
		switch source := s.GetSource().(type) {
		case *kubernetes.SecretEnvVar_Value:
			if source.Value == "" {
				continue
			}
			ref = map[string]interface{}{"name": secretName, "key": s.GetName()}
		case *kubernetes.SecretEnvVar_SecretRef:
			ref = map[string]interface{}{"name": source.SecretRef.GetName(), "key": source.SecretRef.GetKey()}
			if source.SecretRef.GetOptional() {
				ref["optional"] = true
			}
		default:
			continue
		}
		out = append(out, map[string]interface{}{"name": s.GetName(), "valueFrom": map[string]interface{}{"secretKeyRef": ref}})
	}
	return out
}

func resourceRequirements(r *kubernetes.KubernetesCoreV1ApiResourceRequirements) map[string]interface{} {
	out := map[string]interface{}{}
	putStringMap(out, "requests", r.GetRequests())
	putStringMap(out, "limits", r.GetLimits())
	if len(r.GetClaims()) > 0 {
		var claims []interface{}
		for _, c := range r.GetClaims() {
			claim := map[string]interface{}{"name": c.GetName()}
			putString(claim, "request", c.GetRequest())
			claims = append(claims, claim)
		}
		out["claims"] = claims
	}
	return out
}

func volume(v *kubernetesstigmerv1alpha1.KubernetesStigmerVolume) map[string]interface{} {
	out := map[string]interface{}{}
	putString(out, "size", v.GetSize())
	putString(out, "storageClassName", v.GetStorageClassName())
	return out
}

func tolerations(in []*kubernetes.KubernetesCoreV1ApiToleration) []interface{} {
	out := make([]interface{}, 0, len(in))
	for _, t := range in {
		entry := map[string]interface{}{}
		putString(entry, "key", t.GetKey())
		putString(entry, "operator", t.GetOperator())
		putString(entry, "value", t.GetValue())
		putString(entry, "effect", t.GetEffect())
		if t.TolerationSeconds != nil {
			entry["tolerationSeconds"] = int(t.GetTolerationSeconds())
		}
		out = append(out, entry)
	}
	return out
}

func nameList(names []string) []interface{} {
	out := make([]interface{}, 0, len(names))
	for _, n := range names {
		out = append(out, map[string]interface{}{"name": n})
	}
	return out
}

func stringList(in []string) []interface{} {
	out := make([]interface{}, 0, len(in))
	for _, s := range in {
		out = append(out, s)
	}
	return out
}

func putString(m map[string]interface{}, key, value string) {
	if value != "" {
		m[key] = value
	}
}

func putStrings(m map[string]interface{}, key string, values []string) {
	if len(values) > 0 {
		m[key] = stringList(values)
	}
}

func putStringMap(m map[string]interface{}, key string, values map[string]string) {
	if len(values) == 0 {
		return
	}
	out := make(map[string]interface{}, len(values))
	for k, v := range values {
		out[k] = v
	}
	m[key] = out
}

func putMap(m map[string]interface{}, key string, value map[string]interface{}) {
	if len(value) > 0 {
		m[key] = value
	}
}

func putList(m map[string]interface{}, key string, value []interface{}) {
	if len(value) > 0 {
		m[key] = value
	}
}

func putData(m map[string]string, key, value string) {
	if value != "" {
		m[key] = value
	}
}
