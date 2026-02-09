package module

// Namespace Behavior for KubernetesTekton
//
// Tekton manages its own namespace ("tekton-pipelines") through the official
// release manifests applied in tekton.go. There is no create_namespace flag
// because the namespace is embedded in the Tekton Pipeline release YAML.
//
// Therefore, no standalone namespace() function is needed for this component.
// The namespace dependency is implicitly handled by the pipeline manifests
// resource ordering in main.go.
