package module

import (
	"fmt"
	"strconv"

	kubernetesflagdv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflagd/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// Locals holds computed values derived from the IaC input. Every resolution
// has an exact twin in the Terraform module's locals.tf.
type Locals struct {
	Spec *kubernetesflagdv1alpha1.KubernetesFlagdSpec

	// Labels: the Planton governance labels, stamped on every object.
	Labels map[string]string

	// SelectorLabels select the flagd pods (Deployment, Service, PDB and
	// self-spreading topology constraints).
	SelectorLabels map[string]string

	Name      string
	Namespace string
	Image     string

	ServiceAccountName   string
	CreateServiceAccount bool

	Port           int
	ManagementPort int
	SyncPort       int
	OfrepPort      int

	Config *FlagdConfig

	SourcesSecretName  string
	FlagReaderName     string
	ServiceMonitorName string

	EvaluationEndpoint string
	SyncEndpoint       string
	OfrepEndpoint      string
	ManagementEndpoint string
	PortForwardCommand string
}

func initializeLocals(iacInput *kubernetesflagdv1alpha1.KubernetesFlagdIacInput) (*Locals, error) {
	target := iacInput.Target
	spec := target.Spec
	name := target.Metadata.Name
	namespace := spec.Namespace.GetValue()

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesFlagd.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	repository := spec.GetImage().GetRepository()
	if repository == "" {
		repository = vars.DefaultImageRepository
	}
	tag := spec.GetImage().GetTag()
	if tag == "" {
		tag = vars.DefaultImageTag
	}
	if spec.GetImage().GetFips() {
		tag += "-fips"
	}

	cfg, err := buildFlagdConfig(spec, namespace)
	if err != nil {
		return nil, err
	}
	port, managementPort, syncPort, ofrepPort := ports(spec)

	serviceAccountName := name
	createServiceAccount := true
	if existing := spec.GetServiceAccount().GetExistingName(); existing != "" {
		serviceAccountName = existing
		createServiceAccount = false
	}

	host := fmt.Sprintf("%s.%s.svc.cluster.local", name, namespace)
	return &Locals{
		Spec:   spec,
		Labels: labels,
		SelectorLabels: map[string]string{
			"app.kubernetes.io/name":     "flagd",
			"app.kubernetes.io/instance": name,
		},
		Name:                 name,
		Namespace:            namespace,
		Image:                repository + ":" + tag,
		ServiceAccountName:   serviceAccountName,
		CreateServiceAccount: createServiceAccount,
		Port:                 port,
		ManagementPort:       managementPort,
		SyncPort:             syncPort,
		OfrepPort:            ofrepPort,
		Config:               cfg,
		SourcesSecretName:    name + vars.SourcesSecretSuffix,
		FlagReaderName:       name + vars.FlagReaderSuffix,
		ServiceMonitorName:   name + vars.ServiceMonitorSuffix,
		EvaluationEndpoint:   fmt.Sprintf("%s:%d", host, port),
		SyncEndpoint:         fmt.Sprintf("%s:%d", host, syncPort),
		OfrepEndpoint:        fmt.Sprintf("http://%s:%d", host, ofrepPort),
		ManagementEndpoint:   fmt.Sprintf("http://%s:%d", host, managementPort),
		PortForwardCommand:   fmt.Sprintf("kubectl port-forward -n %s svc/%s %d:%d", namespace, name, ofrepPort, ofrepPort),
	}, nil
}
