package module

// Output keys are defined in locals.go as constants (OpNamespace, OpPipelineVersion, etc.)
// This file documents the stack outputs for reference.

// Stack Outputs:
//
// - namespace: The namespace where Tekton is installed (always "tekton-pipelines")
// - pipeline_version: The version of Tekton Pipelines deployed
// - dashboard_version: The version of Tekton Dashboard deployed (if enabled)
// - dashboard_internal_endpoint: Internal cluster endpoint for the dashboard
// - dashboard_external_hostname: External hostname for the dashboard (if ingress enabled)
// - port_forward_dashboard_command: kubectl port-forward command for local access
// - cloud_events_sink_url: The CloudEvents sink URL (if configured)
