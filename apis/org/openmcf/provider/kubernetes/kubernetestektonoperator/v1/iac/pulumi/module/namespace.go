package module

// Namespace Behavior for KubernetesTektonOperator
//
// The Tekton Operator manages its own namespaces:
// - 'tekton-operator' for the operator itself
// - 'tekton-pipelines' for Tekton components (Pipelines, Triggers, Dashboard)
//
// These namespaces are automatically created by the Tekton Operator release
// manifests and cannot be customized. There is no create_namespace flag.
//
// Therefore, no standalone namespace() function is needed for this component.
// The namespace dependency is implicitly handled by the operator manifests
// resource ordering in tekton_operator.go.
