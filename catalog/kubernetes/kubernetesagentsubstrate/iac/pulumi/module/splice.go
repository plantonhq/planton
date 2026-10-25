package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Upstream's installer fills some settings into its manifests as text, at
// comment markers (cmd/ate-setup/internal/steps/overlay.go): a marker line is
// replaced by a block indented like the marker, or removed when the block is
// empty, and exactly the expected number of markers must be found. The
// markers the module meets sit inside ConfigMap data (Envoy's and
// agentgateway's configuration), which survives YAML decoding, so the
// module replaces them the same way. The one marker in container arguments
// (#ATE_EGRESS_INJECT_FLAGS) is a YAML comment; its flags are appended to
// the arguments instead (patchEgress).

// agentgatewayEgressConfig is the egress gateway's agentgateway
// configuration, which carries the credential-provider marker.
const agentgatewayEgressConfig = "atenet-egress-agentgateway-substrate-config"

// replaceMarker replaces each line whose trimmed text starts with marker by
// block, indented like the marker line, and refuses unless exactly want
// lines matched.
func replaceMarker(text, marker, block string, want int) (string, error) {
	var out []string
	found := 0
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), marker) {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			if block != "" {
				for _, l := range strings.Split(block, "\n") {
					if l == "" {
						out = append(out, "")
						continue
					}
					out = append(out, indent+l)
				}
			}
			found++
			continue
		}
		out = append(out, line)
	}
	if found != want {
		return "", errors.Errorf("Agent Substrate %s's configuration carries %d %s markers, not %d; the module's pin needs a catalog fix", vars.Release, found, marker, want)
	}
	return strings.Join(out, "\n"), nil
}

// spliceExtProc adds the additional ext_proc filter to both of Envoy's
// intercepting filter chains, and its cluster, in the egress gateway's Envoy
// configuration (upstream's patchAtenetEgressManifest).
func spliceExtProc(d document, s *settings) error {
	data := child(d, "data")
	envoy, _ := data["envoy.yaml"].(string)
	address := fmt.Sprintf("%s.%s.svc.cluster.local", s.egress.extprocService, s.egress.extprocNamespace)
	serverName := fmt.Sprintf("%s.%s.svc", s.egress.extprocService, s.egress.extprocNamespace)
	envoy, err := replaceMarker(envoy, "#ATE_MITM_EXTPROC_FILTER", extProcFilter, 2)
	if err != nil {
		return err
	}
	envoy, err = replaceMarker(envoy, "#ATE_MITM_EXTPROC_CLUSTER", fmt.Sprintf(extProcCluster, serverName, serverName, address, s.egress.extprocPort), 1)
	if err != nil {
		return err
	}
	data["envoy.yaml"] = envoy
	return nil
}

// spliceAgentgatewayProviders names the credential provider in the
// agentgateway egress policy, or removes the marker when injection is off
// (upstream's patchAgentgatewayEgressInject).
func spliceAgentgatewayProviders(d document, s *settings) error {
	data := child(d, "data")
	config, _ := data["config.yaml"].(string)
	block := ""
	if s.egress.provider != "disabled" {
		block = fmt.Sprintf(agentgatewayProviders, jsonQuote(s.egress.providerName), jsonQuote(s.egress.providerAddress), jsonQuote(s.egress.providerServerName()))
	}
	config, err := replaceMarker(config, "#ATE_AGENTGATEWAY_CREDENTIAL_PROVIDERS", block, 1)
	if err != nil {
		return err
	}
	data["config.yaml"] = config
	return nil
}

// applyAgentgateway swaps both gateways' Envoy dataplane for agentgateway
// with upstream's own components: components/agentgateway (its two
// ConfigMaps, and patches to the router and egress Deployments and the
// router Service), then components/agentgateway-egress (the egress
// configuration, and dropping Envoy's certificate minter). Upstream applies
// them in that order.
func applyAgentgateway(system []document, f releaseFileSet) ([]document, error) {
	configs, err := splitDocuments(f.files[fileAgentgatewayConfig])
	if err != nil {
		return nil, err
	}
	system = append(system, configs...)
	for _, step := range []struct {
		path     string
		required []string
	}{
		{fileAgentgateway, []string{"Deployment/" + vars.RouterName, "Deployment/" + vars.EgressName}},
		{fileAgentgatewayEgress, []string{"Deployment/" + vars.EgressName, "ConfigMap/" + agentgatewayEgressConfig}},
	} {
		component, err := parseKustomization(f.files[step.path])
		if err != nil {
			return nil, err
		}
		if system, err = applyComponent(component, system, step.required...); err != nil {
			return nil, err
		}
	}
	return system, nil
}

// applyCloudSQLSidecar adds upstream's Cloud SQL Auth Proxy native sidecar
// to the API server (cloudsql/proxy-sidecar-patch.yaml, a strategic-merge
// patch over the Deployment).
func applyCloudSQLSidecar(system []document, patchText string) ([]document, error) {
	for _, d := range system {
		if str(d, "kind") == "Deployment" && str(d, "metadata", "name") == vars.APIServerName {
			if _, err := applyPatchText(d, patchText); err != nil {
				return nil, errors.Wrap(err, "adding the Cloud SQL Auth Proxy sidecar")
			}
			return system, nil
		}
	}
	return nil, errors.Errorf("Agent Substrate %s has no %s Deployment for the Cloud SQL sidecar", vars.Release, vars.APIServerName)
}

// mirrorImages moves every image a document runs to the mirror registry,
// keeping each image's path, tag and digest. A Docker Hub name without a
// registry ("postgres:18-alpine", "envoyproxy/envoy:v1") is read the way
// container runtimes read it (docker.io, library/ for a single name).
func mirrorImages(d document, registry string) {
	if registry == "" {
		return
	}
	if !isWorkload(d) {
		return
	}
	spec := podSpec(d)
	for _, field := range []string{"initContainers", "containers"} {
		items, _ := spec[field].([]interface{})
		for _, item := range items {
			if c, ok := item.(map[string]interface{}); ok {
				if image, ok := c["image"].(string); ok {
					c["image"] = mirrored(image, registry)
				}
			}
		}
	}
}

// mirrored is one image moved to registry.
func mirrored(image, registry string) string {
	first, rest, found := strings.Cut(image, "/")
	switch {
	case !found:
		rest = "library/" + image
	case strings.ContainsAny(first, ".:") || first == "localhost":
	default:
		rest = image
	}
	return registry + "/" + rest
}

// extProcFilter is upstream's additional ext_proc filter (overlay.go
// emitAdditionalEgressExtprocFilter).
const extProcFilter = `- name: envoy.filters.http.ext_proc
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.filters.http.ext_proc.v3.ExternalProcessor
    grpc_service:
      envoy_grpc:
        cluster_name: additional_egress_ext_proc
      timeout: 2s
    failure_mode_allow: false
    message_timeout: 2s
    request_attributes:
    - filter_state['dev.ate.actor.identity']
    processing_mode:
      request_header_mode: SEND
      response_header_mode: SKIP
      request_body_mode: NONE
      response_body_mode: NONE
      request_trailer_mode: SKIP
      response_trailer_mode: SKIP
    mutation_rules:
      disallow_system: true
      disallow_is_error: true`

// extProcCluster is upstream's additional ext_proc cluster (overlay.go
// emitAdditionalEgressExtprocCluster): sni, SAN, address, port.
const extProcCluster = `- name: additional_egress_ext_proc
  type: STRICT_DNS
  lb_policy: ROUND_ROBIN
  connect_timeout: 1s
  typed_extension_protocol_options:
    envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
      "@type": type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
      explicit_http_config:
        http2_protocol_options: {}
  transport_socket:
    name: envoy.transport_sockets.tls
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
      sni: %s
      common_tls_context:
        tls_params:
          tls_minimum_protocol_version: TLSv1_3
          tls_maximum_protocol_version: TLSv1_3
        tls_certificate_sds_secret_configs:
        - name: podidentity_client_cert
          sds_config:
            resource_api_version: V3
            path_config_source:
              path: /etc/envoy/sds-podidentity-cert.yaml
        combined_validation_context:
          default_validation_context:
            match_typed_subject_alt_names:
            - san_type: DNS
              matcher:
                exact: %s
          validation_context_sds_secret_config:
            name: servicedns_validation_context
            sds_config:
              resource_api_version: V3
              path_config_source:
                path: /etc/envoy/sds-servicedns-validation.yaml
  load_assignment:
    cluster_name: additional_egress_ext_proc
    endpoints:
    - lb_endpoints:
      - endpoint:
          address:
            socket_address:
              address: %s
              port_value: %d`

// agentgatewayProviders is upstream's credential-provider block for the
// agentgateway egress policy (overlay.go patchAgentgatewayEgressInject):
// name, address, server name, each quoted.
const agentgatewayProviders = `credentialProviders:
- uriAuthority: %s
  target:
    host: %s
    policies:
      backendTLS:
        hostname: %s
        cert: /run/podidentity.podcert.ate.dev/credential-bundle.pem
        key: /run/podidentity.podcert.ate.dev/credential-bundle.pem
        root: /run/servicedns-ca/trust-bundle.pem`
