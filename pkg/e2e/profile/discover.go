package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"

	componentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/qa/componente2eprofile/v1"
	providerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/qa/providere2eprofile/v1"
	sharedpb "github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

// ComponentEntry pairs a component name with its loaded profile.
type ComponentEntry struct {
	Name    string
	Profile *componentv1.ComponentE2EProfile
}

// FilterOpts controls which components are included in discovery.
type FilterOpts struct {
	// Only include components with this status. Empty means all.
	Status componentv1.ComponentE2EProfileSpec_Status
	// Only include components in this tier. 0 means all.
	Tier int32
	// Only include components that have been validated with this provisioner. 0 means all.
	Provisioner sharedpb.IacProvisioner
}

// DiscoverResult holds the full discovery output for a provider.
type DiscoverResult struct {
	Provider   *providerv1.ProviderE2EProfile
	Components []ComponentEntry
}

// Discover scans all component E2E profiles under a provider and applies filters.
func Discover(repoRoot, provider string, opts FilterOpts) (*DiscoverResult, error) {
	pp, err := LoadProviderProfile(repoRoot, provider)
	if err != nil {
		return nil, err
	}

	provDir := ProviderDir(repoRoot, provider)
	entries, err := os.ReadDir(provDir)
	if err != nil {
		return nil, errors.Wrapf(err, "reading provider directory %s", provDir)
	}

	var components []ComponentEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		componentName := entry.Name()

		profilePath := ComponentProfilePath(repoRoot, provider, componentName)
		if _, err := os.Stat(profilePath); err != nil {
			continue
		}

		cp, err := LoadComponentProfile(repoRoot, provider, componentName)
		if err != nil {
			return nil, err
		}

		if !matchesFilter(cp, opts) {
			continue
		}

		components = append(components, ComponentEntry{
			Name:    componentName,
			Profile: cp,
		})
	}

	sort.Slice(components, func(i, j int) bool {
		ti, tj := components[i].Profile.Spec.Tier, components[j].Profile.Spec.Tier
		if ti != tj {
			return ti < tj
		}
		return components[i].Name < components[j].Name
	})

	return &DiscoverResult{Provider: pp, Components: components}, nil
}

func matchesFilter(cp *componentv1.ComponentE2EProfile, opts FilterOpts) bool {
	spec := cp.Spec
	if spec == nil {
		return false
	}

	if opts.Status != componentv1.ComponentE2EProfileSpec_status_unspecified {
		if spec.Status != opts.Status {
			return false
		}
	}

	if opts.Tier > 0 && spec.Tier != opts.Tier {
		return false
	}

	if opts.Provisioner != sharedpb.IacProvisioner_iac_provisioner_unspecified {
		found := false
		for _, vp := range spec.ValidatedProvisioners {
			if vp == opts.Provisioner {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// MatrixCell represents one GitHub Actions matrix entry.
type MatrixCell struct {
	Name           string `json:"name"`
	Tier           int32  `json:"tier"`
	Engine         string `json:"engine"`
	Timeout        int32  `json:"timeout"`
	RunRegex       string `json:"run_regex"`
	ComponentCount int    `json:"component_count"`
}

// Matrix is the top-level GitHub Actions matrix JSON structure.
type Matrix struct {
	Include []MatrixCell `json:"include"`
}

// BuildGitHubMatrix generates the GitHub Actions matrix JSON from discovery results.
// Groups components by tier and provisioner, constructs -run regexes for go test.
func BuildGitHubMatrix(result *DiscoverResult) *Matrix {
	type groupKey struct {
		tier        int32
		provisioner sharedpb.IacProvisioner
	}

	groups := make(map[groupKey][]string)
	timeouts := make(map[groupKey]int32)

	for _, ce := range result.Components {
		spec := ce.Profile.Spec
		if spec == nil || spec.Status != componentv1.ComponentE2EProfileSpec_green {
			continue
		}

		for _, vp := range spec.ValidatedProvisioners {
			key := groupKey{tier: spec.Tier, provisioner: vp}
			groups[key] = append(groups[key], ce.Name)
			if spec.TimeoutMinutes > timeouts[key] {
				timeouts[key] = spec.TimeoutMinutes
			}
		}
	}

	var cells []MatrixCell
	for key, names := range groups {
		engineName := strings.ToLower(sharedpb.IacProvisioner_name[int32(key.provisioner)])
		tierTimeout := timeouts[key]
		if tierTimeout < 15 {
			tierTimeout = 15
		}
		// Buffer: at least 15 min overhead for kind setup/teardown + per-component overhead
		groupTimeout := tierTimeout*int32(len(names)) + 15

		runRegex := buildRunRegex(names, engineName)

		cells = append(cells, MatrixCell{
			Name:           fmt.Sprintf("Tier %d %s", key.tier, capitalize(engineName)),
			Tier:           key.tier,
			Engine:         engineName,
			Timeout:        groupTimeout,
			RunRegex:       runRegex,
			ComponentCount: len(names),
		})
	}

	sort.Slice(cells, func(i, j int) bool {
		if cells[i].Tier != cells[j].Tier {
			return cells[i].Tier < cells[j].Tier
		}
		return cells[i].Engine < cells[j].Engine
	})

	return &Matrix{Include: cells}
}

// MatrixJSON returns the GitHub Actions matrix as a JSON string.
func MatrixJSON(m *Matrix) (string, error) {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", errors.Wrap(err, "marshaling matrix JSON")
	}
	return string(b), nil
}

// buildRunRegex constructs a go test -run regex that matches all test functions
// for the given components and engine. Component names are converted to PascalCase
// Go test function names (e.g., "kubernetesredis" -> "KubernetesRedis").
func buildRunRegex(components []string, engine string) string {
	var parts []string
	for _, name := range components {
		parts = append(parts, toPascalCase(name))
	}
	engineSuffix := capitalize(engine)
	return fmt.Sprintf("Test(%s)_%s", strings.Join(parts, "|"), engineSuffix)
}

// toPascalCase converts a lowercase component name to PascalCase for Go test
// function matching. Handles the "kubernetes" prefix and known component name patterns.
func toPascalCase(name string) string {
	if name == "" {
		return ""
	}

	// Known prefixes that should be capitalized as a single word
	knownPrefixes := []struct {
		lower  string
		pascal string
	}{
		{"kubernetesnamespace", "KubernetesNamespace"},
		{"kubernetesdeployment", "KubernetesDeployment"},
		{"kubernetesstatefulset", "KubernetesStatefulSet"},
		{"kubernetessecret", "KubernetesSecret"},
		{"kubernetesservice", "KubernetesService"},
		{"kubernetescronjob", "KubernetesCronJob"},
		{"kubernetesjob", "KubernetesJob"},
		{"kubernetesdaemonset", "KubernetesDaemonSet"},
		{"kubernetesmanifest", "KubernetesManifest"},
		{"kubernetesredis", "KubernetesRedis"},
		{"kubernetesgrafana", "KubernetesGrafana"},
		{"kubernetesopenbao", "KubernetesOpenBao"},
		{"kubernetesargocd", "KubernetesArgoCD"},
		{"kuberneteslocust", "KubernetesLocust"},
		{"kubernetesnats", "KubernetesNats"},
		{"kubernetesneo4j", "KubernetesNeo4j"},
		{"kubernetesjenkins", "KubernetesJenkins"},
		{"kubernetessolroperator", "KubernetesSolrOperator"},
		{"kubernetesperconamongooperator", "KubernetesPerconaMongoOperator"},
		{"kubernetesperconamysqloperator", "KubernetesPerconaMysqlOperator"},
		{"kubernetesperconapostgresoperator", "KubernetesPerconaPostgresOperator"},
		{"kubernetesgitlab", "KubernetesGitlab"},
		{"kubernetespostgres", "KubernetesPostgres"},
		{"kuberneteskafka", "KubernetesKafka"},
		{"kuberneteselasticsearch", "KubernetesElasticsearch"},
		{"kubernetesmongodb", "KubernetesMongodb"},
		{"kubernetessolr", "KubernetesSolr"},
		{"kubernetesclickhouse", "KubernetesClickHouse"},
		{"kuberneteszalandopostgresoperator", "KubernetesZalandoPostgresOperator"},
		{"kubernetesstrimzikafkaoperator", "KubernetesStrimziKafkaOperator"},
		{"kuberneteselasticoperator", "KubernetesElasticOperator"},
		{"kubernetesaltinityoperator", "KubernetesAltinityOperator"},
		{"kubernetesgatewayapicrds", "KubernetesGatewayApiCrds"},
		{"kubernetesgatewayclass", "KubernetesGatewayClass"},
		{"kubernetesgateway", "KubernetesGateway"},
		{"kuberneteshttproute", "KubernetesHttpRoute"},
		{"kubernetesgrpcroute", "KubernetesGrpcRoute"},
		{"kubernetestcproute", "KubernetesTcpRoute"},
		{"kubernetestlsroute", "KubernetesTlsRoute"},
		{"kubernetesreferencegrant", "KubernetesReferenceGrant"},
		{"kubernetesistiobasecrds", "KubernetesIstioBaseCrds"},
		{"kubernetespeerauthentication", "KubernetesPeerAuthentication"},
		{"kubernetesrequestauthentication", "KubernetesRequestAuthentication"},
		{"kubernetesgharunnerscalesetcontroller", "KubernetesGhaRunnerScaleSetController"},
		{"kubernetesrookcephoperator", "KubernetesRookCephOperator"},
		{"kubernetesexternalsecrets", "KubernetesExternalSecrets"},
		{"kubernetesingressnginx", "KubernetesIngressNginx"},
		{"kubernetestekton", "KubernetesTekton"},
		{"kubernetestektonoperator", "KubernetesTektonOperator"},
		{"kubernetesistio", "KubernetesIstio"},
		{"kuberneteshelmrelease", "KubernetesHelmRelease"},
		{"kubernetescertmanager", "KubernetesCertManager"},
		{"kubernetesexternaldns", "KubernetesExternalDns"},
		{"kubernetesgharunnerscaleset", "KubernetesGhaRunnerScaleSet"},
		{"kubernetesrookcephcluster", "KubernetesRookCephCluster"},
		{"kubernetesprometheus", "KubernetesPrometheus"},
		{"kuberneteskeycloak", "KubernetesKeycloak"},
		{"kubernetesopenfga", "KubernetesOpenFGA"},
		{"kubernetestemporal", "KubernetesTemporal"},
		{"kubernetesharbor", "KubernetesHarbor"},
		{"kubernetessignoz", "KubernetesSigNoz"},
	}

	for _, kp := range knownPrefixes {
		if strings.EqualFold(name, kp.lower) {
			return kp.pascal
		}
	}

	// Fallback: capitalize first letter. This works for most simple names but
	// won't handle multi-word camelCase correctly. Known prefixes above handle
	// all existing Kubernetes components.
	return strings.ToUpper(name[:1]) + name[1:]
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// StatusCounts tallies components by status.
type StatusCounts struct {
	Green    int
	Deferred int
	Skip     int
	Stub     int
	Total    int
}

// CountByStatus counts components in the discovery result by their E2E status.
func CountByStatus(result *DiscoverResult) StatusCounts {
	var sc StatusCounts
	for _, ce := range result.Components {
		sc.Total++
		if ce.Profile.Spec == nil {
			continue
		}
		switch ce.Profile.Spec.Status {
		case componentv1.ComponentE2EProfileSpec_green:
			sc.Green++
		case componentv1.ComponentE2EProfileSpec_deferred:
			sc.Deferred++
		case componentv1.ComponentE2EProfileSpec_skip:
			sc.Skip++
		case componentv1.ComponentE2EProfileSpec_stub:
			sc.Stub++
		}
	}
	return sc
}
