package verify

import (
	"context"
	"strings"
)

// ResourceVerifier knows how to verify a specific Kubernetes resource type.
type ResourceVerifier interface {
	VerifyExists(ctx context.Context, kubeconfig string) error
	VerifyAbsent(ctx context.Context, kubeconfig string) error
}

// operatorKinds lists manifest kind values (lowercased) for operator/controller
// components. Operators install CRD controllers that watch resources but typically
// do not expose a Kubernetes Service. Verification checks namespace + running
// pods only (no service requirement).
var operatorKinds = map[string]bool{
	// Tier 2/3 fixture operators (tested since sessions 3-9)
	"kubernetesperconamongooperator":    true,
	"kubernetesperconamysqloperator":    true,
	"kubernetesperconapostgresoperator": true,
	"kubernetessolroperator":            true,
	"kubernetesaltinityoperator":        true,
	"kuberneteselasticoperator":         true,
	"kubernetesstrimzikafkaoperator":    true,
	"kuberneteszalandopostgresoperator": true,
	// Tier 4 operators with configurable namespace (session 010)
	"kubernetesexternalsecrets":             true,
	"kubernetesgharunnerscalesetcontroller": true,
	"kubernetesrookcephoperator":            true,
}

// crdWorkloadKinds lists manifest kind values (lowercased) for Tier 3
// operator-dependent components. These create Custom Resources (e.g.,
// Zalando Postgresql, Strimzi Kafka, ECK Elasticsearch) that are
// reconciled by their prerequisite operator into pods and services.
// Verification checks namespace + running pods + at least one service.
var crdWorkloadKinds = map[string]bool{
	"kubernetespostgres":      true,
	"kuberneteskafka":         true,
	"kuberneteselasticsearch": true,
	"kubernetesmongodb":       true,
	"kubernetessolr":          true,
	"kubernetesclickhouse":    true,
}

// helmTier2Kinds lists manifest kind values (lowercased) for Helm-based
// Kubernetes components that deploy applications with Services.
// These must match the CloudResourceKind enum names from cloud_resource_kind.proto
// (case-insensitive via lowercasing).
var helmTier2Kinds = map[string]bool{
	// Tier 2 Helm applications
	"kubernetesredis":    true,
	"kubernetesnats":     true,
	"kubernetesgrafana":  true,
	"kubernetesneo4j":    true,
	"kubernetesopenbao":  true,
	"kubernetesopenfga":  true,
	"kubernetesjenkins":  true,
	"kubernetestemporal": true,
	"kubernetesargocd":   true,
	"kubernetesharbor":   true,
	"kubernetesgitlab":   true,
	"kuberneteslocust":   true,
	"kubernetessignoz":   true,
	"kuberneteskeycloak": true,
	// Tier 4 Helm applications with configurable namespace (session 010)
	"kubernetesingressnginx": true,
	"kubernetesistio":        true,
}

// crdInstallKinds maps manifest kind values (lowercased) to their expected CRD
// names for components that only install cluster-scoped CRDs without deploying
// any pods or services.
var crdInstallKinds = map[string][]string{
	"kubernetesgatewayapicrds": {
		"gatewayclasses.gateway.networking.k8s.io",
		"gateways.gateway.networking.k8s.io",
		"httproutes.gateway.networking.k8s.io",
		"referencegrants.gateway.networking.k8s.io",
	},
}

// GetVerifierFromManifest creates the appropriate verifier by parsing the manifest.
func GetVerifierFromManifest(manifestPath string) (ResourceVerifier, error) {
	info, err := ParseManifestInfo(manifestPath)
	if err != nil {
		return nil, err
	}

	component := strings.ToLower(info.Kind)

	switch component {
	case "kubernetesnamespace":
		return &NamespaceVerifier{Name: info.Name}, nil

	case "kubernetesdeployment":
		return &WorkloadVerifier{
			Namespace: info.Namespace,
			Kind:      "deployment",
			Name:      info.Name,
		}, nil

	case "kubernetesstatefulset":
		return &WorkloadVerifier{
			Namespace: info.Namespace,
			Kind:      "statefulset",
			Name:      info.Name,
		}, nil

	case "kubernetessecret":
		return &ResourceExistenceVerifier{
			Namespace: info.Namespace,
			Kind:      "secret",
			Name:      info.Name,
		}, nil

	case "kubernetesservice":
		return &ResourceExistenceVerifier{
			Namespace: info.Namespace,
			Kind:      "service",
			Name:      info.Name,
		}, nil

	case "kubernetescronjob":
		return &ResourceExistenceVerifier{
			Namespace: info.Namespace,
			Kind:      "cronjob",
			Name:      info.Name,
		}, nil

	case "kubernetesjob":
		return &JobVerifier{
			Namespace: info.Namespace,
			Name:      info.Name,
		}, nil

	case "kubernetesdaemonset":
		return &WorkloadVerifier{
			Namespace: info.Namespace,
			Kind:      "daemonset",
			Name:      info.Name,
		}, nil

	case "kubernetesmanifest":
		return &ConfigGroupVerifier{
			Namespace:    info.Namespace,
			ManifestPath: manifestPath,
		}, nil

	// Fixed-namespace components: the proto spec has no namespace field because
	// the upstream tooling (Tekton, etc.) uses hardcoded namespaces. The manifest
	// YAML cannot carry a namespace hint because protojson.Unmarshal rejects
	// unknown fields. Namespace is therefore embedded here.
	case "kubernetestekton":
		return &HelmComponentVerifier{
			Namespace:     "tekton-pipelines",
			ComponentName: info.Name,
		}, nil

	case "kubernetestektonoperator":
		return &OperatorComponentVerifier{
			Namespace:     "tekton-operator",
			ComponentName: info.Name,
		}, nil

	default:
		if crdNames, ok := crdInstallKinds[component]; ok {
			return &CRDInstallVerifier{
				ComponentName: info.Name,
				CRDNames:      crdNames,
			}, nil
		}
		if operatorKinds[component] {
			return &OperatorComponentVerifier{
				Namespace:     info.Namespace,
				ComponentName: info.Name,
			}, nil
		}
		if crdWorkloadKinds[component] {
			return &CRDWorkloadVerifier{
				Namespace:     info.Namespace,
				ComponentName: info.Name,
			}, nil
		}
		if helmTier2Kinds[component] {
			return &HelmComponentVerifier{
				Namespace:     info.Namespace,
				ComponentName: info.Name,
			}, nil
		}
		return &GenericVerifier{Component: component}, nil
	}
}
