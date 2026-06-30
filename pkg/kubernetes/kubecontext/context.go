package kubecontext

import (
	"github.com/plantonhq/planton/pkg/kubernetes/kuberneteslabels"
	"github.com/plantonhq/planton/pkg/reflection/metadatareflect"
	"google.golang.org/protobuf/proto"
)

// ExtractFromManifest extracts the kubectl context from manifest labels.
// Returns:
//   - The context name if the label exists
//   - Empty string if the label is not present (uses default context from kubeconfig)
func ExtractFromManifest(manifest proto.Message) string {
	labels := metadatareflect.ExtractLabels(manifest)
	if labels == nil {
		return ""
	}

	context, ok := labels[kuberneteslabels.KubeContextLabelKey]
	if !ok {
		return ""
	}

	return context
}
