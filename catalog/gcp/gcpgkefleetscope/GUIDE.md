# GcpGkeFleetScope Guide

The judgment this guide protects: a scope is a team's whole slice of the fleet -- its namespaces, its access, and its clusters -- and the clusters are bound by membership name, whichever way they joined the fleet.

## Binding clusters

A cluster joins a fleet in one of two ways. With `GcpGkeCluster.fleetProject` Google creates the membership itself and the cluster exports its name as `fleet_membership`; with `GcpGkeFleetMembership` the membership is declared explicitly and exports `name`. Either value goes into `membershipBindings[].membership`, which is why the bindings live here rather than on the membership: a cluster that joined at creation has no membership block to hold them. The membership must be in the scope's fleet project; both modules derive the binding's location and membership ID from the name.

## Namespaces and access

Each fleet namespace becomes a Kubernetes namespace on every bound cluster (an existing namespace of the same name is onboarded). Google reserves system namespaces such as `default`, `kube-system`, `gke-connect`, and `config-management-system`; the spec refuses them. `namespaceLabels` on the scope apply to all its namespaces and win over a namespace's own labels on the same key. Role bindings grant Google's `ADMIN`, `EDIT`, or `VIEW` in the scope's namespaces to one user or group each; a `customRole` takes effect only when the fleet's `rbacrolebindingactuation` feature allowlists it.

## Lifecycle

IDs are create-time decisions; labels, principals, roles, and a binding's target update in place. Destroy removes the bindings, the namespaces (from every bound cluster), and the scope -- in a chart, before the clusters it binds.
