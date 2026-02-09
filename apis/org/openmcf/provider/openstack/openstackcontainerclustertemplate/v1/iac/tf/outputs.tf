output "template_id" {
  description = "The unique identifier (UUID) of the cluster template"
  value       = openstack_containerinfra_clustertemplate_v1.main.id
}

output "name" {
  description = "The name of the cluster template"
  value       = openstack_containerinfra_clustertemplate_v1.main.name
}

output "coe" {
  description = "The Container Orchestration Engine"
  value       = openstack_containerinfra_clustertemplate_v1.main.coe
}

output "region" {
  description = "The OpenStack region"
  value       = openstack_containerinfra_clustertemplate_v1.main.region
}
