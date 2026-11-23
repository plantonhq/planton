package module

import (
	"fmt"

	kubernetesstigmerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstigmer/v1alpha1"
)

// Output names, the proto field names of KubernetesStigmerOutputs.
const (
	OpNamespace            = "namespace"
	OpOperatorNamespace    = "operator_namespace"
	OpOperatorRelease      = "operator_release"
	OpServerEndpoint       = "server_endpoint"
	OpArtifactsEndpoint    = "artifacts_endpoint"
	OpUrl                  = "url"
	OpPortForwardCommand   = "port_forward_command"
	OpStatusCommand        = "status_command"
	OpOwnerPasswordCommand = "owner_password_command"
	OpKeysSecret           = "keys_secret"
	OpKeysExportCommand    = "keys_export_command"
)

// outputs derives every output from the declaration: the operator names
// everything after the Stigmer, so none needs a read of the cluster. The
// Terraform twin is outputs.tf.
func outputs(locals *Locals) map[string]string {
	ns := locals.Namespace
	url := fmt.Sprintf("http://localhost:%d", vars.ServerPort)
	if hostname := locals.Spec.GetIngress().GetHostname(); hostname != "" {
		url = "https://" + hostname
	}
	ownerPasswordCommand := ""
	if bundledSignIn(locals.Spec) {
		ownerPasswordCommand = fmt.Sprintf("kubectl -n %s get secret %s-owner-setup -o jsonpath='{.data.password}' | base64 -d", ns, vars.StigmerName)
	}
	return map[string]string{
		OpNamespace:            ns,
		OpOperatorNamespace:    locals.OperatorNamespace,
		OpOperatorRelease:      vars.OperatorRelease,
		OpServerEndpoint:       fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", vars.StigmerName, ns, vars.ServerPort),
		OpArtifactsEndpoint:    fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", vars.StigmerName, ns, vars.ArtifactsPort),
		OpUrl:                  url,
		OpPortForwardCommand:   fmt.Sprintf("kubectl -n %s port-forward svc/%s %d:%d %d:%d", ns, vars.StigmerName, vars.ServerPort, vars.ServerPort, vars.ArtifactsPort, vars.ArtifactsPort),
		OpStatusCommand:        fmt.Sprintf("kubectl -n %s get stigmer %s", ns, vars.StigmerName),
		OpOwnerPasswordCommand: ownerPasswordCommand,
		OpKeysSecret:           vars.KeysSecret,
		OpKeysExportCommand:    keysExportCommand(ns),
	}
}

// bundledSignIn is Stigmer's own rule for when the bundled sign-in runs: an
// explicit identity.bundled, or a front door with no identity and without
// allow_unauthenticated_exposure.
func bundledSignIn(spec *kubernetesstigmerv1alpha1.KubernetesStigmerSpec) bool {
	if spec.GetIdentity() != nil {
		return spec.GetIdentity().GetBundled() != nil
	}
	return spec.GetIngress() != nil && !spec.GetServer().GetAllowUnauthenticatedExposure()
}

// keysExportCommand saves the keys Secret as a manifest that carries only its
// name and data, so `kubectl -n <namespace> apply -f stigmer-keys.yaml`
// recreates it in any cluster (a plain `-o yaml` would carry this cluster's
// uid and resourceVersion, which a create refuses).
func keysExportCommand(namespace string) string {
	return fmt.Sprintf(`kubectl -n %s get secret %s -o go-template='{{"apiVersion: v1\nkind: Secret\nmetadata:\n  name: %s\ntype: Opaque\ndata:\n"}}{{range $k, $v := .data}}{{printf "  %%s: %%s\n" $k $v}}{{end}}' > %s.yaml`,
		namespace, vars.KeysSecret, vars.KeysSecret, vars.KeysSecret)
}
